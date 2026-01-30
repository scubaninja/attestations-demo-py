package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Config holds the application configuration
type Config struct {
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

// App represents the application
type App struct {
	Router *mux.Router
	DB     *sql.DB
	Config *Config
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Database  string    `json:"database"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Initialize sets up the application
func (a *App) Initialize() error {
	// Load environment variables
	_ = godotenv.Load()

	// Load configuration
	a.Config = &Config{
		Port:       getEnv("PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "attestations"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Initialize database connection if credentials are provided
	if a.Config.DBPassword != "" {
		connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			a.Config.DBHost,
			a.Config.DBPort,
			a.Config.DBUser,
			a.Config.DBPassword,
			a.Config.DBName,
			a.Config.DBSSLMode,
		)

		var err error
		a.DB, err = sql.Open("postgres", connStr)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}

		// Test the connection
		if err := a.DB.Ping(); err != nil {
			log.Printf("Warning: Database connection failed: %v", err)
			a.DB = nil
		} else {
			log.Println("Database connection established successfully")
			// Initialize database schema
			if err := a.initDB(); err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}
		}
	}

	// Initialize router
	a.Router = mux.NewRouter()
	a.initializeRoutes()

	return nil
}

// initDB creates the necessary database tables
func (a *App) initDB() error {
	if a.DB == nil {
		return nil
	}

	query := `
	CREATE TABLE IF NOT EXISTS attestations (
		id SERIAL PRIMARY KEY,
		artifact_name VARCHAR(255) NOT NULL,
		artifact_digest VARCHAR(255) NOT NULL,
		provenance_uri TEXT,
		sbom_uri TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_artifact_digest ON attestations(artifact_digest);
	CREATE INDEX IF NOT EXISTS idx_created_at ON attestations(created_at);
	`

	_, err := a.DB.Exec(query)
	return err
}

// initializeRoutes sets up the API routes
func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/", a.handleHome).Methods("GET")
	a.Router.HandleFunc("/health", a.handleHealth).Methods("GET")
	a.Router.HandleFunc("/api/v1/attestations", a.handleGetAttestations).Methods("GET")
	a.Router.HandleFunc("/api/v1/attestations", a.handleCreateAttestation).Methods("POST")
}

// handleHome handles the root endpoint
func (a *App) handleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"message": "GitHub Attestations Demo API",
		"version": "1.0.0",
		"docs":    "/health for health check",
	}
	json.NewEncoder(w).Encode(response)
}

// handleHealth handles the health check endpoint
func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbStatus := "not_configured"
	if a.DB != nil {
		if err := a.DB.Ping(); err == nil {
			dbStatus = "connected"
		} else {
			dbStatus = "disconnected"
		}
	}

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Database:  dbStatus,
	}
	json.NewEncoder(w).Encode(response)
}

// Attestation represents an attestation record
type Attestation struct {
	ID             int       `json:"id"`
	ArtifactName   string    `json:"artifact_name"`
	ArtifactDigest string    `json:"artifact_digest"`
	ProvenanceURI  string    `json:"provenance_uri,omitempty"`
	SBOMURI        string    `json:"sbom_uri,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// handleGetAttestations retrieves all attestations
func (a *App) handleGetAttestations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if a.DB == nil {
		json.NewEncoder(w).Encode([]Attestation{})
		return
	}

	rows, err := a.DB.Query(`
		SELECT id, artifact_name, artifact_digest, provenance_uri, sbom_uri, created_at, updated_at
		FROM attestations
		ORDER BY created_at DESC
		LIMIT 100
	`)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch attestations", err)
		return
	}
	defer rows.Close()

	attestations := []Attestation{}
	for rows.Next() {
		var att Attestation
		err := rows.Scan(&att.ID, &att.ArtifactName, &att.ArtifactDigest,
			&att.ProvenanceURI, &att.SBOMURI, &att.CreatedAt, &att.UpdatedAt)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to parse attestation", err)
			return
		}
		attestations = append(attestations, att)
	}

	json.NewEncoder(w).Encode(attestations)
}

// CreateAttestationRequest represents the request body for creating an attestation
type CreateAttestationRequest struct {
	ArtifactName   string `json:"artifact_name"`
	ArtifactDigest string `json:"artifact_digest"`
	ProvenanceURI  string `json:"provenance_uri,omitempty"`
	SBOMURI        string `json:"sbom_uri,omitempty"`
}

// handleCreateAttestation creates a new attestation
func (a *App) handleCreateAttestation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if a.DB == nil {
		respondWithError(w, http.StatusServiceUnavailable, "Database not configured", nil)
		return
	}

	var req CreateAttestationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate required fields
	if req.ArtifactName == "" || req.ArtifactDigest == "" {
		respondWithError(w, http.StatusBadRequest, "artifact_name and artifact_digest are required", nil)
		return
	}

	var attestation Attestation
	err := a.DB.QueryRow(`
		INSERT INTO attestations (artifact_name, artifact_digest, provenance_uri, sbom_uri)
		VALUES ($1, $2, $3, $4)
		RETURNING id, artifact_name, artifact_digest, provenance_uri, sbom_uri, created_at, updated_at
	`, req.ArtifactName, req.ArtifactDigest, req.ProvenanceURI, req.SBOMURI).Scan(
		&attestation.ID, &attestation.ArtifactName, &attestation.ArtifactDigest,
		&attestation.ProvenanceURI, &attestation.SBOMURI, &attestation.CreatedAt, &attestation.UpdatedAt,
	)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create attestation", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(attestation)
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, code int, message string, err error) {
	w.WriteHeader(code)
	response := ErrorResponse{
		Error:   http.StatusText(code),
		Message: message,
	}
	if err != nil {
		log.Printf("Error: %s - %v", message, err)
	}
	json.NewEncoder(w).Encode(response)
}

// Run starts the HTTP server
func (a *App) Run() error {
	addr := ":" + a.Config.Port
	log.Printf("Server starting on port %s", a.Config.Port)
	log.Printf("Health check available at http://localhost:%s/health", a.Config.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      a.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return srv.ListenAndServe()
}

// Close cleanly shuts down the application
func (a *App) Close() error {
	if a.DB != nil {
		return a.DB.Close()
	}
	return nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	app := &App{}

	if err := app.Initialize(); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer app.Close()

	log.Fatal(app.Run())
}

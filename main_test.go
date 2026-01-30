package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestHandleHome tests the home endpoint
func TestHandleHome(t *testing.T) {
	app := &App{}
	app.Initialize()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response["message"] != "GitHub Attestations Demo API" {
		t.Errorf("Unexpected message: got %v", response["message"])
	}
}

// TestHandleHealth tests the health check endpoint
func TestHandleHealth(t *testing.T) {
	app := &App{}
	app.Initialize()

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response HealthResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response.Status)
	}

	if response.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", response.Version)
	}

	if response.Database != "not_configured" && response.Database != "connected" && response.Database != "disconnected" {
		t.Errorf("Unexpected database status: %s", response.Database)
	}
}

// TestHandleGetAttestationsWithoutDB tests getting attestations without a database
func TestHandleGetAttestationsWithoutDB(t *testing.T) {
	app := &App{}
	app.Initialize()

	req, err := http.NewRequest("GET", "/api/v1/attestations", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var attestations []Attestation
	if err := json.NewDecoder(rr.Body).Decode(&attestations); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if len(attestations) != 0 {
		t.Errorf("Expected empty array, got %d items", len(attestations))
	}
}

// TestHandleGetAttestationsWithMockDB tests getting attestations with a mock database
func TestHandleGetAttestationsWithMockDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	app := &App{
		DB: db,
	}
	app.Initialize()

	rows := sqlmock.NewRows([]string{"id", "artifact_name", "artifact_digest", "provenance_uri", "sbom_uri", "created_at", "updated_at"}).
		AddRow(1, "test-artifact", "sha256:abc123", "https://example.com/provenance", "https://example.com/sbom", time.Now(), time.Now())

	mock.ExpectQuery("SELECT (.+) FROM attestations").WillReturnRows(rows)

	req, err := http.NewRequest("GET", "/api/v1/attestations", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var attestations []Attestation
	if err := json.NewDecoder(rr.Body).Decode(&attestations); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if len(attestations) != 1 {
		t.Errorf("Expected 1 attestation, got %d", len(attestations))
	}

	if attestations[0].ArtifactName != "test-artifact" {
		t.Errorf("Expected artifact name 'test-artifact', got '%s'", attestations[0].ArtifactName)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestHandleCreateAttestationWithoutDB tests creating attestation without database
func TestHandleCreateAttestationWithoutDB(t *testing.T) {
	app := &App{}
	app.Initialize()

	reqBody := CreateAttestationRequest{
		ArtifactName:   "test-artifact",
		ArtifactDigest: "sha256:abc123",
		ProvenanceURI:  "https://example.com/provenance",
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/api/v1/attestations", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusServiceUnavailable {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusServiceUnavailable)
	}
}

// TestHandleCreateAttestationWithMockDB tests creating attestation with mock database
func TestHandleCreateAttestationWithMockDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	app := &App{
		DB: db,
	}
	app.Initialize()

	reqBody := CreateAttestationRequest{
		ArtifactName:   "test-artifact",
		ArtifactDigest: "sha256:abc123",
		ProvenanceURI:  "https://example.com/provenance",
		SBOMURI:        "https://example.com/sbom",
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "artifact_name", "artifact_digest", "provenance_uri", "sbom_uri", "created_at", "updated_at"}).
		AddRow(1, reqBody.ArtifactName, reqBody.ArtifactDigest, reqBody.ProvenanceURI, reqBody.SBOMURI, now, now)

	mock.ExpectQuery("INSERT INTO attestations").
		WithArgs(reqBody.ArtifactName, reqBody.ArtifactDigest, reqBody.ProvenanceURI, reqBody.SBOMURI).
		WillReturnRows(rows)

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/api/v1/attestations", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var attestation Attestation
	if err := json.NewDecoder(rr.Body).Decode(&attestation); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if attestation.ArtifactName != reqBody.ArtifactName {
		t.Errorf("Expected artifact name '%s', got '%s'", reqBody.ArtifactName, attestation.ArtifactName)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestHandleCreateAttestationInvalidRequest tests creating attestation with invalid request
func TestHandleCreateAttestationInvalidRequest(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	app := &App{
		DB: db,
	}
	app.Initialize()

	// Test with missing required fields
	reqBody := CreateAttestationRequest{
		ArtifactName: "test-artifact",
		// Missing ArtifactDigest
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/api/v1/attestations", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

// TestGetEnv tests the getEnv function
func TestGetEnv(t *testing.T) {
	tests := []struct {
		key          string
		defaultValue string
		expected     string
	}{
		{"NON_EXISTENT_KEY", "default", "default"},
		{"PORT", "8080", "8080"},
	}

	for _, tt := range tests {
		result := getEnv(tt.key, tt.defaultValue)
		if result != tt.expected {
			t.Errorf("getEnv(%s, %s) = %s; want %s", tt.key, tt.defaultValue, result, tt.expected)
		}
	}
}

// TestAppInitialization tests the app initialization
func TestAppInitialization(t *testing.T) {
	app := &App{}
	err := app.Initialize()
	if err != nil {
		t.Errorf("App initialization failed: %v", err)
	}

	if app.Router == nil {
		t.Error("Router was not initialized")
	}

	if app.Config == nil {
		t.Error("Config was not initialized")
	}
}

// BenchmarkHandleHome benchmarks the home endpoint
func BenchmarkHandleHome(b *testing.B) {
	app := &App{}
	app.Initialize()

	req, _ := http.NewRequest("GET", "/", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		app.Router.ServeHTTP(rr, req)
	}
}

// BenchmarkHandleHealth benchmarks the health endpoint
func BenchmarkHandleHealth(b *testing.B) {
	app := &App{}
	app.Initialize()

	req, _ := http.NewRequest("GET", "/health", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		app.Router.ServeHTTP(rr, req)
	}
}

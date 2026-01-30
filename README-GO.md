# GitHub Attestations Demo - Go Version

This repository demonstrates the use of GitHub Actions attestations with a Go application, including deployment to Azure with PostgreSQL database.

## 🚀 Features

- **Go HTTP API**: RESTful API built with Go and Gorilla Mux
- **PostgreSQL Integration**: Full database support with automatic schema migration
- **Azure Deployment**: Complete Bicep templates for Azure Container Apps and PostgreSQL
- **Comprehensive Tests**: Unit tests with mocking and benchmarks
- **GitHub Attestations**: Build provenance and SBOM generation
- **Docker Support**: Multi-stage Docker builds for optimal image size
- **Best Practices**: Following Go standards and conventions

## 📋 API Endpoints

- `GET /` - API information
- `GET /health` - Health check with database status
- `GET /api/v1/attestations` - List all attestations
- `POST /api/v1/attestations` - Create new attestation

## 🏗️ Architecture

The application is built with:
- **Go 1.21**: Modern Go with generics support
- **Gorilla Mux**: HTTP router for REST API
- **PostgreSQL**: Relational database for storing attestations
- **Azure Container Apps**: Serverless container hosting
- **Azure PostgreSQL Flexible Server**: Managed PostgreSQL database

## 🛠️ Local Development

### Prerequisites

- Go 1.21 or higher
- PostgreSQL (optional, for database features)
- Docker (optional, for containerization)

### Running Locally

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Run the application:**
   ```bash
   go run main.go
   ```

3. **With database (optional):**
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=postgres
   export DB_PASSWORD=yourpassword
   export DB_NAME=attestations
   export DB_SSLMODE=disable
   
   go run main.go
   ```

4. **Access the application:**
   - API: http://localhost:8080
   - Health: http://localhost:8080/health

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. -benchmem ./...
```

### Linting

```bash
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Run linter
golangci-lint run
```

## 🐳 Docker

### Build Docker Image

```bash
docker build -f Dockerfile.go -t attestations-demo:latest .
```

### Run Docker Container

```bash
# Without database
docker run -p 8080:8080 attestations-demo:latest

# With database
docker run -p 8080:8080 \
  -e DB_HOST=host.docker.internal \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=yourpassword \
  -e DB_NAME=attestations \
  -e DB_SSLMODE=disable \
  attestations-demo:latest
```

## ☁️ Azure Deployment

Complete deployment instructions are available in [azure/DEPLOYMENT.md](azure/DEPLOYMENT.md).

### Quick Deploy

```bash
# Set variables
export RESOURCE_GROUP="attestations-demo-rg"
export LOCATION="eastus"
export DB_ADMIN_PASSWORD="YourSecurePassword123!"

# Create resource group
az group create --name $RESOURCE_GROUP --location $LOCATION

# Deploy infrastructure
az deployment group create \
  --resource-group $RESOURCE_GROUP \
  --template-file azure/main.bicep \
  --parameters \
    location=$LOCATION \
    environmentName="attestations-demo" \
    dbAdminUsername="pgadmin" \
    dbAdminPassword=$DB_ADMIN_PASSWORD
```

See [azure/DEPLOYMENT.md](azure/DEPLOYMENT.md) for detailed instructions.

## 🔒 GitHub Attestations

This project demonstrates GitHub Actions attestations for supply chain security:

### Build Provenance Attestation

The workflow automatically generates build provenance attestations for compiled binaries:

```yaml
- name: Attest build provenance
  uses: actions/attest-build-provenance@v1
  with:
    subject-path: "bin/attestations-demo*"
```

### Verification

Verify attestations using GitHub CLI:

```bash
gh attestation verify bin/attestations-demo-linux-amd64 --owner scubaninja
```

## 📊 Database Schema

The application automatically creates the following schema:

```sql
CREATE TABLE attestations (
    id SERIAL PRIMARY KEY,
    artifact_name VARCHAR(255) NOT NULL,
    artifact_digest VARCHAR(255) NOT NULL,
    provenance_uri TEXT,
    sbom_uri TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 🧪 Testing Strategy

The project includes:

- **Unit Tests**: Testing individual functions and handlers
- **Integration Tests**: Testing with mock databases using go-sqlmock
- **Benchmarks**: Performance testing for critical paths
- **Table-driven Tests**: Following Go best practices
- **Race Detection**: Using `-race` flag to detect race conditions

## 📝 Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | (empty) |
| `DB_NAME` | Database name | `attestations` |
| `DB_SSLMODE` | SSL mode | `disable` |

## 🔍 API Examples

### Create Attestation

```bash
curl -X POST http://localhost:8080/api/v1/attestations \
  -H "Content-Type: application/json" \
  -d '{
    "artifact_name": "my-app",
    "artifact_digest": "sha256:abc123...",
    "provenance_uri": "https://github.com/org/repo/attestations/123",
    "sbom_uri": "https://github.com/org/repo/attestations/124"
  }'
```

### List Attestations

```bash
curl http://localhost:8080/api/v1/attestations
```

### Health Check

```bash
curl http://localhost:8080/health
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is open source and available under the MIT License.

## 🔗 Related Resources

- [GitHub Actions Attestations](https://docs.github.com/en/actions/security-guides/using-artifact-attestations-to-establish-provenance-for-builds)
- [SLSA Framework](https://slsa.dev/)
- [Azure Container Apps](https://learn.microsoft.com/azure/container-apps/)
- [Go Best Practices](https://go.dev/doc/effective_go)

## 📚 Migration Notes

This repository was migrated from Python to Go with the following improvements:

- **Performance**: Go's compiled nature provides better performance
- **Concurrency**: Built-in goroutines for concurrent request handling
- **Type Safety**: Static typing catches errors at compile time
- **Deployment**: Single binary deployment simplifies operations
- **Testing**: Comprehensive test suite with mocking
- **Azure Integration**: Native support for Azure services with automatic schema migration

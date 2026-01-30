# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY main.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage - using distroless for minimal size and security
FROM gcr.io/distroless/static-debian11

WORKDIR /

# Copy the binary from builder
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Run the application
CMD ["/main"]

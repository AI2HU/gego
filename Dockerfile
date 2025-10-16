# Production Dockerfile for Gego - GEO Tracker
# Optimized for production deployment with minimal image size

FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata sqlite-dev gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o gego ./cmd/gego/main.go

# Stage 2: Minimal runtime
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Copy binary
COPY --from=builder /app/gego /usr/local/bin/gego

# Copy migration files
COPY --from=builder /app/internal/db/migrations /migrations

# Create directories
RUN mkdir -p /app/data /app/config /app/logs

# Set environment variables
ENV GEGO_CONFIG_PATH=/app/config/config.yaml
ENV GEGO_DATA_PATH=/app/data
ENV GEGO_LOG_PATH=/app/logs

# Create default configuration
RUN echo 'sql_database:\n  provider: sqlite\n  uri: /app/data/gego.db\n  database: gego\n\nnosql_database:\n  provider: mongodb\n  uri: mongodb://mongodb:27017\n  database: gego' > /app/config/config.yaml

# Expose port
EXPOSE 8989

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/usr/local/bin/gego", "api", "--help"] || exit 1

# Default command
CMD ["/usr/local/bin/gego", "api", "--host", "0.0.0.0", "--port", "8989"]

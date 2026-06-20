# Production Dockerfile for Gego - GEO Tracker
# Builds backend and gego-ui into a single image

# Stage 1: Build UI
FROM node:22-alpine AS ui-builder

WORKDIR /app/gego-ui

COPY gego-ui/package.json gego-ui/package-lock.json ./
RUN npm ci

COPY gego-ui/ ./
RUN npm run build-only

# Stage 2: Build Go backend
FROM golang:1.25-alpine AS go-builder

RUN apk add --no-cache git ca-certificates tzdata sqlite-dev gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o gego ./cmd/gego/main.go

# Stage 3: Minimal runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

COPY --from=go-builder /app/gego /usr/local/bin/gego
COPY --from=go-builder /app/internal/db/migrations /migrations
COPY --from=ui-builder /app/gego-ui/dist /app/ui

RUN mkdir -p /app/data /app/config /app/logs

ENV GEGO_CONFIG_PATH=/app/config/config.yaml
ENV GEGO_DATA_PATH=/app/data
ENV GEGO_LOG_PATH=/app/logs

RUN echo 'sql_database:\n  provider: sqlite\n  uri: /app/data/gego.db\n  database: gego\n\nnosql_database:\n  provider: mongodb\n  uri: mongodb://mongodb:27017\n  database: gego' > /app/config/config.yaml

EXPOSE 8989

CMD ["/usr/local/bin/gego", "api", "--host", "0.0.0.0", "--port", "8989"]

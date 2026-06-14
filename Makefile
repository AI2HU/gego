.PHONY: build run clean install test fmt vet help deps build-all \
	ui-install dev dev-api dev-ui

# Build variables
BINARY_NAME=gego
BUILD_DIR=build
MAIN_PATH=cmd/gego/main.go
UI_DIR=gego-ui

# Dev variables (override via environment or: make dev GEGO_JWT_SECRET=...)
API_PORT ?= 8989
UI_PORT ?= 5173
GEGO_JWT_SECRET ?= change-me-to-a-secret-at-least-32-chars
GEGO_BOOTSTRAP_ADMIN_PASSWORD ?= admin1234
GEGO_BOOTSTRAP_ADMIN_USERNAME ?= admin

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Install the application
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(MAIN_PATH)
	@echo "Installation complete"

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run all checks
check: fmt vet test

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Multi-platform build complete"

ui-install:
	@echo "Installing UI dependencies..."
	cd $(UI_DIR) && npm install

dev-api: build
	@echo "Starting API on http://localhost:$(API_PORT)/api/v1"
	GEGO_JWT_SECRET=$(GEGO_JWT_SECRET) \
	GEGO_BOOTSTRAP_ADMIN_PASSWORD=$(GEGO_BOOTSTRAP_ADMIN_PASSWORD) \
	GEGO_BOOTSTRAP_ADMIN_USERNAME=$(GEGO_BOOTSTRAP_ADMIN_USERNAME) \
	./$(BUILD_DIR)/$(BINARY_NAME) api --port $(API_PORT)

dev-ui: ui-install
	@echo "Starting UI on http://localhost:$(UI_PORT)"
	cd $(UI_DIR) && npm run dev -- --port $(UI_PORT)

dev: build ui-install
	@echo "Starting API on http://localhost:$(API_PORT)/api/v1 and UI on http://localhost:$(UI_PORT)"
	@trap 'kill 0' INT TERM EXIT; \
	GEGO_JWT_SECRET=$(GEGO_JWT_SECRET) \
	GEGO_BOOTSTRAP_ADMIN_PASSWORD=$(GEGO_BOOTSTRAP_ADMIN_PASSWORD) \
	GEGO_BOOTSTRAP_ADMIN_USERNAME=$(GEGO_BOOTSTRAP_ADMIN_USERNAME) \
	./$(BUILD_DIR)/$(BINARY_NAME) api --port $(API_PORT) & \
	cd $(UI_DIR) && npm run dev -- --port $(UI_PORT) & \
	wait

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build the application"
	@echo "  install    - Install the application"
	@echo "  run        - Build and run the application"
	@echo "  clean      - Remove build artifacts"
	@echo "  test       - Run tests"
	@echo "  fmt        - Format code"
	@echo "  vet        - Run go vet"
	@echo "  check      - Run fmt, vet, and test"
	@echo "  deps       - Download and tidy dependencies"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  ui-install - Install gego-ui npm dependencies"
	@echo "  dev        - Start API and UI together"
	@echo "  dev-api    - Start API only (port $(API_PORT))"
	@echo "  dev-ui     - Start UI only (port $(UI_PORT))"
	@echo "  help       - Show this help message"

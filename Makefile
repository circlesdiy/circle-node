.PHONY: help build run dev test clean docker-up docker-down docker-logs migrate

# Default target
help:
	@echo "Circles.DIY Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make build        - Build the application binary"
	@echo "  make run          - Run the application"
	@echo "  make dev          - Run in development mode with auto-reload"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docker-up    - Start all services with Docker Compose"
	@echo "  make docker-down  - Stop all services"
	@echo "  make docker-logs  - View logs from all services"
	@echo "  make migrate      - Run database migrations"
	@echo ""

# Build the application
build:
	@echo "Building application..."
	go build -o bin/circles-diy .
	@echo "Build complete: bin/circles-diy"

# Run the application
run: build
	@echo "Starting application..."
	./bin/circles-diy

# Development mode
dev:
	@echo "Starting development server..."
	@echo "Watching for changes..."
	go run main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f static/css/style.css
	@echo "Clean complete"

# Docker operations
docker-up:
	@echo "Starting services..."
	docker-compose up -d
	@echo "Services started. Use 'make docker-logs' to view logs"

docker-down:
	@echo "Stopping services..."
	docker-compose down
	@echo "Services stopped"

docker-logs:
	docker-compose logs -f

docker-build:
	@echo "Building Docker image..."
	docker-compose build
	@echo "Build complete"

# Database operations
migrate:
	@echo "Migrations run automatically on startup"
	@echo "To manually test migrations, ensure database is running and run the app"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatted"

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	golangci-lint run ./...

# Generate documentation
docs:
	@echo "Generating documentation..."
	godoc -http=:6060

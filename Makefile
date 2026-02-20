.PHONY: help build run test test-short test-coverage test-bench test-verbose lint fmt clean package deps

# Default target
help:
	@echo "Mosugo - Makefile targets:"
	@echo "  make build          - Build the executable"
	@echo "  make run            - Run the application in development mode"
	@echo "  make test           - Run tests with race detector and coverage"
	@echo "  make test-short     - Run tests quickly (skip slow tests)"
	@echo "  make test-coverage  - Run tests and open HTML coverage report"
	@echo "  make test-bench     - Run benchmark tests"
	@echo "  make test-verbose   - Run tests with verbose output"
	@echo "  make lint           - Run golangci-lint"
	@echo "  make fmt            - Format code with gofmt and goimports"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make package        - Create packaged executable with Fyne"
	@echo "  make deps           - Download and verify dependencies"

# Build the executable
build:
	@echo "Building Mosugo..."
	go build -o mosugo.exe cmd/mosugo/main.go
	@echo "Build complete: mosugo.exe"

# Run in development mode
run:
	@echo "Running Mosugo..."
	go run cmd/mosugo/main.go

# Run tests with coverage
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "Coverage report generated: coverage.txt"
	@echo "View coverage: go tool cover -html=coverage.txt"

# Run tests quickly (skip slow tests)
test-short:
	@echo "Running short tests..."
	go test -short -v ./internal/...

# Run tests and open HTML coverage report
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./internal/...
	@echo "Opening coverage report in browser..."
	go tool cover -html=coverage.out

# Run benchmark tests
test-bench:
	@echo "Running benchmarks..."
	go test ./internal/... -bench=. -benchmem -run=^Benchmark

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	go test -v -race ./internal/...

# Run linters
lint:
	@echo "Running golangci-lint..."
	golangci-lint run --timeout=5m

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .
	@echo "Code formatted"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@if exist mosugo.exe del /F /Q mosugo.exe
	@if exist Mosugo.exe del /F /Q Mosugo.exe
	@if exist coverage.txt del /F /Q coverage.txt
	@if exist coverage.out del /F /Q coverage.out
	@echo "Clean complete"

# Package with Fyne
package:
	@echo "Packaging Mosugo with Fyne..."
	fyne package -os windows -name Mosugo -icon assets/Mosugo_Icon.png
	@echo "Package complete: Mosugo.exe"

# Download and verify dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod verify
	@echo "Dependencies ready"

# Install development tools
tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install fyne.io/fyne/v2/cmd/fyne@latest
	@echo "Tools installed"

.PHONY: help test test-coverage test-benchmark build clean lint format check-deps install-tools

# Default target
help:
	@echo "Available targets:"
	@echo "  test            - Run all tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  test-benchmark  - Run benchmark tests"
	@echo "  build           - Build the project"
	@echo "  clean           - Clean build artifacts"
	@echo "  lint            - Run linter"
	@echo "  format          - Format code"
	@echo "  check-deps      - Check for outdated dependencies"
	@echo "  install-tools   - Install development tools"

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmark tests
test-benchmark:
	@echo "Running benchmark tests..."
	go test -bench=. -benchmem ./...

# Build the project
build:
	@echo "Building project..."
	go build ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	go clean
	rm -f coverage.out coverage.html

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Format code
format:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Check for outdated dependencies
check-deps:
	@echo "Checking for outdated dependencies..."
	go list -u -m all

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

# Pre-commit hook
pre-commit: format lint test
	@echo "Pre-commit checks passed!"

# CI pipeline
ci: test-coverage test-benchmark lint
	@echo "CI pipeline completed successfully!"

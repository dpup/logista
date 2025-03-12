# Logista Development Guide

## Build & Test Commands
- `make build` - Build binary in dist/
- `make test` - Run all tests
- `make test-coverage` - Run tests with coverage report
- `go test ./internal/formatter -run TestFormatSimpleTemplate` - Run single test
- `make lint` - Run golangci-lint
- `make fmt` - Format code (go fmt, goimports, gofmt)
- `make fmt-check` - Verify formatting
- `make run-test` - Test with sample logs
- `make all` - Run lint, test, and build

## Code Style Guidelines
- Follow standard Go formatting (gofmt)
- Use 4-space indentation
- Employ functional options pattern for configuration
- Group related methods together
- Use descriptive variable names (camelCase)
- Document all exported functions and types
- Write comprehensive table-driven tests
- Return errors rather than using panics
- Use the internal/ directory for non-public packages
- Prefer interfaces for flexibility
- Handle all errors explicitly

## Project Structure
- cmd/ - Command-line interface
- internal/ - Core implementation
- test/ - Testing scripts and data
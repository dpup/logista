.PHONY: build clean test test-coverage shell-test run-test lint fmt fmt-check install release all help

# Variables
BINARY_NAME=logista
DIST_DIR=dist
MAIN_PACKAGE=./cmd/logista
VERSION=$(shell git describe --tags --always 2>/dev/null || echo "dev")
BUILD_FLAGS=-ldflags "-X main.version=${VERSION}"
GOPATH=$(shell go env GOPATH)

# Build the binary
build:
	mkdir -p ${DIST_DIR}
	go build ${BUILD_FLAGS} -o ${DIST_DIR}/${BINARY_NAME} ${MAIN_PACKAGE}

# Clean build artifacts
clean:
	rm -rf ${DIST_DIR}
	go clean

# Run tests
test: shell-test
	go test -v ./...

# Run the shell test script
shell-test:
	@echo "Running shell tests..."
	./test/run_test.sh

# Run the tool manually with test logs
run-test: build
	./test/print_basic_logs.sh
	./test/print_grpc_logs.sh

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Run linters
lint:
	go tool golangci-lint run ./...

# Format code
fmt:
	go fmt ./...
	go tool goimports -w .
	gofmt -s -w .

# Check for formatting errors
fmt-check:
	@if [ -n "$(shell gofmt -l .)" ]; then \
		echo "Go code is not formatted, run 'make fmt'"; \
		exit 1; \
	fi
	@echo "Go code is formatted"

# Install the binary
install:
	go install ${BUILD_FLAGS} ${MAIN_PACKAGE}

# Create a new release
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Use 'make release VERSION=x.y.z'"; \
		exit 1; \
	fi
	git tag -a v$(VERSION) -m "Release $(VERSION)"
	git push origin v$(VERSION)

# Default target
all: lint test build

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run all tests (unit tests and shell tests)"
	@echo "  shell-test    - Run shell tests to verify CLI functionality"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run linters"
	@echo "  fmt           - Format code"
	@echo "  fmt-check     - Check for formatting errors"
	@echo "  install       - Install the binary"
	@echo "  release       - Create a new release (requires VERSION=x.y.z)"
	@echo "  run-test      - Run the tool manually with test logs"
	@echo "  all           - Run lint, test and build"
	@echo "  help          - Show this help"
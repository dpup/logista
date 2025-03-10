#!/bin/bash
set -e

# Navigate to project root
cd "$(dirname "$0")/.."

# Build the binary
make build

# Path to the binary and test data
BINARY="./dist/logista"
TEST_DATA="./test/test-logs.json"

# Tests with default format
echo "=== Testing default format ==="
DEFAULT_OUTPUT=$(cat "$TEST_DATA" | "$BINARY")
echo "$DEFAULT_OUTPUT"

# Check for expected output with default format
if echo "$DEFAULT_OUTPUT" | grep -q "2025-03-10T15:04:05Z info Application started" && \
   echo "$DEFAULT_OUTPUT" | grep -q "2025-03-10T15:04:06Z debug Configuration loaded" && \
   echo "$DEFAULT_OUTPUT" | grep -q "2025-03-10T15:04:07Z error Failed to connect to database"; then
  echo -e "\033[0;32mDefault format test: PASSED\033[0m"
else
  echo -e "\033[0;31mDefault format test: FAILED\033[0m"
  exit 1
fi

# Test with custom format
echo ""
echo "=== Testing custom format ==="
CUSTOM_OUTPUT=$(cat "$TEST_DATA" | "$BINARY" --fmt="{timestamp} [{level}] {message}")
echo "$CUSTOM_OUTPUT"

# Check for expected output with custom format
if echo "$CUSTOM_OUTPUT" | grep -q "2025-03-10T15:04:05Z \[info\] Application started" && \
   echo "$CUSTOM_OUTPUT" | grep -q "2025-03-10T15:04:06Z \[debug\] Configuration loaded" && \
   echo "$CUSTOM_OUTPUT" | grep -q "2025-03-10T15:04:07Z \[error\] Failed to connect to database"; then
  echo -e "\033[0;32mCustom format test: PASSED\033[0m"
else
  echo -e "\033[0;31mCustom format test: FAILED\033[0m"
  exit 1
fi

echo -e "\033[0;32mAll tests passed!\033[0m"
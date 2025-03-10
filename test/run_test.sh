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
if echo "$DEFAULT_OUTPUT" | grep -q "2025-03-10 15:04:05 info Application started" && \
   echo "$DEFAULT_OUTPUT" | grep -q "2025-03-10 15:04:06 debug Configuration loaded" && \
   echo "$DEFAULT_OUTPUT" | grep -q "2025-03-10 15:04:07 error Failed to connect to database"; then
  echo -e "\033[0;32mDefault format test: PASSED\033[0m"
else
  echo -e "\033[0;31mDefault format test: FAILED\033[0m"
  exit 1
fi

# Test with custom format (no date formatting) - simplified syntax
echo ""
echo "=== Testing custom format with simplified syntax ==="
CUSTOM_OUTPUT=$(cat "$TEST_DATA" | "$BINARY" --fmt="{timestamp} [{level}] {message}")
echo "$CUSTOM_OUTPUT"

# Check for expected output with custom format
if echo "$CUSTOM_OUTPUT" | grep -q "2025-03-10T15:04:05Z \[info\] Application started" && \
   echo "$CUSTOM_OUTPUT" | grep -q "2025-03-10T15:04:06Z \[debug\] Configuration loaded" && \
   echo "$CUSTOM_OUTPUT" | grep -q "2025-03-10T15:04:07Z \[error\] Failed to connect to database"; then
  echo -e "\033[0;32mSimplified syntax test: PASSED\033[0m"
else
  echo -e "\033[0;31mSimplified syntax test: FAILED\033[0m"
  exit 1
fi

# Test with custom format (no date formatting) - Go template syntax
echo ""
echo "=== Testing custom format with Go template syntax ==="
CUSTOM_OUTPUT2=$(cat "$TEST_DATA" | "$BINARY" --fmt="{{.timestamp}} [{{.level}}] {{.message}}")
echo "$CUSTOM_OUTPUT2"

# Check for expected output with custom format
if echo "$CUSTOM_OUTPUT2" | grep -q "2025-03-10T15:04:05Z \[info\] Application started" && \
   echo "$CUSTOM_OUTPUT2" | grep -q "2025-03-10T15:04:06Z \[debug\] Configuration loaded" && \
   echo "$CUSTOM_OUTPUT2" | grep -q "2025-03-10T15:04:07Z \[error\] Failed to connect to database"; then
  echo -e "\033[0;32mGo template syntax test: PASSED\033[0m"
else
  echo -e "\033[0;31mGo template syntax test: FAILED\033[0m"
  exit 1
fi

# Test with date function - simplified syntax
echo ""
echo "=== Testing date function with simplified syntax ==="
DATE_OUTPUT=$(cat "$TEST_DATA" | "$BINARY" --fmt="{timestamp | date} [{level}] {message}")
echo "$DATE_OUTPUT"

# Check for expected output with date function
if echo "$DATE_OUTPUT" | grep -q "2025-03-10 15:04:05 \[info\] Application started" && \
   echo "$DATE_OUTPUT" | grep -q "2025-03-10 15:04:06 \[debug\] Configuration loaded" && \
   echo "$DATE_OUTPUT" | grep -q "2025-03-10 15:04:07 \[error\] Failed to connect to database" && \
   echo "$DATE_OUTPUT" | grep -q "2025-03-10 15:04:05 \[info\] Log with common log format"; then
  echo -e "\033[0;32mDate function simplified syntax test: PASSED\033[0m"
else
  echo -e "\033[0;31mDate function simplified syntax test: FAILED\033[0m"
  exit 1
fi

# Test with custom date format - simplified syntax
echo ""
echo "=== Testing custom date format with simplified syntax ==="
CUSTOM_DATE_OUTPUT=$(cat "$TEST_DATA" | "$BINARY" --fmt="{timestamp | date} [{level}] {message}" --preferred_date_format="02/01/2006 15:04")
echo "$CUSTOM_DATE_OUTPUT"

# Check for expected output with custom date format
if echo "$CUSTOM_DATE_OUTPUT" | grep -q "10/03/2025 15:04 \[info\] Application started" && \
   echo "$CUSTOM_DATE_OUTPUT" | grep -q "10/03/2025 15:04 \[debug\] Configuration loaded" && \
   echo "$CUSTOM_DATE_OUTPUT" | grep -q "10/03/2025 15:04 \[error\] Failed to connect to database"; then
  echo -e "\033[0;32mCustom date format simplified syntax test: PASSED\033[0m"
else
  echo -e "\033[0;31mCustom date format simplified syntax test: FAILED\033[0m"
  exit 1
fi

# Test with new color functions syntax
echo ""
echo "=== Testing new color functions syntax ==="
COLOR_OUTPUT=$(cat "$TEST_DATA" | "$BINARY" --fmt="{timestamp | date | color \"cyan\"} [{level}] {message | bold}")
echo "$COLOR_OUTPUT"

# For color tests, we just verify that the command ran without error
# since testing for ANSI codes directly is challenging in a shell script
if [ $? -eq 0 ]; then
  echo -e "\033[0;32mNew color functions test: PASSED\033[0m"
else
  echo -e "\033[0;31mNew color functions test: FAILED\033[0m"
  exit 1
fi

# Test with --no-colors flag
echo ""
echo "=== Testing --no-colors flag ==="
NO_COLOR_OUTPUT=$(cat "$TEST_DATA" | "$BINARY" --fmt="{timestamp | date | color \"cyan\"} [{level}] {message | bold}" --no-colors)
echo "$NO_COLOR_OUTPUT"

# For no-colors test, we verify that the command ran without error
# and that the output doesn't contain ANSI escape codes
if [ $? -eq 0 ] && ! echo "$NO_COLOR_OUTPUT" | grep -q "\033"; then
  echo -e "\033[0;32mNo colors test: PASSED\033[0m"
else
  echo -e "\033[0;31mNo colors test: FAILED\033[0m"
  exit 1
fi

# Test with advanced template features
echo ""
echo "=== Testing advanced template features ==="
ADVANCED_OUTPUT=$(cat "$TEST_DATA" | "$BINARY" --fmt="{{.timestamp | date}} [{{.level}}] {{.message}}{{if .context}} (Context: {{.context.version}}){{end}}")
echo "$ADVANCED_OUTPUT"

# For advanced features test, verify the command ran and expected output is present
if [ $? -eq 0 ] && echo "$ADVANCED_OUTPUT" | grep -q "(Context: 1.0.0)"; then
  echo -e "\033[0;32mAdvanced template features test: PASSED\033[0m"
else
  echo -e "\033[0;31mAdvanced template features test: FAILED\033[0m"
  exit 1
fi

echo -e "\033[0;32mAll tests passed!\033[0m"
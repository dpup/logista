#!/bin/bash
set -e

# Navigate to project root
cd "$(dirname "$0")/.."

# Example showing how log level can be used to format messages differently:
# - Error level messages are red and bold
# - Warning level messages are yellow
# - Info level messages are normal
# - Debug level messages are dim
# - All other levels remain at default style

echo -e "\n=== Using conditionals to power format ==="
cat test/basic-logs.json | dist/logista \
  --date_format='15:04:05' \
  --format='{{.timestamp | date | color "cyan"}} [{{.level | pad 7}}] {{if eq .level "error"}}{{.message | color "red" | bold}}{{else if eq .level "warning"}}{{.message | color "yellow"}}{{else if eq .level "info"}}{{.message}}{{else if eq .level "debug"}}{{.message | dim}}{{else}}{{.message}}{{end}}'

echo -e "\n\n=== Using colorByLevel (simpler alternative) ==="
cat test/basic-logs.json | dist/logista \
  --date_format='15:04:05' \
  --format='{{.timestamp | date | color "cyan"}} [{{.level | pad 7}}] {{.message | colorByLevel .level}}'

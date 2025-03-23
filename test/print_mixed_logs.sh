#!/bin/bash
set -e

cd "$(dirname "$0")/.."

cat test/mixed-input.txt | dist/logista \
  --format='{{.level | pad 7 | colorByLevel .level}} {{.message | colorByLevel .level | bold}}' \
  --handle_non_json


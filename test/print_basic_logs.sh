#!/bin/bash
set -e

# Navigate to project root
cd "$(dirname "$0")/.."

cat test/basic-logs.json | dist/logista  \
	--preferred_date_format='15:04:05' \
	--fmt='{timestamp | date | color "red"} {level | pad 8 | colorByLevel .level} {message | colorByLevel .level}'
#!/bin/bash
set -e

cd "$(dirname "$0")/.."

# Advanced formatting template that:
# 1. Shows timestamp, level, and logger in the first line
# 2. Highlights the message based on log level
# 3. Groups gRPC-related fields into a separate section
# 4. Shows auth information in another section
# 5. Shows remaining fields at the end
#
# THIS IS A PROPOSED VERSION with a hypothetical "field" function that handles both normal and dotted fields

# Note: This is a PROPOSAL and won't work until the "field" function is added to the formatter

cat test/grpc-logs.json | dist/logista \
  --preferred_date_format='15:04:05.000' \
  --fmt='{{field "ts" | date | color "cyan"}} {{field "level" | pad 7 | colorByLevel (field "level")}} [{{field "logger"}}] {{field "msg" | colorByLevel (field "level") | bold}}
{{if field "grpc.service"}}  {{field "grpc.component" | color "blue" | pad 10}} {{field "grpc.service"}}.{{field "grpc.method"}} ({{field "grpc.method_type" | color "magenta"}}){{if field "grpc.time_ms"}} took {{field "grpc.time_ms"}}ms{{end}}{{if field "grpc.code"}} → {{field "grpc.code" | colorByLevel (field "level")}}{{end}}
{{end}}{{if or (field "authz.action") (field "authz.roles")}}  {{color "yellow" "AUTH:     "}} {{if field "authz.action"}}action={{field "authz.action"}}{{end}} {{if field "authz.roles"}}roles={{field "authz.roles"}}{{end}}
{{end}}{{range $key, $value := .}}{{if and 
  (not (isStandardField $key))
  (not (hasPrefix $key "grpc."))
  (not (hasPrefix $key "authz.")) 
}}  {{$key | dim | pad 23}}: {{$value}}
{{end}}{{end}}'
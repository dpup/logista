#!/bin/bash
set -e

cd "$(dirname "$0")/.."

# Advanced formatting template that:
# 1. Shows timestamp, level, and logger in the first line
# 2. Highlights the message based on log level
# 3. Groups gRPC-related fields into a separate section
# 4. Shows auth information in another section
# 5. Shows remaining fields at the end

cat test/grpc-logs.json | dist/logista \
  --preferred_date_format='15:04:05.000' \
  --fmt='{{.ts | date | color "cyan"}} {{.level | pad 7 | colorByLevel .level}} {{.msg | colorByLevel .level | bold}}
{{if index . "grpc.service"}}  {{"grpc" | color "blue"}} {{index . "grpc.component" | color "blue"}} {{index . "grpc.service"}}.{{index . "grpc.method"}} ({{index . "grpc.method_type" | color "magenta"}}){{if index . "grpc.time_ms"}} {{index . "grpc.time_ms"}}ms{{end}}{{if index . "grpc.code"}} → {{index . "grpc.code" | colorByLevel .level}}{{end}}
{{end}}{{if or (index . "authz.action") (index . "authz.roles")}}  {{color "yellow" "auth" | pad 26}} : {{if index . "authz.action"}}action={{index . "authz.action" | color "yellow"}}{{end}} {{if index . "authz.roles"}}roles={{index . "authz.roles" | color "yellow"}}{{end}}
{{end}}{{range $key, $value := .}}{{if and 
  (not (isStandardField $key))
  (not (hasPrefix $key "grpc."))
  (not (hasPrefix $key "authz.")) 
}}  {{$key | dim | pad 26}}: {{$value}}
{{end}}{{end}}'

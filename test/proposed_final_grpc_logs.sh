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
# THIS IS A PROPOSED VERSION with a hypothetical "get" function that makes accessing dotted fields cleaner

# Note: This is a PROPOSAL and won't work until the "get" function is added to the formatter

cat test/grpc-logs.json | dist/logista \
  --preferred_date_format='15:04:05.000' \
  --fmt='{{.ts | date | color "cyan"}} {{.level | pad 7 | colorByLevel .level}} [{{.logger}}] {{.msg | colorByLevel .level | bold}}
{{if get "grpc.service" .}}  {{get "grpc.component" . | color "blue" | pad 10}} {{get "grpc.service" .}}.{{get "grpc.method" .}} ({{get "grpc.method_type" . | color "magenta"}}){{if get "grpc.time_ms" .}} took {{get "grpc.time_ms" .}}ms{{end}}{{if get "grpc.code" .}} → {{get "grpc.code" . | colorByLevel .level}}{{end}}
{{end}}{{if or (get "authz.action" .) (get "authz.roles" .)}}  {{color "yellow" "AUTH:     "}} {{if get "authz.action" .}}action={{get "authz.action" .}}{{end}} {{if get "authz.roles" .}}roles={{get "authz.roles" .}}{{end}}
{{end}}{{range $key, $value := .}}{{if and 
  (not (isStandardField $key))
  (not (hasPrefix $key "grpc."))
  (not (hasPrefix $key "authz.")) 
}}  {{$key | dim | pad 23}}: {{$value}}
{{end}}{{end}}'
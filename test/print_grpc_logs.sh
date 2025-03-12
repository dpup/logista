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
{{if @grpc.service}}  {{"grpc" | color "blue"}} {{@grpc.component | color "blue"}} {{@grpc.service}}.{{@grpc.method}} ({{@grpc.method_type | color "magenta"}}){{if @grpc.time_ms}} {{@grpc.time_ms}}ms{{end}}{{if @grpc.code}} → {{@grpc.code | colorByLevel .level}}{{end}}
{{end}}{{if or (@authz.action) (@authz.roles)}}  {{color "yellow" "auth" | pad 26}} : {{if @authz.action}}action={{@authz.action | color "yellow"}}{{end}} {{if @authz.roles}}roles={{@authz.roles | color "yellow"}}{{end}}
{{end}}{{range $key, $value := .}}{{if and 
  (not (hasPrefix $key "grpc."))
  (not (hasPrefix $key "authz.")) 
}}  {{$key | dim | pad 26}}: {{$value | pretty}}
{{end}}{{end}}'

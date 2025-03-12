# Logista

Logista is a CLI tool for formatting JSON log streams. It's designed to be used
with server applications that output JSON logs, allowing for human-readable log
formatting without requiring the server to have separate production and
development logging formats.

## Installation

```
go install github.com/dpup/logista/cmd/logista@latest
```

## Usage

Formatting uses a super-set of Go's native templates, so any syntax that is a
valid Go template is also a valid formatting string in Logista. This means you
can use Go's powerful templating features, such as conditionals, loops, and
functions, to create highly customized log formats.

```
# Basic usage with default format
my-server | logista

# Simple syntax with custom log formats
my-server | logista --fmt="{timestamp} [{level}] {message}"
my-server | logista --fmt="{timestamp | date} [{level}] {message}"

# @symbol syntax for fields with special characters
my-server | logista --fmt="{@user.name} (ID: {@request-id})"

# Go template syntax (enables advanced features)
my-server | logista --fmt="{{.timestamp}} [{{.level}}] {{.message}}"

# Custom date format
my-server | logista --fmt="{timestamp | date} [{level}] {message}" --preferred_date_format="15:04:05"

# With message colored by log level (colors error red, warning yellow, etc)
my-server | logista --fmt="{timestamp | date} [{level}] {msg | colorByLevel .level}"

# With colored output and other formatting
my-server | logista --fmt="{{.timestamp | date | color \"cyan\"}} [{{.level}}] {{.message | colorByLevel .level | bold}}"

# Pretty-print complex objects like maps and arrays
my-server | logista --fmt="{{.timestamp | date}} [{{.level}}] {{.message}} {{.context | pretty}}"

# Disable colors
my-server | logista --fmt="{{.level | color \"red\"}} {{.message}}" --no-colors

# Help
logista --help
```

## Format Templates

Logista supports several syntax options for formatting logs:

1. **Simple Syntax**: Fields are enclosed in single curly braces. This is convenient for simple formats.

   ```
   {timestamp} [{level}] {message}
   ```

2. **@Symbol Syntax**: Fields can be accessed using the @symbol notation within Go template braces. This is especially useful for fields with special characters like periods, hyphens, or underscores.

   ```
   {{@user.name}} {{@request-id}} {{@response_code}}
   ```

3. **Full Go Template Syntax**: Fields are accessed using `.fieldname` within double curly braces. This enables powerful template features like conditionals, loops, and variable assignments.
   ```
   {{.timestamp}} [{{.level}}] {{.message}} {{.context.user.id}}
   ```

### Template Functions

Logista supports template functions that can transform field values. To use a function, add a pipe `|` after the field name, followed by the function name.

Either syntax supports functions:

```
{timestamp | date} [{level}] {message}
```

Or using full Go template syntax:

```
{{.timestamp | date}} [{{.level}}] {{.message}}
```

#### Available Functions

**Value Formatting Functions**:

| Command    | Description                                                                                                                                                                                                                                                                                                                                                                                        | Example               |
| ---------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------- |
| **date**   | Parses dates in various formats into a standardized format. Works with: ISO 8601 timestamps (`2024-03-10T15:04:05Z`), Unix timestamps (`1741626507`) (seconds since epoch), Unix timestamps with fractional seconds (`1741626507.9066188`), Common log formats (`10/Mar/2024:15:04:05 +0000`), and many others. Use `--preferred_date_format` to set the output format in Go's time format syntax. | `{timestamp \| date}` |
| **pad**    | Pads a string to a specified length.                                                                                                                                                                                                                                                                                                                                                               | `{level \| pad 10}`   |
| **pretty** | Pretty-prints any value with nicely readable formatting.                                                                                                                                                                                                                                                                                                                                           | `{context \| pretty}` |

**Color Functions**:

| Command          | Description                             | Example                           |
| ---------------- | --------------------------------------- | --------------------------------- |
| **color**        | Apply a specific color to a value       | `{level \| color "red"}`          |
| **colorByLevel** | Colors a value based on the level value | `{message \| colorByLevel level}` |
| **bold**         | Makes text bold                         | `{message \| bold}`               |
| **italic**       | Makes text italic                       | `{message \| italic}`             |
| **underline**    | Underlines text                         | `{message \| underline}`          |
| **dim**          | Makes text dim                          | `{timestamp \| dim}`              |

**Field Filtering Functions**:

| Command              | Description                                                | Example                                                            |
| -------------------- | ---------------------------------------------------------- | ------------------------------------------------------------------ |
| **hasPrefix**        | Checks if a string has a specific prefix                   | `{{if hasPrefix $key "grpc."}}`                                    |
| **getFields**        | Returns all fields in the data map                         | `{{range $key, $value := getFields .}}`                            |
| **getFieldsWithout** | Returns fields that don't match any of the provided fields | `{{range $key, $value := getFieldsWithout . "level" "timestamp"}}` |

### Advanced Template Features

When using the full Go template syntax, you get access to powerful template features like conditionals, loops, and variable assignments:

```
{{.ts | date | color "cyan"}} {{.level | colorByLevel .level}} {{.msg | bold}}
{{if hasPrefix "grpc.service" "grpc."}}
  GRPC: {{.grpc.service}}.{{.grpc.method}} ({{.grpc.method_type}})
{{end}}
```

You can iterate over fields and filter them:

```
{{range $key, $value := .}}
  {{if not (eq $key "level" "timestamp" "msg")}}
    {{$key}}: {{$value}}
  {{end}}
{{end}}
```

**Note**: Advanced features like conditionals and loops are only available with the full `{{.field}}` syntax, not the simplified `{field}` syntax.

### Structured Log Example

Here's a comprehensive example that clearly formats structured logs:

```go
{{.ts | date | color "cyan"}} {{.level | colorByLevel .level}} {{.msg | bold}} ({{.logger | dim}})
{{if hasPrefix "grpc.service" "grpc."}}
  GRPC: {{.grpc.service}}.{{.grpc.method}} ({{.grpc.method_type | color "yellow"}})
{{end}}
{{range $key, $value := .}}
  {{if not (eq $key "level" "ts" "msg" "logger" "caller")}}
    {{if not (hasPrefix $key "grpc.")}}
  {{$key | dim}}: {{$value}}
    {{end}}
  {{end}}
{{end}}
```

### Available Colors

The following colors are available for use with color functions:

- Foreground colors: `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`, `gray`
- Bright colors: `brightred`, `brightgreen`, `brightyellow`, `brightblue`, `brightmagenta`, `brightcyan`, `brightwhite`
- Background colors: `bg-black`, `bg-red`, `bg-green`, `bg-yellow`, `bg-blue`, `bg-magenta`, `bg-cyan`, `bg-white`, `bg-gray`
- Bright backgrounds: `bg-brightred`, `bg-brightgreen`, `bg-brightyellow`, `bg-brightblue`, `bg-brightmagenta`, `bg-brightcyan`, `bg-brightwhite`
- Formatting: `bold`, `italic`, `underline`, `dim`

Colors can be disabled with the `--no-colors` flag.

## Building from Source

```
git clone https://github.com/dpup/logista.git
cd logista
make build
```

The binary will be created in the `dist` directory.

## License

MIT

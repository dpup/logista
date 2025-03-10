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

```
# Basic usage with default format
my-server | logista

# Custom format
my-server | logista --fmt="{timestamp} [{level}] {message}"

# Format with template functions
my-server | logista --fmt="{timestamp | date} [{level}] {message}"

# Custom date format
my-server | logista --fmt="{timestamp | date} [{level}] {message}" --preferred_date_format="15:04:05"

# With colored output
my-server | logista --fmt="<cyan>{timestamp | date}</> <bold {level | levelColor}>[{level}]</> {message}"

# Disable colors
my-server | logista --fmt="<red>{level}</> {message}" --no-colors

# Help
logista --help
```

## Format Templates

Format templates use a simple syntax where field names are enclosed in curly
braces `{}`. These field names should match the keys in your JSON log entries.

Example:

```
{timestamp} [{level}] {message} {context.user.id}
```

### Template Functions

Logista supports template functions that can transform field values. To use a function, add a pipe `|` after the field name, followed by the function name:

```
{timestamp | date} [{level}] {message}
```

#### Available Functions

- **date**: Parses dates in various formats into a standardized format. Works with:

  - ISO 8601 timestamps: `2024-03-10T15:04:05Z`
  - Unix timestamps: `1741626507` (seconds since epoch)
  - Unix timestamps with fractional seconds: `1741626507.9066188`
  - Common log formats: `10/Mar/2024:15:04:05 +0000`
  - Many other common formats

  Use `--preferred_date_format` to set the output format in Go's time format syntax.

- **levelColor**: Automatically applies appropriate colors based on log level:

  - `error`, `fatal`, `critical`, etc. → red
  - `warn`, `warning` → yellow
  - `info` → green
  - `debug` → cyan
  - `trace` → blue

  Example use `<bold {level | levelColor}>{message}</>`

### Color Tags

You can add color to your output using HTML-like color tags:

```
<red>This text will be red</>
```

Available color tags:

- Foreground colors: `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`, `gray`
- Bright colors: `brightred`, `brightgreen`, `brightyellow`, `brightblue`, `brightmagenta`, `brightcyan`, `brightwhite`
- Background colors: `bg-black`, `bg-red`, `bg-green`, `bg-yellow`, `bg-blue`, `bg-magenta`, `bg-cyan`, `bg-white`, `bg-gray`
- Bright backgrounds: `bg-brightred`, `bg-brightgreen`, `bg-brightyellow`, `bg-brightblue`, `bg-brightmagenta`, `bg-brightcyan`, `bg-brightwhite`
- Formatting: `bold`, `italic`, `underline`, `dim`

You can combine multiple styles by adding spaces:

```
<bold red>Bold red text</>
```

For simplicity, you can use `</>` as a universal closing tag, or use traditional HTML-style tags like `</red>` if you prefer.

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

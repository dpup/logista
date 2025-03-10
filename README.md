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

## Building from Source

```
git clone https://github.com/dpup/logista.git
cd logista
go build -o logista ./cmd/logista
```

## License

MIT

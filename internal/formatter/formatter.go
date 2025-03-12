package formatter

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// Formatter is an interface for formatting JSON log entries
type Formatter interface {
	// Format converts a log entry map to a formatted string
	Format(data map[string]interface{}) (string, error)
}

// TemplateFormatter formats logs using a Go template
type TemplateFormatter struct {
	template         *template.Template
	preferredDateFmt string
	noColors         bool
	tableKeyPadding  int
}

// FormatterOption is a functional option for configuring the formatter
type FormatterOption func(*TemplateFormatter)

// WithPreferredDateFormat sets the preferred date format for the date function
func WithPreferredDateFormat(format string) FormatterOption {
	return func(tf *TemplateFormatter) {
		tf.preferredDateFmt = format
	}
}

// WithNoColors disables color output
func WithNoColors(noColors bool) FormatterOption {
	return func(tf *TemplateFormatter) {
		tf.noColors = noColors
	}
}

// No longer needed as the filter function can be used directly in templates

// WithTableKeyPadding sets the padding length for keys in table output
func WithTableKeyPadding(padding int) FormatterOption {
	return func(tf *TemplateFormatter) {
		tf.tableKeyPadding = padding
	}
}

// NewTemplateFormatter creates a new TemplateFormatter with the given format string
func NewTemplateFormatter(format string, opts ...FormatterOption) (*TemplateFormatter, error) {
	// Process template with shortcuts via the preprocessor
	format = PreProcessTemplate(format, DefaultPreProcessTemplateOptions())

	// Create the formatter with default values
	formatter := &TemplateFormatter{
		preferredDateFmt: "2006-01-02 15:04:05",
		tableKeyPadding:  19,
	}

	// Apply options
	for _, opt := range opts {
		opt(formatter)
	}

	// Create template with custom functions
	tmpl := template.New("formatter").Funcs(template.FuncMap{
		// Value formatting
		"date":     formatter.dateFunc,
		"pad":      formatter.padFunc,
		"pretty":   formatter.prettyFunc,
		"table":    formatter.tableFunc,
		"duration": formatter.durationFunc,
		"wrap":     formatter.wrapFunc,
		"trunc":    formatter.truncFunc,

		// Color functions
		"color":        formatter.colorFunc,
		"colorByLevel": formatter.colorByLevelFunc,
		"bold":         formatter.boldFunc,
		"italic":       formatter.italicFunc,
		"underline":    formatter.underlineFunc,
		"dim":          formatter.dimFunc,

		// Field filtering and categorization
		"hasPrefix": formatter.hasPrefixFunc,
		"filter":    formatter.filterFunc,
	})

	parsed, err := tmpl.Parse(format)
	if err != nil {
		return nil, err
	}

	formatter.template = parsed
	return formatter, nil
}

// padFunc is a template function that pads a string to a specified length
func (f *TemplateFormatter) padFunc(length int, value interface{}) string {
	if value == nil {
		return strings.Repeat(" ", length)
	}

	str := fmt.Sprintf("%v", value)
	if len(str) >= length {
		return str
	}

	// Add whitespace padding to the right of the string
	return str + strings.Repeat(" ", length-len(str))
}

// dateFunc is a template function that parses various date formats and outputs a standard format
func (f *TemplateFormatter) dateFunc(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		// Try parsing common formats
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05.999999999",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02",
			"Mon Jan 2 15:04:05 2006",
			"Mon Jan 2 15:04:05 MST 2006",
			"Jan 2 15:04:05",
			"Jan 2 15:04:05 2006",
			"02/Jan/2006:15:04:05 -0700", // Common log format
		}

		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t.Format(f.preferredDateFmt)
			}
		}
		return v
	case json.Number:
		// Try parsing as Unix timestamp
		if i, err := v.Int64(); err == nil {
			return time.Unix(i, 0).Format(f.preferredDateFmt)
		}
		// Try parsing as Unix timestamp with fractional seconds
		if floatVal, err := v.Float64(); err == nil {
			sec := int64(floatVal)
			nsec := int64((floatVal - float64(sec)) * 1e9)
			return time.Unix(sec, nsec).Format(f.preferredDateFmt)
		}
		return v.String()
	case int64:
		return time.Unix(v, 0).Format(f.preferredDateFmt)
	case float64:
		sec := int64(v)
		nsec := int64((v - float64(sec)) * 1e9)
		return time.Unix(sec, nsec).Format(f.preferredDateFmt)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// colorFunc applies a specific color to a value
func (f *TemplateFormatter) colorFunc(colorName string, value interface{}) string {
	if f.noColors || value == nil {
		return fmt.Sprintf("%v", value)
	}

	content := fmt.Sprintf("%v", value)
	return ApplyColorToString(content, colorName)
}

// colorByLevelFunc applies color to a value based on the level
// In Go templates with pipes, arguments are passed in reverse order
// so {{.msg | colorByLevel .level}} passes level as first arg and msg as second arg
func (f *TemplateFormatter) colorByLevelFunc(level interface{}, value interface{}) string {
	if f.noColors || value == nil {
		return fmt.Sprintf("%v", value)
	}

	if level == nil {
		return fmt.Sprintf("%v", value)
	}

	content := fmt.Sprintf("%v", value)
	levelStr := fmt.Sprintf("%v", level)

	colorName := ColorByLevelName(levelStr)
	if code, ok := colorCodes[colorName]; ok {
		return fmt.Sprintf("\033[%sm%s%s", code, content, ansiReset)
	}

	return content
}

// boldFunc makes text bold
func (f *TemplateFormatter) boldFunc(value interface{}) string {
	if f.noColors || value == nil {
		return fmt.Sprintf("%v", value)
	}

	content := fmt.Sprintf("%v", value)
	return fmt.Sprintf("\033[1m%s%s", content, ansiReset)
}

// italicFunc makes text italic
func (f *TemplateFormatter) italicFunc(value interface{}) string {
	if f.noColors || value == nil {
		return fmt.Sprintf("%v", value)
	}

	content := fmt.Sprintf("%v", value)
	return fmt.Sprintf("\033[3m%s%s", content, ansiReset)
}

// underlineFunc underlines text
func (f *TemplateFormatter) underlineFunc(value interface{}) string {
	if f.noColors || value == nil {
		return fmt.Sprintf("%v", value)
	}

	content := fmt.Sprintf("%v", value)
	return fmt.Sprintf("\033[4m%s%s", content, ansiReset)
}

// dimFunc makes text dim
func (f *TemplateFormatter) dimFunc(value interface{}) string {
	if f.noColors || value == nil {
		return fmt.Sprintf("%v", value)
	}

	content := fmt.Sprintf("%v", value)
	return fmt.Sprintf("\033[2m%s%s", content, ansiReset)
}

// prettyFunc is a template function that pretty-prints any value, with special handling for maps and arrays
func (f *TemplateFormatter) prettyFunc(value interface{}) string {
	if value == nil {
		return "<nil>"
	}

	// Handle basic types directly
	switch v := value.(type) {
	case string:
		if v == "" {
			return "<empty>"
		}
		return v
	case bool:
		return fmt.Sprintf("%t", v)
	case json.Number:
		return v.String()
	case time.Duration:
		return formatDuration(v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprintf("%v", v)
	case []interface{}:
		return f.prettyArray(v)
	case map[string]interface{}:
		return f.prettyMap(v)
	}

	// For other complex types, use reflection to determine the kind
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		// Convert to []interface{} and use prettyArray
		length := val.Len()
		arr := make([]interface{}, length)
		for i := 0; i < length; i++ {
			arr[i] = val.Index(i).Interface()
		}
		return f.prettyArray(arr)
	} else if val.Kind() == reflect.Map {
		// Convert to map[string]interface{} and use prettyMap if possible
		result := make(map[string]interface{})
		for _, key := range val.MapKeys() {
			// Try to convert key to string
			var keyStr string
			if key.Kind() == reflect.String {
				keyStr = key.String()
			} else {
				keyStr = fmt.Sprintf("%v", key.Interface())
			}
			result[keyStr] = val.MapIndex(key).Interface()
		}
		return f.prettyMap(result)
	}

	// Fallback to standard formatting for other types
	return fmt.Sprintf("%v", value)
}

// prettyArray formats an array as a comma-separated list with dim formatting for commas
func (f *TemplateFormatter) prettyArray(arr []interface{}) string {
	if len(arr) == 0 {
		return "[]"
	}

	var builder strings.Builder
	builder.WriteString("[")

	for i, item := range arr {
		builder.WriteString(f.prettyFunc(item))
		if i < len(arr)-1 {
			if f.noColors {
				builder.WriteString(", ")
			} else {
				builder.WriteString(fmt.Sprintf("\033[2m, \033[0m"))
			}
		}
	}

	builder.WriteString("]")
	return builder.String()
}

// prettyMap formats a map with key=value format where the key part is dim
func (f *TemplateFormatter) prettyMap(m map[string]interface{}) string {
	if len(m) == 0 {
		return "{}"
	}

	var builder strings.Builder
	builder.WriteString("{")

	i := 0
	for key, val := range m {
		if i > 0 {
			if f.noColors {
				builder.WriteString(", ")
			} else {
				builder.WriteString(fmt.Sprintf("\033[2m, \033[0m"))
			}
		}

		// Key part with dim formatting if colors are enabled
		if f.noColors {
			builder.WriteString(fmt.Sprintf("%s=", key))
		} else {
			builder.WriteString(fmt.Sprintf("\033[2m%s=\033[0m", key))
		}

		// Value part with normal formatting
		builder.WriteString(f.prettyFunc(val))
		i++
	}

	builder.WriteString("}")
	return builder.String()
}

// tableFunc formats a map as a table with each field on a new line
// Format is "key: value" with keys right-padded and dimmed
// Empty or nil values are omitted (use with filter function for field exclusion)
func (f *TemplateFormatter) tableFunc(value interface{}) string {
	if value == nil {
		return ""
	}

	// Convert to map if possible
	var dataMap map[string]interface{}
	switch v := value.(type) {
	case map[string]interface{}:
		dataMap = v
	default:
		// For non-map types, return as is using pretty formatting
		return f.prettyFunc(value)
	}

	if len(dataMap) == 0 {
		return ""
	}

	// Get a sorted list of keys for consistent output
	var keys []string
	for key := range dataMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Build the table
	var builder strings.Builder
	for i, key := range keys {
		val := dataMap[key]

		// Skip empty values (nil or empty string)
		isEmpty := val == nil
		if !isEmpty {
			if str, ok := val.(string); ok && str == "" {
				isEmpty = true
			}
		}
		if isEmpty {
			continue
		}

		// Add newline between fields
		if i > 0 {
			builder.WriteString("\n")
		}

		// Format the key with padding and dim effect
		paddedKey := f.padFunc(f.tableKeyPadding, key)
		if f.noColors {
			builder.WriteString(fmt.Sprintf("  %s", paddedKey))
		} else {
			builder.WriteString(fmt.Sprintf("  \033[2m%s\033[0m", paddedKey))
		}

		// Format the value using pretty
		builder.WriteString(f.prettyFunc(val))
	}

	return builder.String()
}

// hasPrefixFunc checks if a string has a specific prefix
func (f *TemplateFormatter) hasPrefixFunc(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// filterFunc returns a filtered map of fields based on patterns
// It can handle exact field names or prefix patterns with wildcards
// Example: filter . "timestamp" "level" - excludes timestamp and level fields
// Example: filter . "grpc.*" - excludes all fields starting with "grpc."
func (f *TemplateFormatter) filterFunc(data map[string]interface{}, excludePatterns ...string) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		exclude := false
		for _, pattern := range excludePatterns {
			// Check if pattern ends with wildcard
			if strings.HasSuffix(pattern, "*") {
				prefix := pattern[:len(pattern)-1]
				if strings.HasPrefix(key, prefix) {
					exclude = true
					break
				}
			} else if key == pattern {
				// Exact field match
				exclude = true
				break
			}
		}

		if !exclude {
			result[key] = value
		}
	}

	return result
}

// formatDuration formats a time.Duration into a human-readable string
// For example: 1h30m45s, 250ms, 1.5s
func formatDuration(d time.Duration) string {
	// Handle special cases for very small durations
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	}

	// For durations less than 1ms, show as microseconds
	if d < time.Millisecond {
		microSeconds := float64(d.Nanoseconds()) / float64(time.Microsecond)
		return fmt.Sprintf("%.2fµs", microSeconds)
	}

	if d < time.Second {
		milliSeconds := float64(d.Nanoseconds()) / float64(time.Millisecond)
		return fmt.Sprintf("%.2fms", milliSeconds)
	}

	// For durations around a few seconds, show decimal seconds
	if d < 10*time.Second {
		seconds := float64(d.Nanoseconds()) / float64(time.Second)
		return fmt.Sprintf("%.2fs", seconds)
	}

	// For longer durations, use the standard Go format (which automatically
	// selects appropriate units like 1h30m45s)
	return d.String()
}

// parseDuration attempts to parse a value as a duration
// It can handle:
// - time.Duration values directly
// - String values (parseable by time.ParseDuration like "1h30m", "500ms")
// - Numeric values (assumed to be milliseconds)
// - json.Number values (assumed to be milliseconds)
func parseDuration(value interface{}) (time.Duration, error) {
	if value == nil {
		return 0, fmt.Errorf("cannot parse nil as duration")
	}

	switch v := value.(type) {
	case time.Duration:
		return v, nil
	case string:
		// Try to parse as a Go duration string (e.g., "1h30m", "500ms")
		if d, err := time.ParseDuration(v); err == nil {
			return d, nil
		}
		// Failed to parse directly, return error
		return 0, fmt.Errorf("cannot parse '%s' as duration", v)
	case json.Number:
		// Parse as milliseconds
		if f, err := v.Float64(); err == nil {
			return time.Duration(f * float64(time.Millisecond)), nil
		}
		return 0, fmt.Errorf("cannot parse '%s' as milliseconds", v)
	case int:
		return time.Duration(v) * time.Millisecond, nil
	case int64:
		return time.Duration(v) * time.Millisecond, nil
	case float64:
		return time.Duration(v * float64(time.Millisecond)), nil
	default:
		return 0, fmt.Errorf("cannot parse '%v' (type %T) as duration", v, v)
	}
}

// durationFunc is a template function that parses a value as duration and formats it nicely
func (f *TemplateFormatter) durationFunc(value interface{}) string {
	duration, err := parseDuration(value)
	if err != nil {
		// If we can't parse as duration, just use pretty formatting
		return f.prettyFunc(value)
	}
	return formatDuration(duration)
}

// truncFunc is a template function that truncates text to a specified length
// and adds an ellipsis if the text was truncated.
// Usage: {{.message | trunc 20}}
func (f *TemplateFormatter) truncFunc(maxLen interface{}, value interface{}) string {
	// Handle nil case
	if value == nil {
		return "<no value>"
	}

	// Get the text to truncate
	text := fmt.Sprintf("%v", value)
	if text == "" {
		return ""
	}

	// Parse maxLen parameter
	maxLength := 20 // Default max length
	if maxLen != nil {
		if l, ok := maxLen.(int); ok {
			maxLength = l
		} else if l, err := strconv.Atoi(fmt.Sprintf("%v", maxLen)); err == nil {
			maxLength = l
		}
	}
	if maxLength <= 0 {
		maxLength = 20
	}

	// If the string is already shorter than max length, return it as is
	if len(text) <= maxLength {
		return text
	}

	// Truncate the string and add ellipsis
	// If maxLength is very small (less than 4), we might not have space for ellipsis
	if maxLength < 4 {
		return text[:maxLength]
	}

	return text[:maxLength-3] + "..."
}

// wrapFunc is a template function that wraps text to a specified width
// It takes a width parameter (required) and an optional indent parameter
// for wrapped lines. Usage: {{.description | wrap 80 2}}
func (f *TemplateFormatter) wrapFunc(width interface{}, indent interface{}, value interface{}) string {
	// Handle nil case
	if value == nil {
		return "<no value>"
	}

	// Get the text to wrap
	text := fmt.Sprintf("%v", value)
	if text == "" {
		return ""
	}

	// Parse width parameter
	widthVal := 80 // Default width
	if width != nil {
		if w, ok := width.(int); ok {
			widthVal = w
		} else if w, err := strconv.Atoi(fmt.Sprintf("%v", width)); err == nil {
			widthVal = w
		}
	}
	if widthVal <= 0 {
		widthVal = 80
	}

	// Parse indent parameter
	indentVal := 0
	if indent != nil {
		if i, ok := indent.(int); ok {
			indentVal = i
		} else if i, err := strconv.Atoi(fmt.Sprintf("%v", indent)); err == nil {
			indentVal = i
		}
	}
	if indentVal < 0 {
		indentVal = 0
	}

	indentStr := strings.Repeat(" ", indentVal)
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	// Calculate line width for wrapped lines
	firstLineWidth := widthVal
	wrappedLineWidth := widthVal - indentVal

	var result strings.Builder
	lineLength := 0
	isFirstLine := true

	for _, word := range words {
		wordLen := len(word)
		currentWidth := firstLineWidth
		if !isFirstLine {
			currentWidth = wrappedLineWidth
		}

		// Check if adding this word would exceed the width
		spaceNeeded := 0
		if lineLength > 0 {
			spaceNeeded = 1 // Need a space between words
		}

		if lineLength+wordLen+spaceNeeded > currentWidth {
			// Start a new line
			result.WriteString("\n")

			// Add indent if not the first line
			if indentVal > 0 {
				result.WriteString(indentStr)
			}

			// Add the word
			result.WriteString(word)
			lineLength = wordLen

			// Mark that we're no longer on the first line
			isFirstLine = false
		} else {
			// Add a space before the word if it's not the first word on the line
			if lineLength > 0 {
				result.WriteString(" ")
				lineLength++
			}

			result.WriteString(word)
			lineLength += wordLen
		}
	}

	return result.String()
}

// Format formats the data according to the template
func (f *TemplateFormatter) Format(data map[string]interface{}) (string, error) {
	var buf strings.Builder
	if err := f.template.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ProcessStream processes JSON logs from a reader and writes formatted output to a writer
func (f *TemplateFormatter) ProcessStream(r io.Reader, w io.Writer, formatter Formatter) error {
	decoder := json.NewDecoder(r)
	decoder.UseNumber() // Use json.Number for numeric values to preserve precision

	for {
		var data map[string]interface{}
		if err := decoder.Decode(&data); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		formatted, err := formatter.Format(data)
		if err != nil {
			return err
		}

		if _, err := io.WriteString(w, formatted+"\n"); err != nil {
			return err
		}
	}

	return nil
}

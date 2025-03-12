package formatter

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
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
	template             *template.Template
	preferredDateFmt     string
	noColors             bool
	tableExcludePrefixes []string
	tableKeyPadding      int
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

// WithTableExcludePrefixes sets prefixes for fields to exclude from table output
func WithTableExcludePrefixes(prefixes []string) FormatterOption {
	return func(tf *TemplateFormatter) {
		tf.tableExcludePrefixes = prefixes
	}
}

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
		preferredDateFmt:     "2006-01-02 15:04:05",
		tableExcludePrefixes: []string{"grpc."},
		tableKeyPadding:      26,
	}

	// Apply options
	for _, opt := range opts {
		opt(formatter)
	}

	// Create template with custom functions
	tmpl := template.New("formatter").Funcs(template.FuncMap{
		// Value formatting
		"date":   formatter.dateFunc,
		"pad":    formatter.padFunc,
		"pretty": formatter.prettyFunc,
		"table":  formatter.tableFunc,

		// Color functions
		"color":        formatter.colorFunc,
		"colorByLevel": formatter.colorByLevelFunc,
		"bold":         formatter.boldFunc,
		"italic":       formatter.italicFunc,
		"underline":    formatter.underlineFunc,
		"dim":          formatter.dimFunc,

		// Field filtering and categorization
		"hasPrefix":        formatter.hasPrefixFunc,
		"getFields":        formatter.getFieldsFunc,
		"getFieldsWithout": formatter.getFieldsWithoutFunc,
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
// Empty or nil values are omitted, and fields with specified prefixes are excluded
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
		// Skip excluded prefixes
		excluded := false
		for _, prefix := range f.tableExcludePrefixes {
			if strings.HasPrefix(key, prefix) {
				excluded = true
				break
			}
		}
		if !excluded {
			keys = append(keys, key)
		}
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
			builder.WriteString(fmt.Sprintf("  %s: ", paddedKey))
		} else {
			builder.WriteString(fmt.Sprintf("  \033[2m%s\033[0m: ", paddedKey))
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

// getFieldsFunc returns all fields in the data map
func (f *TemplateFormatter) getFieldsFunc(data map[string]interface{}) map[string]interface{} {
	return data
}

// getFieldsWithoutFunc returns fields that don't match any of the provided fields
func (f *TemplateFormatter) getFieldsWithoutFunc(data map[string]interface{}, excludeFields ...string) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		exclude := false
		for _, excludeKey := range excludeFields {
			if key == excludeKey {
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

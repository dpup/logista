package formatter

import (
	"encoding/json"
	"fmt"
	"io"
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

// NewTemplateFormatter creates a new TemplateFormatter with the given format string
func NewTemplateFormatter(format string, opts ...FormatterOption) (*TemplateFormatter, error) {
	// Process template with shortcuts via the preprocessor
	format = PreProcessTemplate(format, DefaultPreProcessTemplateOptions())

	// Create the formatter with default values
	formatter := &TemplateFormatter{
		preferredDateFmt: "2006-01-02 15:04:05",
	}

	// Apply options
	for _, opt := range opts {
		opt(formatter)
	}

	// Create template with custom functions
	tmpl := template.New("formatter").Funcs(template.FuncMap{
		// Value formatting
		"date": formatter.dateFunc,
		"pad":  formatter.padFunc,

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
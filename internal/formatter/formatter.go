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
}

// FormatterOption is a functional option for configuring the formatter
type FormatterOption func(*TemplateFormatter)

// WithPreferredDateFormat sets the preferred date format for the date function
func WithPreferredDateFormat(format string) FormatterOption {
	return func(tf *TemplateFormatter) {
		tf.preferredDateFmt = format
	}
}

// NewTemplateFormatter creates a new TemplateFormatter with the given format string
func NewTemplateFormatter(format string, opts ...FormatterOption) (*TemplateFormatter, error) {
	// Replace {field} with {{.field}} for Go template
	goTmplFormat := strings.ReplaceAll(format, "{", "{{.")
	goTmplFormat = strings.ReplaceAll(goTmplFormat, "}", "}}")

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
		"date": formatter.dateFunc,
	})

	parsed, err := tmpl.Parse(goTmplFormat)
	if err != nil {
		return nil, err
	}

	formatter.template = parsed
	return formatter, nil
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
		return v // Return original if we couldn't parse it
	case float64:
		// Try parsing as Unix timestamp (possibly with milliseconds or microseconds)
		sec := int64(v)
		nsec := int64((v - float64(sec)) * 1e9)
		t := time.Unix(sec, nsec)
		return t.Format(f.preferredDateFmt)
	case int:
		// Unix timestamp (seconds only)
		t := time.Unix(int64(v), 0)
		return t.Format(f.preferredDateFmt)
	case int64:
		// Unix timestamp (seconds only)
		t := time.Unix(v, 0)
		return t.Format(f.preferredDateFmt)
	case json.Number:
		// Handle JSON numbers
		if i, err := v.Int64(); err == nil {
			t := time.Unix(i, 0)
			return t.Format(f.preferredDateFmt)
		}
		if floatVal, err := v.Float64(); err == nil {
			sec := int64(floatVal)
			nsec := int64((floatVal - float64(sec)) * 1e9)
			t := time.Unix(sec, nsec)
			return t.Format(f.preferredDateFmt)
		}
		return v.String()
	default:
		// Try to convert to string
		return fmt.Sprintf("%v", v)
	}
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

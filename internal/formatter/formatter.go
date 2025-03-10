package formatter

import (
	"encoding/json"
	"io"
	"strings"
	"text/template"
)

// Formatter is an interface for formatting JSON log entries
type Formatter interface {
	// Format converts a log entry map to a formatted string
	Format(data map[string]interface{}) (string, error)
}

// TemplateFormatter formats logs using a Go template
type TemplateFormatter struct {
	template *template.Template
}

// NewTemplateFormatter creates a new TemplateFormatter with the given format string
func NewTemplateFormatter(format string) (*TemplateFormatter, error) {
	// Replace {field} with {{.field}} for Go template
	goTmplFormat := strings.ReplaceAll(format, "{", "{{.")
	goTmplFormat = strings.ReplaceAll(goTmplFormat, "}", "}}")

	tmpl, err := template.New("formatter").Parse(goTmplFormat)
	if err != nil {
		return nil, err
	}

	return &TemplateFormatter{template: tmpl}, nil
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

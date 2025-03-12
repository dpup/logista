package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"
)

func TestTemplateFormatter(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "simple format",
			format:   "{{.level}} {{.message}}",
			data:     map[string]interface{}{"level": "info", "message": "test message"},
			expected: "info test message",
		},
		{
			name:     "nested fields",
			format:   "{{.level}} {{.context.user.id}}",
			data:     map[string]interface{}{"level": "info", "context": map[string]interface{}{"user": map[string]interface{}{"id": "123"}}},
			expected: "info 123",
		},
		{
			name:     "nested fields with unknown key",
			format:   "{{.level}} {{.context.org.name}}",
			data:     map[string]interface{}{"level": "info", "context": map[string]interface{}{"user": map[string]interface{}{"id": "123"}}},
			expected: "info <no value>",
		},
		{
			name:     "missing field",
			format:   "{{.level}} {{.missing}}",
			data:     map[string]interface{}{"level": "info"},
			expected: "info <no value>",
		},
		{
			name:     "pad function",
			format:   `{{.level | pad 10}} {{.message}}`,
			data:     map[string]interface{}{"level": "info", "message": "test message"},
			expected: "info       test message",
		},
		{
			name:     "pad function with longer text",
			format:   `{{.level | pad 3}} {{.message}}`,
			data:     map[string]interface{}{"level": "warning", "message": "test message"},
			expected: "warning test message",
		},
		{
			name:     "pad function with nil value",
			format:   `{{.missing | pad 5}} {{.message}}`,
			data:     map[string]interface{}{"message": "test message"},
			expected: "      test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter, err := NewTemplateFormatter(tt.format)
			if err != nil {
				t.Fatalf("Failed to create formatter: %v", err)
			}

			result, err := formatter.Format(tt.data)
			if err != nil {
				t.Fatalf("Format failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestDateFunction(t *testing.T) {
	// Use local timezone for testing since our 'date' function
	// doesn't specify timezone in output by default
	loc, _ := time.LoadLocation("Local")
	now := time.Date(2024, 3, 10, 15, 4, 5, 0, loc)
	unixTimestamp := float64(now.Unix())
	isoDate := now.Format(time.RFC3339)

	tests := []struct {
		name         string
		format       string
		data         map[string]interface{}
		expected     string
		dateFormat   string
		expectPrefix string
	}{
		{
			name:       "date function with ISO string",
			format:     "{{.timestamp | date}}",
			data:       map[string]interface{}{"timestamp": isoDate},
			expected:   "2024-03-10 15:04:05",
			dateFormat: "2006-01-02 15:04:05",
		},
		{
			name:       "date function with unix timestamp",
			format:     "{{.timestamp | date}}",
			data:       map[string]interface{}{"timestamp": json.Number(strconv.FormatInt(now.Unix(), 10))},
			expected:   "2024-03-10 15:04:05",
			dateFormat: "2006-01-02 15:04:05",
		},
		{
			name:       "date function with float unix timestamp",
			format:     "{{.timestamp | date}}",
			data:       map[string]interface{}{"timestamp": unixTimestamp},
			expected:   "2024-03-10 15:04:05",
			dateFormat: "2006-01-02 15:04:05",
		},
		{
			name:       "date function with custom format",
			format:     "{{.timestamp | date}}",
			data:       map[string]interface{}{"timestamp": isoDate},
			expected:   "10/03/2024",
			dateFormat: "02/01/2006",
		},
		{
			name:       "date function with common log format",
			format:     "{{.timestamp | date}}",
			data:       map[string]interface{}{"timestamp": "10/Mar/2024:15:04:05 +0000"},
			expected:   "2024-03-10 15:04:05",
			dateFormat: "2006-01-02 15:04:05",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts []FormatterOption
			if tt.dateFormat != "" {
				opts = append(opts, WithPreferredDateFormat(tt.dateFormat))
			}

			formatter, err := NewTemplateFormatter(tt.format, opts...)
			if err != nil {
				t.Fatalf("Failed to create formatter: %v", err)
			}

			result, err := formatter.Format(tt.data)
			if err != nil {
				t.Fatalf("Format failed: %v", err)
			}

			if tt.expectPrefix != "" {
				if !strings.HasPrefix(result, tt.expectPrefix) {
					t.Errorf("Expected result to start with '%s', got '%s'", tt.expectPrefix, result)
				}
			} else if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestStandardTemplateSyntax(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "standard Go template syntax",
			format:   "{{.level}} {{.message}}",
			data:     map[string]interface{}{"level": "info", "message": "test message"},
			expected: "info test message",
		},
		{
			name:     "Go template with function",
			format:   "{{.level | pad 10}} {{.message}}",
			data:     map[string]interface{}{"level": "info", "message": "test message"},
			expected: "info       test message",
		},
		{
			name:     "Go template with color function",
			format:   "{{.level}} {{.message | color \"red\"}}",
			data:     map[string]interface{}{"level": "info", "message": "test message"},
			expected: "info \033[31mtest message\033[0m",
		},
		{
			name:     "Go template with colorByLevel function",
			format:   "{{.level}} {{.message | colorByLevel .level}}",
			data:     map[string]interface{}{"level": "error", "message": "test message"},
			expected: "error \033[31mtest message\033[0m",
		},
		{
			name:     "Go template with index function",
			format:   "{{.level}} {{index . \"grpc.method\"}}",
			data:     map[string]interface{}{"level": "info", "grpc.method": "GetUser"},
			expected: "info GetUser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use options to disable preprocessing for this test
			options := []FormatterOption{
				WithNoColors(false),
			}

			// Create a new formatter with the PreProcessTemplate function manually applied
			// to ensure we're only testing the formatter, not the preprocessor
			rawFormat := tt.format
			formatter, err := NewTemplateFormatter(rawFormat, options...)
			if err != nil {
				t.Fatalf("Failed to create formatter: %v", err)
			}

			result, err := formatter.Format(tt.data)
			if err != nil {
				t.Fatalf("Format failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestProcessStream(t *testing.T) {
	input := `{"level":"info","message":"test1"}
{"level":"error","message":"test2"}`

	expected := "info test1\nerror test2\n"

	formatter, err := NewTemplateFormatter("{{.level}} {{.message}}")
	if err != nil {
		t.Fatalf("Failed to create formatter: %v", err)
	}

	r := strings.NewReader(input)
	var buf bytes.Buffer

	err = formatter.ProcessStream(r, &buf, formatter)
	if err != nil {
		t.Fatalf("ProcessStream failed: %v", err)
	}

	if buf.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, buf.String())
	}
}

func TestColorFunc(t *testing.T) {
	formatter := &TemplateFormatter{noColors: false}

	result := formatter.colorFunc("red", "test")
	expected := "\033[31mtest\033[0m"

	if result != expected {
		t.Errorf("colorFunc failed: expected %q, got %q", expected, result)
	}
}

func TestColorByLevelFunc(t *testing.T) {
	formatter := &TemplateFormatter{noColors: false}

	result := formatter.colorByLevelFunc("error", "test message")
	expected := "\033[31mtest message\033[0m"

	if result != expected {
		t.Errorf("colorByLevelFunc failed: expected %q, got %q", expected, result)
	}
}

func TestBoldFunc(t *testing.T) {
	formatter := &TemplateFormatter{noColors: false}

	result := formatter.boldFunc("test message")
	expected := "\033[1mtest message\033[0m"

	if result != expected {
		t.Errorf("boldFunc failed: expected %q, got %q", expected, result)
	}
}

func TestTemplateWithFunctions(t *testing.T) {
	tmpl, err := template.New("test").Funcs(template.FuncMap{
		"testColor": func(v string) string {
			return "\033[31m" + v + "\033[0m"
		},
	}).Parse("{{.value | testColor}}")

	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, map[string]interface{}{"value": "test"})
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	result := buf.String()
	expected := "\033[31mtest\033[0m"

	if result != expected {
		t.Errorf("Template function failed: expected %q, got %q", expected, result)
	}
}

func TestTemplateColorFunctions(t *testing.T) {
	data := map[string]interface{}{
		"level":   "error",
		"message": "Something went wrong",
	}

	f := &TemplateFormatter{noColors: false}

	// Test direct function call to verify it works as expected
	colorResult := f.colorFunc("red", "Test message")
	colorExpected := "\033[31mTest message\033[0m"

	if colorResult != colorExpected {
		t.Errorf("Direct colorFunc call failed: expected %q, got %q", colorExpected, colorResult)
	}

	// Print what arguments the template passes to our functions
	debugColor := func(a interface{}, b interface{}) string {
		t.Logf("color args: a=%v (%T), b=%v (%T)", a, a, b, b)
		return fmt.Sprintf("a=%v, b=%v", a, b)
	}

	debugTmpl, _ := template.New("debug").Funcs(template.FuncMap{
		"debugColor": debugColor,
	}).Parse(`{{.message | debugColor "red"}}`)

	var buf strings.Builder
	debugTmpl.Execute(&buf, data)

	// Now try the actual color function with Go template
	colorTmpl, _ := template.New("colorTest").Funcs(template.FuncMap{
		"mycolor": func(a interface{}, b string) string {
			// Explicit types to make sure we know what's happening
			content := fmt.Sprintf("%v", a)
			if b == "red" {
				return fmt.Sprintf("\033[31m%s\033[0m", content)
			}
			return content
		},
	}).Parse(`{{.message | mycolor "red"}}`)

	buf.Reset()
	colorTmpl.Execute(&buf, data)
	t.Logf("explicit function result: %q", buf.String())
}

func TestTemplateFieldFiltering(t *testing.T) {
	complexLog := map[string]interface{}{
		"level":            "info",
		"ts":               1741636045.070078,
		"logger":           "/students.Students/ListGroups",
		"caller":           "logging/logging.go:202",
		"msg":              "started call",
		"protocol":         "grpc",
		"grpc.component":   "server",
		"grpc.service":     "students.Students",
		"grpc.method":      "ListGroups",
		"grpc.method_type": "unary",
		"peer.address":     "127.0.0.1:60279",
		"grpc.start_time":  "2025-03-10T19:47:25Z",
		"grpc.time_ms":     "0.011",
	}

	tests := []struct {
		name     string
		format   string
		data     map[string]interface{}
		expected string
		options  []FormatterOption
	}{
		{
			name:     "hasPrefix function",
			format:   "{{.msg}} {{if hasPrefix \"grpc.method\" \"grpc.\"}}(grpc field){{end}}",
			data:     complexLog,
			expected: "started call (grpc field)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter, err := NewTemplateFormatter(tt.format, tt.options...)
			if err != nil {
				t.Fatalf("Failed to create formatter: %v", err)
			}

			result, err := formatter.Format(tt.data)
			if err != nil {
				t.Fatalf("Format failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}

func TestPrettyFunc(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		data      map[string]interface{}
		contains  []string
		notContains []string
		noColors  bool
	}{
		{
			name:   "pretty basic string",
			format: "{{.message | pretty}}",
			data:   map[string]interface{}{"message": "hello world"},
			contains: []string{"hello world"},
		},
		{
			name:   "pretty basic number",
			format: "{{.count | pretty}}",
			data:   map[string]interface{}{"count": 42},
			contains: []string{"42"},
		},
		{
			name:   "pretty map",
			format: "{{.context | pretty}}",
			data: map[string]interface{}{
				"context": map[string]interface{}{
					"user": map[string]interface{}{
						"id":   "user123",
						"name": "John Doe",
					},
					"request_id": "req-456",
				},
			},
			contains: []string{
				"{",
				"user=",
				"id=",
				"user123",
				"name=",
				"John Doe",
				"request_id=",
				"req-456",
				"}",
			},
		},
		{
			name:   "pretty map with no colors",
			format: "{{.context | pretty}}",
			data: map[string]interface{}{
				"context": map[string]interface{}{
					"user": map[string]interface{}{
						"id":   "user123",
					},
					"request_id": "req-456",
				},
			},
			contains: []string{
				"{",
				"user=",
				"id=user123",
				"request_id=req-456",
				"}",
			},
			noColors: true,
		},
		{
			name:   "pretty array",
			format: "{{.items | pretty}}",
			data: map[string]interface{}{
				"items": []interface{}{1, "two", true},
			},
			contains: []string{
				"[",
				"1",
				", ",
				"two",
				", ",
				"true",
				"]",
			},
		},
		{
			name:   "pretty nested structures",
			format: "{{.payload | pretty}}",
			data: map[string]interface{}{
				"payload": map[string]interface{}{
					"users": []interface{}{
						map[string]interface{}{
							"id":   "u1",
							"role": "admin",
						},
						map[string]interface{}{
							"id":   "u2",
							"role": "user",
						},
					},
					"metadata": map[string]interface{}{
						"version": "1.0",
					},
				},
			},
			contains: []string{
				// Map structure
				"{",
				"users=",
				"metadata=",
				
				// Array structure
				"[",
				", ",
				
				// Content checks
				"id=", "u1",
				"role=", "admin",
				"id=", "u2",
				"role=", "user",
				"version=", "1.0",
			},
		},
		{
			name:   "pretty nil value",
			format: "{{.missing | pretty}}",
			data:   map[string]interface{}{},
			contains: []string{"null"},
		},
		{
			name:   "pretty empty map",
			format: "{{.empty | pretty}}",
			data:   map[string]interface{}{"empty": map[string]interface{}{}},
			contains: []string{"{}"},
		},
		{
			name:   "pretty empty array",
			format: "{{.empty | pretty}}",
			data:   map[string]interface{}{"empty": []interface{}{}},
			contains: []string{"[]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := []FormatterOption{}
			if tt.noColors {
				opts = append(opts, WithNoColors(true))
			}
			
			formatter, err := NewTemplateFormatter(tt.format, opts...)
			if err != nil {
				t.Fatalf("Failed to create formatter: %v", err)
			}

			result, err := formatter.Format(tt.data)
			if err != nil {
				t.Fatalf("Format failed: %v", err)
			}

			// Check for expected content
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, but got:\n%s", expected, result)
				}
			}

			// Check for unexpected content
			for _, unexpected := range tt.notContains {
				if strings.Contains(result, unexpected) {
					t.Errorf("Expected result NOT to contain %q, but it does:\n%s", unexpected, result)
				}
			}
		})
	}
}

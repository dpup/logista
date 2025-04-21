package formatter

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
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
			format:   "{{.level | pad 10}} {{.message}}",
			data:     map[string]interface{}{"level": "info", "message": "test message"},
			expected: "info       test message",
		},
		{
			name:     "pad function with longer text",
			format:   "{{.level | pad 3}} {{.message}}",
			data:     map[string]interface{}{"level": "warning", "message": "test message"},
			expected: "warning test message",
		},
		{
			name:     "pad function with nil value",
			format:   "{{.missing | pad 5}} {{.message}}",
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

	err = formatter.ProcessStream(r, &buf, formatter, nil, false)
	if err != nil {
		t.Fatalf("ProcessStream failed: %v", err)
	}

	if buf.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, buf.String())
	}
}

func TestMultFunction(t *testing.T) {
	tests := []struct {
		name     string
		arg      interface{}
		value    interface{}
		expected string
	}{
		{
			name:     "integer * integer",
			arg:      5,
			value:    10,
			expected: "50",
		},
		{
			name:     "float * integer",
			arg:      2.5,
			value:    4,
			expected: "10",
		},
		{
			name:     "integer * float",
			arg:      3,
			value:    1.5,
			expected: "4.50",
		},
		{
			name:     "float * float",
			arg:      2.5,
			value:    3.5,
			expected: "8.75",
		},
		{
			name:     "string number * integer",
			arg:      "5",
			value:    10,
			expected: "50",
		},
		{
			name:     "integer * string number",
			arg:      5,
			value:    "10",
			expected: "50",
		},
		{
			name:     "json.Number * integer",
			arg:      json.Number("5"),
			value:    10,
			expected: "50",
		},
		{
			name:     "non-numeric arg",
			arg:      "abc",
			value:    10,
			expected: "NaN",
		},
		{
			name:     "non-numeric value",
			arg:      5,
			value:    "xyz",
			expected: "NaN",
		},
		{
			name:     "nil arg",
			arg:      nil,
			value:    10,
			expected: "NaN",
		},
		{
			name:     "nil value",
			arg:      5,
			value:    nil,
			expected: "NaN",
		},
	}

	formatter := &TemplateFormatter{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.multFunc(tt.arg, tt.value)
			if result != tt.expected {
				t.Errorf("multFunc(%v, %v) = %v, want %v", tt.arg, tt.value, result, tt.expected)
			}
		})
	}
}

func TestPrintfFunction(t *testing.T) {
	tests := []struct {
		name     string
		format   interface{}
		value    interface{}
		expected string
	}{
		{
			name:     "integer with decimal format",
			format:   "%.2f",
			value:    10,
			expected: "%!f(int=10)",
		},
		{
			name:     "float with precision",
			format:   "%.1f",
			value:    3.14159,
			expected: "3.1",
		},
		{
			name:     "string with padding",
			format:   "%-10s",
			value:    "test",
			expected: "test      ",
		},
		{
			name:     "multiple values in format string",
			format:   "%s: %d",
			value:    "count",
			expected: "count: %!d(MISSING)",
		},
		{
			name:     "nil format",
			format:   nil,
			value:    "test",
			expected: "test",
		},
		{
			name:     "nil value",
			format:   "%.2f",
			value:    nil,
			expected: "<nil>",
		},
		{
			name:     "non-string format",
			format:   123,
			value:    "test",
			expected: "123: test",
		},
	}

	formatter := &TemplateFormatter{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.printfFunc(tt.format, tt.value)
			if result != tt.expected {
				t.Errorf("printfFunc(%v, %v) = %v, want %v", tt.format, tt.value, result, tt.expected)
			}
		})
	}
}

func TestTemplateWithNewFunctions(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "mult function with integers",
			template: "{{.value | mult 5}}",
			data:     map[string]interface{}{"value": 10},
			expected: "50",
		},
		{
			name:     "mult function with non-numeric value",
			template: "{{.value | mult 5}}",
			data:     map[string]interface{}{"value": "abc"},
			expected: "NaN",
		},
		{
			name:     "printf function with float format",
			template: "{{.value | printf \"%.2f\"}}",
			data:     map[string]interface{}{"value": 3.14159},
			expected: "3.14",
		},
		{
			name:     "printf function with string format",
			template: "{{.value | printf \"Value: %s\"}}",
			data:     map[string]interface{}{"value": "test"},
			expected: "Value: test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter, err := NewTemplateFormatter(tt.template)
			if err != nil {
				t.Fatalf("Failed to create formatter: %v", err)
			}

			result, err := formatter.Format(tt.data)
			if err != nil {
				t.Fatalf("Format error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Format result = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestProcessStreamWithNonJSON(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		handleNonJSON   bool
		noColors        bool
		expectedSuccess bool
		expectedOutput  string
	}{
		{
			name:            "valid JSON only",
			input:           `{"level":"info","message":"test1"}` + "\n" + `{"level":"error","message":"test2"}`,
			handleNonJSON:   false, // Default behavior
			noColors:        false,
			expectedSuccess: true,
			expectedOutput:  "info test1\nerror test2\n",
		},
		{
			name:            "non-JSON with handling disabled",
			input:           `{"level":"info","message":"test1"}` + "\n" + `This is not JSON` + "\n" + `{"level":"error","message":"test2"}`,
			handleNonJSON:   false, // Default behavior
			noColors:        false,
			expectedSuccess: false,
			expectedOutput:  "",
		},
		{
			name:            "non-JSON with handling enabled",
			input:           `{"level":"info","message":"test1"}` + "\n" + `This is not JSON` + "\n" + `{"level":"error","message":"test2"}`,
			handleNonJSON:   true,
			noColors:        false,
			expectedSuccess: true,
			expectedOutput:  "info test1\n\n\033[31m>>>\033[0m This is not JSON\n\nerror test2\n",
		},
		{
			name:            "multiple non-JSON lines with handling enabled",
			input:           `{"level":"info","message":"test1"}` + "\n" + `This is not JSON` + "\n" + `Another non-JSON line` + "\n" + `{"level":"error","message":"test2"}`,
			handleNonJSON:   true,
			noColors:        false,
			expectedSuccess: true,
			expectedOutput:  "info test1\n\n\033[31m>>>\033[0m This is not JSON\n\033[31m>>>\033[0m Another non-JSON line\n\nerror test2\n",
		},
		{
			name:            "non-JSON with handling enabled and no colors",
			input:           `{"level":"info","message":"test1"}` + "\n" + `This is not JSON` + "\n" + `{"level":"error","message":"test2"}`,
			handleNonJSON:   true,
			noColors:        true,
			expectedSuccess: true,
			expectedOutput:  "info test1\n\n>>> This is not JSON\n\nerror test2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := []FormatterOption{}
			if tt.noColors {
				options = append(options, WithNoColors(true))
			}

			formatter, err := NewTemplateFormatter("{{.level}} {{.message}}", options...)
			if err != nil {
				t.Fatalf("Failed to create formatter: %v", err)
			}

			r := strings.NewReader(tt.input)
			var buf bytes.Buffer

			err = formatter.ProcessStream(r, &buf, formatter, nil, tt.handleNonJSON)

			// Check if the error result matches expectations
			if tt.expectedSuccess && err != nil {
				t.Fatalf("ProcessStream failed but expected success: %v", err)
			} else if !tt.expectedSuccess && err == nil {
				t.Fatalf("ProcessStream succeeded but expected failure")
			}

			// Only check output if we expected success
			if tt.expectedSuccess {
				if buf.String() != tt.expectedOutput {
					t.Errorf("Expected:\n%s\n\nGot:\n%s", tt.expectedOutput, buf.String())
				}
			}
		})
	}
}
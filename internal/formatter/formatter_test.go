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
			format:   "{level} {message}",
			data:     map[string]interface{}{"level": "info", "message": "test message"},
			expected: "info test message",
		},
		{
			name:     "nested fields",
			format:   "{level} {context.user.id}",
			data:     map[string]interface{}{"level": "info", "context": map[string]interface{}{"user": map[string]interface{}{"id": "123"}}},
			expected: "info 123",
		},
		{
			name:     "missing field",
			format:   "{level} {missing}",
			data:     map[string]interface{}{"level": "info"},
			expected: "info <no value>",
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
			format:     "{timestamp | date}",
			data:       map[string]interface{}{"timestamp": isoDate},
			expected:   "2024-03-10 15:04:05",
			dateFormat: "2006-01-02 15:04:05",
		},
		{
			name:       "date function with unix timestamp",
			format:     "{timestamp | date}",
			data:       map[string]interface{}{"timestamp": json.Number(strconv.FormatInt(now.Unix(), 10))},
			expected:   "2024-03-10 15:04:05",
			dateFormat: "2006-01-02 15:04:05",
		},
		{
			name:       "date function with float unix timestamp",
			format:     "{timestamp | date}",
			data:       map[string]interface{}{"timestamp": unixTimestamp},
			expected:   "2024-03-10 15:04:05",
			dateFormat: "2006-01-02 15:04:05",
		},
		{
			name:       "date function with custom format",
			format:     "{timestamp | date}",
			data:       map[string]interface{}{"timestamp": isoDate},
			expected:   "10/03/2024",
			dateFormat: "02/01/2006",
		},
		{
			name:       "date function with common log format",
			format:     "{timestamp | date}",
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

func TestProcessStream(t *testing.T) {
	input := `{"level":"info","message":"test1"}
{"level":"error","message":"test2"}`

	expected := "info test1\nerror test2\n"

	formatter, err := NewTemplateFormatter("{level} {message}")
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

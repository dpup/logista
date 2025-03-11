package formatter

import (
	"testing"
)

func TestPreProcessTemplate(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expected       string
		options        PreProcessTemplateOptions
		checkFormatted bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
			options:  DefaultPreProcessTemplateOptions(),
		},
		{
			name:     "already valid Go template",
			input:    "{{.level}} {{.message}}",
			expected: "{{.level}} {{.message}}",
			options:  DefaultPreProcessTemplateOptions(),
		},
		{
			name:     "simple syntax",
			input:    "{level} {message}",
			expected: "{{.level}} {{.message}}",
			options:  DefaultPreProcessTemplateOptions(),
		},
		{
			name:     "simple syntax with pipe",
			input:    "{level | pad 10} {message}",
			expected: "{{.level | pad 10}} {{.message}}",
			options:  DefaultPreProcessTemplateOptions(),
		},
		{
			name:     "simple syntax with arguments",
			input:    "{level} {message | color \"red\"}",
			expected: "{{.level}} {{.message | color \"red\"}}",
			options:  DefaultPreProcessTemplateOptions(),
		},
		{
			name:     "@ syntax",
			input:    "{@grpc.service}",
			expected: "{{index . \"grpc.service\"}}",
			options:  DefaultPreProcessTemplateOptions(),
		},
		{
			name:     "@ syntax with pipe",
			input:    "{@grpc.service | color \"blue\"}",
			expected: "{{index . \"grpc.service\" | color \"blue\"}}",
			options:  DefaultPreProcessTemplateOptions(),
		},
		{
			name:     "mixed @ and simple syntax",
			input:    "{level}: {@grpc.service} - {message}",
			expected: "{{.level}}: {{index . \"grpc.service\"}} - {{.message}}",
			options:  DefaultPreProcessTemplateOptions(),
		},
		{
			name:    "only process @ syntax",
			input:   "{level}: {@grpc.service}",
			expected: "{level}: {{index . \"grpc.service\"}}",
			options: PreProcessTemplateOptions{
				EnableAtSyntax:     true,
				EnableSimpleSyntax: false,
			},
		},
		{
			name:    "only process simple syntax",
			input:   "{level}: {@grpc.service}",
			expected: "{{.level}}: {{.@grpc.service}}",
			options: PreProcessTemplateOptions{
				EnableAtSyntax:     false,
				EnableSimpleSyntax: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PreProcessTemplate(tt.input, tt.options)
			if result != tt.expected {
				t.Errorf("PreProcessTemplate(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormattingWithPreprocessor(t *testing.T) {
	// Complex log entry with dotted fields
	logEntry := map[string]interface{}{
		"level":        "info", 
		"message":      "Request completed",
		"grpc.service": "users.UserService",
		"grpc.method":  "GetUser",
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "@ syntax format",
			template: "{@grpc.service} - {@grpc.method}",
			expected: "users.UserService - GetUser",
		},
		{
			name:     "mixed @ and simple syntax",
			template: "{level}: {@grpc.service}.{@grpc.method} - {message}",
			expected: "info: users.UserService.GetUser - Request completed",
		},
		{
			name:     "simple syntax with pipes",
			template: "{level | pad 10} {message}",
			expected: "info       Request completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create formatter with the template
			formatter, err := NewTemplateFormatter(tt.template)
			if err != nil {
				t.Fatalf("Failed to create formatter: %v", err)
			}
			
			// Format the log entry
			result, err := formatter.Format(logEntry)
			if err != nil {
				t.Fatalf("Format failed: %v", err)
			}
			
			// Check result
			if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}
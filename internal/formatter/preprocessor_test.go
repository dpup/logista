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
			name:    "disable simple syntax",
			input:   "{level}: {message}",
			expected: "{level}: {message}",
			options: PreProcessTemplateOptions{
				EnableSimpleSyntax: false,
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
	// Log entry
	logEntry := map[string]interface{}{
		"level":        "info", 
		"message":      "Request completed",
		"grpc_service": "users.UserService",
		"grpc_method":  "GetUser",
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "simple syntax format",
			template: "{grpc_service} - {grpc_method}",
			expected: "users.UserService - GetUser",
		},
		{
			name:     "simple syntax with context",
			template: "{level}: {grpc_service}.{grpc_method} - {message}",
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
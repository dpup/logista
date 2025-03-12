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
			name:     "disable simple syntax",
			input:    "{level}: {message}",
			expected: "{level}: {message}",
			options: PreProcessTemplateOptions{
				EnableSimpleSyntax: false,
			},
		},
		{
			name:     "at symbol syntax",
			input:    "@level @message",
			expected: "(index . \"level\") (index . \"message\")",
			options:  DefaultPreProcessTemplateOptions(),
		},
		{
			name:     "at symbol with complex names",
			input:    "@user.name @request-id @response_code",
			expected: "(index . \"user.name\") (index . \"request-id\") (index . \"response_code\")",
			options:  DefaultPreProcessTemplateOptions(),
		},
		{
			name:     "mixed at symbol and simple syntax",
			input:    "@level: {message}",
			expected: "(index . \"level\"): {{.message}}",
			options:  DefaultPreProcessTemplateOptions(),
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
		"level":         "info",
		"message":       "Request completed",
		"grpc_service":  "users.UserService",
		"grpc_method":   "GetUser",
		"user.name":     "johndoe",
		"request-id":    "abc123",
		"response_code": 200,
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

func TestFormattingWithAtSymbolSyntax(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "at symbol syntax",
			template: "@level: @message",
			expected: "(index . \"level\"): (index . \"message\")",
		},
		{
			name:     "at symbol with complex names",
			template: "User: @user.name, ID: @request-id, Status: @response_code",
			expected: "User: (index . \"user.name\"), ID: (index . \"request-id\"), Status: (index . \"response_code\")",
		},
		{
			name:     "mixed at symbol and simple syntax",
			template: "@level: {grpc_service}.{grpc_method} - @message",
			expected: "(index . \"level\"): {{.grpc_service}}.{{.grpc_method}} - (index . \"message\")",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Process the template
			result := PreProcessTemplate(tt.template, DefaultPreProcessTemplateOptions())

			// Check result
			if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}

func TestTransformAtSymbol(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no at symbols",
			input:    "level message",
			expected: "level message",
		},
		{
			name:     "single at symbol",
			input:    "@level",
			expected: "(index . \"level\")",
		},
		{
			name:     "multiple at symbols",
			input:    "@level @message",
			expected: "(index . \"level\") (index . \"message\")",
		},
		{
			name:     "at symbol with dot",
			input:    "@request.id",
			expected: "(index . \"request.id\")",
		},
		{
			name:     "at symbol with hyphen",
			input:    "@request-id",
			expected: "(index . \"request-id\")",
		},
		{
			name:     "at symbol with underscore",
			input:    "@response_code",
			expected: "(index . \"response_code\")",
		},
		{
			name:     "at symbol in text",
			input:    "Level: @level - Message: @message",
			expected: "Level: (index . \"level\") - Message: (index . \"message\")",
		},
		{
			name:     "at symbol inside template markers",
			input:    "{{@level}} {@level}",
			expected: "{{(index . \"level\")}} {(index . \"level\")}",
		},
		{
			name:     "at sign not part of symbol",
			input:    "email@example.com",
			expected: "email@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformAtSymbol(tt.input)
			if result != tt.expected {
				t.Errorf("transformAtSymbol(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

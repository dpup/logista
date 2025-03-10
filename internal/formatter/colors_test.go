package formatter

import (
	"testing"
)

func TestApplyColorToString(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		colorName string
		noColors  bool
		expected  string
	}{
		{
			name:      "red color",
			content:   "This is red",
			colorName: "red",
			noColors:  false,
			expected:  "\033[31mThis is red\033[0m",
		},
		{
			name:      "blue color",
			content:   "This is blue",
			colorName: "blue",
			noColors:  false,
			expected:  "\033[34mThis is blue\033[0m",
		},
		{
			name:      "bold formatting",
			content:   "This is bold",
			colorName: "bold",
			noColors:  false,
			expected:  "\033[1mThis is bold\033[0m",
		},
		{
			name:      "non-existent color",
			content:   "Not colored",
			colorName: "nonexistent",
			noColors:  false,
			expected:  "Not colored",
		},
		{
			name:      "none color",
			content:   "No color",
			colorName: "none",
			noColors:  false,
			expected:  "No color",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyColorToString(tt.content, tt.colorName)
			if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}

func TestColorByLevelName(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected string
	}{
		{
			name:     "error level",
			level:    "error",
			expected: "red",
		},
		{
			name:     "warning level",
			level:    "warning",
			expected: "yellow",
		},
		{
			name:     "info level",
			level:    "info",
			expected: "green",
		},
		{
			name:     "debug level",
			level:    "debug",
			expected: "cyan",
		},
		{
			name:     "trace level",
			level:    "trace",
			expected: "blue",
		},
		{
			name:     "unknown level",
			level:    "unknown",
			expected: "white",
		},
		{
			name:     "case insensitive",
			level:    "ERROR",
			expected: "red",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ColorByLevelName(tt.level)
			if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}

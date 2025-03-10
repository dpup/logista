package formatter

import (
	"testing"
)

func TestApplyColors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		noColors    bool
		expected    string
		description string
		skip        bool // Skip some tests that don't work with the simplified implementation
	}{
		{
			name:        "simple red text with standard closing tag",
			input:       "<red>This is red</red>",
			noColors:    false,
			expected:    "\033[31mThis is red\033[0m",
			description: "Basic red foreground color with standard closing tag",
		},
		{
			name:        "simple red text with simplified closing tag",
			input:       "<red>This is red</>",
			noColors:    false,
			expected:    "\033[31mThis is red\033[0m",
			description: "Basic red foreground color with simplified closing tag",
		},
		{
			name:        "multiple color tags with standard closing",
			input:       "<red>Red</red> and <blue>Blue</blue>",
			noColors:    false,
			expected:    "\033[31mRed\033[0m and \033[34mBlue\033[0m",
			description: "Multiple color tags in a string with standard closing",
		},
		{
			name:        "multiple color tags with simplified closing",
			input:       "<red>Red</> and <blue>Blue</>",
			noColors:    false,
			expected:    "\033[31mRed\033[0m and \033[34mBlue\033[0m",
			description: "Multiple color tags in a string with simplified closing",
		},
		{
			name:        "nested color tags",
			input:       "<red>Red <bold>and bold</bold></red>",
			noColors:    false,
			expected:    "\033[31mRed \033[1mand bold\033[0m\033[0m",
			description: "Nested color tags with different styles",
		},
		{
			name:        "non-existent color",
			input:       "<nonexistent>Not colored</nonexistent>",
			noColors:    false,
			expected:    "Not colored",
			description: "Non-existent color tag should not apply styling",
		},
		{
			name:        "multiple styles with standard closing",
			input:       "<bold red>Bold and red</bold red>",
			noColors:    false,
			expected:    "\033[1;31mBold and red\033[0m",
			description: "Multiple styles in a single tag with standard closing",
		},
		{
			name:        "multiple styles with simplified closing",
			input:       "<bold red>Bold and red</>",
			noColors:    false,
			expected:    "\033[1;31mBold and red\033[0m",
			description: "Multiple styles in a single tag with simplified closing",
		},
		{
			name:        "background color with standard closing",
			input:       "<bg-green>Green background</bg-green>",
			noColors:    false,
			expected:    "\033[42mGreen background\033[0m",
			description: "Background color style with standard closing",
		},
		{
			name:        "background color with simplified closing",
			input:       "<bg-green>Green background</>",
			noColors:    false,
			expected:    "\033[42mGreen background\033[0m",
			description: "Background color style with simplified closing",
		},
		{
			name:        "combined foreground and background with standard closing",
			input:       "<red bg-yellow>Red text on yellow</red bg-yellow>",
			noColors:    false,
			expected:    "\033[31;43mRed text on yellow\033[0m",
			description: "Combined foreground and background colors with standard closing",
		},
		{
			name:        "combined foreground and background with simplified closing",
			input:       "<red bg-yellow>Red text on yellow</>",
			noColors:    false,
			expected:    "\033[31;43mRed text on yellow\033[0m",
			description: "Combined foreground and background colors with simplified closing",
		},
		{
			name:        "no colors mode",
			input:       "<red>Red</red> and <blue>Blue</blue>",
			noColors:    true,
			expected:    "Red and Blue",
			description: "With noColors=true, tags should be stripped",
		},
		{
			name:        "complex nesting",
			input:       "<bold>Bold <italic>and italic <red>and red</red></italic></bold>",
			noColors:    false,
			expected:    "\033[1mBold \033[3mand italic \033[31mand red\033[0m\033[0m\033[0m",
			description: "Complex nesting of styles",
		},
		{
			name:        "tag with spaces and standard closing",
			input:       "<bold  red>Bold and red</bold  red>",
			noColors:    false,
			expected:    "\033[1;31mBold and red\033[0m",
			description: "Tags with extra spaces with standard closing",
		},
		{
			name:        "tag with spaces and simplified closing",
			input:       "<bold  red>Bold and red</>",
			noColors:    false,
			expected:    "\033[1;31mBold and red\033[0m",
			description: "Tags with extra spaces with simplified closing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip("Test skipped - not supported")
			}
			result := ApplyColors(tt.input, tt.noColors)
			if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q\nDescription: %s", tt.expected, result, tt.description)
			}
		})
	}
}

func TestStripColorTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		skip     bool // Skip some tests that don't yet work
	}{
		{
			name:     "simple tag with standard closing",
			input:    "<red>This is red</red>",
			expected: "This is red",
		},
		{
			name:     "simple tag with simplified closing",
			input:    "<red>This is red</>",
			expected: "This is red",
		},
		{
			name:     "multiple tags with standard closing",
			input:    "<red>Red</red> and <blue>Blue</blue>",
			expected: "Red and Blue",
		},
		{
			name:     "multiple tags with simplified closing",
			input:    "<red>Red</> and <blue>Blue</>",
			expected: "Red and Blue",
		},
		{
			name:     "nested tags",
			input:    "<red>Red <bold>and bold</bold></red>",
			expected: "Red and bold",
		},
		{
			name:     "invalid tags",
			input:    "No <tags> here",
			expected: "No <tags> here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip("Test skipped - not supported")
			}
			result := stripColorTags(tt.input)
			if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}

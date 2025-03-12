package formatter

import (
	"regexp"
	"strings"
)

// PreProcessTemplateOptions holds configuration options for template pre-processing
type PreProcessTemplateOptions struct {
	// Whether to enable the simple {field} syntax
	EnableSimpleSyntax bool
}

// DefaultPreProcessTemplateOptions returns the default options for template pre-processing
func DefaultPreProcessTemplateOptions() PreProcessTemplateOptions {
	return PreProcessTemplateOptions{
		EnableSimpleSyntax: true,
	}
}

// PreProcessTemplate transforms custom logista syntax into standard go template
// syntax.
func PreProcessTemplate(template string, options PreProcessTemplateOptions) string {
	// Skip processing for empty template
	if template == "" {
		return template
	}

	// Transform @symbol to (index . "symbol")
	template = transformAtSymbol(template)

	return transformSimpleSyntax(options, template)
}

// transformSimpleSyntax transforms template strings from simplified syntax to
// full Go template syntax
// It handles:
// {field} -> {{.field}} (when not already using Go template syntax)
func transformSimpleSyntax(options PreProcessTemplateOptions, template string) string {
	// Skip processing if simple syntax is disabled
	if !options.EnableSimpleSyntax {
		return template
	}

	// Check if this is already valid Go template syntax (has {{ but no {)
	if strings.Contains(template, "{{") &&
		!strings.Contains(strings.ReplaceAll(template, "{{", ""), "{") {
		return template
	}

	// Process template character by character to handle all cases
	var result strings.Builder
	i := 0

	for i < len(template) {
		if template[i] == '{' &&
			(i+1 >= len(template) || template[i+1] != '{') {
			// Find closing brace
			start := i
			openBraces := 1
			i++ // Skip past the {

			for i < len(template) {
				if template[i] == '{' {
					openBraces++
				} else if template[i] == '}' {
					openBraces--
					if openBraces == 0 {
						break
					}
				}
				i++
			}

			if i < len(template) { // Found closing brace
				// Replace {field} with {{.field}}
				result.WriteString("{{.")
				result.WriteString(template[start+1 : i])
				result.WriteString("}}")
				i++ // Skip past the closing brace
			} else {
				// No closing brace found, add the original text
				result.WriteString(template[start:])
				i = len(template)
			}
		} else {
			// Copy character as is
			result.WriteByte(template[i])
			i++
		}
	}

	return result.String()
}

// transformAtSymbol transforms @symbol syntax to (index . "symbol")
// The 'symbol' can contain alphanumeric characters, period, hyphen, and underscore.
func transformAtSymbol(template string) string {
	// \B@([a-zA-Z0-9._-]+) matches @symbol where:
	// - \B ensures it's not preceded by a word character (prevents matching email@example.com)
	// - symbol consists of letters, numbers, periods, hyphens, and underscores
	re := regexp.MustCompile(`\B@([a-zA-Z0-9._-]+)`)

	// Replace all occurrences of @symbol with (index . "symbol")
	return re.ReplaceAllString(template, `(index . "$1")`)
}

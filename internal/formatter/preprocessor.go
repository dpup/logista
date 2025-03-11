package formatter

import (
	"strings"
)

// PreProcessTemplateOptions holds configuration options for template pre-processing
type PreProcessTemplateOptions struct {
	// Whether to enable the @ syntax for accessing fields with dots
	EnableAtSyntax bool
	// Whether to enable the simple {field} syntax
	EnableSimpleSyntax bool
}

// DefaultPreProcessTemplateOptions returns the default options for template pre-processing
func DefaultPreProcessTemplateOptions() PreProcessTemplateOptions {
	return PreProcessTemplateOptions{
		EnableAtSyntax:     true,
		EnableSimpleSyntax: true,
	}
}

// PreProcessTemplate transforms template strings from simplified syntax to full Go template syntax
// It handles:
// 1. {@field.with.dots} -> {{index . "field.with.dots"}}
// 2. {field} -> {{.field}} (when not already using Go template syntax)
func PreProcessTemplate(template string, options PreProcessTemplateOptions) string {
	// Skip processing for empty template
	if template == "" {
		return template
	}
	
	// Check if this is already valid Go template syntax (has {{ but no { or {@)
	if strings.Contains(template, "{{") && 
	   !strings.Contains(template, "{@") && 
	   !strings.Contains(strings.ReplaceAll(template, "{{", ""), "{") {
		return template
	}

	// Process template character by character to handle all cases
	var result strings.Builder
	i := 0
	
	for i < len(template) {
		// Process @ syntax
		if options.EnableAtSyntax && i+1 < len(template) && 
		   template[i] == '{' && template[i+1] == '@' {
			// Find closing brace
			start := i
			openBraces := 1
			i += 2 // Skip past the {@
			
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
				// Extract content between {@...}
				content := template[start+2:i]
				
				// Check if there's a pipe in the content
				pipeIndex := strings.Index(content, "|")
				
				if pipeIndex != -1 {
					// Split into field name and pipes
					fieldName := strings.TrimSpace(content[:pipeIndex])
					pipeContent := strings.TrimSpace(content[pipeIndex+1:])
					result.WriteString(`{{index . "` + fieldName + `" | ` + pipeContent + `}}`)
				} else {
					// Just the field name
					fieldName := strings.TrimSpace(content)
					result.WriteString(`{{index . "` + fieldName + `"}}`)
				}
				i++ // Skip past the closing brace
			} else {
				// No closing brace found, add the original text
				result.WriteString(template[start:])
				i = len(template)
			}
		} else if options.EnableSimpleSyntax && template[i] == '{' && 
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
				result.WriteString(template[start+1:i])
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
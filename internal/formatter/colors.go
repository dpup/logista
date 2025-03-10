package formatter

import (
	"fmt"
	"regexp"
	"strings"
)

// Color names and their ANSI color codes
var colorCodes = map[string]string{
	// Foreground colors
	"black":   "30",
	"red":     "31",
	"green":   "32",
	"yellow":  "33",
	"blue":    "34",
	"magenta": "35",
	"cyan":    "36",
	"white":   "37",
	"gray":    "90",

	// Bright foreground colors
	"brightred":     "91",
	"brightgreen":   "92",
	"brightyellow":  "93",
	"brightblue":    "94",
	"brightmagenta": "95",
	"brightcyan":    "96",
	"brightwhite":   "97",

	// Background colors
	"bg-black":   "40",
	"bg-red":     "41",
	"bg-green":   "42",
	"bg-yellow":  "43",
	"bg-blue":    "44",
	"bg-magenta": "45",
	"bg-cyan":    "46",
	"bg-white":   "47",
	"bg-gray":    "100",

	// Bright background colors
	"bg-brightred":     "101",
	"bg-brightgreen":   "102",
	"bg-brightyellow":  "103",
	"bg-brightblue":    "104",
	"bg-brightmagenta": "105",
	"bg-brightcyan":    "106",
	"bg-brightwhite":   "107",

	// Special formatting
	"bold":      "1",
	"italic":    "3",
	"underline": "4",
	"dim":       "2",
}

// Reset code
const ansiReset = "\033[0m"

// ApplyColors processes the input string and replaces color tags with ANSI color codes
func ApplyColors(input string, noColors bool) string {
	if noColors {
		return stripColorTags(input)
	}

	// Simple tag pattern that supports both standard HTML-like tags and simplified </> closing tag
	colorTagPattern := `<([^>]+)>([^<]*)(</[^>]*>|</>)`

	// Process the string
	result := input

	// Iteratively apply color replacements, starting with the innermost tags
	for {
		re := regexp.MustCompile(colorTagPattern)
		matches := re.FindStringSubmatchIndex(result)

		if len(matches) == 0 {
			break // No more color tags
		}

		// Extract tag name and content
		tagNameStart, tagNameEnd := matches[2], matches[3]
		contentStart, contentEnd := matches[4], matches[5]

		tagName := result[tagNameStart:tagNameEnd]
		content := result[contentStart:contentEnd]

		// Apply color codes
		colored := applyColorCode(tagName, content)

		// Replace the tag in the result
		result = result[:matches[0]] + colored + result[matches[1]:]
	}

	return result
}

// applyColorCode applies the ANSI color code for the given tag name to the content
func applyColorCode(tagName string, content string) string {
	// Handle multiple styles specified with spaces
	styles := strings.Fields(tagName)

	var codes []string
	for _, style := range styles {
		if code, ok := colorCodes[strings.ToLower(style)]; ok {
			codes = append(codes, code)
		}
	}

	if len(codes) == 0 {
		// If no valid codes found, return content unchanged
		return content
	}

	// Combine all style codes
	combinedCode := strings.Join(codes, ";")
	return fmt.Sprintf("\033[%sm%s%s", combinedCode, content, ansiReset)
}

// stripColorTags removes color tags from the input string without applying colors
func stripColorTags(input string) string {
	// Pattern that supports both standard HTML-like tags and simplified </> closing tag
	pattern := `<[^>]+>([^<]*)(</[^>]*>|</>)`
	re := regexp.MustCompile(pattern)

	// Iteratively strip tags, from innermost to outermost
	result := input
	for {
		prevResult := result
		result = re.ReplaceAllString(result, "$1")

		// If no changes were made, we're done
		if prevResult == result {
			break
		}
	}

	return result
}

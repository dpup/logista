package formatter

import (
	"fmt"
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

// ApplyColorToString applies a specific color to a string value
func ApplyColorToString(content string, colorName string) string {
	if colorName == "none" {
		return content
	}
	
	if code, ok := colorCodes[colorName]; ok {
		return fmt.Sprintf("\033[%sm%s%s", code, content, ansiReset)
	}
	
	return content // Return unchanged if color not found
}

// ColorByLevelName returns the appropriate color for a log level
func ColorByLevelName(level string) string {
	levelStr := strings.ToLower(level)
	
	switch levelStr {
	case "error", "err", "fatal", "crit", "critical", "alert", "emergency":
		return "red"
	case "warn", "warning":
		return "yellow"
	case "info", "information":
		return "green"
	case "debug":
		return "cyan"
	case "trace":
		return "blue"
	default:
		return "white"
	}
}

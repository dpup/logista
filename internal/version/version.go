package version

import (
	_ "embed" // Required for go:embed directive
	"strings"
)

//go:embed version.txt
var versionFile string

// Version is set during build or read from the embedded version.txt file
var Version string

func init() {
	// Read version from embedded file if not set through build flags
	if Version == "" || Version == "dev" {
		Version = strings.TrimSpace(versionFile)
	}
}

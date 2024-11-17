package debug

import (
	"fmt"
	"os"
	"strings"
)

var isDebug bool

// Initialize sets the debug mode
func Initialize(debug bool) {
	isDebug = debug
}

// IsDebug returns whether debug mode is enabled
func IsDebug() bool {
	return isDebug
}

// Log prints debug information if debug mode is enabled
func Log(format string, args ...interface{}) {
	if isDebug {
		message := fmt.Sprintf(format, args...)
		fmt.Fprintf(os.Stderr, "DEBUG: %s\n", strings.TrimSpace(message))
	}
}

// Dump prints a variable's content in debug mode
func Dump(name string, value interface{}) {
	if isDebug {
		fmt.Fprintf(os.Stderr, "DEBUG: %s = %+v\n", name, value)
	}
} 
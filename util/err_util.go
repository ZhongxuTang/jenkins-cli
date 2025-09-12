package util

import (
	"os"

	"github.com/fatih/color"
)

// HandleError provides a centralized way to handle errors with colored output
func HandleError(err error, context string) {
	if err != nil {
		color.Red("❌ %s: %v", context, err)
	}
}

// HandleErrorWithExit provides error handling with program exit
func HandleErrorWithExit(err error, context string, exitCode int) {
	if err != nil {
		color.Red("❌ %s: %v", context, err)
		os.Exit(exitCode)
	}
}

// Must is deprecated: use explicit error handling instead
// Keeping for backward compatibility but discouraged
func Must(e error) {
	if e != nil {
		color.Red("❌ Critical error: %v", e)
		color.Yellow("💡 Consider using explicit error handling instead of Must()")
		panic(e)
	}
}

// MustAny is deprecated: use explicit error handling instead
// Keeping for backward compatibility but discouraged
func MustAny[A any](a A, e error) A {
	if e != nil {
		color.Red("❌ Critical error: %v", e)
		color.Yellow("💡 Consider using explicit error handling instead of MustAny()")
		panic(e)
	}
	return a
}

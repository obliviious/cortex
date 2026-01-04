// Package ui provides terminal UI components with colors.
package ui

import (
	"fmt"
	"os"
	"runtime"
)

// ANSI color codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"

	// Regular colors
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	// Bright colors
	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	// Background colors
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
)

var colorsEnabled = true

func init() {
	// Disable colors on Windows unless TERM is set
	if runtime.GOOS == "windows" {
		if os.Getenv("TERM") == "" && os.Getenv("WT_SESSION") == "" {
			colorsEnabled = false
		}
	}
	// Respect NO_COLOR env var
	if os.Getenv("NO_COLOR") != "" {
		colorsEnabled = false
	}
}

// SetColorsEnabled enables or disables color output.
func SetColorsEnabled(enabled bool) {
	colorsEnabled = enabled
}

// Colorize wraps text with color codes if colors are enabled.
func Colorize(color, text string) string {
	if !colorsEnabled {
		return text
	}
	return color + text + Reset
}

// Helper functions for common colors
func RedText(text string) string     { return Colorize(Red, text) }
func GreenText(text string) string   { return Colorize(Green, text) }
func YellowText(text string) string  { return Colorize(Yellow, text) }
func BlueText(text string) string    { return Colorize(Blue, text) }
func CyanText(text string) string    { return Colorize(Cyan, text) }
func MagentaText(text string) string { return Colorize(Magenta, text) }
func BoldText(text string) string    { return Colorize(Bold, text) }
func DimText(text string) string     { return Colorize(Dim, text) }

func BrightCyanText(text string) string    { return Colorize(BrightCyan, text) }
func BrightGreenText(text string) string   { return Colorize(BrightGreen, text) }
func BrightYellowText(text string) string  { return Colorize(BrightYellow, text) }
func BrightMagentaText(text string) string { return Colorize(BrightMagenta, text) }

// Success prints a success message
func Success(format string, args ...interface{}) {
	fmt.Printf(GreenText("✓ ")+format+"\n", args...)
}

// Error prints an error message
func Error(format string, args ...interface{}) {
	fmt.Printf(RedText("✗ ")+format+"\n", args...)
}

// Warning prints a warning message
func Warning(format string, args ...interface{}) {
	fmt.Printf(YellowText("⚠ ")+format+"\n", args...)
}

// Info prints an info message
func Info(format string, args ...interface{}) {
	fmt.Printf(CyanText("ℹ ")+format+"\n", args...)
}

// Task prints a task status
func Task(name, status string, success bool) {
	var statusColor string
	if success {
		statusColor = GreenText(status)
	} else {
		statusColor = RedText(status)
	}
	fmt.Printf("  %s %s\n", BoldText(name), statusColor)
}

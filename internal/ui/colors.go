// Package ui provides terminal UI components with colors.
package ui

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
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

	// Custom colors (256-color mode)
	Orange = "\033[38;5;208m" // Claude-like orange

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
func OrangeText(text string) string        { return Colorize(Orange, text) }

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
	fmt.Printf(OrangeText("ℹ ")+format+"\n", args...)
}

// Step prints a setup step with a dot indicator
func Step(format string, args ...interface{}) {
	fmt.Printf("  %s•%s %s"+format+"%s\n", Orange, Reset, Dim, Reset)
}

// StepDone prints a completed step
func StepDone(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s✓%s %s\n", Green, Reset, msg)
}

// PrintSetupStart prints the setup section header
func PrintSetupStart() {
	fmt.Printf("\n  %s○%s Setup\n", Orange, Reset)
}

// PrintSetupStep prints a setup step with green tick
func PrintSetupStep(text string) {
	fmt.Printf("    %s✓%s %s\n", Green, Reset, text)
}

// PrintSetupEnd prints the setup section footer
func PrintSetupEnd() {
	// No footer needed for cleaner look
}

// PrintConfigInfo prints configuration summary
func PrintConfigInfo(levels, maxParallel int, parallel bool) {
	if parallel {
		fmt.Printf("\n  %s⚡%s Parallel: %d levels, %d concurrent\n", Orange, Reset, levels, maxParallel)
	} else {
		fmt.Printf("\n  %s→%s Sequential execution\n", Orange, Reset)
	}
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

// Regex patterns for markdown stripping
var (
	// Headers: # ## ### etc
	headerRegex = regexp.MustCompile(`(?m)^#{1,6}\s+`)
	// Bold: **text** or __text__
	boldRegex = regexp.MustCompile(`(\*\*|__)(.+?)(\*\*|__)`)
	// Italic: *text* or _text_
	italicRegex = regexp.MustCompile(`(\*|_)(.+?)(\*|_)`)
	// Code blocks: ```lang\ncode\n```
	codeBlockRegex = regexp.MustCompile("(?s)```[a-z]*\\n?(.*?)```")
	// Inline code: `code`
	inlineCodeRegex = regexp.MustCompile("`([^`]+)`")
	// Links: [text](url)
	linkRegex = regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`)
	// Images: ![alt](url)
	imageRegex = regexp.MustCompile(`!\[([^\]]*)\]\([^\)]+\)`)
	// Blockquotes: > text
	blockquoteRegex = regexp.MustCompile(`(?m)^>\s+`)
	// Horizontal rules: --- or *** or ___
	hrRegex = regexp.MustCompile(`(?m)^[-*_]{3,}\s*$`)
	// Unordered lists: - or * or +
	ulRegex = regexp.MustCompile(`(?m)^[\s]*[-*+]\s+`)
	// Ordered lists: 1. 2. etc
	olRegex = regexp.MustCompile(`(?m)^[\s]*\d+\.\s+`)
	// Strikethrough: ~~text~~
	strikeRegex = regexp.MustCompile(`~~(.+?)~~`)
)

// StripMarkdown removes markdown formatting and returns plain text
func StripMarkdown(text string) string {
	result := text

	// Remove code blocks first (preserve content)
	result = codeBlockRegex.ReplaceAllString(result, "$1")

	// Remove inline code (preserve content)
	result = inlineCodeRegex.ReplaceAllString(result, "$1")

	// Remove images (replace with alt text or empty)
	result = imageRegex.ReplaceAllString(result, "$1")

	// Remove links (preserve link text)
	result = linkRegex.ReplaceAllString(result, "$1")

	// Remove headers
	result = headerRegex.ReplaceAllString(result, "")

	// Remove bold formatting
	result = boldRegex.ReplaceAllString(result, "$2")

	// Remove italic formatting
	result = italicRegex.ReplaceAllString(result, "$2")

	// Remove strikethrough
	result = strikeRegex.ReplaceAllString(result, "$1")

	// Remove blockquotes
	result = blockquoteRegex.ReplaceAllString(result, "")

	// Remove horizontal rules
	result = hrRegex.ReplaceAllString(result, "")

	// Simplify list markers to plain bullets
	result = ulRegex.ReplaceAllString(result, "  • ")
	result = olRegex.ReplaceAllStringFunc(result, func(match string) string {
		return "  • "
	})

	// Clean up extra blank lines
	result = regexp.MustCompile(`\n{3,}`).ReplaceAllString(result, "\n\n")

	// Trim leading/trailing whitespace
	result = strings.TrimSpace(result)

	return result
}

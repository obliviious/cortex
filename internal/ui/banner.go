package ui

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

// Cortex ASCII art - neural network style
const asciiArt = `
%s       ○━━━━━○       ○━━━━━○%s
%s        ╲   ╱ ╲     ╱ ╲   ╱%s
%s         ○━━━○━━━○━━━○━━━○%s
%s        ╱ ╲ ╱   ╲ ╱   ╲ ╱ ╲%s
%s       ○   ○     ○     ○   ○%s

%s  ██████╗ ██████╗ ██████╗ ████████╗███████╗██╗  ██╗%s
%s ██╔════╝██╔═══██╗██╔══██╗╚══██╔══╝██╔════╝╚██╗██╔╝%s
%s ██║     ██║   ██║██████╔╝   ██║   █████╗   ╚███╔╝%s
%s ██║     ██║   ██║██╔══██╗   ██║   ██╔══╝   ██╔██╗%s
%s ╚██████╗╚██████╔╝██║  ██║   ██║   ███████╗██╔╝ ██╗%s
%s  ╚═════╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚══════╝╚═╝  ╚═╝%s

%s       ○   ○     ○     ○   ○%s
%s        ╲ ╱ ╲   ╱ ╲   ╱ ╲ ╱%s
%s         ○━━━○━━━○━━━○━━━○%s
%s        ╱   ╲ ╱     ╲ ╱   ╲%s
%s       ○━━━━━○       ○━━━━━○%s

%s          ⚡ AI Agent Orchestrator ⚡%s
`

// PrintBanner prints the welcome banner with ASCII art
func PrintBanner(version string) {
	// Get username
	username := "User"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}

	// Get current directory
	cwd, _ := os.Getwd()
	homeDir, _ := os.UserHomeDir()
	displayPath := cwd
	if homeDir != "" && len(cwd) > len(homeDir) && cwd[:len(homeDir)] == homeDir {
		displayPath = "~" + cwd[len(homeDir):]
	}

	fmt.Println()

	// Welcome message
	welcomeMsg := fmt.Sprintf("Welcome back %s!", username)
	fmt.Printf("         %s%s%s\n", Bold+BrightCyan, welcomeMsg, Reset)

	// Print ASCII art with colors
	fmt.Printf(asciiArt,
		// Top neural network (5 lines)
		BrightMagenta, Reset,
		Dim, Reset,
		BrightMagenta, Reset,
		Dim, Reset,
		BrightMagenta, Reset,
		// CORTEX text (6 lines)
		BrightCyan+Bold, Reset,
		BrightCyan+Bold, Reset,
		BrightCyan+Bold, Reset,
		BrightCyan+Bold, Reset,
		BrightCyan+Bold, Reset,
		BrightCyan+Bold, Reset,
		// Bottom neural network (5 lines)
		BrightMagenta, Reset,
		Dim, Reset,
		BrightMagenta, Reset,
		Dim, Reset,
		BrightMagenta, Reset,
		// Tagline
		BrightYellow, Reset,
	)

	// Info line
	fmt.Printf("\n      %sv%s%s · %sCortex%s · %sAI Agent Orchestrator%s\n",
		Dim, version, Reset,
		BrightCyan+Bold, Reset,
		Dim, Reset,
	)
	fmt.Printf("              %s%s%s\n", Dim, displayPath, Reset)
	fmt.Println()
}

// PrintCompactBanner prints a minimal banner
func PrintCompactBanner(version string) {
	fmt.Printf("\n%s⧫ Cortex%s v%s\n\n", BrightCyan+Bold, Reset, version)
}

// PrintSessionInfo prints session information
func PrintSessionInfo(sessionID, outputDir string) {
	fmt.Printf("  %sSession:%s %s\n", Dim, Reset, sessionID)

	// Shorten the output path for display
	homeDir, _ := os.UserHomeDir()
	displayPath := outputDir
	if homeDir != "" && len(outputDir) > len(homeDir) && outputDir[:len(homeDir)] == homeDir {
		displayPath = "~" + outputDir[len(homeDir):]
	}
	fmt.Printf("  %sOutput:%s  %s\n", Dim, Reset, displayPath)
	fmt.Println()
}

// PrintDivider prints a horizontal divider
func PrintDivider() {
	fmt.Printf("%s────────────────────────────────────────%s\n", Dim, Reset)
}

// PrintExecutionPlan prints the execution plan with colors
func PrintExecutionPlan(tasks []TaskInfo) {
	fmt.Printf("%s%sExecution Plan:%s\n", Bold, BrightCyan, Reset)
	for i, task := range tasks {
		deps := ""
		if len(task.Dependencies) > 0 {
			deps = fmt.Sprintf(" %s[needs: %v]%s", Dim, task.Dependencies, Reset)
		}
		fmt.Printf("  %s%d.%s %s%s%s (%s%s%s → %s%s%s)%s\n",
			BrightYellow, i+1, Reset,
			Bold, task.Name, Reset,
			Cyan, task.Agent, Reset,
			Magenta, task.Tool, Reset,
			deps,
		)
		if task.Model != "" {
			fmt.Printf("     %smodel: %s%s\n", Dim, task.Model, Reset)
		}
	}
	fmt.Println()
}

// TaskInfo holds task display information
type TaskInfo struct {
	Name         string
	Agent        string
	Tool         string
	Model        string
	Dependencies []string
}

// PrintTaskStart prints task start message
func PrintTaskStart(index, total int, name, agent, tool, model string) {
	modelStr := ""
	if model != "" {
		modelStr = "/" + model
	}
	fmt.Printf("\n%s[%d/%d]%s %s%s%s\n",
		BrightYellow, index, total, Reset,
		Bold+BrightCyan, name, Reset,
	)
	fmt.Printf("  %sAgent:%s %s (%s%s%s%s)\n",
		Dim, Reset,
		agent,
		Magenta, tool, modelStr, Reset,
	)
}

// PrintTaskStatus prints task status
func PrintTaskStatus(status string, success bool, duration string) {
	var statusStr string
	if success {
		statusStr = fmt.Sprintf("%s✓ %s%s (%s)", BrightGreen, status, Reset, duration)
	} else {
		statusStr = fmt.Sprintf("%s✗ %s%s (%s)", BrightRed, status, Reset, duration)
	}
	fmt.Printf("  %sStatus:%s %s\n", Dim, Reset, statusStr)
}

// PrintTaskRunning prints running status
func PrintTaskRunning() {
	fmt.Printf("  %sStatus:%s %s⟳ Running...%s\n", Dim, Reset, BrightYellow, Reset)
}

// PrintSummary prints the final summary
func PrintSummary(success bool, outputDir string) {
	fmt.Println()
	PrintDivider()

	if success {
		fmt.Printf("%s✓ All tasks completed successfully!%s\n", BrightGreen+Bold, Reset)
	} else {
		fmt.Printf("%s✗ Workflow completed with failures%s\n", BrightRed+Bold, Reset)
	}

	// Shorten output path
	homeDir, _ := os.UserHomeDir()
	displayPath := outputDir
	if homeDir != "" && len(outputDir) > len(homeDir) && outputDir[:len(homeDir)] == homeDir {
		displayPath = "~" + outputDir[len(homeDir):]
	}
	fmt.Printf("%sResults saved to: %s%s\n", Dim, displayPath, Reset)
	fmt.Println()
}

// GetCortexHome returns the cortex home directory (~/.cortex)
func GetCortexHome() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".cortex"), nil
}

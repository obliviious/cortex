package ui

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

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

	// Print banner with clean design (Claude Orange theme)
	border := Orange + "  ╭────────────────────────────────────────────────────────╮" + Reset
	borderB := Orange + "  ╰────────────────────────────────────────────────────────╯" + Reset
	side := Orange + "  │" + Reset
	sideEnd := Orange + "│" + Reset

	fmt.Println(border)
	fmt.Println(side + "                                                          " + sideEnd)
	fmt.Printf("%s   %s ██████╗ ██████╗ ██████╗ ████████╗███████╗██╗  ██╗%s      %s\n", side, Orange+Bold, Reset, sideEnd)
	fmt.Printf("%s   %s██╔════╝██╔═══██╗██╔══██╗╚══██╔══╝██╔════╝╚██╗██╔╝%s      %s\n", side, Orange+Bold, Reset, sideEnd)
	fmt.Printf("%s   %s██║     ██║   ██║██████╔╝   ██║   █████╗   ╚███╔╝%s       %s\n", side, Orange+Bold, Reset, sideEnd)
	fmt.Printf("%s   %s██║     ██║   ██║██╔══██╗   ██║   ██╔══╝   ██╔██╗%s       %s\n", side, Orange+Bold, Reset, sideEnd)
	fmt.Printf("%s   %s╚██████╗╚██████╔╝██║  ██║   ██║   ███████╗██╔╝ ██╗%s      %s\n", side, Orange+Bold, Reset, sideEnd)
	fmt.Printf("%s   %s ╚═════╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚══════╝╚═╝  ╚═╝%s      %s\n", side, Orange+Bold, Reset, sideEnd)
	fmt.Println(side + "                                                          " + sideEnd)
	fmt.Printf("%s            %sAI Agent Orchestrator%s                      %s\n", side, Dim, Reset, sideEnd)
	fmt.Println(side + "                                                          " + sideEnd)
	fmt.Println(borderB)

	// Welcome message
	fmt.Printf("\n  %sWelcome, %s!%s\n", Bold+White, username, Reset)

	// Info line
	fmt.Printf("  %sv%s%s  %s%s%s\n\n",
		Dim, version, Reset,
		Dim, displayPath, Reset,
	)
}

// PrintCompactBanner prints a minimal banner
func PrintCompactBanner(version string) {
	fmt.Printf("\n%s◆ Cortex%s v%s\n\n", Orange+Bold, Reset, version)
}

// PrintSessionInfo prints session information
func PrintSessionInfo(sessionID, outputDir string) {
	// Shorten the output path for display
	homeDir, _ := os.UserHomeDir()
	displayPath := outputDir
	if homeDir != "" && len(outputDir) > len(homeDir) && outputDir[:len(homeDir)] == homeDir {
		displayPath = "~" + outputDir[len(homeDir):]
	}

	fmt.Printf("\n  %s○%s Session: %s\n", Orange, Reset, sessionID)
	fmt.Printf("    %s→%s Output: %s\n", Orange, Reset, displayPath)
	fmt.Println()
}

// PrintDivider prints a horizontal divider
func PrintDivider() {
	fmt.Printf("\n%s─────────────────────────────────────────────%s\n", Dim, Reset)
}

// PrintExecutionPlan prints the execution plan with colors
func PrintExecutionPlan(tasks []TaskInfo) {
	fmt.Printf("\n  %s%s◆ Execution Plan%s\n", Bold, Orange, Reset)
	fmt.Printf("  %s─────────────────%s\n\n", Dim, Reset)

	for i, task := range tasks {
		// Task card with box drawing
		fmt.Printf("  %s┌─%s %s%d%s %s│%s %s%s%s\n",
			Orange, Reset,
			Dim, i+1, Reset,
			Orange, Reset,
			Bold+Orange, task.Name, Reset,
		)

		// Dependencies if any
		if len(task.Dependencies) > 0 {
			fmt.Printf("  %s│%s  %s↳ needs: %v%s\n",
				Orange, Reset,
				Dim, task.Dependencies, Reset,
			)
		}

		// Agent info
		fmt.Printf("  %s│%s  %s◇%s %s%s%s\n",
			Orange, Reset,
			Dim, Reset,
			Orange, task.Agent, Reset,
		)

		// Tool and model
		toolInfo := task.Tool
		if task.Model != "" {
			toolInfo += " · " + task.Model
		}
		fmt.Printf("  %s│%s  %s◇%s %s%s%s\n",
			Orange, Reset,
			Dim, Reset,
			Dim, toolInfo, Reset,
		)

		fmt.Printf("  %s└───────────────────%s\n\n", Orange, Reset)
	}
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
		modelStr = " · " + model
	}
	fmt.Printf("\n%s┌─%s %s[%d/%d]%s %s%s%s\n",
		Orange, Reset,
		Dim, index, total, Reset,
		Bold+Orange, name, Reset,
	)
	fmt.Printf("%s│%s  %s%s%s %s· %s%s%s\n",
		Orange, Reset,
		Orange, agent, Reset,
		Dim, tool, modelStr, Reset,
	)
}

// PrintTaskStatus prints task status
func PrintTaskStatus(status string, success bool, duration string) {
	var statusStr string
	if success {
		statusStr = fmt.Sprintf("%s✓ %s%s %s(%s)%s", Green, status, Reset, Dim, duration, Reset)
	} else {
		statusStr = fmt.Sprintf("%s✗ %s%s %s(%s)%s", Red, status, Reset, Dim, duration, Reset)
	}
	fmt.Printf("%s└─%s %s\n", Orange, Reset, statusStr)
}

// PrintTaskRunning prints running status
func PrintTaskRunning() {
	fmt.Printf("%s│%s  %s● Running...%s\n", Orange, Reset, Orange, Reset)
}

// PrintTaskRunningWithHint prints running status with toggle hint
func PrintTaskRunningWithHint(showHint bool) {
	if showHint {
		fmt.Printf("%s│%s  %s● Running...%s  %s[Ctrl+O to expand]%s\n", Orange, Reset, Orange, Reset, Dim, Reset)
	} else {
		fmt.Printf("%s│%s  %s● Running...%s\n", Orange, Reset, Orange, Reset)
	}
}

// PrintSummary prints the final summary
func PrintSummary(success bool, outputDir string) {
	PrintDivider()

	if success {
		fmt.Printf("\n  %s✓ All tasks completed successfully%s\n", Green+Bold, Reset)
	} else {
		fmt.Printf("\n  %s✗ Workflow completed with failures%s\n", Red+Bold, Reset)
	}

	// Shorten output path
	homeDir, _ := os.UserHomeDir()
	displayPath := outputDir
	if homeDir != "" && len(outputDir) > len(homeDir) && outputDir[:len(homeDir)] == homeDir {
		displayPath = "~" + outputDir[len(homeDir):]
	}
	fmt.Printf("  %sResults: %s%s\n\n", Dim, displayPath, Reset)
}

// GetCortexHome returns the cortex home directory (~/.cortex)
func GetCortexHome() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".cortex"), nil
}

// PrintStreamStart prints a visual separator before streaming output
func PrintStreamStart() {
	fmt.Printf("%s│%s\n", Orange, Reset)
	fmt.Printf("%s│%s  %sAgent output:%s\n", Orange, Reset, Dim, Reset)
	fmt.Printf("%s│%s  %s─────────────%s\n", Orange, Reset, Dim, Reset)
}

// PrintStreamEnd prints a visual separator after streaming output
func PrintStreamEnd() {
	fmt.Printf("%s│%s  %s─────────────%s\n", Orange, Reset, Dim, Reset)
}

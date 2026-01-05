// Package claude implements the Agent interface for Claude Code CLI.
package claude

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/adityaraj/agentflow/internal/runtime"
	"github.com/adityaraj/agentflow/internal/ui"
)

// Adapter implements the Agent interface for claude-code CLI.
type Adapter struct {
	// executable is the name or path of the claude CLI binary
	executable string
	// streamLogs enables real-time output streaming
	streamLogs bool
}

// New creates a new Claude adapter.
// Uses "claude" as the default executable name.
func New() *Adapter {
	return &Adapter{
		executable: "claude",
		streamLogs: false,
	}
}

// NewWithExecutable creates a Claude adapter with a custom executable path.
func NewWithExecutable(executable string) *Adapter {
	return &Adapter{
		executable: executable,
		streamLogs: false,
	}
}

// NewWithOptions creates a Claude adapter with custom options.
func NewWithOptions(executable string, streamLogs bool) *Adapter {
	return &Adapter{
		executable: executable,
		streamLogs: streamLogs,
	}
}

// SetStreamLogs enables or disables real-time log streaming.
func (a *Adapter) SetStreamLogs(enabled bool) {
	a.streamLogs = enabled
}

// Run executes a task using the claude-code CLI.
func (a *Adapter) Run(ctx context.Context, task runtime.Task) (runtime.Result, error) {
	args := a.buildArgs(task)

	cmd := exec.CommandContext(ctx, a.executable, args...)

	var stdout, stderr bytes.Buffer
	var stripper *ui.MarkdownStripWriter

	if a.streamLogs {
		// Print visual separator before streaming
		ui.PrintStreamStart()
		// Use MarkdownStripWriter to strip markdown in real-time as output streams
		stripper = ui.NewMarkdownStripWriter(os.Stdout)
		cmd.Stdout = io.MultiWriter(stripper, &stdout)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	err := cmd.Run()

	if a.streamLogs {
		// Flush any remaining buffered content
		if stripper != nil {
			stripper.Flush()
		}
		// Print visual separator after streaming
		ui.PrintStreamEnd()
	}

	// Strip markdown from stored output as well
	cleanStdout := ui.StripMarkdown(stdout.String())

	result := runtime.Result{
		Stdout:   cleanStdout,
		Stderr:   stderr.String(),
		ExitCode: 0,
		Success:  true,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Success = false
		} else {
			// Command failed to start (e.g., binary not found)
			return result, fmt.Errorf("failed to execute claude: %w", err)
		}
	}

	return result, nil
}

// buildArgs constructs the command-line arguments for claude.
func (a *Adapter) buildArgs(task runtime.Task) []string {
	args := []string{
		"-p",                  // Print mode (non-interactive)
		"--output-format", "text", // Plain text output
	}

	// Add model if specified
	if task.Model != "" {
		args = append(args, "--model", task.Model)
	}

	// If writes are allowed, bypass permission checks
	if task.Write {
		args = append(args, "--dangerously-skip-permissions")
	}

	// Prompt must be the last positional argument
	args = append(args, task.Prompt)

	return args
}

// Check verifies that the claude CLI is available.
func (a *Adapter) Check() error {
	cmd := exec.Command(a.executable, "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("claude CLI not found or not executable: %w", err)
	}
	return nil
}

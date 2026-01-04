// Package opencode implements the Agent interface for OpenCode CLI.
package opencode

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/adityaraj/agentflow/internal/runtime"
)

// Adapter implements the Agent interface for opencode CLI.
type Adapter struct {
	// executable is the name or path of the opencode CLI binary
	executable string
	// streamLogs enables real-time output streaming
	streamLogs bool
}

// New creates a new OpenCode adapter.
// Uses "opencode" as the default executable name.
func New() *Adapter {
	return &Adapter{
		executable: "opencode",
		streamLogs: false,
	}
}

// NewWithExecutable creates an OpenCode adapter with a custom executable path.
func NewWithExecutable(executable string) *Adapter {
	return &Adapter{
		executable: executable,
		streamLogs: false,
	}
}

// SetStreamLogs enables or disables real-time log streaming.
func (a *Adapter) SetStreamLogs(enabled bool) {
	a.streamLogs = enabled
}

// Run executes a task using the opencode CLI.
func (a *Adapter) Run(ctx context.Context, task runtime.Task) (runtime.Result, error) {
	args := a.buildArgs(task)

	cmd := exec.CommandContext(ctx, a.executable, args...)

	var stdout, stderr bytes.Buffer

	if a.streamLogs {
		// Stream to terminal AND capture for result
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	err := cmd.Run()

	result := runtime.Result{
		Stdout:   stdout.String(),
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
			return result, fmt.Errorf("failed to execute opencode: %w", err)
		}
	}

	return result, nil
}

// buildArgs constructs the command-line arguments for opencode.
// Note: OpenCode CLI flags may vary - adjust as needed.
func (a *Adapter) buildArgs(task runtime.Task) []string {
	args := []string{
		"-p", task.Prompt, // Prompt flag (assumes similar to claude)
	}

	// Add model if specified
	if task.Model != "" {
		args = append(args, "--model", task.Model)
	}

	// OpenCode may have different permission flags
	// This is a placeholder - adjust based on actual CLI
	if task.Write {
		args = append(args, "--auto-approve")
	}

	return args
}

// Check verifies that the opencode CLI is available.
func (a *Adapter) Check() error {
	cmd := exec.Command(a.executable, "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("opencode CLI not found or not executable: %w", err)
	}
	return nil
}

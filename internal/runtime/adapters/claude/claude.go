// Package claude implements the Agent interface for Claude Code CLI.
package claude

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/adityaraj/agentflow/internal/runtime"
	"github.com/adityaraj/agentflow/internal/ui"
)

// defaultSystemPrompt provides formatting instructions for clean, readable output.
const defaultSystemPrompt = `Output formatting rules:
1. Use clear numbered points or bullet points
2. No emojis or decorative characters
3. Be concise and direct
4. Structure: Brief summary first, then details if needed
5. Keep responses focused and actionable`

// Adapter implements the Agent interface for claude-code CLI.
type Adapter struct {
	// executable is the name or path of the claude CLI binary
	executable string
	// streamLogs enables real-time output streaming
	streamLogs bool
	// systemPrompt overrides the default system prompt
	systemPrompt string
	// workdir specifies the working directory for Claude
	workdir string
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

// SetSystemPrompt sets a custom system prompt (empty uses default).
func (a *Adapter) SetSystemPrompt(prompt string) {
	a.systemPrompt = prompt
}

// SetWorkdir sets the working directory for Claude execution.
func (a *Adapter) SetWorkdir(dir string) {
	a.workdir = dir
}

// Run executes a task using the claude-code CLI.
func (a *Adapter) Run(ctx context.Context, task runtime.Task) (runtime.Result, error) {
	args := a.buildArgs(task)
	cmd := exec.CommandContext(ctx, a.executable, args...)

	// Streaming mode: use stream-json format and parse NDJSON in real-time
	if a.streamLogs {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return runtime.Result{}, fmt.Errorf("failed to create stdout pipe: %w", err)
		}

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Start(); err != nil {
			return runtime.Result{}, fmt.Errorf("failed to start claude: %w", err)
		}

		ui.PrintStreamStart()

		// Parse NDJSON and stream text content in real-time
		output := a.parseAndStreamNDJSON(stdout, os.Stdout)

		ui.PrintStreamEnd()

		err = cmd.Wait()

		result := runtime.Result{
			Stdout:   ui.StripMarkdown(output),
			Stderr:   stderr.String(),
			ExitCode: 0,
			Success:  true,
		}

		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				result.ExitCode = exitErr.ExitCode()
				result.Success = false
			} else {
				return result, fmt.Errorf("claude execution failed: %w", err)
			}
		}

		return result, nil
	}

	// Non-streaming mode: use buffered text output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

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
			return result, fmt.Errorf("failed to execute claude: %w", err)
		}
	}

	return result, nil
}

// buildArgs constructs the command-line arguments for claude.
func (a *Adapter) buildArgs(task runtime.Task) []string {
	args := []string{
		"-p", // SDK/headless mode
	}

	// Use stream-json for real-time streaming, text for buffered output
	if a.streamLogs {
		args = append(args, "--output-format", "stream-json")
	} else {
		args = append(args, "--output-format", "text")
	}

	// Add system prompt (use default if not overridden)
	systemPrompt := a.systemPrompt
	if systemPrompt == "" {
		systemPrompt = defaultSystemPrompt
	}
	args = append(args, "--system-prompt", systemPrompt)

	// Add working directory if specified (from task or adapter)
	workdir := task.Workdir
	if workdir == "" {
		workdir = a.workdir
	}
	if workdir != "" {
		args = append(args, "--cwd", workdir)
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

// streamMessage represents a single message in the NDJSON stream from Claude
type streamMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Result  string `json:"result"`
}

// parseAndStreamNDJSON reads NDJSON from reader, streams text content to writer,
// and returns the full accumulated output.
func (a *Adapter) parseAndStreamNDJSON(r io.Reader, w io.Writer) string {
	scanner := bufio.NewScanner(r)
	var fullOutput strings.Builder
	stripper := ui.NewMarkdownStripWriter(w)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var msg streamMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			// Not valid JSON, might be raw text - write as-is
			_, _ = stripper.Write([]byte(line + "\n"))
			fullOutput.WriteString(line + "\n")
			continue
		}

		// Stream content in real-time as it arrives
		if msg.Content != "" {
			_, _ = stripper.Write([]byte(msg.Content))
			fullOutput.WriteString(msg.Content)
		}

		// Capture final result (usually contains the complete response)
		if msg.Result != "" && fullOutput.Len() == 0 {
			// Only use result if we haven't accumulated content
			fullOutput.WriteString(msg.Result)
		}
	}

	_ = stripper.Flush()
	return fullOutput.String()
}

// Check verifies that the claude CLI is available.
func (a *Adapter) Check() error {
	cmd := exec.Command(a.executable, "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("claude CLI not found or not executable: %w", err)
	}
	return nil
}

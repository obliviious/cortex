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
	// Note: stream-json requires --verbose flag
	// --include-partial-messages enables real-time character-by-character streaming
	if a.streamLogs {
		args = append(args, "--output-format", "stream-json", "--verbose", "--include-partial-messages")
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
	Subtype string `json:"subtype"`
	Result  string `json:"result"`
	// For stream_event messages (real-time streaming with --include-partial-messages)
	Event *struct {
		Type  string `json:"type"`
		Index int    `json:"index"`
		// For content_block_start (tool_use)
		ContentBlock *struct {
			Type string `json:"type"`
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"content_block"`
		// For content_block_delta
		Delta *struct {
			Type        string `json:"type"`
			Text        string `json:"text"`
			PartialJSON string `json:"partial_json"`
		} `json:"delta"`
	} `json:"event"`
	// For assistant messages (final complete message)
	Message *struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"message"`
}

// toolInput represents common tool input parameters
type toolInput struct {
	FilePath    string `json:"file_path"`
	Path        string `json:"path"`
	Pattern     string `json:"pattern"`
	Command     string `json:"command"`
	Description string `json:"description"`
	Prompt      string `json:"prompt"`
	Query       string `json:"query"`
	URL         string `json:"url"`
	OldString   string `json:"old_string"`
	NewString   string `json:"new_string"`
}

// parseAndStreamNDJSON reads NDJSON from reader, streams text content to writer,
// and returns the full accumulated output.
func (a *Adapter) parseAndStreamNDJSON(r io.Reader, w io.Writer) string {
	scanner := bufio.NewScanner(r)
	// Increase scanner buffer for large JSON lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var fullOutput strings.Builder
	var currentTool string
	var toolInputJSON strings.Builder
	var toolDisplayed bool

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var msg streamMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			// Not valid JSON, might be raw text - write as-is
			_, _ = w.Write([]byte(line + "\n"))
			fullOutput.WriteString(line + "\n")
			continue
		}

		// Handle stream_event messages
		if msg.Type == "stream_event" && msg.Event != nil {
			// Tool use started
			if msg.Event.Type == "content_block_start" && msg.Event.ContentBlock != nil {
				if msg.Event.ContentBlock.Type == "tool_use" {
					currentTool = msg.Event.ContentBlock.Name
					toolInputJSON.Reset()
					toolDisplayed = false
				}
			}

			// Accumulate tool input JSON
			if msg.Event.Type == "content_block_delta" && msg.Event.Delta != nil {
				if msg.Event.Delta.Type == "input_json_delta" && msg.Event.Delta.PartialJSON != "" {
					toolInputJSON.WriteString(msg.Event.Delta.PartialJSON)

					// Try to display tool info early once we have enough data
					if !toolDisplayed && currentTool != "" {
						info := extractToolInfo(currentTool, toolInputJSON.String())
						if info != "" {
							// Tool info on new line with better formatting
							toolMsg := fmt.Sprintf("\n%s  ⚡ %s%s %s%s%s", ui.Orange, currentTool, ui.Reset, ui.Dim, info, ui.Reset)
							_, _ = w.Write([]byte(toolMsg))
							// Show waiting indicator for Task tool (sub-agent)
							if currentTool == "Task" {
								_, _ = w.Write([]byte(fmt.Sprintf(" %s(running sub-agent...)%s", ui.Dim, ui.Reset)))
							}
							_, _ = w.Write([]byte("\n"))
							toolDisplayed = true
						}
					}
				}

				// Text content delta (real-time streaming)
				if msg.Event.Delta.Type == "text_delta" && msg.Event.Delta.Text != "" {
					_, _ = w.Write([]byte(msg.Event.Delta.Text))
					fullOutput.WriteString(msg.Event.Delta.Text)
				}
			}

			// Tool use ended - show if not already displayed
			if msg.Event.Type == "content_block_stop" && currentTool != "" {
				if !toolDisplayed {
					info := extractToolInfo(currentTool, toolInputJSON.String())
					toolMsg := fmt.Sprintf("\n%s  ⚡ %s%s %s%s%s\n", ui.Orange, currentTool, ui.Reset, ui.Dim, info, ui.Reset)
					_, _ = w.Write([]byte(toolMsg))
				}
				currentTool = ""
				toolDisplayed = false
			}
		}

		// Handle final result (fallback if no streaming events received)
		if msg.Type == "result" && msg.Result != "" {
			// Only use result if we haven't accumulated content from stream events
			if fullOutput.Len() == 0 {
				_, _ = w.Write([]byte(msg.Result))
				fullOutput.WriteString(msg.Result)
			}
		}
	}

	return fullOutput.String()
}

// extractToolInfo extracts display info from tool input JSON
func extractToolInfo(toolName, jsonStr string) string {
	var input toolInput
	if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
		// Try to extract partial info from incomplete JSON
		return extractPartialInfo(toolName, jsonStr)
	}

	switch toolName {
	case "Read":
		if input.FilePath != "" {
			return shortenPath(input.FilePath)
		}
	case "Edit":
		if input.FilePath != "" {
			return shortenPath(input.FilePath)
		}
	case "Write":
		if input.FilePath != "" {
			return shortenPath(input.FilePath)
		}
	case "Glob":
		if input.Pattern != "" {
			return input.Pattern
		}
	case "Grep":
		if input.Pattern != "" {
			info := input.Pattern
			if input.Path != "" {
				info += " in " + shortenPath(input.Path)
			}
			return info
		}
	case "Bash":
		if input.Command != "" {
			cmd := input.Command
			if len(cmd) > 50 {
				cmd = cmd[:50] + "..."
			}
			return cmd
		}
	case "WebSearch":
		if input.Query != "" {
			return input.Query
		}
	case "WebFetch":
		if input.URL != "" {
			return input.URL
		}
	case "Task":
		if input.Description != "" {
			return input.Description
		}
	case "LSP":
		if input.FilePath != "" {
			return shortenPath(input.FilePath)
		}
	}
	return ""
}

// extractPartialInfo tries to extract info from incomplete JSON
func extractPartialInfo(toolName, jsonStr string) string {
	// Look for common patterns in partial JSON
	patterns := map[string]string{
		"file_path\":\"": "file_path",
		"path\":\"":      "path",
		"pattern\":\"":   "pattern",
		"command\":\"":   "command",
		"query\":\"":     "query",
	}

	for pattern, _ := range patterns {
		if idx := strings.Index(jsonStr, pattern); idx >= 0 {
			start := idx + len(pattern)
			end := strings.Index(jsonStr[start:], "\"")
			if end > 0 && end < 100 {
				value := jsonStr[start : start+end]
				return shortenPath(value)
			}
		}
	}
	return ""
}

// shortenPath shortens a file path for display
func shortenPath(path string) string {
	// Remove home directory prefix
	if home, err := os.UserHomeDir(); err == nil {
		if strings.HasPrefix(path, home) {
			path = "~" + path[len(home):]
		}
	}
	// If still too long, show just the filename
	if len(path) > 60 {
		parts := strings.Split(path, "/")
		if len(parts) > 2 {
			path = ".../" + parts[len(parts)-2] + "/" + parts[len(parts)-1]
		}
	}
	return path
}

// Check verifies that the claude CLI is available.
func (a *Adapter) Check() error {
	cmd := exec.Command(a.executable, "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("claude CLI not found or not executable: %w", err)
	}
	return nil
}

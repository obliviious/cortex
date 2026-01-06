// Package shell implements the Agent interface for running shell commands.
package shell

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/adityaraj/agentflow/internal/runtime"
	"github.com/adityaraj/agentflow/internal/ui"
)

// Adapter implements the Agent interface for shell command execution.
type Adapter struct {
	// shell is the shell to use (default: /bin/sh)
	shell string
	// streamLogs enables real-time output streaming
	streamLogs bool
	// workdir specifies the working directory for commands
	workdir string
}

// New creates a new Shell adapter with default settings.
func New() *Adapter {
	return &Adapter{
		shell:      "/bin/sh",
		streamLogs: false,
	}
}

// NewWithShell creates a Shell adapter with a custom shell.
func NewWithShell(shell string) *Adapter {
	return &Adapter{
		shell:      shell,
		streamLogs: false,
	}
}

// SetStreamLogs enables or disables real-time log streaming.
func (a *Adapter) SetStreamLogs(enabled bool) {
	a.streamLogs = enabled
}

// SetWorkdir sets the working directory for command execution.
func (a *Adapter) SetWorkdir(dir string) {
	a.workdir = dir
}

// Run executes a shell command.
// For shell agents, task.Prompt contains the command to execute.
func (a *Adapter) Run(ctx context.Context, task runtime.Task) (runtime.Result, error) {
	command := task.Prompt
	if command == "" {
		return runtime.Result{}, fmt.Errorf("no command specified for shell task")
	}

	// Build command with shell
	cmd := exec.CommandContext(ctx, a.shell, "-c", command)

	// Set working directory
	workdir := task.Workdir
	if workdir == "" {
		workdir = a.workdir
	}
	if workdir != "" {
		cmd.Dir = workdir
	}

	// Streaming mode: show output in real-time
	if a.streamLogs {
		return a.runStreaming(cmd, command)
	}

	// Non-streaming mode: capture output
	return a.runBuffered(cmd)
}

// runStreaming executes the command with real-time output streaming.
func (a *Adapter) runStreaming(cmd *exec.Cmd, command string) (runtime.Result, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return runtime.Result{}, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return runtime.Result{}, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return runtime.Result{}, fmt.Errorf("failed to start command: %w", err)
	}

	// Print command being executed
	ui.PrintStreamStart()
	displayCmd := command
	if len(displayCmd) > 80 {
		displayCmd = displayCmd[:80] + "..."
	}
	fmt.Printf("%s  $ %s%s\n", ui.Dim, displayCmd, ui.Reset)

	// Stream stdout and stderr concurrently
	var stdoutBuf, stderrBuf strings.Builder
	done := make(chan struct{}, 2)

	go func() {
		a.streamOutput(stdout, os.Stdout, &stdoutBuf)
		done <- struct{}{}
	}()

	go func() {
		a.streamOutput(stderr, os.Stderr, &stderrBuf)
		done <- struct{}{}
	}()

	// Wait for both streams to finish
	<-done
	<-done

	ui.PrintStreamEnd()

	err = cmd.Wait()

	result := runtime.Result{
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		ExitCode: 0,
		Success:  true,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Success = false
		} else {
			return result, fmt.Errorf("command execution failed: %w", err)
		}
	}

	return result, nil
}

// runBuffered executes the command and captures all output.
func (a *Adapter) runBuffered(cmd *exec.Cmd) (runtime.Result, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

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
			return result, fmt.Errorf("failed to execute command: %w", err)
		}
	}

	return result, nil
}

// streamOutput reads from reader and writes to both writer and buffer.
func (a *Adapter) streamOutput(r io.Reader, w io.Writer, buf *strings.Builder) {
	scanner := bufio.NewScanner(r)
	// Increase buffer size for long lines
	scanBuf := make([]byte, 0, 64*1024)
	scanner.Buffer(scanBuf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		buf.WriteString(line)
		buf.WriteString("\n")
		fmt.Fprintln(w, line)
	}
}

// Check verifies that the shell is available.
func (a *Adapter) Check() error {
	cmd := exec.Command(a.shell, "-c", "echo ok")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("shell %s not available: %w", a.shell, err)
	}
	return nil
}

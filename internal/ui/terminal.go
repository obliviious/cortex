package ui

import (
	"os"
	"sync"

	"golang.org/x/term"
)

// OutputMode defines how output is displayed
type OutputMode int

const (
	// OutputCollapsed shows only summary (first few lines)
	OutputCollapsed OutputMode = iota
	// OutputExpanded shows full streaming output
	OutputExpanded
)

// TerminalController manages interactive terminal features
type TerminalController struct {
	mu         sync.RWMutex
	mode       OutputMode
	maxSummary int // Max lines to show in collapsed mode
	isRawMode  bool
	oldState   *term.State
	toggleChan chan struct{}
	stopChan   chan struct{}
	onToggle   func(OutputMode)
}

// NewTerminalController creates a new terminal controller
func NewTerminalController() *TerminalController {
	return &TerminalController{
		mode:       OutputCollapsed,
		maxSummary: 5,
		toggleChan: make(chan struct{}, 1),
		stopChan:   make(chan struct{}),
	}
}

// SetToggleCallback sets the function to call when output mode is toggled
func (t *TerminalController) SetToggleCallback(fn func(OutputMode)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onToggle = fn
}

// Mode returns the current output mode
func (t *TerminalController) Mode() OutputMode {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.mode
}

// IsExpanded returns true if output is in expanded mode
func (t *TerminalController) IsExpanded() bool {
	return t.Mode() == OutputExpanded
}

// Toggle switches between collapsed and expanded mode
func (t *TerminalController) Toggle() {
	t.mu.Lock()
	if t.mode == OutputCollapsed {
		t.mode = OutputExpanded
	} else {
		t.mode = OutputCollapsed
	}
	callback := t.onToggle
	mode := t.mode
	t.mu.Unlock()

	if callback != nil {
		callback(mode)
	}
}

// SetMode sets the output mode
func (t *TerminalController) SetMode(mode OutputMode) {
	t.mu.Lock()
	t.mode = mode
	t.mu.Unlock()
}

// Start begins listening for keyboard input (Ctrl+O to toggle)
func (t *TerminalController) Start() error {
	// Check if stdin is a terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return nil // Not a terminal, skip raw mode
	}

	// Save current terminal state and switch to raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil // Can't set raw mode, continue without toggle
	}

	t.mu.Lock()
	t.oldState = oldState
	t.isRawMode = true
	t.mu.Unlock()

	// Start key listener goroutine
	go t.listenKeys()

	return nil
}

// Stop restores the terminal to its original state
func (t *TerminalController) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isRawMode && t.oldState != nil {
		close(t.stopChan)
		_ = term.Restore(int(os.Stdin.Fd()), t.oldState)
		t.isRawMode = false
	}
}

// listenKeys reads keyboard input and handles Ctrl+O
func (t *TerminalController) listenKeys() {
	buf := make([]byte, 1)
	for {
		select {
		case <-t.stopChan:
			return
		default:
			n, err := os.Stdin.Read(buf)
			if err != nil || n == 0 {
				continue
			}

			// Ctrl+O is ASCII 15
			if buf[0] == 15 {
				t.Toggle()
			}

			// Ctrl+C is ASCII 3 - propagate interrupt
			if buf[0] == 3 {
				// Send SIGINT to self
				p, _ := os.FindProcess(os.Getpid())
				if p != nil {
					_ = p.Signal(os.Interrupt)
				}
			}
		}
	}
}

// BufferedWriter wraps output to support toggle functionality
type BufferedWriter struct {
	controller *TerminalController
	buffer     []byte
	mu         sync.Mutex
	lineCount  int
}

// NewBufferedWriter creates a writer that buffers output for toggle support
func NewBufferedWriter(ctrl *TerminalController) *BufferedWriter {
	return &BufferedWriter{
		controller: ctrl,
	}
}

// Write implements io.Writer
func (b *BufferedWriter) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Always store in buffer
	b.buffer = append(b.buffer, p...)

	// Count newlines
	for _, c := range p {
		if c == '\n' {
			b.lineCount++
		}
	}

	// Write based on mode
	if b.controller.IsExpanded() {
		return os.Stdout.Write(p)
	}

	// In collapsed mode, only write if under limit
	if b.lineCount <= b.controller.maxSummary {
		return os.Stdout.Write(p)
	}

	return len(p), nil
}

// GetBuffer returns the full buffered output
func (b *BufferedWriter) GetBuffer() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]byte{}, b.buffer...)
}

// Reset clears the buffer
func (b *BufferedWriter) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buffer = nil
	b.lineCount = 0
}

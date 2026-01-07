package ui

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// SelectableItem represents an item that can be selected
type SelectableItem struct {
	Label       string
	Description string
	Value       interface{}
}

// InteractiveSelector provides arrow-key based selection
type InteractiveSelector struct {
	items    []SelectableItem
	selected int
	title    string
	rendered bool // tracks if we've rendered before (to avoid clearing on first render)
}

// NewInteractiveSelector creates a new selector
func NewInteractiveSelector(title string, items []SelectableItem) *InteractiveSelector {
	return &InteractiveSelector{
		items:    items,
		selected: 0,
		title:    title,
	}
}

// Run displays the selector and returns the selected item index, or -1 if cancelled
func (s *InteractiveSelector) Run() int {
	if len(s.items) == 0 {
		return -1
	}

	// Check if stdin is a terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		// Non-interactive mode, just return first item
		return 0
	}

	// Save terminal state and switch to raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return 0
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	// Hide cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h") // Show cursor on exit

	// Initial render
	s.render()

	// Read keys
	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return -1
		}

		if n == 1 {
			switch buf[0] {
			case 13: // Enter
				s.clearDisplay()
				return s.selected
			case 3, 27: // Ctrl+C or Escape (single byte)
				if n == 1 && buf[0] == 27 {
					// Could be escape or start of arrow key sequence
					// Try to read more
					_, _ = os.Stdin.Read(buf[1:])
				}
				if n == 1 && buf[0] == 3 {
					s.clearDisplay()
					return -1
				}
			case 'j', 'J': // Vim down
				s.moveDown()
				s.render()
			case 'k', 'K': // Vim up
				s.moveUp()
				s.render()
			case 'q', 'Q': // Quit
				s.clearDisplay()
				return -1
			}
		}

		if n >= 3 {
			// Arrow key sequences
			if buf[0] == 27 && buf[1] == 91 { // ESC [
				switch buf[2] {
				case 65: // Up arrow
					s.moveUp()
					s.render()
				case 66: // Down arrow
					s.moveDown()
					s.render()
				}
			}
		} else if n == 1 && buf[0] == 27 {
			// Escape key pressed alone - read remaining bytes if any
			remaining := make([]byte, 2)
			_, _ = os.Stdin.Read(remaining)
			if remaining[0] == 91 {
				switch remaining[1] {
				case 65: // Up arrow
					s.moveUp()
					s.render()
				case 66: // Down arrow
					s.moveDown()
					s.render()
				}
			} else if remaining[0] == 0 {
				// Just escape, cancel
				s.clearDisplay()
				return -1
			}
		}
	}
}

func (s *InteractiveSelector) moveUp() {
	if s.selected > 0 {
		s.selected--
	}
}

func (s *InteractiveSelector) moveDown() {
	if s.selected < len(s.items)-1 {
		s.selected++
	}
}

func (s *InteractiveSelector) render() {
	// Only clear if we've rendered before (avoid clearing content above selector on first render)
	if s.rendered {
		s.clearDisplay()
	}
	s.rendered = true

	// Print title
	fmt.Printf("\r%s%s%s %s(↑/↓ to navigate, Enter to select, q to quit)%s\n",
		Bold, Orange, s.title, Dim, Reset)
	fmt.Printf("\r%s%s%s\n", Dim, strings.Repeat("─", 50), Reset)

	// Print items
	for i, item := range s.items {
		if i == s.selected {
			fmt.Printf("\r  %s▸%s %s%s%s\n", Orange, Reset, Bold, item.Label, Reset)
			if item.Description != "" {
				fmt.Printf("\r    %s%s%s\n", Dim, item.Description, Reset)
			}
		} else {
			fmt.Printf("\r    %s%s\n", item.Label, Reset)
			if item.Description != "" {
				fmt.Printf("\r    %s%s%s\n", Dim, item.Description, Reset)
			}
		}
	}
}

func (s *InteractiveSelector) clearDisplay() {
	// Calculate number of lines to clear (title + separator + items with descriptions)
	lines := 2 // title + separator
	for _, item := range s.items {
		lines++ // item label
		if item.Description != "" {
			lines++ // description
		}
	}

	// Move up and clear each line
	for i := 0; i < lines; i++ {
		fmt.Print("\033[A") // Move up
		fmt.Print("\033[K") // Clear line
	}
	fmt.Print("\r") // Return to beginning
}

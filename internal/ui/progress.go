package ui

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// SpinnerFrames contains the animation frames for the spinner
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner provides an animated spinner for terminal output
type Spinner struct {
	frames   []string
	current  int
	message  string
	interval time.Duration
	stop     chan struct{}
	done     chan struct{}
	mu       sync.Mutex
	running  bool
}

// NewSpinner creates a new spinner with default settings
func NewSpinner() *Spinner {
	return &Spinner{
		frames:   SpinnerFrames,
		interval: 80 * time.Millisecond,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
}

// Start begins the spinner animation with the given message
func (s *Spinner) Start(message string) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.message = message
	s.stop = make(chan struct{})
	s.done = make(chan struct{})
	s.mu.Unlock()

	go func() {
		defer close(s.done)
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-s.stop:
				// Clear the spinner line
				fmt.Print("\r\033[K")
				return
			case <-ticker.C:
				s.mu.Lock()
				frame := s.frames[s.current%len(s.frames)]
				msg := s.message
				s.current++
				s.mu.Unlock()

				fmt.Printf("\r%s%s%s %s", Orange, frame, Reset, msg)
			}
		}
	}()
}

// Update changes the spinner message
func (s *Spinner) Update(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}

// Stop stops the spinner animation
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stop)
	<-s.done
}

// ProgressBar represents a progress bar
type ProgressBar struct {
	total     int
	current   int32
	width     int
	startTime time.Time
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int) *ProgressBar {
	return &ProgressBar{
		total:     total,
		width:     20,
		startTime: time.Now(),
	}
}

// Increment increases the progress by 1
func (p *ProgressBar) Increment() {
	atomic.AddInt32(&p.current, 1)
}

// Set sets the current progress value
func (p *ProgressBar) Set(value int) {
	atomic.StoreInt32(&p.current, int32(value))
}

// Render returns the progress bar string
func (p *ProgressBar) Render() string {
	current := int(atomic.LoadInt32(&p.current))
	if p.total == 0 {
		return ""
	}

	ratio := float64(current) / float64(p.total)
	if ratio > 1 {
		ratio = 1
	}

	filled := int(ratio * float64(p.width))
	empty := p.width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	percent := int(ratio * 100)
	elapsed := time.Since(p.startTime).Round(time.Second)

	return fmt.Sprintf("[%s] %d/%d (%d%%) %s", bar, current, p.total, percent, elapsed)
}

// RenderCompact returns a compact progress representation
func (p *ProgressBar) RenderCompact() string {
	current := int(atomic.LoadInt32(&p.current))
	elapsed := time.Since(p.startTime).Round(time.Second)
	return fmt.Sprintf("%d/%d (%s)", current, p.total, elapsed)
}

// ProgressTracker combines spinner and progress tracking
type ProgressTracker struct {
	totalTasks     int
	completedTasks atomic.Int32
	currentLevel   int
	totalLevels    int
	currentTask    string
	startTime      time.Time
	spinner        *Spinner
	enabled        bool
	mu             sync.Mutex
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(totalTasks, totalLevels int) *ProgressTracker {
	return &ProgressTracker{
		totalTasks:  totalTasks,
		totalLevels: totalLevels,
		startTime:   time.Now(),
		spinner:     NewSpinner(),
		enabled:     true,
	}
}

// SetEnabled enables or disables progress display
func (p *ProgressTracker) SetEnabled(enabled bool) {
	p.mu.Lock()
	p.enabled = enabled
	p.mu.Unlock()
}

// StartTask marks a task as started
func (p *ProgressTracker) StartTask(taskName string, level int) {
	p.mu.Lock()
	p.currentTask = taskName
	p.currentLevel = level
	enabled := p.enabled
	p.mu.Unlock()

	if enabled {
		msg := p.formatProgress()
		p.spinner.Start(msg)
	}
}

// CompleteTask marks a task as completed
func (p *ProgressTracker) CompleteTask() {
	p.completedTasks.Add(1)
	p.spinner.Stop()
}

// Stop stops the progress tracker
func (p *ProgressTracker) Stop() {
	p.spinner.Stop()
}

// formatProgress formats the progress message
func (p *ProgressTracker) formatProgress() string {
	completed := int(p.completedTasks.Load())
	elapsed := time.Since(p.startTime).Round(time.Second)

	// Progress bar
	ratio := float64(completed) / float64(p.totalTasks)
	if ratio > 1 {
		ratio = 1
	}
	barWidth := 16
	filled := int(ratio * float64(barWidth))
	empty := barWidth - filled
	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

	return fmt.Sprintf("%s%s%s [%s] %d/%d (Level %d/%d) %s",
		Bold, p.currentTask, Reset,
		bar,
		completed, p.totalTasks,
		p.currentLevel+1, p.totalLevels,
		elapsed,
	)
}

// RenderProgress renders a static progress line
func RenderProgress(completed, total int, taskName string, elapsed time.Duration) string {
	ratio := float64(completed) / float64(total)
	if ratio > 1 {
		ratio = 1
	}

	barWidth := 16
	filled := int(ratio * float64(barWidth))
	empty := barWidth - filled
	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

	return fmt.Sprintf("%s%s%s %s%s [%s] %d/%d%s %s(%s)%s",
		Orange, SpinnerFrames[0], Reset,
		Bold, taskName, bar,
		completed, total, Reset,
		Dim, elapsed.Round(time.Second), Reset,
	)
}

// RenderProgressBar returns just the progress bar portion
func RenderProgressBar(completed, total int) string {
	if total == 0 {
		return ""
	}

	ratio := float64(completed) / float64(total)
	if ratio > 1 {
		ratio = 1
	}

	barWidth := 16
	filled := int(ratio * float64(barWidth))
	empty := barWidth - filled

	return fmt.Sprintf("[%s%s%s%s%s]",
		Cyan, strings.Repeat("█", filled), Reset+Dim, strings.Repeat("░", empty), Reset)
}

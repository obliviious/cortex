// Package webhook provides webhook notification support.
package webhook

import (
	"time"
)

// Event types for webhook notifications.
const (
	EventRunStart     = "run_start"
	EventRunComplete  = "run_complete"
	EventTaskStart    = "task_start"
	EventTaskComplete = "task_complete"
	EventTaskFailed   = "task_failed"
)

// Event represents a webhook event payload.
type Event struct {
	Type      string     `json:"event"`
	Timestamp time.Time  `json:"timestamp"`
	RunID     string     `json:"run_id"`
	Project   string     `json:"project"`
	Task      *TaskEvent `json:"task,omitempty"`
	Run       *RunEvent  `json:"run,omitempty"`
}

// TaskEvent contains task-specific event data.
type TaskEvent struct {
	Name     string `json:"name"`
	Agent    string `json:"agent"`
	Tool     string `json:"tool"`
	Model    string `json:"model,omitempty"`
	Duration string `json:"duration,omitempty"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}

// RunEvent contains run-specific event data.
type RunEvent struct {
	TaskCount int    `json:"task_count"`
	Duration  string `json:"duration"`
	Success   bool   `json:"success"`
}

// NewRunStartEvent creates a run_start event.
func NewRunStartEvent(runID, project string) Event {
	return Event{
		Type:      EventRunStart,
		Timestamp: time.Now(),
		RunID:     runID,
		Project:   project,
	}
}

// NewRunCompleteEvent creates a run_complete event.
func NewRunCompleteEvent(runID, project string, taskCount int, duration time.Duration, success bool) Event {
	return Event{
		Type:      EventRunComplete,
		Timestamp: time.Now(),
		RunID:     runID,
		Project:   project,
		Run: &RunEvent{
			TaskCount: taskCount,
			Duration:  duration.Round(time.Millisecond * 100).String(),
			Success:   success,
		},
	}
}

// NewTaskStartEvent creates a task_start event.
func NewTaskStartEvent(runID, project, taskName, agent, tool, model string) Event {
	return Event{
		Type:      EventTaskStart,
		Timestamp: time.Now(),
		RunID:     runID,
		Project:   project,
		Task: &TaskEvent{
			Name:  taskName,
			Agent: agent,
			Tool:  tool,
			Model: model,
		},
	}
}

// NewTaskCompleteEvent creates a task_complete event.
func NewTaskCompleteEvent(runID, project, taskName, agent, tool, model, duration string, success bool) Event {
	return Event{
		Type:      EventTaskComplete,
		Timestamp: time.Now(),
		RunID:     runID,
		Project:   project,
		Task: &TaskEvent{
			Name:     taskName,
			Agent:    agent,
			Tool:     tool,
			Model:    model,
			Duration: duration,
			Success:  success,
		},
	}
}

// NewTaskFailedEvent creates a task_failed event.
func NewTaskFailedEvent(runID, project, taskName, agent, tool, model, duration, errMsg string) Event {
	return Event{
		Type:      EventTaskFailed,
		Timestamp: time.Now(),
		RunID:     runID,
		Project:   project,
		Task: &TaskEvent{
			Name:     taskName,
			Agent:    agent,
			Tool:     tool,
			Model:    model,
			Duration: duration,
			Success:  false,
			Error:    errMsg,
		},
	}
}

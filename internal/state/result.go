// Package state handles run results and persistence.
package state

import (
	"time"
)

// TaskResult represents the result of executing a single task.
type TaskResult struct {
	TaskName  string    `json:"task_name"`
	Agent     string    `json:"agent"`
	Tool      string    `json:"tool"`
	Model     string    `json:"model,omitempty"`
	Prompt    string    `json:"prompt"`
	Stdout    string    `json:"stdout"`
	Stderr    string    `json:"stderr,omitempty"`
	Success   bool      `json:"success"`
	ExitCode  int       `json:"exit_code"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  string    `json:"duration"` // Human-readable duration
}

// RunResult represents the complete result of an agentflow run.
type RunResult struct {
	RunID     string       `json:"run_id"`
	StartTime time.Time    `json:"start_time"`
	EndTime   time.Time    `json:"end_time"`
	Success   bool         `json:"success"`
	Tasks     []TaskResult `json:"tasks"`
}

// NewTaskResult creates a new TaskResult with timing started.
func NewTaskResult(taskName, agent, tool, model, prompt string) *TaskResult {
	return &TaskResult{
		TaskName:  taskName,
		Agent:     agent,
		Tool:      tool,
		Model:     model,
		Prompt:    prompt,
		StartTime: time.Now(),
	}
}

// Complete marks the task as completed with the given result.
func (r *TaskResult) Complete(stdout, stderr string, exitCode int, success bool) {
	r.Stdout = stdout
	r.Stderr = stderr
	r.ExitCode = exitCode
	r.Success = success
	r.EndTime = time.Now()
	r.Duration = r.EndTime.Sub(r.StartTime).Round(time.Millisecond * 100).String()
}

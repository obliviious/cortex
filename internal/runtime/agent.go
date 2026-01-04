// Package runtime handles agent execution and task orchestration.
package runtime

import (
	"context"
)

// Task represents a task to be executed by an agent.
type Task struct {
	Name   string // Task name
	Agent  string // Agent name
	Tool   string // CLI tool (claude-code, opencode)
	Model  string // Model identifier
	Prompt string // Prompt text (already expanded with template variables)
	Write  bool   // Allow file writes
}

// Result represents the result of executing a task.
type Result struct {
	Stdout   string // Standard output from the agent
	Stderr   string // Standard error from the agent
	ExitCode int    // Exit code (0 = success)
	Success  bool   // Whether the task succeeded
}

// Agent is the interface that all agent adapters must implement.
type Agent interface {
	// Run executes a task and returns the result.
	// The context can be used for cancellation.
	Run(ctx context.Context, task Task) (Result, error)
}

// AgentRegistry holds available agent adapters by tool name.
type AgentRegistry struct {
	adapters map[string]Agent
}

// NewAgentRegistry creates a new registry with no adapters.
func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		adapters: make(map[string]Agent),
	}
}

// Register adds an agent adapter for the given tool name.
func (r *AgentRegistry) Register(tool string, agent Agent) {
	r.adapters[tool] = agent
}

// Get returns the agent adapter for the given tool name.
// Returns nil if no adapter is registered.
func (r *AgentRegistry) Get(tool string) Agent {
	return r.adapters[tool]
}

// Has checks if an adapter is registered for the given tool.
func (r *AgentRegistry) Has(tool string) bool {
	_, ok := r.adapters[tool]
	return ok
}

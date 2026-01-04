package config

import (
	"fmt"
	"strings"
)

// ConfigError represents a configuration error with location information.
type ConfigError struct {
	File    string // File path
	Line    int    // Line number (1-based)
	Column  int    // Column number (1-based, 0 if unknown)
	Message string // Error message
	Hint    string // Optional hint for fixing the error
}

// Error implements the error interface.
func (e *ConfigError) Error() string {
	var sb strings.Builder

	// Location
	if e.File != "" {
		sb.WriteString(e.File)
		if e.Line > 0 {
			sb.WriteString(fmt.Sprintf(":%d", e.Line))
			if e.Column > 0 {
				sb.WriteString(fmt.Sprintf(":%d", e.Column))
			}
		}
		sb.WriteString(": ")
	}

	// Message
	sb.WriteString(e.Message)

	// Hint
	if e.Hint != "" {
		sb.WriteString(fmt.Sprintf("\n  Hint: %s", e.Hint))
	}

	return sb.String()
}

// ConfigErrors represents multiple configuration errors.
type ConfigErrors struct {
	Errors []*ConfigError
}

// Error implements the error interface.
func (e *ConfigErrors) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("found %d configuration errors:\n", len(e.Errors)))
	for i, err := range e.Errors {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err.Error()))
	}
	return sb.String()
}

// Add adds an error to the collection.
func (e *ConfigErrors) Add(err *ConfigError) {
	e.Errors = append(e.Errors, err)
}

// HasErrors returns true if there are any errors.
func (e *ConfigErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// NewConfigError creates a new configuration error.
func NewConfigError(file string, line int, message string) *ConfigError {
	return &ConfigError{
		File:    file,
		Line:    line,
		Message: message,
	}
}

// NewConfigErrorWithHint creates a new configuration error with a hint.
func NewConfigErrorWithHint(file string, line int, message, hint string) *ConfigError {
	return &ConfigError{
		File:    file,
		Line:    line,
		Message: message,
		Hint:    hint,
	}
}

// Common error constructors

// ErrUndefinedAgent creates an error for an undefined agent reference.
func ErrUndefinedAgent(file string, line int, taskName, agentName string, availableAgents []string) *ConfigError {
	hint := ""
	if len(availableAgents) > 0 {
		hint = fmt.Sprintf("Available agents: %s", strings.Join(availableAgents, ", "))
	}
	return &ConfigError{
		File:    file,
		Line:    line,
		Message: fmt.Sprintf("task %q references undefined agent %q", taskName, agentName),
		Hint:    hint,
	}
}

// ErrUnsupportedTool creates an error for an unsupported tool.
func ErrUnsupportedTool(file string, line int, agentName, tool string) *ConfigError {
	return &ConfigError{
		File:    file,
		Line:    line,
		Message: fmt.Sprintf("agent %q uses unsupported tool %q", agentName, tool),
		Hint:    fmt.Sprintf("Supported tools: %s", strings.Join(SupportedTools, ", ")),
	}
}

// ErrUndefinedDependency creates an error for an undefined task dependency.
func ErrUndefinedDependency(file string, line int, taskName, depName string, availableTasks []string) *ConfigError {
	hint := ""
	if len(availableTasks) > 0 {
		hint = fmt.Sprintf("Available tasks: %s", strings.Join(availableTasks, ", "))
	}
	return &ConfigError{
		File:    file,
		Line:    line,
		Message: fmt.Sprintf("task %q depends on undefined task %q", taskName, depName),
		Hint:    hint,
	}
}

// ErrCircularDependency creates an error for circular dependencies.
func ErrCircularDependency(file string, cycle []string) *ConfigError {
	return &ConfigError{
		File:    file,
		Message: fmt.Sprintf("circular dependency detected: %s", strings.Join(cycle, " -> ")),
		Hint:    "Remove one of the dependencies to break the cycle",
	}
}

// ErrNoPrompt creates an error for a task with no prompt defined.
func ErrNoPrompt(file string, line int, taskName string) *ConfigError {
	return &ConfigError{
		File:    file,
		Line:    line,
		Message: fmt.Sprintf("task %q has no prompt defined", taskName),
		Hint:    "Add either 'prompt:' with inline text or 'prompt_file:' with a file path",
	}
}

// ErrPromptFileNotFound creates an error for a missing prompt file.
func ErrPromptFileNotFound(file string, line int, taskName, promptFile string) *ConfigError {
	return &ConfigError{
		File:    file,
		Line:    line,
		Message: fmt.Sprintf("task %q references prompt file that doesn't exist: %s", taskName, promptFile),
		Hint:    "Check the file path and ensure the file exists",
	}
}

// ErrNoAgents creates an error for config with no agents defined.
func ErrNoAgents(file string) *ConfigError {
	return &ConfigError{
		File:    file,
		Message: "no agents defined",
		Hint:    "Add an 'agents:' section with at least one agent",
	}
}

// ErrNoTasks creates an error for config with no tasks defined.
func ErrNoTasks(file string) *ConfigError {
	return &ConfigError{
		File:    file,
		Message: "no tasks defined",
		Hint:    "Add a 'tasks:' section with at least one task",
	}
}

// ErrEmptyAgentName creates an error for an empty agent name.
func ErrEmptyAgentName(file string, line int) *ConfigError {
	return &ConfigError{
		File:    file,
		Line:    line,
		Message: "agent name cannot be empty",
		Hint:    "Provide a valid agent name",
	}
}

// ErrEmptyTaskName creates an error for an empty task name.
func ErrEmptyTaskName(file string, line int) *ConfigError {
	return &ConfigError{
		File:    file,
		Line:    line,
		Message: "task name cannot be empty",
		Hint:    "Provide a valid task name",
	}
}

// ErrYAMLParse creates an error for YAML parsing failures.
func ErrYAMLParse(file string, line int, details string) *ConfigError {
	return &ConfigError{
		File:    file,
		Line:    line,
		Message: fmt.Sprintf("YAML parse error: %s", details),
		Hint:    "Check YAML syntax - ensure proper indentation and formatting",
	}
}

// ErrSelfDependency creates an error for a task that depends on itself.
func ErrSelfDependency(file string, line int, taskName string) *ConfigError {
	return &ConfigError{
		File:    file,
		Line:    line,
		Message: fmt.Sprintf("task %q cannot depend on itself", taskName),
		Hint:    "Remove the self-reference from the 'needs' list",
	}
}

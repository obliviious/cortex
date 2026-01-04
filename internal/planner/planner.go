package planner

import (
	"fmt"

	"github.com/adityaraj/agentflow/internal/config"
)

// ExecutionTask represents a task ready for execution with resolved agent info.
type ExecutionTask struct {
	Name         string   // Task name
	AgentName    string   // Agent reference name
	Tool         string   // CLI tool (claude-code, opencode)
	Model        string   // Model identifier
	Prompt       string   // Prompt text (resolved from prompt_file if needed)
	Write        bool     // Allow file writes
	Dependencies []string // Names of tasks this depends on
}

// ExecutionPlan represents an ordered list of tasks to execute.
type ExecutionPlan struct {
	Tasks []ExecutionTask
	DAG   *DAG // The dependency graph for parallel execution
}

// BuildPlan creates an execution plan from the configuration.
// Returns tasks in dependency order (dependencies before dependents).
func BuildPlan(cfg *config.AgentflowConfig) (*ExecutionPlan, error) {
	// Build DAG from tasks
	dag := BuildDAG(cfg.Tasks)

	// Get topologically sorted task names
	order, err := TopologicalSort(dag)
	if err != nil {
		return nil, fmt.Errorf("failed to sort tasks: %w", err)
	}

	// Build execution tasks with resolved agent info
	tasks := make([]ExecutionTask, 0, len(order))
	for _, name := range order {
		taskCfg := cfg.Tasks[name]
		agentCfg := cfg.Agents[taskCfg.Agent]

		tasks = append(tasks, ExecutionTask{
			Name:         name,
			AgentName:    taskCfg.Agent,
			Tool:         agentCfg.Tool,
			Model:        agentCfg.Model,
			Prompt:       taskCfg.Prompt,
			Write:        taskCfg.Write,
			Dependencies: taskCfg.Needs,
		})
	}

	return &ExecutionPlan{Tasks: tasks, DAG: dag}, nil
}

// String returns a human-readable representation of the execution plan.
func (p *ExecutionPlan) String() string {
	var result string
	for i, task := range p.Tasks {
		line := fmt.Sprintf("  %d. %s (%s -> %s", i+1, task.Name, task.AgentName, task.Tool)
		if task.Model != "" {
			line += "/" + task.Model
		}
		line += ")"
		if len(task.Dependencies) > 0 {
			line += fmt.Sprintf(" [depends: %v]", task.Dependencies)
		}
		result += line + "\n"
	}
	return result
}

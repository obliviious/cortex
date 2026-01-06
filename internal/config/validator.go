package config

import (
	"regexp"
)

// ValidateWithFile checks the configuration for errors, including file path info.
// Returns nil if valid, or a ConfigErrors with all issues found.
func ValidateWithFile(config *AgentflowConfig, filePath string) error {
	errs := &ConfigErrors{}

	// Check for empty config
	if len(config.Agents) == 0 {
		errs.Add(ErrNoAgents(filePath))
	}
	if len(config.Tasks) == 0 {
		errs.Add(ErrNoTasks(filePath))
	}

	// Collect available agent and task names for hints
	availableAgents := make([]string, 0, len(config.Agents))
	for name := range config.Agents {
		availableAgents = append(availableAgents, name)
	}
	availableTasks := make([]string, 0, len(config.Tasks))
	for name := range config.Tasks {
		availableTasks = append(availableTasks, name)
	}

	// Validate agents
	for name, agent := range config.Agents {
		if agent.Tool == "" {
			errs.Add(NewConfigErrorWithHint(filePath, 0,
				"agent \""+name+"\": tool is required",
				"Add 'tool: claude-code', 'tool: opencode', or 'tool: shell'"))
		} else if !IsSupportedTool(agent.Tool) {
			errs.Add(ErrUnsupportedTool(filePath, 0, name, agent.Tool))
		}
	}

	// Validate tasks
	for name, task := range config.Tasks {
		// Check agent reference
		if task.Agent == "" {
			errs.Add(NewConfigErrorWithHint(filePath, 0,
				"task \""+name+"\": agent is required",
				"Add 'agent: <agent_name>' to specify which agent runs this task"))
		} else if _, exists := config.Agents[task.Agent]; !exists {
			errs.Add(ErrUndefinedAgent(filePath, 0, name, task.Agent, availableAgents))
		}

		// Get agent tool type to determine validation rules
		var agentTool string
		if agent, exists := config.Agents[task.Agent]; exists {
			agentTool = agent.Tool
		}

		// Check prompt/command based on agent type
		hasPrompt := task.Prompt != ""
		hasPromptFile := task.PromptFile != ""
		hasCommand := task.Command != ""

		if agentTool == "shell" {
			// Shell agents require 'command' field
			if !hasCommand {
				errs.Add(NewConfigErrorWithHint(filePath, 0,
					"task \""+name+"\": shell agent requires 'command' field",
					"Add 'command: <shell_command>' to specify the command to run"))
			}
			if hasPrompt || hasPromptFile {
				errs.Add(NewConfigErrorWithHint(filePath, 0,
					"task \""+name+"\": shell agent should use 'command', not 'prompt' or 'prompt_file'",
					"Replace 'prompt' or 'prompt_file' with 'command: <shell_command>'"))
			}
		} else {
			// AI agents require prompt or prompt_file
			if !hasPrompt && !hasPromptFile {
				errs.Add(ErrNoPrompt(filePath, 0, name))
			}
			if hasPrompt && hasPromptFile {
				errs.Add(NewConfigErrorWithHint(filePath, 0,
					"task \""+name+"\": cannot have both 'prompt' and 'prompt_file'",
					"Use either inline 'prompt:' or external 'prompt_file:', not both"))
			}
			if hasCommand {
				errs.Add(NewConfigErrorWithHint(filePath, 0,
					"task \""+name+"\": 'command' field is only for shell agents",
					"Use 'prompt' or 'prompt_file' for AI agents, or change agent tool to 'shell'"))
			}
		}

		// Check dependency references
		for _, dep := range task.Needs {
			if _, exists := config.Tasks[dep]; !exists {
				errs.Add(ErrUndefinedDependency(filePath, 0, name, dep, availableTasks))
			}
			if dep == name {
				errs.Add(ErrSelfDependency(filePath, 0, name))
			}
		}

		// Validate template variables reference valid dependencies
		templateErrs := validateTemplateVarsStructured(filePath, name, task.Prompt, task.Needs, config.Tasks)
		for _, e := range templateErrs {
			errs.Add(e)
		}
	}

	// Check for circular dependencies
	if cycle := detectCycleSlice(config.Tasks); cycle != nil {
		errs.Add(ErrCircularDependency(filePath, cycle))
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// Validate checks the configuration for errors (backward compatible).
// Returns nil if valid, or a ConfigErrors with all issues found.
func Validate(config *AgentflowConfig) error {
	return ValidateWithFile(config, "Cortexfile.yml")
}

// templateVarRegex matches {{outputs.taskname}} patterns.
var templateVarRegex = regexp.MustCompile(`\{\{outputs\.([a-zA-Z0-9_-]+)\}\}`)

// validateTemplateVarsStructured checks that all {{outputs.X}} references are valid dependencies.
func validateTemplateVarsStructured(filePath, taskName, prompt string, needs []string, tasks map[string]TaskConfig) []*ConfigError {
	var errs []*ConfigError

	matches := templateVarRegex.FindAllStringSubmatch(prompt, -1)
	needsSet := make(map[string]bool)
	for _, n := range needs {
		needsSet[n] = true
	}

	for _, match := range matches {
		refTask := match[1]

		// Check if referenced task exists
		if _, exists := tasks[refTask]; !exists {
			errs = append(errs, NewConfigErrorWithHint(filePath, 0,
				"task \""+taskName+"\": template references undefined task \""+refTask+"\"",
				"Define the task or fix the template variable name"))
			continue
		}

		// Check if referenced task is in needs
		if !needsSet[refTask] {
			errs = append(errs, NewConfigErrorWithHint(filePath, 0,
				"task \""+taskName+"\": template references \""+refTask+"\" which is not in 'needs'",
				"Add '"+refTask+"' to the 'needs' list to ensure it runs first"))
		}
	}

	return errs
}

// detectCycleSlice uses DFS to find circular dependencies and returns the cycle.
func detectCycleSlice(tasks map[string]TaskConfig) []string {
	// States: 0 = unvisited, 1 = visiting (in current path), 2 = visited
	state := make(map[string]int)
	var path []string

	var visit func(name string) []string
	visit = func(name string) []string {
		if state[name] == 2 {
			return nil // Already fully processed
		}
		if state[name] == 1 {
			// Found cycle - build cycle path for error message
			cycleStart := -1
			for i, p := range path {
				if p == name {
					cycleStart = i
					break
				}
			}
			return append(path[cycleStart:], name)
		}

		state[name] = 1
		path = append(path, name)

		task := tasks[name]
		for _, dep := range task.Needs {
			if cycle := visit(dep); cycle != nil {
				return cycle
			}
		}

		path = path[:len(path)-1]
		state[name] = 2
		return nil
	}

	for name := range tasks {
		if state[name] == 0 {
			if cycle := visit(name); cycle != nil {
				return cycle
			}
		}
	}

	return nil
}

// detectCycles is kept for backward compatibility (returns error).
func detectCycles(tasks map[string]TaskConfig) error {
	cycle := detectCycleSlice(tasks)
	if cycle != nil {
		return ErrCircularDependency("", cycle)
	}
	return nil
}

// ValidationError is kept for backward compatibility.
type ValidationError = ConfigErrors

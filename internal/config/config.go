// Package config handles Agentfile parsing and configuration structures.
package config

import (
	"gopkg.in/yaml.v3"
)

// AgentflowConfig represents the root configuration from Cortexfile.yml.
type AgentflowConfig struct {
	Agents   map[string]AgentConfig `yaml:"agents"`
	Tasks    map[string]TaskConfig  `yaml:"tasks"`
	Settings *SettingsConfig        `yaml:"settings"` // Optional local settings
}

// AgentConfig defines an AI agent's configuration.
type AgentConfig struct {
	Tool  string `yaml:"tool"`  // "claude-code" or "opencode"
	Model string `yaml:"model"` // Optional: model identifier (e.g., "sonnet", "opus")
}

// TaskConfig defines a single task's configuration.
type TaskConfig struct {
	Agent      string     `yaml:"agent"`       // Reference to agent name in agents section
	Prompt     string     `yaml:"prompt"`      // Inline prompt text (option A)
	PromptFile string     `yaml:"prompt_file"` // Path to prompt file (option B)
	Needs      StringList `yaml:"needs"`       // Dependencies: single string or array
	Write      bool       `yaml:"write"`       // Allow file writes (default: false)
}

// StringList is a custom type that can unmarshal from either a single string or an array of strings.
// This allows YAML like:
//
//	needs: task1          # single dependency
//	needs: [task1, task2] # multiple dependencies
type StringList []string

// UnmarshalYAML implements custom unmarshaling for StringList to handle both string and []string.
func (s *StringList) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		// Single string value
		var single string
		if err := node.Decode(&single); err != nil {
			return err
		}
		if single != "" {
			*s = []string{single}
		} else {
			*s = []string{}
		}
		return nil

	case yaml.SequenceNode:
		// Array of strings
		var list []string
		if err := node.Decode(&list); err != nil {
			return err
		}
		*s = list
		return nil

	default:
		// Null or empty
		*s = []string{}
		return nil
	}
}

// SupportedTools lists all valid tool values for agents.
var SupportedTools = []string{"claude-code", "opencode"}

// IsSupportedTool checks if a tool name is valid.
func IsSupportedTool(tool string) bool {
	for _, t := range SupportedTools {
		if t == tool {
			return true
		}
	}
	return false
}

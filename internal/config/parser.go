package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads and parses an Agentfile from the given path.
// It also resolves prompt_file references relative to the Agentfile directory.
func LoadConfig(path string) (*AgentflowConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return ParseConfig(data, filepath.Dir(path))
}

// ParseConfig parses YAML config data and resolves prompt_file references.
// baseDir is used to resolve relative prompt_file paths.
func ParseConfig(data []byte, baseDir string) (*AgentflowConfig, error) {
	var config AgentflowConfig

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Initialize maps if nil (empty config)
	if config.Agents == nil {
		config.Agents = make(map[string]AgentConfig)
	}
	if config.Tasks == nil {
		config.Tasks = make(map[string]TaskConfig)
	}

	// Resolve prompt_file references
	if err := resolvePromptFiles(&config, baseDir); err != nil {
		return nil, err
	}

	return &config, nil
}

// resolvePromptFiles loads content from prompt_file paths into the Prompt field.
func resolvePromptFiles(config *AgentflowConfig, baseDir string) error {
	for name, task := range config.Tasks {
		if task.PromptFile != "" {
			// Resolve path relative to config file directory
			promptPath := task.PromptFile
			if !filepath.IsAbs(promptPath) {
				promptPath = filepath.Join(baseDir, promptPath)
			}

			content, err := os.ReadFile(promptPath)
			if err != nil {
				return fmt.Errorf("task %q: failed to read prompt_file %q: %w", name, task.PromptFile, err)
			}

			// Store the loaded content in Prompt field
			task.Prompt = string(content)
			config.Tasks[name] = task
		}
	}
	return nil
}

// FindCortexfile searches for a Cortexfile in the current directory.
// It looks for: Cortexfile.yml, Cortexfile.yaml, cortexfile.yml, cortexfile.yaml
// Also supports legacy: Agentfile.yml, Agentfile.yaml
func FindCortexfile(dir string) (string, error) {
	candidates := []string{
		"Cortexfile.yml",
		"Cortexfile.yaml",
		"cortexfile.yml",
		"cortexfile.yaml",
		// Legacy support
		"Agentfile.yml",
		"Agentfile.yaml",
	}

	for _, name := range candidates {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no Cortexfile found in %s (tried: %v)", dir, candidates[:4])
}

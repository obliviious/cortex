package config

import (
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// GlobalConfig represents the global ~/.cortex/config.yml configuration.
type GlobalConfig struct {
	Defaults DefaultsConfig  `yaml:"defaults"`
	Settings SettingsConfig  `yaml:"settings"`
	Webhooks []WebhookConfig `yaml:"webhooks"`
}

// DefaultsConfig contains default agent settings.
type DefaultsConfig struct {
	Model string `yaml:"model"` // Default model (e.g., "sonnet")
	Tool  string `yaml:"tool"`  // Default tool (e.g., "claude-code")
}

// SettingsConfig contains execution settings.
type SettingsConfig struct {
	Parallel    bool `yaml:"parallel"`     // Enable parallel execution (default: true)
	MaxParallel int  `yaml:"max_parallel"` // Max concurrent tasks (default: CPU cores)
	Verbose     bool `yaml:"verbose"`      // Verbose output
	Stream      bool `yaml:"stream"`       // Stream agent logs
}

// WebhookConfig defines a webhook endpoint.
type WebhookConfig struct {
	URL     string            `yaml:"url"`
	Events  []string          `yaml:"events"` // Events to trigger on
	Headers map[string]string `yaml:"headers"`
}

// DefaultSettings returns the default settings.
func DefaultSettings() SettingsConfig {
	return SettingsConfig{
		Parallel:    true,
		MaxParallel: runtime.NumCPU(),
		Verbose:     false,
		Stream:      false,
	}
}

// LoadGlobalConfig loads the global configuration from ~/.cortex/config.yml.
// Returns an empty config (with defaults) if the file doesn't exist.
func LoadGlobalConfig() (*GlobalConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return defaultGlobalConfig(), nil
	}

	configPath := filepath.Join(homeDir, ".cortex", "config.yml")
	return LoadGlobalConfigFromPath(configPath)
}

// LoadGlobalConfigFromPath loads global config from a specific path.
func LoadGlobalConfigFromPath(path string) (*GlobalConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultGlobalConfig(), nil
		}
		return nil, err
	}

	var config GlobalConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Apply defaults for unset values
	applyDefaults(&config)

	return &config, nil
}

// defaultGlobalConfig returns a GlobalConfig with all defaults.
func defaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		Settings: DefaultSettings(),
	}
}

// applyDefaults fills in default values for unset fields.
func applyDefaults(config *GlobalConfig) {
	defaults := DefaultSettings()

	if config.Settings.MaxParallel <= 0 {
		config.Settings.MaxParallel = defaults.MaxParallel
	}
	// Note: Parallel defaults to false from YAML, so we check if it was explicitly set
	// This is handled by the caller with CLI flags taking precedence
}

// MergedConfig holds the final merged configuration.
type MergedConfig struct {
	// From AgentflowConfig
	Agents map[string]AgentConfig
	Tasks  map[string]TaskConfig

	// Merged settings
	Settings SettingsConfig

	// From global config
	Webhooks []WebhookConfig

	// Defaults for agents
	Defaults DefaultsConfig
}

// MergeConfigs combines global config, local Cortexfile, and CLI flags.
// Priority: CLI flags > Cortexfile settings > Global config
func MergeConfigs(global *GlobalConfig, local *AgentflowConfig, cliSettings *SettingsConfig) *MergedConfig {
	merged := &MergedConfig{
		Agents:   local.Agents,
		Tasks:    local.Tasks,
		Webhooks: global.Webhooks,
		Defaults: global.Defaults,
	}

	// Start with global settings
	merged.Settings = global.Settings

	// Override with local Cortexfile settings if present
	if local.Settings != nil {
		if local.Settings.MaxParallel > 0 {
			merged.Settings.MaxParallel = local.Settings.MaxParallel
		}
		// Parallel is tricky - we need to know if it was explicitly set
		// For now, local settings override global
		merged.Settings.Parallel = local.Settings.Parallel
		merged.Settings.Verbose = local.Settings.Verbose || merged.Settings.Verbose
		merged.Settings.Stream = local.Settings.Stream || merged.Settings.Stream
	}

	// Override with CLI flags (highest priority)
	if cliSettings != nil {
		if cliSettings.MaxParallel > 0 {
			merged.Settings.MaxParallel = cliSettings.MaxParallel
		}
		// CLI flags always win
		merged.Settings.Verbose = cliSettings.Verbose || merged.Settings.Verbose
		merged.Settings.Stream = cliSettings.Stream || merged.Settings.Stream
	}

	// Apply default model/tool to agents that don't specify them
	for name, agent := range merged.Agents {
		if agent.Model == "" && merged.Defaults.Model != "" {
			agent.Model = merged.Defaults.Model
			merged.Agents[name] = agent
		}
		if agent.Tool == "" && merged.Defaults.Tool != "" {
			agent.Tool = merged.Defaults.Tool
			merged.Agents[name] = agent
		}
	}

	return merged
}

// MatchesEvent checks if a webhook should be triggered for an event.
func (w *WebhookConfig) MatchesEvent(eventType string) bool {
	if len(w.Events) == 0 {
		return true // No filter = all events
	}
	for _, e := range w.Events {
		if e == eventType || e == "*" {
			return true
		}
	}
	return false
}

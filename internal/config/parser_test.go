package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestParseConfig tests YAML parsing functionality.
func TestParseConfig(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		baseDir string
		wantErr bool
		validate func(*testing.T, *AgentflowConfig)
	}{
		{
			name: "valid minimal config",
			yaml: `
agents:
  agent1:
    tool: claude-code

tasks:
  task1:
    agent: agent1
    prompt: "test prompt"
`,
			baseDir: "/tmp",
			wantErr: false,
			validate: func(t *testing.T, cfg *AgentflowConfig) {
				if len(cfg.Agents) != 1 {
					t.Errorf("expected 1 agent, got %d", len(cfg.Agents))
				}
				if len(cfg.Tasks) != 1 {
					t.Errorf("expected 1 task, got %d", len(cfg.Tasks))
				}
				if cfg.Agents["agent1"].Tool != "claude-code" {
					t.Errorf("expected tool claude-code, got %s", cfg.Agents["agent1"].Tool)
				}
				if cfg.Tasks["task1"].Prompt != "test prompt" {
					t.Errorf("expected prompt 'test prompt', got %s", cfg.Tasks["task1"].Prompt)
				}
			},
		},
		{
			name: "empty config",
			yaml: ``,
			baseDir: "/tmp",
			wantErr: false,
			validate: func(t *testing.T, cfg *AgentflowConfig) {
				if cfg.Agents == nil {
					t.Error("expected Agents map to be initialized")
				}
				if cfg.Tasks == nil {
					t.Error("expected Tasks map to be initialized")
				}
			},
		},
		{
			name: "config with model specified",
			yaml: `
agents:
  agent1:
    tool: opencode
    model: sonnet

tasks:
  task1:
    agent: agent1
    prompt: "test"
`,
			baseDir: "/tmp",
			wantErr: false,
			validate: func(t *testing.T, cfg *AgentflowConfig) {
				if cfg.Agents["agent1"].Model != "sonnet" {
					t.Errorf("expected model sonnet, got %s", cfg.Agents["agent1"].Model)
				}
			},
		},
		{
			name: "task with write permission",
			yaml: `
agents:
  agent1:
    tool: claude-code

tasks:
  task1:
    agent: agent1
    prompt: "test"
    write: true
`,
			baseDir: "/tmp",
			wantErr: false,
			validate: func(t *testing.T, cfg *AgentflowConfig) {
				if !cfg.Tasks["task1"].Write {
					t.Error("expected Write to be true")
				}
			},
		},
		{
			name: "task with single dependency as string",
			yaml: `
agents:
  agent1:
    tool: claude-code

tasks:
  task1:
    agent: agent1
    prompt: "test1"
  task2:
    agent: agent1
    prompt: "test2"
    needs: task1
`,
			baseDir: "/tmp",
			wantErr: false,
			validate: func(t *testing.T, cfg *AgentflowConfig) {
				if len(cfg.Tasks["task2"].Needs) != 1 {
					t.Errorf("expected 1 dependency, got %d", len(cfg.Tasks["task2"].Needs))
				}
				if cfg.Tasks["task2"].Needs[0] != "task1" {
					t.Errorf("expected dependency task1, got %s", cfg.Tasks["task2"].Needs[0])
				}
			},
		},
		{
			name: "task with multiple dependencies as array",
			yaml: `
agents:
  agent1:
    tool: claude-code

tasks:
  task1:
    agent: agent1
    prompt: "test1"
  task2:
    agent: agent1
    prompt: "test2"
  task3:
    agent: agent1
    prompt: "test3"
    needs: [task1, task2]
`,
			baseDir: "/tmp",
			wantErr: false,
			validate: func(t *testing.T, cfg *AgentflowConfig) {
				if len(cfg.Tasks["task3"].Needs) != 2 {
					t.Errorf("expected 2 dependencies, got %d", len(cfg.Tasks["task3"].Needs))
				}
			},
		},
		{
			name: "invalid YAML",
			yaml: `
agents:
  agent1:
    tool: claude-code
  invalid yaml here
`,
			baseDir: "/tmp",
			wantErr: true,
		},
		{
			name: "malformed YAML structure",
			yaml: `
agents:
  - this should be a map not a list
`,
			baseDir: "/tmp",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := ParseConfig([]byte(tt.yaml), tt.baseDir)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

// TestLoadConfig tests loading config from file.
func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "agentflow-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test config file
	configContent := `
agents:
  agent1:
    tool: claude-code

tasks:
  task1:
    agent: agent1
    prompt: "test prompt"
`
	configPath := filepath.Join(tmpDir, "test-config.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Test loading the config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if len(cfg.Agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(cfg.Agents))
	}
	if len(cfg.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(cfg.Tasks))
	}
}

// TestLoadConfig_FileNotFound tests error handling for missing files.
func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/to/config.yml")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read config file") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestResolvePromptFiles tests prompt file resolution.
func TestResolvePromptFiles(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "agentflow-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a prompt file
	promptContent := "This is a test prompt from file"
	promptPath := filepath.Join(tmpDir, "prompt.txt")
	if err := os.WriteFile(promptPath, []byte(promptContent), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	tests := []struct {
		name        string
		task        TaskConfig
		baseDir     string
		wantPrompt  string
		wantErr     bool
		wantErrContains string
	}{
		{
			name: "relative prompt file",
			task: TaskConfig{
				Agent:      "agent1",
				PromptFile: "prompt.txt",
			},
			baseDir:    tmpDir,
			wantPrompt: promptContent,
			wantErr:    false,
		},
		{
			name: "absolute prompt file",
			task: TaskConfig{
				Agent:      "agent1",
				PromptFile: promptPath,
			},
			baseDir:    "/some/other/dir",
			wantPrompt: promptContent,
			wantErr:    false,
		},
		{
			name: "no prompt file specified",
			task: TaskConfig{
				Agent:  "agent1",
				Prompt: "inline prompt",
			},
			baseDir:    tmpDir,
			wantPrompt: "inline prompt",
			wantErr:    false,
		},
		{
			name: "nonexistent prompt file",
			task: TaskConfig{
				Agent:      "agent1",
				PromptFile: "nonexistent.txt",
			},
			baseDir: tmpDir,
			wantErr: true,
			wantErrContains: "failed to read prompt_file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &AgentflowConfig{
				Agents: map[string]AgentConfig{
					"agent1": {Tool: "claude-code"},
				},
				Tasks: map[string]TaskConfig{
					"task1": tt.task,
				},
			}

			err := resolvePromptFiles(config, tt.baseDir)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrContains != "" && !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("expected error containing %q, got: %v", tt.wantErrContains, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if config.Tasks["task1"].Prompt != tt.wantPrompt {
				t.Errorf("expected prompt %q, got %q", tt.wantPrompt, config.Tasks["task1"].Prompt)
			}
		})
	}
}

// TestFindCortexfile tests finding Cortexfile in directory.
func TestFindCortexfile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "agentflow-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name         string
		createFiles  []string
		wantFileName string
		wantErr      bool
	}{
		{
			name:         "finds Cortexfile.yml",
			createFiles:  []string{"Cortexfile.yml"},
			wantFileName: "Cortexfile.yml",
			wantErr:      false,
		},
		{
			name:         "finds Cortexfile.yaml",
			createFiles:  []string{"Cortexfile.yaml"},
			wantFileName: "Cortexfile.yaml",
			wantErr:      false,
		},
		{
			name:         "finds cortexfile.yml (lowercase)",
			createFiles:  []string{"cortexfile.yml"},
			wantFileName: "cortexfile.yml",
			wantErr:      false,
		},
		{
			name:         "finds cortexfile.yaml (lowercase)",
			createFiles:  []string{"cortexfile.yaml"},
			wantFileName: "cortexfile.yaml",
			wantErr:      false,
		},
		{
			name:         "finds legacy Agentfile.yml",
			createFiles:  []string{"Agentfile.yml"},
			wantFileName: "Agentfile.yml",
			wantErr:      false,
		},
		{
			name:         "finds legacy Agentfile.yaml",
			createFiles:  []string{"Agentfile.yaml"},
			wantFileName: "Agentfile.yaml",
			wantErr:      false,
		},
		{
			name:         "prefers Cortexfile.yml over legacy",
			createFiles:  []string{"Cortexfile.yml", "Agentfile.yml"},
			wantFileName: "Cortexfile.yml",
			wantErr:      false,
		},
		{
			name:        "no cortexfile found",
			createFiles: []string{"other.yml"},
			wantErr:     true,
		},
		{
			name:        "empty directory",
			createFiles: []string{},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a subdirectory for this test
			testDir, err := os.MkdirTemp(tmpDir, "test-*")
			if err != nil {
				t.Fatalf("failed to create test dir: %v", err)
			}
			defer os.RemoveAll(testDir)

			// Create the test files
			for _, filename := range tt.createFiles {
				path := filepath.Join(testDir, filename)
				if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			}

			// Test FindCortexfile
			foundPath, err := FindCortexfile(testDir)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify the base name matches (case may vary on case-insensitive filesystems)
			foundBasename := filepath.Base(foundPath)
			if !strings.EqualFold(foundBasename, tt.wantFileName) {
				t.Errorf("expected basename %q, got %q", tt.wantFileName, foundBasename)
			}

			// Verify the directory matches
			if filepath.Dir(foundPath) != testDir {
				t.Errorf("expected dir %q, got %q", testDir, filepath.Dir(foundPath))
			}
		})
	}
}

// TestStringList_UnmarshalYAML tests the StringList custom unmarshaling.
func TestStringList_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    []string
		wantErr bool
	}{
		{
			name: "single string",
			yaml: `needs: task1`,
			want: []string{"task1"},
		},
		{
			name: "array of strings",
			yaml: `needs: [task1, task2, task3]`,
			want: []string{"task1", "task2", "task3"},
		},
		{
			name: "empty string",
			yaml: `needs: ""`,
			want: []string{},
		},
		{
			name: "null value",
			yaml: `needs: null`,
			want: []string{},
		},
		{
			name: "no value",
			yaml: `other: value`,
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type testStruct struct {
				Needs StringList `yaml:"needs"`
			}

			var result testStruct
			err := yaml.Unmarshal([]byte(tt.yaml), &result)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Compare slices
			if len(result.Needs) != len(tt.want) {
				t.Errorf("expected %d items, got %d: %v", len(tt.want), len(result.Needs), result.Needs)
				return
			}

			for i, want := range tt.want {
				if result.Needs[i] != want {
					t.Errorf("item %d: expected %q, got %q", i, want, result.Needs[i])
				}
			}
		})
	}
}

// TestParseConfig_RealWorldExample tests a realistic configuration.
func TestParseConfig_RealWorldExample(t *testing.T) {
	yaml := `
agents:
  architect:
    tool: claude-code
    model: opus
  implementer:
    tool: opencode
    model: sonnet

tasks:
  analyze:
    agent: architect
    prompt: "Analyze the codebase and suggest improvements"

  design:
    agent: architect
    prompt: "Based on: {{outputs.analyze}}\nCreate a design plan"
    needs: analyze

  implement:
    agent: implementer
    prompt: "Implement the design: {{outputs.design}}"
    needs: design
    write: true

  test:
    agent: implementer
    prompt: "Write tests for the implementation"
    needs: implement
`

	cfg, err := ParseConfig([]byte(yaml), "/tmp")
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}

	// Validate structure
	if len(cfg.Agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(cfg.Agents))
	}
	if len(cfg.Tasks) != 4 {
		t.Errorf("expected 4 tasks, got %d", len(cfg.Tasks))
	}

	// Validate dependencies
	if len(cfg.Tasks["design"].Needs) != 1 {
		t.Errorf("design task: expected 1 dependency, got %d", len(cfg.Tasks["design"].Needs))
	}

	// Validate write permission
	if !cfg.Tasks["implement"].Write {
		t.Error("implement task should have write permission")
	}
	if cfg.Tasks["test"].Write {
		t.Error("test task should not have write permission by default")
	}

	// Validate models
	if cfg.Agents["architect"].Model != "opus" {
		t.Errorf("architect model: expected opus, got %s", cfg.Agents["architect"].Model)
	}
}

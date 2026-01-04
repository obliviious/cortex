package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Store handles persistence of run results to disk.
type Store struct {
	baseDir    string // Base directory (~/.agentflow)
	runID      string // Current run ID (timestamp-based)
	runDir     string // Full path to current run directory
	projectDir string // Project directory where agentflow was run
}

// NewStore creates a new Store using ~/.cortex as the base directory.
// Creates ~/.cortex/sessions/<project-name>/ structure if it doesn't exist.
func NewStore(projectDir string) (*Store, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, ".cortex")
	runID := time.Now().Format("20060102-150405")

	// Create project-specific session directory
	projectName := filepath.Base(projectDir)
	sessionsDir := filepath.Join(baseDir, "sessions", projectName)
	runDir := filepath.Join(sessionsDir, "run-"+runID)

	if err := os.MkdirAll(runDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create run directory: %w", err)
	}

	return &Store{
		baseDir:    baseDir,
		runID:      runID,
		runDir:     runDir,
		projectDir: projectDir,
	}, nil
}

// NewStoreWithPath creates a Store with a custom base path (for testing).
func NewStoreWithPath(basePath, projectDir string) (*Store, error) {
	runID := time.Now().Format("20060102-150405")
	projectName := filepath.Base(projectDir)
	sessionsDir := filepath.Join(basePath, "sessions", projectName)
	runDir := filepath.Join(sessionsDir, "run-"+runID)

	if err := os.MkdirAll(runDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create run directory: %w", err)
	}

	return &Store{
		baseDir:    basePath,
		runID:      runID,
		runDir:     runDir,
		projectDir: projectDir,
	}, nil
}

// SaveTaskResult saves a task result to disk as JSON.
func (s *Store) SaveTaskResult(result *TaskResult) error {
	filename := filepath.Join(s.runDir, result.TaskName+".json")

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write result file: %w", err)
	}

	return nil
}

// SaveRunResult saves the complete run result to disk.
func (s *Store) SaveRunResult(result *RunResult) error {
	filename := filepath.Join(s.runDir, "run.json")

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal run result: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write run result: %w", err)
	}

	return nil
}

// RunDir returns the path to the current run directory.
func (s *Store) RunDir() string {
	return s.runDir
}

// RunID returns the current run ID.
func (s *Store) RunID() string {
	return s.runID
}

// LoadTaskResult loads a task result from disk.
func (s *Store) LoadTaskResult(taskName string) (*TaskResult, error) {
	filename := filepath.Join(s.runDir, taskName+".json")

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read result file: %w", err)
	}

	var result TaskResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SessionInfo contains summary information about a session.
type SessionInfo struct {
	RunID     string        `json:"run_id"`
	Project   string        `json:"project"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Success   bool          `json:"success"`
	TaskCount int           `json:"task_count"`
	Duration  time.Duration `json:"duration"`
	RunDir    string        `json:"run_dir"`
}

// SessionFilter contains filter options for listing sessions.
type SessionFilter struct {
	Project    string // Filter by project name (empty = all projects)
	Limit      int    // Maximum number of sessions to return (0 = no limit)
	FailedOnly bool   // Only show failed sessions
}

// ListSessions lists all sessions from ~/.cortex/sessions.
func ListSessions(filter SessionFilter) ([]SessionInfo, error) {
	baseDir, err := getCortexDir()
	if err != nil {
		return nil, err
	}

	return ListSessionsFromPath(baseDir, filter)
}

// ListSessionsFromPath lists sessions from a custom base path (for testing).
func ListSessionsFromPath(baseDir string, filter SessionFilter) ([]SessionInfo, error) {
	sessionsDir := filepath.Join(baseDir, "sessions")

	// Check if sessions directory exists
	if _, err := os.Stat(sessionsDir); os.IsNotExist(err) {
		return []SessionInfo{}, nil
	}

	var sessions []SessionInfo

	// If project is specified, only look in that directory
	if filter.Project != "" {
		projectDir := filepath.Join(sessionsDir, filter.Project)
		projectSessions, err := listProjectSessions(projectDir, filter.Project)
		if err != nil {
			if os.IsNotExist(err) {
				return []SessionInfo{}, nil
			}
			return nil, err
		}
		sessions = projectSessions
	} else {
		// List all projects
		projectDirs, err := os.ReadDir(sessionsDir)
		if err != nil {
			return nil, err
		}

		for _, projectEntry := range projectDirs {
			if !projectEntry.IsDir() {
				continue
			}
			projectName := projectEntry.Name()
			projectDir := filepath.Join(sessionsDir, projectName)

			projectSessions, err := listProjectSessions(projectDir, projectName)
			if err != nil {
				continue // Skip projects we can't read
			}
			sessions = append(sessions, projectSessions...)
		}
	}

	// Filter failed only
	if filter.FailedOnly {
		filtered := make([]SessionInfo, 0)
		for _, s := range sessions {
			if !s.Success {
				filtered = append(filtered, s)
			}
		}
		sessions = filtered
	}

	// Sort by start time (newest first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartTime.After(sessions[j].StartTime)
	})

	// Apply limit
	if filter.Limit > 0 && len(sessions) > filter.Limit {
		sessions = sessions[:filter.Limit]
	}

	return sessions, nil
}

// listProjectSessions lists all sessions within a project directory.
func listProjectSessions(projectDir, projectName string) ([]SessionInfo, error) {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return nil, err
	}

	var sessions []SessionInfo

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		runDirName := entry.Name()
		if !strings.HasPrefix(runDirName, "run-") {
			continue
		}

		runDir := filepath.Join(projectDir, runDirName)
		runID := strings.TrimPrefix(runDirName, "run-")

		session, err := loadSessionInfo(runDir, runID, projectName)
		if err != nil {
			// Skip sessions we can't load
			continue
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// loadSessionInfo loads session info from a run directory.
func loadSessionInfo(runDir, runID, project string) (SessionInfo, error) {
	runFile := filepath.Join(runDir, "run.json")

	data, err := os.ReadFile(runFile)
	if err != nil {
		// Try to construct info from directory name
		return SessionInfo{
			RunID:   runID,
			Project: project,
			RunDir:  runDir,
		}, nil
	}

	var runResult RunResult
	if err := json.Unmarshal(data, &runResult); err != nil {
		return SessionInfo{
			RunID:   runID,
			Project: project,
			RunDir:  runDir,
		}, nil
	}

	return SessionInfo{
		RunID:     runResult.RunID,
		Project:   project,
		StartTime: runResult.StartTime,
		EndTime:   runResult.EndTime,
		Success:   runResult.Success,
		TaskCount: len(runResult.Tasks),
		Duration:  runResult.EndTime.Sub(runResult.StartTime),
		RunDir:    runDir,
	}, nil
}

// GetSession loads full session details by run ID.
func GetSession(project, runID string) (*RunResult, error) {
	baseDir, err := getCortexDir()
	if err != nil {
		return nil, err
	}

	return GetSessionFromPath(baseDir, project, runID)
}

// GetSessionFromPath loads session from a custom base path.
func GetSessionFromPath(baseDir, project, runID string) (*RunResult, error) {
	runDir := filepath.Join(baseDir, "sessions", project, "run-"+runID)
	runFile := filepath.Join(runDir, "run.json")

	data, err := os.ReadFile(runFile)
	if err != nil {
		return nil, err
	}

	var result RunResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ListProjects lists all projects with sessions.
func ListProjects() ([]string, error) {
	baseDir, err := getCortexDir()
	if err != nil {
		return nil, err
	}

	return ListProjectsFromPath(baseDir)
}

// ListProjectsFromPath lists projects from a custom base path.
func ListProjectsFromPath(baseDir string) ([]string, error) {
	sessionsDir := filepath.Join(baseDir, "sessions")

	if _, err := os.Stat(sessionsDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return nil, err
	}

	var projects []string
	for _, entry := range entries {
		if entry.IsDir() {
			projects = append(projects, entry.Name())
		}
	}

	sort.Strings(projects)
	return projects, nil
}

// getCortexDir returns the ~/.cortex directory path.
func getCortexDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".cortex"), nil
}

// FormatDuration formats a duration in human-readable form.
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}
	if d < time.Minute {
		return d.Round(time.Second * 1).String()
	}
	if d < time.Hour {
		return d.Round(time.Second * 1).String()
	}
	return d.Round(time.Minute).String()
}

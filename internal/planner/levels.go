package planner

// ExecutionLevel represents a group of tasks that can run in parallel.
// All tasks in the same level have no dependencies on each other.
type ExecutionLevel struct {
	Level int      // Level number (0 = root tasks)
	Tasks []string // Task names at this level
}

// BuildExecutionLevels groups tasks by dependency level for parallel execution.
// Level 0 contains tasks with no dependencies (roots).
// Level N contains tasks that depend only on tasks in levels 0..N-1.
func BuildExecutionLevels(dag *DAG) []ExecutionLevel {
	if dag.Size() == 0 {
		return nil
	}

	// Track remaining in-degree for each task
	remaining := make(map[string]int)
	for name, degree := range dag.InDegree {
		remaining[name] = degree
	}

	// Track which tasks have been assigned to a level
	assigned := make(map[string]bool)

	var levels []ExecutionLevel
	levelNum := 0

	for len(assigned) < dag.Size() {
		// Find all tasks that can run at this level
		// (all dependencies already assigned to previous levels)
		var levelTasks []string

		for name := range dag.Nodes {
			if assigned[name] {
				continue
			}
			if remaining[name] == 0 {
				levelTasks = append(levelTasks, name)
			}
		}

		if len(levelTasks) == 0 {
			// This shouldn't happen with a valid DAG (no cycles)
			break
		}

		// Add this level
		levels = append(levels, ExecutionLevel{
			Level: levelNum,
			Tasks: levelTasks,
		})

		// Mark these tasks as assigned and decrement in-degree of dependents
		for _, taskName := range levelTasks {
			assigned[taskName] = true

			// Decrement in-degree of all tasks that depend on this one
			for _, dependent := range dag.ReverseEdges[taskName] {
				remaining[dependent]--
			}
		}

		levelNum++
	}

	return levels
}

// TotalTasks returns the total number of tasks across all levels.
func TotalTasks(levels []ExecutionLevel) int {
	total := 0
	for _, level := range levels {
		total += len(level.Tasks)
	}
	return total
}

// MaxParallelism returns the maximum number of tasks that could run in parallel.
func MaxParallelism(levels []ExecutionLevel) int {
	max := 0
	for _, level := range levels {
		if len(level.Tasks) > max {
			max = len(level.Tasks)
		}
	}
	return max
}

// LevelForTask returns the level number for a given task, or -1 if not found.
func LevelForTask(levels []ExecutionLevel, taskName string) int {
	for _, level := range levels {
		for _, name := range level.Tasks {
			if name == taskName {
				return level.Level
			}
		}
	}
	return -1
}

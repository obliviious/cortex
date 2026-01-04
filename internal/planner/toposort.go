package planner

import (
	"fmt"
	"sort"
)

// TopologicalSort performs Kahn's algorithm to get tasks in execution order.
// Returns tasks ordered so that all dependencies come before dependents.
// Returns an error if a cycle is detected (should not happen if validation passed).
func TopologicalSort(dag *DAG) ([]string, error) {
	// Copy in-degree map (we'll modify it during the algorithm)
	inDegree := make(map[string]int)
	for name, degree := range dag.InDegree {
		inDegree[name] = degree
	}

	// Initialize queue with all root nodes (no dependencies)
	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	// Sort initial queue for deterministic ordering
	sort.Strings(queue)

	var result []string
	for len(queue) > 0 {
		// Take first element from queue
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Process all tasks that depend on current
		var newReady []string
		for _, dependent := range dag.ReverseEdges[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				newReady = append(newReady, dependent)
			}
		}

		// Sort newly ready tasks for deterministic ordering
		sort.Strings(newReady)
		queue = append(queue, newReady...)
	}

	// Check if all nodes were processed
	if len(result) != dag.Size() {
		return nil, fmt.Errorf("cycle detected: only processed %d of %d tasks", len(result), dag.Size())
	}

	return result, nil
}

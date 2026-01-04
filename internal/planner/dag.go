// Package planner builds execution plans from task configurations.
package planner

import (
	"github.com/adityaraj/agentflow/internal/config"
)

// DAG represents a Directed Acyclic Graph of task dependencies.
type DAG struct {
	// Nodes maps task names to their configuration
	Nodes map[string]config.TaskConfig

	// Edges maps each task to its dependencies (tasks it depends on)
	Edges map[string][]string

	// ReverseEdges maps each task to tasks that depend on it
	ReverseEdges map[string][]string

	// InDegree tracks the number of dependencies for each task
	InDegree map[string]int
}

// BuildDAG constructs a DAG from the task configuration.
func BuildDAG(tasks map[string]config.TaskConfig) *DAG {
	dag := &DAG{
		Nodes:        make(map[string]config.TaskConfig),
		Edges:        make(map[string][]string),
		ReverseEdges: make(map[string][]string),
		InDegree:     make(map[string]int),
	}

	// Initialize nodes and in-degrees
	for name, task := range tasks {
		dag.Nodes[name] = task
		dag.InDegree[name] = 0
		dag.Edges[name] = []string{}
		dag.ReverseEdges[name] = []string{}
	}

	// Build edges from dependencies
	for name, task := range tasks {
		for _, dep := range task.Needs {
			// Edge: name depends on dep (name -> dep in dependency direction)
			dag.Edges[name] = append(dag.Edges[name], dep)

			// Reverse edge: dep is depended on by name
			dag.ReverseEdges[dep] = append(dag.ReverseEdges[dep], name)

			// Increment in-degree of the dependent task
			dag.InDegree[name]++
		}
	}

	return dag
}

// GetRoots returns all tasks with no dependencies (in-degree = 0).
func (d *DAG) GetRoots() []string {
	var roots []string
	for name, degree := range d.InDegree {
		if degree == 0 {
			roots = append(roots, name)
		}
	}
	return roots
}

// GetDependencies returns the direct dependencies of a task.
func (d *DAG) GetDependencies(taskName string) []string {
	return d.Edges[taskName]
}

// GetDependents returns tasks that directly depend on the given task.
func (d *DAG) GetDependents(taskName string) []string {
	return d.ReverseEdges[taskName]
}

// Size returns the number of tasks in the DAG.
func (d *DAG) Size() int {
	return len(d.Nodes)
}

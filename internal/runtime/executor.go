package runtime

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/adityaraj/agentflow/internal/config"
	"github.com/adityaraj/agentflow/internal/planner"
	"github.com/adityaraj/agentflow/internal/state"
	"github.com/adityaraj/agentflow/internal/ui"
)

// Executor runs tasks according to an execution plan.
type Executor struct {
	registry    *AgentRegistry
	store       *state.Store
	outputs     map[string]string // Task outputs for template expansion
	outputsMu   sync.RWMutex      // Protects outputs map
	verbose     bool
	writer      io.Writer // Output writer for logs
	parallel    bool      // Enable parallel execution
	maxParallel int       // Max concurrent tasks (0 = unlimited)
}

// ExecutorConfig holds configuration for creating an Executor.
type ExecutorConfig struct {
	Registry    *AgentRegistry
	Store       *state.Store
	Writer      io.Writer
	Verbose     bool
	Parallel    bool
	MaxParallel int
}

// NewExecutor creates a new Executor with the given registry and store.
func NewExecutor(registry *AgentRegistry, store *state.Store, writer io.Writer, verbose bool) *Executor {
	return &Executor{
		registry:    registry,
		store:       store,
		outputs:     make(map[string]string),
		verbose:     verbose,
		writer:      writer,
		parallel:    false,
		maxParallel: 0,
	}
}

// NewExecutorWithConfig creates a new Executor with full configuration.
func NewExecutorWithConfig(cfg ExecutorConfig) *Executor {
	return &Executor{
		registry:    cfg.Registry,
		store:       cfg.Store,
		outputs:     make(map[string]string),
		verbose:     cfg.Verbose,
		writer:      cfg.Writer,
		parallel:    cfg.Parallel,
		maxParallel: cfg.MaxParallel,
	}
}

// Execute runs all tasks in the execution plan.
// Uses parallel execution if enabled, otherwise sequential.
func (e *Executor) Execute(ctx context.Context, plan *planner.ExecutionPlan) (*state.RunResult, error) {
	if e.parallel {
		return e.executeParallel(ctx, plan)
	}
	return e.executeSequential(ctx, plan)
}

// executeSequential runs all tasks in the execution plan sequentially.
// Stops on the first failure and returns the error.
func (e *Executor) executeSequential(ctx context.Context, plan *planner.ExecutionPlan) (*state.RunResult, error) {
	runResult := &state.RunResult{
		RunID:     e.store.RunID(),
		StartTime: time.Now(),
		Tasks:     make([]state.TaskResult, 0, len(plan.Tasks)),
		Success:   true,
	}

	totalTasks := len(plan.Tasks)
	for i, execTask := range plan.Tasks {
		// Print task start with colors
		ui.PrintTaskStart(i+1, totalTasks, execTask.Name, execTask.AgentName, execTask.Tool, execTask.Model)
		ui.PrintTaskRunning()

		taskResult, err := e.executeTask(ctx, execTask)
		if err != nil {
			runResult.Tasks = append(runResult.Tasks, *taskResult)
			runResult.Success = false
			runResult.EndTime = time.Now()
			e.store.SaveRunResult(runResult)
			return runResult, err
		}

		runResult.Tasks = append(runResult.Tasks, *taskResult)
	}

	runResult.EndTime = time.Now()
	e.store.SaveRunResult(runResult)

	return runResult, nil
}

// executeParallel runs tasks in parallel using execution levels.
// Tasks in the same level run concurrently, levels run sequentially.
func (e *Executor) executeParallel(ctx context.Context, plan *planner.ExecutionPlan) (*state.RunResult, error) {
	runResult := &state.RunResult{
		RunID:     e.store.RunID(),
		StartTime: time.Now(),
		Tasks:     make([]state.TaskResult, 0, len(plan.Tasks)),
		Success:   true,
	}

	// Build task lookup map
	taskMap := make(map[string]planner.ExecutionTask)
	for _, t := range plan.Tasks {
		taskMap[t.Name] = t
	}

	// Build execution levels
	levels := planner.BuildExecutionLevels(plan.DAG)
	totalTasks := len(plan.Tasks)
	completedTasks := 0

	var resultsMu sync.Mutex

	for _, level := range levels {
		// Determine how many tasks to run concurrently
		maxConcurrent := len(level.Tasks)
		if e.maxParallel > 0 && maxConcurrent > e.maxParallel {
			maxConcurrent = e.maxParallel
		}

		// Semaphore for limiting concurrency
		sem := make(chan struct{}, maxConcurrent)
		var wg sync.WaitGroup

		// Channel to collect errors
		errChan := make(chan error, len(level.Tasks))

		for _, taskName := range level.Tasks {
			execTask := taskMap[taskName]

			wg.Add(1)
			go func(task planner.ExecutionTask) {
				defer wg.Done()

				// Acquire semaphore
				sem <- struct{}{}
				defer func() { <-sem }()

				// Check if context is cancelled
				if ctx.Err() != nil {
					errChan <- ctx.Err()
					return
				}

				completedTasks++
				// Print task start
				ui.PrintTaskStart(completedTasks, totalTasks, task.Name, task.AgentName, task.Tool, task.Model)
				ui.PrintTaskRunning()

				// Execute the task
				taskResult, err := e.executeTask(ctx, task)

				resultsMu.Lock()
				runResult.Tasks = append(runResult.Tasks, *taskResult)
				resultsMu.Unlock()

				if err != nil {
					errChan <- err
				}
			}(execTask)
		}

		// Wait for all tasks in this level to complete
		wg.Wait()
		close(errChan)

		// Check for errors
		var firstErr error
		for err := range errChan {
			if firstErr == nil {
				firstErr = err
			}
			runResult.Success = false
		}

		if firstErr != nil {
			runResult.EndTime = time.Now()
			e.store.SaveRunResult(runResult)
			return runResult, firstErr
		}
	}

	runResult.EndTime = time.Now()
	e.store.SaveRunResult(runResult)

	return runResult, nil
}

// executeTask executes a single task and returns its result.
func (e *Executor) executeTask(ctx context.Context, execTask planner.ExecutionTask) (*state.TaskResult, error) {
	// Get the agent adapter
	agent := e.registry.Get(execTask.Tool)
	if agent == nil {
		taskResult := state.NewTaskResult(execTask.Name, execTask.AgentName, execTask.Tool, execTask.Model, "")
		taskResult.Complete("", fmt.Sprintf("no adapter for tool %q", execTask.Tool), 1, false)
		e.store.SaveTaskResult(taskResult)
		ui.PrintTaskStatus("Failed", false, "0s")
		return taskResult, fmt.Errorf("no adapter registered for tool %q", execTask.Tool)
	}

	// Expand template variables in prompt
	e.outputsMu.RLock()
	expandedPrompt := config.ExpandPrompt(execTask.Prompt, e.outputs)
	e.outputsMu.RUnlock()

	// Create task for execution
	task := Task{
		Name:   execTask.Name,
		Agent:  execTask.AgentName,
		Tool:   execTask.Tool,
		Model:  execTask.Model,
		Prompt: expandedPrompt,
		Write:  execTask.Write,
	}

	// Create result tracker
	taskResult := state.NewTaskResult(
		execTask.Name,
		execTask.AgentName,
		execTask.Tool,
		execTask.Model,
		expandedPrompt,
	)

	// Execute the task
	result, err := agent.Run(ctx, task)
	if err != nil {
		taskResult.Complete("", err.Error(), 1, false)
		e.store.SaveTaskResult(taskResult)
		ui.PrintTaskStatus("Failed", false, taskResult.Duration)
		if e.verbose {
			fmt.Fprintf(e.writer, "  %sError:%s %s\n", ui.Dim, ui.Reset, err)
		}
		return taskResult, fmt.Errorf("task %q failed: %w", execTask.Name, err)
	}

	// Complete the task result
	taskResult.Complete(result.Stdout, result.Stderr, result.ExitCode, result.Success)

	// Save task result
	if err := e.store.SaveTaskResult(taskResult); err != nil {
		ui.Warning("Failed to save result: %s", err)
	}

	// Store output for template expansion in dependent tasks
	e.outputsMu.Lock()
	e.outputs[execTask.Name] = result.Stdout
	e.outputsMu.Unlock()

	if result.Success {
		ui.PrintTaskStatus("Success", true, taskResult.Duration)
	} else {
		ui.PrintTaskStatus("Failed", false, taskResult.Duration)
		return taskResult, fmt.Errorf("task %q failed with exit code %d", execTask.Name, result.ExitCode)
	}

	if e.verbose && result.Stdout != "" {
		// Show first few lines of output in verbose mode
		fmt.Fprintf(e.writer, "  %sOutput (truncated):%s\n", ui.Dim, ui.Reset)
		lines := truncateLines(result.Stdout, 5)
		for _, line := range lines {
			fmt.Fprintf(e.writer, "    %s%s%s\n", ui.Dim, line, ui.Reset)
		}
	}

	return taskResult, nil
}

// truncateLines returns the first n lines of text.
func truncateLines(text string, n int) []string {
	var lines []string
	start := 0
	count := 0
	for i, c := range text {
		if c == '\n' {
			lines = append(lines, text[start:i])
			start = i + 1
			count++
			if count >= n {
				lines = append(lines, "...")
				break
			}
		}
	}
	if count < n && start < len(text) {
		lines = append(lines, text[start:])
	}
	return lines
}

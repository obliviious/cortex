package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/adityaraj/agentflow/internal/config"
	"github.com/adityaraj/agentflow/internal/planner"
	"github.com/adityaraj/agentflow/internal/runtime"
	"github.com/adityaraj/agentflow/internal/runtime/adapters/claude"
	"github.com/adityaraj/agentflow/internal/runtime/adapters/opencode"
	"github.com/adityaraj/agentflow/internal/state"
	"github.com/adityaraj/agentflow/internal/ui"
	"github.com/adityaraj/agentflow/internal/webhook"
)

// Version info - set by ldflags during build
var (
	version   = "0.2.0"
	buildTime = "unknown"
)

var (
	configFiles []string
	verbose     bool
	streamLogs  bool
	noStream    bool
	noColor     bool
	compact     bool
	parallel    bool
	sequential  bool
	maxParallel int
	fullOutput  bool
	interactive bool
)

func main() {
	versionStr := version
	if buildTime != "unknown" {
		versionStr = fmt.Sprintf("%s (built %s)", version, buildTime)
	}

	rootCmd := &cobra.Command{
		Use:     "cortex",
		Short:   "AI agent orchestrator",
		Long:    "Cortex orchestrates AI agent workflows defined in YAML.",
		Version: versionStr,
	}

	// Run command
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Execute the Cortexfile workflow",
		Long:  "Loads and executes tasks defined in Cortexfile.yml",
		RunE:  runWorkflow,
	}

	runCmd.Flags().StringArrayVarP(&configFiles, "file", "f", nil, "Path to Cortexfile(s) - supports multiple files and glob patterns")
	runCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	runCmd.Flags().BoolVarP(&streamLogs, "stream", "s", true, "Stream real-time logs from agents (default: on)")
	runCmd.Flags().BoolVar(&noStream, "no-stream", false, "Disable real-time streaming")
	runCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	runCmd.Flags().BoolVar(&compact, "compact", false, "Use compact output (no banner)")
	runCmd.Flags().BoolVar(&parallel, "parallel", false, "Enable parallel execution (default: on)")
	runCmd.Flags().BoolVar(&sequential, "sequential", false, "Force sequential execution")
	runCmd.Flags().IntVar(&maxParallel, "max-parallel", 0, "Max concurrent tasks (0 = use config default)")
	runCmd.Flags().BoolVar(&fullOutput, "full", false, "Show full output (default: summary only)")
	runCmd.Flags().BoolVarP(&interactive, "interactive", "i", true, "Enable interactive mode with Ctrl+O toggle")

	// Validate command
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate the Cortexfile without running",
		Long:  "Checks the Cortexfile for errors without executing tasks",
		RunE:  validateConfig,
	}

	var validateFile string
	validateCmd.Flags().StringVarP(&validateFile, "file", "f", "", "Path to Cortexfile (default: auto-detect)")

	// Sessions command
	sessionsCmd := &cobra.Command{
		Use:   "sessions",
		Short: "List previous run sessions",
		Long:  "Lists all previous run sessions stored in ~/.cortex/sessions/",
		RunE:  listSessions,
	}

	var sessionProject string
	var sessionLimit int
	var sessionFailed bool

	sessionsCmd.Flags().StringVar(&sessionProject, "project", "", "Filter by project name")
	sessionsCmd.Flags().IntVar(&sessionLimit, "limit", 10, "Maximum number of sessions to show")
	sessionsCmd.Flags().BoolVar(&sessionFailed, "failed", false, "Show only failed sessions")

	// Init command - create template files
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Cortexfile in the current directory",
		Long:  "Creates a template Cortexfile.yml that you can customize for your project",
		RunE:  initCortexfile,
	}

	var initMinimal bool
	var initMaster bool
	var initForce bool

	initCmd.Flags().BoolVar(&initMinimal, "minimal", false, "Create a minimal template")
	initCmd.Flags().BoolVar(&initMaster, "master", false, "Create a MasterCortex.yml instead")
	initCmd.Flags().BoolVar(&initForce, "force", false, "Overwrite existing file")

	// Master command - run MasterCortex.yml
	masterCmd := &cobra.Command{
		Use:   "master",
		Short: "Run workflows defined in MasterCortex.yml",
		Long:  "Executes multiple Cortexfiles as defined in MasterCortex.yml",
		RunE:  runMasterWorkflow,
	}

	var masterFile string
	var masterParallel bool
	var masterSequential bool

	masterCmd.Flags().StringVarP(&masterFile, "file", "f", "", "Path to MasterCortex.yml (default: auto-detect)")
	masterCmd.Flags().BoolVar(&masterParallel, "parallel", false, "Force parallel execution")
	masterCmd.Flags().BoolVar(&masterSequential, "sequential", false, "Force sequential execution")
	masterCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	masterCmd.Flags().BoolVar(&compact, "compact", false, "Use compact output")

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(sessionsCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(masterCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runWorkflow(cmd *cobra.Command, args []string) error {
	// Handle color settings
	if noColor {
		ui.SetColorsEnabled(false)
	}

	// Print banner
	if compact {
		ui.PrintCompactBanner(version)
	} else {
		ui.PrintBanner(version)
	}

	// Resolve config files (supports multiple files and globs)
	configPaths, err := resolveConfigFiles()
	if err != nil {
		ui.Error("Failed to resolve config files: %s", err)
		return err
	}

	if len(configPaths) == 0 {
		ui.Error("No Cortexfile found")
		return fmt.Errorf("no Cortexfile found")
	}

	// Run each config file
	var allSuccess = true
	var totalTasks int
	var successfulRuns int

	for i, configPath := range configPaths {
		if len(configPaths) > 1 {
			ui.PrintDivider()
			fmt.Printf("\n%s[%d/%d]%s Running: %s%s%s\n\n",
				ui.Dim, i+1, len(configPaths), ui.Reset,
				ui.Bold, configPath, ui.Reset)
		}

		success, tasks, err := runSingleConfig(cmd, configPath)
		if err != nil {
			ui.Error("Config %s failed: %s", configPath, err)
			allSuccess = false
		} else if success {
			successfulRuns++
		} else {
			allSuccess = false
		}
		totalTasks += tasks
	}

	// Print aggregate summary for multiple configs
	if len(configPaths) > 1 {
		ui.PrintDivider()
		if allSuccess {
			fmt.Printf("\n  %s%s All %d configs completed successfully (%d tasks)%s\n\n",
				ui.Bold, ui.Green, len(configPaths), totalTasks, ui.Reset)
		} else {
			fmt.Printf("\n  %s%s %d/%d configs completed (%d tasks)%s\n\n",
				ui.Bold, ui.Red, successfulRuns, len(configPaths), totalTasks, ui.Reset)
		}
	}

	if !allSuccess {
		return fmt.Errorf("workflow completed with failures")
	}
	return nil
}

func runSingleConfig(cmd *cobra.Command, configPath string) (bool, int, error) {
	// Load global config
	globalCfg, err := config.LoadGlobalConfig()
	if err != nil {
		ui.Warning("Failed to load global config: %s", err)
		globalCfg = &config.GlobalConfig{
			Settings: config.DefaultSettings(),
		}
	}

	// Load local config from specified path
	ui.Info("Loading %s", configPath)
	localCfg, err := config.LoadConfig(configPath)
	if err != nil {
		return false, 0, fmt.Errorf("failed to load config: %w", err)
	}

	ui.Info("Validating configuration...")
	if err := config.ValidateWithFile(localCfg, configPath); err != nil {
		return false, 0, err
	}

	// Build CLI settings override
	cliSettings := &config.SettingsConfig{}
	if cmd.Flags().Changed("max-parallel") {
		cliSettings.MaxParallel = maxParallel
	}
	if cmd.Flags().Changed("verbose") {
		cliSettings.Verbose = verbose
	}
	// Stream is on by default, --no-stream disables it
	cliSettings.Stream = streamLogs && !noStream

	// Merge configs: CLI > local > global
	merged := config.MergeConfigs(globalCfg, localCfg, cliSettings)

	// Handle parallel execution flags
	// Default is parallel ON (from global config)
	useParallel := merged.Settings.Parallel
	if cmd.Flags().Changed("parallel") {
		useParallel = parallel
	}
	if sequential {
		useParallel = false
	}

	// Build execution plan
	ui.Info("Building execution plan...")
	plan, err := planner.BuildPlan(localCfg)
	if err != nil {
		ui.Error("Failed to build plan: %s", err)
		return false, 0, err
	}

	// Show execution mode
	if useParallel {
		levels := planner.BuildExecutionLevels(plan.DAG)
		maxPar := planner.MaxParallelism(levels)
		effectiveMax := merged.Settings.MaxParallel
		if effectiveMax > maxPar {
			effectiveMax = maxPar
		}
		ui.Info("Parallel execution: %s%d%s levels, up to %s%d%s concurrent tasks",
			ui.BrightCyan, len(levels), ui.Reset,
			ui.BrightCyan, effectiveMax, ui.Reset)
	} else {
		ui.Info("Sequential execution mode")
	}

	// Convert plan to TaskInfo for display
	taskInfos := make([]ui.TaskInfo, len(plan.Tasks))
	for i, t := range plan.Tasks {
		taskInfos[i] = ui.TaskInfo{
			Name:         t.Name,
			Agent:        t.AgentName,
			Tool:         t.Tool,
			Model:        t.Model,
			Dependencies: t.Dependencies,
		}
	}
	ui.PrintExecutionPlan(taskInfos)

	// Set up state store
	cwd, err := os.Getwd()
	if err != nil {
		ui.Error("Failed to get working directory: %s", err)
		return false, 0, err
	}

	store, err := state.NewStore(cwd)
	if err != nil {
		ui.Error("Failed to create state store: %s", err)
		return false, 0, err
	}

	// Print session info
	ui.PrintSessionInfo(store.RunID(), store.RunDir())

	// Set up webhook manager
	webhookMgr := webhook.NewManager(merged.Webhooks)
	if webhookMgr.HasWebhooks() {
		ui.Info("Webhooks configured: %d", webhookMgr.Count())
	}

	// Send run_start event
	projectName := filepath.Base(cwd)
	webhookMgr.Send(webhook.NewRunStartEvent(store.RunID(), projectName))

	// Set up agent registry
	registry := runtime.NewAgentRegistry()

	claudeAdapter := claude.New()
	claudeAdapter.SetStreamLogs(merged.Settings.Stream)
	registry.Register("claude-code", claudeAdapter)

	opencodeAdapter := opencode.New()
	opencodeAdapter.SetStreamLogs(merged.Settings.Stream)
	registry.Register("opencode", opencodeAdapter)

	// Create executor with config
	executor := runtime.NewExecutorWithConfig(runtime.ExecutorConfig{
		Registry:    registry,
		Store:       store,
		Writer:      os.Stdout,
		Verbose:     merged.Settings.Verbose,
		Parallel:    useParallel,
		MaxParallel: merged.Settings.MaxParallel,
	})

	// Set up context with cancellation on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Printf("\n%s⚠ Received interrupt, cancelling...%s\n", ui.BrightYellow, ui.Reset)
		cancel()
	}()

	// Execute the plan
	ui.PrintDivider()
	fmt.Printf("%sRunning tasks...%s\n", ui.Bold, ui.Reset)

	startTime := time.Now()
	result, err := executor.Execute(ctx, plan)
	duration := time.Since(startTime)

	// Wait for pending webhooks
	defer webhookMgr.Wait()

	// Send run_complete event
	webhookMgr.Send(webhook.NewRunCompleteEvent(
		store.RunID(),
		projectName,
		len(result.Tasks),
		duration,
		result.Success,
	))

	if err != nil {
		ui.PrintSummary(false, store.RunDir())
		return false, len(result.Tasks), err
	}

	// Print summary
	ui.PrintSummary(result.Success, store.RunDir())

	return result.Success, len(result.Tasks), nil
}

func validateConfig(cmd *cobra.Command, args []string) error {
	ui.PrintCompactBanner(version)

	cfg, configPath, err := loadConfig()
	if err != nil {
		ui.Error("Validation failed: %s", err)
		return err
	}

	// Validate with file path for better error messages
	if err := config.ValidateWithFile(cfg, configPath); err != nil {
		ui.Error("Validation failed:\n%s", err)
		return err
	}

	// Build plan to verify DAG is valid
	plan, err := planner.BuildPlan(cfg)
	if err != nil {
		ui.Error("Plan validation failed: %s", err)
		return err
	}

	ui.Success("Configuration is valid!")
	fmt.Printf("  %sAgents:%s %d\n", ui.Dim, ui.Reset, len(cfg.Agents))
	fmt.Printf("  %sTasks:%s  %d\n", ui.Dim, ui.Reset, len(cfg.Tasks))
	fmt.Println()

	// Show execution levels for parallel info
	levels := planner.BuildExecutionLevels(plan.DAG)
	fmt.Printf("  %sExecution Levels:%s %d\n", ui.Dim, ui.Reset, len(levels))
	fmt.Printf("  %sMax Parallelism:%s  %d\n", ui.Dim, ui.Reset, planner.MaxParallelism(levels))
	fmt.Println()

	// Convert plan to TaskInfo for display
	taskInfos := make([]ui.TaskInfo, len(plan.Tasks))
	for i, t := range plan.Tasks {
		taskInfos[i] = ui.TaskInfo{
			Name:         t.Name,
			Agent:        t.AgentName,
			Tool:         t.Tool,
			Model:        t.Model,
			Dependencies: t.Dependencies,
		}
	}
	ui.PrintExecutionPlan(taskInfos)

	return nil
}

func listSessions(cmd *cobra.Command, args []string) error {
	project, _ := cmd.Flags().GetString("project")
	limit, _ := cmd.Flags().GetInt("limit")
	failedOnly, _ := cmd.Flags().GetBool("failed")

	sessions, err := state.ListSessions(state.SessionFilter{
		Project:    project,
		Limit:      limit,
		FailedOnly: failedOnly,
	})

	if err != nil {
		ui.Error("Failed to list sessions: %s", err)
		return err
	}

	if len(sessions) == 0 {
		fmt.Printf("%sNo sessions found.%s\n", ui.Dim, ui.Reset)
		if project != "" {
			fmt.Printf("%sTry without --project filter.%s\n", ui.Dim, ui.Reset)
		}
		return nil
	}

	fmt.Printf("%s%sSessions%s (%d):\n\n", ui.Bold, ui.BrightCyan, ui.Reset, len(sessions))

	for _, s := range sessions {
		// Status indicator
		statusIcon := fmt.Sprintf("%s✓%s", ui.BrightGreen, ui.Reset)
		if !s.Success {
			statusIcon = fmt.Sprintf("%s✗%s", ui.BrightRed, ui.Reset)
		}

		// Format time
		timeStr := s.StartTime.Format("2006-01-02 15:04:05")
		if s.StartTime.IsZero() {
			timeStr = "unknown"
		}

		// Duration
		durationStr := ""
		if s.Duration > 0 {
			durationStr = fmt.Sprintf(" (%s)", state.FormatDuration(s.Duration))
		}

		fmt.Printf("  %s %s%s%s %s%s%s\n",
			statusIcon,
			ui.Bold, s.RunID, ui.Reset,
			ui.Dim, timeStr, ui.Reset,
		)
		fmt.Printf("      %sProject:%s %s  %sTasks:%s %d%s\n",
			ui.Dim, ui.Reset, s.Project,
			ui.Dim, ui.Reset, s.TaskCount,
			durationStr,
		)
	}

	fmt.Println()
	return nil
}

func loadConfig() (*config.AgentflowConfig, string, error) {
	paths, err := resolveConfigFiles()
	if err != nil {
		return nil, "", err
	}

	if len(paths) == 0 {
		return nil, "", fmt.Errorf("no Cortexfile found")
	}

	// For now, use the first file (multiple files will be handled separately)
	path := paths[0]

	ui.Info("Loading %s", path)

	cfg, err := config.LoadConfig(path)
	if err != nil {
		return nil, path, fmt.Errorf("failed to load config: %w", err)
	}

	ui.Info("Validating configuration...")
	if err := config.ValidateWithFile(cfg, path); err != nil {
		return nil, path, err
	}

	return cfg, path, nil
}

// resolveConfigFiles expands glob patterns and returns all matching config files
func resolveConfigFiles() ([]string, error) {
	if len(configFiles) == 0 {
		// Auto-detect in current directory
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}

		path, err := config.FindCortexfile(cwd)
		if err != nil {
			return nil, err
		}
		return []string{path}, nil
	}

	var result []string
	seen := make(map[string]bool)

	for _, pattern := range configFiles {
		// Check if it's a glob pattern
		if containsGlobChars(pattern) {
			matches, err := filepath.Glob(pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
			}
			for _, m := range matches {
				abs, _ := filepath.Abs(m)
				if !seen[abs] {
					seen[abs] = true
					result = append(result, abs)
				}
			}
		} else {
			// Regular file path
			abs, _ := filepath.Abs(pattern)
			if !seen[abs] {
				seen[abs] = true
				result = append(result, abs)
			}
		}
	}

	// Sort for consistent ordering
	sort.Strings(result)
	return result, nil
}

// containsGlobChars checks if a string contains glob pattern characters
func containsGlobChars(s string) bool {
	for _, c := range s {
		if c == '*' || c == '?' || c == '[' {
			return true
		}
	}
	return false
}

// initCortexfile creates a template Cortexfile or MasterCortex file
func initCortexfile(cmd *cobra.Command, args []string) error {
	minimal, _ := cmd.Flags().GetBool("minimal")
	master, _ := cmd.Flags().GetBool("master")
	force, _ := cmd.Flags().GetBool("force")

	var filename string
	var content string

	if master {
		filename = "MasterCortex.yml"
		content = config.MasterCortexTemplate
	} else if minimal {
		filename = "Cortexfile.yml"
		content = config.MinimalCortexfileTemplate
	} else {
		filename = "Cortexfile.yml"
		content = config.CortexfileTemplate
	}

	// Check if file already exists
	if _, err := os.Stat(filename); err == nil && !force {
		ui.Error("%s already exists. Use --force to overwrite.", filename)
		return fmt.Errorf("file exists")
	}

	// Write the file
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		ui.Error("Failed to create %s: %s", filename, err)
		return err
	}

	ui.Success("Created %s", filename)
	fmt.Printf("\n  %sNext steps:%s\n", ui.Bold, ui.Reset)
	fmt.Printf("  1. Edit %s to define your workflow\n", filename)
	if master {
		fmt.Printf("  2. Run %scortex master%s to execute\n", ui.Bold, ui.Reset)
	} else {
		fmt.Printf("  2. Run %scortex validate%s to check your config\n", ui.Bold, ui.Reset)
		fmt.Printf("  3. Run %scortex run%s to execute\n", ui.Bold, ui.Reset)
	}
	fmt.Println()

	return nil
}

// runMasterWorkflow executes workflows defined in MasterCortex.yml
func runMasterWorkflow(cmd *cobra.Command, args []string) error {
	// Handle color settings
	if noColor {
		ui.SetColorsEnabled(false)
	}

	// Print banner
	if compact {
		ui.PrintCompactBanner(version)
	} else {
		ui.PrintBanner(version)
	}

	// Find MasterCortex file
	masterFile, _ := cmd.Flags().GetString("file")
	var masterPath string
	var err error

	if masterFile != "" {
		masterPath = masterFile
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			ui.Error("Failed to get working directory: %s", err)
			return err
		}
		masterPath, err = config.FindMasterCortex(cwd)
		if err != nil {
			ui.Error("No MasterCortex.yml found. Create one with: cortex init --master")
			return err
		}
	}

	ui.Info("Loading %s", masterPath)

	// Load master config
	masterCfg, err := config.LoadMasterConfig(masterPath)
	if err != nil {
		ui.Error("Failed to load master config: %s", err)
		return err
	}

	// Validate
	if err := config.ValidateMasterConfig(masterCfg); err != nil {
		ui.Error("Invalid master config: %s", err)
		return err
	}

	// Resolve workflow paths
	baseDir := filepath.Dir(masterPath)
	workflows, err := config.ResolveWorkflowPaths(masterCfg, baseDir)
	if err != nil {
		ui.Error("Failed to resolve workflow paths: %s", err)
		return err
	}

	if len(workflows) == 0 {
		ui.Warning("No enabled workflows found")
		return nil
	}

	// Override mode from CLI flags
	forceParallel, _ := cmd.Flags().GetBool("parallel")
	forceSequential, _ := cmd.Flags().GetBool("sequential")

	mode := masterCfg.Mode
	if forceParallel {
		mode = "parallel"
	} else if forceSequential {
		mode = "sequential"
	}

	// Print execution info
	if masterCfg.Name != "" {
		fmt.Printf("  %s%s%s\n", ui.Bold+ui.Orange, masterCfg.Name, ui.Reset)
	}
	if masterCfg.Description != "" {
		fmt.Printf("  %s%s%s\n", ui.Dim, masterCfg.Description, ui.Reset)
	}
	ui.Info("Mode: %s, Workflows: %d", mode, len(workflows))
	fmt.Println()

	// Print workflow list
	fmt.Printf("  %s%sWorkflows%s\n", ui.Bold, ui.Orange, ui.Reset)
	fmt.Printf("  %s─────────%s\n", ui.Dim, ui.Reset)
	for i, w := range workflows {
		deps := ""
		if len(w.Needs) > 0 {
			deps = fmt.Sprintf(" %s← %v%s", ui.Dim, w.Needs, ui.Reset)
		}
		fmt.Printf("  %s%d.%s %s%s%s%s\n", ui.Orange, i+1, ui.Reset, ui.Bold, w.Name, ui.Reset, deps)
		fmt.Printf("     %s%s%s\n", ui.Dim, w.Path, ui.Reset)
	}
	fmt.Println()

	// Execute workflows
	startTime := time.Now()
	var results []workflowResult

	if mode == "parallel" {
		results = executeWorkflowsParallel(cmd, workflows, masterCfg)
	} else {
		results = executeWorkflowsSequential(cmd, workflows, masterCfg)
	}

	duration := time.Since(startTime)

	// Print summary
	ui.PrintDivider()

	successCount := 0
	totalTasks := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
		totalTasks += r.Tasks
	}

	if successCount == len(results) {
		fmt.Printf("\n  %s%s All %d workflows completed successfully%s\n", ui.Bold, ui.Green, len(results), ui.Reset)
	} else {
		fmt.Printf("\n  %s%s %d/%d workflows completed%s\n", ui.Bold, ui.Red, successCount, len(results), ui.Reset)
	}
	fmt.Printf("  %sTotal tasks: %d, Duration: %s%s\n\n", ui.Dim, totalTasks, duration.Round(time.Second), ui.Reset)

	if successCount < len(results) {
		return fmt.Errorf("master workflow completed with failures")
	}
	return nil
}

type workflowResult struct {
	Name    string
	Success bool
	Tasks   int
	Error   error
}

func executeWorkflowsSequential(cmd *cobra.Command, workflows []config.WorkflowEntry, masterCfg *config.MasterConfig) []workflowResult {
	results := make([]workflowResult, 0, len(workflows))
	completed := make(map[string]bool)

	for _, w := range workflows {
		// Check dependencies
		canRun := true
		for _, dep := range w.Needs {
			if !completed[dep] {
				canRun = false
				break
			}
		}

		if !canRun {
			ui.Warning("Skipping %s: dependencies not met", w.Name)
			results = append(results, workflowResult{Name: w.Name, Success: false, Error: fmt.Errorf("dependencies not met")})
			continue
		}

		ui.PrintDivider()
		fmt.Printf("\n%s[%d/%d]%s %s%s%s\n\n",
			ui.Dim, len(results)+1, len(workflows), ui.Reset,
			ui.Bold+ui.Orange, w.Name, ui.Reset)

		// Set configFiles for this workflow
		configFiles = []string{w.Path}

		success, tasks, err := runSingleConfig(cmd, w.Path)
		results = append(results, workflowResult{
			Name:    w.Name,
			Success: success,
			Tasks:   tasks,
			Error:   err,
		})

		if success {
			completed[w.Name] = true
		} else if masterCfg.StopOnError != nil && *masterCfg.StopOnError {
			ui.Error("Stopping due to error in %s", w.Name)
			break
		}
	}

	return results
}

func executeWorkflowsParallel(cmd *cobra.Command, workflows []config.WorkflowEntry, masterCfg *config.MasterConfig) []workflowResult {
	// For parallel execution with dependencies, we need to build execution levels
	// similar to task execution. For simplicity, we'll run all without deps first,
	// then those with deps.

	results := make([]workflowResult, len(workflows))
	var wg sync.WaitGroup
	var mu sync.Mutex
	completed := make(map[string]bool)

	// First pass: run workflows without dependencies
	sem := make(chan struct{}, maxOrDefault(masterCfg.MaxParallel, len(workflows)))

	for i, w := range workflows {
		if len(w.Needs) > 0 {
			continue // Skip workflows with dependencies for now
		}

		wg.Add(1)
		go func(idx int, workflow config.WorkflowEntry) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			fmt.Printf("\n%s[%s]%s Starting...\n", ui.Orange, workflow.Name, ui.Reset)

			success, tasks, err := runSingleConfig(cmd, workflow.Path)

			mu.Lock()
			results[idx] = workflowResult{
				Name:    workflow.Name,
				Success: success,
				Tasks:   tasks,
				Error:   err,
			}
			if success {
				completed[workflow.Name] = true
			}
			mu.Unlock()

			if success {
				fmt.Printf("%s[%s]%s %sCompleted%s\n", ui.Orange, workflow.Name, ui.Reset, ui.Green, ui.Reset)
			} else {
				fmt.Printf("%s[%s]%s %sFailed%s\n", ui.Orange, workflow.Name, ui.Reset, ui.Red, ui.Reset)
			}
		}(i, w)
	}

	wg.Wait()

	// Second pass: run workflows with dependencies (sequentially for simplicity)
	for i, w := range workflows {
		if len(w.Needs) == 0 {
			continue // Already ran
		}

		// Check dependencies
		canRun := true
		for _, dep := range w.Needs {
			if !completed[dep] {
				canRun = false
				break
			}
		}

		if !canRun {
			results[i] = workflowResult{Name: w.Name, Success: false, Error: fmt.Errorf("dependencies not met")}
			continue
		}

		fmt.Printf("\n%s[%s]%s Starting (deps: %v)...\n", ui.Orange, w.Name, ui.Reset, w.Needs)

		success, tasks, err := runSingleConfig(cmd, w.Path)
		results[i] = workflowResult{
			Name:    w.Name,
			Success: success,
			Tasks:   tasks,
			Error:   err,
		}

		if success {
			completed[w.Name] = true
			fmt.Printf("%s[%s]%s %sCompleted%s\n", ui.Orange, w.Name, ui.Reset, ui.Green, ui.Reset)
		} else {
			fmt.Printf("%s[%s]%s %sFailed%s\n", ui.Orange, w.Name, ui.Reset, ui.Red, ui.Reset)
		}
	}

	return results
}

func maxOrDefault(val, def int) int {
	if val > 0 {
		return val
	}
	return def
}

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
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
	configFile  string
	verbose     bool
	streamLogs  bool
	noColor     bool
	compact     bool
	parallel    bool
	sequential  bool
	maxParallel int
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "cortex",
		Short:   "AI agent orchestrator",
		Long:    "Cortex orchestrates AI agent workflows defined in YAML.",
		Version: version,
	}

	// Run command
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Execute the Cortexfile workflow",
		Long:  "Loads and executes tasks defined in Cortexfile.yml",
		RunE:  runWorkflow,
	}

	runCmd.Flags().StringVarP(&configFile, "file", "f", "", "Path to Cortexfile (default: auto-detect)")
	runCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	runCmd.Flags().BoolVarP(&streamLogs, "stream", "s", false, "Stream real-time logs from agents")
	runCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	runCmd.Flags().BoolVar(&compact, "compact", false, "Use compact output (no banner)")
	runCmd.Flags().BoolVar(&parallel, "parallel", false, "Enable parallel execution (default: on)")
	runCmd.Flags().BoolVar(&sequential, "sequential", false, "Force sequential execution")
	runCmd.Flags().IntVar(&maxParallel, "max-parallel", 0, "Max concurrent tasks (0 = use config default)")

	// Validate command
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate the Cortexfile without running",
		Long:  "Checks the Cortexfile for errors without executing tasks",
		RunE:  validateConfig,
	}

	validateCmd.Flags().StringVarP(&configFile, "file", "f", "", "Path to Cortexfile (default: auto-detect)")

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

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(sessionsCmd)

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

	// Load global config
	globalCfg, err := config.LoadGlobalConfig()
	if err != nil {
		ui.Warning("Failed to load global config: %s", err)
		globalCfg = &config.GlobalConfig{
			Settings: config.DefaultSettings(),
		}
	}

	// Find and load local config
	localCfg, configPath, err := loadConfig()
	if err != nil {
		ui.Error("Failed to load config: %s", err)
		return err
	}

	// Build CLI settings override
	cliSettings := &config.SettingsConfig{}
	if cmd.Flags().Changed("max-parallel") {
		cliSettings.MaxParallel = maxParallel
	}
	if cmd.Flags().Changed("verbose") {
		cliSettings.Verbose = verbose
	}
	if cmd.Flags().Changed("stream") {
		cliSettings.Stream = streamLogs
	}

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
		return err
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
		return err
	}

	store, err := state.NewStore(cwd)
	if err != nil {
		ui.Error("Failed to create state store: %s", err)
		return err
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
		return err
	}

	// Print summary
	ui.PrintSummary(result.Success, store.RunDir())

	if !result.Success {
		return fmt.Errorf("workflow completed with failures")
	}

	_ = configPath // Used for future error reporting
	return nil
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
	var path string
	var err error

	if configFile != "" {
		path = configFile
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, "", fmt.Errorf("failed to get working directory: %w", err)
		}

		path, err = config.FindCortexfile(cwd)
		if err != nil {
			return nil, "", err
		}
	}

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

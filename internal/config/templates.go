package config

// CortexfileTemplate is the default template for a new Cortexfile.yml
const CortexfileTemplate = `# Cortexfile.yml - Cortex Workflow Configuration
# Documentation: https://github.com/obliviious/cortex

# ============================================================================
# WORKING DIRECTORY (Optional)
# ============================================================================
# Set working directory for all agents. Can be absolute or relative path.
# workdir: /path/to/project

# ============================================================================
# AGENTS
# ============================================================================
# Define your agents. Each agent has a tool and optional model.
#
# Supported tools:
#   - claude-code : Claude AI via Claude Code CLI
#   - opencode    : OpenCode AI CLI
#   - shell       : Execute shell commands directly
#
# Models (for AI agents):
#   - sonnet      : Claude Sonnet (fast, cost-effective)
#   - opus        : Claude Opus (most capable)
#   - haiku       : Claude Haiku (fastest, cheapest)

agents:
  # AI agent using Claude Code
  analyzer:
    tool: claude-code
    model: sonnet

  # AI agent for code review
  reviewer:
    tool: claude-code
    model: sonnet

  # AI agent for implementation (can use different model)
  coder:
    tool: claude-code
    model: opus

  # Shell agent for running commands (no model needed)
  builder:
    tool: shell

  # Alternative: OpenCode AI agent
  # opencode-agent:
  #   tool: opencode
  #   model: sonnet

# ============================================================================
# TASKS
# ============================================================================
# Define your workflow tasks. Tasks can depend on other tasks using 'needs'.
#
# Task options:
#   - agent      : (required) Reference to agent name defined above
#   - prompt     : (AI agents) Inline prompt text
#   - prompt_file: (AI agents) Path to external prompt file
#   - command    : (shell agents) Shell command to execute
#   - needs      : Dependencies - single task or array of tasks
#   - write      : Allow file writes (default: false)
#
# Template variables:
#   Use {{outputs.task_name}} to reference output from a dependency task

tasks:
  # -------------------------------------------------------------------------
  # Shell task example - build the project
  # -------------------------------------------------------------------------
  build:
    agent: builder
    command: |
      echo "Building project..."
      make build 2>&1 || echo "Build command not found, skipping..."
      echo "Build complete!"

  # -------------------------------------------------------------------------
  # Shell task example - run tests
  # -------------------------------------------------------------------------
  test:
    agent: builder
    command: go test ./... -v
    needs: [build]

  # -------------------------------------------------------------------------
  # AI task - analyze codebase (inline prompt)
  # -------------------------------------------------------------------------
  analyze:
    agent: analyzer
    prompt: |
      Analyze the codebase structure and identify:
      1. Main components and their responsibilities
      2. Key dependencies and their versions
      3. Code architecture patterns used
      4. Potential areas for improvement

      Provide a concise, actionable summary.

  # -------------------------------------------------------------------------
  # AI task - code review
  # -------------------------------------------------------------------------
  review:
    agent: reviewer
    prompt: |
      Review the codebase for:
      1. Code quality issues
      2. Security vulnerabilities
      3. Performance concerns
      4. Best practices violations
      5. Test coverage gaps

      Provide specific, actionable recommendations.

  # -------------------------------------------------------------------------
  # AI task with dependencies - uses template variables
  # -------------------------------------------------------------------------
  implement:
    agent: coder
    needs: [analyze, review, test]
    write: true
    prompt: |
      Based on the analysis:
      {{outputs.analyze}}

      And the review findings:
      {{outputs.review}}

      Test results:
      {{outputs.test}}

      Implement the top 3 most important improvements.
      Focus on code quality and maintainability.

  # -------------------------------------------------------------------------
  # AI task using external prompt file
  # -------------------------------------------------------------------------
  # custom_task:
  #   agent: analyzer
  #   prompt_file: prompts/custom-analysis.md
  #   needs: [analyze]

# ============================================================================
# SETTINGS (Optional)
# ============================================================================
# Local settings that override global ~/.cortex/config.yml

settings:
  # Enable parallel execution of independent tasks (default: true)
  parallel: true

  # Maximum concurrent tasks (default: number of CPU cores)
  max_parallel: 4

  # Show verbose output including task details (default: false)
  verbose: false

  # Stream real-time output from agents (default: true)
  stream: true
`

// MasterCortexTemplate is the default template for a new MasterCortex.yml
const MasterCortexTemplate = `# MasterCortex.yml - Multi-Project Workflow Orchestration
# Documentation: https://github.com/obliviious/cortex
#
# Use this to orchestrate multiple Cortexfile workflows across projects.
# Run with: cortex master

# ============================================================================
# METADATA
# ============================================================================
# Name and description for this master workflow
name: multi-project-workflow
description: Orchestrates multiple Cortex workflows across projects

# ============================================================================
# EXECUTION MODE
# ============================================================================
# Mode: "sequential" or "parallel"
#   - sequential: Run workflows one after another
#   - parallel: Run independent workflows concurrently
mode: sequential

# Maximum parallel workflows (only used in parallel mode, 0 = unlimited)
max_parallel: 2

# Stop on first error (default: true for sequential, false for parallel)
stop_on_error: true

# ============================================================================
# GLOBAL VARIABLES (Optional)
# ============================================================================
# Variables available to all workflows (for future use)
variables:
  environment: development
  output_dir: ./results

# ============================================================================
# WORKFLOWS
# ============================================================================
# Define the workflows to run. Each workflow references a Cortexfile.

workflows:
  # -------------------------------------------------------------------------
  # Main project workflow
  # -------------------------------------------------------------------------
  - name: main-project
    path: ./Cortexfile.yml
    # workdir: ./main-project  # Optional: override working directory
    # enabled: true            # Optional: disable with false

  # -------------------------------------------------------------------------
  # Subdirectory workflows
  # -------------------------------------------------------------------------
  - name: backend
    path: ./backend/Cortexfile.yml
    workdir: ./backend

  - name: frontend
    path: ./frontend/Cortexfile.yml
    workdir: ./frontend
    needs: [backend]  # Wait for backend to complete first

  # -------------------------------------------------------------------------
  # Glob patterns for multiple similar projects
  # -------------------------------------------------------------------------
  # - name: microservices
  #   path: "./services/*/Cortexfile.yml"

  # -------------------------------------------------------------------------
  # Disabled workflow (kept for reference)
  # -------------------------------------------------------------------------
  # - name: experimental
  #   path: ./experimental/Cortexfile.yml
  #   enabled: false
`

// MinimalCortexfileTemplate is a minimal template for quick start
const MinimalCortexfileTemplate = `# Cortexfile.yml - Minimal Template
#
# Supported tools: claude-code, opencode, shell
# Run with: cortex run

agents:
  assistant:
    tool: claude-code
    model: sonnet

  builder:
    tool: shell

tasks:
  # Shell task example
  build:
    agent: builder
    command: echo "Hello from shell!"

  # AI task example
  main:
    agent: assistant
    needs: [build]
    write: true
    prompt: |
      Build output: {{outputs.build}}

      # Your prompt here
      Describe what you want the AI to do.
`

// GlobalConfigTemplate is the template for ~/.cortex/config.yml
const GlobalConfigTemplate = `# ~/.cortex/config.yml - Global Cortex Configuration
# This file contains default settings applied to all Cortex workflows.
# Settings can be overridden per-project in Cortexfile.yml or via CLI flags.

# ============================================================================
# DEFAULT AGENT SETTINGS
# ============================================================================
# These defaults apply to agents that don't specify these values.

defaults:
  # Default AI model (sonnet, opus, haiku)
  model: sonnet

  # Default tool (claude-code, opencode, shell)
  tool: claude-code

# ============================================================================
# EXECUTION SETTINGS
# ============================================================================

settings:
  # Enable parallel execution of independent tasks
  parallel: true

  # Maximum concurrent tasks (0 = use number of CPU cores)
  max_parallel: 0

  # Show verbose output
  verbose: false

  # Stream real-time output from agents
  stream: true

# ============================================================================
# WEBHOOKS (Optional)
# ============================================================================
# Send notifications to external services on workflow events.
#
# Supported events:
#   - run_start    : When a workflow run starts
#   - run_complete : When a workflow run completes
#   - task_start   : When a task starts
#   - task_complete: When a task completes
#   - *            : All events

# webhooks:
#   # Slack notification
#   - url: https://hooks.slack.com/services/YOUR/WEBHOOK/URL
#     events: [run_complete]
#     headers:
#       Content-Type: application/json
#
#   # Custom webhook for all events
#   - url: https://your-server.com/webhook
#     events: ["*"]
#     headers:
#       Authorization: Bearer YOUR_TOKEN
#       Content-Type: application/json
`

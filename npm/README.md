# Cortex

```
   ██████╗ ██████╗ ██████╗ ████████╗███████╗██╗  ██╗
  ██╔════╝██╔═══██╗██╔══██╗╚══██╔══╝██╔════╝╚██╗██╔╝
  ██║     ██║   ██║██████╔╝   ██║   █████╗   ╚███╔╝
  ██║     ██║   ██║██╔══██╗   ██║   ██╔══╝   ██╔██╗
  ╚██████╗╚██████╔╝██║  ██║   ██║   ███████╗██╔╝ ██╗
   ╚═════╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚══════╝╚═╝  ╚═╝

  ⚡ AI Agent Orchestrator ⚡
```

Cortex is a powerful CLI tool that orchestrates AI agent workflows defined in YAML. Run multiple AI agents in parallel, chain their outputs, and automate complex tasks.

## Features

- 🚀 **Parallel Execution** - Run independent tasks concurrently
- 🔗 **Task Dependencies** - Chain tasks with `needs` and pass outputs via templates
- 🤖 **Multi-Agent Support** - Use Claude Code, OpenCode, or other AI CLIs
- 📊 **Session Tracking** - View and manage past run sessions
- 🔔 **Webhooks** - Get notified on task completion/failure
- ⚙️ **Global Config** - Set defaults in `~/.cortex/config.yml`

## Installation

### Quick Install (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/adityaraj/cortex/main/install.sh | bash
```

### Homebrew (macOS/Linux)

```bash
brew tap adityaraj/tap
brew install cortex
```

### Go Install

```bash
go install github.com/adityaraj/cortex/cmd/agentflow@latest
```

### From Source

```bash
git clone https://github.com/adityaraj/cortex.git
cd cortex
make install
```

### Manual Download

Download the latest release for your platform from [GitHub Releases](https://github.com/adityaraj/cortex/releases).

## Quick Start

### 1. Create a Cortexfile

Create `Cortexfile.yml` in your project:

```yaml
agents:
  architect:
    tool: claude-code
    model: sonnet

  reviewer:
    tool: claude-code
    model: sonnet

tasks:
  analyze:
    agent: architect
    prompt: |
      Analyze the codebase structure and identify areas for improvement.
      Be concise and focus on actionable insights.

  review:
    agent: reviewer
    prompt: |
      Review the code for security issues and best practices.

  implement:
    agent: architect
    needs: [analyze, review]
    write: true
    prompt: |
      Based on the analysis and review:

      ## Analysis:
      {{outputs.analyze}}

      ## Review:
      {{outputs.review}}

      Implement the top priority improvement.
```

### 2. Run the Workflow

```bash
cortex run
```

### 3. View Past Sessions

```bash
cortex sessions
```

## Commands

| Command | Description |
|---------|-------------|
| `cortex run` | Execute the Cortexfile workflow |
| `cortex validate` | Validate configuration without running |
| `cortex sessions` | List previous run sessions |

### Run Options

```bash
cortex run [flags]

Flags:
  -f, --file string        Path to Cortexfile (default: auto-detect)
  -v, --verbose            Verbose output
  -s, --stream             Stream real-time logs from agents
      --parallel           Enable parallel execution (default: on)
      --sequential         Force sequential execution
      --max-parallel int   Max concurrent tasks (0 = CPU cores)
      --no-color           Disable colored output
      --compact            Minimal output (no banner)
```

### Sessions Options

```bash
cortex sessions [flags]

Flags:
      --project string   Filter by project name
      --limit int        Max sessions to show (default: 10)
      --failed           Show only failed sessions
```

## Configuration

### Cortexfile.yml

```yaml
# Agents define the AI tools to use
agents:
  my-agent:
    tool: claude-code    # or "opencode"
    model: sonnet        # optional: model override

# Tasks define the workflow
tasks:
  task-name:
    agent: my-agent      # Reference to agent
    prompt: |            # Inline prompt
      Your prompt here
    # OR
    prompt_file: prompts/task.md  # External file

    needs: [other-task]  # Dependencies (optional)
    write: true          # Allow file writes (default: false)

# Local settings (optional)
settings:
  parallel: true
  max_parallel: 4
```

### Global Config (~/.cortex/config.yml)

```yaml
# Default agent settings
defaults:
  model: sonnet
  tool: claude-code

# Execution settings
settings:
  parallel: true
  max_parallel: 4
  verbose: false
  stream: false

# Webhook notifications
webhooks:
  - url: https://hooks.slack.com/services/xxx
    events: [run_complete, task_failed]
    headers:
      Authorization: "Bearer token"
```

## Template Variables

Pass outputs between tasks using template variables:

```yaml
tasks:
  analyze:
    agent: architect
    prompt: Analyze the code...

  implement:
    agent: coder
    needs: [analyze]  # Must declare dependency
    prompt: |
      Based on this analysis:
      {{outputs.analyze}}

      Implement the changes.
```

## Webhooks

Configure webhooks to receive notifications:

```yaml
# In ~/.cortex/config.yml
webhooks:
  - url: https://your-webhook.com/endpoint
    events:
      - run_start
      - run_complete
      - task_start
      - task_complete
      - task_failed
    headers:
      Authorization: "Bearer your-token"
```

### Webhook Payload

```json
{
  "event": "task_complete",
  "timestamp": "2024-01-04T20:00:00Z",
  "run_id": "20240104-200000",
  "project": "my-project",
  "task": {
    "name": "analyze",
    "agent": "architect",
    "tool": "claude-code",
    "duration": "12.3s",
    "success": true
  }
}
```

## Session Storage

Run results are stored in `~/.cortex/sessions/<project>/run-<timestamp>/`:

```
~/.cortex/
├── config.yml          # Global config
└── sessions/
    └── my-project/
        └── run-20240104-200000/
            ├── run.json        # Run summary
            ├── analyze.json    # Task results
            └── review.json
```

## Supported Tools

| Tool | CLI Command | Description |
|------|-------------|-------------|
| `claude-code` | `claude` | Anthropic's Claude Code CLI |
| `opencode` | `opencode` | OpenCode CLI |

## Requirements

- One of the supported AI CLI tools installed
- Go 1.21+ (for building from source)

## License

MIT License - see [LICENSE](LICENSE)

import { useParams } from 'react-router-dom'

const CodeBlock = ({ children, title }) => (
  <div className="bg-dark-400 rounded-xl overflow-hidden my-6">
    {title && (
      <div className="flex items-center gap-2 px-4 py-3 bg-dark-300">
        <div className="w-3 h-3 rounded-full bg-red-400"></div>
        <div className="w-3 h-3 rounded-full bg-yellow-400"></div>
        <div className="w-3 h-3 rounded-full bg-green-400"></div>
        <span className="ml-4 text-cream-300 text-sm">{title}</span>
      </div>
    )}
    <pre className="p-4 text-cream-100 text-sm font-mono overflow-x-auto">
      <code>{children}</code>
    </pre>
  </div>
)

const docsContent = {
  'getting-started': {
    title: 'Quick Start',
    content: (
      <>
        <p className="text-lg text-dark-100 mb-6">
          Get started with Cortex in under 5 minutes. This guide will walk you through installation,
          creating your first Cortexfile, and running your first workflow.
        </p>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">1. Install Cortex</h2>
        <CodeBlock title="terminal">
{`npm install -g @insien/cortex-cli`}
        </CodeBlock>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">2. Create a Cortexfile</h2>
        <p className="text-dark-100 mb-4">
          Generate a template Cortexfile.yml in your project:
        </p>
        <CodeBlock title="terminal">
{`cortex init

# Or create a minimal template
cortex init --minimal`}
        </CodeBlock>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">3. Run the Workflow</h2>
        <CodeBlock title="terminal">
{`cortex run`}
        </CodeBlock>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">4. View Past Sessions</h2>
        <CodeBlock title="terminal">
{`cortex sessions`}
        </CodeBlock>
      </>
    ),
  },

  'installation': {
    title: 'Installation',
    content: (
      <>
        <p className="text-lg text-dark-100 mb-6">
          Cortex can be installed through multiple package managers.
        </p>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">npm (Recommended)</h2>
        <CodeBlock title="terminal">
{`npm install -g @insien/cortex-cli`}
        </CodeBlock>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">Quick Install (Shell)</h2>
        <CodeBlock title="terminal">
{`curl -fsSL https://raw.githubusercontent.com/obliviious/cortex/main/install.sh | bash`}
        </CodeBlock>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">Homebrew (macOS/Linux)</h2>
        <CodeBlock title="terminal">
{`brew tap obliviious/tap
brew install cortex`}
        </CodeBlock>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">Go Install</h2>
        <CodeBlock title="terminal">
{`go install github.com/obliviious/cortex/cmd/agentflow@latest`}
        </CodeBlock>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">From Source</h2>
        <CodeBlock title="terminal">
{`git clone https://github.com/obliviious/cortex.git
cd cortex
make install`}
        </CodeBlock>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">Requirements</h2>
        <ul className="list-disc list-inside text-dark-100 space-y-2">
          <li>One of the supported AI CLI tools installed (Claude Code, OpenCode)</li>
          <li>Go 1.21+ (for building from source)</li>
        </ul>
      </>
    ),
  },

  'cortexfile': {
    title: 'Cortexfile.yml',
    content: (
      <>
        <p className="text-lg text-dark-100 mb-6">
          The Cortexfile.yml is the main configuration file that defines your agents and tasks.
        </p>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">Full Example</h2>
        <CodeBlock title="Cortexfile.yml">
{`# Optional: Working directory for all agents
workdir: /path/to/project

# Agents define the AI tools to use
agents:
  architect:
    tool: claude-code    # or "opencode"
    model: sonnet        # optional: model override

  reviewer:
    tool: claude-code
    model: sonnet

# Tasks define the workflow
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
    needs: [analyze, review]  # Dependencies
    write: true               # Allow file writes
    prompt: |
      Based on the analysis and review:

      ## Analysis:
      {{outputs.analyze}}

      ## Review:
      {{outputs.review}}

      Implement the top priority improvement.

# Local settings (optional)
settings:
  parallel: true
  max_parallel: 4`}
        </CodeBlock>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">Agent Configuration</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-cream-400">
                <th className="text-left py-3 px-4 font-semibold text-dark-400">Field</th>
                <th className="text-left py-3 px-4 font-semibold text-dark-400">Description</th>
              </tr>
            </thead>
            <tbody className="text-dark-100">
              <tr className="border-b border-cream-300">
                <td className="py-3 px-4 font-mono text-coral-500">tool</td>
                <td className="py-3 px-4">AI CLI to use (claude-code, opencode)</td>
              </tr>
              <tr className="border-b border-cream-300">
                <td className="py-3 px-4 font-mono text-coral-500">model</td>
                <td className="py-3 px-4">Optional model override (e.g., sonnet, opus)</td>
              </tr>
            </tbody>
          </table>
        </div>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">Task Configuration</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-cream-400">
                <th className="text-left py-3 px-4 font-semibold text-dark-400">Field</th>
                <th className="text-left py-3 px-4 font-semibold text-dark-400">Description</th>
              </tr>
            </thead>
            <tbody className="text-dark-100">
              <tr className="border-b border-cream-300">
                <td className="py-3 px-4 font-mono text-coral-500">agent</td>
                <td className="py-3 px-4">Reference to agent defined above</td>
              </tr>
              <tr className="border-b border-cream-300">
                <td className="py-3 px-4 font-mono text-coral-500">prompt</td>
                <td className="py-3 px-4">Inline prompt text</td>
              </tr>
              <tr className="border-b border-cream-300">
                <td className="py-3 px-4 font-mono text-coral-500">prompt_file</td>
                <td className="py-3 px-4">Path to external prompt file</td>
              </tr>
              <tr className="border-b border-cream-300">
                <td className="py-3 px-4 font-mono text-coral-500">needs</td>
                <td className="py-3 px-4">Array of task dependencies</td>
              </tr>
              <tr className="border-b border-cream-300">
                <td className="py-3 px-4 font-mono text-coral-500">write</td>
                <td className="py-3 px-4">Allow file writes (default: false)</td>
              </tr>
            </tbody>
          </table>
        </div>
      </>
    ),
  },

  'master-cortex': {
    title: 'MasterCortex.yml',
    content: (
      <>
        <p className="text-lg text-dark-100 mb-6">
          Orchestrate multiple Cortexfiles from a single configuration file.
        </p>

        <CodeBlock title="MasterCortex.yml">
{`# Name and description
name: multi-project-workflow
description: Run workflows across multiple projects

# Execution mode: "sequential" or "parallel"
mode: sequential

# Max concurrent workflows (parallel mode only)
max_parallel: 2

# Stop on first error (default: true for sequential)
stop_on_error: true

# Define workflows to run
workflows:
  # Simple workflow
  - name: main
    path: ./Cortexfile.yml

  # Workflow with custom working directory
  - name: backend
    path: ./backend/Cortexfile.yml
    workdir: ./backend

  # Workflow with dependencies
  - name: frontend
    path: ./frontend/Cortexfile.yml
    needs: [backend]    # Runs after backend completes

  # Glob patterns for multiple projects
  - name: services
    path: "./services/*/Cortexfile.yml"

  # Disabled workflow (skipped)
  - name: experimental
    path: ./experimental/Cortexfile.yml
    enabled: false`}
        </CodeBlock>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">Running MasterCortex</h2>
        <CodeBlock title="terminal">
{`cortex master                 # Auto-detect MasterCortex.yml
cortex master -f custom.yml   # Specify file
cortex master --parallel      # Force parallel mode`}
        </CodeBlock>
      </>
    ),
  },

  'global-config': {
    title: 'Global Config',
    content: (
      <>
        <p className="text-lg text-dark-100 mb-6">
          Set default configuration in <code className="bg-cream-300 px-2 py-1 rounded">~/.cortex/config.yml</code>
        </p>

        <CodeBlock title="~/.cortex/config.yml">
{`# Default agent settings
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
      Authorization: "Bearer token"`}
        </CodeBlock>
      </>
    ),
  },

  'templates': {
    title: 'Template Variables',
    content: (
      <>
        <p className="text-lg text-dark-100 mb-6">
          Pass outputs between tasks using template variables.
        </p>

        <CodeBlock title="Cortexfile.yml">
{`tasks:
  analyze:
    agent: architect
    prompt: Analyze the code...

  implement:
    agent: coder
    needs: [analyze]  # Must declare dependency
    prompt: |
      Based on this analysis:
      {{outputs.analyze}}

      Implement the changes.`}
        </CodeBlock>

        <div className="bg-coral-500/10 border border-coral-500/20 rounded-xl p-4 mt-6">
          <p className="text-dark-400 font-medium">Important</p>
          <p className="text-dark-100 mt-1">
            You must declare the dependency in <code className="bg-cream-300 px-1 rounded">needs</code> to use
            outputs from another task.
          </p>
        </div>
      </>
    ),
  },

  'webhooks': {
    title: 'Webhooks',
    content: (
      <>
        <p className="text-lg text-dark-100 mb-6">
          Configure webhooks to receive notifications about workflow events.
        </p>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">Configuration</h2>
        <CodeBlock title="~/.cortex/config.yml">
{`webhooks:
  - url: https://your-webhook.com/endpoint
    events:
      - run_start
      - run_complete
      - task_start
      - task_complete
      - task_failed
    headers:
      Authorization: "Bearer your-token"`}
        </CodeBlock>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">Webhook Payload</h2>
        <CodeBlock title="JSON payload">
{`{
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
}`}
        </CodeBlock>
      </>
    ),
  },

  'sessions': {
    title: 'Session Storage',
    content: (
      <>
        <p className="text-lg text-dark-100 mb-6">
          Run results are automatically stored for later reference.
        </p>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">Storage Location</h2>
        <CodeBlock title="File structure">
{`~/.cortex/
├── config.yml          # Global config
└── sessions/
    └── my-project/
        └── run-20240104-200000/
            ├── run.json        # Run summary
            ├── analyze.json    # Task results
            └── review.json`}
        </CodeBlock>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">Viewing Sessions</h2>
        <CodeBlock title="terminal">
{`cortex sessions                  # List recent sessions
cortex sessions --limit 20       # Show more sessions
cortex sessions --project myapp  # Filter by project
cortex sessions --failed         # Show only failed sessions`}
        </CodeBlock>
      </>
    ),
  },

  'commands': {
    title: 'Commands',
    content: (
      <>
        <p className="text-lg text-dark-100 mb-6">
          Complete reference for all Cortex CLI commands.
        </p>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">cortex init</h2>
        <p className="text-dark-100 mb-4">Create a template Cortexfile.yml</p>
        <CodeBlock title="terminal">
{`cortex init [flags]

Flags:
      --minimal   Create a minimal template (quick start)
      --master    Create a MasterCortex.yml instead
      --force     Overwrite existing file`}
        </CodeBlock>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">cortex run</h2>
        <p className="text-dark-100 mb-4">Execute the Cortexfile workflow</p>
        <CodeBlock title="terminal">
{`cortex run [flags]

Flags:
  -f, --file stringArray   Path to Cortexfile(s)
  -v, --verbose            Verbose output
  -s, --stream             Stream real-time logs (default: on)
      --no-stream          Disable real-time streaming
      --full               Show full output
  -i, --interactive        Enable Ctrl+O toggle
      --parallel           Enable parallel execution (default: on)
      --sequential         Force sequential execution
      --max-parallel int   Max concurrent tasks (0 = CPU cores)
      --no-color           Disable colored output
      --compact            Minimal output (no banner)`}
        </CodeBlock>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">cortex master</h2>
        <p className="text-dark-100 mb-4">Run multiple workflows from MasterCortex.yml</p>
        <CodeBlock title="terminal">
{`cortex master [flags]

Flags:
  -f, --file string   Path to MasterCortex.yml
      --parallel      Force parallel execution
      --sequential    Force sequential execution
      --no-color      Disable colored output
      --compact       Minimal output`}
        </CodeBlock>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">cortex validate</h2>
        <p className="text-dark-100 mb-4">Validate configuration without running</p>
        <CodeBlock title="terminal">
{`cortex validate`}
        </CodeBlock>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">cortex sessions</h2>
        <p className="text-dark-100 mb-4">List previous run sessions</p>
        <CodeBlock title="terminal">
{`cortex sessions [flags]

Flags:
      --project string   Filter by project name
      --limit int        Max sessions to show (default: 10)
      --failed           Show only failed sessions`}
        </CodeBlock>

        <h2 className="font-display text-2xl text-dark-400 mt-8 mb-4">Supported Tools</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-cream-400">
                <th className="text-left py-3 px-4 font-semibold text-dark-400">Tool</th>
                <th className="text-left py-3 px-4 font-semibold text-dark-400">CLI Command</th>
                <th className="text-left py-3 px-4 font-semibold text-dark-400">Description</th>
              </tr>
            </thead>
            <tbody className="text-dark-100">
              <tr className="border-b border-cream-300">
                <td className="py-3 px-4 font-mono text-coral-500">claude-code</td>
                <td className="py-3 px-4">claude</td>
                <td className="py-3 px-4">Anthropic&apos;s Claude Code CLI</td>
              </tr>
              <tr className="border-b border-cream-300">
                <td className="py-3 px-4 font-mono text-coral-500">opencode</td>
                <td className="py-3 px-4">opencode</td>
                <td className="py-3 px-4">OpenCode CLI</td>
              </tr>
            </tbody>
          </table>
        </div>
      </>
    ),
  },
}

export default function DocsContent() {
  const { section } = useParams()
  const currentSection = section || 'getting-started'
  const doc = docsContent[currentSection] || docsContent['getting-started']

  return (
    <article className="flex-1 min-w-0">
      <h1 className="font-display text-4xl text-dark-400 mb-6">{doc.title}</h1>
      <div className="prose prose-lg max-w-none">{doc.content}</div>
    </article>
  )
}

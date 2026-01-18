# AgentMail

[![Test](https://github.com/UserAd/AgentMail/actions/workflows/test.yml/badge.svg)](https://github.com/UserAd/AgentMail/actions/workflows/test.yml)
[![Release](https://github.com/UserAd/AgentMail/actions/workflows/release.yml/badge.svg)](https://github.com/UserAd/AgentMail/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/UserAd/AgentMail)](https://goreportcard.com/report/github.com/UserAd/AgentMail)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go CLI tool for inter-agent communication within tmux sessions. Agents running in different tmux windows can send and receive messages through a simple file-based mail system.

## Features

- **Asynchronous messaging** - Send messages to agents in other tmux windows without blocking
- **FIFO message queue** - Messages delivered in order, oldest first
- **Simple file-based storage** - Messages stored in `.agentmail/` as JSONL files
- **Concurrent-safe** - File locking ensures atomic operations between agents
- **Minimal dependencies** - Built with Go standard library + lightweight CLI framework
- **Ignore lists** - Filter out windows you don't want to communicate with
- **Stdin support** - Pipe messages from other commands
- **Daemon notifications** - Background mailman daemon monitors mailboxes and notifies agents
- **Agent status tracking** - Agents can set status (ready/work/offline) for smart notifications
- **MCP server** - Model Context Protocol server for Claude Code, Codex CLI, and Gemini CLI
- **Claude Code integration** - Plugin and hooks for AI agent orchestration
- **Cleanup utility** - Remove stale recipients, old messages, and empty mailboxes

## Requirements

- Go 1.25.5 or later
- tmux (must be running inside a tmux session)
- Linux or macOS

## Installation

### Homebrew (macOS/Linux)

The easiest way to install AgentMail is via Homebrew:

```bash
brew install UserAd/agentmail/agentmail
```

Or add the tap first, then install:

```bash
brew tap UserAd/agentmail
brew install agentmail
```

> **Note:** If you have another package named `agentmail` installed, use the full tap path: `brew install UserAd/agentmail/agentmail`

### From Releases

Download the latest binary for your platform from the [Releases](https://github.com/UserAd/AgentMail/releases) page.

Available platforms:

- Linux (amd64, arm64)
- macOS (amd64, arm64)

### From Source

```bash
# Clone the repository
git clone https://github.com/UserAd/AgentMail.git
cd AgentMail

# Build
go build -o agentmail ./cmd/agentmail

# Or install to $GOPATH/bin
go install ./cmd/agentmail
```

## Quick Start

```bash
# Start a tmux session with multiple windows
tmux new-session -s agents -n agent-1
tmux new-window -t agents -n agent-2

# In agent-1 window: send a message
agentmail send agent-2 "Hello from agent-1!"
# Output: Message #xK7mN2pQ sent

# In agent-2 window: receive the message
agentmail receive
# Output:
# From: agent-1
# ID: xK7mN2pQ
#
# Hello from agent-1!
```

## Commands

### send

Send a message to another agent (tmux window).

```bash
agentmail send [flags] [<recipient>] [<message>]
```

**Arguments (positional or flags):**

- `<recipient>` - Target tmux window name (required)
- `<message>` - Message content (optional if using stdin)

**Flags:**

- `-r, --recipient <name>` - Recipient tmux window name
- `-m, --message <text>` - Message content

Flags take precedence over positional arguments.

**Examples:**

```bash
# Send with positional arguments
agentmail send agent-2 "Task completed successfully"

# Send with flags (equivalent)
agentmail send -r agent-2 -m "Task completed successfully"
agentmail send --recipient agent-2 --message "Task completed"

# Send via stdin
echo "Results from analysis" | agentmail send agent-2
echo "Results" | agentmail send -r agent-2

# Send multi-line content
cat report.txt | agentmail send agent-2
```

**Exit codes:**

- `0` - Message sent successfully
- `1` - Error (invalid recipient, missing message, etc.)
- `2` - Not running inside tmux

### receive

Read the oldest unread message from your mailbox.

```bash
agentmail receive [--hook]
```

**Flags:**

- `--hook` - Enable hook mode for Claude Code integration (see [Claude Code Hooks](#claude-code-hooks-manual-setup))

**Output format (normal mode):**

```text
From: <sender>
ID: <message-id>

<message content>
```

Returns "No unread messages" if the mailbox is empty.

**Exit codes (normal mode):**

- `0` - Success (message displayed or no messages)
- `1` - Error reading mailbox
- `2` - Not running inside tmux

**Exit codes (hook mode):**

- `0` - No messages, not in tmux, or error (silent)
- `2` - New message available (output to STDERR)

### recipients

List all available recipients (tmux windows in the current session).

```bash
agentmail recipients
```

**Example output:**

```text
agent-1 [you]
agent-2
agent-3
```

The current window is marked with `[you]`.

**Exit codes:**

- `0` - Success
- `1` - Error listing windows
- `2` - Not running inside tmux

### status

Set your agent's availability status for daemon notifications.

```bash
agentmail status <ready|work|offline>
```

**Statuses:**

- `ready` - Available to receive notifications (daemon will alert you of new mail)
- `work` - Busy working (notifications suppressed until you return to ready)
- `offline` - Not available (notifications suppressed)

**Examples:**

```bash
# Mark yourself as ready to receive notifications
agentmail status ready

# Mark yourself as busy (no notifications until ready again)
agentmail status work

# Go offline
agentmail status offline
```

**Exit codes:**

- `0` - Status updated successfully
- `1` - Invalid status value

### mailman

Start the mailman daemon to monitor mailboxes and notify agents.

```bash
agentmail mailman [--daemon]
```

**Flags:**

- `--daemon` - Run in background (daemonize)

**Behavior:**

- Uses file watching (fsnotify) for instant notification on mailbox changes
- Includes 60-second safety timer that runs alongside watching
- Sends notifications to agents with `ready` status that have unread mail
- Notifications sent via tmux: `tmux send-keys -t <window> "Check your agentmail"`
- Stores PID in `.agentmail/mailman.pid`
- Gracefully shuts down on SIGTERM/SIGINT

**Examples:**

```bash
# Run in foreground (useful for debugging)
agentmail mailman

# Run as background daemon
agentmail mailman --daemon
```

**Exit codes:**

- `0` - Daemon started/stopped successfully
- `1` - Error (failed to start, PID file error, etc.)
- `2` - Daemon already running

### onboard

Output AI-optimized onboarding context about AgentMail.

```bash
agentmail onboard
```

This command outputs a quick reference for AI agents to understand AgentMail's capabilities. Used by Claude Code SessionStart hooks for agent initialization.

**Exit codes:**

- `0` - Success

### cleanup

Remove stale data from the AgentMail system.

```bash
agentmail cleanup [flags]
```

**What gets cleaned:**

- **Offline recipients** - Entries in recipients.jsonl for tmux windows that no longer exist
- **Stale recipients** - Recipients not updated within the threshold (default: 48 hours)
- **Old delivered messages** - Read messages older than the threshold (default: 2 hours)
- **Empty mailboxes** - Mailbox files with zero messages

**Flags:**

- `--stale-hours <N>` - Hours threshold for stale recipients (default: 48)
- `--delivered-hours <N>` - Hours threshold for delivered messages (default: 2)
- `--dry-run` - Report what would be cleaned without deleting

**Examples:**

```bash
# Preview what would be cleaned
agentmail cleanup --dry-run

# Clean with default thresholds (48h stale, 2h delivered)
agentmail cleanup

# Clean with custom thresholds
agentmail cleanup --stale-hours 24 --delivered-hours 1
```

**Output:**

```text
Cleanup complete:
  Recipients removed: 3 (2 offline, 1 stale)
  Messages removed: 15
  Mailboxes removed: 2
```

**Exit codes:**

- `0` - Success
- `1` - Error during cleanup

**Note:** This is an administrative command not intended for AI agent use. It is not exposed via MCP tools or onboarding.

### help

Display usage information.

```bash
agentmail --help
agentmail -h
```

## Configuration

### Ignore List

Create a `.agentmailignore` file in your git repository root to exclude certain windows from the recipients list and prevent sending to them.

```bash
# .agentmailignore
test-runner
debug-window
monitoring
```

- One window name per line
- Whitespace is trimmed
- Your current window is always shown even if listed
- Missing file means no exclusions

## MCP Server

AgentMail includes a built-in MCP (Model Context Protocol) server that enables AI agents to communicate via a standardized interface. The MCP server exposes four tools:

| Tool | Description |
| ---- | ----------- |
| `send` | Send a message to another agent (max 64KB) |
| `receive` | Receive the oldest unread message (FIFO) |
| `status` | Set agent availability (ready/work/offline) |
| `list-recipients` | List available agents in the session |

### Running the MCP Server

```bash
agentmail mcp
```

The server uses STDIO transport and must be run inside a tmux session.

### Claude Code Configuration

Add to `~/.claude/settings.json` or your project's `.mcp.json`:

```json
{
  "mcpServers": {
    "agentmail": {
      "command": "agentmail",
      "args": ["mcp"]
    }
  }
}
```

See [Claude Code MCP documentation](https://code.claude.com/docs/en/mcp) for more details.

### OpenAI Codex CLI Configuration

Add to `~/.codex/config.toml`:

```toml
[mcp_servers.agentmail]
command = "agentmail"
args = ["mcp"]
env_vars = ["TMUX", "TMUX_PANE"]
```

The `env_vars` setting is required to pass TMUX environment variables to the MCP server, enabling tmux session and window detection.

See [Codex MCP documentation](https://developers.openai.com/codex/mcp/) for more details.

### Google Gemini CLI Configuration

Add to `~/.gemini/settings.json`:

```json
{
  "mcpServers": {
    "agentmail": {
      "command": "agentmail",
      "args": ["mcp"]
    }
  }
}
```

See [Gemini CLI MCP documentation](https://geminicli.com/docs/tools/mcp-server/) for more details.

### MCP Tool Responses

**send** returns:

```json
{"message_id": "xK7mN2pQ"}
```

**receive** returns (message available):

```json
{"from": "agent-1", "id": "xK7mN2pQ", "message": "Hello!"}
```

**receive** returns (no messages):

```json
{"status": "No unread messages"}
```

**status** returns:

```json
{"status": "ok"}
```

**list-recipients** returns:

```json
{
  "recipients": [
    {"name": "agent-1", "is_current": true},
    {"name": "agent-2", "is_current": false}
  ]
}
```

## Claude Code Plugin

AgentMail provides a Claude Code plugin for seamless integration with AI agents. The plugin automatically manages agent status and checks for messages.

### Plugin Installation

Install the AgentMail plugin directly in Claude Code:

```bash
# Add the AgentMail marketplace
/plugin marketplace add UserAd/AgentMail

# Install the plugin
/plugin install agentmail@agentmail-marketplace
```

Or install from source:

```bash
# Clone and install locally
git clone https://github.com/UserAd/AgentMail.git
/plugin install ./AgentMail/claude-plugin
```

### What the Plugin Does

The plugin configures hooks that automatically:

| Event | Action | Status |
| ----- | ------ | ------ |
| **SessionStart** | Sets status to ready, runs onboarding | `ready` |
| **UserPromptSubmit** | Sets status to work (agent is busy) | `work` |
| **Stop** (end of turn) | Sets status to ready, checks for messages | `ready` |
| **SessionEnd** | Sets status to offline | `offline` |

### Plugin Commands

After installation, these commands are available:

- `/send` - Send a message to another agent
- `/receive` - Check and read messages
- `/recipients` - List available agents
- `/status` - Set your availability status

### Plugin Skills

The plugin includes the `agentmail` skill with complete documentation for sending and receiving messages between AI agents.

## Claude Code Hooks (Manual Setup)

If you prefer manual configuration instead of the plugin, AgentMail integrates with Claude Code hooks to manage agent status and check for messages automatically.

### Setup

Add to your Claude Code settings (`.claude/settings.json` in your project or `~/.claude/settings.json` globally):

```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "agentmail status ready && agentmail onboard"
          }
        ]
      }
    ],
    "SessionEnd": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "agentmail status offline"
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "agentmail status ready && agentmail receive --hook"
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "agentmail status work"
          }
        ]
      }
    ]
  }
}
```

### How It Works

| Hook Event | Command | Purpose |
| ---------- | ------- | ------- |
| **SessionStart** | `agentmail status ready && agentmail onboard` | Sets agent ready, outputs onboarding context |
| **SessionEnd** | `agentmail status offline` | Marks agent as offline when session ends |
| **Stop** | `agentmail status ready && agentmail receive --hook` | Sets ready after each turn, checks for mail |
| **UserPromptSubmit** | `agentmail status work` | Marks agent as busy when processing |

### Hook Mode Behavior (`--hook` flag)

| Condition | Output | Exit Code |
| --------- | ------ | --------- |
| Unread messages exist | Message to STDERR with "You got new mail" | 2 |
| No unread messages | None | 0 |
| Not in tmux session | None | 0 |
| Any error | None | 0 |

Hook mode is designed to be non-disruptive: errors exit silently rather than interrupting your workflow.

### Example Output

When you have mail (on Stop hook), Claude Code will display:

```text
You got new mail
From: agent-1
ID: xK7mN2pQ

Task completed! Results are in /tmp/output.json
```

## Architecture

```text
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   agent-1   │     │   agent-2   │     │   agent-3   │
│ (tmux win)  │     │ (tmux win)  │     │ (tmux win)  │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │
                    ┌───────────────────┐
                    │  .agentmail/      │
                    │  └── mailboxes/   │
                    │     ├─agent-1.jsonl│
                    │     ├─agent-2.jsonl│
                    │     └─agent-3.jsonl│
                    └───────────────────┘
```

### Message Storage

Messages are stored in `.agentmail/mailboxes/<recipient>.jsonl` with one JSON object per line:

```json
{"id":"xK7mN2pQ","from":"agent-1","to":"agent-2","message":"Hello!","read_flag":false}
```

Each recipient has their own mailbox file, minimizing lock contention.

### Message IDs

Each message gets a unique 8-character base62 ID (a-z, A-Z, 0-9) generated using cryptographically secure random bytes.

### Concurrency

AgentMail uses POSIX file locking (`flock`) to ensure atomic read-modify-write operations. Multiple agents can safely send and receive messages concurrently.

### Daemon System

The mailman daemon provides proactive notifications for agents:

```text
┌─────────────────────────────────────────────────────────────┐
│                     Mailman Daemon                          │
│                                                             │
│  1. Watch .agentmail/ for file changes (fsnotify)          │
│  2. On change (debounced 500ms) or 60s fallback timer:     │
│     - Read recipient states from recipients.jsonl          │
│     - For each "ready" agent with unread messages:         │
│       - If not already notified: send notification         │
│  3. Also notifies stateless agents (no recipient state)    │
│     every 60 seconds if they have unread messages          │
└─────────────────────────────────────────────────────────────┘
```

**Recipient State File** (`.agentmail/recipients.jsonl`):

```json
{"recipient":"agent-1","status":"ready","updated_at":"2024-01-12T10:00:00Z","notified":false}
{"recipient":"agent-2","status":"work","updated_at":"2024-01-12T10:05:00Z","notified":false}
```

**Notification Flow:**

1. Agent sets status to `ready` using `agentmail status ready`
2. Mailman daemon detects unread messages for agent
3. Daemon sends notification via tmux: `Check your agentmail`
4. Agent's `notified` flag is set to prevent duplicate notifications
5. When agent changes to `work` or `offline`, `notified` resets

## Development

### Build

```bash
go build ./...
```

### Test

```bash
# Run all tests
go test ./...

# With verbose output
go test -v ./...

# With race detection and coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out
```

### Lint & Security

```bash
# Format code
go fmt ./...

# Static analysis
go vet ./...

# Vulnerability scanning
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# Security scanning
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec ./...
```

### Testing in CI Environment

To match the CI environment (Go 1.25.5, Linux):

```bash
docker run --rm -v $(pwd):/app -w /app golang:1.25.5 go test -v -race ./...
```

### CI/CD Pipeline

The project uses GitHub Actions for continuous integration and delivery:

**Test Workflow** (on push to main and PRs):

- Code formatting verification (`gofmt`)
- Dependency verification (`go mod verify`)
- Static analysis (`go vet`)
- Tests with race detection (`go test -race`)
- Coverage report generation
- Security scanning (`govulncheck`, `gosec`)

**Release Workflow** (on push to main):

- Pre-release test validation
- Semantic version calculation from commit messages
- Cross-platform builds (Linux/macOS, amd64/arm64)
- GitHub Release creation with auto-generated notes
- Homebrew formula automatic update

## Project Structure

```text
AgentMail/
├── cmd/
│   └── agentmail/          # CLI entry point
├── internal/
│   ├── cli/                # Command implementations
│   ├── daemon/             # Mailman daemon and notification loop
│   ├── mail/               # Message and mailbox logic
│   ├── mcp/                # MCP server implementation
│   └── tmux/               # tmux integration
├── claude-plugin/          # Claude Code plugin
├── .github/workflows/      # CI/CD automation
├── specs/                  # Feature specifications
├── go.mod                  # Go module definition
└── LICENSE                 # MIT License
```

## Security

AgentMail includes several security measures:

- **Path traversal protection** - All file paths are validated to prevent directory traversal attacks
- **Command injection prevention** - tmux pane IDs are validated with regex patterns
- **Atomic file operations** - POSIX file locking prevents race conditions
- **Input validation** - Recipients and status values are validated against whitelists
- **Self-send prevention** - Agents cannot send messages to themselves
- **Local-only operation** - No network communication; purely file-based

## Use Cases

- **Multi-agent AI systems** - Coordinate multiple AI agents running in separate tmux windows
- **Distributed task processing** - Send work items between worker processes
- **Event notifications** - Alert agents about state changes or completed tasks
- **Simple IPC** - Lightweight inter-process communication without complex setup

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please ensure:

- Code passes `go fmt` and `go vet`
- Tests pass with `go test -race ./...`
- New features include appropriate tests

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

# AgentMail

[![Test](https://github.com/UserAd/AgentMail/actions/workflows/test.yml/badge.svg)](https://github.com/UserAd/AgentMail/actions/workflows/test.yml)
[![Release](https://github.com/UserAd/AgentMail/actions/workflows/release.yml/badge.svg)](https://github.com/UserAd/AgentMail/actions/workflows/release.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go CLI tool for inter-agent communication within tmux sessions. Agents running in different tmux windows can send and receive messages through a simple file-based mail system.

## Features

- **Asynchronous messaging** - Send messages to agents in other tmux windows without blocking
- **FIFO message queue** - Messages delivered in order, oldest first
- **Simple file-based storage** - Messages stored in `.git/mail/` as JSONL files
- **Concurrent-safe** - File locking ensures atomic operations between agents
- **Minimal dependencies** - Built with Go standard library + lightweight CLI framework
- **Ignore lists** - Filter out windows you don't want to communicate with
- **Stdin support** - Pipe messages from other commands
- **Claude Code hooks** - Integration with Claude Code for mail notifications

## Requirements

- Go 1.21 or later
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
- Linux (amd64)
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
- `--hook` - Enable hook mode for Claude Code integration (see [Claude Code Hooks](#claude-code-hooks))

**Output format (normal mode):**
```
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
```
agent-1 [you]
agent-2
agent-3
```

The current window is marked with `[you]`.

**Exit codes:**
- `0` - Success
- `1` - Error listing windows
- `2` - Not running inside tmux

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
|-------|--------|--------|
| **SessionStart** | Sets status to ready, runs onboarding | `ready` |
| **SessionEnd** | Sets status to offline | `offline` |
| **Stop** (end of turn) | Sets status to ready, checks for messages | `ready` |

### Plugin Commands

After installation, these commands are available:

- `/send` - Send a message to another agent
- `/receive` - Check and read messages
- `/recipients` - List available agents
- `/status` - Set your availability status

### Plugin Skills

The plugin includes the `agentmail` skill with complete documentation for sending and receiving messages between AI agents.

## Claude Code Hooks (Manual Setup)

If you prefer manual configuration instead of the plugin, AgentMail integrates with Claude Code hooks to notify you when other agents send messages. Configure it as a `user-prompt-submit` hook to check for mail before each prompt.

### Setup

Add to your Claude Code settings (`.claude/settings.json` in your project or `~/.claude/settings.json` globally):

```json
{
  "hooks": {
    "user-prompt-submit": {
      "command": "agentmail receive --hook"
    }
  }
}
```

### How It Works

1. Before each prompt submission, Claude Code runs `agentmail receive --hook`
2. If you have unread mail:
   - The message appears in Claude Code's output (via STDERR)
   - Exit code 2 signals a notification
3. If no mail or not in tmux:
   - Silent exit (no output, exit code 0)
   - Your workflow continues uninterrupted

### Hook Mode Behavior

| Condition | Output | Exit Code |
|-----------|--------|-----------|
| Unread messages exist | Message to STDERR with "You got new mail" | 2 |
| No unread messages | None | 0 |
| Not in tmux session | None | 0 |
| Any error | None | 0 |

Hook mode is designed to be non-disruptive: errors exit silently rather than interrupting your workflow.

### Example Output

When you have mail, Claude Code will display:

```
You got new mail
From: agent-1
ID: xK7mN2pQ

Task completed! Results are in /tmp/output.json
```

## How It Works

### Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   agent-1   │     │   agent-2   │     │   agent-3   │
│ (tmux win)  │     │ (tmux win)  │     │ (tmux win)  │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │
                    ┌──────┴──────┐
                    │  .git/mail/ │
                    ├─────────────┤
                    │ agent-1.jsonl│
                    │ agent-2.jsonl│
                    │ agent-3.jsonl│
                    └─────────────┘
```

### Message Storage

Messages are stored in `.git/mail/<recipient>.jsonl` with one JSON object per line:

```json
{"id":"xK7mN2pQ","from":"agent-1","to":"agent-2","message":"Hello!","read_flag":false}
```

Each recipient has their own mailbox file, minimizing lock contention.

### Message IDs

Each message gets a unique 8-character base62 ID (a-z, A-Z, 0-9) generated using cryptographically secure random bytes.

### Concurrency

AgentMail uses POSIX file locking (`flock`) to ensure atomic read-modify-write operations. Multiple agents can safely send and receive messages concurrently.

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

### Lint

```bash
go fmt ./...
go vet ./...
```

### Testing in CI Environment

To match the CI environment (Go 1.21, Linux):

```bash
docker run --rm -v $(pwd):/app -w /app golang:1.21 go test -v -race ./...
```

## Project Structure

```
AgentMail/
├── cmd/
│   └── agentmail/          # CLI entry point
├── internal/
│   ├── cli/                # Command implementations
│   ├── mail/               # Message and mailbox logic
│   └── tmux/               # tmux integration
├── .github/workflows/      # CI/CD automation
├── specs/                  # Feature specifications
├── go.mod                  # Go module definition
└── LICENSE                 # MIT License
```

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

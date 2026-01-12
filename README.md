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
- **Zero dependencies** - Built with Go standard library only
- **Ignore lists** - Filter out windows you don't want to communicate with
- **Stdin support** - Pipe messages from other commands

## Requirements

- Go 1.21 or later
- tmux (must be running inside a tmux session)
- Linux or macOS

## Installation

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

### From Releases

Download the latest binary for your platform from the [Releases](https://github.com/UserAd/AgentMail/releases) page.

Available platforms:
- Linux (amd64)
- macOS (amd64, arm64)

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
agentmail send <recipient> [message]
```

**Arguments:**
- `<recipient>` - Target tmux window name (required)
- `[message]` - Message content (optional if using stdin)

**Examples:**
```bash
# Send with inline message
agentmail send agent-2 "Task completed successfully"

# Send via stdin
echo "Results from analysis" | agentmail send agent-2

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
agentmail receive
```

**Output format:**
```
From: <sender>
ID: <message-id>

<message content>
```

Returns "No unread messages" if the mailbox is empty.

**Exit codes:**
- `0` - Success (message displayed or no messages)
- `1` - Error reading mailbox
- `2` - Not running inside tmux

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

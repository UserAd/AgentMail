# Quickstart: AgentMail

**Feature**: 001-agent-mail-structure
**Date**: 2026-01-11

## Prerequisites

- Go 1.21 or later
- tmux installed and running
- macOS or Linux

## Build

```bash
# Clone and build
cd AgentMail
go build -o agentmail ./cmd/agentmail

# Optional: install to PATH
go install ./cmd/agentmail
```

## Usage

AgentMail must run inside a tmux session. Each tmux window represents an agent.

### Setup

```bash
# Start tmux with named windows for agents
tmux new-session -s mail -n agent-1
# In another terminal:
tmux new-window -t mail -n agent-2
```

### Send a Message

From `agent-1` window:
```bash
agentmail send agent-2 "Hello from agent-1!"
# Output: xK7mN2pQ (message ID)
```

### Receive a Message

From `agent-2` window:
```bash
agentmail receive
# Output:
# From: agent-1
# ID: xK7mN2pQ
#
# Hello from agent-1!
```

### Check for Messages

```bash
agentmail receive
# Output: No unread messages
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (invalid args, recipient not found, I/O error) |
| 2 | Not running inside tmux |

## Storage

Messages are stored in `.git/mail/<recipient>.jsonl` - one file per recipient. For example:
- `.git/mail/agent-1.jsonl` - Messages for agent-1
- `.git/mail/agent-2.jsonl` - Messages for agent-2

The directory is created automatically on first use.

## Development

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Format code
go fmt ./...

# Run linter
go vet ./...
```

## Project Structure

```
cmd/agentmail/main.go    # CLI entry point
internal/
  mail/                  # Message and mailbox logic
  tmux/                  # tmux integration
  cli/                   # Command implementations
```

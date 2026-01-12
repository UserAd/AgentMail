# Quickstart: Recipients Command, Help Flag, and Stdin Message Input

**Feature**: 003-recipients-help-stdin
**Date**: 2026-01-12

## Prerequisites

- Go 1.21 or later installed
- tmux installed and running
- Inside a tmux session with multiple windows

## Build

```bash
cd /path/to/AgentMail
go build -o agentmail ./cmd/agentmail
```

## Feature Usage

### 1. List Available Recipients

```bash
# Show all windows (current window marked with [you])
agentmail recipients
```

Output (one window per line, current window marked):
```
main [you]
agent1
agent2
worker
```

### 2. Filter Recipients with Ignore File

Create `.agentmailignore` in your git repository root:

```bash
# Create ignore file
cat > .agentmailignore <<EOF
monitor
logs
EOF
```

Now `agentmail recipients` will exclude `monitor` and `logs`.

### 3. Send Message via Stdin

```bash
# Pipe a message
echo "Hello from stdin" | agentmail send agent2

# Multi-line message
cat <<EOF | agentmail send agent2
This is line 1
This is line 2
EOF

# Pipe command output
date | agentmail send agent2
```

### 4. Send Message via Argument (Existing)

```bash
# Same as before - still works
agentmail send agent2 "Hello World"
```

### 5. Get Help

```bash
agentmail --help
```

Output:
```
agentmail - Inter-agent communication for tmux sessions

USAGE:
    agentmail <command> [arguments]
    agentmail --help

COMMANDS:
    send <recipient> [message]    Send a message to a tmux window
                                  Message can be piped via stdin
    receive                       Read the oldest unread message
    recipients                    List available message recipients

EXAMPLES:
    agentmail send agent2 "Hello"
    echo "Hello" | agentmail send agent2
    agentmail receive
    agentmail recipients
```

## Testing the Feature

### Manual Test Sequence

1. **Setup**: Open tmux with 3 windows: `main`, `agent1`, `agent2`

2. **Test recipients command**:
   ```bash
   # From main window
   agentmail recipients
   # Should show: main [you], agent1, agent2
   ```

3. **Test ignore file**:
   ```bash
   echo "agent2" > .agentmailignore
   agentmail recipients
   # Should show: main [you], agent1
   ```

4. **Test blocked send**:
   ```bash
   agentmail send agent2 "test"
   # Should error: recipient not found
   ```

5. **Test stdin send**:
   ```bash
   echo "Hello via stdin" | agentmail send agent1
   # Should succeed

   # In agent1 window:
   agentmail receive
   # Should show: Hello via stdin
   ```

6. **Test help**:
   ```bash
   agentmail --help
   # Should show help text
   ```

## Common Issues

| Issue | Solution |
|-------|----------|
| "not running inside a tmux session" | Run from within tmux |
| "recipient not found" | Check window exists with `tmux list-windows` |
| Recipient excluded unexpectedly | Check `.agentmailignore` in git root |
| No stdin detected | Ensure you're piping, not redirecting from terminal |

## Development

### Run Tests

```bash
go test -v ./...
```

### Check Coverage

```bash
go test -cover ./...
# Must be >= 80% per constitution
```

### Format Code

```bash
go fmt ./...
go vet ./...
```

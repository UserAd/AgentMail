# CLI Interface Contract: Recipients Command, Help Flag, and Stdin Message Input

**Feature**: 003-recipients-help-stdin
**Date**: 2026-01-12

## Command: `agentmail recipients`

### Synopsis
```
agentmail recipients
```

### Description
Lists all tmux windows in the current session. The current window is marked with "[you]" suffix. Windows listed in `.agentmailignore` are excluded.

### Input
- None (no arguments)

### Output

**Success (exit code 0)**:
```
stdout: <window_name>\n
        <window_name> [you]\n
        <window_name>\n
        ...
stderr: (empty)
```

Each window name on its own line. Current window has " [you]" suffix.

**Error (exit code 2)**:
```
stdout: (empty)
stderr: error: not running inside a tmux session\n
```

### Examples
```bash
# List all recipients (running from "main" window)
$ agentmail recipients
main [you]
agent1
agent2
worker

# Only one window
$ agentmail recipients
main [you]
```

---

## Command: `agentmail --help`

### Synopsis
```
agentmail --help
agentmail -h
```

### Description
Displays usage information for all agentmail commands.

### Input
- `--help` or `-h` flag

### Output

**Success (exit code 0)**:
```
stdout:
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

stderr: (empty)
```

---

## Command: `agentmail send` (Modified)

### Synopsis
```
agentmail send <recipient> [message]
echo "message" | agentmail send <recipient>
```

### Description
Sends a message to a recipient. Message can be provided as an argument or piped via stdin. Stdin takes precedence when both are provided.

### Input
- `<recipient>`: Required. Window name to send message to.
- `[message]`: Optional if stdin is provided. Message content.
- `stdin`: Optional. Message content (takes precedence over argument).

### Output

**Success (exit code 0)**:
```
stdout: Message #<ID> sent\n
stderr: (empty)
```

**Error - Recipient not found (exit code 1)**:
```
stdout: (empty)
stderr: error: recipient not found\n
```

Applies when:
- Window does not exist in tmux session
- Window is listed in `.agentmailignore`
- Window is the current window (self-send not allowed)

**Error - No message (exit code 1)**:
```
stdout: (empty)
stderr: error: no message provided\n
        usage: agentmail send <recipient> <message>\n
```

**Error - Not in tmux (exit code 2)**:
```
stdout: (empty)
stderr: error: not running inside a tmux session\n
```

### Examples
```bash
# Send via argument
$ agentmail send agent2 "Hello World"
Message #ABC123 sent

# Send via stdin
$ echo "Hello World" | agentmail send agent2
Message #ABC123 sent

# Multi-line message via stdin
$ cat <<EOF | agentmail send agent2
Line 1
Line 2
EOF
Message #ABC123 sent

# Stdin takes precedence
$ echo "stdin message" | agentmail send agent2 "argument message"
Message #ABC123 sent
# agent2 receives "stdin message"
```

---

## Exit Codes

| Code | Meaning | When |
|------|---------|------|
| 0 | Success | Command completed successfully |
| 1 | Error | Invalid input, recipient not found, no message |
| 2 | Environment Error | Not in tmux session |

---

## File: `.agentmailignore`

### Location
Git repository root (same directory as `.git/`)

### Format
```
<window_name>
<window_name>
...
```

### Behavior
- Lines are trimmed of leading/trailing whitespace
- Empty lines are ignored
- Whitespace-only lines are ignored
- Matching is case-sensitive and exact
- If file is unreadable, treated as if it doesn't exist

### Example
```
# .agentmailignore
monitor
logs
debug
```

Windows `monitor`, `logs`, and `debug` will:
- Not appear in `agentmail recipients` output
- Cause "recipient not found" error when used with `agentmail send`

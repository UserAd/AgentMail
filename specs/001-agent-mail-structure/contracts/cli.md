# CLI Contract: AgentMail

**Feature**: 001-agent-mail-structure
**Date**: 2026-01-11

## Overview

AgentMail provides two CLI commands for inter-agent messaging within tmux sessions.

## Commands

### agentmail send

Send a message to another agent.

**Synopsis**:
```bash
agentmail send <recipient> <message>
```

**Arguments**:
| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `recipient` | string | Yes | Target agent's tmux window name |
| `message` | string | Yes | Message body text |

**Behavior**:
1. Validate running inside tmux (FR-005a, FR-005b)
2. Validate required arguments present (FR-001e)
3. Get sender identity from current tmux window (FR-004)
4. Validate recipient exists in tmux session (FR-001a, FR-001d)
5. Ensure `.git/mail/` directory exists (FR-011)
6. Generate unique message ID (FR-010a)
7. Store message in mailbox (FR-001b, FR-008)
8. Print message ID to stdout (FR-001c, FR-010b)

**Output**:
- **stdout**: Message ID on success (e.g., `xK7mN2pQ`)
- **stderr**: Error messages

**Exit Codes**:
| Code | Condition |
|------|-----------|
| 0 | Message sent successfully |
| 1 | Missing arguments (FR-001e), recipient not found (FR-001d), I/O error |
| 2 | Not running inside tmux (FR-005b) |

**Examples**:
```bash
# Success
$ agentmail send agent-2 "Hello from agent-1"
xK7mN2pQ

# Recipient not found
$ agentmail send nonexistent "Hello"
error: recipient 'nonexistent' not found in tmux session
$ echo $?
1

# Not in tmux
$ agentmail send agent-2 "Hello"
error: agentmail must run inside a tmux session
$ echo $?
2
```

---

### agentmail receive

Receive the oldest unread message for the current agent.

**Synopsis**:
```bash
agentmail receive
```

**Arguments**: None

**Behavior**:
1. Validate running inside tmux (FR-005a, FR-005b)
2. Get receiver identity from current tmux window (FR-004)
3. Validate current window exists in tmux session (FR-006a, FR-006b)
4. Find oldest unread message for receiver in FIFO order (FR-002a, FR-012)
5. If found: display message and mark as read (FR-002a, FR-002b)
6. If not found: print "No unread messages" (FR-003a, FR-003b)

**Output**:
- **stdout** (message found):
  ```
  From: <sender>
  ID: <message-id>

  <message-body>
  ```
- **stdout** (no messages): `No unread messages`
- **stderr**: Error messages

**Exit Codes**:
| Code | Condition |
|------|-----------|
| 0 | Message received OR no unread messages |
| 1 | Current window not found in tmux session (FR-006b), I/O error |
| 2 | Not running inside tmux (FR-005b) |

**Examples**:
```bash
# Message found
$ agentmail receive
From: agent-1
ID: xK7mN2pQ

Hello from agent-1

# No messages
$ agentmail receive
No unread messages

# Not in tmux
$ agentmail receive
error: agentmail must run inside a tmux session
$ echo $?
2

# Current window not in tmux session (FR-006b)
$ agentmail receive
error: current window 'orphan-window' not found in tmux session
$ echo $?
1
```

---

## Common Behaviors

### tmux Detection

Both commands detect tmux context by:
1. Checking `$TMUX` environment variable exists
2. Executing `tmux display-message -p '#W'` to get window name

If either fails, print error to stderr and exit with code 2.

### Recipient Validation

Before sending, validate recipient exists:
```bash
tmux list-windows -F '#{window_name}'
```

Parse output and check if recipient is in the list.

### Message Display Format

When displaying a received message:
```
From: <from-field>
ID: <id-field>

<message-field>
```

- Header fields followed by blank line
- Message body follows
- No trailing newline after message

### Error Message Format

All errors written to stderr with format:
```
error: <description>
```

Examples:
- `error: agentmail must run inside a tmux session`
- `error: recipient 'foo' not found in tmux session`
- `error: missing required arguments: recipient message`
- `error: failed to write message: permission denied`

---

## Testing Contract

### Unit Test Requirements

| Command | Test Case | Expected | Requirement |
|---------|-----------|----------|-------------|
| send | Valid args, tmux present | Exit 0, ID printed | FR-001b, FR-001c |
| send | Missing recipient arg | Exit 1, error message | FR-001e |
| send | Missing message arg | Exit 1, error message | FR-001e |
| send | Invalid recipient | Exit 1, error message | FR-001d |
| send | Not in tmux | Exit 2, error message | FR-005b |
| receive | Unread message exists | Exit 0, message displayed | FR-002a, FR-002b |
| receive | No unread messages | Exit 0, "No unread messages" | FR-003a, FR-003b |
| receive | Not in tmux | Exit 2, error message | FR-005b |
| receive | Current window not in session | Exit 1, error message | FR-006b |
| receive | Multiple messages | Oldest returned (FIFO) | FR-012 |

### Integration Test Requirements

| Scenario | Steps | Expected |
|----------|-------|----------|
| Round-trip | send → receive | Message delivered correctly |
| FIFO order | send 3 msgs → receive 3 times | Oldest first |
| Mark as read | send → receive → receive | Second receive shows next msg |
| Multi-agent | agent-1 sends to agent-2 | agent-2 receives, agent-1 doesn't |
| File isolation | send to agent-1 and agent-2 | Creates separate .git/mail/agent-1.jsonl and .git/mail/agent-2.jsonl |

# CLI Contracts: Mailman Daemon

**Feature**: 006-mailman-daemon
**Date**: 2026-01-12

## Commands

### agentmail mailman

Start the mailman daemon for automated agent notifications.

**Synopsis**:
```
agentmail mailman [--daemon]
```

**Options**:

| Flag | Description |
|------|-------------|
| `--daemon` | Run in background (daemonize), detach from terminal |

**Exit Codes**:

| Code | Meaning |
|------|---------|
| 0 | Daemon started successfully (or exited cleanly on signal) |
| 2 | Daemon already running (environment error) |

**Stdout** (foreground mode):
```
Mailman daemon started (PID: 12345)
Checking for unread messages...
Notified agent1: Check your agentmail
Checking for unread messages...
```

**Stdout** (background mode):
```
Mailman daemon started in background (PID: 12345)
```

**Stderr** (when daemon already running):
```
error: mailman daemon already running (PID: 12345)
```

**Stderr** (stale PID cleanup):
```
Warning: Stale PID file found, cleaning up
Mailman daemon started (PID: 12346)
```

**Behavior**:
1. Check for existing PID file at `.git/mail/mailman.pid`
2. If PID file exists:
   - Parse PID and check if process is running (signal 0)
   - If running: print error to stderr, exit 2
   - If not running: print warning, delete stale file, continue
3. Write current PID to `.git/mail/mailman.pid`
4. Clear stale recipient states (>1 hour old)
5. If `--daemon`: fork, detach from terminal, parent exits 0
6. Enter notification loop (10-second interval):
   - List all mailbox files in `.git/mail/`
   - For each mailbox, check unread messages
   - For ready+unnotified recipients with unread messages: send notification
7. On SIGTERM/SIGINT: delete PID file, exit 0

---

### agentmail status

Set agent availability status for mailman notifications. Designed for hooks integration.

**Synopsis**:
```
agentmail status <STATUS>
```

**Arguments**:

| Argument | Description |
|----------|-------------|
| `STATUS` | One of: `ready`, `work`, `offline` |

**Exit Codes**:

| Code | Meaning |
|------|---------|
| 0 | Status updated (or no-op outside tmux) |
| 1 | Invalid status name |

**Stdout**: Empty (silent on success)

**Stderr** (invalid status):
```
Invalid status: foo. Valid: ready, work, offline
```

**Behavior**:
1. Check if running inside tmux (`$TMUX` env var)
2. If not in tmux: exit 0 silently (no-op for non-tmux environments)
3. Get current tmux window name
4. Parse status argument:
   - If not one of `ready`, `work`, `offline`: print error to stderr, exit 1
5. Update `.git/mail-recipients.jsonl`:
   - Read existing file (create if not exists)
   - Find or create entry for current window
   - Update status and `updated_at` timestamp
   - If transitioning to `work` or `offline`: reset `notified` to false
   - Write back with file locking
6. Exit 0

---

## Notification Protocol

When mailman sends a notification to an agent:

**Step 1**: Send message text
```bash
tmux send-keys -t <window> "Check your agentmail"
```

**Step 2**: Wait 1 second
```go
time.Sleep(1 * time.Second)
```

**Step 3**: Send Enter key
```bash
tmux send-keys -t <window> Enter
```

**Rationale**: The 1-second delay allows the agent's input buffer to receive the text before Enter is pressed, ensuring reliable delivery.

---

## File Contracts

### .git/mail/mailman.pid

**Format**: Plain text, integer PID
**Created by**: `agentmail mailman`
**Deleted by**: `agentmail mailman` on clean shutdown (SIGTERM/SIGINT)

**Example**:
```
12345
```

### .git/mail-recipients.jsonl

**Format**: JSONL, one RecipientState per line
**Created by**: `agentmail status` (on first status update)
**Updated by**: `agentmail status`, `agentmail mailman` (stale cleanup, notified flag)

**Example**:
```json
{"recipient":"agent1","status":"ready","updated_at":"2026-01-12T10:00:00Z","notified":false}
{"recipient":"agent2","status":"work","updated_at":"2026-01-12T10:01:30Z","notified":true}
```

---

## Integration with Existing Commands

The mailman daemon uses existing infrastructure:

| Component | Used For |
|-----------|----------|
| `internal/tmux.InTmux()` | Check if in tmux session |
| `internal/tmux.GetCurrentWindow()` | Get window name for status |
| `internal/mail.FindUnread()` | Check for unread messages |
| `internal/mail.EnsureMailDir()` | Create `.git/mail/` if missing |

---

## Claude Code Hooks Integration

The `agentmail status` command is designed for Claude Code hooks:

**hooks.json example**:
```json
{
  "hooks": [
    {
      "matcher": "SessionStart",
      "hooks": [
        {
          "type": "command",
          "command": "agentmail status ready"
        }
      ]
    },
    {
      "matcher": "Stop",
      "hooks": [
        {
          "type": "command",
          "command": "agentmail status ready"
        }
      ]
    },
    {
      "matcher": "SessionEnd",
      "hooks": [
        {
          "type": "command",
          "command": "agentmail status offline"
        }
      ]
    },
    {
      "matcher": "UserPromptSubmit",
      "hooks": [
        {
          "type": "command",
          "command": "agentmail status work"
        }
      ]
    }
  ]
}
```

**Key Design Points**:
- Silent output: hooks should not pollute agent output
- Always exits 0 outside tmux: works in any environment without errors
- Fast execution: minimal overhead for hook invocation

# Quickstart: Mailman Daemon

**Feature**: 006-mailman-daemon
**Date**: 2026-01-12

## Overview

The mailman daemon provides automated notifications to agents when they have unread messages. Instead of agents polling for messages, the mailman monitors all mailboxes and sends tmux notifications to agents in "ready" state.

## Prerequisites

- AgentMail installed and working (`agentmail send/receive`)
- Running inside a tmux session
- Multiple tmux windows (one per agent)

## Quick Start

### 1. Start the Mailman Daemon

In a dedicated tmux window:

```bash
# Foreground mode (see activity)
agentmail mailman

# Or background mode (daemonize)
agentmail mailman --daemon
```

### 2. Register Agent Status

In each agent's tmux window, register availability:

```bash
# Mark agent as ready to receive notifications
agentmail status ready

# Mark agent as busy (no notifications)
agentmail status work

# Mark agent as offline
agentmail status offline
```

### 3. Send a Message

From any agent:

```bash
agentmail send agent1 "Hello from agent2!"
```

### 4. Receive Notification

If `agent1` is in "ready" status, they will see:
```
Check your agentmail
```
...typed into their tmux window, prompting them to run `agentmail receive`.

## Claude Code Hooks Integration

For automated status management, add hooks to `.claude/hooks.json`:

```json
{
  "hooks": [
    {
      "matcher": "SessionStart",
      "hooks": [{"type": "command", "command": "agentmail status ready"}]
    },
    {
      "matcher": "Stop",
      "hooks": [{"type": "command", "command": "agentmail status ready"}]
    },
    {
      "matcher": "SessionEnd",
      "hooks": [{"type": "command", "command": "agentmail status offline"}]
    },
    {
      "matcher": "UserPromptSubmit",
      "hooks": [{"type": "command", "command": "agentmail status work"}]
    }
  ]
}
```

This automatically:
- Sets "ready" when agent starts or finishes a task
- Sets "work" when processing a user message
- Sets "offline" when agent session ends

## Typical Workflow

```
┌─────────────────────────────────────────────────────────┐
│                    tmux session                          │
├─────────────┬─────────────┬─────────────┬───────────────┤
│   mailman   │   agent1    │   agent2    │    agent3     │
│             │             │             │               │
│  (daemon)   │  (ready)    │  (work)     │  (ready)      │
│             │             │             │               │
│  checks     │  receives   │  busy, no   │  receives     │
│  mailboxes  │  "Check     │  notif      │  "Check       │
│  every 10s  │  your       │             │  your         │
│             │  agentmail" │             │  agentmail"   │
└─────────────┴─────────────┴─────────────┴───────────────┘
```

## Stopping the Daemon

```bash
# Find the PID
cat .git/mail/mailman.pid

# Stop gracefully
kill <pid>

# Or if running in foreground, just Ctrl+C
```

## Troubleshooting

### Daemon won't start

```
Mailman daemon already running (PID: 12345)
```

Check if daemon is actually running:
```bash
ps aux | grep mailman
```

If not running, the PID file is stale. Starting mailman again will clean it up automatically.

### No notifications received

1. Check agent status: Is recipient in "ready" state?
2. Check daemon is running: `cat .git/mail/mailman.pid && ps aux | grep <pid>`
3. Check tmux window name matches: `tmux display-message -p '#W'`
4. Check for unread messages: `agentmail receive` (in recipient window)

### Status command does nothing

This is expected if not running in tmux. The command silently succeeds outside tmux to support hooks in non-tmux environments.

## Files Created

| File | Purpose |
|------|---------|
| `.git/mail/mailman.pid` | Daemon process ID (singleton lock) |
| `.git/mail-recipients.jsonl` | Agent status tracking |

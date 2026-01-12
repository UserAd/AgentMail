# Data Model: Claude Code Hooks Integration

**Feature**: 005-claude-hooks-integration
**Date**: 2026-01-12

## Overview

This feature does not introduce new data entities. It extends existing CLI behavior with a boolean flag.

## Modified Entities

### ReceiveOptions (internal/cli/receive.go)

**Current Structure**:
```go
type ReceiveOptions struct {
    SkipTmuxCheck bool     // Skip tmux environment check
    MockWindows   []string // Mock list of tmux windows
    MockReceiver  string   // Mock receiver window name
    RepoRoot      string   // Repository root (defaults to current directory)
}
```

**Extended Structure**:
```go
type ReceiveOptions struct {
    SkipTmuxCheck bool     // Skip tmux environment check
    MockWindows   []string // Mock list of tmux windows
    MockReceiver  string   // Mock receiver window name
    RepoRoot      string   // Repository root (defaults to current directory)
    HookMode      bool     // NEW: Enable hook mode behavior
}
```

### Exit Code Semantics

| Exit Code | Normal Mode | Hook Mode |
|-----------|-------------|-----------|
| 0 | Success (message displayed or no messages) | Silent exit (no messages, not in tmux, or error) |
| 1 | Error (file read, lock, etc.) | Not used in hook mode |
| 2 | Not in tmux session | Message notification (has unread messages) |

### Output Streams

| Stream | Normal Mode | Hook Mode |
|--------|-------------|-----------|
| STDOUT | Message content, "No unread messages" | Not used |
| STDERR | Error messages | Notification: "You got new mail\n" + message |

## State Transitions

```
┌─────────────────────────────────────────────────────────────────┐
│                        agentmail receive --hook                  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
                    ┌───────────────────────┐
                    │    Check tmux env     │
                    └───────────────────────┘
                                │
                    ┌───────────┴───────────┐
                    │                       │
                    ▼                       ▼
            ┌───────────┐           ┌───────────┐
            │  In tmux  │           │ Not tmux  │
            └───────────┘           └───────────┘
                    │                       │
                    ▼                       ▼
            ┌───────────────┐       ┌───────────────┐
            │ Check mailbox │       │  Exit 0       │
            └───────────────┘       │  (silent)     │
                    │               └───────────────┘
        ┌───────────┼───────────┐
        │           │           │
        ▼           ▼           ▼
┌───────────┐ ┌───────────┐ ┌───────────┐
│  Error    │ │ No msgs   │ │ Has msgs  │
└───────────┘ └───────────┘ └───────────┘
        │           │           │
        ▼           ▼           ▼
┌───────────┐ ┌───────────┐ ┌───────────────────┐
│  Exit 0   │ │  Exit 0   │ │ Write to STDERR:  │
│ (silent)  │ │ (silent)  │ │ "You got new mail"│
└───────────┘ └───────────┘ │ + message content │
                            │ Exit 2            │
                            └───────────────────┘
```

## Validation Rules

| Rule | Requirement | Validation |
|------|-------------|------------|
| HookMode flag | Boolean, defaults to `false` | FR-001a, FR-005 |
| Exit code 0 | No messages, not in tmux, or any error | FR-002, FR-003, FR-004a/b/c |
| Exit code 2 | Unread messages exist (notification) | FR-001b |
| No exit code 1 | Hook mode never uses exit code 1 | FR-004a/b/c |
| Output to STDERR | All output when HookMode is true | FR-005 |
| No output cases | No messages, not in tmux, errors | FR-002, FR-003, FR-004a/b/c |
| Message consumption | Oldest message marked as read | FR-001c |

## No New Storage

This feature does not modify:
- Message format in `.git/mail/*.jsonl`
- Mailbox file structure
- Message ID generation
- File locking behavior

# Data Model: Recipients Command, Help Flag, and Stdin Message Input

**Feature**: 003-recipients-help-stdin
**Date**: 2026-01-12

## Entities

### 1. IgnoreList

Represents the set of window names to exclude from recipient discovery and message sending.

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| entries | map[string]bool | Set of window names to ignore | Non-empty strings only |
| sourcePath | string | Path to `.agentmailignore` file | Must be in git root |

**Lifecycle**:
- Created: When `parseIgnoreFile()` is called
- Updated: Not updated (re-read on each command invocation)
- Deleted: N/A (in-memory only)

**Relationships**:
- Used by: Recipients command, Send command validation

---

### 2. Recipient

Represents a valid message recipient (tmux window eligible for messaging).

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| name | string | Tmux window name | Non-empty, exists in session |
| isCurrent | bool | Whether this is the current window | Used for "[you]" marker display |
| isValid | bool | Whether recipient can receive messages | Not in ignore list, not current window |

**Validation Rules**:
- Must be an active tmux window in current session
- Must NOT be the current window for sending (self-messaging not allowed)
- Must NOT be listed in `.agentmailignore`

**Display Rules**:
- Current window is shown with "[you]" suffix in recipients list
- Current window is NOT a valid send target (still gets "recipient not found" error)

---

### 3. HelpContent (Static)

Represents the help documentation displayed by `--help`.

| Field | Type | Description |
|-------|------|-------------|
| usage | string | General usage pattern |
| commands | []CommandHelp | List of available commands |
| examples | []string | Usage examples |

**CommandHelp Structure**:
| Field | Type | Description |
|-------|------|-------------|
| name | string | Command name (send, receive, recipients) |
| syntax | string | Command syntax with placeholders |
| description | string | One-line description |

---

## State Transitions

### Recipient Display State (for `recipients` command)

```
┌─────────────────┐
│  Window Name    │
│   (from tmux)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     Yes     ┌──────────────────┐
│ In ignore list? ├────────────►│ Skip (not shown) │
└────────┬────────┘             └──────────────────┘
         │ No
         ▼
┌─────────────────┐     Yes     ┌──────────────────┐
│ Is current?     ├────────────►│ Show: "name [you]"│
└────────┬────────┘             └──────────────────┘
         │ No
         ▼
┌─────────────────┐
│ Show: "name"    │
└─────────────────┘
```

### Recipient Validation State (for `send` command)

```
┌─────────────────┐
│  Window Name    │
│   (Input)       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     No      ┌──────────────────┐
│ Exists in tmux? ├────────────►│ Error: not found │
└────────┬────────┘             └──────────────────┘
         │ Yes
         ▼
┌─────────────────┐     Yes     ┌──────────────────┐
│ In ignore list? ├────────────►│ Error: not found │
└────────┬────────┘             └──────────────────┘
         │ No
         ▼
┌─────────────────┐     Yes     ┌──────────────────┐
│ Is self?        ├────────────►│ Error: not found │
└────────┬────────┘             └──────────────────┘
         │ No
         ▼
┌─────────────────┐
│  Valid Recipient│
└─────────────────┘
```

---

## File Formats

### .agentmailignore

**Location**: Git repository root (same directory as `.git/`)

**Format**: Plain text, one window name per line

```text
# Example .agentmailignore
monitor
logs
debug-window
```

**Parsing Rules**:
1. Each line is a window name to exclude
2. Empty lines are ignored
3. Whitespace-only lines are ignored
4. Leading/trailing whitespace is trimmed
5. No comment syntax (lines starting with # are treated as window names)
6. Case-sensitive matching
7. No glob or regex patterns

---

## Integration Points

### Existing Entities (Unchanged)

| Entity | Location | Used By |
|--------|----------|---------|
| Message | internal/mail/message.go | Send command (existing) |
| Mailbox | internal/mail/mailbox.go | Send/Receive commands (existing) |

### New Integration

| New Entity | Integrates With | Purpose |
|------------|-----------------|---------|
| IgnoreList | tmux.ListWindows() | Filter recipients |
| IgnoreList | cli.Send() | Validate recipient |
| Recipient | tmux.WindowExists() | Validate existence |

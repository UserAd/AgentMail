# Data Model: Storage Directory Restructure

**Feature**: 007-storage-restructure
**Date**: 2026-01-13

## Overview

This document describes the data model for AgentMail's restructured storage system. The change is purely structural—file formats remain unchanged.

## Directory Structure

### New Layout

```
<repo-root>/
├── .git/                          # Git repository (required)
└── .agentmail/                    # AgentMail root directory
    ├── mailboxes/                 # Per-recipient mailbox files
    │   ├── <recipient1>.jsonl     # Mailbox for recipient1
    │   ├── <recipient2>.jsonl     # Mailbox for recipient2
    │   └── ...
    ├── recipients.jsonl           # Recipient state tracking
    └── mailman.pid                # Daemon process ID (when running)
```

### Previous Layout (Deprecated)

```
<repo-root>/
├── .git/
│   └── mail/                      # OLD: Mail directory inside .git
│       ├── <recipient>.jsonl      # OLD: Mailbox files
│       └── mailman.pid            # OLD: PID file location
└── .git/mail-recipients.jsonl     # OLD: Recipients at repo root level
```

## Entities

### Storage Root

| Property | Value |
|----------|-------|
| Path | `.agentmail/` |
| Permissions | 0750 (drwxr-x---) |
| Created By | Any agentmail command |
| Contains | mailboxes/, recipients.jsonl, mailman.pid |

### Mailboxes Directory

| Property | Value |
|----------|-------|
| Path | `.agentmail/mailboxes/` |
| Permissions | 0750 (drwxr-x---) |
| Created By | `agentmail send` command |
| Contains | Per-recipient JSONL files |

### Mailbox File

| Property | Value |
|----------|-------|
| Path Pattern | `.agentmail/mailboxes/<recipient>.jsonl` |
| Permissions | 0640 (-rw-r-----) |
| Format | JSONL (one JSON object per line) |
| Locking | File-level exclusive lock during write |

**Message Record Structure** (unchanged):

```json
{
  "id": "string (8 char alphanumeric)",
  "from": "string (sender tmux window name)",
  "to": "string (recipient tmux window name)",
  "body": "string (message content)",
  "timestamp": "string (RFC3339 format)",
  "read": "boolean"
}
```

### Recipients File

| Property | Value |
|----------|-------|
| Path | `.agentmail/recipients.jsonl` |
| Permissions | 0640 (-rw-r-----) |
| Format | JSONL (one JSON object per line) |
| Locking | File-level exclusive lock during write |

**Recipient State Structure** (unchanged):

```json
{
  "recipient": "string (tmux window name)",
  "status": "string (ready|work|offline)",
  "updated_at": "string (RFC3339 format)",
  "notified": "boolean"
}
```

### PID File

| Property | Value |
|----------|-------|
| Path | `.agentmail/mailman.pid` |
| Permissions | 0640 (-rw-r-----) |
| Format | Plain text (single integer) |
| Created By | `agentmail mailman start` |
| Removed By | `agentmail mailman stop` or daemon exit |

**Content**: Single line containing the process ID of the running mailman daemon.

## Path Constants

### Go Constants (New)

```go
// internal/mail/mailbox.go
const (
    RootDir = ".agentmail"           // Storage root directory
    MailDir = ".agentmail/mailboxes" // Mailbox files directory
)

// internal/mail/recipients.go
const RecipientsFile = ".agentmail/recipients.jsonl"

// internal/daemon/daemon.go
const PIDFile = "mailman.pid"  // Relative to RootDir
```

### Path Construction

| Entity | Path Expression |
|--------|-----------------|
| Root | `filepath.Join(repoRoot, RootDir)` |
| Mailboxes | `filepath.Join(repoRoot, MailDir)` |
| Mailbox | `filepath.Join(repoRoot, MailDir, recipient+".jsonl")` |
| Recipients | `filepath.Join(repoRoot, RecipientsFile)` |
| PID | `filepath.Join(repoRoot, RootDir, PIDFile)` |

## Validation Rules

### Directory Names

- `RootDir` must be `.agentmail` (constant)
- `MailDir` must be `.agentmail/mailboxes` (constant)

### Recipient Names

- Must be valid tmux window names
- No path separators (`/`, `\`)
- No path traversal sequences (`..`)
- Sanitized before use in file paths (existing validation)

### File Operations

- Directory creation uses `os.MkdirAll` with 0750 permissions
- File creation uses 0640 permissions
- All write operations use exclusive file locking
- Read operations use shared locking where applicable

## State Transitions

No state machine changes. The storage restructure does not affect:
- Message lifecycle (unread → read)
- Recipient status transitions (ready ↔ work ↔ offline)
- Daemon lifecycle (stopped → running → stopped)

## Migration Path

### Manual Migration Steps

```bash
# 1. Stop daemon if running
agentmail mailman stop

# 2. Create new directory structure
mkdir -p .agentmail/mailboxes

# 3. Move mailbox files
mv .git/mail/*.jsonl .agentmail/mailboxes/

# 4. Move recipients file
mv .git/mail-recipients.jsonl .agentmail/recipients.jsonl

# 5. Clean up old directories (optional)
rm -rf .git/mail
```

### Verification

```bash
# After migration, verify:
ls -la .agentmail/
ls -la .agentmail/mailboxes/
agentmail recipients  # Should list existing recipients
agentmail receive     # Should receive any pending messages
```

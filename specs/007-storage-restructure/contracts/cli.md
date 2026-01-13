# CLI Contract: Storage Directory Restructure

**Feature**: 007-storage-restructure
**Date**: 2026-01-13

## Overview

This document defines the CLI contract changes for the storage restructure. **The CLI interface itself does not change**—only the internal storage paths are modified.

## Commands (Unchanged Interface)

### agentmail send

```
agentmail send <recipient> "<message>"
```

**Storage Change**:
- Old: Creates `.git/mail/<recipient>.jsonl`
- New: Creates `.agentmail/mailboxes/<recipient>.jsonl`

**Directory Creation**:
- Old: Created `.git/mail/` if missing
- New: Creates `.agentmail/` and `.agentmail/mailboxes/` if missing

### agentmail receive

```
agentmail receive
```

**Storage Change**:
- Old: Reads from `.git/mail/<current-window>.jsonl`
- New: Reads from `.agentmail/mailboxes/<current-window>.jsonl`

### agentmail recipients

```
agentmail recipients
```

**Storage Change**:
- Old: Reads from `.git/mail-recipients.jsonl`
- New: Reads from `.agentmail/recipients.jsonl`

### agentmail mailman start

```
agentmail mailman start
```

**Storage Change**:
- Old: Creates `.git/mail/mailman.pid`
- New: Creates `.agentmail/mailman.pid`

**Directory Creation**:
- Old: Created `.git/mail/` if missing
- New: Creates `.agentmail/` if missing

### agentmail mailman stop

```
agentmail mailman stop
```

**Storage Change**:
- Old: Reads/removes `.git/mail/mailman.pid`
- New: Reads/removes `.agentmail/mailman.pid`

### agentmail mailman status

```
agentmail mailman status
```

**Storage Change**:
- Old: Reads `.git/mail/mailman.pid`
- New: Reads `.agentmail/mailman.pid`

### agentmail status

```
agentmail status
```

**Storage Change**:
- Old: Reads from `.git/mail/` and `.git/mail-recipients.jsonl`
- New: Reads from `.agentmail/mailboxes/` and `.agentmail/recipients.jsonl`

## Exit Codes (Unchanged)

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (command failed) |
| 2 | Environment error (not in tmux, no git repo) |

## Error Messages

### New Error Scenarios

| Scenario | Error Message |
|----------|--------------|
| Cannot create `.agentmail/` | `Error: failed to create .agentmail directory: <os error>` |
| Cannot create `.agentmail/mailboxes/` | `Error: failed to create mailboxes directory: <os error>` |

### Unchanged Error Scenarios

- Not in tmux session
- Not in git repository
- Invalid recipient name
- File permission errors

## Internal API Changes

### mail package

```go
// New constant
const RootDir = ".agentmail"

// Updated constant
const MailDir = ".agentmail/mailboxes"  // was ".git/mail"

// Updated function signature (behavior change only)
func EnsureMailDir(repoRoot string) error
// Now creates both RootDir and MailDir

// New function (optional, for clarity)
func EnsureRootDir(repoRoot string) error
// Creates only RootDir
```

### daemon package

```go
// Updated function
func PIDFilePath(repoRoot string) string
// Returns: filepath.Join(repoRoot, mail.RootDir, PIDFile)
// Was: filepath.Join(repoRoot, mail.MailDir, PIDFile)
```

## Backward Compatibility

| Aspect | Compatibility |
|--------|---------------|
| CLI flags | Fully compatible (no changes) |
| CLI output format | Fully compatible (no changes) |
| Exit codes | Fully compatible (no changes) |
| Data format | Fully compatible (JSONL unchanged) |
| Data location | **Breaking change** (new paths) |

## Testing Contract

### Unit Tests

Each test must verify:
1. Correct path constants are used
2. Directory creation creates correct hierarchy
3. File operations target correct paths

### Integration Tests

Each test must verify:
1. End-to-end flow works with new paths
2. No references to old paths in runtime behavior
3. Error messages reference correct paths

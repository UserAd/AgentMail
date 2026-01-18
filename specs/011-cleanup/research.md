# Research: Cleanup Command

**Feature**: 011-cleanup
**Date**: 2026-01-15

## Overview

This document captures research findings for the cleanup command implementation. Since this feature uses only Go standard library and follows existing codebase patterns, research focuses on design decisions rather than technology evaluation.

## Research Topics

### 1. Message Timestamp Field

**Decision**: Add `created_at` field to Message struct using RFC 3339 format (`time.Time` with JSON marshaling)

**Rationale**:
- Matches existing `updated_at` field in RecipientState for consistency
- RFC 3339 is human-readable and sortable
- `time.Time` marshals to RFC 3339 by default in Go's encoding/json
- Using `omitempty` ensures backward compatibility with existing messages

**Alternatives Considered**:
- Unix timestamp (int64): Rejected - less readable, harder to debug
- `sent_at` naming: Rejected - `created_at` matches existing codebase convention

**Implementation**:
```go
type Message struct {
    // ... existing fields ...
    CreatedAt time.Time `json:"created_at,omitempty"` // Timestamp for age-based cleanup
}
```

### 2. Non-blocking File Locking

**Decision**: Use `syscall.Flock` with `LOCK_NB` flag and 1-second retry loop

**Rationale**:
- Existing codebase uses `syscall.Flock` for file locking
- Non-blocking (LOCK_NB) prevents cleanup from blocking indefinitely
- 1-second timeout balances responsiveness with lock acquisition chance
- Skipping locked files is safe - cleanup can be retried

**Alternatives Considered**:
- Blocking lock: Rejected - could hang cleanup on long operations
- No lock timeout (fail immediately): Rejected - too aggressive, transient locks common

**Implementation Pattern**:
```go
// Try non-blocking lock
err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
if err != nil {
    // Retry with timeout
    deadline := time.Now().Add(time.Second)
    for time.Now().Before(deadline) {
        time.Sleep(10 * time.Millisecond)
        err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
        if err == nil {
            break
        }
    }
    if err != nil {
        // Skip this file with warning
        return ErrFileLocked
    }
}
```

### 3. Offline Recipient Detection

**Decision**: Use existing `tmux.ListWindows()` to get current windows, compare against recipients.jsonl entries

**Rationale**:
- ListWindows already implemented and tested
- Simple set comparison (recipients not in windows = offline)
- Gracefully handles non-tmux environment (skip check, warn user)

**Alternatives Considered**:
- Check each recipient individually with WindowExists: Rejected - N calls vs 1 call
- Use `last_read_at` as activity indicator: Rejected - doesn't detect closed windows

### 4. Cleanup Execution Order

**Decision**: Execute cleanup in this order: (1) offline recipients, (2) stale recipients, (3) old messages, (4) empty mailboxes

**Rationale**:
- Removing offline recipients first prevents wasted effort on messages for removed recipients
- Stale check runs after offline check (offline removal may satisfy staleness too)
- Message cleanup before mailbox removal ensures accurate empty detection
- Each phase independent - partial completion still useful

### 5. Summary Output Format

**Decision**: Simple text output with counts

**Rationale**:
- Matches CLI-first principle
- Human-readable for manual execution
- Machine-parseable for scripting

**Format**:
```
Cleanup complete:
  Recipients removed: 3 (2 offline, 1 stale)
  Messages removed: 15
  Mailboxes removed: 2
```

**Dry-run Format**:
```
Cleanup preview (dry-run):
  Recipients to remove: 3 (2 offline, 1 stale)
  Messages to remove: 15
  Mailboxes to remove: 2
```

## Dependencies

No new dependencies required. All functionality implemented with Go standard library:
- `time` - timestamps, duration comparisons
- `syscall` - file locking (existing pattern)
- `os` - file operations
- `encoding/json` - JSONL parsing
- `os/exec` - tmux integration (existing pattern)

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Existing messages lack timestamp | Certain | Medium | Skip messages without `created_at` (documented in spec) |
| Concurrent cleanup races | Low | Low | Non-blocking locks, skip-and-warn pattern |
| tmux not available | Low | Low | Skip offline check, continue with other cleanup |

## Conclusion

No blocking unknowns. Implementation can proceed with Phase 1 design.

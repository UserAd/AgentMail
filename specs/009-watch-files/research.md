# Research: File-Watching for Mailman with Timer Fallback

**Feature**: 009-watch-files
**Date**: 2026-01-13
**Status**: Complete

## Research Questions

### 1. Cross-Platform File Watching Library

**Question**: What Go library should be used for cross-platform file watching?

**Decision**: Use `github.com/fsnotify/fsnotify` v1.9.0

**Rationale**:
- Industry standard for Go file watching (12,768+ importing packages)
- Cross-platform support: Linux (inotify), macOS/BSD (kqueue), Windows (ReadDirectoryChangesW), illumos (FEN)
- Actively maintained (latest release April 4, 2025)
- Simple API: `NewWatcher()`, `Add(path)`, event channel
- Used by major Go projects (Docker, Kubernetes, Hugo, etc.)

**Alternatives Considered**:
| Alternative | Rejected Because |
|-------------|------------------|
| Standard library `os` + syscall | Go stdlib has no cross-platform file watching; would require separate implementations for inotify (Linux), FSEvents (macOS), ReadDirectoryChangesW (Windows) |
| Polling only | Defeats the purpose; 10-second polling is the fallback, not the primary mechanism |
| tilt-dev/fsnotify fork | Fork with minor changes; original is more widely used and maintained |

**Constitution Justification** (IV. Standard Library Preference):
- Standard library is insufficient: Go has no `os.Watch()` or equivalent
- Security/maintenance: fsnotify is widely audited, single-purpose, no transitive dependencies
- fsnotify uses only platform system calls (inotify, kqueue, etc.) - no additional external dependencies

### 2. Debouncing Strategy

**Question**: How to implement 500ms debounce for rapid file change events?

**Decision**: Use a timer-based debounce with event coalescing

**Rationale**:
- File systems often generate multiple events for single logical operation (e.g., Write followed by Chmod)
- JSONL append generates Write event; atomic write generates Create/Rename events
- 500ms window catches most burst scenarios without adding noticeable latency

**Implementation Pattern**:
```go
type Debouncer struct {
    timer    *time.Timer
    duration time.Duration
    mu       sync.Mutex
}

func (d *Debouncer) Trigger(callback func()) {
    d.mu.Lock()
    defer d.mu.Unlock()

    if d.timer != nil {
        d.timer.Stop()
    }
    d.timer = time.AfterFunc(d.duration, callback)
}
```

**Alternatives Considered**:
| Alternative | Rejected Because |
|-------------|------------------|
| Process every event immediately | Would cause duplicate notifications for burst writes |
| Fixed-window batching | Less responsive than trailing-edge debounce |
| Leading-edge debounce | Would miss subsequent events in burst |

### 3. Fallback Detection Strategy

**Question**: How to detect when file watching is unavailable?

**Decision**: Try-catch initialization with automatic fallback

**Rationale**:
- `fsnotify.NewWatcher()` returns error if OS support unavailable
- `watcher.Add(path)` returns error if path can't be watched (e.g., network filesystem)
- Any runtime error during event processing triggers fallback

**Detection Points**:
1. **Initialization**: `NewWatcher()` fails → immediate fallback
2. **Path addition**: `Add()` fails → immediate fallback
3. **Runtime error**: Event channel error → switch to fallback

**Implementation Pattern**:
```go
watcher, err := fsnotify.NewWatcher()
if err != nil {
    log("File watching unavailable, using polling")
    return runPollingMode(opts)
}

err = watcher.Add(mailboxDir)
if err != nil {
    log("Cannot watch %s, using polling: %v", mailboxDir, err)
    watcher.Close()
    return runPollingMode(opts)
}
```

### 4. Directory vs File Watching

**Question**: Should we watch directories or individual files?

**Decision**: Watch directories (`.agentmail/` and `.agentmail/mailboxes/`)

**Rationale**:
- fsnotify docs recommend watching directories, not files
- Editors often use atomic writes (write temp → rename) which breaks file watches
- Directory watches automatically capture new files created after watcher starts
- Only need 2 watches: `.agentmail/` (for recipients.jsonl) and `.agentmail/mailboxes/`

**File Identification**:
- Filter events by `Event.Name` to identify which file changed
- Check if `Event.Name` ends with `.jsonl` for mailbox files
- Check if `Event.Name` matches `recipients.jsonl` for status changes

### 5. Last-Read Timestamp Storage

**Question**: Where and how to store last-read timestamps?

**Decision**: Add `last_read_at` field to existing `RecipientState` struct in `recipients.jsonl`

**Rationale**:
- Reuses existing storage mechanism with file locking
- Consistent with existing recipient state tracking
- No new files or storage formats needed

**Schema Change**:
```go
type RecipientState struct {
    Recipient  string    `json:"recipient"`
    Status     string    `json:"status"`
    UpdatedAt  time.Time `json:"updated_at"`
    Notified   bool      `json:"notified"`
    LastReadAt int64     `json:"last_read_at,omitempty"`  // NEW: Unix timestamp in milliseconds
}
```

**Alternatives Considered**:
| Alternative | Rejected Because |
|-------------|------------------|
| Separate last-read.jsonl file | Adds complexity; recipients.jsonl already tracks agent state |
| Store in mailbox file | Violates separation of concerns; mailbox is for messages |
| In-memory only | Data lost on daemon restart |

### 6. Fallback Timer Duration

**Question**: What interval for the safety fallback timer during file-watching mode?

**Decision**: 60 seconds

**Rationale**:
- Long enough to not interfere with normal event-driven operation
- Short enough to catch missed events within reasonable time
- Spec requires this per FR-009

### 7. Non-Recursive Directory Watching

**Question**: How to handle fsnotify's non-recursive watching limitation?

**Decision**: Watch only two directories explicitly

**Rationale**:
- AgentMail uses flat directory structure (no subdirectories in mailboxes)
- Only need to watch:
  1. `.agentmail/` - for `recipients.jsonl` changes
  2. `.agentmail/mailboxes/` - for mailbox file changes
- No recursive watching needed

## Implementation Notes

### Event Processing Loop

```go
for {
    select {
    case event := <-watcher.Events:
        if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
            debouncer.Trigger(processChanges)
        }
    case err := <-watcher.Errors:
        log("Watcher error: %v, falling back to polling", err)
        watcher.Close()
        return runPollingMode(opts)
    case <-fallbackTicker.C:
        processChanges() // Safety net
    case <-stopChan:
        return
    }
}
```

### Testing Strategy

1. **Unit tests**: Mock fsnotify.Watcher interface for isolated testing
2. **Integration tests**: Use real file operations with short debounce
3. **Fallback tests**: Force watcher initialization failure to verify fallback
4. **Race tests**: `go test -race` to catch concurrency issues

## Sources

- [fsnotify GitHub Repository](https://github.com/fsnotify/fsnotify)
- [fsnotify Go Package Documentation](https://pkg.go.dev/github.com/fsnotify/fsnotify)

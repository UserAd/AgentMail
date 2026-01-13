# Data Model: File-Watching for Mailman with Timer Fallback

**Feature**: 009-watch-files
**Date**: 2026-01-13

## Entity Changes

### RecipientState (Modified)

Existing entity in `internal/mail/recipients.go` that tracks agent availability state.

**Current Schema**:
```go
type RecipientState struct {
    Recipient string    `json:"recipient"`
    Status    string    `json:"status"`      // "ready" | "work" | "offline"
    UpdatedAt time.Time `json:"updated_at"`
    Notified  bool      `json:"notified"`
}
```

**New Schema** (FR-017, FR-018):
```go
type RecipientState struct {
    Recipient  string    `json:"recipient"`
    Status     string    `json:"status"`               // "ready" | "work" | "offline"
    UpdatedAt  time.Time `json:"updated_at"`
    Notified   bool      `json:"notified"`
    LastReadAt int64     `json:"last_read_at,omitempty"` // NEW: Unix timestamp (ms)
}
```

**Field Details**:
| Field | Type | Description | Constraints |
|-------|------|-------------|-------------|
| `recipient` | string | Agent's tmux window name | Required, unique identifier |
| `status` | string | Current availability status | One of: "ready", "work", "offline" |
| `updated_at` | time.Time | Last status change timestamp | ISO 8601 format |
| `notified` | bool | Whether agent has been notified | Reset when status changes to work/offline |
| `last_read_at` | int64 | When agent last read mail | Unix timestamp in milliseconds, optional |

**JSONL Example**:
```jsonl
{"recipient":"agent1","status":"ready","updated_at":"2026-01-13T10:30:00Z","notified":false,"last_read_at":1736764200000}
{"recipient":"agent2","status":"work","updated_at":"2026-01-13T10:35:00Z","notified":true}
```

**Validation Rules**:
- `last_read_at` is optional (agents that never called `receive` won't have it)
- `last_read_at` must be a positive Unix timestamp in milliseconds when present
- `last_read_at` is only updated when `agentmail receive` is called inside tmux (FR-021)

---

## New Entities

### MonitoringMode (New)

Runtime state indicating the daemon's current monitoring strategy.

**Schema**:
```go
type MonitoringMode int

const (
    ModeWatching MonitoringMode = iota  // Event-driven file watching
    ModePolling                          // Timer-based polling (fallback)
)
```

**State Transitions**:
```
                  ┌─────────────────┐
    Start ───────►│   ModeWatching  │
                  └────────┬────────┘
                           │ Watcher error (FR-014b)
                           ▼
                  ┌─────────────────┐
                  │   ModePolling   │◄──── Start (if init fails, FR-003b)
                  └─────────────────┘
                           │
                           │ No recovery (FR-015)
                           ▼
                     (until restart)
```

---

### Debouncer (New)

Mechanism to coalesce rapid file change events.

**Schema**:
```go
type Debouncer struct {
    timer    *time.Timer     // Active timer (nil if no pending trigger)
    duration time.Duration   // Debounce window (500ms per FR-011)
    mu       sync.Mutex      // Protects timer access
}
```

**Behavior**:
- Each `Trigger()` call resets the timer
- Callback fires only after no triggers for `duration`
- Thread-safe via mutex

---

### FileWatcher (New)

Abstraction over fsnotify for file system monitoring.

**Schema**:
```go
type FileWatcher struct {
    watcher      *fsnotify.Watcher
    debouncer    *Debouncer
    mailboxDir   string           // Path to .agentmail/mailboxes/
    agentmailDir string           // Path to .agentmail/
    stopChan     chan struct{}
    mode         MonitoringMode
    mu           sync.Mutex
}
```

**Watched Paths** (FR-004, FR-005, FR-001):
| Path | Purpose |
|------|---------|
| `.agentmail/` | Watch for `recipients.jsonl` changes (FR-005) |
| `.agentmail/mailboxes/` | Watch for mailbox file changes (FR-004) |

**Event Handling**:
| Event Type | Action |
|------------|--------|
| `Write` on mailbox file | Trigger notification check (debounced) |
| `Create` on mailbox file | Trigger notification check (debounced) |
| `Write` on `recipients.jsonl` | Reload states, check notifications (debounced) |
| `Error` | Log and switch to polling mode |

---

## Relationships

```
┌─────────────────┐
│   FileWatcher   │
└────────┬────────┘
         │ uses
         ▼
┌─────────────────┐       triggers      ┌──────────────────┐
│    Debouncer    │ ──────────────────► │ CheckAndNotify() │
└─────────────────┘                     └────────┬─────────┘
                                                 │ reads
                                                 ▼
                                        ┌──────────────────┐
                                        │  RecipientState  │
                                        │  (recipients.jsonl)│
                                        └──────────────────┘
                                                 ▲
                                                 │ updates
┌─────────────────┐                              │
│ Receive command │ ─────────────────────────────┘
└─────────────────┘   sets last_read_at
```

---

## File Locking

All updates to `recipients.jsonl` use existing file locking mechanism via `syscall.Flock`:

1. **Status updates** (`UpdateRecipientState`): Existing, unchanged
2. **Notified flag** (`SetNotifiedFlag`): Existing, unchanged
3. **Last-read updates** (NEW): Uses same locking pattern

**New Function Signature**:
```go
// UpdateLastReadAt sets the last_read_at timestamp for a recipient.
// Creates recipient entry if it doesn't exist.
// Uses file locking to prevent race conditions.
func UpdateLastReadAt(repoRoot string, recipient string, timestamp int64) error
```

---

## Migration

**Backward Compatibility**: The `last_read_at` field uses `omitempty` JSON tag, so:
- Old JSONL files without `last_read_at` will parse correctly
- New daemon reading old files will see `last_read_at` as 0 (zero value)
- Old daemon reading new files will ignore the unknown `last_read_at` field

No migration script required.

# Data Model: Stale Agent Notification Support

**Feature**: `008-stale-agent-mailman`
**Date**: 2026-01-13

## New Types

### StatelessTracker

**Purpose**: Track last notification timestamps for stateless agents to enforce 60-second notification intervals.

**Location**: `internal/daemon/loop.go`

```go
// StatelessTracker tracks notification timestamps for stateless agents.
// It uses in-memory storage that resets on daemon restart.
type StatelessTracker struct {
    mu             sync.Mutex          // Protects concurrent access
    lastNotified   map[string]time.Time // Window name → last notification time
    notifyInterval time.Duration       // Minimum interval between notifications
}
```

**Fields**:
| Field | Type | Description |
|-------|------|-------------|
| `mu` | `sync.Mutex` | Ensures thread-safe access to the map |
| `lastNotified` | `map[string]time.Time` | Maps window names to their last notification timestamp |
| `notifyInterval` | `time.Duration` | Configurable interval (default 60s) |

**Lifecycle**:
- Created once when daemon starts via `NewStatelessTracker()`
- Populated as stateless agents are discovered and notified
- Entries removed via `Cleanup()` when windows no longer have mailboxes
- Entire state discarded on daemon shutdown (acceptable per spec)

## Modified Types

### LoopOptions

**Location**: `internal/daemon/loop.go`

**Change**: Add `StatelessTracker` field

```go
type LoopOptions struct {
    RepoRoot         string             // Repository root path
    Interval         time.Duration      // Loop interval (default 10s)
    StopChan         chan struct{}      // Channel to stop the loop
    SkipTmuxCheck    bool               // Skip tmux check (for testing)
    StatelessTracker *StatelessTracker  // NEW: Tracker for stateless agents
}
```

## Constants

### New Constant

```go
// StatelessNotifyInterval is the interval between notifications for stateless agents.
const StatelessNotifyInterval = 60 * time.Second
```

## Method Contracts

### NewStatelessTracker

```go
// NewStatelessTracker creates a new tracker with the specified notification interval.
func NewStatelessTracker(interval time.Duration) *StatelessTracker
```

**Input**: `interval` - minimum time between notifications
**Output**: Initialized tracker with empty map
**Thread Safety**: N/A (initialization only)

### ShouldNotify

```go
// ShouldNotify returns true if the window is eligible for notification.
// Returns true if: (a) window not in tracker, or (b) interval elapsed since last notification.
func (t *StatelessTracker) ShouldNotify(window string) bool
```

**Input**: `window` - tmux window name
**Output**: `true` if notification should be sent
**Thread Safety**: Acquires mutex for map read

### MarkNotified

```go
// MarkNotified records that a notification was sent to the window.
func (t *StatelessTracker) MarkNotified(window string)
```

**Input**: `window` - tmux window name
**Effect**: Updates `lastNotified[window]` to current time
**Thread Safety**: Acquires mutex for map write

### Cleanup

```go
// Cleanup removes entries for windows that are no longer active.
func (t *StatelessTracker) Cleanup(activeWindows []string)
```

**Input**: `activeWindows` - list of currently active window names (from mailbox list)
**Effect**: Removes map entries not in `activeWindows`
**Thread Safety**: Acquires mutex for map modification

## State Diagram

```text
                    ┌─────────────────────────────────────┐
                    │         StatelessTracker            │
                    └─────────────────────────────────────┘
                                     │
                                     ▼
    ┌───────────────────────────────────────────────────────────────┐
    │                      Window Entry States                       │
    ├───────────────────────────────────────────────────────────────┤
    │                                                               │
    │   ┌─────────────┐    ShouldNotify()    ┌─────────────────┐   │
    │   │  Not Found  │ ─────────────────────►│  Should Notify  │   │
    │   │  (no entry) │       returns true    │  (eligible)     │   │
    │   └─────────────┘                       └────────┬────────┘   │
    │          ▲                                       │            │
    │          │                              MarkNotified()        │
    │          │                                       │            │
    │          │                                       ▼            │
    │          │     Cleanup()              ┌─────────────────┐     │
    │          └────────────────────────────│    Notified     │     │
    │           (window not in list)        │ (entry exists)  │     │
    │                                       └────────┬────────┘     │
    │                                                │              │
    │                                     After 60s elapsed         │
    │                                                │              │
    │                                                ▼              │
    │                                       ┌─────────────────┐     │
    │                                       │  Should Notify  │     │
    │                                       │  (eligible)     │     │
    │                                       └─────────────────┘     │
    │                                                               │
    └───────────────────────────────────────────────────────────────┘
```

## Existing Types (Reference)

These existing types are used but not modified:

### RecipientState (existing)
**Location**: `internal/mail/recipients.go`
```go
type RecipientState struct {
    Recipient string    `json:"recipient"`
    Status    string    `json:"status"`      // "ready", "work", "offline"
    UpdatedAt time.Time `json:"updated_at"`
    Notified  bool      `json:"notified"`
}
```

### Message (existing)
**Location**: `internal/mail/mailbox.go`
```go
type Message struct {
    ID       string `json:"id"`
    From     string `json:"from"`
    To       string `json:"to"`
    Message  string `json:"message"`
    ReadFlag bool   `json:"read_flag"`
}
```

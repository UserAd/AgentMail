# Implementation Plan: Stale Agent Notification Support

**Branch**: `008-stale-agent-mailman` | **Date**: 2026-01-13 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/008-stale-agent-mailman/spec.md`

## Summary

Add fallback notification support for stateless agents (those without Claude hooks) that cannot call `agentmail status`. The daemon will discover agents with mailboxes but no status entry and send periodic notifications at 60-second intervals until messages are read.

## Technical Context

**Language/Version**: Go 1.21+ (per IC-001)
**Primary Dependencies**: Standard library only (time, sync)
**Storage**: JSONL files in `.agentmail/` (existing), in-memory tracker (new)
**Testing**: `go test -v -race ./...` with 80% coverage gate
**Target Platform**: macOS and Linux with tmux installed
**Project Type**: Single CLI application
**Performance Goals**: 10-second loop cycle, 60-second stateless notification interval
**Constraints**: No external dependencies, in-memory tracking acceptable
**Scale/Scope**: Handles all tmux windows in current session

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | PASS | Extends daemon functionality (permitted per v1.1.0) |
| II. Simplicity (YAGNI) | PASS | Solves demonstrated need for non-hooks agents |
| III. Test Coverage | GATE | Must achieve 80% coverage on new code |
| IV. Standard Library | PASS | Uses only time, sync packages |

## Project Structure

### Documentation (this feature)

```text
specs/008-stale-agent-mailman/
├── spec.md              # Feature specification (complete)
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
internal/
├── daemon/
│   ├── loop.go          # MODIFY: Add StatelessTracker, Phase 2 notification
│   ├── loop_test.go     # MODIFY: Add stateless notification tests
│   └── daemon.go        # MODIFY: Initialize StatelessTracker in runForeground()
└── mail/
    ├── recipients.go    # EXISTING: ListMailboxRecipients() already exists
    └── mailbox.go       # EXISTING: FindUnread() already exists
```

**Structure Decision**: Single project layout. All changes are within existing `internal/daemon/` package with minor integration in existing files.

## Complexity Tracking

No constitution violations requiring justification.

---

## Phase 0: Research

### Decision Log

**D1: Tracking Approach**
- **Decision**: In-memory `map[string]time.Time` with mutex
- **Rationale**: Simple, sufficient for use case, daemon restart = fresh state is acceptable
- **Alternatives Rejected**:
  - Persistent file storage: Overkill for this use case, adds complexity
  - Adding to recipients.jsonl: Would conflate stated/stateless concepts

**D2: Discovery Method**
- **Decision**: Use existing `ListMailboxRecipients()` and `ReadAllRecipients()` to compute difference
- **Rationale**: Reuses existing well-tested code
- **Alternatives Rejected**:
  - New tmux scan: Would miss mailboxes for offline windows
  - New filesystem scan: Redundant with existing function

**D3: Notification Interval**
- **Decision**: 60 seconds (6x the stated agent loop)
- **Rationale**: Balances notification responsiveness vs notification spam
- **Alternatives Rejected**:
  - 30 seconds: Too frequent for stateless agents
  - 5 minutes: Too slow for urgent messages

**D4: Integration Point**
- **Decision**: Add Phase 2 inside existing `CheckAndNotifyWithNotifier()` function
- **Rationale**: Single notification cycle handles both stated and stateless
- **Alternatives Rejected**:
  - Separate loop: Would complicate timing and state management
  - New function called from RunLoop: Adds unnecessary indirection

---

## Phase 1: Design

### Data Model

**New Type: StatelessTracker** (in `internal/daemon/loop.go`)

```go
// StatelessTracker tracks notification timestamps for stateless agents.
// It uses in-memory storage that resets on daemon restart.
type StatelessTracker struct {
    mu             sync.Mutex
    lastNotified   map[string]time.Time
    notifyInterval time.Duration
}
```

**Methods:**
- `NewStatelessTracker(interval time.Duration) *StatelessTracker` - Constructor
- `ShouldNotify(window string) bool` - Returns true if window is due for notification
- `MarkNotified(window string)` - Records notification timestamp
- `Cleanup(activeWindows []string)` - Removes entries for non-existent windows

**Modified Type: LoopOptions** (in `internal/daemon/loop.go`)

Add field to pass tracker to the notification function:
```go
type LoopOptions struct {
    RepoRoot         string
    Interval         time.Duration
    StopChan         chan struct{}
    SkipTmuxCheck    bool
    StatelessTracker *StatelessTracker  // NEW: tracker for stateless agents
}
```

### Modified Flow: CheckAndNotifyWithNotifier

```text
CheckAndNotifyWithNotifier()
├── Phase 1: Stated agents (existing, unchanged)
│   ├── Read recipients from recipients.jsonl
│   ├── Filter: Status=ready, Notified=false, has unread
│   └── Notify and set Notified=true
│
└── Phase 2: Stateless agents (NEW)
    ├── Build statedSet from Phase 1 recipients
    ├── List all mailbox recipients via ListMailboxRecipients()
    ├── For each mailbox recipient NOT in statedSet:
    │   ├── Check if has unread messages
    │   ├── Check if tracker.ShouldNotify() returns true
    │   ├── If both: notify and tracker.MarkNotified()
    └── tracker.Cleanup() with current mailbox list
```

### API Contract

No new CLI commands or external APIs. Internal function signatures:

```go
// New functions
func NewStatelessTracker(interval time.Duration) *StatelessTracker
func (t *StatelessTracker) ShouldNotify(window string) bool
func (t *StatelessTracker) MarkNotified(window string)
func (t *StatelessTracker) Cleanup(activeWindows []string)

// Modified function signature (opts gains StatelessTracker field)
func CheckAndNotifyWithNotifier(opts LoopOptions, notify NotifyFunc) error
```

---

## Implementation Tasks

### Task 1: Add StatelessTracker type and methods
**File**: `internal/daemon/loop.go`
**Effort**: Small
**Implements**: FR-009, FR-010, FR-013

Add constant and type:
```go
const StatelessNotifyInterval = 60 * time.Second

type StatelessTracker struct { ... }
```

Implement 4 methods:
- `NewStatelessTracker()` - Initialize map and interval (FR-010)
- `ShouldNotify()` - Check if >= interval since last notification (FR-005)
- `MarkNotified()` - Store current time (FR-009)
- `Cleanup()` - Remove entries not in activeWindows list (FR-011)

### Task 2: Modify LoopOptions and CheckAndNotifyWithNotifier
**File**: `internal/daemon/loop.go`
**Effort**: Medium
**Implements**: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, FR-008, FR-012

1. Add `StatelessTracker` field to `LoopOptions`
2. After existing Phase 1 logic in `CheckAndNotifyWithNotifier()`:
   - Build set of stated recipient names (FR-002, FR-007)
   - Call `mail.ListMailboxRecipients()` (FR-001)
   - For each mailbox recipient NOT in stated set (FR-003):
     - Check `mail.FindUnread()` returns > 0 messages (FR-006)
     - Check `opts.StatelessTracker.ShouldNotify()` (FR-005)
     - If both true and notify != nil, call notify() (FR-004)
     - Call `opts.StatelessTracker.MarkNotified()`
   - Call `opts.StatelessTracker.Cleanup()` with mailbox list (FR-011)

### Task 3: Add error handling
**File**: `internal/daemon/loop.go`
**Effort**: Small
**Implements**: FR-014, FR-015, FR-016, FR-017

Add error handling in Phase 2 stateless logic:
- Log and continue if `ListMailboxRecipients()` fails (FR-014)
- Skip marking notified if `notify()` fails (FR-015)
- Skip agent if `FindUnread()` fails, log warning (FR-016)
- If `ReadAllRecipients()` fails, treat all as stateless (FR-017)

### Task 4: Integrate tracker in daemon.go
**File**: `internal/daemon/daemon.go`
**Effort**: Small
**Implements**: FR-010, FR-012

In `runForeground()`, before starting the loop goroutine:
```go
tracker := NewStatelessTracker(StatelessNotifyInterval)
opts := LoopOptions{
    RepoRoot:         repoRoot,
    Interval:         DefaultLoopInterval,
    StopChan:         loopStopChan,
    SkipTmuxCheck:    false,
    StatelessTracker: tracker,
}
```

### Task 5: Add unit tests for StatelessTracker
**File**: `internal/daemon/loop_test.go`
**Effort**: Medium
**Validates**: FR-005, FR-009, FR-010, FR-011, FR-013, SC-007

Test cases:
- `TestStatelessTracker_ShouldNotify_FirstTime` - Returns true for new window (FR-010)
- `TestStatelessTracker_ShouldNotify_BeforeInterval` - Returns false before 60s (FR-005)
- `TestStatelessTracker_ShouldNotify_AfterInterval` - Returns true after 60s (FR-005)
- `TestStatelessTracker_MarkNotified` - Updates timestamp (FR-009)
- `TestStatelessTracker_Cleanup` - Removes stale entries (FR-011)
- `TestStatelessTracker_ThreadSafety` - Concurrent access is safe (FR-013, SC-007)

### Task 6: Add integration tests for stateless notification
**File**: `internal/daemon/loop_test.go`
**Effort**: Medium
**Validates**: FR-001 through FR-008, SC-001 through SC-005

Test cases:
- `TestStatelessNotification_AgentWithMailboxNoState` - Notifies stateless agent (FR-003, FR-004)
- `TestStatelessNotification_StatedAgentTakesPrecedence` - Skips agents with state (FR-007, SC-003)
- `TestStatelessNotification_RespectInterval` - Doesn't re-notify before 60s (FR-005, SC-002)
- `TestStatelessNotification_NoUnreadMessages` - Skips if mailbox empty (FR-006)
- `TestStatelessNotification_MultipleAgents` - Handles multiple stateless agents
- `TestStatelessNotification_TransitionToStated` - Ceases stateless on status register (FR-008)

### Task 7: Add error handling tests
**File**: `internal/daemon/loop_test.go`
**Effort**: Small
**Validates**: FR-014, FR-015, FR-016, FR-017

Test cases:
- `TestStatelessNotification_MailboxDirReadError` - Continues on directory error (FR-014)
- `TestStatelessNotification_NotifyFailure` - Retries on next interval (FR-015)
- `TestStatelessNotification_MailboxFileReadError` - Skips agent (FR-016)
- `TestStatelessNotification_RecipientsReadError` - Falls back to stateless (FR-017)

---

## Verification

### Automated Testing
```bash
# Run all tests with race detection
go test -v -race ./internal/daemon/...

# Check coverage meets 80% gate
go test -cover ./internal/daemon/...

# Run full quality gate
go vet ./... && go fmt ./... && go test -v -race ./...
```

### Manual Testing
1. Start mailman daemon: `agentmail mailman start`
2. Open a new tmux window (without hooks): `tmux new-window -n test-agent`
3. Send message: `agentmail send test-agent "Hello stateless agent"`
4. Verify notification arrives in test-agent window within 60 seconds
5. Wait 60 seconds, verify notification repeats
6. In test-agent window, run: `agentmail receive`
7. Wait 60 seconds, verify no more notifications arrive

### Expected Test Count
- 6 unit tests for StatelessTracker (including thread-safety)
- 6 integration tests for stateless notification flow
- 4 error handling tests
- **Total new tests**: 16
- All existing tests continue to pass (stated agent behavior unchanged)

### Requirements Coverage Matrix

| Requirement | Test(s) |
|-------------|---------|
| FR-001 | TestStatelessNotification_AgentWithMailboxNoState |
| FR-002 | TestStatelessNotification_AgentWithMailboxNoState |
| FR-003 | TestStatelessNotification_AgentWithMailboxNoState |
| FR-004 | TestStatelessNotification_AgentWithMailboxNoState |
| FR-005 | TestStatelessTracker_ShouldNotify_*, TestStatelessNotification_RespectInterval |
| FR-006 | TestStatelessNotification_NoUnreadMessages |
| FR-007 | TestStatelessNotification_StatedAgentTakesPrecedence |
| FR-008 | TestStatelessNotification_TransitionToStated |
| FR-009 | TestStatelessTracker_MarkNotified |
| FR-010 | TestStatelessTracker_ShouldNotify_FirstTime |
| FR-011 | TestStatelessTracker_Cleanup |
| FR-012 | (integration with existing loop tests) |
| FR-013 | TestStatelessTracker_ThreadSafety |
| FR-014 | TestStatelessNotification_MailboxDirReadError |
| FR-015 | TestStatelessNotification_NotifyFailure |
| FR-016 | TestStatelessNotification_MailboxFileReadError |
| FR-017 | TestStatelessNotification_RecipientsReadError |

# Data Model: Mailman Daemon

**Feature**: 006-mailman-daemon
**Date**: 2026-01-12

## Entities

### 1. RecipientState

Tracks the availability status of an agent for notification purposes.

**File**: `.git/mail-recipients.jsonl`
**Format**: JSONL (one JSON object per line)

| Field | Type | Description | Constraints |
|-------|------|-------------|-------------|
| `recipient` | string | Tmux window name (agent identifier) | Required, unique per file |
| `status` | string | Current availability state | Enum: `ready`, `work`, `offline` |
| `updated_at` | string | ISO 8601 timestamp of last update | Required, RFC3339 format |
| `notified` | bool | Whether agent has been notified since last state transition | Default: false |

**Example**:
```json
{"recipient":"agent1","status":"ready","updated_at":"2026-01-12T10:00:00Z","notified":false}
{"recipient":"agent2","status":"work","updated_at":"2026-01-12T10:01:30Z","notified":true}
{"recipient":"agent3","status":"offline","updated_at":"2026-01-12T09:55:00Z","notified":false}
```

**Go Struct**:
```go
type RecipientState struct {
    Recipient string    `json:"recipient"`
    Status    string    `json:"status"`
    UpdatedAt time.Time `json:"updated_at"`
    Notified  bool      `json:"notified"`
}
```

### 2. PID File

Simple text file containing the daemon's process ID.

**File**: `.git/mail/mailman.pid`
**Format**: Plain text, single line with integer PID

| Content | Type | Description |
|---------|------|-------------|
| PID | int | Process ID of running mailman daemon |

**Example**:
```
12345
```

## State Transitions

### Agent Status State Machine

```
                    ┌─────────────────────────────────────┐
                    │                                     │
                    ▼                                     │
    ┌─────────┐  status ready  ┌─────────┐  status work  │
    │ (none)  │ ────────────▶ │  ready  │ ────────────▶ │
    └─────────┘               └─────────┘               │
                                   │                     │
                                   │ status offline      │
                                   ▼                     │
                              ┌─────────┐               │
                              │ offline │ ──────────────┘
                              └─────────┘   status ready
                                   │
                                   │ status work
                                   ▼
                              ┌─────────┐
                              │  work   │
                              └─────────┘
                                   │
                                   │ status ready/offline
                                   ▼
                              (back to ready/offline)
```

**Transition Rules**:
- Any status → Any status (no restrictions)
- `ready` → `work`: `notified` flag preserved
- `ready` → `offline`: `notified` flag reset to false
- `work` → `ready`: `notified` flag reset to false
- `work` → `offline`: `notified` flag reset to false
- `offline` → `ready`: `notified` flag reset to false
- `offline` → `work`: `notified` flag stays false

**Notification Flag Reset**:
- Reset to `false` when transitioning TO `work` or `offline`
- Set to `true` when notification is sent (from `ready` state only)

### Daemon Lifecycle

```
    ┌─────────┐
    │  start  │
    └────┬────┘
         │
         ▼
    ┌─────────────┐     PID exists & running
    │ check PID   │ ──────────────────────────▶ ERROR (exit 2)
    └──────┬──────┘
           │ no PID or stale
           ▼
    ┌─────────────┐
    │ write PID   │
    └──────┬──────┘
           │
           ▼
    ┌─────────────┐
    │ clean stale │ (>1hr old states)
    └──────┬──────┘
           │
           ▼
    ┌─────────────┐     SIGTERM/SIGINT
    │ run loop    │ ──────────────────────────▶ cleanup PID, exit 0
    └─────────────┘
```

## Relationships

```
┌─────────────────────────────────────────────────────────────┐
│                     .git/mail/                              │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  mailman.pid          mail-recipients.jsonl                 │
│  ┌──────────┐         ┌─────────────────────┐              │
│  │  12345   │         │ recipient: agent1   │              │
│  └──────────┘         │ status: ready       │──────┐       │
│                       │ notified: false     │      │       │
│                       ├─────────────────────┤      │       │
│                       │ recipient: agent2   │      │       │
│                       │ status: work        │      │       │
│                       │ notified: true      │      │       │
│                       └─────────────────────┘      │       │
│                                                    │       │
│  agent1.jsonl ◀────────────────────────────────────┘       │
│  ┌─────────────────────┐                                   │
│  │ Message (existing)  │                                   │
│  │ - id, from, to      │                                   │
│  │ - body, timestamp   │                                   │
│  │ - read_flag         │                                   │
│  └─────────────────────┘                                   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Validation Rules

### RecipientState

1. `recipient` must be non-empty string
2. `status` must be one of: `ready`, `work`, `offline`
3. `updated_at` must be valid RFC3339 timestamp
4. File must be valid JSONL (one JSON object per line)

### PID File

1. Content must be valid integer
2. Integer must be positive
3. File must contain single line (no trailing newline required)

## Stale Data Cleanup

**Rule**: On daemon startup, remove entries from `mail-recipients.jsonl` where:
- `updated_at` is older than 1 hour from current time

**Rationale**: Prevents notification to agents that have crashed or disconnected without proper offline status.

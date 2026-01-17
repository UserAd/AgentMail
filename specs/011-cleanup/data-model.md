# Data Model: Cleanup Command

**Feature**: 011-cleanup
**Date**: 2026-01-15

## Entity Changes

### Message (Modified)

**File**: `internal/mail/message.go`

```go
// Message represents a communication between agents.
type Message struct {
    ID        string    `json:"id"`                   // Short unique identifier (8 chars, base62)
    From      string    `json:"from"`                 // Sender tmux window name
    To        string    `json:"to"`                   // Recipient tmux window name
    Message   string    `json:"message"`              // Body text
    ReadFlag  bool      `json:"read_flag"`            // Read status (default: false)
    CreatedAt time.Time `json:"created_at,omitempty"` // NEW: Timestamp for age-based cleanup
}
```

**Field Details**:

| Field | Type | JSON Key | Required | Description |
|-------|------|----------|----------|-------------|
| CreatedAt | time.Time | created_at | No (omitempty) | RFC 3339 timestamp set when message created |

**Backward Compatibility**:
- `omitempty` ensures existing messages without timestamp serialize correctly
- Messages without `created_at` are skipped during age-based cleanup (not deleted)

### RecipientState (Unchanged)

**File**: `internal/mail/recipients.go`

Existing structure supports cleanup requirements:
- `Recipient` - name for offline check against tmux windows
- `UpdatedAt` - timestamp for staleness check

```go
// RecipientState represents the availability state of a recipient agent
type RecipientState struct {
    Recipient  string    `json:"recipient"`
    Status     string    `json:"status"`
    UpdatedAt  time.Time `json:"updated_at"`
    NotifiedAt time.Time `json:"notified_at,omitempty"`
    LastReadAt int64     `json:"last_read_at,omitempty"`
}
```

## New Types

### CleanupResult

**File**: `internal/cli/cleanup.go`

```go
// CleanupResult holds the counts from a cleanup operation
type CleanupResult struct {
    RecipientsRemoved int // Total recipients removed
    OfflineRemoved    int // Recipients removed because window doesn't exist
    StaleRemoved      int // Recipients removed because updated_at expired
    MessagesRemoved   int // Messages removed (read + old)
    MailboxesRemoved  int // Empty mailbox files removed
    FilesSkipped      int // Files skipped due to lock contention
}
```

### CleanupOptions

**File**: `internal/cli/cleanup.go`

```go
// CleanupOptions configures the Cleanup command behavior
type CleanupOptions struct {
    StaleHours     int  // Hours threshold for stale recipients (default: 48)
    DeliveredHours int  // Hours threshold for delivered messages (default: 2)
    DryRun         bool // If true, report what would be cleaned without deleting
}
```

## Storage Schema

### recipients.jsonl

**Path**: `.agentmail/recipients.jsonl`
**Format**: JSONL (one JSON object per line)

No schema changes. Existing fields sufficient for cleanup:

```jsonl
{"recipient":"agent1","status":"ready","updated_at":"2026-01-15T10:00:00Z","notified_at":"2026-01-15T10:30:00Z","last_read_at":1736940000000}
{"recipient":"agent2","status":"offline","updated_at":"2026-01-13T10:00:00Z"}
```

### Mailbox Files

**Path**: `.agentmail/mailboxes/<recipient>.jsonl`
**Format**: JSONL (one Message JSON object per line)

**Before** (existing):
```jsonl
{"id":"ABC12345","from":"agent1","to":"agent2","message":"Hello","read_flag":false}
```

**After** (with new field):
```jsonl
{"id":"ABC12345","from":"agent1","to":"agent2","message":"Hello","read_flag":false,"created_at":"2026-01-15T12:00:00Z"}
```

## Validation Rules

### Message.CreatedAt

- Must be valid RFC 3339 timestamp when present
- Zero value (missing) is valid - indicates legacy message
- Future timestamps are not validated (clock skew tolerance)

### Cleanup Thresholds

- StaleHours: Must be >= 0 (0 means remove all recipients)
- DeliveredHours: Must be >= 0 (0 means remove all read messages)
- Negative values should be rejected with error

## State Transitions

### Recipient Lifecycle (Cleanup-relevant)

```
[Active in tmux] --window closed--> [Offline - eligible for cleanup]
[Any status] --updated_at expires--> [Stale - eligible for cleanup]
```

### Message Lifecycle (Cleanup-relevant)

```
[Unread] --read by recipient--> [Read]
[Read] --age > threshold--> [Eligible for cleanup]
[Read] --cleanup runs--> [Deleted]
```

Note: Unread messages are NEVER deleted by cleanup regardless of age.

### Mailbox File Lifecycle

```
[Contains messages] --all messages removed--> [Empty]
[Empty] --cleanup runs--> [Deleted]
```

## Relationships

```
RecipientState 1:1 Mailbox File (by recipient name)
Mailbox File 1:N Message (stored in JSONL)
Recipient --> tmux Window (validated by cleanup)
```

## Migration Notes

No data migration required:
- New `created_at` field uses `omitempty` for backward compatibility
- Existing messages without timestamp are skipped (not deleted)
- New messages will automatically include `created_at` (set in Append function)

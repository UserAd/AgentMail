# Data Model: AgentMail

**Feature**: 001-agent-mail-structure
**Date**: 2026-01-11

## Entities

### Message

The core entity representing a communication between agents.

```go
type Message struct {
    ID       string `json:"id"`        // Short unique identifier (8 chars, base62)
    From     string `json:"from"`      // Sender tmux window name
    To       string `json:"to"`        // Recipient tmux window name
    Message  string `json:"message"`   // Body text
    ReadFlag bool   `json:"read_flag"` // Read status (default: false)
}
```

**Field Specifications**:

| Field | Type | Constraints | Source Requirement |
|-------|------|-------------|-------------------|
| `id` | string | 8 chars, alphanumeric (base62) | FR-010a |
| `from` | string | Non-empty, valid tmux window name | FR-009 |
| `to` | string | Non-empty, valid tmux window name | FR-009 |
| `message` | string | Non-empty text | FR-009 |
| `read_flag` | bool | Default: false | FR-009 |

**Validation Rules**:
- `id`: Generated automatically, 8 characters from `[a-zA-Z0-9]` (FR-010a)
- `from`: Must match current tmux window name at send time (FR-004)
- `to`: Must exist in `tmux list-windows` at send time (FR-001a)
- `message`: Must be non-empty string
- `read_flag`: Set to `false` on creation, `true` after receive (FR-002b)

### Agent (Implicit)

An agent is not stored but derived from the tmux environment.

```go
// Agent identity is determined at runtime
func GetCurrentAgent() (string, error) {
    // Executes: tmux display-message -p '#W'
    // Returns: window name or error if not in tmux
}
```

**Identity Resolution**:
- Determined by tmux window name (FR-004)
- Not in tmux → error with exit code 2 (FR-005a, FR-005b)

### Mailbox (Storage)

Each recipient has their own JSONL mailbox file.

**Location**: `.git/mail/<recipient>.jsonl`

**Examples**:
- `.git/mail/agent-1.jsonl` - Messages for agent-1
- `.git/mail/agent-2.jsonl` - Messages for agent-2

**Format**: JSONL (one JSON object per line)

```jsonl
# .git/mail/agent-2.jsonl (messages TO agent-2)
{"id":"xK7mN2pQ","from":"agent-1","to":"agent-2","message":"Hello","read_flag":false}
{"id":"zM9oP4rS","from":"agent-3","to":"agent-2","message":"Meeting?","read_flag":false}
```

**Benefits of Per-Recipient Files**:
- No filtering needed when reading (all messages in file are for recipient)
- Reduced file locking contention (agents write to different files)
- Smaller files, faster reads
- Better concurrent access (different agents don't block each other)

**Operations**:

| Operation | Description | Locking |
|-----------|-------------|---------|
| Append | Add new message to recipient's file | Write lock on recipient file |
| Query | Read all messages from recipient's file | No lock |
| Update | Modify read_flag in recipient's file | Write lock on recipient file |

**State Transitions**:

```text
Message Created (send)    Message Received (receive)
       │                          │
       ▼                          ▼
  read_flag=false  ──────►  read_flag=true
```

## Storage Format

### File Structure

```text
.git/
└── mail/
    ├── agent-1.jsonl    # Messages for agent-1
    ├── agent-2.jsonl    # Messages for agent-2
    └── agent-3.jsonl    # Messages for agent-3
```

**File Naming**: `<recipient-window-name>.jsonl`

**Why `.git/mail/`**:
- Located in git directory but not tracked (add to `.gitignore`)
- Persists with the repository
- Clear ownership per repository
- Easy to inspect/debug

**Why Per-Recipient Files**:
- Send operation only locks recipient's file (not global)
- Receive operation reads only own file (no filtering)
- Concurrent sends to different recipients don't block

### JSONL Specification

- One complete JSON object per line
- No trailing commas
- UTF-8 encoded
- Newline-terminated (each line ends with `\n`)
- Messages appended in chronological order

### Directory Creation

The `.git/mail/` directory is created automatically on first use if it doesn't exist (edge case from spec).

## Relationships

```text
┌─────────────┐         ┌─────────────┐
│   Agent     │ sends   │   Message   │
│ (tmux win)  │────────►│             │
└─────────────┘         └─────────────┘
       │                       │
       │ receives              │ stored in
       │                       ▼
       │                ┌─────────────┐
       └───────────────►│   Mailbox   │
                        │  (JSONL)    │
                        └─────────────┘
```

## Queries

### Find Unread Messages for Agent

```go
// Pseudocode - simplified with per-recipient files
func FindUnreadForAgent(recipient string) []Message {
    // Read only the recipient's file
    messages := readMessagesFromFile(recipient + ".jsonl")
    var unread []Message
    for _, m := range messages {
        if !m.ReadFlag {
            unread = append(unread, m)
        }
    }
    // Already in FIFO order (append-only)
    return unread
}
```

### Mark Message as Read

```go
// Pseudocode - operates on recipient's file only
func MarkAsRead(recipient, messageID string) error {
    messages := readMessagesFromFile(recipient + ".jsonl")
    for i := range messages {
        if messages[i].ID == messageID {
            messages[i].ReadFlag = true
            break
        }
    }
    return writeMessagesToFile(recipient + ".jsonl", messages)
}
```

### Append Message (Send)

```go
// Pseudocode - append to recipient's file
func AppendMessage(msg Message) error {
    filepath := msg.To + ".jsonl"
    // Lock recipient's file, append, unlock
    return appendToFile(filepath, msg)
}
```

## Concurrency Considerations

- Multiple agents may send/receive concurrently
- Per-recipient files improve concurrency:
  - Sends to different recipients don't block each other
  - Each agent's receive only locks their own file
- File locking via `flock` protects writes to each recipient file
- Read-modify-write for marking messages requires lock on recipient's file
- Append-only for new messages is naturally safe with locking

# Data Model: MCP Server for AgentMail

**Feature**: 010-mcp-server
**Date**: 2026-01-14

## Overview

The MCP server feature does not introduce new persistent data entities. It provides an alternative interface (MCP protocol over STDIO) to existing AgentMail functionality. All data operations use the existing mail infrastructure.

## Existing Entities (Reused)

### Message
Defined in `internal/mail/message.go`. Used by `send` and `receive` tools.

```go
type Message struct {
    ID       string `json:"id"`        // Unique message identifier
    From     string `json:"from"`      // Sender tmux window name
    To       string `json:"to"`        // Recipient tmux window name
    Message  string `json:"message"`   // Message content
    ReadFlag bool   `json:"read_flag"` // Whether message has been read
}
```

### Recipient State
Defined in `internal/mail/recipients.go`. Used by `status` and `list-recipients` tools.

```go
type RecipientEntry struct {
    Name       string `json:"name"`                  // Tmux window name
    Status     string `json:"status"`                // ready|work|offline
    Notified   bool   `json:"notified"`              // Notification flag
    LastReadAt int64  `json:"last_read_at,omitempty"` // Unix timestamp (ms)
}
```

## MCP-Specific Types (Runtime Only)

These types exist only in memory during MCP server operation.

### Tool Invocation Request

```go
// Handled by MCP SDK - these are the parameter shapes

// SendParams - parameters for send tool
type SendParams struct {
    Recipient string `json:"recipient"` // Required: target window name
    Message   string `json:"message"`   // Required: message content (max 64KB)
}

// StatusParams - parameters for status tool (sets agent availability)
type StatusParams struct {
    Status string `json:"status"` // Required: "ready", "work", or "offline"
}

// ReceiveParams - no parameters
// ListRecipientsParams - no parameters
```

### Tool Response Formats

```go
// SendResponse - returned by send tool
type SendResponse struct {
    MessageID string `json:"message_id"` // Generated ID, e.g., "ABC123"
}

// ReceiveResponse - returned by receive tool (message available)
type ReceiveResponse struct {
    From    string `json:"from"`    // Sender window name
    ID      string `json:"id"`      // Message ID
    Message string `json:"message"` // Message content
}

// ReceiveEmptyResponse - returned when no messages
type ReceiveEmptyResponse struct {
    Status string `json:"status"` // "No unread messages"
}

// StatusResponse - returned by status tool after setting availability
type StatusResponse struct {
    Status string `json:"status"` // "ok" on success, confirms status was updated
}

// ListRecipientsResponse - returned by list-recipients tool
type ListRecipientsResponse struct {
    Recipients []RecipientInfo `json:"recipients"`
}

type RecipientInfo struct {
    Name      string `json:"name"`       // Window name
    IsCurrent bool   `json:"is_current"` // True if this is the caller's window
}
```

## Storage Locations (Existing)

| File | Purpose | Used By |
|------|---------|---------|
| `.agentmail/mailboxes/<name>.jsonl` | Message storage per recipient | send, receive |
| `.agentmail/recipients.jsonl` | Recipient state registry | status, list-recipients |
| `.agentmailignore` | Windows to exclude from recipients | list-recipients, send |

## Validation Rules

### Message Content (FR-013)
- Maximum size: 64 KB (65,536 bytes)
- Validated before processing send tool

### Status Values (Existing)
- Must be one of: `ready`, `work`, `offline`
- Case-sensitive

### Recipient Names
- Must be valid tmux window name in current session
- Must not be in `.agentmailignore` (for send)
- Must not be sender (no self-send)

## State Transitions

### Message Lifecycle

```text
[Created] --send--> [Unread] --receive--> [Read]
```

- Created: Message written to recipient's mailbox file
- Unread: `read_flag: false`
- Read: `read_flag: true` after receive tool returns message

### Recipient Status Lifecycle

```text
[Any] --status ready--> [Ready]
[Any] --status work--> [Work] (notified flag reset)
[Any] --status offline--> [Offline] (notified flag reset)
```

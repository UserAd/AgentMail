# Research: Stale Agent Notification Support

**Feature**: `008-stale-agent-mailman`
**Date**: 2026-01-13

## Overview

Research findings for implementing fallback notification support for stateless agents in the mailman daemon.

## Decisions

### D1: Tracking Approach

**Decision**: In-memory `map[string]time.Time` with mutex protection

**Rationale**:
- Simple implementation using Go standard library (`sync.Mutex`, `time.Time`)
- No persistence needed - daemon restart allows immediate re-notification (acceptable behavior per spec)
- Mutex provides safe concurrent access from the notification loop
- Memory footprint is negligible (one timestamp per stateless agent)

**Alternatives Rejected**:
| Alternative | Why Rejected |
|-------------|--------------|
| Persistent file storage | Adds complexity for no benefit - restarting daemon can re-notify immediately |
| Adding field to recipients.jsonl | Would conflate stated/stateless concepts; stateless agents shouldn't have state entries |
| Channel-based tracking | Overly complex for simple timestamp comparison |

### D2: Discovery Method

**Decision**: Compute difference between `ListMailboxRecipients()` and `ReadAllRecipients()`

**Rationale**:
- `ListMailboxRecipients()` already exists and returns all agents with mailboxes
- `ReadAllRecipients()` already exists and returns all stated agents
- Set difference (mailbox - stated) = stateless agents with mailboxes
- Reuses well-tested existing code

**Alternatives Rejected**:
| Alternative | Why Rejected |
|-------------|--------------|
| Scan tmux windows directly | Would miss mailboxes for windows that aren't currently running |
| New filesystem scan function | Redundant with existing `ListMailboxRecipients()` |
| Cache discovery results | Adds complexity; discovery is fast enough for 10-second loop |

### D3: Notification Interval

**Decision**: 60 seconds (6x the base loop interval)

**Rationale**:
- Base loop runs every 10 seconds
- 60-second interval provides 6 checks before notification
- Balances responsiveness (1 minute max delay) vs. spam prevention
- Aligns with user expectations from the original plan document

**Alternatives Rejected**:
| Alternative | Why Rejected |
|-------------|--------------|
| 30 seconds | Too frequent - would annoy stateless agents |
| 2 minutes | Acceptable but slightly slower than desired |
| 5 minutes | Too slow for potentially urgent messages |
| Same as stated (10s) | Would spam stateless agents every loop |

### D4: Integration Point

**Decision**: Add Phase 2 logic inside existing `CheckAndNotifyWithNotifier()` function

**Rationale**:
- Single function handles both stated and stateless agents
- Stated logic runs first, stateless second (clean precedence)
- Reuses same `NotifyFunc` callback for notifications
- Reuses same `LoopOptions` struct (with new tracker field)

**Alternatives Rejected**:
| Alternative | Why Rejected |
|-------------|--------------|
| Separate goroutine/loop | Would complicate timing and state management |
| New `CheckStatelessAgents()` function | Adds unnecessary indirection |
| RunLoop calls both functions | Two separate cycles less clean than single function |

## Existing Code Analysis

### Reusable Functions

| Function | Location | Purpose |
|----------|----------|---------|
| `mail.ListMailboxRecipients()` | `internal/mail/recipients.go:200` | List all agents with mailboxes |
| `mail.ReadAllRecipients()` | `internal/mail/recipients.go` | List all stated agents |
| `mail.FindUnread()` | `internal/mail/mailbox.go` | Check for unread messages |
| `NotifyAgent()` | `internal/daemon/loop.go:34` | Send tmux notification |

### Integration Points

| File | Function | Change Needed |
|------|----------|---------------|
| `internal/daemon/loop.go` | `CheckAndNotifyWithNotifier()` | Add Phase 2 stateless logic |
| `internal/daemon/loop.go` | `LoopOptions` | Add `StatelessTracker` field |
| `internal/daemon/daemon.go` | `runForeground()` | Initialize tracker |

## Dependencies

**Standard Library Only** (constitution-compliant):
- `sync` - Mutex for tracker thread safety
- `time` - Time tracking and interval comparison

No external dependencies required.

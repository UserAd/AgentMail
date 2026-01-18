# Feature Specification: Cleanup Command

**Feature Branch**: `011-cleanup`
**Created**: 2026-01-15
**Status**: Draft
**Input**: User description: "I want to have ability to cleanup old recipients/messages/mailboxes."

## Clarifications

### Session 2026-01-15

- Q: Messages currently lack a timestamp field. What should the new timestamp field be named? → A: `created_at` (matches existing `updated_at` naming convention in RecipientState)
- Q: How should the system behave when multiple cleanup processes run simultaneously? → A: Skip locked files (already specified for locked files; applies same pattern - skip and warn, no blocking)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Remove Offline Recipients (Priority: P1)

A system administrator or automated process wants to clean up recipients from the recipients registry who are no longer present as tmux windows in the current session.

**Why this priority**: Removing stale recipients that no longer exist prevents notifications from being sent to non-existent windows and keeps the registry accurate.

**Independent Test**: Can be fully tested by creating recipient entries in recipients.jsonl for windows that don't exist, running cleanup, and verifying those entries are removed.

**Acceptance Scenarios**:

1. **Given** recipients.jsonl contains "agent1" and "agent2" entries, **When** only "agent1" exists as a tmux window and user runs `agentmail cleanup`, **Then** "agent2" is removed from recipients.jsonl and "agent1" is retained.
2. **Given** recipients.jsonl contains entries for windows that all exist, **When** user runs `agentmail cleanup`, **Then** no entries are removed.
3. **Given** recipients.jsonl is empty or doesn't exist, **When** user runs `agentmail cleanup`, **Then** command completes successfully with no errors.

---

### User Story 2 - Remove Stale Recipients by Inactivity (Priority: P1)

A system administrator wants to remove recipients who haven't updated their status for an extended period (configurable, default 48 hours), indicating they are likely abandoned sessions.

**Why this priority**: Long-inactive recipients clutter the registry and may represent crashed or forgotten sessions. Equal priority with Story 1 as both address registry hygiene.

**Independent Test**: Can be fully tested by creating recipient entries with old `updated_at` timestamps, running cleanup with the stale threshold, and verifying old entries are removed.

**Acceptance Scenarios**:

1. **Given** recipients.jsonl contains an entry with `updated_at` older than 48 hours, **When** user runs `agentmail cleanup`, **Then** that entry is removed.
2. **Given** recipients.jsonl contains an entry with `updated_at` within the last 48 hours, **When** user runs `agentmail cleanup`, **Then** that entry is retained.
3. **Given** user specifies `--stale-hours 24`, **When** an entry has `updated_at` older than 24 hours, **Then** that entry is removed.

---

### User Story 3 - Remove Old Delivered Messages (Priority: P2)

A user wants to clean up messages that have been read/delivered and are older than a configurable threshold (default 2 hours) to prevent mailbox files from growing indefinitely.

**Why this priority**: Keeps storage under control while preserving unread messages. Lower than registry cleanup because message accumulation is less disruptive than stale registry entries.

**Independent Test**: Can be fully tested by creating mailbox files with mix of old read and unread messages, running cleanup, and verifying only old read messages are removed.

**Acceptance Scenarios**:

1. **Given** a mailbox contains a message with `read_flag: true` sent more than 2 hours ago, **When** user runs `agentmail cleanup`, **Then** that message is removed.
2. **Given** a mailbox contains a message with `read_flag: false` sent more than 2 hours ago, **When** user runs `agentmail cleanup`, **Then** that message is retained (unread messages are never deleted by age).
3. **Given** a mailbox contains a message with `read_flag: true` sent within the last 2 hours, **When** user runs `agentmail cleanup`, **Then** that message is retained.
4. **Given** user specifies `--delivered-hours 1`, **When** a read message is older than 1 hour, **Then** that message is removed.

---

### User Story 4 - Remove Empty Mailboxes (Priority: P3)

A user wants to remove empty mailbox files to keep the mailboxes directory tidy.

**Why this priority**: Cosmetic cleanup with no functional impact. Empty files don't cause issues but removing them keeps the system clean.

**Independent Test**: Can be fully tested by creating empty .jsonl files in the mailboxes directory, running cleanup, and verifying they are deleted.

**Acceptance Scenarios**:

1. **Given** a mailbox file exists with zero messages (empty or only whitespace), **When** user runs `agentmail cleanup`, **Then** that file is deleted.
2. **Given** a mailbox file contains at least one message, **When** user runs `agentmail cleanup`, **Then** that file is retained.

---

### User Story 5 - Cleanup Not Shown in Onboarding (Priority: P3)

The cleanup command should not appear in the onboarding output since it's an administrative function not needed for day-to-day agent communication.

**Why this priority**: User experience detail that doesn't affect core functionality.

**Independent Test**: Can be fully tested by running `agentmail onboard` and verifying the output does not mention "cleanup".

**Acceptance Scenarios**:

1. **Given** a user runs `agentmail onboard`, **When** the output is displayed, **Then** the word "cleanup" does not appear anywhere in the output.

---

### User Story 6 - Cleanup Not Exposed via MCP (Priority: P3)

The cleanup command should not be available as an MCP tool since it's an administrative function intended for human operators or scheduled jobs, not AI agent invocation.

**Why this priority**: Security consideration - cleanup is a destructive operation that should be human-controlled.

**Independent Test**: Can be fully tested by listing MCP tools and verifying no "cleanup" tool exists.

**Acceptance Scenarios**:

1. **Given** an MCP client lists available tools, **When** the tools response is returned, **Then** no tool named "cleanup" is present.

---

### Edge Cases

- What happens when cleanup runs outside of a tmux session? The offline recipient check cannot determine which windows exist, so this portion is skipped with a warning; other cleanup operations proceed normally.
- What happens when mailbox files are locked by another process during cleanup? Locked files are skipped with a warning to avoid blocking; the user can retry cleanup later.
- What happens when recipients.jsonl doesn't exist? Command completes successfully with zero recipients removed.
- What happens when .agentmail/mailboxes directory doesn't exist? Command completes successfully with zero messages/mailboxes removed.
- What happens when a message lacks a timestamp field? Messages without the `created_at` field cannot be evaluated for age-based cleanup and are skipped (not deleted).
- What happens when multiple cleanup processes run simultaneously? Each process skips files it cannot lock within 1 second and continues; no blocking or failure occurs.

## Requirements *(mandatory)*

### Functional Requirements

#### Offline Recipient Cleanup (US1)

- **FR-001** [Event-Driven]: When the user runs `agentmail cleanup`, the cleanup command shall compare each recipient in recipients.jsonl against current tmux window names.
- **FR-002** [Event-Driven]: When a recipient name does not match any current tmux window, the cleanup command shall remove that recipient from recipients.jsonl.

#### Stale Recipient Cleanup (US2)

- **FR-003** [Ubiquitous]: The cleanup command shall use 48 hours as the default stale threshold.
- **FR-004** [Event-Driven]: When the user provides `--stale-hours <N>`, the cleanup command shall use N hours as the stale threshold.
- **FR-005** [Event-Driven]: When a recipient's `updated_at` timestamp is older than the stale threshold, the cleanup command shall remove that recipient from recipients.jsonl.

#### Message Cleanup (US3)

- **FR-006** [Ubiquitous]: The cleanup command shall use 2 hours as the default delivered threshold.
- **FR-007** [Event-Driven]: When the user provides `--delivered-hours <N>`, the cleanup command shall use N hours as the delivered threshold.
- **FR-008** [Event-Driven]: When a message has `read_flag: true` and `created_at` older than the delivered threshold, the cleanup command shall remove that message from its mailbox file.
- **FR-009** [Ubiquitous]: The cleanup command shall never delete messages with `read_flag: false`.
- **FR-010** [Event-Driven]: When a message lacks a `created_at` field, the cleanup command shall skip that message during age-based cleanup.

#### Mailbox Cleanup (US4)

- **FR-011** [Event-Driven]: When a mailbox file contains zero messages after message cleanup, the cleanup command shall delete that file.

#### Exclusions (US5 & US6)

- **FR-012** [Ubiquitous]: The `agentmail onboard` command shall not include any reference to the cleanup command.
- **FR-013** [Ubiquitous]: The MCP server shall not expose a cleanup tool.

#### Output & Dry-Run

- **FR-014** [Event-Driven]: When cleanup completes, the cleanup command shall output a summary showing counts of removed recipients, removed messages, and removed mailbox files.
- **FR-015** [Event-Driven]: When the user provides `--dry-run`, the cleanup command shall report what would be cleaned without modifying any files.

#### File Locking

- **FR-016** [Ubiquitous]: The cleanup command shall acquire exclusive flock (LOCK_EX) on files before modifying them.
- **FR-017** [Unwanted]: If a file cannot be locked within 1 second, then the cleanup command shall skip that file.
- **FR-018** [Unwanted]: If a file cannot be locked within 1 second, then the cleanup command shall output a warning message.
- **FR-019** [Unwanted]: If a file is skipped due to lock timeout, then the cleanup command shall continue with remaining files.

#### Graceful Handling

- **FR-020** [Unwanted]: If the cleanup command is not running inside a tmux session, then the cleanup command shall skip the offline recipient check.
- **FR-021** [Unwanted]: If the cleanup command is not running inside a tmux session, then the cleanup command shall output a warning about skipped offline check.
- **FR-022** [Unwanted]: If recipients.jsonl does not exist, then the cleanup command shall report zero recipients removed.
- **FR-023** [Unwanted]: If the mailboxes directory does not exist, then the cleanup command shall report zero messages and mailboxes removed.

### Key Entities

- **Recipient State**: Stored in recipients.jsonl with fields: recipient (name), status, updated_at, notified_at, last_read_at
- **Message**: Stored in mailbox files (.agentmail/mailboxes/<name>.jsonl) with fields: id, from, to, message, read_flag, created_at (timestamp for age-based cleanup)
- **Mailbox File**: A JSONL file containing messages for a specific recipient

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Running cleanup removes 100% of recipients that don't correspond to existing tmux windows
- **SC-002**: Running cleanup removes 100% of recipients with updated_at older than the configured threshold
- **SC-003**: Running cleanup removes 100% of read messages older than the configured threshold while preserving 100% of unread messages
- **SC-004**: Running cleanup removes 100% of empty mailbox files
- **SC-005**: Cleanup completes without error when recipients.jsonl or mailboxes directory does not exist
- **SC-006**: Cleanup summary accurately reports the count of removed items (verified by manual counting)
- **SC-007**: Dry-run mode produces identical counts as actual cleanup without modifying any files

## Assumptions

- The Message struct will be extended to include a `created_at` timestamp field (RFC 3339 format) for age-based cleanup. Existing messages without this field will be skipped during age-based cleanup.
- Cleanup is intended for occasional manual or scheduled execution, not continuous background operation.
- The current tmux session is the authoritative source for which recipients are "online" (have corresponding windows).
- File locking uses the same flock mechanism as other AgentMail operations.
- Integer hours provide sufficient granularity for threshold configuration (no need for minutes/seconds).

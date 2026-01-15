# Feature Specification: Claude Code Hooks Integration

**Feature Branch**: `005-claude-hooks-integration`
**Created**: 2026-01-12
**Status**: Implemented
**Input**: User description: "Add integration with Claude hooks for agentmail. Add --hook flag to receive command: (1) no messages or not in tmux = silent exit code 0, (2) new messages = output to STDERR with 'You got new mail' and exit code 2, (3) README.md section for hook installation"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Claude Code Hook Notification (Priority: P1)

A Claude Code user configures agentmail as a hook to get notified when other agents send them messages. When running the receive command with the `--hook` flag, they receive immediate feedback about pending messages without interrupting normal workflow.

**Why this priority**: This is the core functionality that enables the hook integration. Without this, the feature has no value.

**Independent Test**: Can be fully tested by running `agentmail receive --hook` in various scenarios (with/without messages, inside/outside tmux) and verifying exit codes and output streams.

**Acceptance Scenarios**:

1. **Given** the user is in a tmux session with unread messages, **When** they run `agentmail receive --hook`, **Then** the message content is written to STDERR prefixed with "You got new mail" and exit code is 2
2. **Given** the user is in a tmux session with no unread messages, **When** they run `agentmail receive --hook`, **Then** no output is produced and exit code is 0
3. **Given** the user is not in a tmux session, **When** they run `agentmail receive --hook`, **Then** no output is produced and exit code is 0

---

### User Story 2 - Documentation for Hook Setup (Priority: P2)

A user wants to set up agentmail as a Claude Code hook and needs clear instructions on how to configure it properly in their Claude Code settings.

**Why this priority**: Users need documentation to discover and correctly use the hook integration feature.

**Independent Test**: Can be verified by following the README instructions to configure the hook and testing that it works as described.

**Acceptance Scenarios**:

1. **Given** a user reads the README.md Claude Code Hooks section, **When** they follow the setup instructions, **Then** they can successfully configure agentmail as a hook
2. **Given** the README documentation, **When** a user reviews it, **Then** they understand the exit code behavior and when notifications appear

---

### Edge Cases

- What happens when the `--hook` flag is used with the send command? Flag is only valid with `receive` command; other commands ignore it or return an error.
- What happens when mailbox file is corrupted? Exit silently with code 0 (non-disruptive behavior for hooks).
- What happens when file lock cannot be acquired within timeout? Exit silently with code 0 (hooks should not block).
- What happens when multiple messages are pending? Only the first (oldest) message is displayed and consumed.

## Requirements *(mandatory)*

### Functional Requirements

#### Hook Mode Message Notification
- **FR-001a** [Event-Driven]: When the user runs `agentmail receive --hook` and unread messages exist, the agentmail receive command shall write "You got new mail\n" followed by the message content to STDERR.
- **FR-001b** [Event-Driven]: When the user runs `agentmail receive --hook` and unread messages exist, the agentmail receive command shall exit with code 2.
- **FR-001c** [Event-Driven]: When the user runs `agentmail receive --hook` and unread messages exist, the agentmail receive command shall mark the oldest unread message as read.

#### Hook Mode Silent Exit Conditions
- **FR-002** [Event-Driven]: When the user runs `agentmail receive --hook` and no unread messages exist, the agentmail receive command shall exit with code 0 and produce no output.
- **FR-003** [Unwanted Behavior]: If `agentmail receive --hook` is executed outside a tmux session, then the agentmail receive command shall exit with code 0 and produce no output.

#### Hook Mode Error Handling
- **FR-004a** [Unwanted Behavior]: If a file read error occurs during `agentmail receive --hook` execution, then the agentmail receive command shall exit with code 0 and produce no output.
- **FR-004b** [Unwanted Behavior]: If a file lock cannot be acquired during `agentmail receive --hook` execution, then the agentmail receive command shall exit with code 0 and produce no output.
- **FR-004c** [Unwanted Behavior]: If the mailbox file is corrupted during `agentmail receive --hook` execution, then the agentmail receive command shall exit with code 0 and produce no output.

#### Hook Mode Output Stream
- **FR-005** [State-Driven]: While hook mode is enabled (`--hook` flag present), the agentmail receive command shall write all output to STDERR.

#### Documentation
- **FR-006** [Ubiquitous]: The agentmail documentation shall include a "Claude Code Hooks" section containing the `--hook` flag behavior and configuration instructions.

### Key Entities

- **Hook Flag (`--hook`)**: Boolean command-line flag that enables hook mode on the receive command
- **Exit Codes**:
  - `0`: No action required (no messages, not in tmux, or error in hook mode)
  - `2`: Message notification available (unread messages exist in hook mode)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can configure agentmail as a Claude Code hook within 2 minutes using README instructions
- **SC-002**: Hook mode exits silently (no output, exit code 0) 100% of the time when no messages exist or not in tmux
- **SC-003**: Hook mode produces notification (STDERR output, exit code 2) 100% of the time when unread messages exist in tmux
- **SC-004**: Hook execution completes within 500 milliseconds under normal conditions

## Assumptions

- Claude Code hooks execute shell commands and can read STDERR output for display to users
- Exit code 2 is appropriate for indicating "notification" state (distinct from error code 1)
- Users have existing familiarity with Claude Code hooks configuration
- The `--hook` flag is only meaningful for the `receive` command
- One message is consumed per hook invocation (same behavior as regular receive)
- Silent failure (exit 0) is preferable to noisy errors for hook integrations

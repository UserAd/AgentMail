# Feature Specification: Mailman Daemon

**Feature Branch**: `006-mailman-daemon`
**Created**: 2026-01-12
**Status**: Draft
**Input**: User description: "Add mailman command which runs a mailer daemon for inter-agent notifications with state management and singleton process control"

## Clarifications

### Session 2026-01-12

- Q: When should the notification counter for an agent reset (allowing new notifications)? → A: When state changes to work/offline
- Q: Should the mailman daemon run in foreground or background? → A: Configurable via flag (--daemon)
- Q: When the mailman daemon restarts, how should it handle existing agent states? → A: Clear stale states older than 1 hour
- State command interface: Use `agentmail status <STATUS>` for hooks integration - returns exit code 0, empty output, silently succeeds when not in tmux

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Start Mailman Daemon (Priority: P1)

An agent operator wants to start the mailman daemon to enable automatic notification delivery to agents when they have unread messages. The daemon can run in foreground (default) or background mode with the `--daemon` flag.

**Why this priority**: This is the core functionality - without starting the daemon, no notifications can be delivered. All other features depend on this working correctly.

**Independent Test**: Can be fully tested by running `agentmail mailman` and verifying the daemon starts, creates a PID file, and begins its monitoring loop.

**Acceptance Scenarios**:

1. **Given** no mailman daemon is running, **When** `agentmail mailman` is invoked, **Then** the daemon starts in foreground mode and outputs status to stdout
2. **Given** no mailman daemon is running, **When** `agentmail mailman --daemon` is invoked, **Then** the daemon starts in background mode and detaches from the terminal
3. **Given** the daemon is running, **When** checking the PID file in `.git/mail/`, **Then** a valid PID file exists with the current process ID

---

### User Story 2 - Singleton Process Control (Priority: P1)

An agent operator accidentally tries to start a second mailman daemon. The system prevents duplicate daemons to avoid race conditions and duplicate notifications.

**Why this priority**: Critical for system integrity - duplicate daemons would cause double notifications and potential file corruption.

**Independent Test**: Can be fully tested by starting one daemon, then attempting to start a second one and verifying the error message and exit code.

**Acceptance Scenarios**:

1. **Given** a mailman daemon is already running, **When** `agentmail mailman` is invoked again, **Then** the command displays an error message and exits with code 2
2. **Given** a stale PID file exists (process no longer running), **When** `agentmail mailman` is invoked, **Then** a warning is displayed and the daemon starts successfully

---

### User Story 3 - Agent State Management (Priority: P1)

Agents need to register their availability status so the mailman knows when to send notifications. The Claude Code hooks integration invokes `agentmail status <STATUS>` commands automatically during agent lifecycle events. The command is designed for hooks: silent operation, always succeeds.

**Why this priority**: State tracking is essential for the mailman to know which agents can receive notifications. Without it, notifications would be sent to offline or busy agents.

**Independent Test**: Can be fully tested by invoking `agentmail status ready`, `agentmail status work`, and `agentmail status offline` commands and verifying state changes in `.git/mail-recipients.jsonl`.

**Acceptance Scenarios**:

1. **Given** an agent session starts in tmux, **When** `agentmail status ready` is invoked, **Then** the agent is marked as "ready" in the recipients state file, command outputs nothing, and exits with code 0
2. **Given** a user submits a message in tmux, **When** `agentmail status work` is invoked, **Then** the agent is marked as "work" in the recipients state file, command outputs nothing, and exits with code 0
3. **Given** an agent session ends in tmux, **When** `agentmail status offline` is invoked, **Then** the agent is marked as "offline" in the recipients state file, command outputs nothing, and exits with code 0
4. **Given** a STOP hook triggers in tmux, **When** `agentmail status ready` is invoked, **Then** the agent transitions back to "ready" status
5. **Given** the command is invoked outside of tmux, **When** `agentmail status ready` is invoked, **Then** the command does nothing, outputs nothing, and exits with code 0

---

### User Story 4 - Notification Delivery (Priority: P2)

The mailman daemon monitors all mailboxes and sends notifications to agents in "ready" state when they have unread messages. This enables proactive message checking without agents polling constantly. Notification tracking resets when an agent transitions to work or offline state.

**Why this priority**: This is the main value proposition of the mailman, but depends on the daemon being running and agents having registered states.

**Independent Test**: Can be fully tested by setting an agent to "ready" state, sending a message to that agent, and verifying the tmux notification is delivered.

**Acceptance Scenarios**:

1. **Given** an agent is in "ready" status with unread messages, **When** the mailman checks for unread messages, **Then** it sends "Check your agentmail" to the recipient's tmux window using tmux send-keys followed by Enter after 1 second
2. **Given** an agent has already been notified about unread messages, **When** the mailman checks again before state changes, **Then** no additional notification is sent
3. **Given** an agent is in "work" or "offline" status, **When** the mailman checks for unread messages, **Then** no notification is sent to that agent
4. **Given** an agent was previously notified and then transitions to work/offline then back to ready, **When** the mailman checks for unread messages, **Then** a new notification is sent

---

### Edge Cases

- What happens when the `.git/mail/` directory does not exist? The mailman shall create it.
- What happens when the PID file is corrupted (non-numeric content)? The mailman shall treat it as stale and warn.
- What happens when tmux is not available or the session is not valid? The mailman shall skip notification for that recipient and log a warning.
- What happens when the recipients state file does not exist? The mailman shall create it on first state update.
- What happens when the mailman receives SIGTERM/SIGINT? The mailman shall clean up the PID file and exit gracefully.
- What happens when agent states are older than 1 hour on daemon startup? The mailman shall clear those stale states.
- What happens when `agentmail status` is invoked outside tmux? The command shall do nothing and exit with code 0.
- What happens when `agentmail status` is invoked with an invalid status name? The command shall output an error and exit with code 1.

## Requirements *(mandatory)*

### Functional Requirements

**Daemon Lifecycle:**

- **FR-001**: When `agentmail mailman` is invoked without flags, the mailman shall run in foreground mode and output status messages to stdout.
- **FR-002**: When `agentmail mailman --daemon` is invoked, the mailman shall detach from the terminal and run in background mode.
- **FR-003**: When `agentmail mailman` is invoked, the mailman shall check for an existing PID file at `.git/mail/mailman.pid`.
- **FR-004**: If a PID file exists with a running process, then the mailman shall display "Mailman daemon already running (PID: <pid>)" and exit with code 2.
- **FR-005**: If a PID file exists but the process is not running, then the mailman shall display "Warning: Stale PID file found, cleaning up" and proceed to start the daemon.
- **FR-006**: When the mailman daemon starts successfully, the mailman shall write its PID to `.git/mail/mailman.pid`.
- **FR-007**: When the mailman daemon receives SIGTERM or SIGINT, the mailman shall delete the PID file and exit with code 0.

**Status Command (for Hooks Integration):**

- **FR-008**: When `agentmail status ready` is invoked inside tmux, the status command shall set the current tmux window's state to "ready" in `.git/mail-recipients.jsonl`.
- **FR-009**: When `agentmail status work` is invoked inside tmux, the status command shall set the current tmux window's state to "work" in `.git/mail-recipients.jsonl`.
- **FR-010**: When `agentmail status offline` is invoked inside tmux, the status command shall set the current tmux window's state to "offline" in `.git/mail-recipients.jsonl`.
- **FR-011**: When `agentmail status <STATUS>` completes successfully, the status command shall output nothing and exit with code 0.
- **FR-012**: When `agentmail status <STATUS>` is invoked outside of a tmux session, the status command shall do nothing, output nothing, and exit with code 0.
- **FR-013**: If `agentmail status` is invoked with an invalid status name (not ready/work/offline), then the status command shall output "Invalid status: <name>. Valid: ready, work, offline" and exit with code 1.
- **FR-014**: The status command shall store each recipient's state as a JSONL entry containing recipient name, status, last updated timestamp, and notified flag.
- **FR-015**: The status command shall use file locking when updating the recipients state file to prevent race conditions.

**State Cleanup:**

- **FR-016**: When the mailman daemon starts, the mailman shall clear agent states with timestamps older than 1 hour from `.git/mail-recipients.jsonl`.

**Notification Loop:**

- **FR-017**: While the daemon is running, the mailman shall check for unread messages across all recipients every 10 seconds.
- **FR-018**: When a recipient has unread messages and is in "ready" state and has not been notified, the mailman shall send "Check your agentmail" to the recipient's tmux window using `tmux send-keys`.
- **FR-019**: When a notification is sent, the mailman shall wait 1 second and then send Enter key to the recipient's tmux window.
- **FR-020**: When a notification is sent, the mailman shall set the notified flag to true for that recipient.
- **FR-021**: When a recipient's state changes to "work" or "offline", the status command shall reset the notified flag to false for that recipient.
- **FR-022**: While a recipient is in "work" or "offline" state, the mailman shall not send notifications to that recipient.

**Error Handling:**

- **FR-023**: If the `.git/mail/` directory does not exist, then the mailman shall create it before proceeding.
- **FR-024**: If the tmux window specified in a recipient's mailbox does not exist, then the mailman shall skip notification for that recipient.

### Key Entities

- **Mailman PID File**: Singleton lock file at `.git/mail/mailman.pid` containing the daemon's process ID
- **Recipients State**: JSONL file at `.git/mail-recipients.jsonl` tracking each agent's current status (ready/work/offline), last update timestamp, and notification flag
- **Agent Status**: Enumerated state value - "ready" (available for notifications), "work" (busy processing), "offline" (session ended)
- **Notification Cycle**: 10-second interval during which the daemon checks all mailboxes and sends notifications
- **Notified Flag**: Boolean indicating whether an agent has been notified since last state transition to work/offline

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Only one mailman daemon instance can run at a time within a git repository
- **SC-002**: Agents in "ready" state receive notification within 10 seconds of message arrival
- **SC-003**: Agents in "work" or "offline" state receive zero notifications while in that state
- **SC-004**: State changes take effect within 1 second of command invocation
- **SC-005**: Daemon cleanly shuts down without leaving stale PID files when receiving termination signals
- **SC-006**: 100% of stale PID files are detected and cleaned up on daemon startup
- **SC-007**: Agents receive at most one notification per ready session (reset only on state transition)
- **SC-008**: Agent states older than 1 hour are cleared on daemon startup
- **SC-009**: Status command produces zero output on success (empty stdout and stderr)
- **SC-010**: Status command always exits with code 0 when run outside tmux (graceful no-op)

## Assumptions

- The git repository root contains a `.git/` directory (standard git structure)
- tmux is installed and available in the system PATH
- File locking is supported on the target filesystem
- Agents use Claude Code hooks to invoke status commands automatically
- Status command is designed for hook usage - silent, always succeeds in non-tmux environments

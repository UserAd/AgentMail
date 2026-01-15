# Feature Specification: Stale Agent Notification Support

**Feature Branch**: `008-stale-agent-mailman`
**Created**: 2026-01-13
**Status**: Implemented
**Input**: User description: "Implement stale agent handling and notifying - fallback notification support for agents without Claude hooks that cannot register their status"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Stateless Agent Receives Mail Notifications (Priority: P1)

An agent running in a tmux window without Claude hooks support receives mail from another agent. Since this agent cannot call `agentmail status ready/work/offline`, it has no entry in the recipients state file. Despite this, the mailman daemon discovers it has unread mail and sends a notification to the tmux window at a reasonable interval.

**Why this priority**: This is the core value proposition - ensuring all agents can receive mail notifications regardless of their ability to register status, preventing messages from going unnoticed indefinitely.

**Independent Test**: Can be fully tested by sending mail to an agent without hooks and verifying notifications arrive periodically until mail is read.

**Acceptance Scenarios**:

1. **Given** an agent is running in tmux window "agent1" without status registration, **When** another agent sends mail to "agent1", **Then** the mailman daemon sends a notification to the agent1 window within 60 seconds.

2. **Given** an agent has unread mail and has not registered status, **When** 60 seconds elapse since the last notification, **Then** the mailman daemon sends another notification.

3. **Given** an agent has received a notification about unread mail, **When** the agent reads all unread messages, **Then** the mailman daemon stops sending notifications to that agent.

---

### User Story 2 - Stated Agents Take Precedence (Priority: P2)

An agent that was previously stateless (receiving periodic notifications) registers its status using `agentmail status ready`. The mailman daemon recognizes this transition and applies the stated agent notification logic (one-time notification when ready) instead of the periodic stateless notifications.

**Why this priority**: Ensures seamless transition between stateless and stated modes without duplicate or conflicting notifications.

**Independent Test**: Can be tested by starting with a stateless agent, sending mail, observing periodic notifications, then registering status and verifying notification behavior changes.

**Acceptance Scenarios**:

1. **Given** an agent has been receiving periodic stateless notifications, **When** the agent registers status as "ready", **Then** the mailman daemon applies stated agent notification rules.

2. **Given** a stated agent with status "ready" has unread mail, **When** the agent is notified once, **Then** no further notifications occur until the agent returns to "ready" status.

---

### User Story 3 - Daemon Restart Clears Tracking State (Priority: P3)

The mailman daemon is restarted while stateless agents have unread mail. Upon restart, the in-memory tracking of stateless agent notification times is cleared, allowing immediate notification to all eligible stateless agents.

**Why this priority**: Ensures no mail notifications are missed due to daemon restarts, accepting that agents may receive an immediate notification after restart.

**Independent Test**: Can be tested by sending mail to a stateless agent, waiting for notification, restarting daemon, and verifying immediate notification occurs.

**Acceptance Scenarios**:

1. **Given** a stateless agent was last notified 30 seconds ago, **When** the mailman daemon restarts, **Then** the agent receives a notification within 10 seconds of daemon startup (at next loop iteration).

---

### Edge Cases

- What happens when a stateless agent's tmux window disappears? The daemon removes the tracking entry for windows that no longer exist.
- How does the system handle an agent that rapidly switches between stated and stateless? The stated agent logic always takes precedence when a status entry exists.
- What if multiple stateless agents have unread mail? All eligible stateless agents are checked and notified independently each loop iteration.

## Requirements *(mandatory)*

### Functional Requirements

#### Discovery & Identification

- **FR-001**: When the daemon completes a 10-second loop iteration, the daemon shall scan all mailbox files in `.agentmail/mailboxes/` to identify recipients.

- **FR-002**: When the daemon identifies a mailbox recipient, the daemon shall check whether that recipient has an entry in `recipients.jsonl`.

- **FR-003**: When a mailbox recipient has no entry in `recipients.jsonl` and has unread messages, the daemon shall classify that recipient as a stateless agent.

#### Stateless Notification Behavior

- **FR-004**: When the daemon completes a loop iteration and a stateless agent has unread messages, the daemon shall send a notification to that agent's tmux window if either: (a) the agent has not been notified before, or (b) 60 seconds have elapsed since the agent's last notification.

- **FR-006**: When a stateless agent's mailbox contains zero unread messages, the daemon shall skip notification for that agent.

#### State Precedence

- **FR-007**: While an agent has an entry in `recipients.jsonl`, the daemon shall apply stated agent notification logic instead of stateless notification logic.

- **FR-008**: When a stateless agent registers status via `agentmail status`, the daemon shall immediately cease stateless notifications for that agent.

#### Tracking & Lifecycle

- **FR-009**: The daemon shall maintain an in-memory map of stateless agent window names to last-notification timestamps.

- **FR-010**: When the mailman daemon starts, the daemon shall initialize the stateless tracker with an empty map.

- **FR-011**: When the daemon detects a mailbox file no longer exists for a tracked window, the daemon shall remove that window from the stateless tracker.

#### Concurrency & Loop Integration

- **FR-012**: The daemon shall perform stateless agent checks within the existing 10-second notification loop, after stated agent processing completes.

- **FR-013**: The daemon shall protect the stateless tracker map with a mutex to ensure thread-safe access.

#### Error Handling

- **FR-014**: If the daemon fails to read the mailboxes directory, then the daemon shall log the error and continue to the next loop iteration.

- **FR-015**: If the daemon fails to send a notification to a stateless agent's tmux window, then the daemon shall skip marking that agent as notified and retry on the next eligible interval.

- **FR-016**: If the daemon fails to read a mailbox file for unread message count, then the daemon shall skip that agent for the current iteration and log a warning.

- **FR-017**: If the daemon fails to read `recipients.jsonl`, then the daemon shall treat all mailbox recipients as potentially stateless and apply stateless notification logic.

### Key Entities

- **Stateless Agent**: An agent running in a tmux window that has a mailbox with unread messages but no entry in the recipients state file. Identified by the presence of a mailbox file without corresponding status registration.

- **Stateless Tracker**: An in-memory data structure mapping tmux window names to their last notification timestamp. Used to enforce the 60-second notification interval for stateless agents.

- **Stated Agent**: An agent that has registered its status using `agentmail status`. Has an entry in `recipients.jsonl` and follows existing notification rules.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Stateless agents receive their first mail notification within 10 seconds of mail arrival (at the next daemon loop iteration).

- **SC-002**: Stateless agents receive repeated notifications at 60-second intervals (±5 seconds tolerance) until all messages are read.

- **SC-003**: When an agent transitions from stateless to stated, zero duplicate notifications occur in the same loop iteration.

- **SC-004**: After daemon restart, stateless agents with unread mail receive notification within 10 seconds (first loop iteration).

- **SC-005**: All existing stated agent tests pass with zero modifications required.

- **SC-006**: New code achieves minimum 80% test coverage as measured by `go test -cover`.

- **SC-007**: Zero race conditions detected when running `go test -race ./internal/daemon/...`.

### Requirements Traceability

| Success Criteria | Traced Requirements |
|-----------------|---------------------|
| SC-001 | FR-001, FR-003, FR-004 |
| SC-002 | FR-004, FR-006 |
| SC-003 | FR-007, FR-008 |
| SC-004 | FR-010 |
| SC-005 | FR-007, FR-012 |
| SC-006 | (quality gate) |
| SC-007 | FR-013 |

## Assumptions

- The mailman daemon loop continues to run at 10-second intervals as currently implemented.
- Stateless agents are identified by the presence of a mailbox file in `.agentmail/mailboxes/` combined with absence from `recipients.jsonl`.
- The 60-second interval for stateless notifications is appropriate to avoid excessive notifications while ensuring timely delivery.
- In-memory tracking is acceptable (vs persistent storage) since daemon restart simply allows immediate re-notification.
- Tmux window existence can be determined via the existing tmux integration used for notifications.

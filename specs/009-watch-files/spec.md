# Feature Specification: File-Watching for Mailman with Timer Fallback

**Feature Branch**: `009-watch-files`
**Created**: 2026-01-13
**Status**: Implemented
**Input**: User description: "I want to mailman watch for file changes instead of timer (with fallback if OS not support watching) to check mailbox adding/updating and agents status changing. As additional thing I want to agentmail tracks in recipients.jsonl when agent reads his mail last time."

## Clarifications

### Session 2026-01-13

- Q: What identifier should be used for last-read tracking when `agentmail receive` is invoked outside of tmux? → A: Skip tracking - do not update last-read timestamp when outside tmux
- Q: Should the mailman daemon periodically attempt to re-enable file-watching after falling back to polling mode? → A: No recovery - stay in polling mode until daemon restart

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Instant Notification on New Mail (Priority: P1)

An agent sends a message to another agent. Instead of waiting up to 10 seconds for the polling interval, the mailman daemon detects the mailbox file change immediately via OS file-watching and processes the notification within 1 second of the message being written.

**Why this priority**: This is the core value proposition - reducing notification latency from up to 10 seconds to near-instant, significantly improving inter-agent communication responsiveness.

**Independent Test**: Can be fully tested by starting mailman, sending a message, and measuring the time between send completion and notification delivery.

**Acceptance Scenarios**:

1. **Given** an agent is in "ready" status and mailman is running with file-watching enabled, **When** another agent sends a message to the ready agent, **Then** the ready agent receives notification within 2 seconds of message delivery.

2. **Given** the mailman daemon is running with file-watching enabled, **When** a new mailbox file is created (first message to a new recipient), **Then** the daemon detects the new file and processes it within 2 seconds.

---

### User Story 2 - Instant Status Change Detection (Priority: P1)

An agent changes its status from "work" to "ready" using `agentmail status ready`. The mailman daemon immediately detects this change and checks for pending mail, rather than waiting for the next polling cycle.

**Why this priority**: Status changes are time-critical - when an agent becomes ready, it should receive pending notifications immediately, not after up to 10 seconds of delay.

**Independent Test**: Can be fully tested by setting status to "work", sending a message, then setting status to "ready" and verifying notification arrives within 2 seconds.

**Acceptance Scenarios**:

1. **Given** an agent has status "work" with unread messages, **When** the agent sets status to "ready", **Then** the agent receives notification within 2 seconds of the status change.

2. **Given** the mailman daemon is running with file-watching, **When** any agent updates their status, **Then** the daemon detects the `recipients.jsonl` file change and processes it within 2 seconds.

---

### User Story 3 - Graceful Fallback to Polling (Priority: P2)

An operator runs AgentMail on a system or filesystem that does not support file-system watching (e.g., some network filesystems, older systems, or when watch resources are exhausted). The mailman daemon automatically falls back to timer-based polling without user intervention.

**Why this priority**: Ensures AgentMail works reliably across all environments - the system should "just work" rather than fail or require manual configuration.

**Independent Test**: Can be fully tested by simulating watch unavailability and verifying the daemon continues operating with polling.

**Acceptance Scenarios**:

1. **Given** the mailman daemon starts on a system without file-watching support, **When** watch initialization fails, **Then** the daemon logs a warning and continues operating with 10-second polling.

2. **Given** the mailman daemon is running with file-watching, **When** the file watcher encounters an error mid-operation, **Then** the daemon logs a warning and falls back to polling mode.

3. **Given** the mailman daemon is running in fallback polling mode, **When** the daemon completes a polling cycle, **Then** the daemon checks both mailboxes and recipient states as before.

---

### User Story 4 - Track Agent Last-Read Time (Priority: P2)

An agent reads their mail using `agentmail receive`. The system records when this agent last read their mailbox, enabling future features like "time since last active" metrics and helping operators understand agent responsiveness patterns.

**Why this priority**: Provides visibility into agent activity patterns and enables future observability features. Useful for debugging and monitoring inter-agent communication health.

**Independent Test**: Can be fully tested by reading mail and verifying the last-read timestamp is updated in `recipients.jsonl`.

**Acceptance Scenarios**:

1. **Given** an agent has unread messages, **When** the agent runs `agentmail receive`, **Then** the system updates the agent's last-read timestamp in `recipients.jsonl`.

2. **Given** an agent has no unread messages, **When** the agent runs `agentmail receive`, **Then** the system still updates the agent's last-read timestamp (agent checked mail, even if empty).

3. **Given** an agent has never been tracked in `recipients.jsonl`, **When** the agent runs `agentmail receive` for the first time, **Then** a new entry is created with the current timestamp.

---

### Edge Cases

- What happens when the watch directory does not exist? The daemon shall create it and set up watches after creation.
- What happens when the watcher receives a burst of events? The daemon shall debounce events to avoid processing the same change multiple times within 500 milliseconds.
- What happens when `recipients.jsonl` is watched but doesn't exist yet? The daemon shall watch the parent directory and create the watch when the file is created.
- How does the daemon handle macOS, Linux, and Windows differences? The daemon shall use a cross-platform file-watching approach that handles OS-specific behaviors transparently.
- What happens when an agent reads mail outside of tmux? The receive command shall skip updating the last-read timestamp (tracking only applies to tmux agents).

## Requirements *(mandatory)*

### Functional Requirements

**File-Watching Infrastructure:**

- **FR-001**: When `agentmail mailman` starts, the mailman shall attempt to initialize file-system watchers for the `.agentmail/` directory and the `.agentmail/mailboxes/` directory.

- **FR-002a**: When file-watching initialization succeeds, the mailman shall log "File watching enabled" to stdout.

- **FR-002b**: When file-watching initialization succeeds, the mailman shall enter file-watching mode and use event-driven monitoring.

- **FR-003a**: If file-watching initialization fails, then the mailman shall log "File watching unavailable, using polling" to stdout.

- **FR-003b**: If file-watching initialization fails, then the mailman shall enter polling mode with 10-second intervals.

- **FR-004**: While in file-watching mode, the mailman shall monitor all files in the `.agentmail/mailboxes/` directory for write and create events.

- **FR-005**: While in file-watching mode, the mailman shall monitor the `.agentmail/recipients.jsonl` file for write events.

**Directory and File Creation:**

- **FR-006**: When the `.agentmail/` directory does not exist at daemon startup, the mailman shall create the directory before initializing watchers.

- **FR-007**: When the `.agentmail/mailboxes/` directory does not exist at daemon startup, the mailman shall create the directory before initializing watchers.

- **FR-008**: When `recipients.jsonl` does not exist at daemon startup, the mailman shall watch the parent `.agentmail/` directory and detect when the file is created.

**Event Processing:**

- **FR-009**: When a write or create event is received for a mailbox file in `.agentmail/mailboxes/`, the mailman shall trigger a notification check for that mailbox within 2 seconds.

- **FR-010a**: When a write event is received for `recipients.jsonl`, the mailman shall reload all agent states from the file within 2 seconds.

- **FR-010b**: When a write event is received for `recipients.jsonl`, the mailman shall process pending notifications for newly-ready agents within 2 seconds.

- **FR-011**: When multiple file change events occur within 500 milliseconds of each other, the mailman shall coalesce them using trailing-edge debouncing and process only once after the 500-millisecond window closes.

- **FR-012**: While in file-watching mode, the mailman shall execute a fallback notification check every 60 seconds to detect any events missed by the file watcher.

**Fallback Behavior:**

- **FR-013**: While in polling mode, the mailman shall check all mailboxes and recipient states for changes every 10 seconds.

- **FR-014a**: If the file watcher encounters a runtime error after initialization, then the mailman shall log the error message to stdout.

- **FR-014b**: If the file watcher encounters a runtime error after initialization, then the mailman shall switch from file-watching mode to polling mode.

- **FR-015**: While in polling mode after a fallback from file-watching mode, the mailman shall remain in polling mode until the daemon process is terminated and restarted.

- **FR-016**: The mailman shall automatically detect file-watching capability at startup without requiring user configuration or command-line flags.

**Last-Read Tracking:**

- **FR-017**: When `agentmail receive` is invoked inside a tmux session, the receive command shall update the `last_read_at` field for the current agent in `recipients.jsonl` to the current Unix timestamp in milliseconds.

- **FR-018**: The receive command shall store the `last_read_at` value as a Unix timestamp with millisecond precision (13-digit integer).

- **FR-019**: When `agentmail receive` is invoked for an agent that has no existing entry in `recipients.jsonl`, the receive command shall create a new entry with the `last_read_at` field set to the current timestamp.

- **FR-020**: When updating `recipients.jsonl`, the receive command shall acquire an exclusive file lock using `syscall.Flock` before writing and release it after writing completes.

- **FR-021**: When `agentmail receive` is invoked outside of a tmux session, the receive command shall not update the `last_read_at` field and shall not create a new entry in `recipients.jsonl`.

**Cross-Platform Compatibility:**

- **FR-022**: The mailman shall support file-watching on macOS using the kqueue mechanism.

- **FR-023**: The mailman shall support file-watching on Linux using the inotify mechanism.

- **FR-024**: The mailman shall support file-watching on Windows using the ReadDirectoryChangesW mechanism.

### Key Entities

- **File Watcher**: Component that monitors filesystem events for the `.agentmail/` directory tree and triggers processing on changes.
- **Debounce Timer**: Mechanism to coalesce rapid file change events within a 500-millisecond window.
- **Last-Read Timestamp**: Unix timestamp (milliseconds) recording when an agent last invoked `agentmail receive`.
- **Monitoring Mode**: Runtime state indicating whether the daemon is using event-driven watching or timer-based polling.
- **Fallback Timer**: 60-second safety timer that runs even during file-watching mode to catch missed events.

### Requirements Traceability

| User Story | Requirements | Description |
|------------|--------------|-------------|
| US-1: Instant Notification | FR-001, FR-002a/b, FR-004, FR-009, FR-011 | File-watching infrastructure and mailbox event processing |
| US-2: Instant Status Change | FR-005, FR-010a/b | Recipients.jsonl monitoring and state reload |
| US-3: Graceful Fallback | FR-003a/b, FR-013, FR-014a/b, FR-015, FR-016 | Fallback detection, polling mode, and recovery behavior |
| US-4: Last-Read Tracking | FR-017, FR-018, FR-019, FR-020, FR-021 | Timestamp storage, format, and file locking |
| Edge Cases | FR-006, FR-007, FR-008, FR-012 | Directory creation, file creation detection, and fallback timer |
| Cross-Platform | FR-022, FR-023, FR-024 | macOS, Linux, and Windows support |

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: While in file-watching mode, agents in "ready" state shall receive notification within 2 seconds of message arrival. *(Validates: FR-009)*

- **SC-002**: While in file-watching mode, status changes shall trigger notification checks within 2 seconds. *(Validates: FR-010a, FR-010b)*

- **SC-003**: When file-watching initialization fails, the daemon shall continue operating with 10-second polling intervals. *(Validates: FR-003b, FR-013)*

- **SC-004**: When multiple file events occur within 500 milliseconds, the system shall produce exactly one notification check (zero duplicates). *(Validates: FR-011)*

- **SC-005**: After `agentmail receive` is invoked inside tmux, the `recipients.jsonl` file shall contain a `last_read_at` timestamp within 1 second of the current time. *(Validates: FR-017, FR-018)*

- **SC-006**: All existing mailman tests shall pass without modification after implementing file-watching. *(Validates: backward compatibility)*

- **SC-007**: File-watching shall function correctly on macOS, Linux, and Windows without requiring OS-specific configuration flags. *(Validates: FR-022, FR-023, FR-024, FR-016)*

- **SC-008**: While in file-watching mode, the 60-second fallback timer shall execute notification checks to catch any missed file events. *(Validates: FR-012)*

- **SC-009**: When transitioning from file-watching mode to polling mode due to error, the daemon shall not terminate and shall continue processing notifications. *(Validates: FR-014a, FR-014b, FR-015)*

## Assumptions

- Go's standard library or commonly available packages support file-system watching on macOS, Linux, and Windows.
- File-watching is reliable enough for production use on local filesystems (may not work on network mounts).
- The 500-millisecond debounce window is sufficient to catch burst writes without adding noticeable latency.
- The 60-second fallback timer provides adequate safety margin without impacting normal operation.
- Updating `recipients.jsonl` on receive is acceptable overhead for tracking visibility.
- The `.agentmail/` directory is on a local filesystem (not a network mount) in typical usage.

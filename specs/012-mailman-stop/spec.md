# Feature Specification: Mailman Daemon Stop Command

**Feature Branch**: `012-mailman-stop`
**Created**: 2026-01-20
**Status**: Implemented
**Input**: User description: "Add ability to stop mailman daemon via file-based signaling mechanism"

## Clarifications

### Session 2026-01-20

- Q: Should the stop command verify the daemon has terminated, or return immediately after sending SIGTERM? → A: Fire-and-forget (create stop file and exit immediately)
- Q: Should the stop command validate that a daemon is running before creating the .stop file? → A: No, just create the file unconditionally
- Q: How should the daemon detect the .stop file? → A: File watcher event (use existing fsnotify)
- Q: What should happen if .stop file already exists when stop command runs? → A: Report "stop already pending" and exit with code 1

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Stop Running Daemon (Priority: P1)

An agent operator wants to gracefully stop a running mailman daemon. They invoke `agentmail mailman stop` which creates a stop signal file. The daemon detects this file via its file watcher and shuts down gracefully.

**Why this priority**: This is the core functionality - the primary use case for the stop command is to signal the daemon to terminate.

**Independent Test**: Can be fully tested by starting a daemon with `agentmail mailman`, then running `agentmail mailman stop` and verifying the daemon terminates and cleans up its files.

**Acceptance Scenarios**:

1. **Given** no `.stop` file exists in `.agentmail/`, **When** `agentmail mailman stop` is invoked, **Then** the command creates `.agentmail/.stop` file, outputs "Stop signal sent" and exits with code 0
2. **Given** a mailman daemon is running with file watcher, **When** `.agentmail/.stop` file is created, **Then** the daemon detects the file, removes it, and initiates graceful shutdown
3. **Given** a mailman daemon is shutting down, **When** shutdown completes, **Then** the daemon removes the PID file at `.agentmail/mailman.pid`

---

### User Story 2 - Stop Already Pending (Priority: P2)

An agent operator runs the stop command when a previous stop is already pending (stop file exists). The system informs them that a stop is already in progress.

**Why this priority**: Important for user feedback to avoid confusion about multiple stop attempts.

**Independent Test**: Can be fully tested by creating `.agentmail/.stop` file manually, then invoking `agentmail mailman stop` and verifying the error message.

**Acceptance Scenarios**:

1. **Given** `.agentmail/.stop` file already exists, **When** `agentmail mailman stop` is invoked, **Then** the command outputs "Stop already pending" to stderr and exits with code 1

## Requirements *(mandatory)*

### Functional Requirements

**Stop Command:**

- **FR-001**: When `agentmail mailman stop` is invoked, the CLI shall attempt to create the file `.agentmail/.stop`.
- **FR-002**: When the `.stop` file is created successfully, the CLI shall output "Stop signal sent" to stdout and exit with code 0.
- **FR-003**: If the `.stop` file already exists, then the CLI shall output "Stop already pending" to stderr and exit with code 1.
- **FR-004**: If the `.stop` file cannot be created due to a filesystem error, then the CLI shall output "Failed to send stop signal: <error>" to stderr and exit with code 1.

**Daemon Stop File Detection:**

- **FR-005**: While the daemon is running, the daemon shall monitor the `.agentmail/` directory for file creation events using the existing file watcher.
- **FR-006**: When the daemon detects creation of `.agentmail/.stop` file, the daemon shall initiate graceful shutdown.
- **FR-007**: When the daemon initiates shutdown, the daemon shall remove the `.agentmail/.stop` file.
- **FR-008**: When the daemon completes shutdown, the daemon shall remove the `.agentmail/mailman.pid` file.

### Key Entities

- **Stop Signal File**: File at `.agentmail/.stop` used to signal the daemon to shut down
- **Mailman PID File**: File at `.agentmail/mailman.pid` containing the daemon's process ID (existing)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Stop command creates signal file and returns within 1 second (fire-and-forget)
- **SC-002**: Daemon detects stop file within 1 second of creation (via file watcher event)
- **SC-003**: Daemon removes both `.stop` and `.pid` files during shutdown in 100% of cases
- **SC-004**: Stop command correctly identifies pending stop in 100% of cases
- **SC-005**: All error conditions produce clear, actionable error messages to stderr

## Assumptions

- The daemon's existing file watcher (fsnotify) can detect file creation in `.agentmail/` directory
- The daemon already has graceful shutdown logic via signal handling that can be reused
- File creation is atomic enough to serve as a reliable IPC mechanism
- The `.agentmail/` directory exists (created by daemon on startup)

# Feature Specification: Storage Directory Restructure

**Feature Branch**: `007-storage-restructure`
**Created**: 2026-01-13
**Status**: Draft
**Input**: User description: "Change storage structure for AgentMail to use .agentmail/ directory instead of .git/mail/"

## Clarifications

### Session 2026-01-13

- Q: Should AgentMail still require running inside a git repository, or should it work in any directory? → A: Git still required (maintains existing behavior, simpler implementation, repo root detection unchanged)
- Q: When old data exists in .git/mail/, should AgentMail display a one-time migration warning? → A: No warning (silent transition, users must discover migration need themselves)
- Q: Should AgentMail automatically add .agentmail/ to the repository's .gitignore file? → A: User manages .gitignore (no automatic changes, user decides whether to track mail data)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Send and Receive Messages with New Storage (Priority: P1)

Agents sending and receiving messages should work seamlessly with the new storage location. When an agent sends a message, it should be stored in `.agentmail/mailboxes/`. When an agent receives messages, they should be read from the same location.

**Why this priority**: This is the core functionality of AgentMail. If messaging doesn't work with the new storage structure, the entire application is broken.

**Independent Test**: Can be fully tested by sending a message from one tmux window to another and verifying the message is stored in `.agentmail/mailboxes/<recipient>.jsonl` and can be received.

**Acceptance Scenarios**:

1. **Given** no `.agentmail/` directory exists, **When** a user sends a message, **Then** the directory structure `.agentmail/mailboxes/` is created and the message is stored in `<recipient>.jsonl`
2. **Given** a message exists in `.agentmail/mailboxes/<recipient>.jsonl`, **When** the recipient runs `agentmail receive`, **Then** the message is displayed and marked as read
3. **Given** the `.agentmail/` directory already exists, **When** a user sends a message, **Then** the message is appended to the existing mailbox file

---

### User Story 2 - Mailman Daemon Uses New PID Location (Priority: P1)

The mailman daemon should store its PID file at `.agentmail/mailman.pid` instead of `.git/mail/mailman.pid`.

**Why this priority**: The daemon is essential for recipient status notifications. If the PID file location is inconsistent, daemon management (start/stop/status) will fail.

**Independent Test**: Can be fully tested by starting the mailman daemon and verifying the PID file is created at `.agentmail/mailman.pid`, then stopping it and verifying cleanup.

**Acceptance Scenarios**:

1. **Given** no daemon is running, **When** a user runs `agentmail mailman start`, **Then** the PID is written to `.agentmail/mailman.pid`
2. **Given** a daemon is running with PID in `.agentmail/mailman.pid`, **When** a user runs `agentmail mailman stop`, **Then** the daemon stops and the PID file is removed
3. **Given** a daemon is running, **When** a user runs `agentmail mailman status`, **Then** the status is read from `.agentmail/mailman.pid`

---

### User Story 3 - Recipients State Uses New Location (Priority: P1)

The recipients state file should be stored at `.agentmail/recipients.jsonl` instead of `.git/mail-recipients.jsonl`.

**Why this priority**: Recipients state is used by the mailman daemon and status commands. Inconsistent location would break status tracking.

**Independent Test**: Can be fully tested by registering a recipient status and verifying the state is stored in `.agentmail/recipients.jsonl`.

**Acceptance Scenarios**:

1. **Given** no recipients file exists, **When** the mailman daemon updates a recipient status, **Then** `.agentmail/recipients.jsonl` is created with the state
2. **Given** recipients exist in `.agentmail/recipients.jsonl`, **When** a user runs `agentmail recipients`, **Then** the list is read from the new location
3. **Given** recipients exist in `.agentmail/recipients.jsonl`, **When** the mailman daemon polls, **Then** it reads from and writes to the new location

---

### User Story 4 - Migration from Old Storage Location (Priority: P2)

Users with existing data in `.git/mail/` should be able to continue using AgentMail. The system should handle the transition gracefully.

**Why this priority**: Important for existing users but not critical for new installations. Ensures backward compatibility during transition.

**Independent Test**: Can be tested by creating old-format data in `.git/mail/` and verifying the application handles it appropriately.

**Acceptance Scenarios**:

1. **Given** data exists only in `.git/mail/`, **When** the user runs any agentmail command, **Then** the application operates using the new `.agentmail/` location (old data is not automatically migrated)
2. **Given** data exists in both locations, **When** the user runs any agentmail command, **Then** the application uses only the new `.agentmail/` location
3. **Given** users need to migrate data, **When** they manually move files from `.git/mail/` to `.agentmail/mailboxes/`, **Then** the application correctly reads the migrated data

---

### Edge Cases

- **Permission failure**: If `.agentmail/` directory cannot be created due to permissions, the system displays an error message indicating the failure reason (FR-009).
- **Concurrent access**: File locking mechanisms continue to operate identically in the new location, ensuring atomic operations.
- **No git repository**: AgentMail continues to require a git repository for operation; if no `.git/` directory exists, existing error handling applies.
- **Partial directory structure**: If `.agentmail/mailboxes/` exists but `.agentmail/` root doesn't (impossible state), the system creates missing parent directories as needed.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The AgentMail shall store all application files in the `.agentmail/` directory at the repository root.
- **FR-002**: The AgentMail shall store mailbox files in the `.agentmail/mailboxes/` subdirectory.
- **FR-003**: When a message is sent, the AgentMail shall create the file `.agentmail/mailboxes/<recipient>.jsonl` if it does not exist.
- **FR-004**: The AgentMail shall store the mailman daemon PID in `.agentmail/mailman.pid`.
- **FR-005**: The AgentMail shall store recipient state information in `.agentmail/recipients.jsonl`.
- **FR-006**: When any AgentMail command is executed, the AgentMail shall create the `.agentmail/` directory if it does not exist.
- **FR-007**: When the send command is executed, the AgentMail shall create the `.agentmail/mailboxes/` directory if it does not exist.
- **FR-008a**: The AgentMail shall use file permissions of 0750 for directories within `.agentmail/`.
- **FR-008b**: The AgentMail shall use file permissions of 0640 for files within `.agentmail/`.
- **FR-009**: If the `.agentmail/` directory cannot be created, then the AgentMail shall display an error message indicating the failure reason.
- **FR-010a**: The AgentMail shall maintain the existing JSONL file format for mailbox files.
- **FR-010b**: The AgentMail shall maintain the existing JSONL file format for the recipients file.
- **FR-011**: The AgentMail shall no longer read from or write to the `.git/mail/` directory.
- **FR-012**: The AgentMail shall no longer read from or write to the `.git/mail-recipients.jsonl` file.
- **FR-013**: The AgentMail shall continue to require execution within a git repository (presence of `.git/` directory).
- **FR-014**: The AgentMail shall not display any warning when old data exists in `.git/mail/`.
- **FR-015**: The AgentMail shall not automatically modify the repository's `.gitignore` file.

### Key Entities

- **Storage Root**: The `.agentmail/` directory at the repository root, containing all AgentMail data
- **Mailboxes Directory**: The `.agentmail/mailboxes/` subdirectory containing per-recipient JSONL files
- **Mailbox File**: A JSONL file at `.agentmail/mailboxes/<recipient>.jsonl` containing messages for a specific recipient
- **PID File**: The file at `.agentmail/mailman.pid` containing the process ID of the running mailman daemon
- **Recipients File**: The file at `.agentmail/recipients.jsonl` containing recipient availability states

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All AgentMail data files are stored within the `.agentmail/` directory hierarchy
- **SC-002**: No AgentMail operations read from or write to `.git/mail/` or `.git/mail-recipients.jsonl`
- **SC-003**: Existing functionality (send, receive, mailman, recipients, status) works correctly with the new storage location
- **SC-004**: Directory and file creation with appropriate permissions succeeds within 100 milliseconds
- **SC-005**: All existing tests pass after updating path constants
- **SC-006**: Users can manually migrate data by moving files from old to new locations

## Assumptions

- The repository root directory is writable by the user running AgentMail
- Users with existing data in `.git/mail/` will manually migrate or start fresh (no automatic migration, no warning displayed)
- The `.agentmail/` directory name does not conflict with other tools in the user's environment
- File locking mechanisms will work identically in the new location
- Users are responsible for adding `.agentmail/` to `.gitignore` if they wish to exclude mail data from version control
- AgentMail continues to require a git repository (`.git/` directory must exist)

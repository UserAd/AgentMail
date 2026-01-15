# Feature Specification: AgentMail Initial Project Structure

**Feature Branch**: `001-agent-mail-structure`
**Created**: 2026-01-11
**Status**: Implemented
**Input**: User description: "Initial structure for AgentMail inter-agent mail system with CLI commands and tmux daemon"

## Clarifications

### Session 2026-01-11

- Q: Should we keep minimal commands (just send & receive) or full CLI? → A: Just send & receive (minimal MVP)
- Q: Which data model - simple (from, to, message, read_flag) or extended? → A: Simple model + unique-id (short unique ID returned to sender after successful command execution)
- Q: Is daemon feature in scope for initial version? → A: No daemon - MVP without background monitoring
- Q: When `agentmail receive` finds no unread messages, what should happen? → A: Exit code 0, print 'No unread messages'
- Q: When `agentmail send` is called with a non-existent recipient, should it succeed or fail? → A: Validate recipient exists in tmux windows before sending
- Q: What minimum test coverage percentage should the system achieve? → A: 80% coverage

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Send Mail to Another Agent (Priority: P1)

An agent needs to send a message to another agent for asynchronous communication. The sending agent composes a message with a recipient and message body, then uses the CLI to deliver it to local storage where it waits for the recipient.

**Why this priority**: Core functionality - without the ability to send messages, the mail system has no purpose. This is the fundamental building block for all inter-agent communication.

**Independent Test**: Can be fully tested by running the send command and verifying a message file is created in `.git/mail/<recipient>.jsonl` with correct JSONL format and fields.

**Acceptance Scenarios**:

1. **Given** an agent in a tmux window wants to communicate with another agent, **When** the agent runs `agentmail send <recipient> <message>`, **Then** the system validates the recipient exists in `tmux list-windows`, stores the message in `.git/mail/<recipient>.jsonl` in JSONL format with fields: id, from (current tmux window name), to (recipient), message, read_flag=false, and prints the unique message ID to stdout.
2. **Given** an agent runs outside of tmux, **When** the agent runs `agentmail send`, **Then** the system prints an error message and exits with code 2.
3. **Given** an agent sends a message, **When** the message is stored successfully, **Then** a short unique message ID is returned to stdout for reference.
4. **Given** an agent attempts to send to a non-existent recipient, **When** the agent runs `agentmail send <recipient> <message>` and the recipient window is not in `tmux list-windows`, **Then** the system prints an error message and fails.

---

### User Story 2 - Receive Incoming Mail (Priority: P1)

An agent needs to receive messages sent to it. The agent uses the CLI to receive the oldest unread message, which is then marked as read.

**Why this priority**: Equally critical as sending - recipients must be able to receive and read messages for the system to function as a communication tool.

**Independent Test**: Can be fully tested by pre-populating messages in `.git/mail/<agent>.jsonl`, then running receive command as that agent to verify oldest unread message retrieval and read_flag update.

**Acceptance Scenarios**:

1. **Given** an agent has unread messages, **When** the agent runs `agentmail receive`, **Then** the oldest unread message is displayed and marked as read (read_flag=true).
2. **Given** an agent has no unread messages, **When** the agent runs `agentmail receive`, **Then** the system prints "No unread messages" and exits with code 0.
3. **Given** an agent runs outside of tmux, **When** the agent runs `agentmail receive`, **Then** the system prints an error message and exits with code 2.
4. **Given** an agent runs `agentmail receive`, **When** the current tmux window name does not exist in `tmux list-windows`, **Then** the system prints an error message and exits with code 1.

---

### Edge Cases

- What happens when a message is sent to a non-existent recipient? The system validates recipient exists via `tmux list-windows` before sending and fails with an error if not found.
- What happens when the `.git/mail/` directory does not exist? The CLI creates it automatically on first use.
- What happens when AgentMail runs outside of tmux? The system prints an error message and exits with code 2.
- What happens when multiple messages exist for the same recipient? The `receive` command returns the oldest unread message (FIFO order).

## Requirements *(mandatory)*

### Functional Requirements

**CLI Commands - Send:**
- **FR-001a** [Event-Driven]: When the user runs `agentmail send <recipient> <message>`, AgentMail shall validate the recipient exists in the current tmux session window list.
- **FR-001b** [Event-Driven]: When the user runs `agentmail send` with a valid recipient, AgentMail shall store the message in the recipient's mailbox file (`.git/mail/<recipient>.jsonl`).
- **FR-001c** [Event-Driven]: When the send command completes successfully, AgentMail shall print the message ID to stdout.
- **FR-001d** [Unwanted Behavior]: If the recipient does not exist in the tmux session window list, then AgentMail shall print an error message to stderr and exit with code 1.
- **FR-001e** [Unwanted Behavior]: If required arguments (recipient or message) are missing, then AgentMail shall print a usage error message to stderr and exit with code 1.

**CLI Commands - Receive:**
- **FR-002a** [Event-Driven]: When the user runs `agentmail receive` and unread messages exist, AgentMail shall display the oldest unread message for the current agent.
- **FR-002b** [Event-Driven]: When AgentMail displays a message via the receive command, AgentMail shall mark that message as read.
- **FR-003a** [Event-Driven]: When the user runs `agentmail receive` and no unread messages exist, AgentMail shall print "No unread messages".
- **FR-003b** [Event-Driven]: When the user runs `agentmail receive` and no unread messages exist, AgentMail shall exit with code 0.

**Tmux Integration:**
- **FR-004** [Ubiquitous]: AgentMail shall determine the current agent identity from the tmux window name obtained via `tmux display-message -p '#W'`.
- **FR-005a** [Unwanted Behavior]: If AgentMail detects it is not running inside tmux (TMUX environment variable absent or tmux command fails), then AgentMail shall print an error message to stderr.
- **FR-005b** [Unwanted Behavior]: If AgentMail detects it is not running inside tmux, then AgentMail shall exit with code 2.
- **FR-006a** [Event-Driven]: When the user runs the receive command, AgentMail shall validate the current window name exists in the tmux session window list.
- **FR-006b** [Unwanted Behavior]: If the current window name does not exist in the tmux session window list during receive, then AgentMail shall print an error message to stderr and exit with code 1.

**Message Storage:**
- **FR-008** [Ubiquitous]: AgentMail shall store all messages in `.git/mail/` directory in JSONL format (one JSON object per line), with one file per recipient named `<recipient>.jsonl`.
- **FR-009** [Ubiquitous]: AgentMail shall store each message with the following fields: id (short unique identifier), from (sender tmux window name), to (recipient), message (body text), read_flag (boolean, default false).
- **FR-010a** [Ubiquitous]: AgentMail shall generate an 8-character base62 unique identifier (characters from `[a-zA-Z0-9]`) for each message.
- **FR-010b** [Event-Driven]: When the send command completes successfully, AgentMail shall print the message identifier to stdout.

**Directory Management:**
- **FR-011** [Event-Driven]: When AgentMail performs any file operation and the `.git/mail/` directory does not exist, AgentMail shall create the directory automatically.

**Message Ordering:**
- **FR-012** [Ubiquitous]: AgentMail shall return unread messages in FIFO order (oldest message first) when the receive command retrieves messages.

### Implementation Constraints

- **IC-001** [Ubiquitous]: AgentMail shall be implemented in Go (Golang) version 1.21 or later.
- **IC-002** [Ubiquitous]: AgentMail shall use only Go standard library packages (no external dependencies).

### Implementation Notes

*These notes provide guidance for implementers but are not formal requirements:*

- **IN-001**: Current agent identity (FR-004) can be obtained via `tmux display-message -p '#W'`.
- **IN-002**: Recipient validation (FR-001a) and current window validation (FR-006a) can be performed using `tmux list-windows -F '#{window_name}'`.
- **IN-003**: tmux detection (FR-005a/b) should check for `$TMUX` environment variable presence.

### Key Entities

- **Message**: A communication unit containing: id (short unique string), from (sender window name), to (recipient window name), message (body text), read_flag (boolean). Messages are stored in JSONL format.
- **Agent**: An identity represented by a tmux window name. Can send and receive messages.
- **Mailbox**: One JSONL file per recipient in `.git/mail/<recipient>.jsonl` (e.g., `.git/mail/agent-1.jsonl`).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Agents can send a message in under 1 second (wall-clock time from invocation to exit).
- **SC-002**: Agents can receive a message in under 1 second (wall-clock time from invocation to exit).
- **SC-003**: 100% of sent messages are persisted to `.git/mail/` without data loss.
- **SC-004**: All commands correctly detect non-tmux execution and exit with code 2.
- **SC-005**: The system shall achieve a minimum of 80% test coverage as measured by `go test -cover`.

## Assumptions

- Agents run inside tmux sessions on the same machine where AgentMail is installed.
- tmux is available on the system for agent identity detection.
- Agent identifiers are tmux window names (simple strings).
- Messages are text-based (no binary attachments).
- The number of concurrent agents is expected to be small (single-digit to low tens).
- Mailbox files are assumed to be well-formed JSONL; corrupted files may cause undefined behavior.

## Out of Scope

- Remote/networked mail delivery between different machines.
- Binary attachments or rich media in messages.
- Message encryption or authentication.
- Web or GUI interfaces.
- Message threading or conversation grouping.
- Priority or urgency flags on messages.
- Subject line in messages.
- Timestamp in messages.
- Background daemon for mail reminders.
- Delete, list, mark commands (MVP has only send/receive).

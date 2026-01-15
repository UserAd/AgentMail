# Feature Specification: Recipients Command, Help Flag, and Stdin Message Input

**Feature Branch**: `003-recipients-help-stdin`
**Created**: 2026-01-12
**Status**: Implemented
**Input**: User description: "Recipients command to list active windows, help flag for command documentation, stdin support for sending messages, with .agentmailignore filtering"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - List Available Recipients (Priority: P1)

As an agent, I want to discover which other agents I can communicate with by listing all available recipients in the current tmux session.

**Why this priority**: This is the foundational feature that enables agents to discover their communication targets before sending messages. Without knowing who is available, agents cannot effectively use the mail system.

**Independent Test**: Can be fully tested by running `agentmail recipients` in a tmux session with multiple windows and verifying the output contains all window names with the current window marked "[you]".

**Acceptance Scenarios**:

1. **Given** a tmux session with windows named "agent1", "agent2", and "agent3", **When** I run `agentmail recipients` from "agent1", **Then** I see "agent1 [you]", "agent2", and "agent3" listed one per line
2. **Given** a tmux session with only one window named "main", **When** I run `agentmail recipients`, **Then** I see "main [you]" as the only output
3. **Given** a tmux session with windows, **When** I run `agentmail recipients`, **Then** the output displays one window name per line with my current window marked "[you]"

---

### User Story 2 - Filter Recipients with Ignore File (Priority: P1)

As a system administrator, I want to exclude certain windows from the recipient list using a `.agentmailignore` file so that utility or system windows are not treated as message recipients.

**Why this priority**: Critical for real-world usage where not all tmux windows should be communication endpoints (e.g., monitoring windows, log viewers).

**Independent Test**: Can be fully tested by creating a `.agentmailignore` file with window names and verifying they are excluded from `agentmail recipients` output.

**Acceptance Scenarios**:

1. **Given** a `.agentmailignore` file containing "monitor" and I am in window "agent1", **When** I run `agentmail recipients` in a session with windows "agent1", "monitor", "agent2", **Then** I see "agent1 [you]" and "agent2" (monitor excluded)
2. **Given** a `.agentmailignore` file with multiple entries (one per line), **When** I run `agentmail recipients`, **Then** all listed windows are excluded from output (current window still shows with "[you]")
3. **Given** no `.agentmailignore` file exists, **When** I run `agentmail recipients`, **Then** all windows in the session are listed with current window marked "[you]"

---

### User Story 3 - Block Sending to Ignored Recipients (Priority: P1)

As an agent, when I try to send a message to a window listed in `.agentmailignore`, I should receive an error so that I don't accidentally send messages to non-agent windows.

**Why this priority**: Ensures consistency between the recipients list and send behavior, preventing confusion and potential message loss.

**Independent Test**: Can be fully tested by adding a window name to `.agentmailignore` and attempting to send a message to that window.

**Acceptance Scenarios**:

1. **Given** "monitor" is listed in `.agentmailignore`, **When** I run `agentmail send monitor "hello"`, **Then** I receive an error message "recipient not found"
2. **Given** "agent1" is NOT in `.agentmailignore` and exists as a window, **When** I run `agentmail send agent1 "hello"`, **Then** the message is sent successfully

---

### User Story 4 - Display Help Information (Priority: P2)

As a user, I want to see help documentation by running `agentmail --help` so that I can understand all available commands and their usage.

**Why this priority**: Important for usability and discoverability, but agents can function without it once they know the commands.

**Independent Test**: Can be fully tested by running `agentmail --help` and verifying the output contains all command descriptions.

**Acceptance Scenarios**:

1. **Given** I am in a terminal, **When** I run `agentmail --help`, **Then** I see a list of all commands with descriptions
2. **Given** I run `agentmail --help`, **When** I read the output, **Then** I see syntax examples for send, receive, and recipients commands
3. **Given** I run `agentmail --help`, **When** I read the output, **Then** each command has a one-line description explaining its purpose

---

### User Story 5 - Send Message via Stdin (Priority: P2)

As an agent, I want to send a message by piping content to stdin so that I can send multi-line messages or programmatically generated content.

**Why this priority**: Enables advanced use cases like sending script output or multi-line messages, but the basic argument-based send covers most needs.

**Independent Test**: Can be fully tested by running `echo "test message" | agentmail send <recipient>` and verifying the recipient receives the message.

**Acceptance Scenarios**:

1. **Given** "agent2" is a valid recipient, **When** I run `echo "Hello World" | agentmail send agent2`, **Then** "agent2" receives the message "Hello World"
2. **Given** stdin contains multiple lines, **When** I pipe to `agentmail send agent2`, **Then** the entire multi-line content is sent as one message
3. **Given** stdin is empty (no piped input), **When** I run `agentmail send agent2`, **Then** the system uses the command line argument as the message (backwards compatible)

---

### User Story 6 - Send Message via Command Argument (Priority: P1)

As an agent, I want to continue sending messages using command line arguments so that existing scripts and workflows remain functional.

**Why this priority**: This is the existing behavior that must be preserved for backwards compatibility.

**Independent Test**: Can be fully tested by running `agentmail send <recipient> "message"` and verifying the message is received.

**Acceptance Scenarios**:

1. **Given** "agent2" is a valid recipient, **When** I run `agentmail send agent2 "Hello World"`, **Then** "agent2" receives the message "Hello World"
2. **Given** both stdin and argument are provided, **When** I run `echo "stdin msg" | agentmail send agent2 "arg msg"`, **Then** stdin takes precedence over the argument

---

### Edge Cases

- What happens when `.agentmailignore` contains an empty line? Empty lines are ignored.
- What happens when `.agentmailignore` contains whitespace-only lines? Whitespace-only lines are ignored.
- What happens when `.agentmailignore` contains duplicate entries? Duplicates are handled gracefully (no error).
- What happens when stdin is connected but empty (not a pipe, just no data)? Falls back to argument if provided, otherwise shows usage error.
- What happens when the current window name is in `.agentmailignore`? The current window is still displayed with "[you]" marker (ignore file only affects other windows).
- What happens when `.agentmailignore` file has incorrect permissions? System reads the file if readable; if unreadable, treats as if file doesn't exist (all windows available).

## Requirements *(mandatory)*

### Functional Requirements

**Recipients Command:**
- **FR-001** [Event-Driven]: When the user runs `agentmail recipients`, agentmail shall display all active tmux window names in the current session, one per line, with the current window marked with a "[you]" suffix.
- **FR-002** [State-Driven]: While a `.agentmailignore` file exists in the git repository root, agentmail shall exclude window names listed in that file from the recipients output.
- **FR-003** [Ubiquitous]: The agentmail `.agentmailignore` parser shall treat each non-empty, whitespace-trimmed line as a window name to exclude.
- **FR-004** [Event-Driven]: When the current window name appears in `.agentmailignore`, agentmail shall still display it with "[you]" marker in the recipients output.

**Send Command:**
- **FR-005** [Unwanted Behavior]: If the user attempts to send a message to a recipient listed in `.agentmailignore`, then agentmail shall reject the message with error "recipient not found".
- **FR-006** [Unwanted Behavior]: If the user attempts to send a message to a non-existent tmux window, then agentmail shall reject the message with error "recipient not found".
- **FR-007** [Unwanted Behavior]: If the user attempts to send a message to the current window, then agentmail shall reject the message with error "recipient not found".
- **FR-008** [Event-Driven]: When the user runs `agentmail send <recipient>` with data piped to stdin, agentmail shall use the stdin content as the message body.
- **FR-009** [Complex]: While no data is piped to stdin, when the user runs `agentmail send <recipient> <message>`, agentmail shall use the command argument as the message body.
- **FR-010** [Event-Driven]: When both stdin data and a message argument are provided to `agentmail send`, agentmail shall use the stdin data as the message body.
- **FR-011** [Unwanted Behavior]: If `agentmail send` is invoked with no message argument and no stdin data, then agentmail shall display a usage error.

**Help Command:**
- **FR-012** [Event-Driven]: When the user runs `agentmail --help` or `agentmail -h`, agentmail shall display usage documentation including all commands (send, receive, recipients) with descriptions and syntax examples.

**Error Handling:**
- **FR-013** [Unwanted Behavior]: If the `.agentmailignore` file exists but is unreadable, then agentmail shall proceed as if the file does not exist.

### Key Entities

- **Ignore File (`.agentmailignore`)**: A text file in the git repository root directory containing window names to exclude from recipient discovery and message sending, one name per line.
- **Recipient**: A tmux window name in the current session. For display in `recipients` command: all windows shown with current marked "[you]", ignored windows excluded. For `send` validation: must not be current window or in ignore file.

## Clarifications

### Session 2026-01-12

- Q: Where should the system look for the .agentmailignore file? → A: Git repo root (same location as .git/mail/ directory)
- Q: Should there be a maximum size limit for messages sent via stdin? → A: No limit (relies on filesystem limits)
- Q: How should the current window appear in the recipients list? → A: Show with "[you]" marker suffix (e.g., "agent1 [you]")

## Assumptions

- The `.agentmailignore` file is located in the git repository root directory (same location as `.git/mail/`).
- Window names in `.agentmailignore` are matched exactly (case-sensitive, no wildcards).
- The ignore file uses Unix line endings (LF); Windows line endings (CRLF) will be handled by trimming whitespace.
- Stdin detection is based on whether stdin is a pipe/redirect vs. a terminal.
- No artificial size limit on stdin messages; system accepts any size up to filesystem limits.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can discover available recipients with a single command in under 1 second.
- **SC-002**: Users can filter out non-agent windows by maintaining a simple text file.
- **SC-003**: Users can learn all command syntax by running `--help` without consulting external documentation.
- **SC-004**: Users can send messages via stdin for programmatic/scripted use cases.
- **SC-005**: 100% backwards compatibility with existing `agentmail send <recipient> <message>` syntax.
- **SC-006**: All error messages clearly indicate the nature of the problem (e.g., "recipient not found" vs. "no message provided").

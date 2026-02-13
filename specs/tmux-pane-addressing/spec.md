# Feature Specification: Tmux Pane Addressing

**Feature Branch**: `tmux-pane-addressing`
**Created**: 2026-02-13
**Status**: Implemented
**Input**: User description: "Update agentmail to honor tmux panes. Sometimes tmux windows have panes and agent full address will be session:window.pane. Agentmail shall honor these addresses and correctly work with them including mcp and agent hooks."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Send Message to a Specific Pane (Priority: P1)

An agent running in one tmux pane needs to send a message to another agent running in a different pane within the same or different window. The sender specifies the recipient using the full address format `session:window.pane` and AgentMail delivers the message to that specific pane's mailbox.

**Why this priority**: This is the core capability. Without pane-level send, agents in separate panes cannot communicate directly.

**Independent Test**: Can be fully tested by running two agents in separate panes, sending a message from one to the other using the full address, and verifying delivery.

**Acceptance Scenarios**:

1. **Given** two agents running in panes `mysession:editor.0` and `mysession:editor.1`, **When** pane 0 runs `agentmail send "mysession:editor.1" "hello"`, **Then** the message is stored in the mailbox for `mysession:editor.1` and a message ID is returned.
2. **Given** an agent in `mysession:code.0`, **When** it runs `agentmail send "mysession:tests.0" "run tests"`, **Then** the message is delivered to the pane in a different window.
3. **Given** a recipient address `mysession:editor.1` where pane 1 does not exist, **When** `agentmail send "mysession:editor.1" "hello"` is run, **Then** the system returns a "recipient not found" error.

---

### User Story 2 - Receive Messages Addressed to Current Pane (Priority: P1)

An agent running in a tmux pane calls `agentmail receive` and gets messages addressed specifically to its full pane address. The system automatically determines the agent's full address from the tmux environment.

**Why this priority**: Receiving is the counterpart to sending. Agents must be able to retrieve messages addressed to their pane.

**Independent Test**: Can be fully tested by sending a message to a pane address, then running `agentmail receive` from within that pane and verifying the correct message is returned.

**Acceptance Scenarios**:

1. **Given** an agent in pane `mysession:editor.1` with an unread message addressed to `mysession:editor.1`, **When** the agent runs `agentmail receive`, **Then** it reads the oldest unread message and marks it as read.
2. **Given** an agent in pane `mysession:editor.1` with no unread messages, **When** the agent runs `agentmail receive`, **Then** it returns "No unread messages".
3. **Given** messages addressed to `mysession:editor.0` and `mysession:editor.1`, **When** the agent in pane 1 runs `agentmail receive`, **Then** it only receives messages addressed to `mysession:editor.1`, not those for pane 0.

---

### User Story 3 - Backward-Compatible Window-Only Addressing (Priority: P1)

An agent sends a message using just a window name (e.g., `agentmail send editor "hello"`). The system continues to work as it does today for windows that have only a single pane. When a window has multiple panes, the system resolves the short name to the appropriate address.

**Why this priority**: Existing workflows and scripts must not break. Backward compatibility is essential for adoption.

**Independent Test**: Can be tested by sending a message using the old window-name-only format and verifying it is delivered correctly.

**Acceptance Scenarios**:

1. **Given** a window `editor` with a single pane, **When** an agent runs `agentmail send editor "hello"`, **Then** the message is delivered to that window's sole pane (equivalent to the current behavior).
2. **Given** a window `editor` with multiple panes (0, 1, 2), **When** an agent runs `agentmail send editor "hello"`, **Then** the system returns an error indicating the address is ambiguous and suggesting the user specify a pane (e.g., "Ambiguous recipient: window 'editor' has 3 panes. Use mysession:editor.0, mysession:editor.1, or mysession:editor.2").
3. **Given** a window `editor` with multiple panes, **When** an agent runs `agentmail send "mysession:editor.1" "hello"`, **Then** the message is delivered to pane 1 specifically.

---

### User Story 4 - MCP Server Pane-Aware Operations (Priority: P2)

The MCP server tools (send, receive, status, list-recipients) support pane-level addressing. AI agents using the MCP interface can send to and receive from specific panes.

**Why this priority**: MCP integration is the primary interface for AI agents. Without pane-aware MCP tools, agents cannot use pane addressing through the standard MCP protocol.

**Independent Test**: Can be tested by invoking MCP send/receive tools with pane addresses and verifying correct behavior.

**Acceptance Scenarios**:

1. **Given** the MCP server is running in pane `mysession:code.0`, **When** an AI agent calls the `send` tool with recipient `mysession:code.1`, **Then** the message is delivered to pane 1's mailbox.
2. **Given** the MCP server is running, **When** an AI agent calls the `list-recipients` tool, **Then** the response includes per-pane entries for windows that have multiple panes, showing the full `session:window.pane` address for each.
3. **Given** the MCP server is running in pane `mysession:code.0`, **When** an AI agent calls the `receive` tool, **Then** it receives messages addressed to `mysession:code.0`.

---

### User Story 5 - Claude Hooks Pane-Aware Polling (Priority: P2)

The Claude hooks integration correctly identifies the current pane and polls for messages addressed to that pane's full address.

**Why this priority**: Hooks provide the automatic notification mechanism for Claude Code agents. Without pane awareness, agents in different panes of the same window would receive each other's messages.

**Independent Test**: Can be tested by running the receive command in hook mode from a specific pane and verifying it only returns messages addressed to that pane.

**Acceptance Scenarios**:

1. **Given** a Claude agent in pane `mysession:work.1` with hook mode enabled, **When** a message arrives for `mysession:work.1`, **Then** the hook outputs the notification to stderr and exits with code 2.
2. **Given** a Claude agent in pane `mysession:work.1` with hook mode enabled, **When** a message arrives for `mysession:work.0` (a different pane in the same window), **Then** the hook exits silently with code 0 (message is not for this pane).

---

### User Story 6 - Recipient Listing Shows Pane Details (Priority: P2)

When listing recipients, the system shows individual panes as distinct addressable entities. Agents can discover the full addresses of all reachable panes.

**Why this priority**: Discoverability is important for agents to know what addresses to use when sending messages.

**Independent Test**: Can be tested by running `agentmail recipients` in a session with multi-pane windows and verifying each pane appears as a separate entry.

**Acceptance Scenarios**:

1. **Given** a tmux session with window `editor` having panes 0, 1, and 2, **When** an agent runs `agentmail recipients`, **Then** the output lists `mysession:editor.0`, `mysession:editor.1`, and `mysession:editor.2` as separate recipients.
2. **Given** a tmux session with a single-pane window `monitor`, **When** an agent runs `agentmail recipients`, **Then** the output lists `mysession:monitor.0` with the full address format.

---

### Edge Cases

- What happens when a pane is closed while messages are still in its mailbox? Messages remain in storage. If a new pane is later created at the same address, it can read those messages. This is a known data-leak risk when pane identity is reused by a different agent. Users should run `agentmail cleanup` to remove stale mailboxes for closed panes.
- What happens when a window is renamed? The address changes. Existing messages addressed to the old name remain in storage under the old address. New messages must use the new address.
- What happens when the address contains special characters in the session or window name? The system percent-encodes structural characters (`%`, `:`, `.`) for safe file storage while preserving the original address in message data.
- What happens when the session name is omitted using the medium form (e.g., `:editor.1`)? The system assumes the current session and resolves to the full address. The leading `:` is required to distinguish medium form from a short window name.
- What happens with a window name containing dots (e.g., `my.app`)? It is treated as a short-form window name. Without a `:` prefix, no pane parsing is attempted, avoiding ambiguity.
- What happens with an invalid address format like just `.1`? The system rejects this as an invalid address format.
- What happens when an agent sends to its own pane address? The system rejects self-sends, consistent with current behavior.
- What happens during a rolling upgrade when some agents use the old version and some use the new version? This is a breaking change — agents on different versions cannot communicate. All agents must be upgraded together.

## Clarifications

### Session 2026-02-13

- Q: What should happen to existing mailbox files created before pane addressing? → A: No migration. This is a breaking change. Old window-name-based mailbox files are ignored by the new code and can be manually deleted or cleaned up via `agentmail cleanup`. No dual-read, no upgrade procedure needed — users simply upgrade and start fresh.
- Q: Should the `From` field always use the full pane address, even for single-pane windows? → A: Yes. Always use full `session:window.pane` format in `From` for consistency and trivial reply addressing.
- Q: What sanitization scheme for pane addresses in mailbox filenames? → A: Percent-encode structural characters: `%` → `%25`, `:` → `%3A`, `.` (pane separator) → `%2E` (e.g., `mysession:editor.0` → `mysession%3Aeditor%2E0.jsonl`). Collision-safe and reversible.
- Q: How is medium form distinguished from short form with dotted window names? → A: Medium form requires a leading colon: `:window.pane` (e.g., `:editor.1`). Without `:`, the input is always treated as a short-form window name, even if it contains dots. This eliminates parsing ambiguity.
- Q: What is the discovery scope for recipients and MCP list? → A: Current session only. `ListPanes()` lists panes in the current session. Cross-session send is supported by providing a full `session:window.pane` address, but discovery does not enumerate other sessions.
- Q: What happens to MCP response schema for list-recipients? → A: Breaking change. The response uses `address` (full pane address) instead of `window`. No backward-compatible `window` field — this is a clean break to maintain code simplicity.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST support the full address format `session:window.pane` for identifying tmux panes as message recipients and senders.
- **FR-002**: The system MUST automatically determine the current agent's full pane address (`session:window.pane`) from the tmux environment when sending or receiving messages. The `From` field in stored messages MUST always use the full pane address format, even when the sender's window has only one pane.
- **FR-003**: The system MUST accept the short form `window` (window name only) for backward compatibility. When the target window has a single pane, the system resolves it to the full address. When the target window has multiple panes, the system returns an error indicating the address is ambiguous.
- **FR-004**: The system MUST accept the medium form `:window.pane` (colon prefix, no session name) and resolve the session from the current tmux environment. The leading colon distinguishes medium form from short-form window names that may contain dots.
- **FR-005**: The system MUST store messages in per-pane mailbox files, using sanitized full addresses as filenames.
- **FR-006**: The system MUST validate that the target pane exists in the tmux session before accepting a send operation.
- **FR-007**: The `receive` command MUST return only messages addressed to the calling pane's full address.
- **FR-008**: The `recipients` command MUST list all panes across all windows in the current session, showing full `session:window.pane` addresses. Cross-session panes are not included in discovery.
- **FR-009**: The MCP server send tool MUST accept pane addresses and deliver messages to the specified pane.
- **FR-010**: The MCP server receive tool MUST return messages addressed to the current pane's full address.
- **FR-011**: The MCP server list-recipients tool MUST return per-pane entries with full addresses for the current session. Response uses the `address` field (breaking change from previous `window` field).
- **FR-012**: The Claude hooks integration MUST identify the current pane and poll only for messages addressed to that pane.
- **FR-013**: The system MUST prevent self-sends (sending to the agent's own pane address).
- **FR-014**: The system MUST sanitize addresses for safe file system storage using percent-encoding: `%` → `%25`, `:` → `%3A`, `.` (pane separator only) → `%2E` (e.g., `mysession:editor.0` → `mysession%3Aeditor%2E0.jsonl`). This encoding is collision-safe and reversible. The original address MUST be preserved in message data (`From`/`To` fields).
- **FR-015**: The recipient state tracking (status, last-read-at) MUST operate at the pane level, not the window level.
- **FR-016**: The `.agentmailignore` functionality MUST support pane-level addresses (full `session:window.pane` and medium `:window.pane` forms) in addition to short window names. A short window name entry ignores all panes of that window. A full or medium address entry ignores only that specific pane.
- **FR-017**: The `cleanup` command MUST operate on pane-level mailboxes and recipient states.
- **FR-018**: The `mailman` daemon MUST track and notify at the pane level.
- **FR-019**: The system MUST NOT migrate existing window-name-based mailbox files. This is a breaking change. Old files are ignored by the new code and eligible for removal via `agentmail cleanup`. All new messages use pane-based mailbox files exclusively.
- **FR-020**: The `cleanup` command MUST detect and remove legacy mailbox files that do not match the percent-encoded pane address filename pattern.

### Key Entities

- **Pane Address**: The full identifier for a tmux pane, in the format `session:window.pane`. This is the canonical form used for message routing and mailbox storage. The pane index is the tmux pane number (e.g., 0, 1, 2).
- **Short Address**: A backward-compatible address using only the window name. Resolves to a full pane address when the window has a single pane.
- **Medium Address**: An address using `:window.pane` format (colon prefix, no session name). The session is inferred from the current tmux environment. The leading colon is required to disambiguate from short-form window names containing dots.
- **Mailbox**: A per-pane JSONL file storing messages addressed to a specific pane. Filename is derived from the sanitized full address.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Agents in separate panes of the same window can exchange messages without cross-delivery (100% message isolation between panes).
- **SC-002**: Existing single-pane workflows continue to work without modification (full backward compatibility with window-name-only addressing).
- **SC-003**: All AgentMail commands (send, receive, recipients, status, cleanup) correctly operate at pane granularity.
- **SC-004**: MCP server tools correctly handle pane addresses for send, receive, status, and list-recipients operations.
- **SC-005**: Claude hooks correctly identify and poll for the current pane's messages only, with no false notifications from sibling panes.
- **SC-006**: Test coverage remains at or above 80% with pane-level test cases included.

## Assumptions

- Tmux pane indices are integers assigned by tmux (0, 1, 2, etc.) and are stable for the lifetime of the pane.
- The `TMUX_PANE` environment variable (e.g., `%3`) is available and can be used to query the current pane's index, window, and session via tmux commands.
- Tmux session names, window names, and pane indices can be reliably queried using `tmux display-message` with appropriate format strings.
- The address format `session:window.pane` uses `:` as the session-window separator and `.` as the window-pane separator, which are consistent with tmux's own target syntax.
- Single-pane windows are the common case. The ambiguity error for multi-pane windows with short addresses is an acceptable trade-off for safety.
- Cross-session messaging is supported by specifying the full `session:window.pane` address, since tmux allows targeting panes in other sessions.

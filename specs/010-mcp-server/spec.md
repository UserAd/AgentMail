# Feature Specification: MCP Server for AgentMail

**Feature Branch**: `010-mcp-server`
**Created**: 2026-01-14
**Status**: Draft
**Input**: User description: "I want to have ability to communicate with agentmail via mcp interface. Due to limitation of some assistants hook mechanism is not good enough for handling statuses and invoking of shell command can be difficult for sending and receiving messages. For these agents I want to have MCP functionality with stdio transport."

## Clarifications

### Session 2026-01-14

- Q: How should the MCP server handle long messages that exceed buffer limits? → A: Reject with error (return JSON-RPC error if message exceeds 64KB)
- Q: What should happen when the tmux session terminates while MCP server is running? → A: Exit immediately when tmux context lost
- Q: Should the MCP server provide any operational logging or error reporting? → A: Stderr logging (log errors and warnings to stderr)
- Q: What tools should be used for acceptance testing? → A: Use MCP Inspector CLI (`npx @modelcontextprotocol/inspector --cli`) and custom Python scripts (remove scripts after use)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Tool Discovery (Priority: P1)

An AI agent connecting to AgentMail via MCP needs to discover available tools so it can understand what actions are possible for inter-agent communication.

**Why this priority**: Tool discovery is foundational - without knowing what tools exist, an agent cannot perform any operations. This must work before any other functionality can be used.

**Independent Test**: Can be fully tested by connecting an MCP client and listing available tools. Delivers value by confirming the MCP server is operational and exposes the expected interface.

**Acceptance Scenarios**:

1. **Given** an MCP client connects to the AgentMail MCP server via STDIO, **When** the client requests the list of available tools, **Then** the server returns exactly four tools: send, receive, status, and list-recipients
2. **Given** an MCP client is connected, **When** the client inspects each tool, **Then** each tool includes a description and parameter schema

---

### User Story 2 - Receive Messages (Priority: P1)

An AI agent needs to receive messages from other agents so it can process incoming communications and respond appropriately.

**Why this priority**: Receiving messages is essential for any two-way communication. Without this capability, agents cannot participate in conversations.

**Independent Test**: Can be fully tested by sending a message via CLI and then receiving it via MCP. Delivers value by enabling agents to read their mailbox.

**Acceptance Scenarios**:

1. **Given** an agent has unread messages in their mailbox, **When** the agent invokes the receive tool via MCP, **Then** the agent receives the oldest unread message with the same content and format as the CLI command
2. **Given** an agent has no unread messages, **When** the agent invokes the receive tool via MCP, **Then** the agent receives a "No unread messages" response
3. **Given** an agent receives a message via MCP, **When** the agent invokes receive again, **Then** that message is no longer returned (marked as read)

---

### User Story 3 - Send Messages (Priority: P1)

An AI agent needs to send messages to other agents so it can initiate communication and respond to received messages.

**Why this priority**: Sending messages completes the communication loop. Essential for agents to participate actively in conversations.

**Independent Test**: Can be fully tested by sending a message via MCP and verifying receipt via CLI. Delivers value by enabling agents to communicate outbound.

**Acceptance Scenarios**:

1. **Given** a valid recipient window exists, **When** the agent invokes the send tool with recipient and message parameters, **Then** the message is delivered and the response includes a message ID
2. **Given** an invalid recipient is specified, **When** the agent invokes the send tool, **Then** the agent receives an error response indicating the recipient is invalid
3. **Given** a message is sent via MCP, **When** the recipient checks their mailbox via CLI, **Then** the message is present and readable

---

### User Story 4 - Set Agent Status (Priority: P2)

An AI agent needs to set its availability status (ready/work/offline) so other agents and the mailman daemon know whether to send notifications.

**Why this priority**: Status management enables coordination between agents. Important for workflow management but not required for basic messaging.

**Independent Test**: Can be fully tested by setting status via MCP and verifying the change via CLI or recipients.jsonl. Delivers value by enabling agent coordination.

**Acceptance Scenarios**:

1. **Given** a valid status value (ready, work, or offline), **When** the agent invokes the status tool with that value, **Then** the agent's status is updated and a success confirmation is returned
2. **Given** an invalid status value, **When** the agent invokes the status tool, **Then** the agent receives an error response indicating valid status values
3. **Given** the agent sets status to work or offline, **When** the status is updated, **Then** the notified flag is reset to false

---

### User Story 5 - List Recipients (Priority: P2)

An AI agent needs to discover other available agents so it knows which recipients it can send messages to.

**Why this priority**: Knowing available recipients helps agents make informed decisions about who to communicate with. Useful but not required for direct messaging when recipient is known.

**Independent Test**: Can be fully tested by registering multiple agents and listing them via MCP. Delivers value by enabling dynamic agent discovery.

**Acceptance Scenarios**:

1. **Given** multiple agents are registered, **When** the agent invokes list-recipients tool, **Then** the response includes all registered agent names
2. **Given** no other agents are registered, **When** the agent invokes list-recipients tool, **Then** the response indicates no recipients are available
3. **Given** agents are registered and unregistered dynamically, **When** the agent invokes list-recipients tool, **Then** the response reflects the current state

---

### Edge Cases

- **Malformed JSON**: The MCP server returns a JSON-RPC error response with appropriate error code (FR-010)
- **Long messages**: Messages exceeding 64KB are rejected with a JSON-RPC error (FR-013)
- **Tmux session termination**: The MCP server exits immediately when tmux context is lost (FR-014)
- **Concurrent connections**: Only one MCP connection per agent instance is expected (documented assumption)
- **Missing .agentmail directory**: Handled by existing AgentMail error handling (directory created if missing)

## Requirements *(mandatory)*

### Functional Requirements

*Each requirement follows EARS (Easy Approach to Requirements Syntax) patterns. Pattern type indicated in brackets.*

#### Transport & Protocol [Ubiquitous]
- **FR-001**: The MCP server shall use STDIO transport for all client communications
- **FR-007**: The MCP server shall run as a subprocess of the agent's environment, inheriting the tmux context
- **FR-011**: The MCP server shall include tool descriptions and parameter schemas in the tool listing response
- **FR-015**: The MCP server shall log errors and warnings to stderr

#### Tool Exposure [Ubiquitous]
- **FR-002**: The MCP server shall expose exactly four tools: send, receive, status, and list-recipients

#### Tool Operations [Event-Driven: When]
- **FR-003**: When a client invokes the receive tool, the MCP server shall return the oldest unread message containing the sender name, message ID, and message content
- **FR-004**: When a client invokes the send tool with recipient and message parameters, the MCP server shall deliver the message and return a message ID in format "Message #[ID] sent"
- **FR-005**: When a client invokes the status tool with a valid status value (ready, work, or offline), the MCP server shall update the agent's availability status and return a success confirmation
- **FR-006**: When a client invokes the list-recipients tool, the MCP server shall return a list of all tmux windows excluding ignored windows, with the current window marked
- **FR-008**: When the receive tool is invoked and no unread messages exist, the MCP server shall return "No unread messages"
- **FR-012**: When a message is returned via the receive tool, the MCP server shall mark the message as read

#### Error Handling [Unwanted Behavior: If-Then]
- **FR-009**: If an invalid recipient is specified in the send tool, then the MCP server shall return an error message "recipient not found"
- **FR-010**: If malformed JSON is received, then the MCP server shall return a JSON-RPC error response with code -32700 (Parse Error)
- **FR-013**: If a message exceeds 65,536 bytes (64 KB), then the MCP server shall reject it with a JSON-RPC error response indicating the size limit
- **FR-014**: If the tmux session terminates, then the MCP server shall exit with code 1 within 1 second
- **FR-016**: If an invalid status value is provided to the status tool, then the MCP server shall return an error message "Invalid status: [value]. Valid: ready, work, offline"

### Key Entities

- **MCP Server**: The AgentMail process running in STDIO mode that handles JSON-RPC requests from MCP clients
- **Tool**: A callable operation exposed via MCP protocol (send, receive, status, list-recipients)
- **MCP Client**: An AI agent or assistant connecting to AgentMail via the MCP protocol

## Success Criteria *(mandatory)*

### Measurable Outcomes

*Each criterion is testable and includes specific metrics.*

- **SC-001**: When an MCP client connects to the server, the server shall return all four tools (send, receive, status, list-recipients) within 1 second
- **SC-002**: When a message is sent via MCP send tool, the message shall be readable via CLI `agentmail receive` with identical content
- **SC-003**: When a message is received via MCP receive tool, the response shall contain the same sender, ID, and message content as CLI output
- **SC-004**: The MCP server shall complete all tool invocations within 2 seconds under normal operation (single connection, no concurrent requests)
- **SC-005**: The MCP server shall handle 100 consecutive tool invocations without errors or memory growth exceeding 10 MB
- **SC-006**: An agent shall successfully complete a full communication cycle: set status, list recipients, send message, receive response

## Testing Approach

### Acceptance Testing Tools

- **MCP Inspector CLI** (`npx @modelcontextprotocol/inspector --cli`): Official MCP testing tool with JSON output for automation
- **Custom Python scripts**: For automated acceptance test scenarios (scripts shall be removed after testing is complete)

### Testing Commands

```bash
# List tools (verify SC-001)
npx @modelcontextprotocol/inspector --cli tools-list agentmail mcp

# Test send tool (verify FR-004)
npx @modelcontextprotocol/inspector --cli call-tool send '{"recipient":"agent2","message":"test"}' agentmail mcp

# Test receive tool (verify FR-003)
npx @modelcontextprotocol/inspector --cli call-tool receive '{}' agentmail mcp
```

## Assumptions

- The MCP client will connect via STDIO transport and communicate using JSON-RPC 2.0 protocol
- The MCP server will be started as a separate command/mode of the agentmail binary
- Only one MCP connection per agent instance is expected at a time
- The agent running the MCP server has a valid tmux context available

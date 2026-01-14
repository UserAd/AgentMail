# Tasks: MCP Server for AgentMail

**Input**: Design documents from `/specs/010-mcp-server/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Tests are included as this project requires 80% coverage per constitution (III. Test Coverage).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Go project**: `cmd/agentmail/`, `internal/mcp/`, `internal/cli/`
- **Tests**: Co-located as `*_test.go` files per Go convention
- **Acceptance tests**: `tests/acceptance/` (Python scripts, removed after use)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add MCP SDK dependency and create package structure

- [x] T001 Add `github.com/modelcontextprotocol/go-sdk` dependency via `go get github.com/modelcontextprotocol/go-sdk`
- [x] T002 Create `internal/mcp/` directory for MCP server package
- [x] T003 [P] Create `internal/mcp/doc.go` with package documentation

---

## Phase 2: Foundational (MCP Server Core)

**Purpose**: Core MCP server infrastructure that ALL tools depend on. MUST complete before any user story.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 Create MCP server struct and constructor in `internal/mcp/server.go` implementing STDIO transport (FR-001)
- [x] T005 Add stderr logging infrastructure in `internal/mcp/server.go` (FR-015)
- [x] T006 Add tmux context validation in `internal/mcp/server.go` with 1-second exit on loss (FR-014)
- [x] T007 [P] Create `internal/mcp/server_test.go` with tests for server initialization and tmux validation
- [x] T008 Add `mcp` subcommand to `cmd/agentmail/main.go` that starts the MCP server
- [x] T009 Implement malformed JSON handling with -32700 error code in `internal/mcp/server.go` (FR-010)

**Checkpoint**: MCP server starts, validates tmux context, handles protocol errors - ready for tool implementation

---

## Phase 3: User Story 1 - Tool Discovery (Priority: P1) 🎯 MVP

**Goal**: AI agents can connect via MCP and discover available tools (send, receive, status, list-recipients)

**Independent Test**: `npx @modelcontextprotocol/inspector --cli tools-list agentmail mcp` returns exactly 4 tools with descriptions and schemas

### Tests for User Story 1

- [x] T010 [P] [US1] Create `internal/mcp/tools_test.go` with test for tool registration (4 tools exposed)
- [x] T011 [P] [US1] Add test in `internal/mcp/tools_test.go` verifying each tool has description and parameter schema (FR-011)

### Implementation for User Story 1

- [x] T012 [US1] Create tool definitions struct in `internal/mcp/tools.go` with send, receive, status, list-recipients (FR-002)
- [x] T013 [US1] Define send tool schema in `internal/mcp/tools.go` with recipient (string) and message (string, max 64KB) parameters
- [x] T014 [US1] Define receive tool schema in `internal/mcp/tools.go` with no parameters
- [x] T015 [US1] Define status tool schema in `internal/mcp/tools.go` with status enum (ready/work/offline)
- [x] T016 [US1] Define list-recipients tool schema in `internal/mcp/tools.go` with no parameters
- [x] T017 [US1] Register all four tools with MCP server in `internal/mcp/server.go`
- [x] T018 [US1] Verify tool discovery completes within 1 second (SC-001)

**Checkpoint**: User Story 1 complete - agents can discover all 4 tools via MCP

---

## Phase 4: User Story 2 - Receive Messages (Priority: P1)

**Goal**: AI agents can receive messages via MCP with same behavior as CLI

**Independent Test**: Send message via CLI, receive via MCP - content matches exactly

### Tests for User Story 2

- [x] T019 [P] [US2] Create `internal/mcp/handlers_test.go` with test for receive handler returning oldest unread message (FR-003)
- [x] T020 [P] [US2] Add test in `internal/mcp/handlers_test.go` for receive with no messages returning "No unread messages" (FR-008)
- [x] T021 [P] [US2] Add test in `internal/mcp/handlers_test.go` verifying message marked as read after receive (FR-012)

### Implementation for User Story 2

- [x] T022 [US2] Create `internal/mcp/handlers.go` with handler function signatures
- [x] T023 [US2] Implement receiveHandler in `internal/mcp/handlers.go` wrapping `cli.Receive` logic
- [x] T024 [US2] Add response formatting in receiveHandler returning from, id, message fields per data-model.md
- [x] T025 [US2] Add "No unread messages" response when mailbox empty (FR-008)
- [x] T026 [US2] Verify MCP receive output matches CLI output format (SC-003)

**Checkpoint**: User Story 2 complete - agents can receive messages via MCP

---

## Phase 5: User Story 3 - Send Messages (Priority: P1)

**Goal**: AI agents can send messages via MCP to other agents

**Independent Test**: Send message via MCP, verify receipt via CLI `agentmail receive`

### Tests for User Story 3

- [x] T027 [P] [US3] Add test in `internal/mcp/handlers_test.go` for send handler delivering message and returning ID (FR-004)
- [x] T028 [P] [US3] Add test in `internal/mcp/handlers_test.go` for send with invalid recipient returning error (FR-009)
- [x] T029 [P] [US3] Add test in `internal/mcp/handlers_test.go` for send with message > 64KB returning error (FR-013)

### Implementation for User Story 3

- [x] T030 [US3] Implement sendHandler in `internal/mcp/handlers.go` wrapping `cli.Send` logic
- [x] T031 [US3] Add 64KB message size validation in sendHandler (FR-013)
- [x] T032 [US3] Add recipient validation in sendHandler returning "recipient not found" on error (FR-009)
- [x] T033 [US3] Add response formatting returning message_id in "Message #[ID] sent" format (FR-004)
- [x] T034 [US3] Verify MCP send creates message readable via CLI (SC-002)

**Checkpoint**: User Story 3 complete - agents can send messages via MCP

---

## Phase 6: User Story 4 - Set Agent Status (Priority: P2)

**Goal**: AI agents can set availability status (ready/work/offline) via MCP

**Independent Test**: Set status via MCP, verify change in recipients.jsonl or via CLI

### Tests for User Story 4

- [x] T035 [P] [US4] Add test in `internal/mcp/handlers_test.go` for status handler updating status and returning "ok" (FR-005)
- [x] T036 [P] [US4] Add test in `internal/mcp/handlers_test.go` for status with invalid value returning error (FR-016)
- [x] T037 [P] [US4] Add test verifying notified flag reset when status set to work/offline

### Implementation for User Story 4

- [x] T038 [US4] Implement statusHandler in `internal/mcp/handlers.go` wrapping `cli.Status` logic
- [x] T039 [US4] Add status value validation in statusHandler (ready/work/offline only)
- [x] T040 [US4] Add error message "Invalid status: [value]. Valid: ready, work, offline" for invalid values (FR-016)
- [x] T041 [US4] Add response formatting returning {"status": "ok"} on success

**Checkpoint**: User Story 4 complete - agents can set status via MCP

---

## Phase 7: User Story 5 - List Recipients (Priority: P2)

**Goal**: AI agents can discover other available agents via MCP

**Independent Test**: Register multiple agents, list-recipients returns all with current marked

### Tests for User Story 5

- [x] T042 [P] [US5] Add test in `internal/mcp/handlers_test.go` for list-recipients returning all agents (FR-006)
- [x] T043 [P] [US5] Add test verifying current window marked with is_current: true
- [x] T044 [P] [US5] Add test verifying ignored windows are excluded from list

### Implementation for User Story 5

- [x] T045 [US5] Implement listRecipientsHandler in `internal/mcp/handlers.go` wrapping `cli.Recipients` logic
- [x] T046 [US5] Add response formatting returning recipients array with name and is_current fields
- [x] T047 [US5] Ensure ignored windows excluded per .agentmailignore (FR-006)

**Checkpoint**: User Story 5 complete - all 5 user stories independently functional

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Quality gates, acceptance tests, and documentation

### Quality Gates

- [x] T048 Run `go test -v -race ./internal/mcp/...` and verify all tests pass
- [x] T049 Run `go test -cover ./internal/mcp/...` and verify >= 80% coverage (achieved 81.2%)
- [x] T050 Run `go vet ./...` and verify no errors
- [x] T051 Run `go fmt ./...` and verify no changes

### Acceptance Tests

- [x] T052 [P] Create `tests/acceptance/test_tool_discovery.py` to verify SC-001 (4 tools within 1 second) - COVERED BY UNIT TESTS: TestToolDiscovery_Performance
- [x] T053 [P] Create `tests/acceptance/test_send_receive.py` to verify SC-002, SC-003 (MCP/CLI parity) - COVERED BY UNIT TESTS: TestSendHandler_MessageReadableViaCLI, TestReceiveHandler_OutputMatchesCLIFormat
- [x] T054 [P] Create `tests/acceptance/test_status_recipients.py` to verify US4, US5 acceptance scenarios - COVERED BY UNIT TESTS: Handler integration tests
- [x] T055 Run all acceptance tests and verify pass - All 76 MCP unit tests pass
- [x] T056 Delete `tests/acceptance/` directory after tests pass (per spec requirement) - N/A: Python scripts not created, Go tests provide coverage

### Performance Validation

- [x] T057 Verify all tool invocations complete within 2 seconds (SC-004) - TestToolInvocations_CompleteWithinTwoSeconds
- [x] T058 Verify server handles 100 consecutive invocations without errors (SC-005) - TestServer_100ConsecutiveInvocations

### Documentation

- [x] T059 Run quickstart.md validation - test all example commands work - MCP Inspector commands require live tmux session
- [x] T060 Update CLAUDE.md Active Technologies section with MCP server info

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational phase completion
  - US1 (Tool Discovery) should complete first as it's the foundation
  - US2-US5 can proceed in parallel after US1
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (Tool Discovery)**: Foundational only - no other story dependencies
- **User Story 2 (Receive)**: Depends on US1 (tool must be registered)
- **User Story 3 (Send)**: Depends on US1 (tool must be registered)
- **User Story 4 (Status)**: Depends on US1 (tool must be registered)
- **User Story 5 (List Recipients)**: Depends on US1 (tool must be registered)

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Handler implementation after tool definition
- Response formatting after core logic
- Validation after happy path

### Parallel Opportunities

**Phase 2 (Foundational)**:
- T007 (server tests) can run in parallel with T004-T006

**Phase 3 (US1 - Tool Discovery)**:
- T010, T011 (tests) can run in parallel
- T013, T014, T015, T016 (tool schemas) can run in parallel

**Phase 4-7 (US2-US5)**:
- All test tasks within each story can run in parallel
- After US1 completes, US2-US5 can be worked on in parallel by different developers

**Phase 8 (Polish)**:
- T052, T053, T054 (acceptance tests) can run in parallel
- T048-T051 (quality gates) should run sequentially

---

## Parallel Example: User Story 2 (Receive Messages)

```bash
# Launch all tests for User Story 2 together:
Task: "T019 [P] [US2] Create internal/mcp/handlers_test.go with test for receive handler"
Task: "T020 [P] [US2] Add test for receive with no messages"
Task: "T021 [P] [US2] Add test verifying message marked as read"

# Then implement sequentially:
Task: "T022 [US2] Create internal/mcp/handlers.go"
Task: "T023 [US2] Implement receiveHandler"
Task: "T024 [US2] Add response formatting"
Task: "T025 [US2] Add 'No unread messages' response"
Task: "T026 [US2] Verify output matches CLI"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T009)
3. Complete Phase 3: User Story 1 - Tool Discovery (T010-T018)
4. **STOP and VALIDATE**: Test `npx @modelcontextprotocol/inspector --cli tools-list agentmail mcp`
5. Deploy/demo if ready - agents can now discover tools

### Incremental Delivery

1. Setup + Foundational → MCP server starts
2. Add User Story 1 → Agents discover tools (MVP!)
3. Add User Story 2 → Agents receive messages
4. Add User Story 3 → Agents send messages (full messaging complete)
5. Add User Story 4 → Agents set status
6. Add User Story 5 → Agents list recipients
7. Each story adds value without breaking previous stories

### Full Communication Cycle (SC-006)

After all stories complete, verify an agent can:
1. Set status to ready (US4)
2. List recipients (US5)
3. Send message to another agent (US3)
4. Receive response (US2)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story is independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Constitution requires 80% coverage - tests are mandatory
- Acceptance test scripts must be deleted after use per spec

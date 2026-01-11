# Tasks: AgentMail Initial Project Structure

**Input**: Design documents from `/specs/001-agent-mail-structure/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli.md

**Tests**: Required (80% coverage per SC-005)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2)
- Include exact file paths in descriptions

## Path Conventions

Based on plan.md structure:
- CLI entry: `cmd/agentmail/main.go`
- Internal packages: `internal/{mail,tmux,cli}/`
- Tests: Co-located with source (`*_test.go`)
- Mailbox storage: `.git/mail/<recipient>.jsonl` (one file per recipient)

---

## Phase 1: Setup (Project Initialization)

**Purpose**: Create Go project structure and initialize module

- [x] T001 Create directory structure: `cmd/agentmail/`, `internal/mail/`, `internal/tmux/`, `internal/cli/`
- [x] T002 Initialize Go module with `go mod init` in project root (go.mod)
- [x] T003 Create placeholder main.go in cmd/agentmail/main.go with basic CLI skeleton

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure shared by ALL user stories - MUST complete before any story work

**⚠️ CRITICAL**: Both US1 and US2 depend on tmux integration and message/mailbox types

### Tests for Foundational Components

- [ ] T004 [P] Write tests for tmux detection (InTmux, GetCurrentWindow) in internal/tmux/tmux_test.go
- [ ] T005 [P] Write tests for tmux window listing (ListWindows, WindowExists) in internal/tmux/tmux_test.go
- [ ] T006 [P] Write tests for Message struct JSON marshaling in internal/mail/message_test.go
- [ ] T007 [P] Write tests for unique ID generation (8-char base62) in internal/mail/message_test.go

### Implementation for Foundational Components

- [ ] T008 [P] Implement Message struct with JSON tags in internal/mail/message.go
- [ ] T009 [P] Implement GenerateID function (crypto/rand, base62) in internal/mail/message.go
- [ ] T010 Implement InTmux detection (check $TMUX env var) in internal/tmux/tmux.go
- [ ] T011 Implement GetCurrentWindow (tmux display-message -p '#W') in internal/tmux/tmux.go
- [ ] T012 Implement ListWindows (tmux list-windows -F '#{window_name}') in internal/tmux/tmux.go
- [ ] T013 Implement WindowExists helper function in internal/tmux/tmux.go

**Checkpoint**: tmux integration and message types ready - user story implementation can begin

---

## Phase 3: User Story 1 - Send Mail to Another Agent (Priority: P1) MVP

**Goal**: Enable agents to send messages to other agents via `agentmail send <recipient> <message>`

**Independent Test**: Run `agentmail send agent-2 "Hello"` and verify `.git/mail/agent-2.jsonl` created with correct JSONL format

### Tests for User Story 1

- [ ] T014 [P] [US1] Write tests for mailbox Append to recipient file in internal/mail/mailbox_test.go
- [ ] T015 [P] [US1] Write tests for send command argument validation in internal/cli/send_test.go
- [ ] T016 [P] [US1] Write tests for send command recipient validation in internal/cli/send_test.go
- [ ] T017 [P] [US1] Write tests for send command success path (ID output) in internal/cli/send_test.go

### Implementation for User Story 1

- [ ] T018 [US1] Implement EnsureMailDir (create .git/mail/ if missing) in internal/mail/mailbox.go
- [ ] T019 [US1] Implement Append function (appends to .git/mail/<recipient>.jsonl with file locking) in internal/mail/mailbox.go
- [ ] T020 [US1] Implement Send command structure in internal/cli/send.go
- [ ] T021 [US1] Add tmux validation to Send (exit code 2 if not in tmux) in internal/cli/send.go
- [ ] T022 [US1] Add recipient validation to Send (check WindowExists) in internal/cli/send.go
- [ ] T023 [US1] Add message storage and ID output to Send in internal/cli/send.go
- [ ] T024 [US1] Wire up send subcommand in cmd/agentmail/main.go

**Checkpoint**: `agentmail send` is fully functional - can test independently by sending messages

---

## Phase 4: User Story 2 - Receive Incoming Mail (Priority: P1)

**Goal**: Enable agents to receive their oldest unread message via `agentmail receive`

**Independent Test**: Pre-populate `.git/mail/agent-1.jsonl` with messages, run receive as agent-1, verify oldest unread displayed and marked read

### Tests for User Story 2

- [ ] T025 [P] [US2] Write tests for mailbox ReadAll from recipient file in internal/mail/mailbox_test.go
- [ ] T026 [P] [US2] Write tests for mailbox FindUnread (filter by read_flag only) in internal/mail/mailbox_test.go
- [ ] T027 [P] [US2] Write tests for mailbox MarkAsRead operation in internal/mail/mailbox_test.go
- [ ] T028 [P] [US2] Write tests for receive command no-messages case in internal/cli/receive_test.go
- [ ] T029 [P] [US2] Write tests for receive command success path in internal/cli/receive_test.go

### Implementation for User Story 2

- [ ] T030 [US2] Implement ReadAll function (read .git/mail/<recipient>.jsonl) in internal/mail/mailbox.go
- [ ] T031 [US2] Implement FindUnread query (filter by read_flag only, no recipient filter needed) in internal/mail/mailbox.go
- [ ] T032 [US2] Implement WriteAll function (write to recipient file with locking) in internal/mail/mailbox.go
- [ ] T033 [US2] Implement MarkAsRead function in internal/mail/mailbox.go
- [ ] T034 [US2] Implement Receive command structure in internal/cli/receive.go
- [ ] T035 [US2] Add tmux validation to Receive (exit code 2 if not in tmux) in internal/cli/receive.go
- [ ] T036 [US2] Add message retrieval and display formatting to Receive in internal/cli/receive.go
- [ ] T037 [US2] Add "No unread messages" handling (exit code 0) to Receive in internal/cli/receive.go
- [ ] T038 [US2] Wire up receive subcommand in cmd/agentmail/main.go

**Checkpoint**: Both `agentmail send` and `agentmail receive` are fully functional

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Integration testing, coverage verification, and final cleanup

- [ ] T039 Write integration test: send → receive round-trip in internal/cli/integration_test.go
- [ ] T040 Write integration test: FIFO ordering (send 3, receive 3) in internal/cli/integration_test.go
- [ ] T041 Write integration test: multi-agent file isolation (separate .jsonl per recipient) in internal/cli/integration_test.go
- [ ] T042 Run `go test -cover ./...` and verify 80% coverage (SC-005)
- [ ] T043 Run `go vet ./...` and fix any issues
- [ ] T044 Run `go fmt ./...` to ensure consistent formatting
- [ ] T045 Validate quickstart.md scenarios work end-to-end
- [ ] T046 Verify send/receive complete in <1 second (SC-001, SC-002)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - start immediately
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational (Phase 2)
- **User Story 2 (Phase 4)**: Depends on Foundational (Phase 2), can run in parallel with US1
- **Polish (Phase 5)**: Depends on both US1 and US2 completion

### User Story Dependencies

| Story | Depends On | Can Parallel With |
|-------|------------|-------------------|
| US1 (Send) | Foundational | US2 (different files) |
| US2 (Receive) | Foundational | US1 (different files) |

### Within Each Phase

```
Phase 2 (Foundational):
  T004-T007 (tests) → in parallel
  T008-T009 (message) → in parallel, no deps
  T010-T013 (tmux) → sequential (T010 → T011 → T012 → T013)

Phase 3 (US1 - Send):
  T014-T017 (tests) → in parallel, write first
  T018-T019 (mailbox) → sequential
  T020-T024 (CLI) → sequential

Phase 4 (US2 - Receive):
  T025-T029 (tests) → in parallel, write first
  T030-T033 (mailbox) → sequential
  T034-T038 (CLI) → sequential
```

---

## Parallel Opportunities

### Foundational Phase Parallelism

```bash
# All test files can be written in parallel:
Task: "Write tests for tmux detection in internal/tmux/tmux_test.go"
Task: "Write tests for tmux window listing in internal/tmux/tmux_test.go"
Task: "Write tests for Message struct in internal/mail/message_test.go"
Task: "Write tests for unique ID generation in internal/mail/message_test.go"

# Message implementation (parallel with tmux):
Task: "Implement Message struct in internal/mail/message.go"
Task: "Implement GenerateID in internal/mail/message.go"
```

### User Story Parallelism

```bash
# US1 and US2 can run in parallel after Foundational:
# Developer A: US1 (send)
# Developer B: US2 (receive)

# Within US1, tests can be parallel:
Task: "Write tests for mailbox Append in internal/mail/mailbox_test.go"
Task: "Write tests for send argument validation in internal/cli/send_test.go"
Task: "Write tests for send recipient validation in internal/cli/send_test.go"
Task: "Write tests for send success path in internal/cli/send_test.go"
```

---

## Implementation Strategy

### MVP First (Recommended)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T013)
3. Complete Phase 3: User Story 1 - Send (T014-T024)
4. **VALIDATE**: Test send command independently
5. Complete Phase 4: User Story 2 - Receive (T025-T038)
6. **VALIDATE**: Test both commands, verify round-trip
7. Complete Phase 5: Polish (T039-T045)
8. **FINAL**: Verify 80% coverage

### Incremental Delivery

| Increment | Stories Included | Deliverable |
|-----------|-----------------|-------------|
| MVP | US1 only | Can send messages (store only) |
| Full | US1 + US2 | Complete send/receive flow |
| Polished | All + tests | Production-ready with 80% coverage |

---

## Notes

- All tests follow TDD: write test → verify fail → implement → verify pass
- File locking (syscall.Flock) required for mailbox write operations
- Exit codes: 0 (success), 1 (error), 2 (not in tmux)
- Message ID format: 8-character base62 string
- JSONL storage: `.git/mail/<recipient>.jsonl` (one file per recipient)
- Per-recipient files simplify queries (no recipient filtering needed)
- Commit after each task or logical group for easy rollback

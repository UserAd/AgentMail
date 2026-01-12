# Tasks: Mailman Daemon

**Input**: Design documents from `/specs/006-mailman-daemon/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Required per constitution (80%+ coverage mandate)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

This project uses Go standard structure:
- **Source**: `internal/` for internal packages, `cmd/` for entry points
- **Tests**: `*_test.go` files alongside source (Go convention)

---

## Phase 1: Setup

**Purpose**: Create new package structure for daemon functionality

- [x] T001 Create internal/daemon/ directory for daemon package
- [x] T002 [P] Create internal/mail/recipients.go file with package declaration
- [x] T003 [P] Create internal/tmux/sendkeys.go file with package declaration

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### RecipientState Data Model

- [x] T004 [P] Define RecipientState struct in internal/mail/recipients.go per data-model.md
- [x] T005 [P] Implement ReadAllRecipients() to parse .git/mail-recipients.jsonl in internal/mail/recipients.go
- [x] T006 [P] Implement WriteAllRecipients() with file locking in internal/mail/recipients.go
- [x] T007 Implement UpdateRecipientState() for atomic read-modify-write in internal/mail/recipients.go
- [x] T008 [P] Write unit tests for recipients.go in internal/mail/recipients_test.go

### Tmux SendKeys Wrapper

- [x] T009 [P] Implement SendKeys(window, text string) in internal/tmux/sendkeys.go
- [x] T010 [P] Implement SendEnter(window string) in internal/tmux/sendkeys.go
- [x] T011 [P] Write unit tests for sendkeys.go in internal/tmux/sendkeys_test.go

**Checkpoint**: Foundation ready - RecipientState and SendKeys available for all user stories

---

## Phase 3: User Story 1 - Start Mailman Daemon (Priority: P1) 🎯 MVP

**Goal**: Enable starting the mailman daemon in foreground or background mode with PID file creation

**Independent Test**: Run `agentmail mailman` and verify daemon starts, creates PID file, outputs status

### Tests for User Story 1

- [ ] T012 [P] [US1] Write test for PID file creation in internal/daemon/daemon_test.go
- [ ] T013 [P] [US1] Write test for foreground mode startup in internal/daemon/daemon_test.go
- [ ] T014 [P] [US1] Write test for background mode (--daemon flag) in internal/daemon/daemon_test.go

### Implementation for User Story 1

- [ ] T015 [P] [US1] Implement ReadPID() to read .git/mail/mailman.pid in internal/daemon/daemon.go
- [ ] T016 [P] [US1] Implement WritePID() to write current PID in internal/daemon/daemon.go
- [ ] T017 [P] [US1] Implement DeletePID() to remove PID file in internal/daemon/daemon.go
- [ ] T018 [US1] Implement StartDaemon() with foreground mode in internal/daemon/daemon.go
- [ ] T019 [US1] Implement daemonize logic for --daemon flag in internal/daemon/daemon.go
- [ ] T020 [US1] Add mailman command to CLI dispatch in cmd/agentmail/main.go
- [ ] T021 [US1] Implement RunMailman() command handler in internal/cli/mailman.go
- [ ] T022 [P] [US1] Write integration test for mailman command in internal/cli/mailman_test.go

**Checkpoint**: Daemon can start in foreground/background, PID file created

---

## Phase 4: User Story 2 - Singleton Process Control (Priority: P1)

**Goal**: Prevent duplicate daemon instances, detect and clean stale PID files

**Independent Test**: Start daemon, try starting second instance, verify exit code 2 and error message

### Tests for User Story 2

- [ ] T023 [P] [US2] Write test for duplicate daemon detection in internal/daemon/daemon_test.go
- [ ] T024 [P] [US2] Write test for stale PID detection in internal/daemon/daemon_test.go
- [ ] T025 [P] [US2] Write test for corrupted PID file handling in internal/daemon/daemon_test.go

### Implementation for User Story 2

- [ ] T026 [US2] Implement IsProcessRunning(pid int) using signal 0 in internal/daemon/daemon.go
- [ ] T027 [US2] Implement CheckExistingDaemon() that returns running/stale/none in internal/daemon/daemon.go
- [ ] T028 [US2] Update StartDaemon() to check for existing daemon before starting in internal/daemon/daemon.go
- [ ] T029 [US2] Add stale PID cleanup logic with warning output in internal/daemon/daemon.go
- [ ] T030 [US2] Implement signal handling for SIGTERM/SIGINT cleanup in internal/daemon/daemon.go
- [ ] T031 [P] [US2] Write integration test for singleton behavior in internal/cli/mailman_test.go

**Checkpoint**: Only one daemon can run, stale PIDs detected and cleaned

---

## Phase 5: User Story 3 - Agent State Management (Priority: P1)

**Goal**: Enable agents to register status via `agentmail status` command for hooks integration

**Independent Test**: Run `agentmail status ready/work/offline` and verify state in .git/mail-recipients.jsonl

### Tests for User Story 3

- [ ] T032 [P] [US3] Write test for status ready command in internal/cli/status_test.go
- [ ] T033 [P] [US3] Write test for status work command in internal/cli/status_test.go
- [ ] T034 [P] [US3] Write test for status offline command in internal/cli/status_test.go
- [ ] T035 [P] [US3] Write test for non-tmux silent exit in internal/cli/status_test.go
- [ ] T036 [P] [US3] Write test for invalid status name error in internal/cli/status_test.go

### Implementation for User Story 3

- [ ] T037 [US3] Add status command to CLI dispatch in cmd/agentmail/main.go
- [ ] T038 [US3] Implement RunStatus() command handler in internal/cli/status.go
- [ ] T039 [US3] Implement ValidateStatus() for ready/work/offline enum in internal/cli/status.go
- [ ] T040 [US3] Handle non-tmux case (silent exit 0) in internal/cli/status.go
- [ ] T041 [US3] Implement notified flag reset on work/offline transition in internal/cli/status.go
- [ ] T042 [P] [US3] Write integration test for status command in internal/cli/status_test.go

**Checkpoint**: Status command works, state persisted to JSONL, hooks-compatible (silent)

---

## Phase 6: User Story 4 - Notification Delivery (Priority: P2)

**Goal**: Daemon monitors mailboxes and sends tmux notifications to ready agents

**Independent Test**: Set agent to ready, send message, verify "Check your agentmail" notification in tmux

### Tests for User Story 4

- [ ] T043 [P] [US4] Write test for notification loop interval (10s) in internal/daemon/loop_test.go
- [ ] T044 [P] [US4] Write test for ready agent notification in internal/daemon/loop_test.go
- [ ] T045 [P] [US4] Write test for work/offline agent skip in internal/daemon/loop_test.go
- [ ] T046 [P] [US4] Write test for notified flag prevents duplicate in internal/daemon/loop_test.go
- [ ] T047 [P] [US4] Write test for stale state cleanup (>1hr) in internal/daemon/loop_test.go

### Implementation for User Story 4

- [ ] T048 [US4] Implement ListMailboxRecipients() to find all .jsonl files in internal/mail/recipients.go
- [ ] T049 [US4] Implement CleanStaleStates() for >1hr old entries in internal/mail/recipients.go
- [ ] T050 [US4] Implement CheckAndNotify() single cycle logic in internal/daemon/loop.go
- [ ] T051 [US4] Implement NotifyAgent() with send-keys + 1s delay + Enter in internal/daemon/loop.go
- [ ] T052 [US4] Implement RunLoop() with 10-second interval in internal/daemon/loop.go
- [ ] T053 [US4] Integrate loop into StartDaemon() in internal/daemon/daemon.go
- [ ] T054 [US4] Update notified flag after successful notification in internal/daemon/loop.go
- [ ] T055 [P] [US4] Write integration test for notification delivery in internal/daemon/loop_test.go

**Checkpoint**: Full notification flow works - daemon monitors and notifies ready agents

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup

- [ ] T056 Run go vet ./... and fix any issues
- [ ] T057 Run go fmt ./... and verify no changes
- [ ] T058 Run go test -cover ./... and verify >= 80% coverage
- [ ] T059 Run go test -race ./... to check for race conditions
- [ ] T060 Update help text in internal/cli/help.go for mailman and status commands
- [ ] T061 Validate all acceptance scenarios from spec.md pass
- [ ] T062 Run quickstart.md validation manually

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1: Setup ──────────────────────────────────────────┐
                                                          │
Phase 2: Foundational ◀──────────────────────────────────┘
         (RecipientState + SendKeys)
              │
              ├──────────────┬──────────────┬──────────────┐
              ▼              ▼              ▼              ▼
         Phase 3         Phase 4        Phase 5        Phase 6
         US1: Start      US2: Single    US3: Status    US4: Notify
         Daemon          ton Control    Command        Loop
              │              │              │              │
              └──────────────┴──────────────┴──────────────┘
                                    │
                                    ▼
                              Phase 7: Polish
```

### User Story Dependencies

- **User Story 1 (P1)**: Depends on Foundational - No other story dependencies
- **User Story 2 (P1)**: Depends on US1 (extends StartDaemon with checks)
- **User Story 3 (P1)**: Depends on Foundational only - Independent of US1/US2
- **User Story 4 (P2)**: Depends on US1 (daemon must run), US3 (status must exist)

### Within Each User Story

- Tests FIRST - ensure they FAIL before implementation
- Models/utilities before services
- Services before CLI handlers
- Core logic before integration

### Parallel Opportunities

**Setup Phase**:
```
T002 [P] recipients.go    ──┐
                            ├── All parallel
T003 [P] sendkeys.go      ──┘
```

**Foundational Phase**:
```
T004-T008 Recipients  ──┐
                        ├── Run in parallel
T009-T011 SendKeys    ──┘
```

**User Story Phases** (after Foundational):
```
US1 + US3 can run in parallel (independent)
US2 depends on US1
US4 depends on US1 + US3
```

---

## Parallel Example: Foundational Phase

```bash
# Launch all foundational tasks in parallel:
# Recipients package:
Task: "Define RecipientState struct in internal/mail/recipients.go"
Task: "Implement ReadAllRecipients() in internal/mail/recipients.go"
Task: "Implement WriteAllRecipients() in internal/mail/recipients.go"
Task: "Write unit tests in internal/mail/recipients_test.go"

# SendKeys package (parallel with Recipients):
Task: "Implement SendKeys() in internal/tmux/sendkeys.go"
Task: "Implement SendEnter() in internal/tmux/sendkeys.go"
Task: "Write unit tests in internal/tmux/sendkeys_test.go"
```

---

## Implementation Strategy

### MVP First (User Stories 1-3)

1. Complete Phase 1: Setup (3 tasks)
2. Complete Phase 2: Foundational (8 tasks) - BLOCKS all stories
3. Complete Phase 3: User Story 1 - Daemon starts (11 tasks)
4. Complete Phase 4: User Story 2 - Singleton control (9 tasks)
5. Complete Phase 5: User Story 3 - Status command (11 tasks)
6. **STOP and VALIDATE**: Test mailman + status independently
7. Deploy/demo if ready - daemon works but doesn't notify yet

### Full Feature (Add US4)

8. Complete Phase 6: User Story 4 - Notification loop (13 tasks)
9. Complete Phase 7: Polish (7 tasks)
10. Final validation with quickstart.md

### Incremental Delivery

| Checkpoint | What Works | Tasks |
|------------|------------|-------|
| After US1 | Daemon starts, PID file created | T001-T022 |
| After US2 | Singleton control, stale cleanup | T023-T031 |
| After US3 | Status command, hooks integration | T032-T042 |
| After US4 | Full notification delivery | T043-T055 |
| After Polish | Production ready | T056-T062 |

---

## Summary

| Phase | Tasks | Purpose |
|-------|-------|---------|
| Phase 1: Setup | 3 | Create package structure |
| Phase 2: Foundational | 8 | RecipientState + SendKeys |
| Phase 3: US1 Start Daemon | 11 | Daemon foreground/background |
| Phase 4: US2 Singleton | 9 | Duplicate prevention |
| Phase 5: US3 Status | 11 | Hooks integration |
| Phase 6: US4 Notify | 13 | Notification loop |
| Phase 7: Polish | 7 | Validation & cleanup |
| **Total** | **62** | |

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Tests are REQUIRED per constitution (80%+ coverage)
- US1 + US3 can proceed in parallel after Foundational
- US2 extends US1, US4 integrates all previous stories
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently

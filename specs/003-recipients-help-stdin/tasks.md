# Tasks: Recipients Command, Help Flag, and Stdin Message Input

**Input**: Design documents from `/specs/003-recipients-help-stdin/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Included per constitution requirement (80% minimum coverage)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Based on plan.md structure:
- **CLI commands**: `internal/cli/`
- **Mail utilities**: `internal/mail/`
- **Tmux integration**: `internal/tmux/`
- **Main entry**: `cmd/agentmail/`

---

## Phase 1: Setup

**Purpose**: No new setup needed - extending existing project structure

- [X] T001 Verify existing project builds with `go build ./...`
- [X] T002 Verify existing tests pass with `go test ./...`

---

## Phase 2: Foundational (Shared Infrastructure)

**Purpose**: Core infrastructure used by multiple user stories

**⚠️ CRITICAL**: US2, US3, US5 depend on these foundational components

- [X] T003 [P] Implement `FindGitRoot()` function in internal/mail/ignore.go (per research.md pattern)
- [X] T004 [P] Implement `LoadIgnoreList()` function in internal/mail/ignore.go (per research.md pattern) - implements FR-003 (parser treats non-empty, whitespace-trimmed lines as exclusions)
- [X] T005 [P] Write tests for ignore.go in internal/mail/ignore_test.go
- [X] T006 [P] Implement `IsStdinPipe()` helper function in internal/cli/stdin.go (per research.md pattern)
- [X] T007 [P] Write tests for stdin.go in internal/cli/stdin_test.go

**Checkpoint**: Foundation ready - `go test ./...` passes with new files

---

## Phase 3: User Story 1 - List Available Recipients (Priority: P1) 🎯 MVP

**Goal**: Display all tmux windows with current window marked "[you]"

**Independent Test**: Run `agentmail recipients` and verify output shows all windows with current marked

### Tests for User Story 1

- [X] T008 [P] [US1] Write unit tests for Recipients() in internal/cli/recipients_test.go
- [X] T009 [P] [US1] Write test: lists all windows one per line
- [X] T010 [P] [US1] Write test: marks current window with "[you]" suffix
- [X] T011 [P] [US1] Write test: returns exit code 2 when not in tmux

### Implementation for User Story 1

- [X] T012 [US1] Create Recipients() function in internal/cli/recipients.go
- [X] T013 [US1] Implement window listing using tmux.ListWindows()
- [X] T014 [US1] Implement current window detection using tmux.GetCurrentWindow()
- [X] T015 [US1] Add "[you]" marker formatting for current window
- [X] T016 [US1] Add `recipients` command routing in cmd/agentmail/main.go
- [X] T017 [US1] Verify tests pass: `go test -v ./internal/cli/... -run Recipients`

**Checkpoint**: `agentmail recipients` shows all windows with "[you]" marker

---

## Phase 4: User Story 2 - Filter Recipients with Ignore File (Priority: P1)

**Goal**: Exclude windows listed in `.agentmailignore` from recipients output

**Independent Test**: Create `.agentmailignore` with window names and verify they don't appear in output

**Depends on**: Phase 2 (ignore.go), US1

### Tests for User Story 2

- [X] T018 [P] [US2] Write test: excludes windows listed in .agentmailignore
- [X] T019 [P] [US2] Write test: handles missing .agentmailignore gracefully (shows all windows)
- [X] T020 [P] [US2] Write test: ignores empty and whitespace-only lines in ignore file
- [X] T021 [P] [US2] Write test: handles unreadable ignore file (per FR-013)

### Implementation for User Story 2

- [X] T022 [US2] Integrate LoadIgnoreList() into Recipients() function in internal/cli/recipients.go
- [X] T023 [US2] Filter windows against ignore list before output
- [X] T024 [US2] Current window shown with "[you]" even if in ignore list (per FR-004)
- [X] T025 [US2] Verify tests pass: `go test -v ./internal/cli/... -run Recipients`

**Checkpoint**: Recipients list correctly filters ignored windows

---

## Phase 5: User Story 3 - Block Sending to Ignored Recipients (Priority: P1)

**Goal**: Reject send attempts to windows in `.agentmailignore` with "recipient not found"

**Independent Test**: Add window to `.agentmailignore` and verify send fails

**Depends on**: Phase 2 (ignore.go)

### Tests for User Story 3

- [X] T026 [P] [US3] Write test: send to ignored recipient returns "recipient not found" error
- [X] T027 [P] [US3] Write test: send to valid recipient (not ignored) succeeds
- [X] T028 [P] [US3] Write test: send to self returns "recipient not found" error

### Implementation for User Story 3

- [X] T029 [US3] Add ignore list validation to Send() function in internal/cli/send.go
- [X] T030 [US3] Return exit code 1 with "recipient not found" for ignored recipients
- [X] T031 [US3] Verify tests pass: `go test -v ./internal/cli/... -run Send`

**Checkpoint**: Sending to ignored windows fails with proper error

---

## Phase 6: User Story 4 - Display Help Information (Priority: P2)

**Goal**: Show usage documentation via `agentmail --help`

**Independent Test**: Run `agentmail --help` and verify output contains all commands

### Tests for User Story 4

- [X] T032 [P] [US4] Write unit tests for Help() in internal/cli/help_test.go
- [X] T033 [P] [US4] Write test: help output includes send command with syntax
- [X] T034 [P] [US4] Write test: help output includes receive command with syntax
- [X] T035 [P] [US4] Write test: help output includes recipients command with syntax
- [X] T036 [P] [US4] Write test: help output includes examples section

### Implementation for User Story 4

- [X] T037 [US4] Create Help() function in internal/cli/help.go with help text (per research.md)
- [X] T038 [US4] Add --help and -h flag handling in cmd/agentmail/main.go (before command parsing)
- [X] T039 [US4] Ensure help returns exit code 0
- [X] T040 [US4] Verify tests pass: `go test -v ./internal/cli/... -run Help`

**Checkpoint**: `agentmail --help` displays complete usage information

---

## Phase 7: User Story 5 - Send Message via Stdin (Priority: P2)

**Goal**: Accept message content from stdin pipe

**Independent Test**: Run `echo "test" | agentmail send <recipient>` and verify message received

**Depends on**: Phase 2 (stdin.go)

### Tests for User Story 5

- [X] T041 [P] [US5] Write test: message from stdin is used when piped
- [X] T042 [P] [US5] Write test: multi-line stdin content sent as single message
- [X] T043 [P] [US5] Write test: stdin takes precedence over argument (per FR-010)
- [X] T044 [P] [US5] Write test: falls back to argument when no stdin data

### Implementation for User Story 5

- [X] T045 [US5] Modify Send() in internal/cli/send.go to accept io.Reader for stdin
- [X] T046 [US5] Implement stdin detection using IsStdinPipe() helper
- [X] T047 [US5] Read stdin content with io.ReadAll when pipe detected
- [X] T048 [US5] Implement stdin precedence logic (stdin > argument)
- [X] T049 [US5] Update main.go to pass os.Stdin to Send()
- [X] T050 [US5] Verify tests pass: `go test -v ./internal/cli/... -run Send`

**Checkpoint**: `echo "msg" | agentmail send recipient` works correctly

---

## Phase 8: User Story 6 - Send Message via Command Argument (Priority: P1)

**Goal**: Ensure backwards compatibility with existing send syntax

**Independent Test**: Run `agentmail send <recipient> "message"` and verify it works

**Note**: This is existing behavior - just verify no regressions

### Tests for User Story 6

- [ ] T051 [P] [US6] Write test: argument-based send still works after stdin changes
- [ ] T052 [P] [US6] Write test: no message argument and no stdin returns usage error (FR-011)

### Implementation for User Story 6

- [ ] T053 [US6] Verify existing send tests still pass
- [ ] T054 [US6] Add regression test for argument-only send path
- [ ] T055 [US6] Verify tests pass: `go test -v ./internal/cli/... -run Send`

**Checkpoint**: Existing send behavior unchanged

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Quality gates and final validation

- [ ] T056 Run full test suite: `go test -v ./...`
- [ ] T057 Verify coverage meets 80%: `go test -cover ./...`
- [ ] T058 Run static analysis: `go vet ./...`
- [ ] T059 Run formatting check: `go fmt ./...`
- [ ] T060 Manual validation: Run quickstart.md test sequence
- [ ] T061 Update usage message in main.go to include `recipients` command

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1 (Setup)
     │
     ▼
Phase 2 (Foundational) ─────────────────────────────────────┐
     │                                                       │
     ├─────────────┬─────────────┬─────────────┐            │
     ▼             ▼             ▼             ▼            ▼
Phase 3 (US1) → Phase 4 (US2)   Phase 5 (US3) Phase 6 (US4) Phase 7 (US5)
                     │                                           │
                     └──────────────────────────────────────────┘
                                        │
                                        ▼
                               Phase 8 (US6) - Regression
                                        │
                                        ▼
                               Phase 9 (Polish)
```

### User Story Dependencies

- **US1**: Foundational (Phase 2) - No other story dependencies
- **US2**: Foundational + US1 (builds on recipients output)
- **US3**: Foundational only (modifies send.go independently)
- **US4**: None (independent help system)
- **US5**: Foundational only (modifies send.go independently)
- **US6**: US3 + US5 complete (regression testing after send changes)

### Within Each User Story

1. Tests written FIRST (should fail initially)
2. Implementation follows
3. Tests must pass before checkpoint

### Parallel Opportunities

**Phase 2 (Foundational)**:
```bash
# All can run in parallel:
Task T003: Implement FindGitRoot() in internal/mail/ignore.go
Task T004: Implement LoadIgnoreList() in internal/mail/ignore.go
Task T006: Implement IsStdinPipe() in internal/cli/stdin.go
```

**Phase 3-8 (User Stories)**:
```bash
# After Phase 2, these can run in parallel:
- US1 (Phase 3): Recipients command
- US4 (Phase 6): Help command
- US5 (Phase 7): Stdin support (partial - needs stdin.go from Phase 2)
```

**Within each story** - all tests marked [P] can run in parallel

---

## Parallel Example: Foundational Phase

```bash
# Launch all foundational tasks in parallel:
Task: "Implement FindGitRoot() in internal/mail/ignore.go"
Task: "Implement LoadIgnoreList() in internal/mail/ignore.go"
Task: "Implement IsStdinPipe() in internal/cli/stdin.go"
Task: "Write tests for ignore.go in internal/mail/ignore_test.go"
Task: "Write tests for stdin.go in internal/cli/stdin_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup verification
2. Complete Phase 2: Foundational (ignore.go, stdin.go)
3. Complete Phase 3: User Story 1 (recipients command)
4. **STOP and VALIDATE**: `agentmail recipients` works
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 (recipients) → Test → MVP!
3. Add US2 (ignore filtering) → Test
4. Add US3 (send validation) → Test
5. Add US4 (help) → Test
6. Add US5 (stdin) → Test
7. US6 (regression) → Test
8. Polish → Ship

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: US1 → US2 (recipients chain)
   - Developer B: US4 (help - independent)
   - Developer C: US3 + US5 (send modifications)
3. Developer D: US6 regression after send changes complete
4. All: Polish phase

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Constitution requires 80% test coverage - tests included for all new code
- All tests use Go standard library `testing` package
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently

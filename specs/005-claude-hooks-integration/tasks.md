# Tasks: Claude Code Hooks Integration

**Input**: Design documents from `/specs/005-claude-hooks-integration/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Status**: Implementation complete - tasks documented for reference

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2)
- Exact file paths included in descriptions

## Path Conventions

- **Project type**: Single CLI application
- **Source**: `cmd/agentmail/`, `internal/cli/`
- **Tests**: `internal/cli/*_test.go`

---

## Phase 1: Setup (CLI Framework Migration)

**Purpose**: Migrate from manual argument parsing to ffcli for robust flag handling

- [x] T001 Add ff/v3 dependency in `go.mod`
- [x] T002 Restructure CLI with ffcli subcommands in `cmd/agentmail/main.go`
- [x] T003 [P] Add `-r`/`--recipient` flags for send command in `cmd/agentmail/main.go`
- [x] T004 [P] Add `-m`/`--message` flags for send command in `cmd/agentmail/main.go`
- [x] T005 Run `go mod tidy` to fetch dependencies

**Checkpoint**: CLI framework ready, flag parsing infrastructure in place

---

## Phase 2: User Story 1 - Claude Code Hook Notification (Priority: P1) 🎯 MVP

**Goal**: Enable `agentmail receive --hook` to notify Claude Code users of pending messages

**Independent Test**: Run `agentmail receive --hook` in various scenarios and verify exit codes and output streams

### Implementation for User Story 1

- [x] T006 [US1] Add `--hook` flag definition in `cmd/agentmail/main.go`
- [x] T007 [US1] Add `HookMode bool` field to `ReceiveOptions` in `internal/cli/receive.go`
- [x] T008 [US1] Implement FR-003: Silent exit (code 0) when not in tmux in `internal/cli/receive.go`
- [x] T009 [US1] Implement FR-004a/b/c: Silent exit (code 0) on errors in `internal/cli/receive.go`
- [x] T010 [US1] Implement FR-002: Silent exit (code 0) when no messages in `internal/cli/receive.go`
- [x] T011 [US1] Implement FR-001a: Write "You got new mail\n" + message to STDERR in `internal/cli/receive.go`
- [x] T012 [US1] Implement FR-001b: Exit code 2 when messages exist in `internal/cli/receive.go`
- [x] T013 [US1] Implement FR-005: All output to STDERR in hook mode in `internal/cli/receive.go`

### Tests for User Story 1

- [x] T014 [P] [US1] Test hook mode with messages (STDERR output, exit 2) in `internal/cli/receive_test.go`
- [x] T015 [P] [US1] Test hook mode with no messages (silent, exit 0) in `internal/cli/receive_test.go`
- [x] T016 [P] [US1] Test hook mode not in tmux (silent, exit 0) in `internal/cli/receive_test.go`
- [x] T017 [P] [US1] Test hook mode with errors (silent, exit 0) in `internal/cli/receive_test.go`
- [x] T018 [P] [US1] Test hook mode output stream verification in `internal/cli/receive_test.go`
- [x] T019 [P] [US1] Regression test: normal mode unchanged in `internal/cli/receive_test.go`

**Checkpoint**: Hook mode fully functional and tested - MVP complete

---

## Phase 3: User Story 2 - Documentation for Hook Setup (Priority: P2)

**Goal**: Provide clear instructions for configuring agentmail as a Claude Code hook

**Independent Test**: Follow README instructions to configure hook and verify it works

### Implementation for User Story 2

- [x] T020 [US2] Add "Claude Code Hooks" section to `README.md`
- [x] T021 [US2] Document setup instructions with settings.json example in `README.md`
- [x] T022 [US2] Document hook mode behavior table in `README.md`
- [x] T023 [US2] Update send command documentation with flag syntax in `README.md`
- [x] T024 [US2] Update receive command documentation with --hook flag in `README.md`

**Checkpoint**: Documentation complete - users can configure hooks

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup

- [x] T025 Update research.md with ff/v3 dependency justification in `specs/005-claude-hooks-integration/research.md`
- [x] T026 Run `go test ./...` to verify all tests pass
- [x] T027 Run `go vet ./...` to verify no issues
- [x] T028 Run `go build ./...` to verify build succeeds

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - starts immediately
- **User Story 1 (Phase 2)**: Depends on Setup completion
- **User Story 2 (Phase 3)**: Can run in parallel with Phase 2 (different files)
- **Polish (Phase 4)**: Depends on all user stories complete

### User Story Dependencies

- **User Story 1 (P1)**: Core functionality - no dependencies on other stories
- **User Story 2 (P2)**: Documentation - can be done in parallel with US1

### Within Each User Story

- Implementation tasks before test tasks (TDD not required for this feature)
- T006-T007 (flag setup) before T008-T013 (behavior implementation)
- All tests (T014-T019) can run in parallel

### Parallel Opportunities

- T003 and T004 can run in parallel (different flag definitions)
- T014-T019 can all run in parallel (independent test cases)
- T020-T024 can run in parallel with T006-T019 (different files)

---

## Parallel Example: User Story 1 Tests

```bash
# Launch all tests for User Story 1 together:
Task: "Test hook mode with messages in internal/cli/receive_test.go"
Task: "Test hook mode with no messages in internal/cli/receive_test.go"
Task: "Test hook mode not in tmux in internal/cli/receive_test.go"
Task: "Test hook mode with errors in internal/cli/receive_test.go"
Task: "Test hook mode output stream in internal/cli/receive_test.go"
Task: "Regression test normal mode in internal/cli/receive_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. ✅ Complete Phase 1: Setup (CLI framework)
2. ✅ Complete Phase 2: User Story 1 (hook mode)
3. ✅ **VALIDATED**: All tests pass, hook works correctly
4. Ready for deployment

### Incremental Delivery

1. ✅ Setup + US1 → Hook functionality works → MVP complete
2. ✅ Add US2 → Documentation complete → Full feature ready
3. ✅ Polish → All validation passes → Ready for merge

---

## Summary

| Metric | Value |
|--------|-------|
| Total Tasks | 28 |
| User Story 1 Tasks | 14 (T006-T019) |
| User Story 2 Tasks | 5 (T020-T024) |
| Setup Tasks | 5 (T001-T005) |
| Polish Tasks | 4 (T025-T028) |
| Parallel Opportunities | 12 tasks marked [P] |
| Status | ✅ All tasks complete |

## Notes

- All 28 tasks completed
- 7 new hook mode tests added to receive_test.go
- ff/v3 dependency added and justified per constitution
- README updated with comprehensive hook documentation
- All tests pass: `go test ./...` successful

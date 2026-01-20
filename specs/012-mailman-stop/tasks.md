# Tasks: Mailman Daemon Stop Command

**Input**: Design documents from `/specs/012-mailman-stop/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli.md, quickstart.md

**Tests**: Included (constitution requires >= 80% coverage)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Path Conventions

Based on plan.md structure (Go CLI project):

- `cmd/agentmail/main.go` - CLI entry point
- `internal/daemon/` - Daemon package (existing + modifications)
- `internal/cli/` - CLI handlers (existing + new)

---

## Phase 1: Setup

**Purpose**: Add stop file constant and path function

- [x] T001 Add StopFile constant ".stop" in internal/daemon/daemon.go
- [x] T002 Add StopFilePath(repoRoot string) function in internal/daemon/daemon.go

**Checkpoint**: Stop file path infrastructure ready

---

## Phase 2: User Story 1 - Stop Running Daemon (Priority: P1) 🎯 MVP

**Goal**: Create stop signal file and have daemon detect it and shut down

**Independent Test**: Start daemon, run `agentmail mailman stop`, verify daemon terminates and cleans up files

### Tests for User Story 1

- [x] T003 [P] [US1] Write test for successful stop file creation in internal/cli/mailman_stop_test.go
- [x] T004 [P] [US1] Write test for "Stop signal sent" message and exit code 0 in internal/cli/mailman_stop_test.go
- [x] T005 [P] [US1] Write test for daemon stop file detection in internal/daemon/watcher_test.go

### Implementation for User Story 1

- [x] T006 [US1] Create MailmanStopOptions struct in internal/cli/mailman_stop.go
- [x] T007 [US1] Implement MailmanStop function with atomic file creation (O_CREATE|O_EXCL) in internal/cli/mailman_stop.go
- [x] T008 [US1] Add success message "Stop signal sent" output in internal/cli/mailman_stop.go
- [x] T009 [US1] Add stop file detection (fsnotify.Create event for .stop) in internal/daemon/watcher.go
- [x] T010 [US1] Add StopChan() method to FileWatcher to signal shutdown in internal/daemon/watcher.go
- [x] T011 [US1] Update runForeground() to select on fileWatcher.StopChan() in internal/daemon/daemon.go
- [x] T012 [US1] Add stop file removal during daemon shutdown in internal/daemon/daemon.go
- [x] T013 [US1] Register 'stop' subcommand under mailmanCmd in cmd/agentmail/main.go
- [x] T014 [US1] Run tests to verify US1: `go test -v ./internal/cli/... -run MailmanStop && go test -v ./internal/daemon/... -run Stop`

**Checkpoint**: User Story 1 complete - can stop a running daemon via file signal

---

## Phase 3: User Story 2 - Stop Already Pending (Priority: P2)

**Goal**: Detect and report when a stop is already pending

**Independent Test**: Create `.agentmail/.stop` file manually, run `agentmail mailman stop`, verify error message

### Tests for User Story 2

- [x] T015 [P] [US2] Write test for "Stop already pending" when file exists in internal/cli/mailman_stop_test.go
- [x] T016 [P] [US2] Write test for exit code 1 when file exists in internal/cli/mailman_stop_test.go

### Implementation for User Story 2

- [x] T017 [US2] Add os.IsExist error handling for "Stop already pending" message in internal/cli/mailman_stop.go
- [x] T018 [US2] Add generic filesystem error handling "Failed to send stop signal: \<error\>" in internal/cli/mailman_stop.go
- [x] T019 [US2] Run tests to verify US2: `go test -v ./internal/cli/... -run MailmanStop`

**Checkpoint**: Both user stories complete - full stop functionality with error handling

---

## Phase 4: Polish & Quality Gates

**Purpose**: Ensure code meets all quality requirements

- [x] T020 [P] Update mailman command help text in cmd/agentmail/main.go to show stop subcommand
- [x] T021 [P] Run gofmt and fix any formatting issues: `gofmt -w .`
- [x] T022 [P] Run go vet and fix any issues: `go vet ./...`
- [x] T023 Run full test suite with coverage: `go test -v -race -coverprofile=coverage.out ./...`
- [x] T024 Verify coverage >= 80%: `go tool cover -func=coverage.out | grep total`
- [x] T025 Run govulncheck: `govulncheck ./...`
- [x] T026 Run gosec: `gosec ./...`
- [x] T027 Build and manual test per quickstart.md: `go build -o agentmail ./cmd/agentmail`

**Checkpoint**: All quality gates pass - ready for merge

---

## Dependencies & Execution Order

### Phase Dependencies

```text
Phase 1 (Setup)           → No dependencies
Phase 2 (US1)             → Depends on Phase 1
Phase 3 (US2)             → Depends on Phase 2 (shares MailmanStop function)
Phase 4 (Polish)          → Depends on all user stories
```

### User Story Dependencies

- **User Story 1 (P1)**: Requires Phase 1 setup (StopFilePath function)
- **User Story 2 (P2)**: Builds on US1 implementation (error handling in same function)

### Within Each Phase

- Tests MUST be written and FAIL before implementation
- Implementation tasks are sequential within a story
- [P] tasks can run in parallel

---

## Parallel Opportunities

### User Story 1 Tests (T003-T005)

```bash
# Run in parallel - different test files:
T003: Test stop file creation (cli tests)
T004: Test success message (cli tests)
T005: Test daemon detection (watcher tests)
```

### User Story 2 Tests (T015-T016)

```bash
# Run in parallel - different test cases:
T015: Test "already pending" message
T016: Test exit code 1
```

### Polish Phase (T020-T022)

```bash
# Run in parallel - independent operations:
T020: Update help text
T021: Run gofmt
T022: Run go vet
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: User Story 1 (T003-T014)
3. **STOP and VALIDATE**: Test `agentmail mailman stop` with running daemon
4. Can deploy MVP with basic stop functionality

### Full Implementation

1. Complete Setup → StopFilePath ready
2. Add User Story 1 → Can stop running daemons
3. Add User Story 2 → Handles "already pending" case
4. Run Polish phase → Quality gates pass
5. Ready for merge

### Estimated Effort

| Phase | Tasks | Complexity |
|-------|-------|------------|
| Setup | 2 | Simple (constants) |
| US1 | 12 | Medium (file ops + watcher) |
| US2 | 5 | Simple (error handling) |
| Polish | 8 | Standard |
| **Total** | **27** | **~2 hours** |

---

## Notes

- File-based approach is simpler than SIGTERM (no process validation needed)
- Uses existing fsnotify watcher infrastructure
- O_CREATE|O_EXCL provides atomic file creation (prevents race conditions)
- Daemon removes both .stop and .pid files during shutdown
- Cross-platform compatible (works on Windows, macOS, Linux)

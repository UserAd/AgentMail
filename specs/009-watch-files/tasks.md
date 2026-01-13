# Tasks: File-Watching for Mailman with Timer Fallback

**Input**: Design documents from `/specs/009-watch-files/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

**Tests**: Not explicitly requested - implementation tasks only.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

Based on plan.md structure:
- `internal/daemon/` - File watcher and daemon loop modifications
- `internal/mail/` - RecipientState modifications
- `internal/cli/` - Receive command modifications

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add fsnotify dependency and prepare project structure

- [x] T001 Add fsnotify dependency: `go get github.com/fsnotify/fsnotify@v1.9.0`
- [x] T002 Run `go mod tidy` to update go.sum

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T003 Add `LastReadAt int64` field with `json:"last_read_at,omitempty"` tag to `RecipientState` struct in `internal/mail/recipients.go`
- [ ] T004 [P] Create `MonitoringMode` type constants (`ModeWatching`, `ModePolling`) in `internal/daemon/watcher.go`
- [ ] T005 [P] Create `Debouncer` struct with `timer`, `duration`, `mu` fields in `internal/daemon/watcher.go`
- [ ] T006 Implement `Debouncer.Trigger(callback func())` method with trailing-edge debounce (500ms) in `internal/daemon/watcher.go`
- [ ] T007 Implement `Debouncer.Stop()` method to cancel pending timer in `internal/daemon/watcher.go`

**Checkpoint**: Foundation ready - Debouncer and MonitoringMode available for user story implementation

---

## Phase 3: User Story 1 - Instant Notification on New Mail (Priority: P1) 🎯 MVP

**Goal**: Replace 10-second polling with file-watching for mailbox changes - notifications within 2 seconds

**Independent Test**: Start mailman, send message, verify notification arrives within 2 seconds

### Implementation for User Story 1

- [ ] T008 [US1] Create `FileWatcher` struct in `internal/daemon/watcher.go` with fields: `watcher *fsnotify.Watcher`, `debouncer *Debouncer`, `mailboxDir string`, `agentmailDir string`, `stopChan chan struct{}`, `mode MonitoringMode`, `mu sync.Mutex`
- [ ] T009 [US1] Implement `NewFileWatcher(repoRoot string) (*FileWatcher, error)` constructor in `internal/daemon/watcher.go` - creates fsnotify watcher, creates directories if needed (FR-006, FR-007)
- [ ] T010 [US1] Implement `FileWatcher.AddWatches()` method to add watches for `.agentmail/` and `.agentmail/mailboxes/` directories in `internal/daemon/watcher.go` (FR-001, FR-004)
- [ ] T011 [US1] Implement `FileWatcher.isMailboxEvent(event fsnotify.Event) bool` helper to identify mailbox file events (Write/Create on `.jsonl` files in mailboxes/) in `internal/daemon/watcher.go`
- [ ] T012 [US1] Implement `FileWatcher.Run(processFunc func()) error` main event loop in `internal/daemon/watcher.go` - handles Write/Create events for mailbox files, uses debouncer (FR-009, FR-011)
- [ ] T013 [US1] Add 60-second fallback ticker to `FileWatcher.Run()` for safety net notification checks in `internal/daemon/watcher.go` (FR-012)
- [ ] T014 [US1] Implement `FileWatcher.Close()` to stop watcher and debouncer in `internal/daemon/watcher.go`
- [ ] T015 [US1] Modify `runForeground()` in `internal/daemon/daemon.go` to attempt FileWatcher initialization before falling back to RunLoop
- [ ] T016 [US1] Add "File watching enabled" log message when FileWatcher initializes successfully in `internal/daemon/daemon.go` (FR-002a)

**Checkpoint**: User Story 1 complete - mailbox changes trigger instant notifications via file-watching

---

## Phase 4: User Story 2 - Instant Status Change Detection (Priority: P1)

**Goal**: Watch recipients.jsonl for status changes - reload states and check notifications within 2 seconds

**Independent Test**: Set status to "work", send message, set status to "ready", verify notification arrives within 2 seconds

### Implementation for User Story 2

- [ ] T017 [US2] Implement `FileWatcher.isRecipientsEvent(event fsnotify.Event) bool` helper to identify `recipients.jsonl` write events in `internal/daemon/watcher.go` (FR-005)
- [ ] T018 [US2] Extend `FileWatcher.Run()` to handle Write events for `recipients.jsonl` - reload states and trigger notification check (FR-010a, FR-010b) in `internal/daemon/watcher.go`
- [ ] T019 [US2] Handle case when `recipients.jsonl` doesn't exist at startup - watch parent directory for file creation (FR-008) in `internal/daemon/watcher.go`

**Checkpoint**: User Stories 1 AND 2 complete - both mailbox and status changes trigger instant notifications

---

## Phase 5: User Story 3 - Graceful Fallback to Polling (Priority: P2)

**Goal**: Automatic fallback to 10-second polling when file-watching is unavailable

**Independent Test**: Simulate watcher initialization failure, verify daemon continues with 10-second polling

### Implementation for User Story 3

- [ ] T020 [US3] Add "File watching unavailable, using polling" log message when FileWatcher init fails in `internal/daemon/daemon.go` (FR-003a)
- [ ] T021 [US3] Implement fallback to existing `RunLoop()` when FileWatcher initialization fails in `internal/daemon/daemon.go` (FR-003b, FR-016)
- [ ] T022 [US3] Implement runtime error handling in `FileWatcher.Run()` - log error and switch to polling mode (FR-014a, FR-014b) in `internal/daemon/watcher.go`
- [ ] T023 [US3] Implement `FileWatcher.SwitchToPolling(opts LoopOptions)` method to transition from watching to polling mode in `internal/daemon/watcher.go` (FR-015)
- [ ] T024 [US3] Ensure existing `RunLoop()` in `internal/daemon/loop.go` continues to work unchanged for polling mode (FR-013)

**Checkpoint**: User Stories 1, 2, AND 3 complete - file-watching with automatic fallback to polling

---

## Phase 6: User Story 4 - Track Agent Last-Read Time (Priority: P2)

**Goal**: Record when agents last read their mailbox in recipients.jsonl

**Independent Test**: Run `agentmail receive` inside tmux, verify `last_read_at` timestamp is set in recipients.jsonl

### Implementation for User Story 4

- [ ] T025 [US4] Implement `UpdateLastReadAt(repoRoot, recipient string, timestamp int64) error` function in `internal/mail/recipients.go` - uses file locking (FR-020)
- [ ] T026 [US4] Handle case when recipient entry doesn't exist - create new entry with `last_read_at` in `UpdateLastReadAt()` function (FR-019)
- [ ] T027 [US4] Modify `Receive()` function in `internal/cli/receive.go` to call `UpdateLastReadAt()` with current Unix timestamp in milliseconds when inside tmux (FR-017, FR-018)
- [ ] T028 [US4] Add tmux check in `Receive()` before updating last-read timestamp - skip update when outside tmux (FR-021) in `internal/cli/receive.go`

**Checkpoint**: All 4 user stories complete - file-watching with fallback and last-read tracking

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Testing, validation, and quality assurance

- [ ] T029 [P] Add unit tests for `Debouncer` in `internal/daemon/watcher_test.go`
- [ ] T030 [P] Add unit tests for `FileWatcher` initialization and fallback in `internal/daemon/watcher_test.go`
- [ ] T031 [P] Add unit tests for `UpdateLastReadAt()` function in `internal/mail/recipients_test.go`
- [ ] T032 [P] Add unit tests for last-read tracking in `Receive()` in `internal/cli/receive_test.go`
- [ ] T033 Run `go test -v -race ./...` to verify all tests pass with race detection
- [ ] T034 Run `go vet ./...` to check for code issues
- [ ] T035 Run `go fmt ./...` to ensure code formatting
- [ ] T036 Run `govulncheck ./...` for security vulnerabilities
- [ ] T037 Run `gosec ./...` for security issues
- [ ] T038 Build and manually test file-watching with tmux: send message, verify < 2 second notification
- [ ] T039 Manually test fallback: simulate watcher failure, verify polling mode works
- [ ] T040 Manually test last-read tracking: run `agentmail receive`, check `recipients.jsonl` for `last_read_at`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - US1 and US2 are both P1 priority but US2 builds on US1's FileWatcher
  - US3 depends on US1/US2 infrastructure
  - US4 is independent (different component)
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2)
- **User Story 2 (P1)**: Depends on US1 (extends FileWatcher.Run)
- **User Story 3 (P2)**: Depends on US1/US2 (fallback for file-watching)
- **User Story 4 (P2)**: Can start after Foundational (Phase 2) - independent of US1-3

### Within Each User Story

- Create structs and types first
- Then implement methods
- Then integrate with existing code
- Story complete before moving to next

### Parallel Opportunities

- **Phase 2**: T004 and T005 can run in parallel (different aspects of watcher.go)
- **Phase 7**: All test tasks (T029-T032) can run in parallel (different test files)
- **US1 + US4**: Can be developed in parallel by different developers (different files)

---

## Parallel Example: Foundational Phase

```bash
# Launch these in parallel (different concerns in same file):
Task: "Create MonitoringMode type constants in internal/daemon/watcher.go"
Task: "Create Debouncer struct in internal/daemon/watcher.go"
```

## Parallel Example: Polish Phase

```bash
# Launch all unit test tasks in parallel:
Task: "Add unit tests for Debouncer in internal/daemon/watcher_test.go"
Task: "Add unit tests for FileWatcher in internal/daemon/watcher_test.go"
Task: "Add unit tests for UpdateLastReadAt in internal/mail/recipients_test.go"
Task: "Add unit tests for last-read tracking in internal/cli/receive_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (add fsnotify)
2. Complete Phase 2: Foundational (Debouncer, types)
3. Complete Phase 3: User Story 1 (file-watching for mailboxes)
4. **STOP and VALIDATE**: Test mailbox notification < 2 seconds
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test mailbox watching → Demo (MVP!)
3. Add User Story 2 → Test status change watching → Demo
4. Add User Story 3 → Test fallback behavior → Demo
5. Add User Story 4 → Test last-read tracking → Demo
6. Complete Polish phase → Production ready

### Recommended Order

Since US2 extends US1's FileWatcher.Run(), recommended sequential order:
1. Setup → Foundational
2. US1 (mailbox watching)
3. US2 (recipients.jsonl watching)
4. US3 (fallback behavior)
5. US4 (last-read tracking) - can be done in parallel with US3
6. Polish

---

## Notes

- [P] tasks = different files or independent concerns, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- FR references map to functional requirements in spec.md

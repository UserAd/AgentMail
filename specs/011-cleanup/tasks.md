# Tasks: Cleanup Command

**Input**: Design documents from `/specs/011-cleanup/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, quickstart.md

**Tests**: Tests are included as this project has an 80% coverage requirement per constitution.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Go CLI**: `cmd/agentmail/`, `internal/cli/`, `internal/mail/`, `internal/tmux/`
- Tests colocated with source: `*_test.go` files

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Foundation for cleanup command - Message struct modification

- [X] T001 Add `CreatedAt time.Time` field with `json:"created_at,omitempty"` to Message struct in internal/mail/message.go
- [X] T002 Update Append function to set `CreatedAt` to `time.Now()` when creating new messages in internal/mail/mailbox.go
- [X] T003 [P] Add unit tests for Message serialization with and without CreatedAt in internal/mail/message_test.go

**Checkpoint**: Message struct ready with timestamp support

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core cleanup infrastructure that all user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Define CleanupOptions struct (StaleHours, DeliveredHours, DryRun) in internal/cli/cleanup.go
- [X] T005 Define CleanupResult struct (RecipientsRemoved, OfflineRemoved, StaleRemoved, MessagesRemoved, MailboxesRemoved, FilesSkipped) in internal/cli/cleanup.go
- [X] T006 Implement non-blocking file lock helper with 1-second timeout in internal/mail/mailbox.go
- [X] T007 [P] Add tests for non-blocking lock timeout behavior in internal/mail/mailbox_test.go
- [X] T008 Create Cleanup function signature and stub in internal/cli/cleanup.go
- [X] T009 Register cleanup subcommand with flags (--stale-hours, --delivered-hours, --dry-run) in cmd/agentmail/main.go

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Remove Offline Recipients (Priority: P1) 🎯 MVP

**Goal**: Remove recipients from recipients.jsonl whose tmux windows no longer exist

**Independent Test**: Create recipient entries for non-existent windows, run cleanup, verify removal

### Tests for User Story 1

- [X] T010 [P] [US1] Test offline recipient removal when window doesn't exist in internal/cli/cleanup_test.go
- [X] T011 [P] [US1] Test retention of recipients whose windows still exist in internal/cli/cleanup_test.go
- [X] T012 [P] [US1] Test cleanup completes successfully when recipients.jsonl is empty or missing in internal/cli/cleanup_test.go
- [X] T013 [P] [US1] Test non-tmux environment skips offline check with warning in internal/cli/cleanup_test.go

### Implementation for User Story 1

- [X] T014 [US1] Implement CleanOfflineRecipients function in internal/mail/recipients.go that compares recipients against tmux.ListWindows()
- [X] T015 [US1] Add offline recipient cleanup logic to Cleanup function in internal/cli/cleanup.go
- [X] T016 [US1] Handle non-tmux environment gracefully (skip offline check, warn, continue) in internal/cli/cleanup.go
- [X] T017 [US1] Track and return OfflineRemoved count in CleanupResult

**Checkpoint**: User Story 1 complete - offline recipient cleanup works independently

---

## Phase 4: User Story 2 - Remove Stale Recipients (Priority: P1)

**Goal**: Remove recipients whose `updated_at` is older than configured threshold (default 48h)

**Independent Test**: Create recipients with old timestamps, run cleanup, verify removal based on threshold

### Tests for User Story 2

- [X] T018 [P] [US2] Test stale recipient removal with default 48-hour threshold in internal/cli/cleanup_test.go
- [X] T019 [P] [US2] Test retention of recently updated recipients in internal/cli/cleanup_test.go
- [X] T020 [P] [US2] Test custom --stale-hours flag (e.g., 24h) in internal/cli/cleanup_test.go

### Implementation for User Story 2

- [X] T021 [US2] Extend existing CleanStaleStates function in internal/mail/recipients.go to accept configurable threshold
- [X] T022 [US2] Add stale recipient cleanup logic to Cleanup function using StaleHours option in internal/cli/cleanup.go
- [X] T023 [US2] Track and return StaleRemoved count in CleanupResult (distinct from OfflineRemoved)

**Checkpoint**: User Stories 1 AND 2 complete - both recipient cleanup types work

---

## Phase 5: User Story 3 - Remove Old Delivered Messages (Priority: P2)

**Goal**: Remove messages with `read_flag: true` older than threshold (default 2h), never delete unread

**Independent Test**: Create mailbox with mix of read/unread and old/new messages, verify only old read messages removed

### Tests for User Story 3

- [X] T024 [P] [US3] Test removal of old read messages (read_flag: true, created_at > 2h ago) in internal/cli/cleanup_test.go
- [X] T025 [P] [US3] Test retention of unread messages regardless of age in internal/cli/cleanup_test.go
- [X] T026 [P] [US3] Test retention of recent read messages (created_at within threshold) in internal/cli/cleanup_test.go
- [X] T027 [P] [US3] Test custom --delivered-hours flag in internal/cli/cleanup_test.go
- [X] T028 [P] [US3] Test messages without created_at field are skipped (not deleted) in internal/cli/cleanup_test.go

### Implementation for User Story 3

- [X] T029 [US3] Implement CleanOldMessages function in internal/mail/mailbox.go that filters messages by read_flag and created_at
- [X] T030 [US3] Iterate all mailbox files in .agentmail/mailboxes/ and apply message cleanup in internal/cli/cleanup.go
- [X] T031 [US3] Track and return MessagesRemoved count in CleanupResult
- [X] T032 [US3] Handle locked mailbox files (skip with warning, increment FilesSkipped) in internal/cli/cleanup.go

**Checkpoint**: User Story 3 complete - message cleanup works independently

---

## Phase 6: User Story 4 - Remove Empty Mailboxes (Priority: P3)

**Goal**: Delete mailbox files that contain zero messages after cleanup

**Independent Test**: Create empty .jsonl files, run cleanup, verify deletion

### Tests for User Story 4

- [X] T033 [P] [US4] Test removal of empty mailbox files in internal/cli/cleanup_test.go
- [X] T034 [P] [US4] Test retention of non-empty mailbox files in internal/cli/cleanup_test.go
- [X] T035 [P] [US4] Test cleanup succeeds when mailboxes directory doesn't exist in internal/cli/cleanup_test.go

### Implementation for User Story 4

- [X] T036 [US4] Implement RemoveEmptyMailboxes function in internal/mail/mailbox.go
- [X] T037 [US4] Add empty mailbox removal to Cleanup function (after message cleanup) in internal/cli/cleanup.go
- [X] T038 [US4] Track and return MailboxesRemoved count in CleanupResult

**Checkpoint**: User Story 4 complete - empty mailbox cleanup works

---

## Phase 7: User Story 5 & 6 - Exclusions (Priority: P3)

**Goal**: Ensure cleanup is NOT exposed in onboarding or MCP tools

**Independent Test**: Run `agentmail onboard` and list MCP tools, verify no cleanup reference

### Tests for User Stories 5 & 6

- [X] T039 [P] [US5] Test that `agentmail onboard` output does not contain "cleanup" in internal/cli/onboard_test.go
- [X] T040 [P] [US6] Test that MCP tools list does not include cleanup in internal/mcp/tools_test.go

### Implementation for User Stories 5 & 6

- [X] T041 [US5] Verify onboard.go does not reference cleanup command (no implementation change expected)
- [X] T042 [US6] Verify tools.go does not register a cleanup tool (no implementation change expected)

**Checkpoint**: Exclusion requirements verified

---

## Phase 8: Output & Dry-Run

**Goal**: Summary output and dry-run mode per FR-013 and FR-015

### Tests

- [X] T043 [P] Test cleanup outputs summary with correct counts (recipients, messages, mailboxes) in internal/cli/cleanup_test.go
- [X] T044 [P] Test dry-run mode reports counts without making changes in internal/cli/cleanup_test.go
- [X] T045 [P] Test warning output when files are skipped due to locking in internal/cli/cleanup_test.go

### Implementation

- [X] T046 Implement summary output formatting in Cleanup function (format: "Cleanup complete: Recipients removed: X...") in internal/cli/cleanup.go
- [X] T047 Implement dry-run mode that collects counts without deletions in internal/cli/cleanup.go
- [X] T048 Implement warning output for skipped locked files in internal/cli/cleanup.go

**Checkpoint**: Full cleanup command functionality complete

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup

- [ ] T049 Run `go test -v -race -cover ./...` and verify >= 80% coverage on new code
- [ ] T050 Run `go vet ./...` and fix any issues
- [ ] T051 Run `go fmt ./...` and commit any formatting changes
- [ ] T052 Run `gosec ./...` and verify no security issues
- [ ] T053 Manual test: Run quickstart.md examples and verify expected output
- [ ] T054 Update CLAUDE.md with cleanup command reference in Active Technologies section

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion
- **User Stories (Phase 3-7)**: All depend on Foundational phase completion
- **Output & Dry-Run (Phase 8)**: Depends on all user story phases
- **Polish (Phase 9)**: Depends on all phases complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Phase 2 - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Phase 2 - Independent of US1
- **User Story 3 (P2)**: Can start after Phase 2 - Independent of US1/US2
- **User Story 4 (P3)**: Should run AFTER US3 (message cleanup may empty mailboxes)
- **User Stories 5&6 (P3)**: Can run anytime after Phase 2 - Verification only

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Implementation follows test-first approach
- Story complete before moving to next

### Parallel Opportunities

**Phase 1 (Setup)**:
```
T001, T002, T003 can run in parallel (different files)
```

**Phase 2 (Foundational)**:
```
T004, T005 in parallel (same file - types only)
T006, T007 in parallel with T004/T005 (different file)
```

**User Story Tests** (within each story):
```
All test tasks marked [P] can run in parallel
```

**Cross-Story Parallelism**:
```
US1, US2, US3 can run in parallel after Phase 2
US5, US6 can run in parallel with any story
```

---

## Implementation Strategy

### MVP First (User Stories 1 & 2)

1. Complete Phase 1: Setup (Message timestamp)
2. Complete Phase 2: Foundational (types, lock helper, command registration)
3. Complete Phase 3: User Story 1 (offline recipients)
4. Complete Phase 4: User Story 2 (stale recipients)
5. **STOP and VALIDATE**: Test cleanup with recipients only
6. Deploy if recipient cleanup is sufficient

### Incremental Delivery

1. MVP → Recipient cleanup (US1 + US2)
2. Add Message cleanup (US3)
3. Add Empty mailbox cleanup (US4)
4. Add Output polish (Phase 8)
5. Each increment adds value without breaking previous

### Parallel Team Strategy

With multiple developers:
1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 + 2 (recipient cleanup)
   - Developer B: User Story 3 (message cleanup)
   - Developer C: User Story 4 + 5 + 6 (mailbox + verification)
3. Integrate at Phase 8

---

## Task Summary

| Phase | Task Count | Description |
|-------|------------|-------------|
| Phase 1 (Setup) | 3 | Message struct, timestamp |
| Phase 2 (Foundational) | 6 | Types, locking, command |
| Phase 3 (US1) | 8 | Offline recipients |
| Phase 4 (US2) | 6 | Stale recipients |
| Phase 5 (US3) | 9 | Old messages |
| Phase 6 (US4) | 6 | Empty mailboxes |
| Phase 7 (US5&6) | 4 | Exclusions |
| Phase 8 (Output) | 6 | Summary, dry-run |
| Phase 9 (Polish) | 6 | Quality gates |
| **Total** | **54** | |

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Constitution requires 80% test coverage - tests included in each phase
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently

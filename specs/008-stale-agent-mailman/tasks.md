# Tasks: Stale Agent Notification Support

**Input**: Design documents from `/specs/008-stale-agent-mailman/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md

**Tests**: Tests ARE requested per constitution (80% coverage gate, SC-006, SC-007). TDD approach will be used.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Single project**: Go project at repository root
- **Source files**: `internal/daemon/loop.go`, `internal/daemon/daemon.go`
- **Test files**: `internal/daemon/loop_test.go`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add foundational types required for all user stories

- [X] T001 Add StatelessNotifyInterval constant (60s) in internal/daemon/loop.go
- [X] T002 Add StatelessTracker struct with mu, lastNotified, notifyInterval fields in internal/daemon/loop.go
- [X] T003 Add StatelessTracker field to LoopOptions struct in internal/daemon/loop.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core StatelessTracker methods that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Tests (write first, ensure they FAIL)

- [X] T004 [P] Add TestStatelessTracker_ShouldNotify_FirstTime - returns true for new window in internal/daemon/loop_test.go (FR-010)
- [X] T005 [P] Add TestStatelessTracker_ShouldNotify_BeforeInterval - returns false before 60s in internal/daemon/loop_test.go (FR-005)
- [X] T006 [P] Add TestStatelessTracker_ShouldNotify_AfterInterval - returns true after 60s in internal/daemon/loop_test.go (FR-005)
- [X] T007 [P] Add TestStatelessTracker_MarkNotified - updates timestamp in internal/daemon/loop_test.go (FR-009)
- [X] T008 [P] Add TestStatelessTracker_Cleanup - removes stale entries in internal/daemon/loop_test.go (FR-011)
- [X] T009 [P] Add TestStatelessTracker_ThreadSafety - concurrent access is safe in internal/daemon/loop_test.go (FR-013, SC-007)

### Implementation

- [X] T010 Implement NewStatelessTracker(interval time.Duration) constructor in internal/daemon/loop.go (FR-010)
- [X] T011 Implement ShouldNotify(window string) bool method in internal/daemon/loop.go (FR-005)
- [X] T012 Implement MarkNotified(window string) method in internal/daemon/loop.go (FR-009)
- [X] T013 Implement Cleanup(activeWindows []string) method in internal/daemon/loop.go (FR-011)

**Checkpoint**: All T004-T009 tests should now PASS. Foundation ready - user story implementation can now begin.

---

## Phase 3: User Story 1 - Stateless Agent Receives Mail Notifications (Priority: P1) 🎯 MVP

**Goal**: Agents without Claude hooks support receive periodic mail notifications at 60-second intervals

**Independent Test**: Send mail to an agent without status registration and verify notifications arrive every 60 seconds until mail is read

**Implements**: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-012
**Validates**: SC-001, SC-002

### Tests for User Story 1 (write first, ensure they FAIL)

- [ ] T014 [P] [US1] Add TestStatelessNotification_AgentWithMailboxNoState in internal/daemon/loop_test.go (FR-003, FR-004)
- [ ] T015 [P] [US1] Add TestStatelessNotification_RespectInterval - first notification immediate, subsequent at 60s intervals in internal/daemon/loop_test.go (FR-004, SC-002)
- [ ] T016 [P] [US1] Add TestStatelessNotification_NoUnreadMessages in internal/daemon/loop_test.go (FR-006)
- [ ] T017 [P] [US1] Add TestStatelessNotification_MultipleAgents in internal/daemon/loop_test.go

### Implementation for User Story 1

- [ ] T018 [US1] In CheckAndNotifyWithNotifier, build statedSet from recipients slice in internal/daemon/loop.go (FR-002)
- [ ] T019 [US1] In CheckAndNotifyWithNotifier, call mail.ListMailboxRecipients() after Phase 1 logic in internal/daemon/loop.go (FR-001)
- [ ] T020 [US1] In CheckAndNotifyWithNotifier, iterate mailbox recipients and skip those in statedSet in internal/daemon/loop.go (FR-003)
- [ ] T021 [US1] For each stateless agent, check FindUnread() and skip if empty in internal/daemon/loop.go (FR-006)
- [ ] T022 [US1] For each stateless agent, check tracker.ShouldNotify() and send notification if due in internal/daemon/loop.go (FR-004, FR-005)
- [ ] T023 [US1] Call tracker.MarkNotified() after successful notification in internal/daemon/loop.go
- [ ] T024 [US1] Call tracker.Cleanup() with mailbox list at end of Phase 2 logic in internal/daemon/loop.go (FR-011)
- [ ] T025 [US1] Initialize StatelessTracker in runForeground() and pass to LoopOptions in internal/daemon/daemon.go (FR-010, FR-012)

**Checkpoint**: User Story 1 complete. Tests T014-T017 should PASS. Stateless agents now receive periodic notifications.

---

## Phase 4: User Story 2 - Stated Agents Take Precedence (Priority: P2)

**Goal**: When a stateless agent registers status, it immediately switches to stated agent notification logic

**Independent Test**: Start with stateless agent receiving periodic notifications, then register status and verify notification behavior changes

**Implements**: FR-007, FR-008
**Validates**: SC-003, SC-005

### Tests for User Story 2 (write first, ensure they FAIL)

- [ ] T026 [P] [US2] Add TestStatelessNotification_StatedAgentTakesPrecedence in internal/daemon/loop_test.go (FR-007, SC-003)
- [ ] T027 [P] [US2] Add TestStatelessNotification_TransitionToStated in internal/daemon/loop_test.go (FR-008)

### Implementation for User Story 2

- [ ] T028 [US2] Verify statedSet check correctly excludes agents with recipients.jsonl entries in internal/daemon/loop.go (FR-007)
- [ ] T029 [US2] Verify existing stated agent tests still pass (no modifications needed) (SC-005)

**Checkpoint**: User Story 2 complete. Tests T026-T027 should PASS. Stated agents take precedence over stateless logic.

---

## Phase 5: User Story 3 - Daemon Restart Clears Tracking State (Priority: P3)

**Goal**: On daemon restart, all stateless agents with unread mail become immediately eligible for notification

**Independent Test**: Send mail to stateless agent, wait for notification, restart daemon, verify immediate notification

**Implements**: FR-010
**Validates**: SC-004

### Tests for User Story 3

- [ ] T030 [US3] Add TestStatelessNotification_DaemonRestart_ImmediateEligibility in internal/daemon/loop_test.go (SC-004)

### Implementation for User Story 3

- [ ] T031 [US3] Verify NewStatelessTracker initializes empty map (already done in T010, verify behavior) in internal/daemon/loop.go

**Checkpoint**: User Story 3 complete. Test T030 should PASS. Daemon restart behavior works correctly.

---

## Phase 6: Error Handling

**Purpose**: Robust error handling for all failure modes

**Implements**: FR-014, FR-015, FR-016, FR-017

### Tests for Error Handling

- [ ] T032 [P] Add TestStatelessNotification_MailboxDirReadError in internal/daemon/loop_test.go (FR-014)
- [ ] T033 [P] Add TestStatelessNotification_NotifyFailure in internal/daemon/loop_test.go (FR-015)
- [ ] T034 [P] Add TestStatelessNotification_MailboxFileReadError in internal/daemon/loop_test.go (FR-016)
- [ ] T035 [P] Add TestStatelessNotification_RecipientsReadError in internal/daemon/loop_test.go (FR-017)

### Implementation for Error Handling

- [ ] T036 Add error handling: log and continue if ListMailboxRecipients() fails in internal/daemon/loop.go (FR-014)
- [ ] T037 Add error handling: skip MarkNotified if notify() fails in internal/daemon/loop.go (FR-015)
- [ ] T038 Add error handling: skip agent if FindUnread() fails, log warning in internal/daemon/loop.go (FR-016)
- [ ] T039 Add error handling: if ReadAllRecipients() fails, treat all as stateless in internal/daemon/loop.go (FR-017)

**Checkpoint**: Error handling complete. Tests T032-T035 should PASS.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Quality gates and verification

- [ ] T040 Run go test -v -race ./internal/daemon/... and verify zero race conditions (SC-007)
- [ ] T041 Run go test -cover ./internal/daemon/... and verify >= 80% coverage (SC-006)
- [ ] T042 Run go vet ./... and verify no errors
- [ ] T043 Run go fmt ./... and verify no changes
- [ ] T044 Run existing daemon tests and verify all pass (SC-005)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User stories can then proceed in priority order (P1 → P2 → P3)
  - US2 and US3 are relatively small and can be done after US1
- **Error Handling (Phase 6)**: Can be done in parallel with or after user stories
- **Polish (Phase 7)**: Depends on all implementation being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Verifies existing behavior
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Verifies initialization behavior

### Within Each Phase

- Tests MUST be written and FAIL before implementation
- Implementation tasks are sequential within each phase
- Verify tests PASS after implementation

### Parallel Opportunities

- All Foundational tests (T004-T009) can run in parallel
- All User Story 1 tests (T014-T017) can run in parallel
- All User Story 2 tests (T026-T027) can run in parallel
- All Error Handling tests (T032-T035) can run in parallel
- Error Handling (Phase 6) can run in parallel with User Stories 2 and 3

---

## Parallel Example: Foundational Tests

```bash
# Launch all foundational tests together:
Task: "Add TestStatelessTracker_ShouldNotify_FirstTime in internal/daemon/loop_test.go"
Task: "Add TestStatelessTracker_ShouldNotify_BeforeInterval in internal/daemon/loop_test.go"
Task: "Add TestStatelessTracker_ShouldNotify_AfterInterval in internal/daemon/loop_test.go"
Task: "Add TestStatelessTracker_MarkNotified in internal/daemon/loop_test.go"
Task: "Add TestStatelessTracker_Cleanup in internal/daemon/loop_test.go"
Task: "Add TestStatelessTracker_ThreadSafety in internal/daemon/loop_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T013)
3. Complete Phase 3: User Story 1 (T014-T025)
4. **STOP and VALIDATE**: Test stateless notifications independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Core functionality working (MVP!)
3. Add User Story 2 → Test independently → State precedence working
4. Add User Story 3 → Test independently → Restart behavior working
5. Add Error Handling → Test independently → Robust error handling
6. Polish → Quality gates pass → Ready for merge

### Full Implementation Order

```
T001 → T002 → T003 (Setup)
  ↓
T004-T009 (Foundational tests - parallel)
  ↓
T010 → T011 → T012 → T013 (Foundational implementation)
  ↓
T014-T017 (US1 tests - parallel)
  ↓
T018 → T019 → T020 → T021 → T022 → T023 → T024 → T025 (US1 implementation)
  ↓
T026-T027 (US2 tests - parallel)
  ↓
T028 → T029 (US2 implementation)
  ↓
T030 → T031 (US3)
  ↓
T032-T035 (Error tests - parallel)
  ↓
T036 → T037 → T038 → T039 (Error implementation)
  ↓
T040 → T041 → T042 → T043 → T044 (Polish)
```

---

## Notes

- [P] tasks = different files or independent test functions, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- TDD approach: verify tests fail before implementing, then pass after
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- All tasks target internal/daemon/ package only (loop.go, loop_test.go, daemon.go)

---

## Summary

| Metric | Count |
|--------|-------|
| **Total Tasks** | 44 |
| **Setup Tasks** | 3 |
| **Foundational Tasks** | 10 |
| **User Story 1 Tasks** | 12 |
| **User Story 2 Tasks** | 4 |
| **User Story 3 Tasks** | 2 |
| **Error Handling Tasks** | 8 |
| **Polish Tasks** | 5 |
| **Total Test Tasks** | 16 |
| **Parallel Opportunities** | 18 tasks marked [P] |

### Requirements Coverage

All 17 functional requirements (FR-001 through FR-017) are covered by tasks.
All 7 success criteria (SC-001 through SC-007) are validated by tests.

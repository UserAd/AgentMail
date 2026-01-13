# Tasks: Storage Directory Restructure

**Input**: Design documents from `/specs/007-storage-restructure/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Constitution requires 80% coverage. Test updates are included to maintain existing coverage.

**Organization**: Tasks grouped by user story. Note: US1-US3 are all P1 priority but are organized as foundational changes that enable all messaging, daemon, and recipients functionality.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Go project**: `internal/`, `cmd/` at repository root
- Tests co-located with source as `*_test.go` files

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: No setup required - existing Go project with all dependencies in place

*No tasks in this phase - project already initialized*

---

## Phase 2: Foundational (Path Constants)

**Purpose**: Update core path constants that ALL user stories depend on

**⚠️ CRITICAL**: These constant changes are prerequisites for all user story functionality

- [x] T001 Add `RootDir` constant with value `.agentmail` in internal/mail/mailbox.go
- [x] T002 Update `MailDir` constant from `.git/mail` to `.agentmail/mailboxes` in internal/mail/mailbox.go
- [x] T003 Update `RecipientsFile` constant from `.git/mail-recipients.jsonl` to `.agentmail/recipients.jsonl` in internal/mail/recipients.go
- [x] T004 Update `EnsureMailDir` function to create both RootDir and MailDir in internal/mail/mailbox.go
- [x] T005 Update `PIDFilePath` function to use `mail.RootDir` instead of `mail.MailDir` in internal/daemon/daemon.go

**Checkpoint**: Core path constants updated - user story verification can begin

---

## Phase 3: User Story 1 - Send and Receive Messages (Priority: P1) 🎯 MVP

**Goal**: Messages stored in `.agentmail/mailboxes/<recipient>.jsonl` instead of `.git/mail/<recipient>.jsonl`

**Independent Test**: Send a message and verify it creates `.agentmail/mailboxes/<recipient>.jsonl`, then receive and verify message displayed

### Test Updates for User Story 1

- [x] T006 [P] [US1] Update path assertions in internal/mail/mailbox_test.go to expect `.agentmail/mailboxes/`
- [x] T007 [P] [US1] Update path assertions in internal/cli/send_test.go to expect new directory structure
- [x] T008 [P] [US1] Update path assertions in internal/cli/receive_test.go to expect new mailbox paths

### Verification for User Story 1

- [x] T009 [US1] Run `go test ./internal/mail/... ./internal/cli/...` to verify send/receive tests pass

**Checkpoint**: Send and receive functionality works with new `.agentmail/mailboxes/` path

---

## Phase 4: User Story 2 - Mailman Daemon PID Location (Priority: P1)

**Goal**: PID file stored at `.agentmail/mailman.pid` instead of `.git/mail/mailman.pid`

**Independent Test**: Start daemon and verify `.agentmail/mailman.pid` created, stop and verify removed

### Test Updates for User Story 2

- [x] T010 [P] [US2] Update path assertions in internal/daemon/daemon_test.go to expect `.agentmail/mailman.pid`
- [x] T011 [P] [US2] Update path assertions in internal/daemon/loop_test.go if any reference PID paths
- [x] T012 [P] [US2] Update path assertions in internal/cli/mailman_test.go to expect new PID location

### Verification for User Story 2

- [x] T013 [US2] Run `go test ./internal/daemon/... ./internal/cli/...` to verify daemon tests pass

**Checkpoint**: Mailman daemon start/stop/status works with new `.agentmail/mailman.pid` path

---

## Phase 5: User Story 3 - Recipients State Location (Priority: P1)

**Goal**: Recipients state stored at `.agentmail/recipients.jsonl` instead of `.git/mail-recipients.jsonl`

**Independent Test**: Run recipients command and verify state read from `.agentmail/recipients.jsonl`

### Test Updates for User Story 3

- [x] T014 [P] [US3] Update path assertions in internal/mail/recipients_test.go to expect `.agentmail/recipients.jsonl`
- [x] T015 [P] [US3] Update path assertions in internal/cli/recipients_test.go to expect new path
- [x] T016 [P] [US3] Update path assertions in internal/cli/status_test.go if any reference recipients path

### Verification for User Story 3

- [x] T017 [US3] Run `go test ./internal/mail/... ./internal/cli/...` to verify recipients tests pass

**Checkpoint**: Recipients functionality works with new `.agentmail/recipients.jsonl` path

---

## Phase 6: User Story 4 - Migration Compatibility (Priority: P2)

**Goal**: Application ignores old `.git/mail/` data, uses only new `.agentmail/` location

**Independent Test**: Create data in both locations, verify only `.agentmail/` data is read

### Implementation for User Story 4

- [x] T018 [US4] Verify no code references old paths (grep for `.git/mail` in internal/)
- [x] T019 [US4] Run integration tests in internal/cli/integration_test.go to confirm end-to-end behavior

**Checkpoint**: Application correctly ignores legacy storage locations

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, quality gates, and final validation

### Documentation Updates

- [x] T020 [P] Update storage location section in README.md (change `.git/mail/` to `.agentmail/`)
- [x] T021 [P] Update Message Storage section in CLAUDE.md (change `.git/mail/` to `.agentmail/mailboxes/`)
- [x] T022 [P] Update Technology Constraints in .specify/memory/constitution.md (change storage path)

### Quality Gates

- [x] T023 Run `go fmt ./...` and verify no formatting changes needed
- [x] T024 Run `go vet ./...` and verify no static analysis errors
- [x] T025 Run `go test -cover ./...` and verify coverage >= 80%
- [x] T026 Run full test suite `go test -race ./...` to verify no race conditions

### Final Validation

- [x] T027 Run quickstart.md manual verification steps in a tmux session
- [x] T028 Verify `.gitignore` does NOT auto-modify (FR-015 compliance)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Skipped - project already initialized
- **Foundational (Phase 2)**: No dependencies - BLOCKS all user stories (T001-T005 must complete first)
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - US1, US2, US3 can proceed in parallel after Phase 2
  - US4 can proceed after Phase 2 (no code changes, just verification)
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

```
Phase 2 (Foundational)
    ├── T001 RootDir constant
    ├── T002 MailDir constant
    ├── T003 RecipientsFile constant
    ├── T004 EnsureMailDir function
    └── T005 PIDFilePath function
         │
         ▼ ALL MUST COMPLETE
    ┌────┴────┬────────┬────────┐
    ▼         ▼        ▼        ▼
  US1       US2      US3      US4
 (send/    (daemon) (recip)  (verify)
  recv)
    │         │        │        │
    └────┬────┴────────┴────────┘
         ▼
    Phase 7 (Polish)
```

### Within Each User Story

- Test updates marked [P] can run in parallel
- Verification task runs after all test updates

### Parallel Opportunities

**Phase 2** (Sequential - same files):
- T001 → T002 → T004 (all in mailbox.go)
- T003 (recipients.go - parallel with above)
- T005 (daemon.go - parallel with above)

**Phase 3-6** (Parallel across stories):
- All [P] test updates can run in parallel
- Different user stories can proceed in parallel

**Phase 7** (Parallel documentation):
- T020, T021, T022 can run in parallel (different files)
- T023-T026 are sequential (quality gates)

---

## Parallel Example: Test Updates

```bash
# Launch all Phase 3-5 test updates in parallel:
Task: "Update path assertions in internal/mail/mailbox_test.go"
Task: "Update path assertions in internal/cli/send_test.go"
Task: "Update path assertions in internal/cli/receive_test.go"
Task: "Update path assertions in internal/daemon/daemon_test.go"
Task: "Update path assertions in internal/cli/mailman_test.go"
Task: "Update path assertions in internal/mail/recipients_test.go"
Task: "Update path assertions in internal/cli/recipients_test.go"
```

---

## Implementation Strategy

### MVP First (Phase 2 + User Story 1)

1. Complete Phase 2: Foundational constants (T001-T005)
2. Complete Phase 3: User Story 1 test updates (T006-T008)
3. Run verification (T009)
4. **STOP and VALIDATE**: `agentmail send` and `agentmail receive` work with new paths
5. Deploy if ready - core messaging works

### Incremental Delivery

1. Phase 2 → Constants updated
2. Add US1 → Test send/receive → Messaging works (MVP!)
3. Add US2 → Test daemon → Daemon works
4. Add US3 → Test recipients → Recipients works
5. Add US4 → Verify migration → Legacy data ignored
6. Phase 7 → Polish → Documentation updated, quality gates pass

### Recommended Sequence (Single Developer)

```
T001 → T002 → T004 (mailbox.go changes)
    ↓ (parallel)
T003 (recipients.go)
    ↓ (parallel)
T005 (daemon.go)
    ↓
T006-T008 (US1 tests - parallel)
    ↓
T009 (verify US1)
    ↓
T010-T012 (US2 tests - parallel)
    ↓
T013 (verify US2)
    ↓
T014-T016 (US3 tests - parallel)
    ↓
T017 (verify US3)
    ↓
T018-T019 (US4 verification)
    ↓
T020-T022 (docs - parallel)
    ↓
T023-T028 (quality gates - sequential)
```

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Constitution requires 80% test coverage - test updates maintain this
- No new test files needed - only update assertions in existing tests
- FR-014/FR-015 compliance verified in T028 (no warnings, no .gitignore changes)
- Total: 28 tasks (5 foundational, 12 test updates, 4 verifications, 3 docs, 4 quality gates)

# Tasks: Tmux Pane Addressing

**Input**: Design documents from `/specs/tmux-pane-addressing/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/cli.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Foundational — Address Parsing & Tmux Layer

**Purpose**: Create the PaneAddress type and all new tmux query functions. This is the foundation that ALL user stories depend on.

**CRITICAL**: No user story work can begin until this phase is complete.

- [ ] T001 Create PaneAddress struct with ParseAddress(), FormatAddress(), SanitizeForFilename(), and UnsanitizeFromFilename() functions in internal/tmux/address.go. PaneAddress has fields: Session (string), Window (string), Pane (int), Full (string). ParseAddress(input, currentSession) handles three forms: full (contains ":" not at position 0), medium (starts with ":"), short (no ":" — entire input is window name, even if it contains dots). FormatAddress returns "session:window.pane". SanitizeForFilename percent-encodes structural characters: "%" → "%25", ":" → "%3A", "." (pane separator only, the last dot) → "%2E" (e.g., "mysession:editor.0" → "mysession%3Aeditor%2E0"). UnsanitizeFromFilename reverses the encoding to recover canonical addresses. Include comprehensive unit tests in internal/tmux/address_test.go covering: full/medium/short parsing, window names containing dots (e.g., "my.app", "logs.1"), medium form ":editor.1", invalid formats (e.g., ".1", empty string, ":" alone), sanitization roundtrips proving no collisions (e.g., "s_a:w.0" vs "s:a_w.0" produce different filenames), names containing "%" characters.

- [ ] T002 Add GetCurrentPaneAddress(), GetCurrentSession(), ListPanes(), ListAllPanes(), and PaneExists() functions to internal/tmux/tmux.go. GetCurrentPaneAddress() runs `tmux display-message -t $TMUX_PANE -p '#{session_name}:#{window_name}.#{pane_index}'` and returns the full address string. GetCurrentSession() runs `tmux display-message -t $TMUX_PANE -p '#{session_name}'`. ListPanes() runs `tmux list-panes -s -F '#{session_name}:#{window_name}.#{pane_index}'` (current session only — this is the default discovery scope). ListAllPanes() runs `tmux list-panes -a -F '...'` (all sessions — not exposed in commands/MCP, available for future use). PaneExists(address) runs `tmux display-message -t '<address>' -p '#{pane_id}'` and returns true if exit code is 0. Note: PaneExists and ListAllPanes do NOT require TMUX_PANE validation since they work with explicit targets. GetCurrentPaneAddress, GetCurrentSession, and ListPanes DO require TMUX_PANE via existing GetCurrentPaneID(). Add tests in internal/tmux/tmux_test.go.

- [ ] T003 Update SendKeys() and SendEnter() in internal/tmux/sendkeys.go to accept full pane addresses (e.g., "mysession:editor.1") as the target parameter instead of just window names. The tmux `-t` flag already supports this format natively. Update tests if they exist.

- [ ] T004 Update internal/mail/mailbox.go path construction: modify all calls to safePath() that build mailbox file paths to use SanitizeForFilename() on pane addresses. Currently paths are built as `msg.To + ".jsonl"` — change to `tmux.SanitizeForFilename(msg.To) + ".jsonl"`. This affects Append(), ReadAll(), FindUnread(), MarkAsRead(), WriteAll(), CleanOldMessages(). IMPORTANT type boundary: ListMailboxRecipients() MUST return canonical pane addresses (by calling UnsanitizeFromFilename on each filename), NOT raw percent-encoded filenames. All public APIs deal in canonical addresses; percent-encoding is internal to the storage layer. Update tests in internal/mail/mailbox_test.go to use pane-address-style recipient strings and verify ListMailboxRecipients returns decoded addresses.

- [ ] T005 Update internal/mail/recipients.go: change all functions that key on window name to use full pane addresses instead. ReadAllRecipients(), UpdateRecipientState(), WriteAllRecipients(), CleanStaleStates(), SetNotifiedAt(), UpdateLastReadAt(), and CleanOfflineRecipients() all use the `Recipient` field which was previously a window name — now it will be a full pane address. CleanOfflineRecipients() must compare recipient pane addresses against valid pane addresses (passed by caller) instead of window names. Update tests in internal/mail/recipients_test.go.

- [ ] T006 Update internal/mail/ignore.go: modify LoadIgnoreList() to return a map that supports pane-level matching. Add a new IsIgnored(address string, ignoreList map[string]bool, currentSession string) function that checks: (1) exact match on full pane address (e.g., "mysession:editor.1"), (2) match on medium form ":window.pane" resolved against currentSession, (3) match on window name part only (short form matches all panes of that window). This ensures ".agentmailignore" entries like "editor" ignore all panes of the editor window, ":editor.1" ignores pane 1 in the current session, and "mysession:editor.1" ignores only that specific pane.

**Checkpoint**: Foundation ready — all address parsing, tmux queries, mail storage, and ignore matching use pane addresses. User story implementation can begin.

---

## Phase 2: User Story 1 — Send Message to a Specific Pane (Priority: P1) MVP

**Goal**: Agents can send messages to specific panes using full, medium, or short address formats.

**Independent Test**: Send a message from one pane to another using `agentmail send "session:window.pane" "message"` and verify the message is stored in the correct pane-specific mailbox file.

### Implementation for User Story 1

- [ ] T007 [US1] Update internal/cli/send.go: Replace GetCurrentWindow() call with GetCurrentPaneAddress() for sender identification. Replace WindowExists() recipient validation with address resolution logic: (1) call ParseAddress(recipient, currentSession) to determine address form, (2) for full/medium forms call PaneExists() to validate, (3) for short form call ListPanes() to find matching panes — if exactly one pane matches the window name resolve to full address, if multiple panes match return ambiguity error with suggestion listing all matching pane addresses. Note: medium form now requires leading ":" (e.g., ":editor.1"). Update self-send check to compare full pane addresses. Update ignore list check to use IsIgnored(). Add MockPaneAddress, MockSession, MockPanes fields to SendOptions for testing. Update tests in internal/cli/send_test.go with cases: full address send, medium address ":editor.1" send, short address single-pane resolve, short address multi-pane ambiguity error, dotted window name "my.app" treated as short form, invalid address rejection, self-send rejection, ignore list with pane addresses.

**Checkpoint**: `agentmail send` works with pane addresses. Messages are stored in pane-specific mailbox files.

---

## Phase 3: User Story 2 — Receive Messages Addressed to Current Pane (Priority: P1)

**Goal**: Agents receive only messages addressed to their specific pane.

**Independent Test**: Send a message to a pane address, run `agentmail receive` from that pane, verify correct message returned and pane isolation.

### Implementation for User Story 2

- [ ] T008 [US2] Update internal/cli/receive.go: Replace GetCurrentWindow() with GetCurrentPaneAddress() for receiver identification. The receiver's full pane address is used to look up the mailbox (FindUnread now uses sanitized pane address for file path). Update window existence validation to use PaneExists() on the receiver's own address. Update LastReadAt call to use pane address. Add MockPaneAddress to ReceiveOptions. Update tests in internal/cli/receive_test.go with cases: receive from pane-specific mailbox, pane isolation (messages to different pane not returned), no unread messages, hook mode with pane address.

**Checkpoint**: `agentmail receive` returns only messages for the calling pane.

---

## Phase 4: User Story 3 — Backward-Compatible Window-Only Addressing (Priority: P1)

**Goal**: Existing `agentmail send <window> "message"` commands continue to work for single-pane windows and produce clear errors for multi-pane windows.

**Independent Test**: Run `agentmail send editor "hello"` with editor having a single pane — should succeed. Run the same with editor having multiple panes — should return ambiguity error with suggestions.

### Implementation for User Story 3

- [ ] T009 [US3] Verify and add backward compatibility tests in internal/cli/send_test.go: Add specific test cases that verify short-form addressing behavior: (1) window with single pane resolves and delivers, (2) window with multiple panes returns ambiguity error message containing all pane addresses as suggestions, (3) medium form ":window.pane" resolves correctly with inferred session, (4) dotted window name "logs.1" is treated as short form (NOT medium form), (5) dotted window name "my.app" with single pane resolves correctly. The core logic was implemented in T007 — this task focuses on ensuring the acceptance scenarios from spec.md US3 are explicitly covered by tests. Also verify that the ambiguity error message format matches: "Ambiguous recipient: window '<name>' has N panes. Use <addr1>, <addr2>, ..."

**Checkpoint**: All backward-compatible addressing scenarios pass.

---

## Phase 5: User Story 4 — MCP Server Pane-Aware Operations (Priority: P2)

**Goal**: MCP server tools handle pane-level addressing for AI agent integration.

**Independent Test**: Call MCP send tool with a pane address, verify delivery. Call MCP list-recipients, verify pane-level entries.

### Implementation for User Story 4

- [ ] T010 [P] [US4] Update MCP tool schemas in internal/mcp/tools.go: Update the `send` tool description to document that `recipient` accepts pane addresses (full `session:window.pane`, medium `:window.pane`, short `window` forms). Update ListRecipientsResponse: replace `Window` field with `Address` field (string, canonical pane address) and add `IsCurrent` bool. This is a breaking change — no deprecated `Window` field. Update description strings for all tools to reference pane-level addressing.

- [ ] T011 [US4] Update MCP handlers in internal/mcp/handlers.go: (1) doSend() — replace GetCurrentWindow() with GetCurrentPaneAddress() for sender, replace WindowExists() with the same address resolution logic as CLI send (ParseAddress + PaneExists/ListPanes), update self-send and ignore checks. (2) doReceive() — replace GetCurrentWindow() with GetCurrentPaneAddress() for receiver. (3) doStatus() — replace GetCurrentWindow() with GetCurrentPaneAddress() for agent identity. (4) doListRecipients() — replace ListWindows() with ListPanes() (current session scope), use full pane addresses in response with Address field (replacing Window), mark current pane with IsCurrent. Add MockPaneAddress, MockSession, MockPanes to HandlerOptions. Update tests in internal/mcp/handlers_test.go with pane address test cases for all four handlers.

**Checkpoint**: All MCP tools work with pane addresses.

---

## Phase 6: User Story 5 — Claude Hooks Pane-Aware Polling (Priority: P2)

**Goal**: Hook mode correctly identifies current pane and only polls that pane's mailbox.

**Independent Test**: Run receive in hook mode from a specific pane, verify it only checks that pane's mailbox and ignores sibling panes.

### Implementation for User Story 5

- [ ] T012 [US5] Verify hook mode pane isolation in internal/cli/receive_test.go: The core hook mode logic already uses the same receive path updated in T008. Add explicit test cases that verify: (1) hook mode with message for current pane outputs notification to stderr and exits code 2, (2) hook mode with message only for a sibling pane (different pane in same window) returns exit code 0 with no output, (3) hook mode when not in tmux exits silently code 0. These tests confirm FR-012 and SC-005.

**Checkpoint**: Hook mode notifications are pane-isolated.

---

## Phase 7: User Story 6 — Recipient Listing Shows Pane Details (Priority: P2)

**Goal**: `agentmail recipients` lists all panes with full addresses.

**Independent Test**: Run `agentmail recipients` in a session with multi-pane windows, verify each pane listed separately with full address.

### Implementation for User Story 6

- [ ] T013 [US6] Update internal/cli/recipients.go: Replace ListWindows() with ListPanes() to get all pane addresses. Replace GetCurrentWindow() with GetCurrentPaneAddress() for marking current pane with "[you]". Update ignore list filtering to use IsIgnored() with full pane addresses. Output each pane address on its own line. Add MockPanes, MockPaneAddress to RecipientsOptions. Update tests in internal/cli/recipients_test.go with cases: multi-pane window shows separate entries, single-pane window shows full address, current pane marked with "[you]", ignored panes filtered, current pane shown even if in ignore list.

**Checkpoint**: `agentmail recipients` shows pane-level listing.

---

## Phase 8: Daemon & Cleanup — Cross-Cutting Pane Updates (Priority: P2)

**Goal**: The mailman daemon and cleanup command operate at pane granularity.

### Implementation

- [ ] T014 [P] Update internal/cli/status.go: Replace GetCurrentWindow() with GetCurrentPaneAddress() for agent identity. The rest of the status logic (UpdateRecipientState with status and resetNotified) is unchanged since recipients.go was already updated in T005 to key by pane address. Add MockPaneAddress to StatusOptions. Update tests in internal/cli/status_test.go.

- [ ] T015 [P] Update internal/cli/cleanup.go: (1) Phase 1 (offline recipients) — replace ListWindows() with ListPanes() for valid pane list, pass pane addresses to CleanOfflineRecipients(). (2) Phase 2 (stale recipients) — no change needed (CleanStaleStates works on timestamps, already updated in T005). (3) Phase 3 (old messages) — ListMailboxRecipients() now returns canonical pane addresses (decoded from filenames), iterate and clean. (4) Phase 4 (empty mailboxes) — no change needed. (5) Add orphaned legacy file detection (FR-020): validate each mailbox filename against the percent-encoded pane address pattern (`<name>%3A<name>%2E<digits>.jsonl`). Files that do not match this pattern are legacy window-name files and should be removed. A simple "%3A" substring check is insufficient — use a pattern match to avoid false positives from malformed files. Update tests in internal/cli/cleanup_test.go with cases: valid pane file kept, legacy "editor.jsonl" removed, malformed file with "%3A" but no "%2E" removed.

- [ ] T016 Update internal/daemon/loop.go: (1) Replace ListWindows() with ListPanes() (current session scope) in CheckAndNotifyWithNotifier(). (2) Update NotifyAgent() to accept pane addresses and pass them to SendKeys(). (3) Phase 1 (stated agents) — recipients are now keyed by pane address, compare against ListPanes() output. (4) Phase 2 (stateless agents) — ListMailboxRecipients() returns canonical pane addresses (decoded), check against recipient state by pane address. (5) StatelessTracker keys change from window names to pane addresses. (6) Cleanup tracker entries for closed panes (compare against ListPanes()). Update tests in internal/daemon/loop_test.go.

**Checkpoint**: Daemon notifies at pane level. Cleanup removes pane-level data.

---

## Phase 9: Polish & Validation

**Purpose**: Final validation, documentation, and quality gates.

- [ ] T017 Run full test suite: `go test -v -race -coverprofile=coverage.out ./...` and verify >= 80% coverage. Fix any failing tests.
- [ ] T018 Run quality gates: `gofmt -l .`, `go vet ./...`, `go mod verify`. Fix any issues.
- [ ] T019 Update README.md: Document pane addressing format (`session:window.pane`), backward compatibility behavior, updated command examples with pane addresses, and the address resolution rules (full/medium/short).
- [ ] T020 Update spec status: Mark spec.md status as "Implemented".

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Foundational)**: No dependencies — start immediately
- **Phase 2-4 (US1, US2, US3 — P1 stories)**: Depend on Phase 1. Execute sequentially (US1 → US2 → US3) since US2 depends on US1 for send/receive integration, and US3 validates US1 backward compat.
- **Phase 5-7 (US4, US5, US6 — P2 stories)**: Depend on Phase 1. Can run in parallel with each other after Phase 1.
- **Phase 8 (Daemon & Cleanup)**: Depends on Phase 1. Can run in parallel with P2 stories.
- **Phase 9 (Polish)**: Depends on all previous phases.

### User Story Dependencies

- **US1 (Send)**: Depends on Phase 1 only
- **US2 (Receive)**: Depends on Phase 1 only (but logically benefits from US1 being done)
- **US3 (Backward Compat)**: Depends on US1 (tests validate US1 behavior)
- **US4 (MCP)**: Depends on Phase 1 only
- **US5 (Hooks)**: Depends on US2 (hooks use receive path)
- **US6 (Recipients)**: Depends on Phase 1 only

### Parallel Opportunities

Within Phase 1:
- T001 (address parsing) must complete first
- T002 and T003 can run in parallel after T001
- T004, T005, T006 can run in parallel after T001

After Phase 1 completes:
- US4 (T010-T011), US6 (T013), and Phase 8 (T014-T016) can all run in parallel
- US5 (T012) can run after US2 (T008) completes

---

## Parallel Example: Phase 1

```text
Sequential:
  T001 (address.go) → then parallel:
    T002 (tmux.go new functions)
    T003 (sendkeys.go update)
  → then parallel:
    T004 (mailbox.go paths)
    T005 (recipients.go keys)
    T006 (ignore.go matching)
```

## Parallel Example: After Phase 1

```text
After Phase 1 completes, these can run in parallel:
  Stream A: T007 (US1 send) → T008 (US2 receive) → T009 (US3 compat) → T012 (US5 hooks)
  Stream B: T010 + T011 (US4 MCP)
  Stream C: T013 (US6 recipients)
  Stream D: T014 + T015 + T016 (daemon/cleanup/status)
```

---

## Implementation Strategy

### MVP First (User Stories 1-3)

1. Complete Phase 1: Foundational address parsing + tmux + mail storage
2. Complete Phase 2: US1 — Send with pane addresses
3. Complete Phase 3: US2 — Receive with pane addresses
4. Complete Phase 4: US3 — Backward compatibility verification
5. **STOP and VALIDATE**: Core send/receive with pane isolation works
6. At this point, agents in separate panes can communicate

### Incremental Delivery

1. Phase 1 → Foundation ready
2. US1 + US2 + US3 → Core pane messaging works (MVP!)
3. US4 → MCP integration for AI agents
4. US5 → Hook notifications pane-isolated
5. US6 → Discoverability via recipients listing
6. Phase 8 → Daemon and cleanup pane-aware
7. Phase 9 → Polish and validate

---

## Notes

- [P] tasks = different files, no dependencies on incomplete tasks
- [Story] label maps task to specific user story for traceability
- Each user story is independently testable after its phase completes
- Commit after each task or logical group
- The Message struct in internal/mail/message.go does NOT need modification — only the content of From/To fields changes, not the struct definition
- SendOptions, ReceiveOptions, etc. need new mock fields (MockPaneAddress, MockPanes, MockSession) to replace MockSender/MockWindows/MockReceiver
- This is a BREAKING CHANGE: no data migration, no MCP backward-compat fields, no upgrade procedure. Old mailbox files are ignored. MCP list-recipients replaces `window` with `address`.
- All public APIs use canonical pane addresses (e.g., "mysession:editor.0"). Percent-encoding is internal to the mail storage layer only — never exposed to callers.

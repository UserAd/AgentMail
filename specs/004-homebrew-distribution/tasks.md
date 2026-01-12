# Tasks: Homebrew Distribution

**Input**: Design documents from `/specs/004-homebrew-distribution/`
**Prerequisites**: plan.md, spec.md, research.md, quickstart.md

**Tests**: No automated tests requested. Formula validated via `brew audit`.

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Includes exact file paths and repository locations

## Path Conventions

This feature spans two repositories:
- **AgentMail repo**: `.github/workflows/release.yml`, `README.md`
- **homebrew-agentmail repo** (new): `Formula/agentmail.rb`, `README.md`

---

## Phase 1: Setup (External Prerequisites)

**Purpose**: Create external infrastructure required before any implementation

- [X] T001 Create GitHub repository `UserAd/homebrew-agentmail` (public, with README)
- [X] T002 Create Fine-Grained PAT named `homebrew-tap-updater` with Contents (Read/Write) permission for `UserAd/homebrew-agentmail` only
- [X] T003 Add repository secret `HOMEBREW_TAP_TOKEN` in `UserAd/AgentMail` → Settings → Secrets

**Checkpoint**: External infrastructure ready - implementation can begin

---

## Phase 2: Foundational (Tap Repository Structure)

**Purpose**: Create the homebrew-agentmail repository structure that ALL user stories depend on

**⚠️ CRITICAL**: User stories cannot be tested until this phase is complete

- [X] T004 Create `Formula/` directory in `UserAd/homebrew-agentmail` repository
- [X] T005 [P] Create initial `Formula/agentmail.rb` with formula structure per research.md template in `UserAd/homebrew-agentmail`
- [X] T006 [P] Create tap `README.md` with installation instructions in `UserAd/homebrew-agentmail`

**Checkpoint**: Tap repository ready - formula can be installed manually

---

## Phase 3: User Story 1 - Install AgentMail via Homebrew (Priority: P1) 🎯 MVP

**Goal**: Enable users to install AgentMail with `brew install UserAd/agentmail/agentmail`

**Independent Test**: Run `brew install UserAd/agentmail/agentmail` on macOS and verify `agentmail --help` works

**Requirements Covered**: FR-001, FR-002, FR-003, FR-006, FR-007, FR-008, FR-009

### Implementation for User Story 1

- [X] T007 [US1] Add macOS arm64 support block with URL and SHA256 placeholder in `Formula/agentmail.rb`
- [X] T008 [US1] Add macOS amd64 support block with URL and SHA256 placeholder in `Formula/agentmail.rb`
- [X] T009 [US1] Add Linux amd64 support block with URL and SHA256 placeholder in `Formula/agentmail.rb`
- [X] T010 [US1] Implement `install` method with architecture detection in `Formula/agentmail.rb`
- [X] T011 [US1] Implement `test` block to verify installation in `Formula/agentmail.rb`
- [X] T012 [US1] Calculate SHA256 checksums for current release binaries and update formula
- [X] T013 [US1] Run `brew audit --new --formula ./Formula/agentmail.rb` to validate formula syntax
- [X] T014 [US1] Test installation locally with `brew install --formula ./Formula/agentmail.rb`
- [X] T015 [US1] Push completed formula to `UserAd/homebrew-agentmail` repository
- [X] T016 [US1] Verify `brew install UserAd/agentmail/agentmail` works from remote tap

**Checkpoint**: User Story 1 complete - users can install AgentMail via Homebrew

---

## Phase 4: User Story 2 - Upgrade AgentMail via Homebrew (Priority: P2)

**Goal**: Enable automated formula updates when new releases are published

**Independent Test**: Trigger a release, verify formula is updated, run `brew upgrade agentmail`

**Requirements Covered**: FR-004

### Implementation for User Story 2

- [X] T017 [US2] Add `update-homebrew` job to `.github/workflows/release.yml` with `needs: [version, release]`
- [X] T018 [US2] Add checkout step for `UserAd/homebrew-agentmail` using `HOMEBREW_TAP_TOKEN` secret in `.github/workflows/release.yml`
- [X] T019 [US2] Add step to download release binaries and calculate SHA256 checksums in `.github/workflows/release.yml`
- [X] T020 [US2] Add step to generate updated formula with version and checksums in `.github/workflows/release.yml`
- [X] T021 [US2] Add step to commit and push formula update in `.github/workflows/release.yml`
- [ ] T022 [US2] Test workflow by triggering a test release (or dry-run validation)

**Checkpoint**: User Story 2 complete - formula auto-updates on release

---

## Phase 5: User Story 3 - Find Installation Instructions in README (Priority: P3)

**Goal**: Document Homebrew installation in main repository README

**Independent Test**: Read README.md, follow instructions, verify installation succeeds

**Requirements Covered**: FR-005, FR-010

### Implementation for User Story 3

- [ ] T023 [US3] Add "Homebrew (macOS/Linux)" subsection under Installation in `README.md`
- [ ] T024 [US3] Add single-command install example `brew install UserAd/agentmail/agentmail` in `README.md`
- [ ] T025 [US3] Add two-step tap + install example in `README.md`
- [ ] T026 [US3] Add note about using full tap path if name conflict exists in `README.md`

**Checkpoint**: User Story 3 complete - README documents Homebrew installation

---

## Phase 6: Polish & Validation

**Purpose**: Final validation and cross-cutting concerns

- [ ] T027 Verify SC-001: Installation completes in under 30 seconds on macOS
- [ ] T028 [P] Verify SC-002: Formula installs on both Intel and Apple Silicon Macs
- [ ] T029 [P] Verify SC-003: SHA256 checksum verification passes
- [ ] T030 Verify SC-004: README instructions enable successful installation
- [ ] T031 Verify SC-005: Formula upgrade works after new release
- [ ] T032 [P] Verify SC-006: Formula installs on Linux via Linuxbrew (if available)
- [ ] T033 Run quickstart.md verification steps from specs/004-homebrew-distribution/quickstart.md

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - external GitHub/PAT setup
- **Foundational (Phase 2)**: Depends on Phase 1 - tap repo must exist
- **User Story 1 (Phase 3)**: Depends on Phase 2 - formula structure must exist
- **User Story 2 (Phase 4)**: Depends on Phase 3 - manual formula must work first
- **User Story 3 (Phase 5)**: Can run in parallel with Phase 4 (different repo/files)
- **Polish (Phase 6)**: Depends on all user stories

### User Story Dependencies

```
Phase 1 (Setup)
    ↓
Phase 2 (Foundational)
    ↓
Phase 3 (US1: Install) ──→ Phase 4 (US2: Upgrade)
    ↓                            ↓
Phase 5 (US3: README) ←──────────┘
    ↓
Phase 6 (Polish)
```

- **User Story 1 (P1)**: BLOCKS User Story 2 (upgrade requires working install)
- **User Story 2 (P2)**: Independent of User Story 3
- **User Story 3 (P3)**: Can run in parallel with User Story 2 after US1 completes

### Within Each User Story

- Formula structure before content
- macOS support before Linux support
- Local testing before remote testing
- Manual process before automation

### Parallel Opportunities

**Phase 2** (after T004):
```
T005 (Formula file) ←→ T006 (README)
```

**Phase 3** (T007-T009 can be combined in single edit):
```
T007 (arm64) + T008 (amd64) + T009 (linux) → single formula edit
```

**Phase 5** (after Phase 3):
```
T023-T026 can run in parallel with Phase 4 tasks
```

**Phase 6** (verification):
```
T028 ←→ T029 ←→ T032 (different platforms/aspects)
```

---

## Parallel Example: Phase 3 Formula Creation

```bash
# These can be combined into a single formula edit (T007-T011):
- Add on_macos/on_arm block
- Add on_macos/on_intel block
- Add on_linux block
- Add install method
- Add test block

# Then sequential validation:
T012 → T013 → T014 → T015 → T016
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (GitHub repo, PAT, secret)
2. Complete Phase 2: Foundational (tap structure)
3. Complete Phase 3: User Story 1 (manual formula)
4. **STOP and VALIDATE**: `brew install UserAd/agentmail/agentmail` works
5. Deploy/demo: Users can now install via Homebrew!

### Incremental Delivery

1. Setup + Foundational → Tap exists
2. Add User Story 1 → Manual install works (MVP!)
3. Add User Story 2 → Auto-updates on release
4. Add User Story 3 → Documentation complete
5. Each story adds value independently

### Single Developer Strategy

Execute phases sequentially:
1. Phase 1-2: ~15 minutes (GitHub setup)
2. Phase 3: ~30 minutes (formula creation and testing)
3. Phase 4: ~20 minutes (CI workflow update)
4. Phase 5: ~10 minutes (README update)
5. Phase 6: ~15 minutes (validation)

Total estimated time: ~90 minutes

---

## Notes

- [P] tasks = different files/repos, no dependencies
- [Story] label maps task to specific user story
- External repo tasks (homebrew-agentmail) require GitHub web UI or separate clone
- PAT creation is manual GitHub UI task (T002)
- Secret creation is manual GitHub UI task (T003)
- SHA256 checksums must be calculated from actual release binaries
- Formula audit (`brew audit`) validates syntax before publishing
- Commit after each logical task group

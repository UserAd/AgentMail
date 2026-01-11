# Tasks: GitHub CI/CD Pipeline

**Input**: Design documents from `/specs/002-github-ci-cd/`
**Prerequisites**: plan.md (required), spec.md (required), research.md

**Tests**: Not explicitly requested in spec - no test tasks included.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

This feature creates GitHub Actions workflow files:
- `.github/workflows/test.yml` - PR testing (User Story 1)
- `.github/workflows/release.yml` - Release creation (User Stories 2 & 3)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the GitHub workflows directory structure

- [x] T001 Create `.github/workflows/` directory structure

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: No foundational tasks needed - workflow files are independent configurations

**⚠️ NOTE**: This feature has no blocking prerequisites. User stories can begin immediately after setup.

**Checkpoint**: Setup ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Automatic Test Execution on PRs (Priority: P1) 🎯 MVP

**Goal**: Run Go tests, vet, and fmt checks automatically on every pull request

**Independent Test**: Open a PR with passing tests → green checkmark. Open a PR with failing tests → red X with details.

### Implementation for User Story 1

- [x] T002 [US1] Create test workflow file at `.github/workflows/test.yml`
- [x] T003 [US1] Configure PR triggers (opened, synchronize, reopened) in `.github/workflows/test.yml`
- [x] T004 [US1] Add Go setup with `actions/setup-go@v5` in `.github/workflows/test.yml`
- [x] T005 [US1] Add `go fmt` check step in `.github/workflows/test.yml`
- [x] T006 [US1] Add `go vet` check step in `.github/workflows/test.yml`
- [x] T007 [US1] Add `go test -v -race -coverprofile=coverage.out ./...` step in `.github/workflows/test.yml`
- [x] T008 [US1] Set minimal permissions (`contents: read`) in `.github/workflows/test.yml`

**Checkpoint**: At this point, User Story 1 should be fully functional. Test by opening a PR.

---

## Phase 4: User Story 2 - Automatic Release Creation (Priority: P2)

**Goal**: Automatically create GitHub releases with semantic versioning when code is merged to main

**Independent Test**: Merge a PR to main → new GitHub Release created with correct version bump based on commit messages.

### Implementation for User Story 2

- [ ] T009 [US2] Create release workflow file at `.github/workflows/release.yml`
- [ ] T010 [US2] Configure main branch push trigger in `.github/workflows/release.yml`
- [ ] T011 [US2] Add checkout with `fetch-depth: 0` for version history in `.github/workflows/release.yml`
- [ ] T012 [US2] Add semantic version calculation with `PaulHatch/semantic-version@v5` in `.github/workflows/release.yml`
- [ ] T013 [US2] Configure conventional commit patterns (feat:, fix:, BREAKING CHANGE:) in `.github/workflows/release.yml`
- [ ] T014 [US2] Add Git tag creation step in `.github/workflows/release.yml`
- [ ] T015 [US2] Add release creation with `softprops/action-gh-release@v2` in `.github/workflows/release.yml`
- [ ] T016 [US2] Configure auto-generated release notes in `.github/workflows/release.yml`
- [ ] T017 [US2] Set write permissions (`contents: write`) in `.github/workflows/release.yml`

**Checkpoint**: At this point, User Story 2 should be fully functional. Test by merging a PR to main.

---

## Phase 5: User Story 3 - Multi-Platform Binary Distribution (Priority: P3)

**Goal**: Build and attach binaries for Linux (amd64) and macOS (amd64, arm64) to each release

**Independent Test**: After release is created, download binaries and verify they execute on target platforms.

### Implementation for User Story 3

- [ ] T018 [US3] Add build job with matrix strategy to `.github/workflows/release.yml`
- [ ] T019 [US3] Configure build matrix for linux/amd64, darwin/amd64, darwin/arm64 in `.github/workflows/release.yml`
- [ ] T020 [US3] Add Go setup in build job in `.github/workflows/release.yml`
- [ ] T021 [US3] Add cross-compilation step with GOOS/GOARCH and CGO_ENABLED=0 in `.github/workflows/release.yml`
- [ ] T021b [US3] Add `continue-on-error: true` to build matrix for graceful partial failure handling (FR-008b) in `.github/workflows/release.yml`
- [ ] T022 [US3] Configure binary naming (`agentmail-{os}-{arch}`) in `.github/workflows/release.yml`
- [ ] T023 [US3] Add ldflags for smaller binaries (`-s -w`) in `.github/workflows/release.yml`
- [ ] T024 [US3] Upload build artifacts with `actions/upload-artifact@v4` in `.github/workflows/release.yml`
- [ ] T025 [US3] Download artifacts in release job with `actions/download-artifact@v4` in `.github/workflows/release.yml`
- [ ] T026 [US3] Attach all binaries to GitHub Release in `.github/workflows/release.yml`

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and edge case handling

- [ ] T027 Verify fork PR handling in `.github/workflows/test.yml`: ensure `pull_request` trigger (not `pull_request_target`) for fork security, and secrets are not exposed
- [ ] T028 Add workflow status badges to README.md
- [ ] T029 Run quickstart.md validation - verify all instructions work and test pipeline completes within 5 minutes (SC-001)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: N/A - no foundational tasks
- **User Stories (Phase 3+)**: Can start immediately after setup
  - US2 and US3 share `release.yml` but tasks are sequential within the file
- **Polish (Final Phase)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Independent - creates `test.yml`
- **User Story 2 (P2)**: Independent - creates `release.yml`
- **User Story 3 (P3)**: Depends on User Story 2 - adds build matrix to existing `release.yml`

### Within Each User Story

- Tasks are sequential (same file modifications)
- Each task builds on previous task within the story

### Parallel Opportunities

- **Between Stories**: US1 and US2 can be developed in parallel (different files)
- **Within Stories**: Tasks within a story are sequential (same file)

---

## Parallel Example: User Stories 1 and 2

```bash
# These can run in parallel (different files):
Task: "Create test workflow file at .github/workflows/test.yml" (US1)
Task: "Create release workflow file at .github/workflows/release.yml" (US2)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001)
2. Complete Phase 3: User Story 1 (T002-T008)
3. **STOP and VALIDATE**: Open a test PR, verify checks run
4. Deploy to main branch

### Incremental Delivery

1. Setup → Create workflow directory
2. Add User Story 1 → Test workflow on PRs → Merge (MVP!)
3. Add User Story 2 → Release creation on main → Merge
4. Add User Story 3 → Binary distribution → Merge
5. Polish → Badges, validation → Final merge

### Single Developer Strategy

Since US2 and US3 share `release.yml`:
1. Complete US1 first (test.yml) - can merge independently
2. Complete US2 + US3 together (release.yml) - single coherent workflow
3. Polish and validate

---

## Notes

- All tasks modify YAML workflow files - no Go code changes
- US3 must come after US2 (adds to same file)
- Test US1 by opening a PR before merging
- Test US2+US3 by merging to main after PR tests pass
- Constitution compliance: Feature enables quality gates (test coverage, vet, fmt)

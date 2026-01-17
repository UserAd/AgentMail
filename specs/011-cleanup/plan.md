# Implementation Plan: Cleanup Command

**Branch**: `011-cleanup` | **Date**: 2026-01-15 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/011-cleanup/spec.md`

## Summary

Add an `agentmail cleanup` command that removes offline recipients, stale recipients (inactive >48h), old delivered messages (read >2h), and empty mailbox files. The command supports configurable thresholds (`--stale-hours`, `--delivered-hours`) and a `--dry-run` mode. This feature requires extending the Message struct to include a `created_at` timestamp field.

## Technical Context

**Language/Version**: Go 1.21+ (per constitution IC-001, project uses Go 1.25.3)
**Primary Dependencies**: Standard library only (os/exec, encoding/json, syscall, time, os)
**Storage**: JSONL files in `.agentmail/` directory (recipients.jsonl, mailboxes/*.jsonl)
**Testing**: `go test -v -race ./...` with 80% minimum coverage
**Target Platform**: macOS and Linux with tmux installed
**Project Type**: Single CLI application
**Performance Goals**: N/A (occasional manual execution, not performance-critical)
**Constraints**: Non-blocking file locking with 1-second timeout; graceful handling of missing files
**Scale/Scope**: Single-user CLI tool; handles typical agent session sizes (10s of recipients, 100s of messages)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | ✅ PASS | New subcommand with text I/O, deterministic exit codes |
| II. Simplicity (YAGNI) | ✅ PASS | Clear use case: prevent stale data accumulation |
| III. Test Coverage (NON-NEGOTIABLE) | ⚠️ PENDING | Must achieve 80% coverage |
| IV. Standard Library Preference | ✅ PASS | Uses only stdlib: os, time, syscall, encoding/json |

**Gate Status**: PASS - No violations. Proceed to Phase 0.

## Project Structure

### Documentation (this feature)

```text
specs/011-cleanup/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # N/A (CLI, no API contracts)
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
cmd/agentmail/
└── main.go              # Add cleanup subcommand registration

internal/
├── cli/
│   ├── cleanup.go       # NEW: Cleanup command implementation
│   └── cleanup_test.go  # NEW: Cleanup command tests
├── mail/
│   ├── message.go       # MODIFY: Add CreatedAt field to Message struct
│   ├── message_test.go  # MODIFY: Update tests for new field
│   ├── mailbox.go       # MODIFY: Add cleanup functions for messages
│   ├── mailbox_test.go  # MODIFY: Tests for cleanup functions
│   ├── recipients.go    # EXISTS: Already has CleanStaleStates, add offline cleanup
│   └── recipients_test.go # MODIFY: Tests for offline cleanup
└── tmux/
    └── tmux.go          # EXISTS: ListWindows for offline check
```

**Structure Decision**: Follows existing single CLI application structure. New `cleanup.go` in `internal/cli/` mirrors existing command files (send.go, receive.go, status.go). Mail package extended with cleanup-specific functions.

## Complexity Tracking

> No complexity violations. Feature fits within existing patterns.

| Aspect | Approach | Justification |
|--------|----------|---------------|
| Message timestamp | Add `created_at` field | Minimal change, backward compatible (omitempty) |
| File locking | Reuse existing flock pattern | Consistent with existing code |
| Non-blocking locks | 1-second timeout | Prevents cleanup blocking on active operations |

## Constitution Check (Post-Design)

*Re-evaluated after Phase 1 design completion.*

| Principle | Status | Verification |
|-----------|--------|--------------|
| I. CLI-First Design | ✅ PASS | `agentmail cleanup` with flags, text output, exit codes 0/1 |
| II. Simplicity (YAGNI) | ✅ PASS | No abstractions; direct file operations following existing patterns |
| III. Test Coverage | ⚠️ PENDING | Implementation must achieve 80% coverage |
| IV. Standard Library | ✅ PASS | No new dependencies; uses time, syscall, os, encoding/json |

**Post-Design Gate Status**: PASS - Ready for task generation.

## Generated Artifacts

| Artifact | Path | Status |
|----------|------|--------|
| Implementation Plan | `specs/011-cleanup/plan.md` | ✅ Complete |
| Research | `specs/011-cleanup/research.md` | ✅ Complete |
| Data Model | `specs/011-cleanup/data-model.md` | ✅ Complete |
| Quickstart | `specs/011-cleanup/quickstart.md` | ✅ Complete |
| Contracts | N/A | CLI command, no API contracts |
| Tasks | `specs/011-cleanup/tasks.md` | 🔜 Next: `/speckit.tasks` |

# Implementation Plan: File-Watching for Mailman with Timer Fallback

**Branch**: `009-watch-files` | **Date**: 2026-01-13 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/009-watch-files/spec.md`

## Summary

Replace the current 10-second polling loop in the mailman daemon with OS file-system watching for instant notification delivery. Watch `.agentmail/mailboxes/` for new/updated mail and `.agentmail/recipients.jsonl` for status changes. Automatically fall back to 10-second polling if file-watching is unavailable. Additionally, track when agents last called `agentmail receive` by updating a `last_read_at` timestamp in `recipients.jsonl`.

## Technical Context

**Language/Version**: Go 1.21+ (per constitution IC-001)
**Primary Dependencies**: Standard library only (os, time, syscall) + fsnotify (external - requires justification)
**Storage**: JSONL files in `.agentmail/` directory
**Testing**: `go test -v -race ./...` with minimum 80% coverage
**Target Platform**: macOS, Linux, Windows (per SC-007)
**Project Type**: Single CLI tool
**Performance Goals**: Notification delivery within 2 seconds of file change (per SC-001, SC-002)
**Constraints**: 500ms debounce window, 60-second fallback timer
**Scale/Scope**: Single daemon process monitoring local filesystem

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | ✅ Pass | Enhances existing CLI daemon, no new GUI |
| II. Simplicity (YAGNI) | ⚠️ Review | External dependency (fsnotify) requires justification |
| III. Test Coverage | ✅ Pass | Will maintain 80%+ coverage |
| IV. Standard Library Preference | ⚠️ Review | fsnotify is external; standard library lacks cross-platform file watching |

**Gate Decision**: Proceed to Phase 0 to research fsnotify justification.

## Project Structure

### Documentation (this feature)

```text
specs/009-watch-files/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (N/A - no API contracts for CLI)
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
cmd/agentmail/           # Main entry point (no changes expected)
internal/
├── daemon/
│   ├── daemon.go        # StartDaemon, runForeground (modify for watcher integration)
│   ├── loop.go          # RunLoop, CheckAndNotify (modify for event-driven mode)
│   ├── watcher.go       # NEW: File watcher abstraction
│   └── watcher_test.go  # NEW: Watcher tests
├── mail/
│   ├── recipients.go    # RecipientState (add LastReadAt field)
│   └── recipients_test.go # Update tests for new field
└── cli/
    ├── receive.go       # Receive command (add last-read tracking)
    └── receive_test.go  # Update tests for last-read tracking
```

**Structure Decision**: Extend existing `internal/daemon/` package with new `watcher.go` file. Modify `recipients.go` to add `LastReadAt` field. Modify `receive.go` to update last-read timestamp.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| External dependency (fsnotify) | Cross-platform file watching required for macOS, Linux, Windows support per SC-007 | Go standard library has no cross-platform file watching API; using syscall directly would require OS-specific code for inotify (Linux), FSEvents (macOS), and ReadDirectoryChangesW (Windows) |

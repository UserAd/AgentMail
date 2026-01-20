# Implementation Plan: Mailman Daemon Stop Command

**Branch**: `012-mailman-stop` | **Date**: 2026-01-20 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/012-mailman-stop/spec.md`

## Summary

Add `agentmail mailman stop` subcommand to gracefully terminate the mailman daemon using file-based signaling. The command creates a `.stop` file in `.agentmail/` directory. The daemon detects this file via its existing fsnotify watcher and initiates graceful shutdown, removing both the stop file and PID file. This approach is simpler and more cross-platform than signal-based termination.

## Technical Context

**Language/Version**: Go 1.25.5 (minimum 1.21+ per IC-001)
**Primary Dependencies**: Standard library only (os for file operations)
**Storage**: `.agentmail/.stop` (new stop signal file), `.agentmail/mailman.pid` (existing)
**Testing**: `go test -v -race -cover` with >= 80% coverage
**Target Platform**: macOS and Linux (per constitution, with tmux installed)
**Project Type**: Single CLI tool
**Performance Goals**: Stop command returns within 1 second (fire-and-forget)
**Constraints**: No external dependencies for stop functionality
**Scale/Scope**: Single daemon process per repository

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | ✅ PASS | New subcommand with text I/O, exit codes 0/1 |
| II. Simplicity (YAGNI) | ✅ PASS | File-based IPC, reuses existing fsnotify watcher |
| III. Test Coverage | ✅ REQUIRED | Must achieve >= 80% coverage on new code |
| IV. Standard Library | ✅ PASS | Uses only stdlib (os package for file operations) |

**Quality Gates Required**:
1. `gofmt -l .` - no output
2. `go mod verify` - pass
3. `go vet ./...` - pass
4. `go test -v -race -coverprofile=coverage.out ./...` - pass with >= 80%
5. `govulncheck ./...` - no vulnerabilities
6. `gosec ./...` - no issues

## Project Structure

### Documentation (this feature)

```text
specs/012-mailman-stop/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0: File-based IPC research
├── data-model.md        # Phase 1: Stop signal file entity
├── quickstart.md        # Phase 1: Quick implementation guide
├── contracts/           # Phase 1: CLI contract
│   └── cli.md
└── tasks.md             # Phase 2: Implementation tasks
```

### Source Code (repository root)

```text
cmd/
└── agentmail/
    └── main.go          # Add 'stop' subcommand to mailman command

internal/
├── cli/
│   ├── mailman.go       # Existing mailman command handler
│   ├── mailman_stop.go  # NEW: Stop command implementation
│   └── mailman_stop_test.go  # NEW: Tests for stop command
└── daemon/
    ├── daemon.go        # MODIFY: Add StopFile constant, StopFilePath, stop file removal
    ├── watcher.go       # MODIFY: Add stop file detection, StopChan() method
    └── watcher_test.go  # MODIFY: Add stop file detection tests
```

**Structure Decision**: Follows existing project layout. New `mailman_stop.go` in `internal/cli/` for the stop command handler. Modifications to existing `daemon.go` and `watcher.go` for stop file detection.

## Complexity Tracking

No violations - this feature aligns with all constitution principles. The file-based approach is actually simpler than the originally planned SIGTERM approach:

| Aspect | SIGTERM Approach | File-Based Approach |
|--------|------------------|---------------------|
| New files | 4 (process.go, process_test.go, stop.go, stop_test.go) | 2 (stop.go, stop_test.go) |
| Dependencies | os, os/exec, syscall, strings | os only |
| Platform support | Unix only | macOS/Linux |
| Process validation | Required | Not needed |
| Error scenarios | Many (permissions, wrong process) | Few (file exists, filesystem error) |

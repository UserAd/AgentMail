# Implementation Plan: Mailman Daemon

**Branch**: `006-mailman-daemon` | **Date**: 2026-01-12 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/006-mailman-daemon/spec.md`

## Summary

Add `agentmail mailman` daemon command for automated notification delivery to agents, plus `agentmail status` command for hooks-based agent state management. The daemon monitors mailboxes and sends tmux notifications to agents in "ready" state.

## Technical Context

**Language/Version**: Go 1.21+ (per constitution IC-001, project uses Go 1.25.3)
**Primary Dependencies**: Standard library only (os/exec, encoding/json, syscall, time, os/signal)
**Storage**: JSONL files - `.git/mail/mailman.pid` (PID), `.git/mail-recipients.jsonl` (state)
**Testing**: `go test -v ./...` with 80%+ coverage requirement
**Target Platform**: macOS and Linux with tmux installed
**Project Type**: Single CLI project (existing structure)
**Performance Goals**: 10-second notification cycle, 1-second state change latency
**Constraints**: Single daemon instance per repo, file locking for concurrency
**Scale/Scope**: Single repository, multiple agents (tmux windows)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*
*Constitution v1.1.0 - daemon processes now permitted when enhancing CLI workflows*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | PASS | Daemon enhances CLI workflows (per constitution v1.1.0) |
| II. Simplicity (YAGNI) | PASS | Minimal implementation, std lib only |
| III. Test Coverage | PASS | 80%+ coverage required, TDD approach |
| IV. Standard Library Preference | PASS | No external dependencies |

**Quality Gates**:
- [x] `go test -cover ./...` >= 80%
- [x] `go vet ./...` passes
- [x] `go fmt ./...` no changes
- [x] All acceptance scenarios pass

## Project Structure

### Documentation (this feature)

```text
specs/006-mailman-daemon/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── cli.md           # CLI command contracts
└── tasks.md             # Phase 2 output (via /speckit.tasks)
```

### Source Code (repository root)

```text
cmd/
└── agentmail/
    └── main.go          # CLI entry point (existing)

internal/
├── cli/
│   ├── mailman.go       # NEW: mailman command handler
│   ├── mailman_test.go  # NEW: mailman tests
│   ├── status.go        # NEW: status command handler
│   └── status_test.go   # NEW: status tests
├── daemon/
│   ├── daemon.go        # NEW: daemon lifecycle (PID, signals)
│   ├── daemon_test.go   # NEW: daemon tests
│   ├── loop.go          # NEW: notification loop
│   └── loop_test.go     # NEW: loop tests
├── mail/
│   ├── mailbox.go       # Existing (reuse FindUnread)
│   ├── recipients.go    # NEW: recipients state JSONL
│   └── recipients_test.go # NEW: recipients tests
└── tmux/
    ├── tmux.go          # Existing (reuse InTmux, GetCurrentWindow)
    ├── sendkeys.go      # NEW: tmux send-keys wrapper
    └── sendkeys_test.go # NEW: sendkeys tests
```

**Structure Decision**: Extends existing single-project structure with new `internal/daemon/` package for daemon-specific logic and new files in existing packages.

## Complexity Tracking

> **No constitution violations** - All principles pass as of constitution v1.1.0.

The daemon is compliant with the updated CLI-First Design principle which explicitly permits daemon processes when they enhance CLI workflows. The mailman daemon:
- Is started via CLI (`agentmail mailman`)
- Is controlled via CLI (`agentmail status`)
- Uses standard exit codes
- Enhances rather than replaces CLI-based message access

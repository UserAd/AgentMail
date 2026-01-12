# Implementation Plan: Recipients Command, Help Flag, and Stdin Message Input

**Branch**: `003-recipients-help-stdin` | **Date**: 2026-01-12 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-recipients-help-stdin/spec.md`

## Summary

This feature adds three capabilities to agentmail:
1. **Recipients command** - List all active tmux windows with current window marked "[you]", other windows filtered by `.agentmailignore`
2. **Help flag** - Display usage documentation via `--help`
3. **Stdin support** - Accept message content from stdin for programmatic use

Technical approach: Extend existing CLI structure with new `recipients` command, add `--help` handling in main.go, and modify send command to check stdin before using argument.

## Technical Context

**Language/Version**: Go 1.21+ (per constitution IC-001, project uses Go 1.25.3)
**Primary Dependencies**: Standard library only (os/exec, encoding/json, bufio, os)
**Storage**: JSONL files in `.git/mail/` directory (existing)
**Testing**: `go test` with 80% minimum coverage (per constitution)
**Target Platform**: macOS and Linux with tmux installed
**Project Type**: Single CLI application
**Performance Goals**: Command execution in under 1 second (per SC-001)
**Constraints**: No external dependencies (per constitution Principle IV)
**Scale/Scope**: Single-user CLI tool, operates within one tmux session

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | PASS | All features are CLI commands with text I/O |
| II. Simplicity (YAGNI) | PASS | Minimal implementation, extends existing patterns |
| III. Test Coverage (80%) | PENDING | Tests must be written for all new functions |
| IV. Standard Library Preference | PASS | Uses only os, bufio, strings (standard library) |

**Quality Gates**:
- [ ] `go test -cover ./...` >= 80%
- [ ] `go vet ./...` passes
- [ ] `go fmt ./...` produces no changes
- [ ] All acceptance scenarios from spec.md pass

## Project Structure

### Documentation (this feature)

```text
specs/003-recipients-help-stdin/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (CLI interface contracts)
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
cmd/agentmail/
└── main.go              # Add --help handling, recipients command routing

internal/
├── cli/
│   ├── send.go          # Modify to support stdin input
│   ├── send_test.go     # Add stdin tests
│   ├── receive.go       # Existing (unchanged)
│   ├── receive_test.go  # Existing (unchanged)
│   ├── recipients.go    # NEW: recipients command implementation
│   ├── recipients_test.go # NEW: recipients command tests
│   ├── help.go          # NEW: help text generation
│   ├── help_test.go     # NEW: help tests
│   └── integration_test.go # Add integration tests for new features
├── mail/
│   ├── mailbox.go       # Existing (unchanged)
│   ├── mailbox_test.go  # Existing (unchanged)
│   ├── message.go       # Existing (unchanged)
│   ├── message_test.go  # Existing (unchanged)
│   ├── ignore.go        # NEW: .agentmailignore parsing
│   └── ignore_test.go   # NEW: ignore file tests
└── tmux/
    ├── tmux.go          # Existing - has ListWindows(), GetCurrentWindow()
    └── tmux_test.go     # Existing (unchanged)
```

**Structure Decision**: Extends existing single-project structure. New files follow established patterns in internal/cli/ and internal/mail/.

## Complexity Tracking

No constitution violations. All features use standard library and extend existing patterns.

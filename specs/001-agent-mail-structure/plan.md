# Implementation Plan: AgentMail Initial Project Structure

**Branch**: `001-agent-mail-structure` | **Date**: 2026-01-11 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-agent-mail-structure/spec.md`

## Summary

AgentMail is a CLI tool for inter-agent asynchronous communication within tmux sessions. The MVP provides two commands (`send` and `receive`) that allow agents (identified by tmux window names) to exchange messages stored in JSONL format in `.git/mail/`. The system validates tmux context and recipient existence before operations.

## Technical Context

**Language/Version**: Go 1.21+ (per IC-001)
**Primary Dependencies**: Standard library only (os/exec for tmux, encoding/json for JSONL)
**Storage**: JSONL files in `.git/mail/<recipient>.jsonl` (one file per recipient)
**Testing**: `go test` with 80% coverage target (SC-005)
**Target Platform**: macOS/Linux with tmux installed
**Project Type**: Single CLI application
**Performance Goals**: <1 second command execution (SC-001, SC-002)
**Constraints**: Must run inside tmux session; single-machine only
**Scale/Scope**: Small number of concurrent agents (single-digit to low tens)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Status**: PASS (Constitution v1.0.0)

| Principle | Compliance | Evidence |
|-----------|------------|----------|
| I. CLI-First Design | ✅ | Text I/O, exit codes 0/1/2, no GUI/daemon |
| II. Simplicity (YAGNI) | ✅ | MVP send/receive only, stdlib dependencies |
| III. Test Coverage (80%) | ✅ | SC-005 mandates 80%, T042 verifies |
| IV. Standard Library | ✅ | os/exec, encoding/json, crypto/rand, syscall |

## Project Structure

### Documentation (this feature)

```text
specs/001-agent-mail-structure/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   └── cli.md           # CLI contract specification
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
cmd/
└── agentmail/
    └── main.go          # CLI entry point

internal/
├── mail/
│   ├── message.go       # Message struct and JSONL operations
│   ├── message_test.go  # Unit tests for Message struct
│   ├── mailbox.go       # Mailbox file operations (read/write/query)
│   └── mailbox_test.go  # Unit tests for mailbox
├── tmux/
│   ├── tmux.go          # tmux detection and window operations
│   └── tmux_test.go     # Unit tests for tmux integration
└── cli/
    ├── send.go          # Send command implementation
    ├── send_test.go     # Send command tests
    ├── receive.go       # Receive command implementation
    ├── receive_test.go  # Receive command tests
    └── integration_test.go  # CLI integration tests

go.mod                   # Go module definition
go.sum                   # Dependency checksums
```

**Structure Decision**: Single Go CLI application using standard `cmd/` and `internal/` layout. The `internal/` package prevents external imports and organizes code by domain (mail, tmux, cli). This follows Go project conventions and keeps the codebase simple for an MVP.

## Complexity Tracking

*No violations - design follows simplicity principles.*

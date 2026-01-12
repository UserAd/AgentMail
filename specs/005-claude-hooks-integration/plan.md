# Implementation Plan: Claude Code Hooks Integration

**Branch**: `005-claude-hooks-integration` | **Date**: 2026-01-12 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/005-claude-hooks-integration/spec.md`

## Summary

Add a `--hook` flag to the `agentmail receive` command that modifies behavior for Claude Code hooks integration. When enabled, the command outputs to STDERR (not STDOUT), exits with code 2 when messages exist, and exits with code 0 producing no output for all other conditions (no messages, not in tmux, or errors).

## Requirements Traceability

| Requirement | Description | Implementation Location |
|-------------|-------------|------------------------|
| FR-001a | Write notification to STDERR | `internal/cli/receive.go` |
| FR-001b | Exit code 2 on messages | `internal/cli/receive.go` |
| FR-001c | Mark message as read | `internal/cli/receive.go` (existing) |
| FR-002 | Exit 0 + no output when no messages | `internal/cli/receive.go` |
| FR-003 | Exit 0 + no output when not in tmux | `internal/cli/receive.go` |
| FR-004a/b/c | Exit 0 + no output on errors | `internal/cli/receive.go` |
| FR-005 | All output to STDERR in hook mode | `internal/cli/receive.go` |
| FR-006 | Documentation section | `README.md` |

## Technical Context

**Language/Version**: Go 1.21+ (per constitution IC-001, project uses Go 1.25.3)
**Primary Dependencies**: Standard library only (os, fmt, io - already used)
**Storage**: JSONL files in `.git/mail/` directory (existing infrastructure)
**Testing**: `go test` with table-driven tests (existing pattern)
**Target Platform**: macOS and Linux with tmux
**Project Type**: Single CLI application
**Performance Goals**: Hook execution within 500ms (per SC-004)
**Constraints**: Silent failure for hooks (exit 0 on errors), STDERR output only
**Scale/Scope**: Minimal change - extends existing `receive` command with flag

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|----------|
| **I. CLI-First Design** | PASS | Adds CLI flag (`--hook`), uses exit codes (0, 2), outputs to stderr |
| **II. Simplicity (YAGNI)** | PASS | Minimal change - single boolean flag, reuses existing receive logic |
| **III. Test Coverage (NON-NEGOTIABLE)** | PENDING | Tests required for hook mode behavior (80% minimum) |
| **IV. Standard Library Preference** | PASS | No new dependencies required |

**Technology Constraints Check**:
- Language: Go 1.21+ ✓
- Storage: Uses existing JSONL in `.git/mail/` ✓
- Platform: macOS/Linux with tmux ✓
- Build: Standard `go build` ✓

**Quality Gates** (to verify at completion):
1. `go test -cover ./...` >= 80%
2. `go vet ./...` passes
3. `go fmt ./...` produces no changes
4. All acceptance scenarios pass

## Project Structure

### Documentation (this feature)

```text
specs/005-claude-hooks-integration/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output (minimal for this feature)
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
cmd/
└── agentmail/
    └── main.go          # CLI entry point (modify for --hook flag parsing)

internal/
├── cli/
│   ├── receive.go       # PRIMARY: Add HookMode to ReceiveOptions
│   └── receive_test.go  # Add hook mode test cases
├── mail/
│   ├── mailbox.go       # Existing (no changes needed)
│   └── message.go       # Existing (no changes needed)
└── tmux/
    └── tmux.go          # Existing (no changes needed)

README.md                # Add Claude Code Hooks section
```

**Structure Decision**: Single CLI application. Modifications confined to:
1. `cmd/agentmail/main.go` - Flag parsing for `--hook`
2. `internal/cli/receive.go` - Hook mode logic in `ReceiveOptions` struct
3. `internal/cli/receive_test.go` - Test coverage for hook behaviors
4. `README.md` - Documentation section

## Complexity Tracking

No constitution violations - no complexity justification needed.

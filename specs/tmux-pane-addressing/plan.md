# Implementation Plan: Tmux Pane Addressing

**Branch**: `tmux-pane-addressing` | **Date**: 2026-02-13 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/tmux-pane-addressing/spec.md`

## Summary

Update AgentMail to use full tmux pane addresses (`session:window.pane`) instead of window names for all agent identification. This affects the tmux integration layer, mail storage, all CLI commands, MCP server, Claude hooks, and the mailman daemon. The approach introduces an address parsing module, extends tmux query functions to return pane-level data, and updates all consumers to use pane addresses. No new dependencies are required. Backward compatibility is maintained for window-name-only addressing when the target window has a single pane.

## Technical Context

**Language/Version**: Go 1.25.7 (minimum 1.21+ per constitution IC-001)
**Primary Dependencies**: Standard library only (`os/exec`, `encoding/json`, `strings`, `strconv`, `fmt`, `regexp`). Existing approved deps: `fsnotify`, `go-sdk` (MCP), `ff/v3` (CLI flags)
**Storage**: JSONL files in `.agentmail/` directory — mailbox files renamed from `<window>.jsonl` to `<percent-encoded-pane-address>.jsonl` (collision-safe encoding)
**Testing**: `go test -v -race -coverprofile=coverage.out ./...` with >= 80% coverage
**Target Platform**: macOS and Linux with tmux installed
**Project Type**: Single Go project (CLI tool)
**Performance Goals**: N/A — address parsing adds negligible overhead (string operations only)
**Constraints**: Standard library preference (Constitution IV), no CGO
**Scale/Scope**: ~15 files modified, 1 new file created, ~500-800 lines of new/changed code

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | PASS | All pane addressing accessible via CLI commands. Text I/O protocol unchanged. Exit codes unchanged. |
| II. Simplicity (YAGNI) | PASS | Feature justified by demonstrated need (agents in separate panes). No migration complexity (orphan old files). Address parsing is minimal string logic. |
| III. Test Coverage | PASS | All new functions will have unit tests. Integration tests for CLI flows. 80% coverage target maintained. |
| IV. Standard Library Preference | PASS | No new external dependencies. All address parsing uses `strings`, `strconv`, `fmt`. |
| Quality Gate 1: gofmt | Will verify | All new code formatted with gofmt. |
| Quality Gate 2: go mod verify | Will verify | No module changes expected. |
| Quality Gate 3: go vet | Will verify | All new code passes vet. |
| Quality Gate 4: Tests >= 80% | Will verify | New tests for address parsing, tmux queries, updated CLI/MCP flows. |
| Quality Gate 5: govulncheck | Will verify | No new dependencies to introduce vulnerabilities. |
| Quality Gate 6: gosec | Will verify | Address sanitization prevents path traversal. Input validation on pane indices. |
| Quality Gate 7: Spec compliance | Will verify | All acceptance scenarios from spec.md covered by tests. |

**Post-Phase 1 Re-check**: PASS — No design decisions violate any constitution principles. No complexity tracking entries needed.

## Project Structure

### Documentation (this feature)

```text
specs/tmux-pane-addressing/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0: research findings
├── data-model.md        # Phase 1: entity and storage changes
├── quickstart.md        # Phase 1: implementation guide
├── contracts/
│   └── cli.md           # Phase 1: CLI and MCP contract changes
├── checklists/
│   └── requirements.md  # Spec quality checklist
├── tasks.md             # Phase 2 output (/speckit.tasks)
└── worklog.md           # Implementation log (/speckit.implement)
```

### Source Code (repository root)

```text
internal/
├── tmux/
│   ├── address.go          # NEW: PaneAddress struct, ParseAddress, SanitizeForFilename
│   ├── address_test.go     # NEW: Address parsing tests
│   ├── tmux.go             # MODIFY: Add GetCurrentPaneAddress, GetCurrentSession, ListPanes, ListAllPanes, PaneExists
│   ├── tmux_test.go        # MODIFY: Add tests for new functions
│   └── sendkeys.go         # MODIFY: Accept pane addresses in SendKeys
├── mail/
│   ├── mailbox.go          # MODIFY: Path construction uses sanitized pane addresses
│   ├── mailbox_test.go     # MODIFY: Tests with pane-based paths
│   ├── recipients.go       # MODIFY: Key by pane address
│   ├── recipients_test.go  # MODIFY: Tests with pane-based keys
│   └── ignore.go           # MODIFY: Support pane-level ignore matching
├── cli/
│   ├── send.go             # MODIFY: Address parsing for recipient, pane address for sender
│   ├── send_test.go        # MODIFY: Pane address test cases
│   ├── receive.go          # MODIFY: Pane address for receiver
│   ├── receive_test.go     # MODIFY: Pane address test cases
│   ├── recipients.go       # MODIFY: List panes instead of windows
│   ├── recipients_test.go  # MODIFY: Pane address test cases
│   ├── status.go           # MODIFY: Pane address for agent
│   ├── status_test.go      # MODIFY: Pane address test cases
│   ├── cleanup.go          # MODIFY: Pane-level cleanup
│   └── cleanup_test.go     # MODIFY: Pane address test cases
├── mcp/
│   ├── handlers.go         # MODIFY: Use pane addresses in all handlers
│   ├── handlers_test.go    # MODIFY: Pane address test cases
│   ├── tools.go            # MODIFY: Update descriptions, response types
│   └── server.go           # NO CHANGE
└── daemon/
    ├── loop.go             # MODIFY: Pane-level notification, pane-level tracking
    ├── loop_test.go        # MODIFY: Pane address test cases
    ├── watcher.go          # NO CHANGE (file watching is address-agnostic)
    └── daemon.go           # NO CHANGE (lifecycle is address-agnostic)
```

**Structure Decision**: Existing Go project structure is preserved. One new file (`internal/tmux/address.go` + test) is added for address parsing. All other changes are modifications to existing files. No structural reorganization needed.

## Complexity Tracking

No constitution violations. No complexity justifications needed.

# Quickstart: Tmux Pane Addressing

**Feature Branch**: `tmux-pane-addressing`
**Date**: 2026-02-13

## Overview

This feature updates AgentMail to use full tmux pane addresses (`session:window.pane`) instead of window names only. Each pane becomes a distinct addressable agent.

## File Change Map

Changes are organized by layer, bottom-up:

### Layer 1: Address Parsing (new)

| File | Action | Description |
|------|--------|-------------|
| `internal/tmux/address.go` | **Create** | `PaneAddress` struct, `ParseAddress()`, `FormatAddress()`, `SanitizeForFilename()`, `UnsanitizeFromFilename()` |
| `internal/tmux/address_test.go` | **Create** | Unit tests for all address parsing cases |

### Layer 2: Tmux Integration (modify)

| File | Action | Description |
|------|--------|-------------|
| `internal/tmux/tmux.go` | **Modify** | Add `GetCurrentPaneAddress()`, `GetCurrentSession()`, `ListPanes()`, `ListAllPanes()` (not exposed in commands), `PaneExists()` |
| `internal/tmux/tmux_test.go` | **Modify** | Add tests for new tmux functions |
| `internal/tmux/sendkeys.go` | **Modify** | Update `SendKeys()` to accept pane addresses |

### Layer 3: Mail Storage (modify)

| File | Action | Description |
|------|--------|-------------|
| `internal/mail/mailbox.go` | **Modify** | Update `safePath()` and all path construction to use sanitized pane addresses |
| `internal/mail/recipients.go` | **Modify** | Update all functions to key by pane address instead of window name |
| `internal/mail/ignore.go` | **Modify** | Update `LoadIgnoreList()` to support pane-level matching |
| `internal/mail/mailbox_test.go` | **Modify** | Update tests for pane-based paths |
| `internal/mail/recipients_test.go` | **Modify** | Update tests for pane-based keys |

### Layer 4: CLI Commands (modify)

| File | Action | Description |
|------|--------|-------------|
| `internal/cli/send.go` | **Modify** | Use address parsing for recipient, `GetCurrentPaneAddress()` for sender |
| `internal/cli/receive.go` | **Modify** | Use `GetCurrentPaneAddress()` for receiver identity |
| `internal/cli/recipients.go` | **Modify** | Use `ListPanes()` instead of `ListWindows()` |
| `internal/cli/status.go` | **Modify** | Use `GetCurrentPaneAddress()` for agent identity |
| `internal/cli/cleanup.go` | **Modify** | Update offline detection to use pane addresses |
| `internal/cli/*_test.go` | **Modify** | Update all CLI tests with pane address mocks |

### Layer 5: MCP Server (modify)

| File | Action | Description |
|------|--------|-------------|
| `internal/mcp/handlers.go` | **Modify** | Update all handlers to use pane addresses |
| `internal/mcp/tools.go` | **Modify** | Update tool descriptions and response types |
| `internal/mcp/handlers_test.go` | **Modify** | Update MCP handler tests |

### Layer 6: Daemon (modify)

| File | Action | Description |
|------|--------|-------------|
| `internal/daemon/loop.go` | **Modify** | Update notification loop to use pane addresses, target panes with `send-keys` |
| `internal/daemon/watcher.go` | **No change** | File watching is address-agnostic |
| `internal/daemon/daemon.go` | **No change** | Daemon lifecycle is address-agnostic |

## Implementation Order

1. **Address parsing** (`internal/tmux/address.go`) — no dependencies, foundation for everything
2. **Tmux query functions** (`internal/tmux/tmux.go`) — depends on address parsing
3. **Mail storage** (`internal/mail/`) — depends on address format for filenames
4. **CLI commands** (`internal/cli/`) — depends on tmux and mail layers
5. **MCP server** (`internal/mcp/`) — depends on tmux and mail layers (parallel with CLI)
6. **Daemon** (`internal/daemon/`) — depends on tmux and mail layers (parallel with CLI/MCP)

## Key Design Decisions

1. **Always full address**: `From`/`To` fields always store `session:window.pane`, even for single-pane windows
2. **Breaking change, no migration**: Old mailbox files are ignored; `cleanup` removes them. No dual-read, no upgrade procedure.
3. **Filename sanitization**: Percent-encoding (`%` → `%25`, `:` → `%3A`, `.` → `%2E`) — collision-safe and reversible
4. **Medium form requires colon**: `:window.pane` (not `window.pane`) to avoid ambiguity with dotted window names
5. **Short form backward compat**: Single-pane window auto-resolves; multi-pane returns ambiguity error
6. **Discovery scope**: Current session only for recipients/MCP list; cross-session send requires full address
7. **MCP breaking change**: `list-recipients` uses `address` field (replaces `window`, no deprecated field)
8. **Standard library only**: No new dependencies needed
9. **Type boundary**: All public APIs use canonical pane addresses; percent-encoding is internal to storage layer only

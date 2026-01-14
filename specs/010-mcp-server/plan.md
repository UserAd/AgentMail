# Implementation Plan: MCP Server for AgentMail

**Branch**: `010-mcp-server` | **Date**: 2026-01-14 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/010-mcp-server/spec.md`

## Summary

Add MCP (Model Context Protocol) server capability to AgentMail, enabling AI agents to communicate via STDIO transport without shell command execution. The server exposes four tools (send, receive, status, list-recipients) that wrap existing CLI functionality, providing an alternative interface for assistants where hook mechanisms or shell invocation is difficult.

**Technical Approach**: Add `agentmail mcp` subcommand using the official `modelcontextprotocol/go-sdk` package. The MCP server runs as a subprocess, inheriting the tmux context from the parent process, and communicates via JSON-RPC over STDIO.

## Technical Context

**Language/Version**: Go 1.23 (per go.mod, constitution requires 1.21+)
**Primary Dependencies**:
- Existing: `github.com/peterbourgon/ff/v3` (CLI framework), `github.com/fsnotify/fsnotify` (file watching)
- New: `github.com/modelcontextprotocol/go-sdk` (official MCP SDK)

**Storage**: JSONL files in `.agentmail/` directory (existing infrastructure)
**Testing**: `go test` with race detection, MCP Inspector CLI mode for integration tests
**Target Platform**: macOS and Linux with tmux installed
**Project Type**: Single CLI application
**Performance Goals**: Tool invocations complete within 2 seconds (SC-004), tool discovery within 1 second (SC-001)
**Constraints**: 64KB message size limit (FR-013), single MCP connection per agent
**Scale/Scope**: Single agent per MCP server instance

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. CLI-First Design
- [x] **PASS**: MCP server is CLI subcommand (`agentmail mcp`)
- [x] **PASS**: STDIO transport (text-based I/O)
- [x] **PASS**: Errors logged to stderr (FR-015)
- [x] **PASS**: No GUI or web interface

### II. Simplicity (YAGNI)
- [x] **PASS**: Only 4 tools exposed (per spec requirement)
- [x] **PASS**: Reuses existing CLI logic (Send, Receive, Status, Recipients)
- [x] **PASS**: No premature abstractions - thin wrapper over existing code

### III. Test Coverage (NON-NEGOTIABLE)
- [x] **PLANNED**: Unit tests for MCP tool handlers
- [x] **PLANNED**: Integration tests using MCP Inspector CLI mode
- [x] **PLANNED**: Coverage gate >= 80%

### IV. Standard Library Preference
- [!] **REQUIRES JUSTIFICATION**: New external dependency `modelcontextprotocol/go-sdk`

**Justification** (documented in research.md):
- Standard library insufficient for MCP protocol compliance
- MCP requires complex handshakes, capability negotiation, JSON Schema generation
- Official SDK maintained by protocol authors (Anthropic + Google)
- Widely adopted (8M+ downloads), actively maintained
- Alternative: Custom JSON-RPC implementation would be significant effort and error-prone

### Quality Gates
- [ ] `go test -cover ./...` >= 80%
- [ ] `go vet ./...` passes
- [ ] `go fmt ./...` produces no changes
- [ ] All acceptance scenarios from spec.md pass

## Project Structure

### Documentation (this feature)

```text
specs/010-mcp-server/
├── plan.md              # This file
├── research.md          # Phase 0 output (complete)
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (MCP tool schemas)
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
cmd/
└── agentmail/
    └── main.go              # Add mcp subcommand

internal/
├── cli/
│   ├── send.go              # Existing - reused by MCP
│   ├── receive.go           # Existing - reused by MCP
│   ├── status.go            # Existing - reused by MCP
│   ├── recipients.go        # Existing - reused by MCP
│   └── ...
├── mcp/                     # NEW: MCP server package
│   ├── server.go            # MCP server initialization and main loop
│   ├── server_test.go       # Unit tests for server
│   ├── tools.go             # Tool definitions and handlers
│   ├── tools_test.go        # Unit tests for tools
│   └── handlers.go          # Tool handler implementations
├── mail/                    # Existing - unchanged
├── tmux/                    # Existing - unchanged
└── daemon/                  # Existing - unchanged

tests/
└── acceptance/              # NEW: Acceptance test scripts (removed after use)
    ├── test_tool_discovery.py
    ├── test_send_receive.py
    └── test_status_recipients.py
```

**Structure Decision**: Follows existing Go project layout. New `internal/mcp/` package contains all MCP-specific code, keeping it isolated from existing CLI infrastructure while reusing the same underlying logic.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| External dependency: `modelcontextprotocol/go-sdk` | MCP protocol compliance requires complex handshakes, JSON Schema generation, capability negotiation | Custom JSON-RPC implementation would be 500+ lines, error-prone, and may drift from spec |

## Implementation Overview

### Phase 1: Core MCP Infrastructure
1. Add `modelcontextprotocol/go-sdk` dependency
2. Create `internal/mcp/server.go` with server initialization
3. Create `internal/mcp/tools.go` with tool definitions (schemas)
4. Add `mcp` subcommand to `cmd/agentmail/main.go`

### Phase 2: Tool Handlers
1. Implement `sendHandler` wrapping `cli.Send`
2. Implement `receiveHandler` wrapping `cli.Receive`
3. Implement `statusHandler` wrapping `cli.Status`
4. Implement `listRecipientsHandler` wrapping `cli.Recipients`

### Phase 3: Error Handling & Edge Cases
1. Add 64KB message size validation (FR-013)
2. Add tmux context loss detection with 1-second exit (FR-014)
3. Add stderr logging (FR-015)
4. Add malformed JSON handling with -32700 error code (FR-010)
5. Add invalid status value handling (FR-016)

### Phase 4: Testing
1. Unit tests for each tool handler
2. Integration tests using MCP Inspector CLI mode
3. Acceptance tests using Python scripts
4. Coverage verification (>= 80%)

# Research: MCP Server for AgentMail

**Date**: 2026-01-14
**Feature**: 010-mcp-server
**Status**: Complete

## Research Topics

### 1. Go MCP SDK Selection

**Decision**: Use the official `modelcontextprotocol/go-sdk` package

**Rationale**:
- Official SDK maintained in collaboration with Google
- Native STDIO transport support via `mcp.StdioTransport{}`
- Supports MCP specification versions from 2024-11-05 through 2025-11-25
- Clean API for tool registration: `mcp.AddTool(server, &mcp.Tool{...}, handler)`
- Well-documented with examples in `/examples` directory

**Alternatives Considered**:

| Library | Pros | Cons | Why Rejected |
|---------|------|------|--------------|
| [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) | Popular community SDK, simple API | External dependency, not official | Prefer official SDK per constitution IV |
| [metoro-io/mcp-golang](https://github.com/metoro-io/mcp-golang) | Type-safe, minimal boilerplate | Less mature, unofficial | Prefer official SDK per constitution IV |
| Custom JSON-RPC implementation | No external dependencies | Significant effort, error-prone | MCP spec is complex, SDK is justified |

**Dependency Justification (per Constitution IV)**:
- Standard library is insufficient for full MCP protocol compliance
- MCP protocol includes complex handshakes, capability negotiation, and JSON Schema generation
- Official SDK is maintained by protocol authors, ensuring specification compliance
- Security/maintenance: Official SDK is actively maintained and widely adopted (8M+ downloads)

### 2. MCP Testing Tools

**Decision**: Use MCP Inspector CLI mode as primary testing tool (replaces mcp-cli)

**Rationale**:
- Official tool from Model Context Protocol team
- CLI mode with `--cli` flag outputs JSON for automation
- Supports STDIO transport directly
- No installation required (`npx @modelcontextprotocol/inspector --cli`)
- Can be combined with `jq` for test assertions

**Alternative Tools Evaluated**:

| Tool | Type | Pros | Cons |
|------|------|------|------|
| [MCP Inspector](https://github.com/modelcontextprotocol/inspector) | Official | CLI mode, JSON output, no install | Requires Node.js/npx |
| [f/mcptools](https://github.com/f/mcptools) | Community (Go) | Native Go, multiple formats | External tool, less mature |
| [apify/mcp-cli](https://github.com/apify/mcp-cli) | Community | JSON output, schema validation | External dependency |
| Direct JSON-RPC piping | Manual | No dependencies | Manual construction, error-prone |

**Testing Strategy**:

1. **Unit Tests**: Standard Go tests with mocked interfaces
   - Test MCP tool handlers in isolation
   - Mock existing CLI functions (Send, Receive, Status, Recipients)
   - Test JSON-RPC message parsing and generation

2. **Integration Tests**: MCP Inspector CLI mode
   ```bash
   # List tools
   npx @modelcontextprotocol/inspector --cli tools-list ./agentmail mcp

   # Call a tool
   npx @modelcontextprotocol/inspector --cli call-tool send '{"recipient":"agent2","message":"Hello"}' ./agentmail mcp
   ```

3. **Acceptance Tests**: Custom Python scripts (per spec clarification)
   - Use Python `subprocess` to spawn MCP server
   - Send JSON-RPC requests via stdin
   - Validate responses match expected format
   - Scripts removed after testing (per spec)

### 3. MCP Server Architecture Pattern

**Decision**: Add `mcp` subcommand to existing `agentmail` binary

**Rationale**:
- Maintains single binary distribution (per constitution I: CLI-First Design)
- Reuses existing CLI logic (Send, Receive, Status, Recipients functions)
- No daemon process required - runs as subprocess of AI agent
- STDIO transport means no network configuration needed

**Implementation Pattern**:

```go
// New subcommand: agentmail mcp
mcpCmd := &ffcli.Command{
    Name:       "mcp",
    ShortUsage: "agentmail mcp",
    ShortHelp:  "Start MCP server (STDIO transport)",
    Exec: func(ctx context.Context, args []string) error {
        return runMCPServer(ctx)
    },
}

func runMCPServer(ctx context.Context) error {
    server := mcp.NewServer(&mcp.Implementation{
        Name:    "agentmail",
        Version: version,
    }, nil)

    // Register tools using existing CLI logic
    mcp.AddTool(server, &mcp.Tool{Name: "send", ...}, sendHandler)
    mcp.AddTool(server, &mcp.Tool{Name: "receive", ...}, receiveHandler)
    mcp.AddTool(server, &mcp.Tool{Name: "status", ...}, statusHandler)
    mcp.AddTool(server, &mcp.Tool{Name: "list-recipients", ...}, recipientsHandler)

    return server.Run(ctx, &mcp.StdioTransport{})
}
```

### 4. Tool Parameter Schema Design

**Decision**: Use JSON Schema for tool parameters (MCP specification requirement)

**Tool Schemas**:

| Tool | Parameters | Returns |
|------|------------|---------|
| `send` | `recipient` (string, required), `message` (string, required) | `{"message_id": "ABC123"}` or error |
| `receive` | None | Message object or "No unread messages" |
| `status` | `status` (enum: ready/work/offline, required) | Success confirmation or error |
| `list-recipients` | None | Array of recipient names |

### 5. Error Handling Pattern

**Decision**: Return MCP-compliant error responses

**Pattern**:
```go
// Success response
return mcp.NewToolResultText("Message #ABC123 sent"), nil

// Error response (MCP-compliant)
return mcp.NewToolResultError("recipient not found"), nil

// JSON-RPC errors (protocol level)
// Handled automatically by SDK for malformed requests
```

### 6. Tmux Context Handling

**Decision**: Inherit tmux context from parent process

**Rationale**:
- MCP server runs as subprocess of AI agent (e.g., Claude Code)
- Agent runs inside tmux window
- Environment variables (`$TMUX`, `$TMUX_PANE`) inherited automatically
- No special handling required - existing tmux detection works

**Edge Case**: If tmux context lost mid-session (FR-014):
```go
// Detect tmux loss on each tool call
if !tmux.InTmux() {
    // Log to stderr (FR-015)
    log.Println("tmux session terminated")
    os.Exit(1)  // Exit within 1 second (FR-014)
}
```

### 7. JSON-RPC Error Codes

**Decision**: Use standard JSON-RPC 2.0 error codes

| Error | Code | When Used |
|-------|------|-----------|
| Parse Error | -32700 | Malformed JSON received (FR-010) |
| Invalid Request | -32600 | Invalid JSON-RPC structure |
| Method Not Found | -32601 | Unknown tool name |
| Invalid Params | -32602 | Missing or invalid tool parameters |

## Sources

- [Official Go SDK](https://github.com/modelcontextprotocol/go-sdk) - Model Context Protocol official Go implementation
- [MCP Inspector](https://modelcontextprotocol.io/docs/tools/inspector) - Official testing tool documentation
- [f/mcptools](https://github.com/f/mcptools) - Community CLI testing tool
- [apify/mcp-cli](https://github.com/apify/mcp-cli) - Universal MCP command-line client
- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) - Community Go SDK
- [MCP Specification](https://modelcontextprotocol.io/specification/2025-06-18/server/tools) - Tools specification

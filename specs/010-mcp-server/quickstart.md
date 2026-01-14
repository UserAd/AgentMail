# Quickstart: MCP Server for AgentMail

**Feature**: 010-mcp-server
**Date**: 2026-01-14

## Overview

This guide explains how to use AgentMail's MCP server for inter-agent communication.

## Starting the MCP Server

```bash
# Start MCP server (STDIO transport)
agentmail mcp
```

The server reads JSON-RPC requests from stdin and writes responses to stdout. Logs and errors go to stderr.

## Configuring AI Assistants

### Claude Code / Claude Desktop

Add to `~/.claude/settings.json` or your project's `.mcp.json`:

```json
{
  "mcpServers": {
    "agentmail": {
      "command": "agentmail",
      "args": ["mcp"]
    }
  }
}
```

### VS Code with Copilot

Add to your settings or MCP configuration:

```json
{
  "mcp.servers": {
    "agentmail": {
      "command": "agentmail",
      "args": ["mcp"]
    }
  }
}
```

## Available Tools

### send

Send a message to another agent.

**Parameters:**
- `recipient` (string, required): Target tmux window name
- `message` (string, required): Message content (max 64KB)

**Example:**
```json
{
  "name": "send",
  "arguments": {
    "recipient": "agent2",
    "message": "Hello from agent1!"
  }
}
```

**Response:**
```json
{
  "message_id": "ABC123"
}
```

### receive

Read the oldest unread message.

**Parameters:** None

**Example:**
```json
{
  "name": "receive",
  "arguments": {}
}
```

**Response (message available):**
```json
{
  "from": "agent2",
  "id": "XYZ789",
  "message": "Hello back!"
}
```

**Response (no messages):**
```json
{
  "status": "No unread messages"
}
```

### status

Set your availability status for agent coordination. Setting status to `work` or `offline` resets the notification flag, allowing future notifications when returning to `ready`.

**Parameters:**
- `status` (string, required): One of `ready`, `work`, `offline`

**Example:**
```json
{
  "name": "status",
  "arguments": {
    "status": "ready"
  }
}
```

**Response:**
```json
{
  "status": "ok"
}
```

### list-recipients

List available agents.

**Parameters:** None

**Example:**
```json
{
  "name": "list-recipients",
  "arguments": {}
}
```

**Response:**
```json
{
  "recipients": [
    {"name": "agent1", "is_current": true},
    {"name": "agent2", "is_current": false},
    {"name": "agent3", "is_current": false}
  ]
}
```

## Testing with MCP Inspector

### List Available Tools

```bash
npx @modelcontextprotocol/inspector --cli tools-list agentmail mcp
```

### Call a Tool

```bash
# Send a message
npx @modelcontextprotocol/inspector --cli call-tool send \
  '{"recipient":"agent2","message":"Test message"}' \
  agentmail mcp

# Receive a message
npx @modelcontextprotocol/inspector --cli call-tool receive '{}' \
  agentmail mcp

# Set status
npx @modelcontextprotocol/inspector --cli call-tool status \
  '{"status":"ready"}' \
  agentmail mcp

# List recipients
npx @modelcontextprotocol/inspector --cli call-tool list-recipients '{}' \
  agentmail mcp
```

## Requirements

- Must be running inside a tmux session
- AgentMail binary must be in PATH
- Git repository with `.agentmail/` directory

## Error Handling

Errors are returned as MCP tool result errors:

```json
{
  "error": "recipient not found"
}
```

Common errors:
- `recipient not found` - Invalid or ignored recipient
- `no message provided` - Empty message in send
- `Invalid status: X. Valid: ready, work, offline` - Bad status value
- `message exceeds 64KB limit` - Message too large

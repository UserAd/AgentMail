// Package mcp provides an MCP (Model Context Protocol) server implementation
// for AgentMail, enabling AI agents to communicate via STDIO transport.
//
// The MCP server exposes AgentMail functionality through four tools:
//
//   - send: Send a message to another agent in the tmux session
//   - receive: Receive the oldest unread message from the agent's mailbox
//   - status: Set the agent's availability status (ready/work/offline)
//   - list-recipients: List all available agents in the current tmux session
//
// The server uses the official MCP Go SDK from github.com/modelcontextprotocol/go-sdk
// and communicates over STDIO transport, making it suitable for integration with
// AI agents and IDE extensions that support the MCP protocol.
//
// Usage:
//
//	agentmail mcp
package mcp

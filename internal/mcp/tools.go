package mcp

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MaxMessageSize is the maximum allowed message size (64KB per FR-002).
const MaxMessageSize = 65536

// Tool names as constants for consistent reference.
const (
	ToolSend           = "send"
	ToolReceive        = "receive"
	ToolStatus         = "status"
	ToolListRecipients = "list-recipients"
)

// SendArgs represents the input parameters for the send tool.
type SendArgs struct {
	// Recipient is the tmux window name of the recipient agent.
	Recipient string `json:"recipient"`
	// Message is the message content to send (max 64KB).
	Message string `json:"message"`
}

// ReceiveArgs represents the input parameters for the receive tool.
// It has no parameters.
type ReceiveArgs struct{}

// StatusArgs represents the input parameters for the status tool.
type StatusArgs struct {
	// Status is the availability status to set.
	Status string `json:"status"`
}

// ListRecipientsArgs represents the input parameters for the list-recipients tool.
// It has no parameters.
type ListRecipientsArgs struct{}

// sendToolSchema returns the JSON schema for the send tool input.
// We define this manually to include maxLength constraint on message.
func sendToolSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"recipient": {
				"type": "string",
				"description": "The tmux window name of the recipient agent"
			},
			"message": {
				"type": "string",
				"description": "The message content to send (max 64KB)",
				"maxLength": 65536
			}
		},
		"required": ["recipient", "message"]
	}`)
}

// receiveToolSchema returns the JSON schema for the receive tool input.
func receiveToolSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {}
	}`)
}

// statusToolSchema returns the JSON schema for the status tool input.
// Includes enum constraint for status values.
func statusToolSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"status": {
				"type": "string",
				"description": "The availability status to set",
				"enum": ["ready", "work", "offline"]
			}
		},
		"required": ["status"]
	}`)
}

// listRecipientsToolSchema returns the JSON schema for the list-recipients tool input.
func listRecipientsToolSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {}
	}`)
}

// RegisterTools registers all AgentMail tools with the MCP server.
// Tool handlers are stubs that return "not implemented" - actual implementation
// will be added in later phases.
func RegisterTools(s *Server) {
	mcpServer := s.MCPServer()

	// Register send tool with explicit schema
	mcpServer.AddTool(&mcp.Tool{
		Name:        ToolSend,
		Description: "Send a message to another agent in a tmux window",
		InputSchema: sendToolSchema(),
	}, sendHandler)

	// Register receive tool with explicit schema
	mcpServer.AddTool(&mcp.Tool{
		Name:        ToolReceive,
		Description: "Read the oldest unread message from your mailbox",
		InputSchema: receiveToolSchema(),
	}, receiveHandler)

	// Register status tool with explicit schema (includes enum)
	mcpServer.AddTool(&mcp.Tool{
		Name:        ToolStatus,
		Description: "Set your agent's availability status",
		InputSchema: statusToolSchema(),
	}, statusHandler)

	// Register list-recipients tool with explicit schema
	mcpServer.AddTool(&mcp.Tool{
		Name:        ToolListRecipients,
		Description: "List all available agents that can receive messages",
		InputSchema: listRecipientsToolSchema(),
	}, listRecipientsHandler)
}

// sendHandler handles the send tool invocation.
// Delegates to handleSend in handlers.go for actual implementation.
func sendHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return handleSend(ctx, req)
}

// receiveHandler handles the receive tool invocation.
// Delegates to handleReceive in handlers.go for actual implementation.
func receiveHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return handleReceive(ctx, req)
}

// statusHandler handles the status tool invocation.
// Delegates to handleStatus in handlers.go for actual implementation.
func statusHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return handleStatus(ctx, req)
}

// listRecipientsHandler handles the list-recipients tool invocation.
// Delegates to handleListRecipients in handlers.go for actual implementation.
func listRecipientsHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return handleListRecipients(ctx, req)
}

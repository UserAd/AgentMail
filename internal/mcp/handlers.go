package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"agentmail/internal/mail"
	"agentmail/internal/tmux"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// HandlerOptions configures handler behavior for testing.
type HandlerOptions struct {
	// SkipTmuxCheck disables tmux validation (for testing).
	SkipTmuxCheck bool
	// MockReceiver is the mock receiver window name (for testing).
	MockReceiver string
	// RepoRoot is the repository root (defaults to git root).
	RepoRoot string
}

// handlerOptions holds the current handler options.
// Set via SetHandlerOptions for testing, nil for production.
var handlerOptions *HandlerOptions

// SetHandlerOptions sets the handler options for testing.
// Pass nil to reset to production behavior.
func SetHandlerOptions(opts *HandlerOptions) {
	handlerOptions = opts
}

// ReceiveResponse represents a successful receive response with a message.
type ReceiveResponse struct {
	From    string `json:"from"`    // Sender window name
	ID      string `json:"id"`      // Message ID
	Message string `json:"message"` // Message content
}

// ReceiveEmptyResponse represents a response when no messages are available.
type ReceiveEmptyResponse struct {
	Status string `json:"status"` // "No unread messages"
}

// doReceive implements the receive handler logic.
// It returns the response as a map for JSON encoding, or an error.
func doReceive(ctx context.Context) (any, error) {
	opts := handlerOptions
	if opts == nil {
		opts = &HandlerOptions{}
	}

	// Get receiver identity
	var receiver string
	if opts.MockReceiver != "" {
		receiver = opts.MockReceiver
	} else {
		var err error
		receiver, err = tmux.GetCurrentWindow()
		if err != nil {
			return nil, fmt.Errorf("failed to get current window: %w", err)
		}
	}

	// Determine repository root
	repoRoot := opts.RepoRoot
	if repoRoot == "" {
		var err error
		repoRoot, err = mail.FindGitRoot()
		if err != nil {
			return nil, fmt.Errorf("not in a git repository: %w", err)
		}
	}

	// Find unread messages for receiver (FR-003: FIFO order)
	unread, err := mail.FindUnread(repoRoot, receiver)
	if err != nil {
		return nil, fmt.Errorf("failed to read messages: %w", err)
	}

	// FR-008: Handle no unread messages
	if len(unread) == 0 {
		return ReceiveEmptyResponse{
			Status: "No unread messages",
		}, nil
	}

	// Get oldest unread message (FIFO - first in list) per FR-003
	msg := unread[0]

	// FR-012: Mark as read
	if err := mail.MarkAsRead(repoRoot, receiver, msg.ID); err != nil {
		return nil, fmt.Errorf("failed to mark message as read: %w", err)
	}

	// Return response with from, id, message fields per data-model.md
	return ReceiveResponse{
		From:    msg.From,
		ID:      msg.ID,
		Message: msg.Message,
	}, nil
}

// handleReceive is the MCP handler function for the receive tool.
// It wraps doReceive and formats the response as MCP content.
func handleReceive(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	response, err := doReceive(ctx)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: err.Error()},
			},
		}, nil
	}

	// Encode response as JSON
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("failed to encode response: %v", err)},
			},
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}

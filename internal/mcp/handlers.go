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
	// MockSender is the mock sender window name (for testing).
	MockSender string
	// MockWindows is the mock list of tmux windows (for testing).
	MockWindows []string
	// MockIgnoreList is the mock ignore list (for testing).
	MockIgnoreList map[string]bool
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

// SendResponse represents a successful send response.
type SendResponse struct {
	MessageID string `json:"message_id"` // Generated message ID
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

// StatusResponse represents a successful status update response.
type StatusResponse struct {
	Status string `json:"status"` // "ok" on success
}

// ListRecipientsResponse represents the response from the list-recipients tool.
type ListRecipientsResponse struct {
	Recipients []RecipientInfo `json:"recipients"`
}

// RecipientInfo represents a single recipient in the list-recipients response.
type RecipientInfo struct {
	Name      string `json:"name"`       // Window name
	IsCurrent bool   `json:"is_current"` // True if this is the caller's window
}

// MaxMessageSizeBytes is the maximum allowed message size (64KB per FR-013).
const MaxMessageSizeBytes = 65536

// doSend implements the send handler logic.
// It validates the message, stores it, and returns the response or an error.
func doSend(ctx context.Context, recipient, message string) (any, error) {
	opts := handlerOptions
	if opts == nil {
		opts = &HandlerOptions{}
	}

	// Validate message is not empty
	if message == "" {
		return nil, fmt.Errorf("no message provided")
	}

	// FR-013: Validate message size (64KB limit)
	if len(message) > MaxMessageSizeBytes {
		return nil, fmt.Errorf("message exceeds maximum size of 64KB")
	}

	// Get sender identity
	var sender string
	if opts.MockSender != "" {
		sender = opts.MockSender
	} else {
		var err error
		sender, err = tmux.GetCurrentWindow()
		if err != nil {
			return nil, fmt.Errorf("failed to get current window: %w", err)
		}
	}

	// FR-009: Validate recipient exists
	var recipientExists bool
	if opts.MockWindows != nil {
		for _, w := range opts.MockWindows {
			if w == recipient {
				recipientExists = true
				break
			}
		}
	} else {
		var err error
		recipientExists, err = tmux.WindowExists(recipient)
		if err != nil {
			return nil, fmt.Errorf("failed to check recipient: %w", err)
		}
	}

	if !recipientExists {
		return nil, fmt.Errorf("recipient not found")
	}

	// Check if sending to self (not allowed)
	if recipient == sender {
		return nil, fmt.Errorf("cannot send message to self")
	}

	// Load and check ignore list
	var ignoreList map[string]bool
	if opts.MockIgnoreList != nil {
		ignoreList = opts.MockIgnoreList
	} else {
		// Determine git root for loading ignore list
		gitRoot := opts.RepoRoot
		if gitRoot == "" {
			gitRoot, _ = mail.FindGitRoot()
		}
		if gitRoot != "" {
			ignoreList, _ = mail.LoadIgnoreList(gitRoot)
		}
	}

	// Check if recipient is in ignore list
	if ignoreList != nil && ignoreList[recipient] {
		return nil, fmt.Errorf("recipient not found")
	}

	// Generate message ID
	id, err := mail.GenerateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate message ID: %w", err)
	}

	// Determine repository root
	repoRoot := opts.RepoRoot
	if repoRoot == "" {
		repoRoot, err = mail.FindGitRoot()
		if err != nil {
			return nil, fmt.Errorf("not in a git repository: %w", err)
		}
	}

	// Store message
	msg := mail.Message{
		ID:       id,
		From:     sender,
		To:       recipient,
		Message:  message,
		ReadFlag: false,
	}

	if err := mail.Append(repoRoot, msg); err != nil {
		return nil, fmt.Errorf("failed to write message: %w", err)
	}

	// FR-004: Return response with message_id
	return SendResponse{
		MessageID: id,
	}, nil
}

// sendParams holds the unmarshaled parameters for the send tool.
type sendParams struct {
	Recipient string `json:"recipient"`
	Message   string `json:"message"`
}

// handleSend is the MCP handler function for the send tool.
// It wraps doSend and formats the response as MCP content.
func handleSend(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters from request by unmarshaling JSON
	var params sendParams
	if req.Params != nil && req.Params.Arguments != nil {
		if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("failed to parse arguments: %v", err)},
				},
			}, nil
		}
	}

	response, err := doSend(ctx, params.Recipient, params.Message)
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

// ValidStatus values for the status tool.
var validStatuses = map[string]bool{
	mail.StatusReady:   true,
	mail.StatusWork:    true,
	mail.StatusOffline: true,
}

// validateStatus checks if the provided status is valid.
// Returns true if valid, false otherwise.
func validateStatus(status string) bool {
	return validStatuses[status]
}

// doStatus implements the status handler logic.
// It validates the status, updates the recipient state, and returns the response or an error.
func doStatus(ctx context.Context, status string) (any, error) {
	opts := handlerOptions
	if opts == nil {
		opts = &HandlerOptions{}
	}

	// T039: Validate status value (ready/work/offline only)
	if !validateStatus(status) {
		// T040: Return error message matching FR-016 format
		return nil, fmt.Errorf("Invalid status: %s. Valid: ready, work, offline", status)
	}

	// Get agent identity (current tmux window)
	var agent string
	if opts.MockReceiver != "" {
		agent = opts.MockReceiver
	} else {
		var err error
		agent, err = tmux.GetCurrentWindow()
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

	// Reset notified flag when status is work or offline
	resetNotified := (status == mail.StatusWork || status == mail.StatusOffline)

	// Update recipient state using existing mail infrastructure
	if err := mail.UpdateRecipientState(repoRoot, agent, status, resetNotified); err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	// T041: Return {"status": "ok"} on success
	return StatusResponse{
		Status: "ok",
	}, nil
}

// statusParams holds the unmarshaled parameters for the status tool.
type statusParams struct {
	Status string `json:"status"`
}

// handleStatus is the MCP handler function for the status tool.
// It wraps doStatus and formats the response as MCP content.
func handleStatus(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters from request by unmarshaling JSON
	var params statusParams
	if req.Params != nil && req.Params.Arguments != nil {
		if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("failed to parse arguments: %v", err)},
				},
			}, nil
		}
	}

	response, err := doStatus(ctx, params.Status)
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

// doListRecipients implements the list-recipients handler logic.
// It returns all available agents (tmux windows) with the current window marked.
// Ignored windows are excluded, but current window is always shown.
func doListRecipients(ctx context.Context) (any, error) {
	opts := handlerOptions
	if opts == nil {
		opts = &HandlerOptions{}
	}

	// Get current window (agent identity)
	var currentWindow string
	if opts.MockReceiver != "" {
		currentWindow = opts.MockReceiver
	} else {
		var err error
		currentWindow, err = tmux.GetCurrentWindow()
		if err != nil {
			return nil, fmt.Errorf("failed to get current window: %w", err)
		}
	}

	// Get list of all windows
	var windows []string
	if opts.MockWindows != nil {
		windows = opts.MockWindows
	} else {
		var err error
		windows, err = tmux.ListWindows()
		if err != nil {
			return nil, fmt.Errorf("failed to list windows: %w", err)
		}
	}

	// Load ignore list
	var ignoreList map[string]bool
	if opts.MockIgnoreList != nil {
		ignoreList = opts.MockIgnoreList
	} else {
		// Determine git root for loading ignore list
		gitRoot := opts.RepoRoot
		if gitRoot == "" {
			gitRoot, _ = mail.FindGitRoot()
		}
		if gitRoot != "" {
			ignoreList, _ = mail.LoadIgnoreList(gitRoot)
		}
	}

	// Build recipients list, filtering ignored windows but always including current
	recipients := []RecipientInfo{}
	for _, window := range windows {
		// Current window is always shown (even if in ignore list)
		if window == currentWindow {
			recipients = append(recipients, RecipientInfo{
				Name:      window,
				IsCurrent: true,
			})
		} else if ignoreList == nil || !ignoreList[window] {
			// Only show non-current windows if they're not in the ignore list
			recipients = append(recipients, RecipientInfo{
				Name:      window,
				IsCurrent: false,
			})
		}
	}

	return ListRecipientsResponse{
		Recipients: recipients,
	}, nil
}

// handleListRecipients is the MCP handler function for the list-recipients tool.
// It wraps doListRecipients and formats the response as MCP content.
func handleListRecipients(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	response, err := doListRecipients(ctx)
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

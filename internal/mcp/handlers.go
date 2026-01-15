package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

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

// getSender returns the sender window name.
func (opts *HandlerOptions) getSender() (string, error) {
	if opts.MockSender != "" {
		return opts.MockSender, nil
	}
	return tmux.GetCurrentWindow()
}

// getReceiver returns the receiver window name.
func (opts *HandlerOptions) getReceiver() (string, error) {
	if opts.MockReceiver != "" {
		return opts.MockReceiver, nil
	}
	return tmux.GetCurrentWindow()
}

// windowExists checks if a window exists.
func (opts *HandlerOptions) windowExists(window string) (bool, error) {
	if opts.MockWindows != nil {
		for _, w := range opts.MockWindows {
			if w == window {
				return true, nil
			}
		}
		return false, nil
	}
	return tmux.WindowExists(window)
}

// listWindows returns all tmux windows.
func (opts *HandlerOptions) listWindows() ([]string, error) {
	if opts.MockWindows != nil {
		return opts.MockWindows, nil
	}
	return tmux.ListWindows()
}

// loadIgnoreList loads the ignore list.
func (opts *HandlerOptions) loadIgnoreList() map[string]bool {
	if opts.MockIgnoreList != nil {
		return opts.MockIgnoreList
	}
	gitRoot := opts.RepoRoot
	if gitRoot == "" {
		gitRoot, _ = mail.FindGitRoot()
	}
	if gitRoot == "" {
		return nil
	}
	ignoreList, _ := mail.LoadIgnoreList(gitRoot)
	return ignoreList
}

// getRepoRoot returns the repository root.
func (opts *HandlerOptions) getRepoRoot() (string, error) {
	if opts.RepoRoot != "" {
		return opts.RepoRoot, nil
	}
	return mail.FindGitRoot()
}

// handlerOptions holds the current handler options.
// Set via SetHandlerOptions for testing, nil for production.
// Protected by handlerOptionsMu for thread-safe access.
var (
	handlerOptions   *HandlerOptions
	handlerOptionsMu sync.RWMutex
)

// SetHandlerOptions sets the handler options for testing.
// Pass nil to reset to production behavior.
func SetHandlerOptions(opts *HandlerOptions) {
	handlerOptionsMu.Lock()
	defer handlerOptionsMu.Unlock()
	handlerOptions = opts
}

// getHandlerOptions returns the current handler options in a thread-safe manner.
func getHandlerOptions() *HandlerOptions {
	handlerOptionsMu.RLock()
	defer handlerOptionsMu.RUnlock()
	return handlerOptions
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

// doSend implements the send handler logic.
// It validates the message, stores it, and returns the response or an error.
func doSend(ctx context.Context, recipient, message string) (any, error) {
	opts := getHandlerOptions()
	if opts == nil {
		opts = &HandlerOptions{}
	}

	// Validate message
	if message == "" {
		return nil, fmt.Errorf("no message provided")
	}
	if len(message) > MaxMessageSize {
		return nil, fmt.Errorf("message exceeds maximum size of 64KB")
	}

	// Get sender identity
	sender, err := opts.getSender()
	if err != nil {
		return nil, fmt.Errorf("failed to get current window: %w", err)
	}

	// Validate recipient exists
	exists, err := opts.windowExists(recipient)
	if err != nil {
		return nil, fmt.Errorf("failed to check recipient: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("recipient not found")
	}

	// Check self-send
	if recipient == sender {
		return nil, fmt.Errorf("cannot send message to self")
	}

	// Check ignore list
	if ignoreList := opts.loadIgnoreList(); ignoreList != nil && ignoreList[recipient] {
		return nil, fmt.Errorf("recipient not found")
	}

	// Get repository root
	repoRoot, err := opts.getRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("not in a git repository: %w", err)
	}

	// Generate ID and store message
	id, err := mail.GenerateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate message ID: %w", err)
	}

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

	return SendResponse{MessageID: id}, nil
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
	opts := getHandlerOptions()
	if opts == nil {
		opts = &HandlerOptions{}
	}

	// Get receiver identity
	receiver, err := opts.getReceiver()
	if err != nil {
		return nil, fmt.Errorf("failed to get current window: %w", err)
	}

	// Get repository root
	repoRoot, err := opts.getRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("not in a git repository: %w", err)
	}

	// Find unread messages (FIFO order)
	unread, err := mail.FindUnread(repoRoot, receiver)
	if err != nil {
		return nil, fmt.Errorf("failed to read messages: %w", err)
	}
	if len(unread) == 0 {
		return ReceiveEmptyResponse{Status: "No unread messages"}, nil
	}

	// Get oldest and mark as read
	msg := unread[0]
	if err := mail.MarkAsRead(repoRoot, receiver, msg.ID); err != nil {
		return nil, fmt.Errorf("failed to mark message as read: %w", err)
	}

	return ReceiveResponse{From: msg.From, ID: msg.ID, Message: msg.Message}, nil
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
	opts := getHandlerOptions()
	if opts == nil {
		opts = &HandlerOptions{}
	}

	// Validate status
	if !validateStatus(status) {
		return nil, fmt.Errorf("Invalid status: %s. Valid: ready, work, offline", status)
	}

	// Get agent identity
	agent, err := opts.getReceiver()
	if err != nil {
		return nil, fmt.Errorf("failed to get current window: %w", err)
	}

	// Get repository root
	repoRoot, err := opts.getRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("not in a git repository: %w", err)
	}

	// Update state (reset notified for work/offline)
	resetNotified := status == mail.StatusWork || status == mail.StatusOffline
	if err := mail.UpdateRecipientState(repoRoot, agent, status, resetNotified); err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	return StatusResponse{Status: "ok"}, nil
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
	opts := getHandlerOptions()
	if opts == nil {
		opts = &HandlerOptions{}
	}

	// Get current window
	currentWindow, err := opts.getReceiver()
	if err != nil {
		return nil, fmt.Errorf("failed to get current window: %w", err)
	}

	// Get all windows
	windows, err := opts.listWindows()
	if err != nil {
		return nil, fmt.Errorf("failed to list windows: %w", err)
	}

	// Build filtered recipients list
	ignoreList := opts.loadIgnoreList()
	recipients := make([]RecipientInfo, 0, len(windows))
	for _, window := range windows {
		if window == currentWindow {
			recipients = append(recipients, RecipientInfo{Name: window, IsCurrent: true})
		} else if ignoreList == nil || !ignoreList[window] {
			recipients = append(recipients, RecipientInfo{Name: window, IsCurrent: false})
		}
	}

	return ListRecipientsResponse{Recipients: recipients}, nil
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

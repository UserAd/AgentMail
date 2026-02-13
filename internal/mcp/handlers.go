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
	// MockPaneAddress is the mock pane address for sender/receiver (for testing).
	MockPaneAddress string
	// MockSession is the mock session name (for testing).
	MockSession string
	// MockPanes is the mock list of tmux panes (for testing).
	MockPanes []string
	// MockIgnoreList is the mock ignore list (for testing).
	MockIgnoreList map[string]bool
	// RepoRoot is the repository root (defaults to git root).
	RepoRoot string
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
	From    string `json:"from"`    // Sender pane address
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
	Address   string `json:"address"`    // Pane address (session:window.pane)
	IsCurrent bool   `json:"is_current"` // True if this is the caller's pane
}

// doSend implements the send handler logic.
// It validates the message, stores it, and returns the response or an error.
func doSend(ctx context.Context, recipient, message string) (any, error) {
	opts := getHandlerOptions()
	if opts == nil {
		opts = &HandlerOptions{}
	}

	// Validate message is not empty
	if message == "" {
		return nil, fmt.Errorf("no message provided")
	}

	// FR-013: Validate message size (64KB limit)
	if len(message) > MaxMessageSize {
		return nil, fmt.Errorf("message exceeds maximum size of 64KB")
	}

	// Get sender identity (pane address)
	var sender string
	if opts.MockPaneAddress != "" {
		sender = opts.MockPaneAddress
	} else {
		var err error
		sender, err = tmux.GetCurrentPaneAddress()
		if err != nil {
			return nil, fmt.Errorf("failed to get current pane address: %w", err)
		}
	}

	// Get current session for address resolution
	var currentSession string
	if opts.MockSession != "" {
		currentSession = opts.MockSession
	} else {
		// Try to extract session from sender address
		addr, err := tmux.ParseAddress(sender, "")
		if err == nil {
			currentSession = addr.Session
		} else {
			var sessionErr error
			currentSession, sessionErr = tmux.GetCurrentSession()
			if sessionErr != nil {
				return nil, fmt.Errorf("failed to get current session: %w", sessionErr)
			}
		}
	}

	// Parse recipient address
	addr, err := tmux.ParseAddress(recipient, currentSession)
	if err != nil {
		return nil, fmt.Errorf("recipient not found")
	}

	// Resolve recipient address
	var resolvedRecipient string
	if addr.Pane == -1 {
		// Short form - need to resolve window name to pane address
		var panes []string
		if opts.MockPanes != nil {
			panes = opts.MockPanes
		} else {
			panes, err = tmux.ListPanes()
			if err != nil {
				return nil, fmt.Errorf("failed to list panes: %w", err)
			}
		}

		// Find matching panes
		var matches []string
		for _, pane := range panes {
			paneAddr, parseErr := tmux.ParseAddress(pane, currentSession)
			if parseErr == nil && paneAddr.Window == addr.Window {
				matches = append(matches, pane)
			}
		}

		if len(matches) == 0 {
			return nil, fmt.Errorf("recipient not found")
		} else if len(matches) > 1 {
			// Ambiguous - suggest all matching panes
			return nil, fmt.Errorf("Ambiguous recipient: window '%s' has %d panes. Use %s", addr.Window, len(matches), formatAddressList(matches))
		}
		resolvedRecipient = matches[0]
	} else {
		// Full or medium form - validate pane exists
		fullAddr := tmux.FormatAddress(addr)
		var exists bool
		if opts.MockPanes != nil {
			for _, p := range opts.MockPanes {
				if p == fullAddr {
					exists = true
					break
				}
			}
		} else {
			exists, err = tmux.PaneExists(fullAddr)
			if err != nil {
				return nil, fmt.Errorf("failed to check recipient: %w", err)
			}
		}
		if !exists {
			return nil, fmt.Errorf("recipient not found")
		}
		resolvedRecipient = fullAddr
	}

	// Check if sending to self (not allowed)
	if resolvedRecipient == sender {
		return nil, fmt.Errorf("cannot send message to self")
	}

	// Load and check ignore list
	var ignoreList map[string]bool
	if opts.MockIgnoreList != nil {
		ignoreList = opts.MockIgnoreList
	} else {
		gitRoot := opts.RepoRoot
		if gitRoot == "" {
			gitRoot, _ = mail.FindGitRoot()
		}
		if gitRoot != "" {
			ignoreList, _ = mail.LoadIgnoreList(gitRoot)
		}
	}

	// Check if recipient is in ignore list
	if mail.IsIgnored(resolvedRecipient, ignoreList, currentSession) {
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
		To:       resolvedRecipient,
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

// formatAddressList formats a list of addresses for error messages.
func formatAddressList(addresses []string) string {
	if len(addresses) == 0 {
		return ""
	}
	if len(addresses) == 1 {
		return addresses[0]
	}
	result := ""
	for i, addr := range addresses {
		if i > 0 {
			if i == len(addresses)-1 {
				result += ", or "
			} else {
				result += ", "
			}
		}
		result += addr
	}
	return result
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

	// Get receiver identity (pane address)
	var receiver string
	if opts.MockPaneAddress != "" {
		receiver = opts.MockPaneAddress
	} else {
		var err error
		receiver, err = tmux.GetCurrentPaneAddress()
		if err != nil {
			return nil, fmt.Errorf("failed to get current pane address: %w", err)
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
	opts := getHandlerOptions()
	if opts == nil {
		opts = &HandlerOptions{}
	}

	// T039: Validate status value (ready/work/offline only)
	if !validateStatus(status) {
		// T040: Return error message matching FR-016 format
		return nil, fmt.Errorf("Invalid status: %s. Valid: ready, work, offline", status)
	}

	// Get agent identity (current tmux pane address)
	var agent string
	if opts.MockPaneAddress != "" {
		agent = opts.MockPaneAddress
	} else {
		var err error
		agent, err = tmux.GetCurrentPaneAddress()
		if err != nil {
			return nil, fmt.Errorf("failed to get current pane address: %w", err)
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
// It returns all available agents (tmux panes) with the current pane marked.
// Ignored panes are excluded, but current pane is always shown.
func doListRecipients(ctx context.Context) (any, error) {
	opts := getHandlerOptions()
	if opts == nil {
		opts = &HandlerOptions{}
	}

	// Get current pane (agent identity)
	var currentPane string
	if opts.MockPaneAddress != "" {
		currentPane = opts.MockPaneAddress
	} else {
		var err error
		currentPane, err = tmux.GetCurrentPaneAddress()
		if err != nil {
			return nil, fmt.Errorf("failed to get current pane address: %w", err)
		}
	}

	// Get current session for ignore matching
	var currentSession string
	if opts.MockSession != "" {
		currentSession = opts.MockSession
	} else {
		addr, err := tmux.ParseAddress(currentPane, "")
		if err == nil {
			currentSession = addr.Session
		} else {
			var sessionErr error
			currentSession, sessionErr = tmux.GetCurrentSession()
			if sessionErr != nil {
				return nil, fmt.Errorf("failed to get current session: %w", sessionErr)
			}
		}
	}

	// Get list of all panes
	var panes []string
	if opts.MockPanes != nil {
		panes = opts.MockPanes
	} else {
		var err error
		panes, err = tmux.ListPanes()
		if err != nil {
			return nil, fmt.Errorf("failed to list panes: %w", err)
		}
	}

	// Load ignore list
	var ignoreList map[string]bool
	if opts.MockIgnoreList != nil {
		ignoreList = opts.MockIgnoreList
	} else {
		gitRoot := opts.RepoRoot
		if gitRoot == "" {
			gitRoot, _ = mail.FindGitRoot()
		}
		if gitRoot != "" {
			ignoreList, _ = mail.LoadIgnoreList(gitRoot)
		}
	}

	// Build recipients list, filtering ignored panes but always including current
	recipients := []RecipientInfo{}
	for _, pane := range panes {
		// Current pane is always shown (even if in ignore list)
		if pane == currentPane {
			recipients = append(recipients, RecipientInfo{
				Address:   pane,
				IsCurrent: true,
			})
		} else if !mail.IsIgnored(pane, ignoreList, currentSession) {
			// Only show non-current panes if they're not ignored
			recipients = append(recipients, RecipientInfo{
				Address:   pane,
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

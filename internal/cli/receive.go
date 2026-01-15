package cli

import (
	"fmt"
	"io"
	"time"

	"agentmail/internal/mail"
	"agentmail/internal/tmux"
)

// ReceiveOptions configures the Receive command behavior.
// Used for testing to mock tmux and file system operations.
type ReceiveOptions struct {
	SkipTmuxCheck bool     // Skip tmux environment check
	MockWindows   []string // Mock list of tmux windows
	MockReceiver  string   // Mock receiver window name
	RepoRoot      string   // Repository root (defaults to current directory)
	HookMode      bool     // Enable hook mode for Claude Code integration
}

// receiveError represents an error during receive with context.
type receiveError struct {
	msg      string
	exitCode int
}

// getReceiver returns the receiver window name.
func (opts *ReceiveOptions) getReceiver() (string, error) {
	if opts.MockReceiver != "" {
		return opts.MockReceiver, nil
	}
	return tmux.GetCurrentWindow()
}

// windowExists checks if a window exists in the tmux session.
func (opts *ReceiveOptions) windowExists(window string) (bool, error) {
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

// getRepoRoot returns the repository root path.
func (opts *ReceiveOptions) getRepoRoot() (string, error) {
	if opts.RepoRoot != "" {
		return opts.RepoRoot, nil
	}
	return mail.FindGitRoot()
}

// receiveCore contains the core receive logic, returning message or error info.
func (opts *ReceiveOptions) receiveCore() (*mail.Message, *receiveError) {
	// Get receiver identity
	receiver, err := opts.getReceiver()
	if err != nil {
		return nil, &receiveError{"failed to get current window: " + err.Error(), 1}
	}

	// Validate current window exists
	exists, err := opts.windowExists(receiver)
	if err != nil {
		return nil, &receiveError{"failed to check window: " + err.Error(), 1}
	}
	if !exists {
		return nil, &receiveError{fmt.Sprintf("current window '%s' not found in tmux session", receiver), 1}
	}

	// Get repository root
	repoRoot, err := opts.getRepoRoot()
	if err != nil {
		return nil, &receiveError{"not in a git repository: " + err.Error(), 1}
	}

	// Find unread messages
	unread, err := mail.FindUnread(repoRoot, receiver)
	if err != nil {
		return nil, &receiveError{"failed to read messages: " + err.Error(), 1}
	}

	// No messages case
	if len(unread) == 0 {
		return nil, nil
	}

	// Get oldest message and mark as read
	msg := unread[0]
	if err := mail.MarkAsRead(repoRoot, receiver, msg.ID); err != nil {
		return nil, &receiveError{"failed to mark message as read: " + err.Error(), 1}
	}

	// Update last_read_at timestamp (best-effort)
	if !opts.SkipTmuxCheck {
		_ = mail.UpdateLastReadAt(repoRoot, receiver, time.Now().UnixMilli())
	}

	return &msg, nil
}

// outputMessage writes the message to the given writer.
func outputMessage(w io.Writer, msg *mail.Message, prefix string) {
	if prefix != "" {
		fmt.Fprintln(w, prefix)
	}
	fmt.Fprintf(w, "From: %s\n", msg.From)
	fmt.Fprintf(w, "ID: %s\n", msg.ID)
	fmt.Fprintln(w)
	fmt.Fprint(w, msg.Message)
}

// Receive implements the agentmail receive command.
// Hook mode: silent on errors/no messages, output to stderr, exit 2 on message.
// Normal mode: verbose errors, output to stdout, exit 0 on success.
func Receive(stdout, stderr io.Writer, opts ReceiveOptions) int {
	// Validate tmux environment
	if !opts.SkipTmuxCheck && !tmux.InTmux() {
		if opts.HookMode {
			return 0
		}
		fmt.Fprintln(stderr, "error: agentmail must run inside a tmux session")
		return 2
	}

	// Execute core logic
	msg, recvErr := opts.receiveCore()

	// Handle errors
	if recvErr != nil {
		if opts.HookMode {
			return 0
		}
		fmt.Fprintf(stderr, "error: %s\n", recvErr.msg)
		return recvErr.exitCode
	}

	// Handle no messages
	if msg == nil {
		if opts.HookMode {
			return 0
		}
		fmt.Fprintln(stdout, "No unread messages")
		return 0
	}

	// Output message
	if opts.HookMode {
		outputMessage(stderr, msg, "You got new mail")
		return 2
	}

	outputMessage(stdout, msg, "")
	return 0
}

package cli

import (
	"fmt"
	"io"
	"os"

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
}

// Receive implements the agentmail receive command.
// T034: Implement Receive command structure
// T035: Add tmux validation (exit code 2 if not in tmux)
// T036: Add message retrieval and display formatting
// T037: Add "No unread messages" handling (exit code 0)
func Receive(stdout, stderr io.Writer, opts ReceiveOptions) int {
	// T035: Validate running inside tmux
	if !opts.SkipTmuxCheck {
		if !tmux.InTmux() {
			fmt.Fprintln(stderr, "error: agentmail must run inside a tmux session")
			return 2
		}
	}

	// Get receiver identity
	var receiver string
	if opts.MockReceiver != "" {
		receiver = opts.MockReceiver
	} else {
		var err error
		receiver, err = tmux.GetCurrentWindow()
		if err != nil {
			fmt.Fprintf(stderr, "error: failed to get current window: %v\n", err)
			return 1
		}
	}

	// Validate current window exists in tmux session (FR-006a, FR-006b)
	var receiverExists bool
	if opts.MockWindows != nil {
		for _, w := range opts.MockWindows {
			if w == receiver {
				receiverExists = true
				break
			}
		}
	} else {
		var err error
		receiverExists, err = tmux.WindowExists(receiver)
		if err != nil {
			fmt.Fprintf(stderr, "error: failed to check window: %v\n", err)
			return 1
		}
	}

	if !receiverExists {
		fmt.Fprintf(stderr, "error: current window '%s' not found in tmux session\n", receiver)
		return 1
	}

	// Determine repository root
	repoRoot := opts.RepoRoot
	if repoRoot == "" {
		var err error
		repoRoot, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(stderr, "error: failed to get current directory: %v\n", err)
			return 1
		}
	}

	// T036: Find unread messages for receiver
	unread, err := mail.FindUnread(repoRoot, receiver)
	if err != nil {
		fmt.Fprintf(stderr, "error: failed to read messages: %v\n", err)
		return 1
	}

	// T037: Handle no unread messages
	if len(unread) == 0 {
		fmt.Fprintln(stdout, "No unread messages")
		return 0
	}

	// Get oldest unread message (FIFO - first in list)
	msg := unread[0]

	// Mark as read
	if err := mail.MarkAsRead(repoRoot, receiver, msg.ID); err != nil {
		fmt.Fprintf(stderr, "error: failed to mark message as read: %v\n", err)
		return 1
	}

	// Display message in format:
	// From: <sender>
	// ID: <id>
	//
	// <message>
	fmt.Fprintf(stdout, "From: %s\n", msg.From)
	fmt.Fprintf(stdout, "ID: %s\n", msg.ID)
	fmt.Fprintln(stdout)
	fmt.Fprint(stdout, msg.Message)

	return 0
}

package cli

import (
	"fmt"
	"io"

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

// Receive implements the agentmail receive command.
// T034: Implement Receive command structure
// T035: Add tmux validation (exit code 2 if not in tmux)
// T036: Add message retrieval and display formatting
// T037: Add "No unread messages" handling (exit code 0)
//
// Hook mode behavior (FR-001 through FR-005):
// - FR-001a/b/c: Write notification to STDERR, exit 2, mark as read when messages exist
// - FR-002: Exit 0 with no output when no messages
// - FR-003: Exit 0 with no output when not in tmux
// - FR-004a/b/c: Exit 0 with no output on any error
// - FR-005: All output to STDERR in hook mode
func Receive(stdout, stderr io.Writer, opts ReceiveOptions) int {
	// T035: Validate running inside tmux
	if !opts.SkipTmuxCheck {
		if !tmux.InTmux() {
			// FR-003: Hook mode exits silently when not in tmux
			if opts.HookMode {
				return 0
			}
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
			// FR-004a: Hook mode exits silently on errors
			if opts.HookMode {
				return 0
			}
			fmt.Fprintf(stderr, "error: failed to get current window: %v\n", err)
			return 1
		}
	}

	// Validate current window exists in tmux session
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
			// FR-004a: Hook mode exits silently on errors
			if opts.HookMode {
				return 0
			}
			fmt.Fprintf(stderr, "error: failed to check window: %v\n", err)
			return 1
		}
	}

	if !receiverExists {
		// FR-004a: Hook mode exits silently on errors
		if opts.HookMode {
			return 0
		}
		fmt.Fprintf(stderr, "error: current window '%s' not found in tmux session\n", receiver)
		return 1
	}

	// Determine repository root (find git root, not current directory)
	repoRoot := opts.RepoRoot
	if repoRoot == "" {
		var err error
		repoRoot, err = mail.FindGitRoot()
		if err != nil {
			// FR-004a: Hook mode exits silently on errors
			if opts.HookMode {
				return 0
			}
			fmt.Fprintf(stderr, "error: not in a git repository: %v\n", err)
			return 1
		}
	}

	// T036: Find unread messages for receiver
	unread, err := mail.FindUnread(repoRoot, receiver)
	if err != nil {
		// FR-004a/b/c: Hook mode exits silently on file/lock/corruption errors
		if opts.HookMode {
			return 0
		}
		fmt.Fprintf(stderr, "error: failed to read messages: %v\n", err)
		return 1
	}

	// T037: Handle no unread messages
	if len(unread) == 0 {
		// FR-002: Hook mode exits silently with no messages
		if opts.HookMode {
			return 0
		}
		fmt.Fprintln(stdout, "No unread messages")
		return 0
	}

	// Get oldest unread message (FIFO - first in list)
	msg := unread[0]

	// FR-001c: Mark as read
	if err := mail.MarkAsRead(repoRoot, receiver, msg.ID); err != nil {
		// FR-004a: Hook mode exits silently on errors
		if opts.HookMode {
			return 0
		}
		fmt.Fprintf(stderr, "error: failed to mark message as read: %v\n", err)
		return 1
	}

	// FR-005: Hook mode writes all output to STDERR
	// FR-001a: Hook mode prefixes with "You got new mail\n"
	if opts.HookMode {
		fmt.Fprintln(stderr, "You got new mail")
		fmt.Fprintf(stderr, "From: %s\n", msg.From)
		fmt.Fprintf(stderr, "ID: %s\n", msg.ID)
		fmt.Fprintln(stderr)
		fmt.Fprint(stderr, msg.Message)
		// FR-001b: Hook mode exits with code 2 when messages exist
		return 2
	}

	// Normal mode: Display message to stdout
	// Format:
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

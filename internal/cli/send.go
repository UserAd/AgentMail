package cli

import (
	"fmt"
	"io"
	"os"

	"agentmail/internal/mail"
	"agentmail/internal/tmux"
)

// SendOptions configures the Send command behavior.
// Used for testing to mock tmux and file system operations.
type SendOptions struct {
	SkipTmuxCheck  bool            // Skip tmux environment check
	MockWindows    []string        // Mock list of tmux windows
	MockSender     string          // Mock sender window name
	RepoRoot       string          // Repository root (defaults to current directory)
	MockIgnoreList map[string]bool // Mock ignore list (nil = load from file)
	MockGitRoot    string          // Mock git root (for testing)
}

// Send implements the agentmail send command.
// T020: Implement Send command structure
// T021: Add tmux validation (exit code 2 if not in tmux)
// T022: Add recipient validation (check WindowExists)
// T023: Add message storage and ID output
func Send(args []string, stdout, stderr io.Writer, opts SendOptions) int {
	// T021: Validate running inside tmux
	if !opts.SkipTmuxCheck {
		if !tmux.InTmux() {
			fmt.Fprintln(stderr, "error: agentmail must run inside a tmux session")
			return 2
		}
	}

	// Validate required arguments
	if len(args) < 2 {
		if len(args) == 0 {
			fmt.Fprintln(stderr, "error: missing required arguments: recipient message")
		} else {
			fmt.Fprintln(stderr, "error: missing required argument: message")
		}
		return 1
	}

	recipient := args[0]
	message := args[1]

	// Get sender identity
	var sender string
	if opts.MockSender != "" {
		sender = opts.MockSender
	} else {
		var err error
		sender, err = tmux.GetCurrentWindow()
		if err != nil {
			fmt.Fprintf(stderr, "error: failed to get current window: %v\n", err)
			return 1
		}
	}

	// T022: Validate recipient exists
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
			fmt.Fprintf(stderr, "error: failed to check recipient: %v\n", err)
			return 1
		}
	}

	if !recipientExists {
		fmt.Fprintln(stderr, "error: recipient not found")
		return 1
	}

	// T029: Check if recipient is the sender (self-send not allowed)
	if recipient == sender {
		fmt.Fprintln(stderr, "error: recipient not found")
		return 1
	}

	// T029: Load and check ignore list
	var ignoreList map[string]bool
	if opts.MockIgnoreList != nil {
		ignoreList = opts.MockIgnoreList
	} else {
		// Load from .agentmailignore file
		var gitRoot string
		if opts.MockGitRoot != "" {
			gitRoot = opts.MockGitRoot
		} else {
			gitRoot, _ = mail.FindGitRoot()
			// Errors from FindGitRoot mean not in a git repo - proceed without ignore list
		}
		if gitRoot != "" {
			ignoreList, _ = mail.LoadIgnoreList(gitRoot)
			// Errors from LoadIgnoreList are treated as no ignore file
		}
	}

	// T030: Check if recipient is in ignore list
	if ignoreList != nil && ignoreList[recipient] {
		fmt.Fprintln(stderr, "error: recipient not found")
		return 1
	}

	// Generate message ID
	id, err := mail.GenerateID()
	if err != nil {
		fmt.Fprintf(stderr, "error: failed to generate message ID: %v\n", err)
		return 1
	}

	// Determine repository root
	repoRoot := opts.RepoRoot
	if repoRoot == "" {
		repoRoot, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(stderr, "error: failed to get current directory: %v\n", err)
			return 1
		}
	}

	// T023: Store message
	msg := mail.Message{
		ID:       id,
		From:     sender,
		To:       recipient,
		Message:  message,
		ReadFlag: false,
	}

	if err := mail.Append(repoRoot, msg); err != nil {
		fmt.Fprintf(stderr, "error: failed to write message: %v\n", err)
		return 1
	}

	// Output message confirmation
	fmt.Fprintf(stdout, "Message #%s sent\n", id)
	return 0
}

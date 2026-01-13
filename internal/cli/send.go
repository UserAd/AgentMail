package cli

import (
	"fmt"
	"io"
	"strings"

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
	StdinContent   string          // Mock stdin content (empty = no stdin)
	StdinIsPipe    bool            // Mock whether stdin is a pipe
}

// Send implements the agentmail send command.
// T020: Implement Send command structure
// T021: Add tmux validation (exit code 2 if not in tmux)
// T022: Add recipient validation (check WindowExists)
// T023: Add message storage and ID output
// T045: Accept io.Reader for stdin
func Send(args []string, stdin io.Reader, stdout, stderr io.Writer, opts SendOptions) int {
	// T021: Validate running inside tmux
	if !opts.SkipTmuxCheck {
		if !tmux.InTmux() {
			fmt.Fprintln(stderr, "error: agentmail must run inside a tmux session")
			return 2
		}
	}

	// Validate recipient argument is provided
	if len(args) == 0 {
		fmt.Fprintln(stderr, "error: missing required arguments: recipient message")
		return 1
	}

	recipient := args[0]

	// T046-T048: Get message from stdin or argument
	var message string

	// Check if stdin is a pipe
	isStdinPipe := opts.StdinIsPipe
	if !opts.SkipTmuxCheck { // If not mocking, use real detection
		isStdinPipe = IsStdinPipe()
	}

	if isStdinPipe {
		// T047: Read from stdin (use mock or real)
		var stdinContent []byte
		if opts.StdinContent != "" {
			stdinContent = []byte(opts.StdinContent)
		} else if stdin != nil {
			var err error
			stdinContent, err = io.ReadAll(stdin)
			if err != nil {
				fmt.Fprintf(stderr, "error: failed to read stdin: %v\n", err)
				return 1
			}
		}
		if len(stdinContent) > 0 {
			// Trim trailing newline only (preserve internal newlines for multi-line messages)
			message = strings.TrimSuffix(string(stdinContent), "\n")
		}
	}

	// T048: Fall back to argument if no stdin content
	if message == "" && len(args) >= 2 {
		message = args[1]
	}

	// Error if no message provided
	if message == "" {
		fmt.Fprintln(stderr, "error: no message provided")
		fmt.Fprintln(stderr, "usage: agentmail send <recipient> <message>")
		return 1
	}

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

	// Determine repository root (find git root, not current directory)
	repoRoot := opts.RepoRoot
	if repoRoot == "" {
		repoRoot, err = mail.FindGitRoot()
		if err != nil {
			fmt.Fprintf(stderr, "error: not in a git repository: %v\n", err)
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

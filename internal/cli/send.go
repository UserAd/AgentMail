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

// extractMessage reads message from stdin or args.
// Returns the message content and any error encountered.
func (opts *SendOptions) extractMessage(args []string, stdin io.Reader) (string, error) {
	var message string

	// Check if stdin is a pipe
	isStdinPipe := opts.StdinIsPipe
	if !opts.SkipTmuxCheck {
		isStdinPipe = IsStdinPipe()
	}

	if isStdinPipe {
		var stdinContent []byte
		if opts.StdinContent != "" {
			stdinContent = []byte(opts.StdinContent)
		} else if stdin != nil {
			var err error
			stdinContent, err = io.ReadAll(stdin)
			if err != nil {
				return "", fmt.Errorf("failed to read stdin: %w", err)
			}
		}
		if len(stdinContent) > 0 {
			message = strings.TrimSuffix(string(stdinContent), "\n")
		}
	}

	// Fall back to argument if no stdin content
	if message == "" && len(args) >= 2 {
		message = args[1]
	}

	return message, nil
}

// getSender returns the sender window name.
func (opts *SendOptions) getSender() (string, error) {
	if opts.MockSender != "" {
		return opts.MockSender, nil
	}
	return tmux.GetCurrentWindow()
}

// windowExists checks if a window exists in the tmux session.
func (opts *SendOptions) windowExists(window string) (bool, error) {
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

// loadIgnoreList loads the ignore list from file or returns mock.
func (opts *SendOptions) loadIgnoreList() map[string]bool {
	if opts.MockIgnoreList != nil {
		return opts.MockIgnoreList
	}

	gitRoot := opts.MockGitRoot
	if gitRoot == "" {
		gitRoot, _ = mail.FindGitRoot()
	}
	if gitRoot == "" {
		return nil
	}

	ignoreList, _ := mail.LoadIgnoreList(gitRoot)
	return ignoreList
}

// getRepoRoot returns the repository root path.
func (opts *SendOptions) getRepoRoot() (string, error) {
	if opts.RepoRoot != "" {
		return opts.RepoRoot, nil
	}
	return mail.FindGitRoot()
}

// Send implements the agentmail send command.
func Send(args []string, stdin io.Reader, stdout, stderr io.Writer, opts SendOptions) int {
	// Validate running inside tmux
	if !opts.SkipTmuxCheck && !tmux.InTmux() {
		fmt.Fprintln(stderr, "error: agentmail must run inside a tmux session")
		return 2
	}

	// Validate recipient argument is provided
	if len(args) == 0 {
		fmt.Fprintln(stderr, "error: missing required arguments: recipient message")
		return 1
	}

	recipient := args[0]

	// Extract message from stdin or argument
	message, err := opts.extractMessage(args, stdin)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	if message == "" {
		fmt.Fprintln(stderr, "error: no message provided")
		fmt.Fprintln(stderr, "usage: agentmail send <recipient> <message>")
		return 1
	}

	// Get sender identity
	sender, err := opts.getSender()
	if err != nil {
		fmt.Fprintf(stderr, "error: failed to get current window: %v\n", err)
		return 1
	}

	// Validate recipient exists
	recipientExists, err := opts.windowExists(recipient)
	if err != nil {
		fmt.Fprintf(stderr, "error: failed to check recipient: %v\n", err)
		return 1
	}
	if !recipientExists || recipient == sender {
		fmt.Fprintln(stderr, "error: recipient not found")
		return 1
	}

	// Check ignore list
	if ignoreList := opts.loadIgnoreList(); ignoreList != nil && ignoreList[recipient] {
		fmt.Fprintln(stderr, "error: recipient not found")
		return 1
	}

	// Get repository root
	repoRoot, err := opts.getRepoRoot()
	if err != nil {
		fmt.Fprintf(stderr, "error: not in a git repository: %v\n", err)
		return 1
	}

	// Generate message ID and store
	id, err := mail.GenerateID()
	if err != nil {
		fmt.Fprintf(stderr, "error: failed to generate message ID: %v\n", err)
		return 1
	}

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

	fmt.Fprintf(stdout, "Message #%s sent\n", id)
	return 0
}

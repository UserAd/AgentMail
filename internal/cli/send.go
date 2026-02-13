package cli

import (
	"fmt"
	"io"
	"strings"

	"agentmail/internal/mail"
	"agentmail/internal/tmux"
)

// formatPaneList formats a list of pane addresses for error messages.
func formatPaneList(panes []string) string {
	if len(panes) == 0 {
		return ""
	}
	if len(panes) == 1 {
		return panes[0]
	}
	if len(panes) == 2 {
		return panes[0] + ", " + panes[1]
	}
	// For 3+ panes, use commas with "or" before the last one
	parts := make([]string, len(panes))
	for i := 0; i < len(panes)-1; i++ {
		parts[i] = panes[i]
	}
	return strings.Join(parts[:len(panes)-1], ", ") + " or " + panes[len(panes)-1]
}

// SendOptions configures the Send command behavior.
// Used for testing to mock tmux and file system operations.
type SendOptions struct {
	SkipTmuxCheck   bool            // Skip tmux environment check
	MockWindows     []string        // Mock list of tmux windows (deprecated, use MockPanes)
	MockPanes       []string        // Mock list of full pane addresses
	MockSender      string          // Mock sender window name (deprecated, use MockPaneAddress)
	MockPaneAddress string          // Mock sender pane address
	MockSession     string          // Mock current session for address resolution
	RepoRoot        string          // Repository root (defaults to current directory)
	MockIgnoreList  map[string]bool // Mock ignore list (nil = load from file)
	MockGitRoot     string          // Mock git root (for testing)
	StdinContent    string          // Mock stdin content (empty = no stdin)
	StdinIsPipe     bool            // Mock whether stdin is a pipe
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

	// Get sender identity (pane address)
	var sender string
	if opts.MockPaneAddress != "" {
		sender = opts.MockPaneAddress
	} else if opts.MockSender != "" {
		// Backward compatibility for old tests
		sender = opts.MockSender
	} else {
		var err error
		sender, err = tmux.GetCurrentPaneAddress()
		if err != nil {
			fmt.Fprintf(stderr, "error: failed to get current pane: %v\n", err)
			return 1
		}
	}

	// Get current session for address resolution
	var currentSession string
	if opts.MockSession != "" {
		currentSession = opts.MockSession
	} else if opts.SkipTmuxCheck {
		// In tests without mock session, extract from sender
		addr, err := tmux.ParseAddress(sender, "")
		if err == nil {
			currentSession = addr.Session
		}
	} else {
		var err error
		currentSession, err = tmux.GetCurrentSession()
		if err != nil {
			fmt.Fprintf(stderr, "error: failed to get current session: %v\n", err)
			return 1
		}
	}

	// Resolve recipient address
	addr, err := tmux.ParseAddress(recipient, currentSession)
	if err != nil {
		fmt.Fprintln(stderr, "error: recipient not found")
		return 1
	}

	var resolvedRecipient string
	if addr.Pane == -1 {
		// Short form - need to resolve window name to pane address
		var panes []string
		if opts.MockPanes != nil {
			panes = opts.MockPanes
		} else if opts.MockWindows != nil {
			// Backward compatibility for old tests
			panes = opts.MockWindows
		} else {
			panes, err = tmux.ListPanes()
			if err != nil {
				fmt.Fprintf(stderr, "error: failed to list panes: %v\n", err)
				return 1
			}
		}

		// Find matching panes
		var matches []string
		for _, pane := range panes {
			paneAddr, err := tmux.ParseAddress(pane, currentSession)
			if err == nil && paneAddr.Window == addr.Window {
				matches = append(matches, pane)
			}
		}

		if len(matches) == 0 {
			fmt.Fprintln(stderr, "error: recipient not found")
			return 1
		} else if len(matches) > 1 {
			fmt.Fprintf(stderr, "Ambiguous recipient: window '%s' has %d panes. Use %s\n",
				addr.Window, len(matches), formatPaneList(matches))
			return 1
		}
		resolvedRecipient = matches[0]
	} else {
		// Full or medium form - validate pane exists
		resolvedRecipient = tmux.FormatAddress(addr)
		var paneExists bool
		if opts.MockPanes != nil {
			for _, p := range opts.MockPanes {
				if p == resolvedRecipient {
					paneExists = true
					break
				}
			}
		} else if opts.MockWindows != nil {
			// Backward compatibility
			paneExists = true
		} else {
			paneExists, err = tmux.PaneExists(resolvedRecipient)
			if err != nil {
				fmt.Fprintf(stderr, "error: failed to check recipient: %v\n", err)
				return 1
			}
		}

		if !paneExists {
			fmt.Fprintln(stderr, "error: recipient not found")
			return 1
		}
	}

	// Check if recipient is the sender (self-send not allowed)
	if resolvedRecipient == sender {
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

	// Check if recipient is in ignore list
	if mail.IsIgnored(resolvedRecipient, ignoreList, currentSession) {
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

	// Store message
	msg := mail.Message{
		ID:       id,
		From:     sender,
		To:       resolvedRecipient,
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

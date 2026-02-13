package cli

import (
	"fmt"
	"io"

	"agentmail/internal/mail"
	"agentmail/internal/tmux"
)

// RecipientsOptions configures the Recipients command behavior.
type RecipientsOptions struct {
	SkipTmuxCheck   bool            // Skip tmux environment check
	MockPanes       []string        // Mock list of tmux panes
	MockPaneAddress string          // Mock current pane address
	MockSession     string          // Mock session name
	MockIgnoreList  map[string]bool // Mock ignore list (nil = load from file)
	MockGitRoot     string          // Mock git root (for testing)
}

// Recipients implements the agentmail recipients command.
// It lists all tmux panes with the current pane marked "[you]".
func Recipients(stdout, stderr io.Writer, opts RecipientsOptions) int {
	// Validate running inside tmux
	if !opts.SkipTmuxCheck {
		if !tmux.InTmux() {
			fmt.Fprintln(stderr, "error: not running inside a tmux session")
			return 2
		}
	}

	// Get list of panes
	var panes []string
	if opts.MockPanes != nil {
		panes = opts.MockPanes
	} else {
		var err error
		panes, err = tmux.ListPanes()
		if err != nil {
			fmt.Fprintf(stderr, "error: failed to list panes: %v\n", err)
			return 1
		}
	}

	// Get current pane address
	var currentPane string
	if opts.MockPaneAddress != "" {
		currentPane = opts.MockPaneAddress
	} else {
		var err error
		currentPane, err = tmux.GetCurrentPaneAddress()
		if err != nil {
			fmt.Fprintf(stderr, "error: failed to get current pane address: %v\n", err)
			return 1
		}
	}

	// Get current session for ignore matching
	var currentSession string
	if opts.MockSession != "" {
		currentSession = opts.MockSession
	} else {
		// Try to extract session from current pane address
		addr, err := tmux.ParseAddress(currentPane, "")
		if err == nil {
			currentSession = addr.Session
		} else {
			var sessionErr error
			currentSession, sessionErr = tmux.GetCurrentSession()
			if sessionErr != nil {
				fmt.Fprintf(stderr, "error: failed to get current session: %v\n", sessionErr)
				return 1
			}
		}
	}

	// Load ignore list
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
			// Errors from LoadIgnoreList are treated as no ignore file (per FR-013)
		}
	}

	// Output panes with current marked, filtering ignored panes
	for _, pane := range panes {
		// Current pane is always shown (per FR-004), even if in ignore list
		if pane == currentPane {
			fmt.Fprintf(stdout, "%s [you]\n", pane)
		} else if !mail.IsIgnored(pane, ignoreList, currentSession) {
			// Only show non-current panes if they're not ignored
			fmt.Fprintf(stdout, "%s\n", pane)
		}
	}

	return 0
}

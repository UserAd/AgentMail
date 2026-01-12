package cli

import (
	"fmt"
	"io"

	"agentmail/internal/mail"
	"agentmail/internal/tmux"
)

// RecipientsOptions configures the Recipients command behavior.
type RecipientsOptions struct {
	SkipTmuxCheck  bool            // Skip tmux environment check
	MockWindows    []string        // Mock list of tmux windows
	MockCurrent    string          // Mock current window name
	MockIgnoreList map[string]bool // Mock ignore list (nil = load from file)
	MockGitRoot    string          // Mock git root (for testing)
}

// Recipients implements the agentmail recipients command.
// It lists all tmux windows with the current window marked "[you]".
func Recipients(stdout, stderr io.Writer, opts RecipientsOptions) int {
	// Validate running inside tmux
	if !opts.SkipTmuxCheck {
		if !tmux.InTmux() {
			fmt.Fprintln(stderr, "error: not running inside a tmux session")
			return 2
		}
	}

	// Get list of windows
	var windows []string
	if opts.MockWindows != nil {
		windows = opts.MockWindows
	} else {
		var err error
		windows, err = tmux.ListWindows()
		if err != nil {
			fmt.Fprintf(stderr, "error: failed to list windows: %v\n", err)
			return 1
		}
	}

	// Get current window
	// In mock mode (MockWindows is set), use MockCurrent even if empty
	var currentWindow string
	if opts.MockWindows != nil {
		currentWindow = opts.MockCurrent
	} else {
		var err error
		currentWindow, err = tmux.GetCurrentWindow()
		if err != nil {
			fmt.Fprintf(stderr, "error: failed to get current window: %v\n", err)
			return 1
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

	// Output windows with current marked, filtering ignored windows
	for _, window := range windows {
		// Current window is always shown (per FR-004), even if in ignore list
		if window == currentWindow {
			fmt.Fprintf(stdout, "%s [you]\n", window)
		} else if ignoreList == nil || !ignoreList[window] {
			// Only show non-current windows if they're not in the ignore list
			fmt.Fprintf(stdout, "%s\n", window)
		}
	}

	return 0
}

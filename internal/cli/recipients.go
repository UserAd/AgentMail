package cli

import (
	"fmt"
	"io"

	"agentmail/internal/tmux"
)

// RecipientsOptions configures the Recipients command behavior.
type RecipientsOptions struct {
	SkipTmuxCheck bool     // Skip tmux environment check
	MockWindows   []string // Mock list of tmux windows
	MockCurrent   string   // Mock current window name
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
	var currentWindow string
	if opts.MockCurrent != "" {
		currentWindow = opts.MockCurrent
	} else {
		var err error
		currentWindow, err = tmux.GetCurrentWindow()
		if err != nil {
			fmt.Fprintf(stderr, "error: failed to get current window: %v\n", err)
			return 1
		}
	}

	// Output windows with current marked
	for _, window := range windows {
		if window == currentWindow {
			fmt.Fprintf(stdout, "%s [you]\n", window)
		} else {
			fmt.Fprintf(stdout, "%s\n", window)
		}
	}

	return 0
}

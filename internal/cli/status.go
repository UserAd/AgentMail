package cli

import (
	"fmt"
	"io"

	"agentmail/internal/mail"
	"agentmail/internal/tmux"
)

// StatusOptions configures the Status command behavior.
// Used for testing to mock tmux and file system operations.
type StatusOptions struct {
	SkipTmuxCheck bool   // Skip tmux environment check
	MockWindow    string // Mock current window name
	RepoRoot      string // Repository root (defaults to finding git root)
}

// ValidateStatus checks if the provided status is a valid status value.
// T039: Implement ValidateStatus() for ready/work/offline enum
// Valid values: ready, work, offline (case sensitive)
func ValidateStatus(status string) bool {
	switch status {
	case mail.StatusReady, mail.StatusWork, mail.StatusOffline:
		return true
	default:
		return false
	}
}

// Status implements the agentmail status command.
// T038: Implement Status() command handler
//
// Contract from cli.md:
// agentmail status <STATUS>
//
// Arguments:
// - STATUS: One of `ready`, `work`, `offline`
//
// Exit Codes:
// - 0: Status updated (or no-op outside tmux)
// - 1: Invalid status name
//
// Stdout: Empty (silent on success)
//
// Stderr (invalid status):
// Invalid status: foo. Valid: ready, work, offline
//
// Behavior:
// 1. Check if running inside tmux ($TMUX env var)
// 2. If not in tmux: exit 0 silently (no-op for non-tmux environments)
// 3. Get current tmux window name
// 4. Parse status argument - if not ready/work/offline: print error to stderr, exit 1
// 5. Update .git/mail-recipients.jsonl
// 6. If transitioning to `work` or `offline`: reset `notified` to false
// 7. Exit 0
func Status(args []string, stdout, stderr io.Writer, opts StatusOptions) int {
	// T040: Handle non-tmux case (silent exit 0)
	if !opts.SkipTmuxCheck {
		if !tmux.InTmux() {
			// Exit 0 silently (no-op for non-tmux environments)
			return 0
		}
	}

	// Validate status argument is provided
	if len(args) == 0 {
		fmt.Fprintln(stderr, "error: missing required argument: status")
		fmt.Fprintln(stderr, "usage: agentmail status <ready|work|offline>")
		return 1
	}

	status := args[0]

	// T039: Validate status value
	if !ValidateStatus(status) {
		fmt.Fprintf(stderr, "Invalid status: %s. Valid: ready, work, offline\n", status)
		return 1
	}

	// Get current window name
	var window string
	if opts.MockWindow != "" {
		window = opts.MockWindow
	} else {
		var err error
		window, err = tmux.GetCurrentWindow()
		if err != nil {
			fmt.Fprintf(stderr, "error: failed to get current window: %v\n", err)
			return 1
		}
	}

	// Determine repository root
	repoRoot := opts.RepoRoot
	if repoRoot == "" {
		var err error
		repoRoot, err = mail.FindGitRoot()
		if err != nil {
			fmt.Fprintf(stderr, "error: not in a git repository: %v\n", err)
			return 1
		}
	}

	// T041: Implement notified flag reset on work/offline transition
	// Reset notified flag only when transitioning to work or offline
	resetNotified := (status == mail.StatusWork || status == mail.StatusOffline)

	// Update recipient state using existing infrastructure
	if err := mail.UpdateRecipientState(repoRoot, window, status, resetNotified); err != nil {
		fmt.Fprintf(stderr, "error: failed to update status: %v\n", err)
		return 1
	}

	// Silent success - no output
	return 0
}

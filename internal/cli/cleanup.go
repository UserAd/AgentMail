package cli

import (
	"fmt"
	"io"
	"time"

	"agentmail/internal/mail"
	"agentmail/internal/tmux"
)

// CleanupOptions configures the Cleanup command behavior
type CleanupOptions struct {
	StaleHours     int  // Hours threshold for stale recipients (default: 48)
	DeliveredHours int  // Hours threshold for delivered messages (default: 2)
	DryRun         bool // If true, report what would be cleaned without deleting

	// Testing options
	RepoRoot      string   // Repository root (defaults to "." if empty)
	SkipTmuxCheck bool     // Skip real tmux check (for testing)
	MockInTmux    bool     // Mocked value for InTmux() when SkipTmuxCheck is true
	MockWindows   []string // Mock list of tmux windows (for testing, nil means use real tmux)
}

// CleanupResult holds the counts from a cleanup operation
type CleanupResult struct {
	RecipientsRemoved int // Total recipients removed
	OfflineRemoved    int // Recipients removed because window doesn't exist
	StaleRemoved      int // Recipients removed because updated_at expired
	MessagesRemoved   int // Messages removed (read + old)
	MailboxesRemoved  int // Empty mailbox files removed
	FilesSkipped      int // Files skipped due to lock contention
}

// Cleanup removes stale data from the AgentMail system.
// It removes offline recipients, stale recipients, old delivered messages, and empty mailboxes.
//
// FR-001: Compare each recipient in recipients.jsonl against current tmux window names
// FR-002: Remove recipients whose names don't match any current tmux window
// FR-020: Skip offline recipient check when not in tmux
// FR-021: Output warning when offline check is skipped due to non-tmux environment
// FR-022: Report zero recipients removed when recipients.jsonl doesn't exist
func Cleanup(stdout, stderr io.Writer, opts CleanupOptions) int {
	result := CleanupResult{}

	// Determine repository root
	repoRoot := opts.RepoRoot
	if repoRoot == "" {
		repoRoot = "."
	}

	// Determine if we're in tmux
	var inTmux bool
	var isMocking bool

	if opts.SkipTmuxCheck {
		// Testing mode - use mocked values
		inTmux = opts.MockInTmux
		isMocking = true
	} else {
		// Production mode - use real tmux check
		inTmux = tmux.InTmux()
		isMocking = false
	}

	// Phase 1: Clean offline recipients (US1)
	if inTmux {
		// Get list of valid tmux windows
		var windows []string
		if isMocking {
			windows = opts.MockWindows
		} else {
			var err error
			windows, err = tmux.ListWindows()
			if err != nil {
				fmt.Fprintf(stderr, "Warning: failed to list tmux windows: %v\n", err)
				// Continue without offline cleanup
				windows = nil
			}
		}

		// Clean offline recipients if we have a window list
		if windows != nil && !opts.DryRun {
			removed, err := mail.CleanOfflineRecipients(repoRoot, windows)
			if err != nil {
				fmt.Fprintf(stderr, "Error cleaning offline recipients: %v\n", err)
				return 1
			}
			result.OfflineRemoved = removed
			result.RecipientsRemoved += removed
		}
	} else {
		// FR-020 & FR-021: Not in tmux - skip offline check with warning
		fmt.Fprintln(stderr, "Warning: not running in tmux session, skipping offline recipient check")
	}

	// Phase 2: Clean stale recipients (US2)
	// Remove recipients whose updated_at is older than StaleHours threshold
	if !opts.DryRun {
		staleThreshold := time.Duration(opts.StaleHours) * time.Hour
		staleRemoved, err := mail.CleanStaleStates(repoRoot, staleThreshold)
		if err != nil {
			fmt.Fprintf(stderr, "Error cleaning stale recipients: %v\n", err)
			return 1
		}
		result.StaleRemoved = staleRemoved
		result.RecipientsRemoved += staleRemoved
	}

	// Phase 3: Clean old delivered messages (US3)
	// Remove read messages older than DeliveredHours threshold
	if !opts.DryRun {
		deliveredThreshold := time.Duration(opts.DeliveredHours) * time.Hour

		// List all mailbox files
		recipients, err := mail.ListMailboxRecipients(repoRoot)
		if err != nil {
			fmt.Fprintf(stderr, "Error listing mailboxes: %v\n", err)
			return 1
		}

		// Clean each mailbox
		for _, recipient := range recipients {
			removed, err := mail.CleanOldMessages(repoRoot, recipient, deliveredThreshold)
			if err != nil {
				// Check if it's a lock failure - skip file and continue
				if err == mail.ErrFileLocked {
					result.FilesSkipped++
					continue
				}
				fmt.Fprintf(stderr, "Error cleaning mailbox %s: %v\n", recipient, err)
				return 1
			}
			result.MessagesRemoved += removed
		}
	}

	// Suppress unused variable warning - result will be used for summary output
	_ = result

	// TODO: Phase 4 - Remove empty mailboxes (US4)
	// TODO: Phase 5 - Output summary (FR-014)

	return 0
}

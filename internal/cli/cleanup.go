package cli

import (
	"io"
)

// CleanupOptions configures the Cleanup command behavior
type CleanupOptions struct {
	StaleHours     int  // Hours threshold for stale recipients (default: 48)
	DeliveredHours int  // Hours threshold for delivered messages (default: 2)
	DryRun         bool // If true, report what would be cleaned without deleting
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
func Cleanup(stdout, stderr io.Writer, opts CleanupOptions) int {
	// Stub implementation - returns 0 for success
	return 0
}

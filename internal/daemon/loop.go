// Package daemon provides functionality for the mailman daemon process.
// This file contains the notification loop implementation.
package daemon

import (
	"fmt"
	"io"
	"sync"
	"time"

	"agentmail/internal/mail"
	"agentmail/internal/tmux"
)

// DefaultStaleThreshold is the default threshold for cleaning stale states.
const DefaultStaleThreshold = time.Hour

// StatelessNotifyInterval is the interval between notifications for stateless agents (T001).
const StatelessNotifyInterval = 60 * time.Second

// StatelessTracker tracks notification timestamps for stateless agents (T002).
// It uses in-memory storage that resets on daemon restart.
type StatelessTracker struct {
	mu             sync.Mutex           // Protects concurrent access
	lastNotified   map[string]time.Time // Window name → last notification time
	notifyInterval time.Duration        // Minimum interval between notifications
}

// NewStatelessTracker creates a new tracker with the specified notification interval (T010).
func NewStatelessTracker(interval time.Duration) *StatelessTracker {
	return &StatelessTracker{
		lastNotified:   make(map[string]time.Time),
		notifyInterval: interval,
	}
}

// ShouldNotify returns true if the window is eligible for notification (T011).
// Returns true if: (a) window not in tracker, or (b) interval elapsed since last notification.
func (t *StatelessTracker) ShouldNotify(window string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	lastTime, exists := t.lastNotified[window]
	if !exists {
		return true
	}

	return time.Since(lastTime) >= t.notifyInterval
}

// MarkNotified records that a notification was sent to the window (T012).
func (t *StatelessTracker) MarkNotified(window string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.lastNotified[window] = time.Now()
}

// Cleanup removes entries for windows that are no longer active (T013).
func (t *StatelessTracker) Cleanup(activeWindows []string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Build a set of active windows for O(1) lookup
	activeSet := make(map[string]struct{}, len(activeWindows))
	for _, w := range activeWindows {
		activeSet[w] = struct{}{}
	}

	// Remove entries not in the active set
	for window := range t.lastNotified {
		if _, exists := activeSet[window]; !exists {
			delete(t.lastNotified, window)
		}
	}
}

// LoopOptions configures the notification check.
type LoopOptions struct {
	RepoRoot         string            // Repository root path
	StopChan         chan struct{}     // Channel to stop the watcher
	SkipTmuxCheck    bool              // Skip tmux check (for testing)
	StatelessTracker *StatelessTracker // Tracker for stateless agents (T003)
	Logger           io.Writer         // Logger for foreground mode (nil = no logging)
}

// log writes a formatted message to the logger if configured.
func (opts *LoopOptions) log(format string, args ...interface{}) {
	if opts.Logger != nil {
		fmt.Fprintf(opts.Logger, "[mailman] "+format+"\n", args...)
	}
}

// NotifyFunc is the function signature for notifying an agent.
type NotifyFunc func(window string) error

// NotifyAgent sends a notification to an agent's tmux window.
// Notification protocol:
// 1. tmux send-keys -t <window> "Check your agentmail"
// 2. time.Sleep(1 * time.Second)
// 3. tmux send-keys -t <window> Enter
func NotifyAgent(window string) error {
	// Send the notification message
	if err := tmux.SendKeys(window, "Check your agentmail"); err != nil {
		return err
	}

	// Wait 1 second before sending Enter
	time.Sleep(1 * time.Second)

	// Send Enter to execute the command
	if err := tmux.SendEnter(window); err != nil {
		return err
	}

	return nil
}

// CheckAndNotify performs a single notification cycle.
// It reads recipient states, checks for ready agents with unread messages,
// and sends notifications to those who haven't been notified yet.
// If SkipTmuxCheck is true, notifications are skipped but the notified flag is still updated.
func CheckAndNotify(opts LoopOptions) error {
	if opts.SkipTmuxCheck {
		// In test mode, skip actual notifications but still update flags
		return CheckAndNotifyWithNotifier(opts, nil)
	}
	return CheckAndNotifyWithNotifier(opts, NotifyAgent)
}

// CheckAndNotifyWithNotifier performs a single notification cycle with a custom notifier.
// This allows for testing without actual tmux calls.
// When notify is non-nil, it will be called for each agent that should be notified.
// The function handles two types of agents:
// - Phase 1: Stated agents (with recipient state in recipients.jsonl)
// - Phase 2: Stateless agents (mailbox but no recipient state)
func CheckAndNotifyWithNotifier(opts LoopOptions, notify NotifyFunc) error {
	opts.log("Starting notification cycle")

	// =========================================================================
	// Phase 1: Stated agents (existing logic)
	// =========================================================================

	// Read all recipient states
	recipients, err := mail.ReadAllRecipients(opts.RepoRoot)
	if err != nil {
		opts.log("Error reading recipients: %v", err)
		return err
	}

	opts.log("Found %d stated agents", len(recipients))

	// T018: Build statedSet from recipients for Phase 2 lookup (FR-002)
	statedSet := make(map[string]struct{}, len(recipients))
	for _, r := range recipients {
		statedSet[r.Recipient] = struct{}{}
	}

	// Check each recipient
	for _, recipient := range recipients {
		// Only process ready agents that haven't been notified
		if recipient.Status != mail.StatusReady {
			opts.log("Skipping stated agent %q: status=%s (not ready)", recipient.Recipient, recipient.Status)
			continue
		}
		if recipient.Notified {
			opts.log("Skipping stated agent %q: already notified", recipient.Recipient)
			continue
		}

		// Check for unread messages
		unread, err := mail.FindUnread(opts.RepoRoot, recipient.Recipient)
		if err != nil {
			opts.log("Error reading mailbox for stated agent %q: %v", recipient.Recipient, err)
			continue
		}

		if len(unread) == 0 {
			opts.log("Skipping stated agent %q: no unread messages", recipient.Recipient)
			continue
		}

		opts.log("Stated agent %q has %d unread message(s)", recipient.Recipient, len(unread))

		// Send notification
		if notify != nil {
			opts.log("Notifying stated agent %q", recipient.Recipient)
			if err := notify(recipient.Recipient); err != nil {
				opts.log("Notification failed for stated agent %q: %v", recipient.Recipient, err)
				continue
			}
			opts.log("Notification sent to stated agent %q", recipient.Recipient)
		}

		// Update notified flag
		if err := mail.SetNotifiedFlag(opts.RepoRoot, recipient.Recipient, true); err != nil {
			opts.log("Error setting notified flag for %q: %v", recipient.Recipient, err)
			continue
		}
		opts.log("Marked stated agent %q as notified", recipient.Recipient)
	}

	// =========================================================================
	// Phase 2: Stateless agents (T019-T024)
	// =========================================================================

	// Skip Phase 2 if no tracker is configured
	if opts.StatelessTracker == nil {
		opts.log("Stateless tracking disabled, skipping Phase 2")
		return nil
	}

	// T019: Get all mailbox recipients (FR-001)
	mailboxRecipients, err := mail.ListMailboxRecipients(opts.RepoRoot)
	if err != nil {
		opts.log("Error listing mailbox recipients: %v", err)
		return nil
	}

	opts.log("Found %d mailbox recipients, checking for stateless agents", len(mailboxRecipients))

	// T020-T024: Process each stateless agent
	statelessCount := 0
	for _, mailboxRecipient := range mailboxRecipients {
		// T020: Skip agents that have recipient state (FR-003)
		if _, isStated := statedSet[mailboxRecipient]; isStated {
			continue
		}
		statelessCount++

		// T021: Check for unread messages (FR-006)
		unread, err := mail.FindUnread(opts.RepoRoot, mailboxRecipient)
		if err != nil {
			opts.log("Error reading mailbox for stateless agent %q: %v", mailboxRecipient, err)
			continue
		}
		if len(unread) == 0 {
			opts.log("Skipping stateless agent %q: no unread messages", mailboxRecipient)
			continue
		}

		opts.log("Stateless agent %q has %d unread message(s)", mailboxRecipient, len(unread))

		// T022: Check if notification is due (FR-004, FR-005)
		if !opts.StatelessTracker.ShouldNotify(mailboxRecipient) {
			opts.log("Skipping stateless agent %q: interval not elapsed", mailboxRecipient)
			continue
		}

		// Send notification
		if notify != nil {
			opts.log("Notifying stateless agent %q", mailboxRecipient)
			if err := notify(mailboxRecipient); err != nil {
				opts.log("Notification failed for stateless agent %q: %v", mailboxRecipient, err)
				continue
			}
			opts.log("Notification sent to stateless agent %q", mailboxRecipient)
		}

		// T023: Mark as notified
		opts.StatelessTracker.MarkNotified(mailboxRecipient)
		opts.log("Marked stateless agent %q in tracker", mailboxRecipient)
	}

	opts.log("Found %d stateless agents", statelessCount)

	// T024: Cleanup tracker with current mailbox list (FR-011)
	opts.StatelessTracker.Cleanup(mailboxRecipients)
	opts.log("Cleaned up stale entries from stateless tracker")

	opts.log("Notification cycle complete")
	return nil
}

// cleanStaleStates removes recipient states older than the threshold.
func cleanStaleStates(repoRoot string, logger io.Writer) {
	if logger != nil {
		fmt.Fprintf(logger, "[mailman] Cleaning stale recipient states (threshold: %v)\n", DefaultStaleThreshold)
	}
	_ = mail.CleanStaleStates(repoRoot, DefaultStaleThreshold) // G104: best-effort cleanup, errors don't stop the daemon
}

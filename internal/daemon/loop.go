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

// WindowCheckerFunc is the function signature for checking if a window exists.
type WindowCheckerFunc func(window string) (bool, error)

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
		return CheckAndNotifyWithNotifier(opts, nil, nil)
	}
	return CheckAndNotifyWithNotifier(opts, NotifyAgent, tmux.WindowExists)
}

// notifyStatedAgents processes agents with recipient state in recipients.jsonl.
// Returns the set of stated agent names for Phase 2 exclusion.
func notifyStatedAgents(opts LoopOptions, notify NotifyFunc) (map[string]struct{}, error) {
	recipients, err := mail.ReadAllRecipients(opts.RepoRoot)
	if err != nil {
		opts.log("Error reading recipients: %v", err)
		return nil, err
	}

	opts.log("Found %d stated agents", len(recipients))

	statedSet := make(map[string]struct{}, len(recipients))
	for _, r := range recipients {
		statedSet[r.Recipient] = struct{}{}
	}

	for _, recipient := range recipients {
		processStatedAgent(opts, notify, recipient)
	}

	return statedSet, nil
}

// processStatedAgent handles notification logic for a single stated agent.
func processStatedAgent(opts LoopOptions, notify NotifyFunc, recipient mail.RecipientState) {
	// Skip non-ready agents
	if recipient.Status != mail.StatusReady {
		if recipient.IsProtected() {
			opts.log("Skipping stated agent %q: status=%s, protected for 1h", recipient.Recipient, recipient.Status)
		} else {
			opts.log("Skipping stated agent %q: status=%s (not ready)", recipient.Recipient, recipient.Status)
		}
		return
	}

	// Check 60s debounce for ready agents
	if !recipient.ShouldNotify() {
		opts.log("Skipping stated agent %q: notified within last 60s", recipient.Recipient)
		return
	}

	// Check for unread messages
	unread, err := mail.FindUnread(opts.RepoRoot, recipient.Recipient)
	if err != nil {
		opts.log("Error reading mailbox for stated agent %q: %v", recipient.Recipient, err)
		return
	}
	if len(unread) == 0 {
		opts.log("Skipping stated agent %q: no unread messages", recipient.Recipient)
		return
	}

	opts.log("Stated agent %q has %d unread message(s)", recipient.Recipient, len(unread))

	// Send notification
	if notify != nil {
		opts.log("Notifying stated agent %q", recipient.Recipient)
		if err := notify(recipient.Recipient); err != nil {
			opts.log("Notification failed for stated agent %q: %v", recipient.Recipient, err)
			return
		}
		opts.log("Notification sent to stated agent %q", recipient.Recipient)
	}

	// Update notified flag
	if err := mail.SetNotifiedFlag(opts.RepoRoot, recipient.Recipient, true); err != nil {
		opts.log("Error setting notified flag for %q: %v", recipient.Recipient, err)
		return
	}
	opts.log("Marked stated agent %q as notified", recipient.Recipient)
}

// notifyStatelessAgents processes agents with mailboxes but no recipient state.
func notifyStatelessAgents(opts LoopOptions, notify NotifyFunc, windowChecker WindowCheckerFunc, statedSet map[string]struct{}) {
	if opts.StatelessTracker == nil {
		opts.log("Stateless tracking disabled, skipping Phase 2")
		return
	}

	mailboxRecipients, err := mail.ListMailboxRecipients(opts.RepoRoot)
	if err != nil {
		opts.log("Error listing mailbox recipients: %v", err)
		return
	}

	opts.log("Found %d mailbox recipients, checking for stateless agents", len(mailboxRecipients))

	statelessCount := 0
	for _, agent := range mailboxRecipients {
		if _, isStated := statedSet[agent]; isStated {
			continue
		}
		statelessCount++
		processStatelessAgent(opts, notify, windowChecker, agent)
	}

	opts.log("Found %d stateless agents", statelessCount)

	opts.StatelessTracker.Cleanup(mailboxRecipients)
	opts.log("Cleaned up stale entries from stateless tracker")
}

// processStatelessAgent handles notification logic for a single stateless agent.
func processStatelessAgent(opts LoopOptions, notify NotifyFunc, windowChecker WindowCheckerFunc, agent string) {
	// Check for unread messages
	unread, err := mail.FindUnread(opts.RepoRoot, agent)
	if err != nil {
		opts.log("Error reading mailbox for stateless agent %q: %v", agent, err)
		return
	}
	if len(unread) == 0 {
		opts.log("Skipping stateless agent %q: no unread messages", agent)
		return
	}

	opts.log("Stateless agent %q has %d unread message(s)", agent, len(unread))

	// Check if notification is due
	if !opts.StatelessTracker.ShouldNotify(agent) {
		opts.log("Skipping stateless agent %q: interval not elapsed", agent)
		return
	}

	// Verify window exists
	if windowChecker != nil {
		exists, err := windowChecker(agent)
		if err != nil {
			opts.log("Error checking window existence for %q: %v", agent, err)
			return
		}
		if !exists {
			opts.log("Skipping stateless agent %q: window does not exist", agent)
			opts.StatelessTracker.MarkNotified(agent)
			return
		}
	}

	// Send notification
	if notify != nil {
		opts.log("Notifying stateless agent %q", agent)
		if err := notify(agent); err != nil {
			opts.log("Notification failed for stateless agent %q: %v", agent, err)
			opts.StatelessTracker.MarkNotified(agent)
			opts.log("Marked stateless agent %q in tracker (after failure)", agent)
			return
		}
		opts.log("Notification sent to stateless agent %q", agent)
	}

	opts.StatelessTracker.MarkNotified(agent)
	opts.log("Marked stateless agent %q in tracker", agent)
}

// CheckAndNotifyWithNotifier performs a single notification cycle with a custom notifier.
// This allows for testing without actual tmux calls.
// The function handles two types of agents:
// - Phase 1: Stated agents (with recipient state in recipients.jsonl)
// - Phase 2: Stateless agents (mailbox but no recipient state)
func CheckAndNotifyWithNotifier(opts LoopOptions, notify NotifyFunc, windowChecker WindowCheckerFunc) error {
	opts.log("Starting notification cycle")

	// Phase 1: Stated agents
	statedSet, err := notifyStatedAgents(opts, notify)
	if err != nil {
		return err
	}

	// Phase 2: Stateless agents
	notifyStatelessAgents(opts, notify, windowChecker, statedSet)

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

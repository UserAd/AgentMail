// Package daemon provides functionality for the mailman daemon process.
// This file contains the notification loop implementation.
package daemon

import (
	"sync"
	"time"

	"agentmail/internal/mail"
	"agentmail/internal/tmux"
)

// DefaultLoopInterval is the default interval between notification checks.
const DefaultLoopInterval = 10 * time.Second

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

// LoopOptions configures the notification loop.
type LoopOptions struct {
	RepoRoot         string            // Repository root path
	Interval         time.Duration     // Loop interval (default 10s)
	StopChan         chan struct{}     // Channel to stop the loop
	SkipTmuxCheck    bool              // Skip tmux check (for testing)
	StatelessTracker *StatelessTracker // Tracker for stateless agents (T003)
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
func CheckAndNotifyWithNotifier(opts LoopOptions, notify NotifyFunc) error {
	// Read all recipient states
	recipients, err := mail.ReadAllRecipients(opts.RepoRoot)
	if err != nil {
		return err
	}

	// Check each recipient
	for _, recipient := range recipients {
		// Only process ready agents that haven't been notified
		if recipient.Status != mail.StatusReady {
			continue
		}
		if recipient.Notified {
			continue
		}

		// Check for unread messages
		unread, err := mail.FindUnread(opts.RepoRoot, recipient.Recipient)
		if err != nil {
			// Log error but continue with other recipients
			continue
		}

		if len(unread) == 0 {
			// No unread messages, skip notification
			continue
		}

		// Send notification
		if notify != nil {
			if err := notify(recipient.Recipient); err != nil {
				// Notification failed, don't mark as notified
				continue
			}
		}

		// Update notified flag
		if err := mail.SetNotifiedFlag(opts.RepoRoot, recipient.Recipient, true); err != nil {
			// Log error but continue
			continue
		}
	}

	return nil
}

// RunLoop runs the notification loop at the configured interval.
// It stops when the StopChan is closed.
func RunLoop(opts LoopOptions) {
	// Set default interval if not specified
	interval := opts.Interval
	if interval == 0 {
		interval = DefaultLoopInterval
	}

	// Create ticker for the loop
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run initial check immediately
	_ = CheckAndNotify(opts) // G104: errors are logged but don't stop the loop

	// Also clean stale states periodically
	cleanStaleStates(opts.RepoRoot)

	// Loop until stopped
	for {
		select {
		case <-opts.StopChan:
			return
		case <-ticker.C:
			// Perform notification check
			_ = CheckAndNotify(opts) // G104: errors are logged but don't stop the loop

			// Clean stale states periodically
			cleanStaleStates(opts.RepoRoot)
		}
	}
}

// cleanStaleStates removes recipient states older than the threshold.
func cleanStaleStates(repoRoot string) {
	_ = mail.CleanStaleStates(repoRoot, DefaultStaleThreshold) // G104: best-effort cleanup, errors don't stop the daemon
}

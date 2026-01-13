// Package daemon provides functionality for the mailman daemon process.
// This file contains the file watcher abstraction for instant notifications.
package daemon

import (
	"sync"
	"time"
)

// MonitoringMode indicates the daemon's current monitoring strategy.
type MonitoringMode int

const (
	// ModeWatching indicates event-driven file watching is active.
	ModeWatching MonitoringMode = iota
	// ModePolling indicates timer-based polling (fallback mode).
	ModePolling
)

// DefaultDebounceWindow is the default debounce window for file events (500ms per FR-011).
const DefaultDebounceWindow = 500 * time.Millisecond

// FallbackTimerInterval is the interval for the safety net notification check in watching mode (60s per FR-012).
const FallbackTimerInterval = 60 * time.Second

// Debouncer coalesces rapid file change events using a trailing-edge debounce.
// It ensures that a callback is only triggered after no triggers have occurred
// for the specified duration.
type Debouncer struct {
	timer    *time.Timer   // Active timer (nil if no pending trigger)
	duration time.Duration // Debounce window (500ms per FR-011)
	mu       sync.Mutex    // Protects timer access
}

// NewDebouncer creates a new Debouncer with the specified duration.
func NewDebouncer(duration time.Duration) *Debouncer {
	return &Debouncer{
		duration: duration,
	}
}

// Trigger schedules the callback to be called after the debounce window.
// If Trigger is called again before the window expires, the timer is reset.
// This implements a trailing-edge debounce pattern.
func (d *Debouncer) Trigger(callback func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Stop any existing timer
	if d.timer != nil {
		d.timer.Stop()
	}

	// Start a new timer
	d.timer = time.AfterFunc(d.duration, callback)
}

// Stop cancels any pending timer. Should be called when shutting down.
func (d *Debouncer) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
}

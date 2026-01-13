// Package daemon provides functionality for the mailman daemon process.
// This file contains the file watcher abstraction for instant notifications.
package daemon

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"agentmail/internal/mail"

	"github.com/fsnotify/fsnotify"
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

// FileWatcher provides file system watching for instant notifications.
// It watches .agentmail/ and .agentmail/mailboxes/ directories for changes.
type FileWatcher struct {
	watcher      *fsnotify.Watcher // Underlying fsnotify watcher
	debouncer    *Debouncer        // Debouncer for coalescing events
	mailboxDir   string            // Path to .agentmail/mailboxes/
	agentmailDir string            // Path to .agentmail/
	stopChan     chan struct{}     // Signal to stop the watcher
	mode         MonitoringMode    // Current monitoring mode
	mu           sync.Mutex        // Protects mode
	logger       io.Writer         // Logger for foreground mode (nil = no logging)
}

// log writes a formatted message to the logger if configured.
func (fw *FileWatcher) log(format string, args ...interface{}) {
	if fw.logger != nil {
		fmt.Fprintf(fw.logger, "[mailman] "+format+"\n", args...)
	}
}

// NewFileWatcher creates a new FileWatcher for the given repository root.
// Returns an error if the watcher cannot be initialized (e.g., OS doesn't support it).
// Creates the .agentmail/ and .agentmail/mailboxes/ directories if they don't exist (FR-006, FR-007).
func NewFileWatcher(repoRoot string) (*FileWatcher, error) {
	// Create directories if needed (FR-006, FR-007)
	if err := mail.EnsureMailDir(repoRoot); err != nil {
		return nil, err
	}

	// Create fsnotify watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fw := &FileWatcher{
		watcher:      watcher,
		debouncer:    NewDebouncer(DefaultDebounceWindow),
		mailboxDir:   filepath.Join(repoRoot, mail.MailDir),
		agentmailDir: filepath.Join(repoRoot, mail.RootDir),
		stopChan:     make(chan struct{}),
		mode:         ModeWatching,
	}

	return fw, nil
}

// SetLogger sets the logger for the file watcher.
func (fw *FileWatcher) SetLogger(logger io.Writer) {
	fw.logger = logger
}

// AddWatches adds watches for .agentmail/ and .agentmail/mailboxes/ directories (FR-001, FR-004).
func (fw *FileWatcher) AddWatches() error {
	// Watch .agentmail/ for recipients.jsonl changes (FR-005)
	if err := fw.watcher.Add(fw.agentmailDir); err != nil {
		return err
	}
	fw.log("Watching directory: %s", fw.agentmailDir)

	// Watch .agentmail/mailboxes/ for mailbox file changes (FR-004)
	// Check if directory exists first
	if _, err := os.Stat(fw.mailboxDir); err == nil {
		if err := fw.watcher.Add(fw.mailboxDir); err != nil {
			return err
		}
		fw.log("Watching directory: %s", fw.mailboxDir)
	}
	// Note: If mailboxes/ doesn't exist yet, we'll still get events when it's created
	// because we're watching the parent .agentmail/ directory

	return nil
}

// isMailboxEvent checks if the event is for a mailbox file (Write/Create on .jsonl in mailboxes/).
func (fw *FileWatcher) isMailboxEvent(event fsnotify.Event) bool {
	// Must be Write or Create operation
	if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) {
		return false
	}

	// Must be in mailboxes directory
	dir := filepath.Dir(event.Name)
	if dir != fw.mailboxDir {
		return false
	}

	// Must be a .jsonl file
	return strings.HasSuffix(event.Name, ".jsonl")
}

// isRecipientsEvent checks if the event is a Write to recipients.jsonl (FR-005).
func (fw *FileWatcher) isRecipientsEvent(event fsnotify.Event) bool {
	// Must be Write operation (status changes are writes, not creates)
	if !event.Has(fsnotify.Write) {
		return false
	}

	// Must be recipients.jsonl in .agentmail/
	expectedPath := filepath.Join(fw.agentmailDir, "recipients.jsonl")
	return event.Name == expectedPath
}

// isMailboxDirCreate checks if the mailboxes directory was just created.
// This handles the case where mailboxes/ doesn't exist at startup (FR-008).
func (fw *FileWatcher) isMailboxDirCreate(event fsnotify.Event) bool {
	if !event.Has(fsnotify.Create) {
		return false
	}
	return event.Name == fw.mailboxDir
}

// Run starts the file watcher event loop.
// It calls processFunc when mailbox or recipients.jsonl changes are detected (debounced).
// Returns when Close() is called or an error occurs.
func (fw *FileWatcher) Run(processFunc func()) error {
	// Create fallback ticker for safety net (FR-012)
	fallbackTicker := time.NewTicker(FallbackTimerInterval)
	defer fallbackTicker.Stop()

	fw.log("Starting file watcher event loop (fallback interval: %v)", FallbackTimerInterval)

	for {
		select {
		case <-fw.stopChan:
			fw.log("Received stop signal, shutting down file watcher")
			return nil

		case event, ok := <-fw.watcher.Events:
			if !ok {
				return nil
			}

			// Check if this is a mailbox event (FR-009)
			if fw.isMailboxEvent(event) {
				fw.log("Mailbox change detected: %s (%s)", filepath.Base(event.Name), event.Op)
				// Trigger debounced notification check (FR-011)
				fw.debouncer.Trigger(processFunc)
				continue
			}

			// Check if this is a recipients.jsonl event (FR-005, FR-010a, FR-010b)
			if fw.isRecipientsEvent(event) {
				fw.log("Recipients state change detected (%s)", event.Op)
				// Trigger debounced notification check - reloads states and checks notifications
				fw.debouncer.Trigger(processFunc)
				continue
			}

			// Check if mailboxes directory was created (FR-008)
			if fw.isMailboxDirCreate(event) {
				fw.log("Mailboxes directory created, adding watch")
				// Add watch for the newly created mailboxes directory
				_ = fw.watcher.Add(fw.mailboxDir) // G104: best-effort, errors handled by fallback
			}

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return nil
			}
			fw.log("Watcher error: %v", err)
			// Return error to allow caller to handle fallback (FR-014a, FR-014b)
			return err

		case <-fallbackTicker.C:
			fw.log("Fallback timer tick: running notification check")
			// Safety net: check for notifications even if no events (FR-012)
			processFunc()
		}
	}
}

// Close stops the file watcher and releases resources.
func (fw *FileWatcher) Close() error {
	// Signal stop
	close(fw.stopChan)

	// Stop debouncer
	fw.debouncer.Stop()

	// Close fsnotify watcher
	return fw.watcher.Close()
}

// Mode returns the current monitoring mode.
func (fw *FileWatcher) Mode() MonitoringMode {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	return fw.mode
}

// SetMode sets the monitoring mode.
func (fw *FileWatcher) SetMode(mode MonitoringMode) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.mode = mode
}

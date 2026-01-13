package daemon

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// T029: Unit tests for Debouncer
// =============================================================================

func TestNewDebouncer_CreatesDebouncerWithDuration(t *testing.T) {
	d := NewDebouncer(100 * time.Millisecond)
	if d == nil {
		t.Fatal("NewDebouncer returned nil")
	}
	if d.duration != 100*time.Millisecond {
		t.Errorf("Expected duration 100ms, got %v", d.duration)
	}
	if d.timer != nil {
		t.Error("Expected timer to be nil initially")
	}
	if d.ready == nil {
		t.Error("Expected ready channel to be initialized")
	}
}

func TestDebouncer_Trigger_SignalsReadyAfterDuration(t *testing.T) {
	d := NewDebouncer(50 * time.Millisecond)
	defer d.Stop()

	d.Trigger()

	// Ready channel should not have signal immediately
	select {
	case <-d.Ready():
		t.Error("Ready channel should not signal immediately")
	default:
		// Expected - no signal yet
	}

	// Wait for debounce window to pass
	select {
	case <-d.Ready():
		// Expected - signal received
	case <-time.After(100 * time.Millisecond):
		t.Error("Ready channel should have signaled after debounce window")
	}
}

func TestDebouncer_Trigger_ResetsTimerOnMultipleCalls(t *testing.T) {
	d := NewDebouncer(100 * time.Millisecond)
	defer d.Stop()

	var signalCount atomic.Int32

	// Trigger multiple times within the debounce window
	d.Trigger()
	time.Sleep(30 * time.Millisecond)
	d.Trigger()
	time.Sleep(30 * time.Millisecond)
	d.Trigger()

	// Wait for debounce window to pass after last trigger
	time.Sleep(150 * time.Millisecond)

	// Drain all signals from ready channel
	for {
		select {
		case <-d.Ready():
			signalCount.Add(1)
		default:
			goto done
		}
	}
done:

	// Should only signal once (trailing-edge debounce, buffered channel size 1)
	if signalCount.Load() != 1 {
		t.Errorf("Expected ready channel to signal once due to debouncing, got %d", signalCount.Load())
	}
}

func TestDebouncer_Stop_CancelsPendingTimer(t *testing.T) {
	d := NewDebouncer(100 * time.Millisecond)

	d.Trigger()
	d.Stop()

	// Wait for what would have been the debounce window
	time.Sleep(150 * time.Millisecond)

	// Ready channel should not have signal
	select {
	case <-d.Ready():
		t.Error("Ready channel should not signal after Stop")
	default:
		// Expected - no signal
	}
}

func TestDebouncer_Stop_SafeToCallMultipleTimes(t *testing.T) {
	d := NewDebouncer(50 * time.Millisecond)

	// Should not panic
	d.Stop()
	d.Stop()
	d.Stop()
}

func TestDebouncer_Stop_SafeToCallWithoutTrigger(t *testing.T) {
	d := NewDebouncer(50 * time.Millisecond)

	// Should not panic when stopping without ever triggering
	d.Stop()
}

// =============================================================================
// T030: Unit tests for FileWatcher initialization and fallback
// =============================================================================

func TestNewFileWatcher_CreatesWatcherAndDirectories(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-watcher-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fw, err := NewFileWatcher(tmpDir)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer fw.Close()

	// Check directories were created
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if _, err := os.Stat(agentmailDir); os.IsNotExist(err) {
		t.Error(".agentmail directory was not created")
	}

	mailboxDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if _, err := os.Stat(mailboxDir); os.IsNotExist(err) {
		t.Error(".agentmail/mailboxes directory was not created")
	}

	// Check watcher fields
	if fw.agentmailDir != agentmailDir {
		t.Errorf("Expected agentmailDir %s, got %s", agentmailDir, fw.agentmailDir)
	}
	if fw.mailboxDir != mailboxDir {
		t.Errorf("Expected mailboxDir %s, got %s", mailboxDir, fw.mailboxDir)
	}
	if fw.mode != ModeWatching {
		t.Error("Expected initial mode to be ModeWatching")
	}
}

func TestNewFileWatcher_ReturnsErrorForInvalidPath(t *testing.T) {
	// Use a path that cannot be created (file instead of directory)
	tmpFile, err := os.CreateTemp("", "agentmail-watcher-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// Try to create watcher with a file path (should fail to create directories)
	_, err = NewFileWatcher(tmpFile.Name())
	if err == nil {
		t.Error("Expected error when creating watcher with invalid path")
	}
}

func TestFileWatcher_AddWatches_AddsWatchesToDirectories(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-watcher-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fw, err := NewFileWatcher(tmpDir)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer fw.Close()

	err = fw.AddWatches()
	if err != nil {
		t.Fatalf("AddWatches failed: %v", err)
	}

	// The watcher should have added watches (we can't easily verify this
	// without triggering events, but at least it shouldn't error)
}

func TestFileWatcher_Close_StopsWatcher(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-watcher-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fw, err := NewFileWatcher(tmpDir)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}

	err = fw.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}
}

func TestFileWatcher_Mode_ReturnsCurrentMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-watcher-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fw, err := NewFileWatcher(tmpDir)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer fw.Close()

	if fw.Mode() != ModeWatching {
		t.Error("Expected initial mode to be ModeWatching")
	}

	fw.SetMode(ModePolling)
	if fw.Mode() != ModePolling {
		t.Error("Expected mode to be ModePolling after SetMode")
	}
}

func TestMonitoringMode_Constants(t *testing.T) {
	// Verify constants have expected values
	if ModeWatching != 0 {
		t.Error("ModeWatching should be 0 (iota)")
	}
	if ModePolling != 1 {
		t.Error("ModePolling should be 1")
	}
}

func TestDefaultDebounceWindow_Is500ms(t *testing.T) {
	if DefaultDebounceWindow != 500*time.Millisecond {
		t.Errorf("DefaultDebounceWindow should be 500ms, got %v", DefaultDebounceWindow)
	}
}

func TestFallbackTimerInterval_Is60s(t *testing.T) {
	if FallbackTimerInterval != 60*time.Second {
		t.Errorf("FallbackTimerInterval should be 60s, got %v", FallbackTimerInterval)
	}
}

// =============================================================================
// T031: Tests for FileWatcher.Run event handling
// =============================================================================

func TestFileWatcher_Run_CallsProcessFuncOnMailboxEvent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-watcher-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fw, err := NewFileWatcher(tmpDir)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}

	err = fw.AddWatches()
	if err != nil {
		t.Fatalf("AddWatches failed: %v", err)
	}

	var called atomic.Int32
	processFunc := func() {
		called.Add(1)
	}

	// Run watcher in goroutine
	done := make(chan struct{})
	go func() {
		_ = fw.Run(processFunc)
		close(done)
	}()

	// Give watcher time to start
	time.Sleep(50 * time.Millisecond)

	// Trigger a mailbox event by writing to a .jsonl file in mailboxes/
	mailboxFile := filepath.Join(tmpDir, ".agentmail", "mailboxes", "test-agent.jsonl")
	if err := os.WriteFile(mailboxFile, []byte("{\"test\":1}\n"), 0600); err != nil {
		t.Fatalf("Failed to write mailbox file: %v", err)
	}

	// Wait for debounce window (500ms) + buffer
	time.Sleep(700 * time.Millisecond)

	// Close watcher
	_ = fw.Close()
	<-done

	// processFunc should have been called at least once (debounced)
	if called.Load() < 1 {
		t.Errorf("Expected processFunc to be called at least once on mailbox event, got %d", called.Load())
	}
}

func TestFileWatcher_Run_CallsProcessFuncOnRecipientsEvent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-watcher-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fw, err := NewFileWatcher(tmpDir)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}

	err = fw.AddWatches()
	if err != nil {
		t.Fatalf("AddWatches failed: %v", err)
	}

	var called atomic.Int32
	processFunc := func() {
		called.Add(1)
	}

	// Run watcher in goroutine
	done := make(chan struct{})
	go func() {
		_ = fw.Run(processFunc)
		close(done)
	}()

	// Give watcher time to start
	time.Sleep(50 * time.Millisecond)

	// First create the recipients file (this is a Create event, not Write)
	recipientsFile := filepath.Join(tmpDir, ".agentmail", "recipients.jsonl")
	if err := os.WriteFile(recipientsFile, []byte("{\"recipient\":\"agent1\"}\n"), 0600); err != nil {
		t.Fatalf("Failed to write recipients file: %v", err)
	}

	// Give some time for initial event
	time.Sleep(100 * time.Millisecond)

	// Now write to it again (this is a Write event which triggers recipients event)
	if err := os.WriteFile(recipientsFile, []byte("{\"recipient\":\"agent2\"}\n"), 0600); err != nil {
		t.Fatalf("Failed to write recipients file: %v", err)
	}

	// Wait for debounce window (500ms) + buffer
	time.Sleep(700 * time.Millisecond)

	// Close watcher
	_ = fw.Close()
	<-done

	// processFunc should have been called at least once
	if called.Load() < 1 {
		t.Errorf("Expected processFunc to be called at least once on recipients event, got %d", called.Load())
	}
}

func TestFileWatcher_Run_StopsOnClose(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-watcher-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fw, err := NewFileWatcher(tmpDir)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}

	err = fw.AddWatches()
	if err != nil {
		t.Fatalf("AddWatches failed: %v", err)
	}

	processFunc := func() {}

	// Run watcher in goroutine
	done := make(chan struct{})
	go func() {
		_ = fw.Run(processFunc)
		close(done)
	}()

	// Give watcher time to start
	time.Sleep(50 * time.Millisecond)

	// Close watcher
	err = fw.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}

	// Run should return after Close
	select {
	case <-done:
		// Success - watcher stopped
	case <-time.After(1 * time.Second):
		t.Error("Watcher did not stop within timeout after Close")
	}
}

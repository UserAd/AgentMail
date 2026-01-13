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
}

func TestDebouncer_Trigger_CallsCallbackAfterDuration(t *testing.T) {
	d := NewDebouncer(50 * time.Millisecond)
	defer d.Stop()

	var called atomic.Int32
	callback := func() {
		called.Add(1)
	}

	d.Trigger(callback)

	// Should not be called immediately
	if called.Load() != 0 {
		t.Error("Callback should not be called immediately")
	}

	// Wait for debounce window to pass
	time.Sleep(100 * time.Millisecond)

	if called.Load() != 1 {
		t.Errorf("Expected callback to be called once, got %d", called.Load())
	}
}

func TestDebouncer_Trigger_ResetsTimerOnMultipleCalls(t *testing.T) {
	d := NewDebouncer(100 * time.Millisecond)
	defer d.Stop()

	var called atomic.Int32
	callback := func() {
		called.Add(1)
	}

	// Trigger multiple times within the debounce window
	d.Trigger(callback)
	time.Sleep(30 * time.Millisecond)
	d.Trigger(callback)
	time.Sleep(30 * time.Millisecond)
	d.Trigger(callback)

	// Wait for debounce window to pass after last trigger
	time.Sleep(150 * time.Millisecond)

	// Should only be called once (trailing-edge debounce)
	if called.Load() != 1 {
		t.Errorf("Expected callback to be called once due to debouncing, got %d", called.Load())
	}
}

func TestDebouncer_Stop_CancelsPendingTimer(t *testing.T) {
	d := NewDebouncer(100 * time.Millisecond)

	var called atomic.Int32
	callback := func() {
		called.Add(1)
	}

	d.Trigger(callback)
	d.Stop()

	// Wait for what would have been the debounce window
	time.Sleep(150 * time.Millisecond)

	// Callback should not have been called
	if called.Load() != 0 {
		t.Error("Callback should not be called after Stop")
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

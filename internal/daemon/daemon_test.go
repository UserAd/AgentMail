package daemon

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// =============================================================================
// T012: Tests for PID file operations (ReadPID, WritePID, DeletePID)
// =============================================================================

func TestWritePID_CreatesFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-daemon-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git/mail directory
	mailDir := filepath.Join(tmpDir, ".git", "mail")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	// Write PID
	testPID := 12345
	err = WritePID(tmpDir, testPID)
	if err != nil {
		t.Fatalf("WritePID failed: %v", err)
	}

	// Verify file exists with correct content
	pidFile := filepath.Join(mailDir, "mailman.pid")
	content, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("Failed to read PID file: %v", err)
	}

	expected := "12345\n"
	if string(content) != expected {
		t.Errorf("PID file content mismatch. Expected %q, got %q", expected, string(content))
	}
}

func TestWritePID_OverwritesExisting(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-daemon-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git/mail directory
	mailDir := filepath.Join(tmpDir, ".git", "mail")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	// Write initial PID file
	pidFile := filepath.Join(mailDir, "mailman.pid")
	if err := os.WriteFile(pidFile, []byte("99999\n"), 0644); err != nil {
		t.Fatalf("Failed to create initial PID file: %v", err)
	}

	// Overwrite with new PID
	newPID := 54321
	err = WritePID(tmpDir, newPID)
	if err != nil {
		t.Fatalf("WritePID failed: %v", err)
	}

	// Verify file has new content
	content, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("Failed to read PID file: %v", err)
	}

	expected := "54321\n"
	if string(content) != expected {
		t.Errorf("PID file should be overwritten. Expected %q, got %q", expected, string(content))
	}
}

func TestReadPID_ReadsExistingFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-daemon-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git/mail directory and PID file
	mailDir := filepath.Join(tmpDir, ".git", "mail")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	pidFile := filepath.Join(mailDir, "mailman.pid")
	if err := os.WriteFile(pidFile, []byte("67890\n"), 0644); err != nil {
		t.Fatalf("Failed to create PID file: %v", err)
	}

	// Read PID
	pid, err := ReadPID(tmpDir)
	if err != nil {
		t.Fatalf("ReadPID failed: %v", err)
	}

	if pid != 67890 {
		t.Errorf("Expected PID 67890, got %d", pid)
	}
}

func TestReadPID_ReturnsZeroForMissingFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-daemon-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git/mail directory but no PID file
	mailDir := filepath.Join(tmpDir, ".git", "mail")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	// Read PID - should return 0, no error for missing file
	pid, err := ReadPID(tmpDir)
	if err != nil {
		t.Fatalf("ReadPID should not error for missing file: %v", err)
	}

	if pid != 0 {
		t.Errorf("Expected PID 0 for missing file, got %d", pid)
	}
}

func TestReadPID_ErrorsOnInvalidContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-daemon-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git/mail directory and invalid PID file
	mailDir := filepath.Join(tmpDir, ".git", "mail")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	pidFile := filepath.Join(mailDir, "mailman.pid")
	if err := os.WriteFile(pidFile, []byte("not-a-number\n"), 0644); err != nil {
		t.Fatalf("Failed to create PID file: %v", err)
	}

	// Read PID - should error on invalid content
	_, err = ReadPID(tmpDir)
	if err == nil {
		t.Error("ReadPID should error for invalid PID content")
	}
}

func TestDeletePID_RemovesFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-daemon-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git/mail directory and PID file
	mailDir := filepath.Join(tmpDir, ".git", "mail")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	pidFile := filepath.Join(mailDir, "mailman.pid")
	if err := os.WriteFile(pidFile, []byte("12345\n"), 0644); err != nil {
		t.Fatalf("Failed to create PID file: %v", err)
	}

	// Delete PID
	err = DeletePID(tmpDir)
	if err != nil {
		t.Fatalf("DeletePID failed: %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
		t.Error("PID file should be deleted")
	}
}

func TestDeletePID_NoErrorForMissingFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-daemon-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git/mail directory but no PID file
	mailDir := filepath.Join(tmpDir, ".git", "mail")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	// Delete PID - should not error for missing file
	err = DeletePID(tmpDir)
	if err != nil {
		t.Errorf("DeletePID should not error for missing file: %v", err)
	}
}

// =============================================================================
// T013: Tests for foreground mode startup
// =============================================================================

func TestStartDaemon_ForegroundMode_WritesPID(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-daemon-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git/mail directory
	mailDir := filepath.Join(tmpDir, ".git", "mail")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// Start daemon in foreground mode (daemonize=false)
	exitCode := StartDaemon(tmpDir, false, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Verify PID file was created
	pidFile := filepath.Join(mailDir, "mailman.pid")
	content, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("PID file should exist: %v", err)
	}

	// Verify PID is current process PID
	pid, err := strconv.Atoi(string(content[:len(content)-1])) // strip newline
	if err != nil {
		t.Fatalf("PID file should contain valid number: %v", err)
	}

	currentPID := os.Getpid()
	if pid != currentPID {
		t.Errorf("PID file should contain current PID %d, got %d", currentPID, pid)
	}
}

func TestStartDaemon_ForegroundMode_OutputsStartupMessage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-daemon-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git/mail directory
	mailDir := filepath.Join(tmpDir, ".git", "mail")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	exitCode := StartDaemon(tmpDir, false, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify startup message format: "Mailman daemon started (PID: 12345)"
	output := stdout.String()
	currentPID := os.Getpid()
	expected := "Mailman daemon started (PID: " + strconv.Itoa(currentPID) + ")\n"

	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

func TestStartDaemon_AlreadyRunning_ExitsWithCode2(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-daemon-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git/mail directory
	mailDir := filepath.Join(tmpDir, ".git", "mail")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	// Create PID file with current process PID (simulating running daemon)
	// Using current PID ensures IsRunning returns true
	currentPID := os.Getpid()
	pidFile := filepath.Join(mailDir, "mailman.pid")
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(currentPID)+"\n"), 0644); err != nil {
		t.Fatalf("Failed to create PID file: %v", err)
	}

	var stdout, stderr bytes.Buffer

	exitCode := StartDaemon(tmpDir, false, &stdout, &stderr)

	if exitCode != 2 {
		t.Errorf("Expected exit code 2 (daemon already running), got %d", exitCode)
	}

	// Verify error message
	if !bytes.Contains(stderr.Bytes(), []byte("already running")) {
		t.Errorf("Expected error message about daemon already running, got: %s", stderr.String())
	}
}

func TestStartDaemon_StalePID_OverwritesAndStarts(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-daemon-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git/mail directory
	mailDir := filepath.Join(tmpDir, ".git", "mail")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	// Create PID file with a definitely non-existent PID (PID 1 is init, use very high PID)
	// PID 99999999 should not exist on any system
	stalePID := 99999999
	pidFile := filepath.Join(mailDir, "mailman.pid")
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(stalePID)+"\n"), 0644); err != nil {
		t.Fatalf("Failed to create PID file: %v", err)
	}

	var stdout, stderr bytes.Buffer

	exitCode := StartDaemon(tmpDir, false, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for stale PID, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Verify PID file was overwritten with current PID
	content, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("PID file should exist: %v", err)
	}

	currentPID := os.Getpid()
	expected := strconv.Itoa(currentPID) + "\n"
	if string(content) != expected {
		t.Errorf("PID file should be overwritten with current PID. Expected %q, got %q", expected, string(content))
	}
}

// =============================================================================
// T014: Tests for background mode (--daemon flag)
// =============================================================================

// Note: Testing actual daemonization is difficult in unit tests because:
// 1. The child process would be a separate process
// 2. We can't easily verify child behavior from parent test
//
// We test the StartDaemon function behavior with daemonize=true through
// integration tests. Here we test that the function signature accepts
// the daemonize flag and the basic flow.

func TestStartDaemon_BackgroundMode_OutputsBackgroundMessage(t *testing.T) {
	// This test verifies the parent behavior in background mode.
	// In real daemon mode, the function would fork and the parent returns immediately.
	// For unit tests, we mock this by testing with a special test flag.

	// Skip actual daemonization test - this would spawn a real process.
	// Integration tests will cover the full daemon flow.
	t.Skip("Background mode requires integration test - see internal/cli/mailman_test.go")
}

// =============================================================================
// Helper function tests
// =============================================================================

func TestIsRunning_ReturnsTrueForCurrentProcess(t *testing.T) {
	currentPID := os.Getpid()
	if !IsRunning(currentPID) {
		t.Error("IsRunning should return true for current process")
	}
}

func TestIsRunning_ReturnsFalseForNonExistent(t *testing.T) {
	// PID 99999999 should not exist
	if IsRunning(99999999) {
		t.Error("IsRunning should return false for non-existent PID")
	}
}

func TestIsRunning_ReturnsFalseForZero(t *testing.T) {
	if IsRunning(0) {
		t.Error("IsRunning should return false for PID 0")
	}
}

func TestPIDFilePath(t *testing.T) {
	expected := filepath.Join("/test/repo", ".git/mail/mailman.pid")
	result := PIDFilePath("/test/repo")
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

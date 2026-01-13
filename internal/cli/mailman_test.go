package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"agentmail/internal/daemon"
)

// =============================================================================
// Test helper functions
// =============================================================================

// setupTestStopChannel creates a stop channel for testing and ensures cleanup.
func setupTestStopChannel() (chan struct{}, func()) {
	stopCh := make(chan struct{})
	daemon.SetStopChannel(stopCh)
	return stopCh, func() {
		daemon.SetStopChannel(nil)
	}
}

// =============================================================================
// T022: Integration tests for Mailman CLI command
// =============================================================================

func TestMailman_ForegroundMode_Success(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-mailman-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.MkdirAll(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	// Set up test stop channel
	stopCh, cleanup := setupTestStopChannel()
	defer cleanup()

	var stdout, stderr bytes.Buffer
	var exitCode int

	// Start daemon in goroutine
	done := make(chan struct{})
	go func() {
		exitCode = Mailman(&stdout, &stderr, MailmanOptions{
			Daemonize: false,
			RepoRoot:  tmpDir,
		})
		close(done)
	}()

	// Give daemon time to start
	time.Sleep(50 * time.Millisecond)

	// Read PID file while daemon is running (safe - not being written to)
	pidFile := filepath.Join(agentmailDir, "mailman.pid")
	content, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("PID file should exist: %v", err)
	}

	currentPID := os.Getpid()
	expectedPID := strconv.Itoa(currentPID) + "\n"
	if string(content) != expectedPID {
		t.Errorf("PID file content mismatch. Expected %q, got %q", expectedPID, string(content))
	}

	// Stop the daemon first to avoid race when reading stdout
	close(stopCh)
	<-done

	// Verify startup message format (safe to read after daemon stopped)
	output := stdout.String()
	expected := "Mailman daemon started (PID: " + strconv.Itoa(currentPID) + ")"

	if !strings.Contains(output, expected) {
		t.Errorf("Expected output to contain %q, got %q", expected, output)
	}

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}
}

func TestMailman_AlreadyRunning_ExitsCode2(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-mailman-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.MkdirAll(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	// Create PID file with current process PID (simulating running daemon)
	currentPID := os.Getpid()
	pidFile := filepath.Join(agentmailDir, "mailman.pid")
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(currentPID)+"\n"), 0644); err != nil {
		t.Fatalf("Failed to create PID file: %v", err)
	}

	var stdout, stderr bytes.Buffer

	exitCode := Mailman(&stdout, &stderr, MailmanOptions{
		Daemonize: false,
		RepoRoot:  tmpDir,
	})

	if exitCode != 2 {
		t.Errorf("Expected exit code 2 (daemon already running), got %d", exitCode)
	}

	// Verify error message
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "already running") {
		t.Errorf("Expected error message about daemon already running, got: %s", errOutput)
	}

	// Verify stdout is empty
	if stdout.String() != "" {
		t.Errorf("Expected empty stdout, got: %s", stdout.String())
	}
}

func TestMailman_StalePID_StartsSuccessfully(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-mailman-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.MkdirAll(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	// Create PID file with a non-existent PID (stale)
	stalePID := 99999999
	pidFile := filepath.Join(agentmailDir, "mailman.pid")
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(stalePID)+"\n"), 0644); err != nil {
		t.Fatalf("Failed to create PID file: %v", err)
	}

	// Set up test stop channel
	stopCh, cleanup := setupTestStopChannel()
	defer cleanup()

	var stdout, stderr bytes.Buffer
	var exitCode int

	// Start daemon in goroutine
	done := make(chan struct{})
	go func() {
		exitCode = Mailman(&stdout, &stderr, MailmanOptions{
			Daemonize: false,
			RepoRoot:  tmpDir,
		})
		close(done)
	}()

	// Give daemon time to start
	time.Sleep(50 * time.Millisecond)

	// Read PID file while daemon is running (safe - not being written to)
	content, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("PID file should exist: %v", err)
	}

	currentPID := os.Getpid()
	expectedPID := strconv.Itoa(currentPID) + "\n"
	if string(content) != expectedPID {
		t.Errorf("PID file should be overwritten. Expected %q, got %q", expectedPID, string(content))
	}

	// Stop the daemon first to avoid race when reading stderr
	close(stopCh)
	<-done

	// Verify stale PID warning was output (safe to read after daemon stopped)
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Stale PID file found") {
		t.Errorf("Expected warning about stale PID file, got: %s", errOutput)
	}

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for stale PID, got %d. Stderr: %s", exitCode, errOutput)
	}
}

func TestMailman_NoPIDFile_StartsSuccessfully(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-mailman-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .agentmail directory (no PID file)
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.MkdirAll(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	// Set up test stop channel
	stopCh, cleanup := setupTestStopChannel()
	defer cleanup()

	var stdout, stderr bytes.Buffer
	var exitCode int

	// Start daemon in goroutine
	done := make(chan struct{})
	go func() {
		exitCode = Mailman(&stdout, &stderr, MailmanOptions{
			Daemonize: false,
			RepoRoot:  tmpDir,
		})
		close(done)
	}()

	// Give daemon time to start
	time.Sleep(50 * time.Millisecond)

	// Verify PID file was created
	pidFile := filepath.Join(agentmailDir, "mailman.pid")
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		t.Error("PID file should be created")
	}

	// Stop the daemon
	close(stopCh)
	<-done

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}
}

func TestMailman_CreatesMailDirIfNeeded(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-mailman-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Don't create .agentmail directory - daemon should create it

	// Set up test stop channel
	stopCh, cleanup := setupTestStopChannel()
	defer cleanup()

	var stdout, stderr bytes.Buffer
	var exitCode int

	// Start daemon in goroutine
	done := make(chan struct{})
	go func() {
		exitCode = Mailman(&stdout, &stderr, MailmanOptions{
			Daemonize: false,
			RepoRoot:  tmpDir,
		})
		close(done)
	}()

	// Give daemon time to start
	time.Sleep(50 * time.Millisecond)

	// Verify .agentmail directory was created
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	info, err := os.Stat(agentmailDir)
	if err != nil {
		t.Fatalf(".agentmail directory should be created: %v", err)
	}
	if !info.IsDir() {
		t.Error(".agentmail should be a directory")
	}

	// Verify PID file was created
	pidFile := filepath.Join(agentmailDir, "mailman.pid")
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		t.Error("PID file should be created")
	}

	// Stop the daemon
	close(stopCh)
	<-done

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}
}

// =============================================================================
// T031: Singleton integration test
// =============================================================================

func TestMailman_Singleton_SecondInstanceExitsCode2(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-mailman-singleton-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.MkdirAll(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	// Set up test stop channel for first daemon
	stopCh, cleanup := setupTestStopChannel()
	defer cleanup()

	var stdout1, stderr1 bytes.Buffer
	var exitCode1 int

	// Start first daemon in goroutine
	done1 := make(chan struct{})
	go func() {
		exitCode1 = Mailman(&stdout1, &stderr1, MailmanOptions{
			Daemonize: false,
			RepoRoot:  tmpDir,
		})
		close(done1)
	}()

	// Give first daemon time to start and write PID file
	time.Sleep(50 * time.Millisecond)

	// Verify PID file exists (safe to check file existence)
	pidFile := filepath.Join(agentmailDir, "mailman.pid")
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		t.Fatal("PID file should exist after first daemon start")
	}

	// Try to start second daemon - should fail with exit code 2
	// Note: Second daemon doesn't use stopCh, it exits immediately
	var stdout2, stderr2 bytes.Buffer
	exitCode2 := Mailman(&stdout2, &stderr2, MailmanOptions{
		Daemonize: false,
		RepoRoot:  tmpDir,
	})

	if exitCode2 != 2 {
		t.Errorf("Second daemon should exit with code 2, got %d", exitCode2)
	}

	// Verify error message about already running
	errOutput := stderr2.String()
	if !strings.Contains(errOutput, "already running") {
		t.Errorf("Expected error about daemon already running, got: %s", errOutput)
	}

	// Verify error message includes PID
	if !strings.Contains(errOutput, "PID:") {
		t.Errorf("Expected error message to include PID, got: %s", errOutput)
	}

	// Stop first daemon
	close(stopCh)
	<-done1

	// Verify first daemon started successfully (safe to read after stopped)
	if !strings.Contains(stdout1.String(), "Mailman daemon started") {
		t.Errorf("First daemon should have started, got stdout: %s", stdout1.String())
	}

	if exitCode1 != 0 {
		t.Errorf("First daemon should exit with code 0, got %d", exitCode1)
	}
}

// Note: Background mode (--daemon) is tested differently because it spawns
// a separate process. These tests focus on the foreground mode behavior
// which can be tested directly in unit tests.
//
// Full integration tests for daemon mode should be done externally
// by actually running the binary and checking the behavior.

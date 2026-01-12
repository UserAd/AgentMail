package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

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

	// Create .git/mail directory
	mailDir := filepath.Join(tmpDir, ".git", "mail")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	exitCode := Mailman(&stdout, &stderr, MailmanOptions{
		Daemonize: false,
		RepoRoot:  tmpDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Verify startup message format
	output := stdout.String()
	currentPID := os.Getpid()
	expected := "Mailman daemon started (PID: " + strconv.Itoa(currentPID) + ")\n"

	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}

	// Verify PID file was created
	pidFile := filepath.Join(mailDir, "mailman.pid")
	content, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("PID file should exist: %v", err)
	}

	expectedPID := strconv.Itoa(currentPID) + "\n"
	if string(content) != expectedPID {
		t.Errorf("PID file content mismatch. Expected %q, got %q", expectedPID, string(content))
	}
}

func TestMailman_AlreadyRunning_ExitsCode2(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-mailman-test-*")
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
	currentPID := os.Getpid()
	pidFile := filepath.Join(mailDir, "mailman.pid")
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

	// Create .git/mail directory
	mailDir := filepath.Join(tmpDir, ".git", "mail")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	// Create PID file with a non-existent PID (stale)
	stalePID := 99999999
	pidFile := filepath.Join(mailDir, "mailman.pid")
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(stalePID)+"\n"), 0644); err != nil {
		t.Fatalf("Failed to create PID file: %v", err)
	}

	var stdout, stderr bytes.Buffer

	exitCode := Mailman(&stdout, &stderr, MailmanOptions{
		Daemonize: false,
		RepoRoot:  tmpDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for stale PID, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Verify PID file was overwritten with current PID
	content, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("PID file should exist: %v", err)
	}

	currentPID := os.Getpid()
	expectedPID := strconv.Itoa(currentPID) + "\n"
	if string(content) != expectedPID {
		t.Errorf("PID file should be overwritten. Expected %q, got %q", expectedPID, string(content))
	}
}

func TestMailman_NoPIDFile_StartsSuccessfully(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-mailman-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git/mail directory (no PID file)
	mailDir := filepath.Join(tmpDir, ".git", "mail")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	exitCode := Mailman(&stdout, &stderr, MailmanOptions{
		Daemonize: false,
		RepoRoot:  tmpDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Verify PID file was created
	pidFile := filepath.Join(mailDir, "mailman.pid")
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		t.Error("PID file should be created")
	}
}

func TestMailman_CreatesMailDirIfNeeded(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-mailman-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create only .git directory (no mail subdirectory)
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	exitCode := Mailman(&stdout, &stderr, MailmanOptions{
		Daemonize: false,
		RepoRoot:  tmpDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Verify .git/mail directory was created
	mailDir := filepath.Join(tmpDir, ".git", "mail")
	info, err := os.Stat(mailDir)
	if err != nil {
		t.Fatalf("Mail directory should be created: %v", err)
	}
	if !info.IsDir() {
		t.Error("Mail directory should be a directory")
	}

	// Verify PID file was created
	pidFile := filepath.Join(mailDir, "mailman.pid")
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		t.Error("PID file should be created")
	}
}

// Note: Background mode (--daemon) is tested differently because it spawns
// a separate process. These tests focus on the foreground mode behavior
// which can be tested directly in unit tests.
//
// Full integration tests for daemon mode should be done externally
// by actually running the binary and checking the behavior.

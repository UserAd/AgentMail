package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agentmail/internal/daemon"
)

// =============================================================================
// T003-T004: Tests for MailmanStop - successful stop file creation
// =============================================================================

func TestMailmanStop_CreatesStopFile(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-stop-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .agentmail directory (simulating existing daemon setup)
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.MkdirAll(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// Run stop command
	exitCode := MailmanStop(&stdout, &stderr, MailmanStopOptions{
		RepoRoot: tmpDir,
	})

	// Verify exit code 0
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. stderr: %s", exitCode, stderr.String())
	}

	// Verify stop file was created
	stopPath := daemon.StopFilePath(tmpDir)
	if _, err := os.Stat(stopPath); os.IsNotExist(err) {
		t.Error("Stop file was not created")
	}
}

func TestMailmanStop_OutputsSuccessMessage(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-stop-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.MkdirAll(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// Run stop command
	exitCode := MailmanStop(&stdout, &stderr, MailmanStopOptions{
		RepoRoot: tmpDir,
	})

	// Verify exit code 0
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify success message
	expectedMsg := "Stop signal sent\n"
	if stdout.String() != expectedMsg {
		t.Errorf("Expected stdout %q, got %q", expectedMsg, stdout.String())
	}

	// Verify no stderr output
	if stderr.String() != "" {
		t.Errorf("Expected empty stderr, got %q", stderr.String())
	}
}

// =============================================================================
// T015-T016: Tests for MailmanStop - stop already pending
// =============================================================================

func TestMailmanStop_StopAlreadyPending_ReturnsError(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-stop-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.MkdirAll(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	// Pre-create the stop file to simulate pending stop
	stopPath := daemon.StopFilePath(tmpDir)
	if err := os.WriteFile(stopPath, []byte{}, 0600); err != nil {
		t.Fatalf("Failed to create stop file: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// Run stop command
	exitCode := MailmanStop(&stdout, &stderr, MailmanStopOptions{
		RepoRoot: tmpDir,
	})

	// Verify exit code 1
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	// Verify error message
	expectedMsg := "Stop already pending\n"
	if stderr.String() != expectedMsg {
		t.Errorf("Expected stderr %q, got %q", expectedMsg, stderr.String())
	}

	// Verify no stdout output
	if stdout.String() != "" {
		t.Errorf("Expected empty stdout, got %q", stdout.String())
	}
}

func TestMailmanStop_FilesystemError_ReturnsError(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-stop-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Do NOT create .agentmail directory - this should cause a filesystem error
	// when trying to create the stop file

	var stdout, stderr bytes.Buffer

	// Run stop command
	exitCode := MailmanStop(&stdout, &stderr, MailmanStopOptions{
		RepoRoot: tmpDir,
	})

	// Verify exit code 1
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	// Verify error message contains expected prefix
	expectedPrefix := "Failed to send stop signal:"
	if !strings.HasPrefix(stderr.String(), expectedPrefix) {
		t.Errorf("Expected stderr to start with %q, got %q", expectedPrefix, stderr.String())
	}

	// Verify no stdout output
	if stdout.String() != "" {
		t.Errorf("Expected empty stdout, got %q", stdout.String())
	}
}

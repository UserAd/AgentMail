package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// T015: Tests for send command argument validation

func TestSendCommand_MissingRecipient(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Send([]string{}, &stdout, &stderr, SendOptions{
		// Skip tmux check for this test
		SkipTmuxCheck: true,
	})

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	if stderr.String() == "" {
		t.Error("Expected error message in stderr")
	}
}

func TestSendCommand_MissingMessage(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Send([]string{"agent-2"}, &stdout, &stderr, SendOptions{
		SkipTmuxCheck: true,
	})

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	if stderr.String() == "" {
		t.Error("Expected error message in stderr")
	}
}

// T016: Tests for send command recipient validation

func TestSendCommand_RecipientNotFound(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Send([]string{"nonexistent", "Hello"}, &stdout, &stderr, SendOptions{
		SkipTmuxCheck: true,
		// Mock window list without the recipient
		MockWindows: []string{"agent-1", "agent-3"},
		MockSender:  "agent-1",
	})

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for nonexistent recipient, got %d", exitCode)
	}

	stderrStr := stderr.String()
	if stderrStr != "error: recipient not found\n" {
		t.Errorf("Expected 'error: recipient not found\\n', got: %q", stderrStr)
	}
}

// T017: Tests for send command success path (ID output)

func TestSendCommand_Success(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
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

	exitCode := Send([]string{"agent-2", "Hello from agent-1"}, &stdout, &stderr, SendOptions{
		SkipTmuxCheck: true,
		MockWindows:   []string{"agent-1", "agent-2"},
		MockSender:    "agent-1",
		RepoRoot:      tmpDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Should output "Message #ID sent"
	output := stdout.String()
	if !strings.HasPrefix(output, "Message #") {
		t.Errorf("Expected output to start with 'Message #', got: %s", output)
	}
	if !strings.HasSuffix(output, " sent\n") {
		t.Errorf("Expected output to end with ' sent', got: %s", output)
	}

	// Verify file was created
	filePath := filepath.Join(mailDir, "agent-2.jsonl")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Message file should have been created")
	}
}

func TestSendCommand_NotInTmux(t *testing.T) {
	// Save and restore TMUX env var
	original := os.Getenv("TMUX")
	defer os.Setenv("TMUX", original)
	os.Unsetenv("TMUX")

	var stdout, stderr bytes.Buffer

	exitCode := Send([]string{"agent-2", "Hello"}, &stdout, &stderr, SendOptions{
		// Don't skip tmux check - we want to test it
		SkipTmuxCheck: false,
	})

	if exitCode != 2 {
		t.Errorf("Expected exit code 2 (not in tmux), got %d", exitCode)
	}
}

// T026: Test sending to ignored recipient returns "recipient not found" error
func TestSendCommand_IgnoredRecipient(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
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

	// Send to agent-2 which is in the ignore list
	exitCode := Send([]string{"agent-2", "Hello from agent-1"}, &stdout, &stderr, SendOptions{
		SkipTmuxCheck:  true,
		MockWindows:    []string{"agent-1", "agent-2", "agent-3"},
		MockSender:     "agent-1",
		MockIgnoreList: map[string]bool{"agent-2": true, "monitor": true},
		RepoRoot:       tmpDir,
	})

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	stderrStr := stderr.String()
	if stderrStr != "error: recipient not found\n" {
		t.Errorf("Expected 'error: recipient not found\\n', got: %q", stderrStr)
	}

	if stdout.String() != "" {
		t.Errorf("Expected empty stdout, got: %s", stdout.String())
	}

	// Verify no file was created
	filePath := filepath.Join(mailDir, "agent-2.jsonl")
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("Message file should NOT have been created for ignored recipient")
	}
}

// T027: Test sending to valid (non-ignored) recipient succeeds
func TestSendCommand_ValidRecipient(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
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

	// Send to agent-3 which is NOT in the ignore list
	exitCode := Send([]string{"agent-3", "Hello from agent-1"}, &stdout, &stderr, SendOptions{
		SkipTmuxCheck:  true,
		MockWindows:    []string{"agent-1", "agent-2", "agent-3"},
		MockSender:     "agent-1",
		MockIgnoreList: map[string]bool{"agent-2": true, "monitor": true},
		RepoRoot:       tmpDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Should output "Message #ID sent"
	output := stdout.String()
	if !strings.HasPrefix(output, "Message #") {
		t.Errorf("Expected output to start with 'Message #', got: %s", output)
	}
	if !strings.HasSuffix(output, " sent\n") {
		t.Errorf("Expected output to end with ' sent', got: %s", output)
	}

	// Verify file was created
	filePath := filepath.Join(mailDir, "agent-3.jsonl")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Message file should have been created")
	}
}

// T028: Test sending to self returns "recipient not found" error
func TestSendCommand_SendToSelf(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
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

	// Send to self (agent-1 sending to agent-1)
	exitCode := Send([]string{"agent-1", "Hello to myself"}, &stdout, &stderr, SendOptions{
		SkipTmuxCheck: true,
		MockWindows:   []string{"agent-1", "agent-2", "agent-3"},
		MockSender:    "agent-1",
		RepoRoot:      tmpDir,
	})

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	stderrStr := stderr.String()
	if stderrStr != "error: recipient not found\n" {
		t.Errorf("Expected 'error: recipient not found\\n', got: %q", stderrStr)
	}

	if stdout.String() != "" {
		t.Errorf("Expected empty stdout, got: %s", stdout.String())
	}

	// Verify no file was created
	filePath := filepath.Join(mailDir, "agent-1.jsonl")
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("Message file should NOT have been created for self-send")
	}
}

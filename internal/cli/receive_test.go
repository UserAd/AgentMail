package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// T028: Tests for receive command no-messages case

func TestReceiveCommand_NoMessages(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git/mail directory but no file
	mailDir := filepath.Join(tmpDir, ".git", "mail")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	exitCode := Receive(&stdout, &stderr, ReceiveOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "agent-2",
		MockWindows:   []string{"agent-1", "agent-2"},
		RepoRoot:      tmpDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "No unread messages") {
		t.Errorf("Expected 'No unread messages' in output, got: %s", output)
	}
}

func TestReceiveCommand_AllMessagesRead(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git/mail directory and file with all messages read
	mailDir := filepath.Join(tmpDir, ".git", "mail")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	content := `{"id":"id1","from":"agent-1","to":"agent-2","message":"Hello","read_flag":true}
`
	filePath := filepath.Join(mailDir, "agent-2.jsonl")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	var stdout, stderr bytes.Buffer

	exitCode := Receive(&stdout, &stderr, ReceiveOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "agent-2",
		MockWindows:   []string{"agent-1", "agent-2"},
		RepoRoot:      tmpDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "No unread messages") {
		t.Errorf("Expected 'No unread messages' in output, got: %s", output)
	}
}

// T029: Tests for receive command success path

func TestReceiveCommand_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git/mail directory and file
	mailDir := filepath.Join(tmpDir, ".git", "mail")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	content := `{"id":"testID01","from":"agent-1","to":"agent-2","message":"Hello from agent-1!","read_flag":false}
`
	filePath := filepath.Join(mailDir, "agent-2.jsonl")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	var stdout, stderr bytes.Buffer

	exitCode := Receive(&stdout, &stderr, ReceiveOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "agent-2",
		MockWindows:   []string{"agent-1", "agent-2"},
		RepoRoot:      tmpDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	output := stdout.String()

	// Check format: From: <sender>\nID: <id>\n\n<message>
	if !strings.Contains(output, "From: agent-1") {
		t.Errorf("Expected 'From: agent-1' in output, got: %s", output)
	}
	if !strings.Contains(output, "ID: testID01") {
		t.Errorf("Expected 'ID: testID01' in output, got: %s", output)
	}
	if !strings.Contains(output, "Hello from agent-1!") {
		t.Errorf("Expected message body in output, got: %s", output)
	}

	// Verify message was marked as read
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if !strings.Contains(string(fileContent), `"read_flag":true`) {
		t.Errorf("Message should be marked as read after receive")
	}
}

func TestReceiveCommand_FIFO_Order(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git/mail directory and file with multiple messages
	mailDir := filepath.Join(tmpDir, ".git", "mail")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	// Three unread messages
	content := `{"id":"first123","from":"agent-1","to":"agent-2","message":"First message","read_flag":false}
{"id":"second12","from":"agent-3","to":"agent-2","message":"Second message","read_flag":false}
{"id":"third123","from":"agent-1","to":"agent-2","message":"Third message","read_flag":false}
`
	filePath := filepath.Join(mailDir, "agent-2.jsonl")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	var stdout, stderr bytes.Buffer

	exitCode := Receive(&stdout, &stderr, ReceiveOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "agent-2",
		MockWindows:   []string{"agent-1", "agent-2", "agent-3"},
		RepoRoot:      tmpDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Should receive oldest (first) message
	output := stdout.String()
	if !strings.Contains(output, "ID: first123") {
		t.Errorf("Expected first message (FIFO), got: %s", output)
	}
	if !strings.Contains(output, "First message") {
		t.Errorf("Expected first message body, got: %s", output)
	}
}

func TestReceiveCommand_NotInTmux(t *testing.T) {
	// Save and restore TMUX env var
	original := os.Getenv("TMUX")
	defer os.Setenv("TMUX", original)
	os.Unsetenv("TMUX")

	var stdout, stderr bytes.Buffer

	exitCode := Receive(&stdout, &stderr, ReceiveOptions{
		SkipTmuxCheck: false,
	})

	if exitCode != 2 {
		t.Errorf("Expected exit code 2 (not in tmux), got %d", exitCode)
	}
}

func TestReceiveCommand_CurrentWindowNotInSession(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mailDir := filepath.Join(tmpDir, ".git", "mail")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	exitCode := Receive(&stdout, &stderr, ReceiveOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "orphan-window",
		MockWindows:   []string{"agent-1", "agent-2"}, // orphan-window not in list
		RepoRoot:      tmpDir,
	})

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 (window not in session), got %d", exitCode)
	}

	if !strings.Contains(stderr.String(), "not found in tmux session") {
		t.Errorf("Expected error about window not found, got: %s", stderr.String())
	}
}

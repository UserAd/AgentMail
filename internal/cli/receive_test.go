package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agentmail/internal/mail"
)

// T028: Tests for receive command no-messages case

func TestReceiveCommand_NoMessages(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .agentmail/mailboxes directory but no file
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
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

	// Create .agentmail/mailboxes directory and file with all messages read
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
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

	// Create .agentmail/mailboxes directory and file
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
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

	// Create .agentmail/mailboxes directory and file with multiple messages
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
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
	// Use t.Setenv for thread-safe environment variable manipulation
	t.Setenv("TMUX", "")

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

	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
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

// Hook mode tests (FR-001 through FR-005)

func TestReceiveCommand_HookMode_WithMessages(t *testing.T) {
	// FR-001a/b/c: Hook mode with messages should output to STDERR and exit 2
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	content := `{"id":"hookID01","from":"agent-1","to":"agent-2","message":"Hook test message","read_flag":false}
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
		HookMode:      true,
	})

	// FR-001b: Exit code should be 2
	if exitCode != 2 {
		t.Errorf("Hook mode with messages: expected exit code 2, got %d", exitCode)
	}

	// FR-005: Output should be on STDERR, not STDOUT
	if stdout.Len() > 0 {
		t.Errorf("Hook mode should not output to STDOUT, got: %s", stdout.String())
	}

	stderrOutput := stderr.String()

	// FR-001a: Should have "You got new mail" prefix
	if !strings.HasPrefix(stderrOutput, "You got new mail\n") {
		t.Errorf("Hook mode output should start with 'You got new mail\\n', got: %s", stderrOutput)
	}

	// Should contain message details
	if !strings.Contains(stderrOutput, "From: agent-1") {
		t.Errorf("Expected 'From: agent-1' in STDERR, got: %s", stderrOutput)
	}
	if !strings.Contains(stderrOutput, "ID: hookID01") {
		t.Errorf("Expected 'ID: hookID01' in STDERR, got: %s", stderrOutput)
	}
	if !strings.Contains(stderrOutput, "Hook test message") {
		t.Errorf("Expected message body in STDERR, got: %s", stderrOutput)
	}

	// FR-001c: Message should be marked as read
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if !strings.Contains(string(fileContent), `"read_flag":true`) {
		t.Errorf("Hook mode should mark message as read")
	}
}

func TestReceiveCommand_HookMode_NoMessages(t *testing.T) {
	// FR-002: Hook mode with no messages should exit silently with code 0
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	exitCode := Receive(&stdout, &stderr, ReceiveOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "agent-2",
		MockWindows:   []string{"agent-1", "agent-2"},
		RepoRoot:      tmpDir,
		HookMode:      true,
	})

	// FR-002: Exit code should be 0
	if exitCode != 0 {
		t.Errorf("Hook mode with no messages: expected exit code 0, got %d", exitCode)
	}

	// FR-002: Should produce no output
	if stdout.Len() > 0 {
		t.Errorf("Hook mode with no messages should not output to STDOUT, got: %s", stdout.String())
	}
	if stderr.Len() > 0 {
		t.Errorf("Hook mode with no messages should not output to STDERR, got: %s", stderr.String())
	}
}

func TestReceiveCommand_HookMode_NotInTmux(t *testing.T) {
	// FR-003: Hook mode not in tmux should exit silently with code 0
	t.Setenv("TMUX", "")

	var stdout, stderr bytes.Buffer

	exitCode := Receive(&stdout, &stderr, ReceiveOptions{
		SkipTmuxCheck: false,
		HookMode:      true,
	})

	// FR-003: Exit code should be 0 (not 2 like normal mode)
	if exitCode != 0 {
		t.Errorf("Hook mode not in tmux: expected exit code 0, got %d", exitCode)
	}

	// FR-003: Should produce no output
	if stdout.Len() > 0 {
		t.Errorf("Hook mode not in tmux should not output to STDOUT, got: %s", stdout.String())
	}
	if stderr.Len() > 0 {
		t.Errorf("Hook mode not in tmux should not output to STDERR, got: %s", stderr.String())
	}
}

func TestReceiveCommand_HookMode_WindowNotInSession(t *testing.T) {
	// FR-004a: Hook mode with error should exit silently with code 0
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	exitCode := Receive(&stdout, &stderr, ReceiveOptions{
		SkipTmuxCheck: true,
		MockReceiver:  "orphan-window",
		MockWindows:   []string{"agent-1", "agent-2"}, // orphan-window not in list
		RepoRoot:      tmpDir,
		HookMode:      true,
	})

	// FR-004a: Exit code should be 0 (silent error)
	if exitCode != 0 {
		t.Errorf("Hook mode with error: expected exit code 0, got %d", exitCode)
	}

	// FR-004a: Should produce no output
	if stdout.Len() > 0 {
		t.Errorf("Hook mode with error should not output to STDOUT, got: %s", stdout.String())
	}
	if stderr.Len() > 0 {
		t.Errorf("Hook mode with error should not output to STDERR, got: %s", stderr.String())
	}
}

func TestReceiveCommand_HookMode_AllMessagesRead(t *testing.T) {
	// Hook mode with all messages read should behave like no messages
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
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
		HookMode:      true,
	})

	// Exit code should be 0 (no unread messages)
	if exitCode != 0 {
		t.Errorf("Hook mode with all read: expected exit code 0, got %d", exitCode)
	}

	// Should produce no output
	if stdout.Len() > 0 || stderr.Len() > 0 {
		t.Errorf("Hook mode with all read should produce no output")
	}
}

func TestReceiveCommand_NormalMode_Unchanged(t *testing.T) {
	// Regression test: ensure normal mode behavior is unchanged
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	content := `{"id":"normalID","from":"agent-1","to":"agent-2","message":"Normal test","read_flag":false}
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
		HookMode:      false, // Explicitly false
	})

	// Normal mode should exit 0 with message
	if exitCode != 0 {
		t.Errorf("Normal mode: expected exit code 0, got %d", exitCode)
	}

	// Output should be on STDOUT
	if stdout.Len() == 0 {
		t.Errorf("Normal mode should output to STDOUT")
	}
	if stderr.Len() > 0 {
		t.Errorf("Normal mode should not output to STDERR, got: %s", stderr.String())
	}

	// Should NOT have "You got new mail" prefix
	if strings.Contains(stdout.String(), "You got new mail") {
		t.Errorf("Normal mode should not have hook prefix")
	}
}

// T008: Tests for pane address support in receive

func TestReceiveCommand_PaneSpecificMailbox(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create message for specific pane
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	content := `{"id":"testID1","from":"mysession:sender.0","to":"mysession:editor.1","message":"Hello pane 1","read_flag":false}
`
	// Use sanitized filename
	mailboxFile := filepath.Join(mailDir, "mysession%3Aeditor%2E1.jsonl")
	if err := os.WriteFile(mailboxFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write mailbox file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	exitCode := Receive(&stdout, &stderr, ReceiveOptions{
		SkipTmuxCheck:   true,
		MockPaneAddress: "mysession:editor.1",
		RepoRoot:        tmpDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	if !strings.Contains(stdout.String(), "Hello pane 1") {
		t.Errorf("Expected message content, got: %s", stdout.String())
	}
}

// T012: Hook mode pane isolation test

func TestReceiveCommand_HookMode_PaneIsolation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	// Create message for pane 0 (sibling pane)
	pane0Content := `{"id":"msg0","from":"mysession:sender.0","to":"mysession:editor.0","message":"Message for pane 0","read_flag":false}
`
	pane0File := filepath.Join(mailDir, "mysession%3Aeditor%2E0.jsonl")
	if err := os.WriteFile(pane0File, []byte(pane0Content), 0644); err != nil {
		t.Fatalf("Failed to write pane 0 mailbox: %v", err)
	}

	// Receive from pane 1 (different pane in same window)
	var stdout, stderr bytes.Buffer
	exitCode := Receive(&stdout, &stderr, ReceiveOptions{
		SkipTmuxCheck:   true,
		MockPaneAddress: "mysession:editor.1",
		RepoRoot:        tmpDir,
		HookMode:        true,
	})

	// Should exit 0 with no output (no messages for pane 1)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 (no messages for this pane), got %d", exitCode)
	}

	if stdout.String() != "" {
		t.Errorf("Expected no stdout output, got: %s", stdout.String())
	}

	if stderr.String() != "" {
		t.Errorf("Expected no stderr output (pane isolation), got: %s", stderr.String())
	}
}

func TestReceiveCommand_PaneIsolation_NoMessages(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	// Create message for a different pane in the same window
	otherPaneContent := `{"id":"msg1","from":"mysession:sender.0","to":"mysession:editor.0","message":"For pane 0","read_flag":false}
`
	otherPaneFile := filepath.Join(mailDir, "mysession%3Aeditor%2E0.jsonl")
	if err := os.WriteFile(otherPaneFile, []byte(otherPaneContent), 0644); err != nil {
		t.Fatalf("Failed to write mailbox: %v", err)
	}

	// Try to receive from pane 1
	var stdout, stderr bytes.Buffer
	exitCode := Receive(&stdout, &stderr, ReceiveOptions{
		SkipTmuxCheck:   true,
		MockPaneAddress: "mysession:editor.1",
		RepoRoot:        tmpDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout.String(), "No unread messages") {
		t.Errorf("Expected 'No unread messages', got: %s", stdout.String())
	}
}

func TestReceiveCommand_NoRepoRoot(t *testing.T) {
	// Change to a temp directory without .git so FindGitRoot() fails
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	defer os.Chdir(origDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	var stdout, stderr bytes.Buffer
	exitCode := Receive(
		&stdout,
		&stderr,
		ReceiveOptions{
			SkipTmuxCheck:   true,
			MockPaneAddress: "mysession:receiver.0",
			RepoRoot:        "", // Force it to search for git root
		},
	)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for no repo root, got %d", exitCode)
	}

	if !strings.Contains(stderr.String(), "not in a git repository") {
		t.Errorf("Expected 'not in a git repository' error, got: %s", stderr.String())
	}
}

func TestReceiveCommand_MultipleUnreadFIFO(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mailbox with multiple unread messages
	messages := []mail.Message{
		{ID: "first123", From: "mysession:sender.0", To: "mysession:receiver.0", Message: "First message", ReadFlag: false, CreatedAt: time.Now().Add(-3 * time.Hour)},
		{ID: "second12", From: "mysession:sender.0", To: "mysession:receiver.0", Message: "Second message", ReadFlag: false, CreatedAt: time.Now().Add(-2 * time.Hour)},
		{ID: "third123", From: "mysession:sender.0", To: "mysession:receiver.0", Message: "Third message", ReadFlag: false, CreatedAt: time.Now().Add(-1 * time.Hour)},
	}

	if err := mail.WriteAll(tmpDir, "mysession:receiver.0", messages); err != nil {
		t.Fatalf("Failed to create test messages: %v", err)
	}

	var stdout, stderr bytes.Buffer
	exitCode := Receive(
		&stdout,
		&stderr,
		ReceiveOptions{
			SkipTmuxCheck:   true,
			MockPaneAddress: "mysession:receiver.0",
			RepoRoot:        tmpDir,
		},
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Should receive oldest message first (FIFO)
	if !strings.Contains(stdout.String(), "First message") {
		t.Errorf("Expected 'First message' (oldest), got: %s", stdout.String())
	}

	if strings.Contains(stdout.String(), "Second message") || strings.Contains(stdout.String(), "Third message") {
		t.Errorf("Should only receive oldest message, got: %s", stdout.String())
	}
}

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

	exitCode := Send([]string{}, nil, &stdout, &stderr, SendOptions{
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

	exitCode := Send([]string{"agent-2"}, nil, &stdout, &stderr, SendOptions{
		SkipTmuxCheck: true,
		MockWindows:   []string{"agent-1", "agent-2"},
		MockSender:    "agent-1",
	})

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	stderrStr := stderr.String()
	expectedErr := "error: no message provided\nusage: agentmail send <recipient> <message>\n"
	if stderrStr != expectedErr {
		t.Errorf("Expected stderr %q, got: %q", expectedErr, stderrStr)
	}
}

// T016: Tests for send command recipient validation

func TestSendCommand_RecipientNotFound(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Send([]string{"nonexistent", "Hello"}, nil, &stdout, &stderr, SendOptions{
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

	// Create .agentmail/mailboxes directory
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	exitCode := Send([]string{"agent-2", "Hello from agent-1"}, nil, &stdout, &stderr, SendOptions{
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
	// Use t.Setenv for thread-safe environment variable manipulation
	t.Setenv("TMUX", "")

	var stdout, stderr bytes.Buffer

	exitCode := Send([]string{"agent-2", "Hello"}, nil, &stdout, &stderr, SendOptions{
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

	// Create .agentmail/mailboxes directory
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// Send to agent-2 which is in the ignore list
	exitCode := Send([]string{"agent-2", "Hello from agent-1"}, nil, &stdout, &stderr, SendOptions{
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

	// Create .agentmail/mailboxes directory
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// Send to agent-3 which is NOT in the ignore list
	exitCode := Send([]string{"agent-3", "Hello from agent-1"}, nil, &stdout, &stderr, SendOptions{
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

	// Create .agentmail/mailboxes directory
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// Send to self (agent-1 sending to agent-1)
	exitCode := Send([]string{"agent-1", "Hello to myself"}, nil, &stdout, &stderr, SendOptions{
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

// T041: Test message from stdin is used when piped
func TestSendCommand_StdinMessage(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .agentmail/mailboxes directory
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// Only recipient argument, message comes from stdin
	exitCode := Send([]string{"agent-2"}, nil, &stdout, &stderr, SendOptions{
		SkipTmuxCheck: true,
		MockWindows:   []string{"agent-1", "agent-2"},
		MockSender:    "agent-1",
		RepoRoot:      tmpDir,
		StdinIsPipe:   true,
		StdinContent:  "Hello from stdin\n",
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

	// Read the file and verify the message content
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read message file: %v", err)
	}

	if !strings.Contains(string(data), "Hello from stdin") {
		t.Errorf("Message file should contain 'Hello from stdin', got: %s", string(data))
	}
}

// T042: Test multi-line stdin content sent as single message
func TestSendCommand_MultiLineStdin(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .agentmail/mailboxes directory
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	multiLineMsg := "Line 1\nLine 2\nLine 3\n"

	exitCode := Send([]string{"agent-2"}, nil, &stdout, &stderr, SendOptions{
		SkipTmuxCheck: true,
		MockWindows:   []string{"agent-1", "agent-2"},
		MockSender:    "agent-1",
		RepoRoot:      tmpDir,
		StdinIsPipe:   true,
		StdinContent:  multiLineMsg,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Verify file was created and contains multi-line content
	filePath := filepath.Join(mailDir, "agent-2.jsonl")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read message file: %v", err)
	}

	// The message should contain the multi-line content (trailing newline stripped)
	if !strings.Contains(string(data), "Line 1\\nLine 2\\nLine 3") {
		t.Errorf("Message file should contain multi-line content, got: %s", string(data))
	}
}

// T043: Test stdin takes precedence over argument (per FR-010)
func TestSendCommand_StdinPrecedence(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .agentmail/mailboxes directory
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// Both argument and stdin provided - stdin should take precedence
	exitCode := Send([]string{"agent-2", "Argument message"}, nil, &stdout, &stderr, SendOptions{
		SkipTmuxCheck: true,
		MockWindows:   []string{"agent-1", "agent-2"},
		MockSender:    "agent-1",
		RepoRoot:      tmpDir,
		StdinIsPipe:   true,
		StdinContent:  "Stdin message\n",
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Verify file contains stdin message, not argument
	filePath := filepath.Join(mailDir, "agent-2.jsonl")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read message file: %v", err)
	}

	dataStr := string(data)
	if !strings.Contains(dataStr, "Stdin message") {
		t.Errorf("Message file should contain 'Stdin message', got: %s", dataStr)
	}
	if strings.Contains(dataStr, "Argument message") {
		t.Errorf("Message file should NOT contain 'Argument message', got: %s", dataStr)
	}
}

// T044: Test falls back to argument when no stdin data
func TestSendCommand_FallbackToArgument(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .agentmail/mailboxes directory
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// Stdin is a pipe but empty - should fall back to argument
	exitCode := Send([]string{"agent-2", "Argument message"}, nil, &stdout, &stderr, SendOptions{
		SkipTmuxCheck: true,
		MockWindows:   []string{"agent-1", "agent-2"},
		MockSender:    "agent-1",
		RepoRoot:      tmpDir,
		StdinIsPipe:   true,
		StdinContent:  "", // Empty stdin
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Verify file contains argument message
	filePath := filepath.Join(mailDir, "agent-2.jsonl")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read message file: %v", err)
	}

	if !strings.Contains(string(data), "Argument message") {
		t.Errorf("Message file should contain 'Argument message', got: %s", string(data))
	}
}

// =============================================================================
// US6 Backward Compatibility Regression Tests (Phase 8)
// =============================================================================

// T051: Regression test - argument-based send still works after stdin changes
// This explicitly verifies backward compatibility per US6: "existing argument-based
// send command continues to work unchanged"
func TestSendCommand_ArgumentBasedSend_Regression(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .agentmail/mailboxes directory
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// US6 Backward Compatibility: Original argument-based send must work
	// Command: agentmail send <recipient> <message>
	// No stdin involved (StdinIsPipe: false by default)
	exitCode := Send([]string{"agent-2", "Hello via argument"}, nil, &stdout, &stderr, SendOptions{
		SkipTmuxCheck: true,
		MockWindows:   []string{"agent-1", "agent-2"},
		MockSender:    "agent-1",
		RepoRoot:      tmpDir,
		// StdinIsPipe defaults to false - simulating terminal input (not a pipe)
	})

	if exitCode != 0 {
		t.Errorf("US6 Regression: Expected exit code 0 for argument-based send, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Should output "Message #ID sent"
	output := stdout.String()
	if !strings.HasPrefix(output, "Message #") {
		t.Errorf("US6 Regression: Expected output to start with 'Message #', got: %s", output)
	}
	if !strings.HasSuffix(output, " sent\n") {
		t.Errorf("US6 Regression: Expected output to end with ' sent', got: %s", output)
	}

	// Verify file was created with correct content
	filePath := filepath.Join(mailDir, "agent-2.jsonl")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("US6 Regression: Failed to read message file: %v", err)
	}

	if !strings.Contains(string(data), "Hello via argument") {
		t.Errorf("US6 Regression: Message file should contain 'Hello via argument', got: %s", string(data))
	}
}

// T052: Regression test - no message argument and no stdin returns usage error (FR-011)
// Verifies that when neither argument nor stdin provides a message, the proper
// error is returned per FR-011
func TestSendCommand_NoMessageProvided_Regression(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// FR-011: No message argument AND no stdin content should error
	// This simulates: agentmail send agent-2 (with no stdin pipe)
	exitCode := Send([]string{"agent-2"}, nil, &stdout, &stderr, SendOptions{
		SkipTmuxCheck: true,
		MockWindows:   []string{"agent-1", "agent-2"},
		MockSender:    "agent-1",
		// StdinIsPipe defaults to false - no stdin available
	})

	// Verify exit code 1
	if exitCode != 1 {
		t.Errorf("FR-011 Regression: Expected exit code 1 for missing message, got %d", exitCode)
	}

	// Verify exact error message format per FR-011
	expectedErr := "error: no message provided\nusage: agentmail send <recipient> <message>\n"
	stderrStr := stderr.String()
	if stderrStr != expectedErr {
		t.Errorf("FR-011 Regression: Expected stderr %q, got: %q", expectedErr, stderrStr)
	}

	// Verify no stdout output
	if stdout.String() != "" {
		t.Errorf("FR-011 Regression: Expected empty stdout, got: %s", stdout.String())
	}
}

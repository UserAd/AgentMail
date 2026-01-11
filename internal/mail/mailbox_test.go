package mail

import (
	"os"
	"path/filepath"
	"testing"
)

// T014: Tests for mailbox Append to recipient file

func TestEnsureMailDir_CreatesDirectory(t *testing.T) {
	// Create temp dir as git root
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git directory to simulate git repo
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	mailDir := filepath.Join(gitDir, "mail")

	// Verify mail dir doesn't exist yet
	if _, err := os.Stat(mailDir); !os.IsNotExist(err) {
		t.Fatal("Mail dir should not exist before test")
	}

	// Call EnsureMailDir
	err = EnsureMailDir(tmpDir)
	if err != nil {
		t.Fatalf("EnsureMailDir failed: %v", err)
	}

	// Verify mail dir now exists
	info, err := os.Stat(mailDir)
	if err != nil {
		t.Fatalf("Mail dir should exist after EnsureMailDir: %v", err)
	}
	if !info.IsDir() {
		t.Error("Mail dir should be a directory")
	}
}

func TestEnsureMailDir_ExistingDirectory(t *testing.T) {
	// Create temp dir with existing mail directory
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

	// Should not error on existing directory
	err = EnsureMailDir(tmpDir)
	if err != nil {
		t.Errorf("EnsureMailDir should not error on existing dir: %v", err)
	}
}

func TestAppend_CreatesFileAndWritesMessage(t *testing.T) {
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

	msg := Message{
		ID:       "testID01",
		From:     "agent-1",
		To:       "agent-2",
		Message:  "Hello",
		ReadFlag: false,
	}

	err = Append(tmpDir, msg)
	if err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	// Verify file exists with correct content
	filePath := filepath.Join(mailDir, "agent-2.jsonl")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expected := `{"id":"testID01","from":"agent-1","to":"agent-2","message":"Hello","read_flag":false}` + "\n"
	if string(content) != expected {
		t.Errorf("File content mismatch.\nExpected: %s\nGot: %s", expected, string(content))
	}
}

func TestAppend_AppendsToExistingFile(t *testing.T) {
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

	// Append first message
	msg1 := Message{
		ID:       "firstID1",
		From:     "agent-1",
		To:       "agent-2",
		Message:  "First message",
		ReadFlag: false,
	}
	if err := Append(tmpDir, msg1); err != nil {
		t.Fatalf("First Append failed: %v", err)
	}

	// Append second message
	msg2 := Message{
		ID:       "secndID2",
		From:     "agent-3",
		To:       "agent-2",
		Message:  "Second message",
		ReadFlag: false,
	}
	if err := Append(tmpDir, msg2); err != nil {
		t.Fatalf("Second Append failed: %v", err)
	}

	// Verify both messages in file
	filePath := filepath.Join(mailDir, "agent-2.jsonl")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	lines := string(content)
	expectedLine1 := `{"id":"firstID1","from":"agent-1","to":"agent-2","message":"First message","read_flag":false}`
	expectedLine2 := `{"id":"secndID2","from":"agent-3","to":"agent-2","message":"Second message","read_flag":false}`

	if lines != expectedLine1+"\n"+expectedLine2+"\n" {
		t.Errorf("File content mismatch.\nExpected:\n%s\n%s\nGot:\n%s", expectedLine1, expectedLine2, lines)
	}
}

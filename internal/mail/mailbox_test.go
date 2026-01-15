package mail

import (
	"os"
	"path/filepath"
	"testing"
)

// T014: Tests for mailbox Append to recipient file

func TestEnsureMailDir_CreatesDirectory(t *testing.T) {
	// Create temp dir as root
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")

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

	// Create .agentmail/mailboxes directory
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
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

	// Create .agentmail/mailboxes directory
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
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

	// Create .agentmail/mailboxes directory
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
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

// T025: Tests for mailbox ReadAll from recipient file

func TestReadAll_ReadsMessagesFromFile(t *testing.T) {
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

	// Write test data
	content := `{"id":"id1","from":"agent-1","to":"agent-2","message":"Hello","read_flag":false}
{"id":"id2","from":"agent-3","to":"agent-2","message":"Hi there","read_flag":true}
`
	filePath := filepath.Join(mailDir, "agent-2.jsonl")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	messages, err := ReadAll(tmpDir, "agent-2")
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	if messages[0].ID != "id1" || messages[0].From != "agent-1" || messages[0].Message != "Hello" || messages[0].ReadFlag != false {
		t.Errorf("First message mismatch: %+v", messages[0])
	}

	if messages[1].ID != "id2" || messages[1].From != "agent-3" || messages[1].Message != "Hi there" || messages[1].ReadFlag != true {
		t.Errorf("Second message mismatch: %+v", messages[1])
	}
}

func TestReadAll_ReturnsEmptyForMissingFile(t *testing.T) {
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

	messages, err := ReadAll(tmpDir, "nonexistent")
	if err != nil {
		t.Fatalf("ReadAll should not error for missing file: %v", err)
	}

	if len(messages) != 0 {
		t.Errorf("Expected 0 messages for missing file, got %d", len(messages))
	}
}

// T026: Tests for mailbox FindUnread (filter by read_flag only)

func TestFindUnread_ReturnsUnreadMessages(t *testing.T) {
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

	// Write test data with mixed read flags
	content := `{"id":"id1","from":"agent-1","to":"agent-2","message":"Read message","read_flag":true}
{"id":"id2","from":"agent-3","to":"agent-2","message":"Unread message","read_flag":false}
{"id":"id3","from":"agent-1","to":"agent-2","message":"Another unread","read_flag":false}
`
	filePath := filepath.Join(mailDir, "agent-2.jsonl")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	unread, err := FindUnread(tmpDir, "agent-2")
	if err != nil {
		t.Fatalf("FindUnread failed: %v", err)
	}

	if len(unread) != 2 {
		t.Fatalf("Expected 2 unread messages, got %d", len(unread))
	}

	// Should be in FIFO order
	if unread[0].ID != "id2" {
		t.Errorf("First unread should be id2, got %s", unread[0].ID)
	}
	if unread[1].ID != "id3" {
		t.Errorf("Second unread should be id3, got %s", unread[1].ID)
	}
}

func TestFindUnread_ReturnsEmptyWhenAllRead(t *testing.T) {
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

	// Write test data with all messages read
	content := `{"id":"id1","from":"agent-1","to":"agent-2","message":"Read1","read_flag":true}
{"id":"id2","from":"agent-3","to":"agent-2","message":"Read2","read_flag":true}
`
	filePath := filepath.Join(mailDir, "agent-2.jsonl")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	unread, err := FindUnread(tmpDir, "agent-2")
	if err != nil {
		t.Fatalf("FindUnread failed: %v", err)
	}

	if len(unread) != 0 {
		t.Errorf("Expected 0 unread messages when all read, got %d", len(unread))
	}
}

// T027: Tests for mailbox MarkAsRead operation

func TestMarkAsRead_UpdatesMessageFlag(t *testing.T) {
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

	// Write test data
	content := `{"id":"id1","from":"agent-1","to":"agent-2","message":"Hello","read_flag":false}
{"id":"id2","from":"agent-3","to":"agent-2","message":"Hi","read_flag":false}
`
	filePath := filepath.Join(mailDir, "agent-2.jsonl")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Mark first message as read
	err = MarkAsRead(tmpDir, "agent-2", "id1")
	if err != nil {
		t.Fatalf("MarkAsRead failed: %v", err)
	}

	// Verify message was marked as read
	messages, err := ReadAll(tmpDir, "agent-2")
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	if !messages[0].ReadFlag {
		t.Error("First message should be marked as read")
	}
	if messages[1].ReadFlag {
		t.Error("Second message should still be unread")
	}
}

func TestMarkAsRead_NonexistentMessage(t *testing.T) {
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

	// Write test data
	content := `{"id":"id1","from":"agent-1","to":"agent-2","message":"Hello","read_flag":false}
`
	filePath := filepath.Join(mailDir, "agent-2.jsonl")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Try to mark nonexistent message
	err = MarkAsRead(tmpDir, "agent-2", "nonexistent")
	// Should not error, just not find anything to update
	if err != nil {
		t.Errorf("MarkAsRead should not error for nonexistent message: %v", err)
	}
}

// Tests for safePath security function (path traversal prevention)

func TestSafePath_ValidFilename(t *testing.T) {
	result, err := safePath("/base/dir", "file.jsonl")
	if err != nil {
		t.Errorf("safePath should accept valid filename: %v", err)
	}
	expected := filepath.Join("/base/dir", "file.jsonl")
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestSafePath_DirectoryTraversal(t *testing.T) {
	tests := []struct {
		name     string
		baseDir  string
		filename string
	}{
		{"simple traversal", "/base/dir", "../etc/passwd"},
		{"double traversal", "/base/dir", "../../etc/passwd"},
		{"triple traversal", "/base/dir", "../../../etc/passwd"},
		{"hidden traversal", "/base/dir", "foo/../../../etc/passwd"},
		{"traversal with extension", "/base/dir", "../secret.jsonl"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := safePath(tt.baseDir, tt.filename)
			if err != ErrInvalidPath {
				t.Errorf("safePath should reject %q with ErrInvalidPath, got: %v", tt.filename, err)
			}
		})
	}
}

func TestSafePath_AbsolutePath(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{"absolute unix path", "/etc/passwd"},
		{"absolute with extension", "/var/log/secret.jsonl"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := safePath("/base/dir", tt.filename)
			if err != ErrInvalidPath {
				t.Errorf("safePath should reject absolute path %q with ErrInvalidPath, got: %v", tt.filename, err)
			}
		})
	}
}

func TestSafePath_ValidSubdirectory(t *testing.T) {
	result, err := safePath("/base/dir", "subdir/file.jsonl")
	if err != nil {
		t.Errorf("safePath should accept valid subdirectory path: %v", err)
	}
	expected := filepath.Join("/base/dir", "subdir/file.jsonl")
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestSafePath_DotInFilename(t *testing.T) {
	// Single dot in path component should be cleaned but allowed
	result, err := safePath("/base/dir", "./file.jsonl")
	if err != nil {
		t.Errorf("safePath should accept path with single dot: %v", err)
	}
	expected := filepath.Join("/base/dir", "file.jsonl")
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

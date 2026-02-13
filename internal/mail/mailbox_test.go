package mail

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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

	beforeAppend := time.Now()
	err = Append(tmpDir, msg)
	if err != nil {
		t.Fatalf("Append failed: %v", err)
	}
	afterAppend := time.Now()

	// Verify file exists
	filePath := filepath.Join(mailDir, "agent-2.jsonl")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("File should exist after Append")
	}

	// Verify message can be read back with correct fields
	messages, err := ReadAll(tmpDir, "agent-2")
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	restored := messages[0]
	if restored.ID != "testID01" {
		t.Errorf("Expected ID 'testID01', got '%s'", restored.ID)
	}
	if restored.From != "agent-1" {
		t.Errorf("Expected From 'agent-1', got '%s'", restored.From)
	}
	if restored.To != "agent-2" {
		t.Errorf("Expected To 'agent-2', got '%s'", restored.To)
	}
	if restored.Message != "Hello" {
		t.Errorf("Expected Message 'Hello', got '%s'", restored.Message)
	}
	if restored.ReadFlag != false {
		t.Errorf("Expected ReadFlag false, got %v", restored.ReadFlag)
	}

	// Verify CreatedAt was set to a reasonable time
	if restored.CreatedAt.Before(beforeAppend) || restored.CreatedAt.After(afterAppend) {
		t.Errorf("CreatedAt should be between %v and %v, got %v", beforeAppend, afterAppend, restored.CreatedAt)
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
	beforeFirst := time.Now()
	if err := Append(tmpDir, msg1); err != nil {
		t.Fatalf("First Append failed: %v", err)
	}
	afterFirst := time.Now()

	// Append second message
	msg2 := Message{
		ID:       "secndID2",
		From:     "agent-3",
		To:       "agent-2",
		Message:  "Second message",
		ReadFlag: false,
	}
	beforeSecond := time.Now()
	if err := Append(tmpDir, msg2); err != nil {
		t.Fatalf("Second Append failed: %v", err)
	}
	afterSecond := time.Now()

	// Verify both messages can be read back
	messages, err := ReadAll(tmpDir, "agent-2")
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	// Verify first message
	if messages[0].ID != "firstID1" {
		t.Errorf("First message ID mismatch: expected 'firstID1', got '%s'", messages[0].ID)
	}
	if messages[0].From != "agent-1" {
		t.Errorf("First message From mismatch: expected 'agent-1', got '%s'", messages[0].From)
	}
	if messages[0].Message != "First message" {
		t.Errorf("First message Message mismatch: expected 'First message', got '%s'", messages[0].Message)
	}
	if messages[0].CreatedAt.Before(beforeFirst) || messages[0].CreatedAt.After(afterFirst) {
		t.Errorf("First message CreatedAt should be between %v and %v, got %v", beforeFirst, afterFirst, messages[0].CreatedAt)
	}

	// Verify second message
	if messages[1].ID != "secndID2" {
		t.Errorf("Second message ID mismatch: expected 'secndID2', got '%s'", messages[1].ID)
	}
	if messages[1].From != "agent-3" {
		t.Errorf("Second message From mismatch: expected 'agent-3', got '%s'", messages[1].From)
	}
	if messages[1].Message != "Second message" {
		t.Errorf("Second message Message mismatch: expected 'Second message', got '%s'", messages[1].Message)
	}
	if messages[1].CreatedAt.Before(beforeSecond) || messages[1].CreatedAt.After(afterSecond) {
		t.Errorf("Second message CreatedAt should be between %v and %v, got %v", beforeSecond, afterSecond, messages[1].CreatedAt)
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

// T007: Tests for TryLockWithTimeout non-blocking file lock

func TestTryLockWithTimeout_Success(t *testing.T) {
	// Create a temp file for testing
	tmpFile, err := os.CreateTemp("", "agentmail-lock-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Lock on an unlocked file should succeed immediately
	err = TryLockWithTimeout(tmpFile, 100*time.Millisecond)
	if err != nil {
		t.Errorf("TryLockWithTimeout should succeed on unlocked file: %v", err)
	}
}

func TestTryLockWithTimeout_Timeout(t *testing.T) {
	// Create a temp file for testing
	tmpFile, err := os.CreateTemp("", "agentmail-lock-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// First, acquire an exclusive lock on the file (simulating another process holding it)
	// We need to open the file again to simulate another process
	tmpFile2, err := os.Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open temp file for second lock: %v", err)
	}
	defer tmpFile2.Close()

	// Acquire lock using the first file handle
	if err := TryLockWithTimeout(tmpFile, 100*time.Millisecond); err != nil {
		t.Fatalf("Failed to acquire initial lock: %v", err)
	}

	// Try to lock with the second file handle - should timeout
	start := time.Now()
	err = TryLockWithTimeout(tmpFile2, 100*time.Millisecond)
	elapsed := time.Since(start)

	if err != ErrFileLocked {
		t.Errorf("TryLockWithTimeout should return ErrFileLocked on already-locked file, got: %v", err)
	}

	// Verify timeout was respected (should be at least 100ms, give some margin for scheduling)
	if elapsed < 90*time.Millisecond {
		t.Errorf("TryLockWithTimeout should wait for timeout, elapsed: %v", elapsed)
	}
}

// Tests for pane address support

func TestAppend_WithPaneAddress(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	msg := Message{
		ID:       "testID01",
		From:     "mysession:sender.0",
		To:       "mysession:editor.0",
		Message:  "Hello from pane",
		ReadFlag: false,
	}

	err = Append(tmpDir, msg)
	if err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	// Verify file is created with sanitized filename
	sanitizedFilename := "mysession%3Aeditor%2E0.jsonl"
	filePath := filepath.Join(mailDir, sanitizedFilename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("File should exist with sanitized name: %s", sanitizedFilename)
	}

	// Verify message can be read back
	messages, err := ReadAll(tmpDir, "mysession:editor.0")
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].To != "mysession:editor.0" {
		t.Errorf("Expected To 'mysession:editor.0', got '%s'", messages[0].To)
	}
}

func TestReadAll_WithPaneAddress(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	// Write test data with sanitized filename
	content := `{"id":"id1","from":"mysession:sender.0","to":"mysession:editor.1","message":"Hello","read_flag":false}
`
	sanitizedFilename := "mysession%3Aeditor%2E1.jsonl"
	filePath := filepath.Join(mailDir, sanitizedFilename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	messages, err := ReadAll(tmpDir, "mysession:editor.1")
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].To != "mysession:editor.1" {
		t.Errorf("Expected To 'mysession:editor.1', got '%s'", messages[0].To)
	}
}

func TestListMailboxRecipients_ReturnsDecodedAddresses(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	// Create mailbox files with encoded pane addresses
	files := []string{
		"mysession%3Aeditor%2E0.jsonl",
		"mysession%3Aeditor%2E1.jsonl",
		"s%3Amy.app%2E0.jsonl",
	}

	for _, filename := range files {
		filePath := filepath.Join(mailDir, filename)
		if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filename, err)
		}
	}

	recipients, err := ListMailboxRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ListMailboxRecipients failed: %v", err)
	}

	if len(recipients) != 3 {
		t.Fatalf("Expected 3 recipients, got %d", len(recipients))
	}

	// Verify decoded addresses
	expected := []string{"mysession:editor.0", "mysession:editor.1", "s:my.app.0"}
	for _, exp := range expected {
		found := false
		for _, rec := range recipients {
			if rec == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected recipient '%s' not found in list %v", exp, recipients)
		}
	}
}

// Tests for cleanup functions

func TestCleanOldMessages(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	// Create messages with different ages
	oldTime := time.Now().Add(-48 * time.Hour)
	recentTime := time.Now().Add(-30 * time.Minute)

	messages := []Message{
		{ID: "old1", From: "sender", To: "mysession:editor.0", Message: "Old read", ReadFlag: true, CreatedAt: oldTime},
		{ID: "recent1", From: "sender", To: "mysession:editor.0", Message: "Recent read", ReadFlag: true, CreatedAt: recentTime},
		{ID: "unread1", From: "sender", To: "mysession:editor.0", Message: "Unread old", ReadFlag: false, CreatedAt: oldTime},
	}

	if err := WriteAll(tmpDir, "mysession:editor.0", messages); err != nil {
		t.Fatalf("Failed to write messages: %v", err)
	}

	// Clean messages older than 1 hour
	removed, err := CleanOldMessages(tmpDir, "mysession:editor.0", 1*time.Hour)
	if err != nil {
		t.Fatalf("CleanOldMessages failed: %v", err)
	}

	if removed != 1 {
		t.Errorf("Expected 1 message removed, got %d", removed)
	}

	// Verify remaining messages
	remaining, err := ReadAll(tmpDir, "mysession:editor.0")
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(remaining) != 2 {
		t.Errorf("Expected 2 remaining messages, got %d", len(remaining))
	}
}

func TestCountOldMessages(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	oldTime := time.Now().Add(-48 * time.Hour)
	recentTime := time.Now().Add(-30 * time.Minute)

	messages := []Message{
		{ID: "old1", From: "sender", To: "mysession:editor.0", Message: "Old read", ReadFlag: true, CreatedAt: oldTime},
		{ID: "recent1", From: "sender", To: "mysession:editor.0", Message: "Recent read", ReadFlag: true, CreatedAt: recentTime},
		{ID: "unread1", From: "sender", To: "mysession:editor.0", Message: "Unread", ReadFlag: false, CreatedAt: oldTime},
	}

	if err := WriteAll(tmpDir, "mysession:editor.0", messages); err != nil {
		t.Fatalf("Failed to write messages: %v", err)
	}

	count, err := CountOldMessages(tmpDir, "mysession:editor.0", 1*time.Hour)
	if err != nil {
		t.Fatalf("CountOldMessages failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

func TestRemoveEmptyMailboxes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	// Create empty mailbox
	emptyFile := filepath.Join(mailDir, "mysession%3Aeditor%2E0.jsonl")
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	// Create non-empty mailbox
	messages := []Message{
		{ID: "msg1", From: "sender", To: "mysession:editor.1", Message: "Hello", ReadFlag: false},
	}
	if err := WriteAll(tmpDir, "mysession:editor.1", messages); err != nil {
		t.Fatalf("Failed to write messages: %v", err)
	}

	removed, err := RemoveEmptyMailboxes(tmpDir)
	if err != nil {
		t.Fatalf("RemoveEmptyMailboxes failed: %v", err)
	}

	if removed != 1 {
		t.Errorf("Expected 1 mailbox removed, got %d", removed)
	}

	// Verify empty file is gone
	if _, err := os.Stat(emptyFile); !os.IsNotExist(err) {
		t.Error("Empty mailbox file should be removed")
	}
}

func TestCountEmptyMailboxes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}

	// Create empty mailbox
	emptyFile := filepath.Join(mailDir, "mysession%3Aeditor%2E0.jsonl")
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	// Create non-empty mailbox
	messages := []Message{
		{ID: "msg1", From: "sender", To: "mysession:editor.1", Message: "Hello", ReadFlag: false},
	}
	if err := WriteAll(tmpDir, "mysession:editor.1", messages); err != nil {
		t.Fatalf("Failed to write messages: %v", err)
	}

	count, err := CountEmptyMailboxes(tmpDir)
	if err != nil {
		t.Fatalf("CountEmptyMailboxes failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

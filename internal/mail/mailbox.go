package mail

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// RootDir is the root directory for AgentMail storage
const RootDir = ".agentmail"

// MailDir is the directory name for mailbox storage
const MailDir = ".agentmail/mailboxes"

// ErrInvalidPath is returned when a path traversal attack is detected.
var ErrInvalidPath = errors.New("invalid path: directory traversal detected")

// safePath constructs a safe file path and validates it stays within the base directory.
// This prevents path traversal attacks (G304) by ensuring the cleaned path
// is still under the expected base directory.
func safePath(baseDir, filename string) (string, error) {
	// Clean the filename to remove any .. or . components
	cleanName := filepath.Clean(filename)

	// Reject if the cleaned name tries to escape (starts with .. or /)
	if strings.HasPrefix(cleanName, "..") || filepath.IsAbs(cleanName) {
		return "", ErrInvalidPath
	}

	// Construct the full path
	fullPath := filepath.Join(baseDir, cleanName)

	// Verify the result is still under baseDir
	// Use Clean on baseDir too for consistent comparison
	cleanBase := filepath.Clean(baseDir)
	if !strings.HasPrefix(fullPath, cleanBase+string(filepath.Separator)) && fullPath != cleanBase {
		return "", ErrInvalidPath
	}

	return fullPath, nil
}

// EnsureMailDir creates the .agentmail/ and .agentmail/mailboxes/ directories if they don't exist.
func EnsureMailDir(repoRoot string) error {
	// Create root directory first
	rootPath := filepath.Join(repoRoot, RootDir)
	if err := os.MkdirAll(rootPath, 0750); err != nil { // G301: restricted directory permissions
		return err
	}

	// Create mailboxes directory
	mailPath := filepath.Join(repoRoot, MailDir)
	return os.MkdirAll(mailPath, 0750) // G301: restricted directory permissions
}

// Append adds a message to the recipient's mailbox file with file locking.
func Append(repoRoot string, msg Message) error {
	// Ensure mail directory exists
	if err := EnsureMailDir(repoRoot); err != nil {
		return err
	}

	// Build file path for recipient with path traversal protection (G304)
	mailDir := filepath.Join(repoRoot, MailDir)
	filePath, err := safePath(mailDir, msg.To+".jsonl")
	if err != nil {
		return err
	}

	// Open file for appending (create if not exists)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600) // #nosec G304 - path validated by safePath; G302 - restricted file permissions
	if err != nil {
		return err
	}

	// Acquire exclusive lock on the file
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		_ = file.Close() // G104: error intentionally ignored in cleanup path
		return err
	}

	// Marshal message to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN) // G104: error intentionally ignored in cleanup path
		_ = file.Close()
		return err
	}

	// Write JSON line (append newline)
	_, err = file.Write(append(data, '\n'))

	// Unlock before close (correct order)
	_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN) // G104: unlock errors don't affect the write result
	_ = file.Close()                                   // G104: close errors don't affect the write result
	return err
}

// ReadAll reads all messages from a recipient's mailbox file.
func ReadAll(repoRoot string, recipient string) ([]Message, error) {
	// Build file path with path traversal protection (G304)
	mailDir := filepath.Join(repoRoot, MailDir)
	filePath, err := safePath(mailDir, recipient+".jsonl")
	if err != nil {
		return nil, err
	}

	// Open file for reading
	file, err := os.Open(filePath) // #nosec G304 - path validated by safePath
	if err != nil {
		if os.IsNotExist(err) {
			return []Message{}, nil
		}
		return nil, err
	}
	defer file.Close()

	// Acquire shared lock
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_SH); err != nil {
		return nil, err
	}
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var messages []Message
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		var msg Message
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// FindUnread returns all unread messages for a recipient in FIFO order.
func FindUnread(repoRoot string, recipient string) ([]Message, error) {
	messages, err := ReadAll(repoRoot, recipient)
	if err != nil {
		return nil, err
	}

	var unread []Message
	for _, msg := range messages {
		if !msg.ReadFlag {
			unread = append(unread, msg)
		}
	}

	return unread, nil
}

// writeAllLocked writes all messages to an already-locked file.
// The caller is responsible for locking and unlocking.
func writeAllLocked(file *os.File, messages []Message) error {
	// Truncate the file
	if err := file.Truncate(0); err != nil {
		return err
	}
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	// Write each message as a JSON line
	for _, msg := range messages {
		data, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		if _, err := file.Write(append(data, '\n')); err != nil {
			return err
		}
	}

	return nil
}

// WriteAll writes all messages to a recipient's mailbox file with locking.
func WriteAll(repoRoot string, recipient string, messages []Message) error {
	// Ensure mail directory exists
	if err := EnsureMailDir(repoRoot); err != nil {
		return err
	}

	// Build file path with path traversal protection (G304)
	mailDir := filepath.Join(repoRoot, MailDir)
	filePath, err := safePath(mailDir, recipient+".jsonl")
	if err != nil {
		return err
	}

	// Open file for read/write
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0600) // #nosec G304 - path validated by safePath; G302 - restricted file permissions
	if err != nil {
		return err
	}

	// Acquire exclusive lock
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		_ = file.Close() // G104: error intentionally ignored in cleanup path
		return err
	}

	// Write messages
	writeErr := writeAllLocked(file, messages)

	// Unlock before close (correct order)
	_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN) // G104: unlock errors don't affect the write result
	_ = file.Close()                                   // G104: close errors don't affect the write result
	return writeErr
}

// MarkAsRead marks a specific message as read in the recipient's mailbox.
// This function is atomic - it holds a lock during the entire read-modify-write cycle.
func MarkAsRead(repoRoot string, recipient string, messageID string) error {
	// Ensure mail directory exists
	if err := EnsureMailDir(repoRoot); err != nil {
		return err
	}

	// Build file path with path traversal protection (G304)
	mailDir := filepath.Join(repoRoot, MailDir)
	filePath, err := safePath(mailDir, recipient+".jsonl")
	if err != nil {
		return err
	}

	// Open file for read/write
	file, err := os.OpenFile(filePath, os.O_RDWR, 0600) // #nosec G304 - path validated by safePath; G302 - restricted file permissions
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No messages to mark
		}
		return err
	}

	// Acquire exclusive lock for atomic read-modify-write
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		_ = file.Close() // G104: error intentionally ignored in cleanup path
		return err
	}

	// Read all messages while holding lock
	data, err := os.ReadFile(filePath) // #nosec G304 - path validated by safePath
	if err != nil {
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN) // G104: error intentionally ignored in cleanup path
		_ = file.Close()
		return err
	}

	var messages []Message
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		var msg Message
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN) // G104: error intentionally ignored in cleanup path
			_ = file.Close()
			return err
		}
		messages = append(messages, msg)
	}

	// Find and update the message
	for i := range messages {
		if messages[i].ID == messageID {
			messages[i].ReadFlag = true
			break
		}
	}

	// Write back while still holding lock
	writeErr := writeAllLocked(file, messages)

	// Unlock before close (correct order)
	_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN) // G104: unlock errors don't affect the write result
	_ = file.Close()                                   // G104: close errors don't affect the write result
	return writeErr
}

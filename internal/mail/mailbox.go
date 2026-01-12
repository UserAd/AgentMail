package mail

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// MailDir is the directory name for mail storage within .git
const MailDir = ".git/mail"

// EnsureMailDir creates the .git/mail/ directory if it doesn't exist.
// T018: Create .git/mail/ if missing
func EnsureMailDir(repoRoot string) error {
	mailPath := filepath.Join(repoRoot, MailDir)
	return os.MkdirAll(mailPath, 0755)
}

// Append adds a message to the recipient's mailbox file with file locking.
// T019: Append to .git/mail/<recipient>.jsonl with file locking
func Append(repoRoot string, msg Message) error {
	// Ensure mail directory exists
	if err := EnsureMailDir(repoRoot); err != nil {
		return err
	}

	// Build file path for recipient
	filePath := filepath.Join(repoRoot, MailDir, msg.To+".jsonl")

	// Open file for appending (create if not exists)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// Acquire exclusive lock on the file
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		file.Close()
		return err
	}

	// Marshal message to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		file.Close()
		return err
	}

	// Write JSON line (append newline)
	_, err = file.Write(append(data, '\n'))

	// Unlock before close (correct order)
	syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	file.Close()
	return err
}

// ReadAll reads all messages from a recipient's mailbox file.
// T030: Read .git/mail/<recipient>.jsonl
func ReadAll(repoRoot string, recipient string) ([]Message, error) {
	filePath := filepath.Join(repoRoot, MailDir, recipient+".jsonl")

	// Open file for reading
	file, err := os.Open(filePath)
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
// T031: Filter by read_flag only, no recipient filter needed (per-recipient files)
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
// T032: Write to recipient file with locking
func WriteAll(repoRoot string, recipient string, messages []Message) error {
	// Ensure mail directory exists
	if err := EnsureMailDir(repoRoot); err != nil {
		return err
	}

	filePath := filepath.Join(repoRoot, MailDir, recipient+".jsonl")

	// Open file for read/write
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	// Acquire exclusive lock
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		file.Close()
		return err
	}

	// Write messages
	writeErr := writeAllLocked(file, messages)

	// Unlock before close (correct order)
	syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	file.Close()
	return writeErr
}

// MarkAsRead marks a specific message as read in the recipient's mailbox.
// T033: Implement MarkAsRead function
// This function is atomic - it holds a lock during the entire read-modify-write cycle.
func MarkAsRead(repoRoot string, recipient string, messageID string) error {
	// Ensure mail directory exists
	if err := EnsureMailDir(repoRoot); err != nil {
		return err
	}

	filePath := filepath.Join(repoRoot, MailDir, recipient+".jsonl")

	// Open file for read/write
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No messages to mark
		}
		return err
	}

	// Acquire exclusive lock for atomic read-modify-write
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		file.Close()
		return err
	}

	// Read all messages while holding lock
	data, err := os.ReadFile(filePath)
	if err != nil {
		syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		file.Close()
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
			syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
			file.Close()
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
	syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	file.Close()
	return writeErr
}

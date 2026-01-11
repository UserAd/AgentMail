package mail

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	defer file.Close()

	// Acquire exclusive lock on the file
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

	// Marshal message to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Write JSON line (append newline)
	_, err = file.Write(append(data, '\n'))
	return err
}

// ReadAll reads all messages from a recipient's mailbox file.
// T030: Read .git/mail/<recipient>.jsonl
func ReadAll(repoRoot string, recipient string) ([]Message, error) {
	filePath := filepath.Join(repoRoot, MailDir, recipient+".jsonl")

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Message{}, nil
		}
		return nil, err
	}

	var messages []Message
	lines := splitLines(data)

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var msg Message
		if err := json.Unmarshal(line, &msg); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// splitLines splits byte data by newlines, preserving each line as []byte.
func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	// Handle last line without trailing newline
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
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

// WriteAll writes all messages to a recipient's mailbox file with locking.
// T032: Write to recipient file with locking
func WriteAll(repoRoot string, recipient string, messages []Message) error {
	// Ensure mail directory exists
	if err := EnsureMailDir(repoRoot); err != nil {
		return err
	}

	filePath := filepath.Join(repoRoot, MailDir, recipient+".jsonl")

	// Open file for writing (truncate)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Acquire exclusive lock
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

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

// MarkAsRead marks a specific message as read in the recipient's mailbox.
// T033: Implement MarkAsRead function
func MarkAsRead(repoRoot string, recipient string, messageID string) error {
	messages, err := ReadAll(repoRoot, recipient)
	if err != nil {
		return err
	}

	// Find and update the message
	for i := range messages {
		if messages[i].ID == messageID {
			messages[i].ReadFlag = true
			break
		}
	}

	// Write back all messages
	return WriteAll(repoRoot, recipient, messages)
}

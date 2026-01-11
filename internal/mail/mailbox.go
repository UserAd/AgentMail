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

package mail

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// RecipientsFile is the filename for recipients state storage within .git
const RecipientsFile = ".git/mail-recipients.jsonl"

// Status constants for recipient availability
const (
	StatusReady   = "ready"
	StatusWork    = "work"
	StatusOffline = "offline"
)

// RecipientState represents the availability state of a recipient agent
type RecipientState struct {
	Recipient string    `json:"recipient"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
	Notified  bool      `json:"notified"`
}

// ReadAllRecipients reads and parses all recipient states from .git/mail-recipients.jsonl
func ReadAllRecipients(repoRoot string) ([]RecipientState, error) {
	filePath := filepath.Join(repoRoot, RecipientsFile)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []RecipientState{}, nil
		}
		return nil, err
	}

	var recipients []RecipientState
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		var state RecipientState
		if err := json.Unmarshal([]byte(line), &state); err != nil {
			return nil, err
		}
		recipients = append(recipients, state)
	}

	return recipients, nil
}

// writeAllRecipientsLocked writes all recipient states to an already-locked file.
// The caller is responsible for locking and unlocking.
func writeAllRecipientsLocked(file *os.File, recipients []RecipientState) error {
	// Truncate the file
	if err := file.Truncate(0); err != nil {
		return err
	}
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	// Write each recipient state as a JSON line
	for _, state := range recipients {
		data, err := json.Marshal(state)
		if err != nil {
			return err
		}
		if _, err := file.Write(append(data, '\n')); err != nil {
			return err
		}
	}

	return nil
}

// WriteAllRecipients writes all recipient states to .git/mail-recipients.jsonl with file locking
func WriteAllRecipients(repoRoot string, recipients []RecipientState) error {
	filePath := filepath.Join(repoRoot, RecipientsFile)

	// Ensure parent directory exists (.git should exist, but be safe)
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

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

	// Write recipients
	writeErr := writeAllRecipientsLocked(file, recipients)

	// Unlock before close (correct order)
	syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	file.Close()
	return writeErr
}

// UpdateRecipientState performs an atomic read-modify-write to update a recipient's state.
// If the recipient doesn't exist, it will be added.
// If resetNotified is true, the Notified field will be set to false.
func UpdateRecipientState(repoRoot string, recipient string, status string, resetNotified bool) error {
	filePath := filepath.Join(repoRoot, RecipientsFile)

	// Ensure parent directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Open file for read/write (create if not exists)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	// Acquire exclusive lock for atomic read-modify-write
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		file.Close()
		return err
	}

	// Read all recipient states while holding lock
	data, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		file.Close()
		return err
	}

	var recipients []RecipientState
	if len(data) > 0 {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			var state RecipientState
			if err := json.Unmarshal([]byte(line), &state); err != nil {
				syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
				file.Close()
				return err
			}
			recipients = append(recipients, state)
		}
	}

	// Find and update the recipient, or add new
	found := false
	now := time.Now()
	for i := range recipients {
		if recipients[i].Recipient == recipient {
			recipients[i].Status = status
			recipients[i].UpdatedAt = now
			if resetNotified {
				recipients[i].Notified = false
			}
			found = true
			break
		}
	}

	if !found {
		newState := RecipientState{
			Recipient: recipient,
			Status:    status,
			UpdatedAt: now,
			Notified:  false,
		}
		recipients = append(recipients, newState)
	}

	// Write back while still holding lock
	writeErr := writeAllRecipientsLocked(file, recipients)

	// Unlock before close (correct order)
	syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	file.Close()
	return writeErr
}

// ListMailboxRecipients returns a list of all recipients who have mailbox files.
// It scans the .git/mail/ directory for .jsonl files (excluding non-mailbox files).
func ListMailboxRecipients(repoRoot string) ([]string, error) {
	mailPath := filepath.Join(repoRoot, MailDir)

	entries, err := os.ReadDir(mailPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var recipients []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Only include .jsonl files (mailbox files)
		if !strings.HasSuffix(name, ".jsonl") {
			continue
		}
		// Extract recipient name (remove .jsonl suffix)
		recipient := strings.TrimSuffix(name, ".jsonl")
		recipients = append(recipients, recipient)
	}

	return recipients, nil
}

// CleanStaleStates removes recipient states that haven't been updated within the threshold.
// This is used to clean up states for agents that are no longer active.
func CleanStaleStates(repoRoot string, threshold time.Duration) error {
	filePath := filepath.Join(repoRoot, RecipientsFile)

	// Open file for read/write (create if not exists)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Acquire exclusive lock for atomic read-modify-write
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		file.Close()
		return err
	}

	// Read all recipient states while holding lock
	data, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		file.Close()
		return err
	}

	var recipients []RecipientState
	if len(data) > 0 {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			var state RecipientState
			if err := json.Unmarshal([]byte(line), &state); err != nil {
				syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
				file.Close()
				return err
			}
			recipients = append(recipients, state)
		}
	}

	// Filter out stale states
	cutoff := time.Now().Add(-threshold)
	var fresh []RecipientState
	for _, r := range recipients {
		if r.UpdatedAt.After(cutoff) {
			fresh = append(fresh, r)
		}
	}

	// Write back while still holding lock
	writeErr := writeAllRecipientsLocked(file, fresh)

	// Unlock before close (correct order)
	syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	file.Close()
	return writeErr
}

// SetNotifiedFlag sets the Notified flag for a specific recipient.
// If the recipient doesn't exist, this is a no-op (doesn't create new state).
func SetNotifiedFlag(repoRoot string, recipient string, notified bool) error {
	filePath := filepath.Join(repoRoot, RecipientsFile)

	// Open file for read/write
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No recipients file, nothing to update
		}
		return err
	}

	// Acquire exclusive lock for atomic read-modify-write
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		file.Close()
		return err
	}

	// Read all recipient states while holding lock
	data, err := os.ReadFile(filePath)
	if err != nil {
		syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		file.Close()
		return err
	}

	var recipients []RecipientState
	if len(data) > 0 {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			var state RecipientState
			if err := json.Unmarshal([]byte(line), &state); err != nil {
				syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
				file.Close()
				return err
			}
			recipients = append(recipients, state)
		}
	}

	// Find and update the recipient
	found := false
	for i := range recipients {
		if recipients[i].Recipient == recipient {
			recipients[i].Notified = notified
			found = true
			break
		}
	}

	if !found {
		// Recipient doesn't exist, don't create it
		syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		file.Close()
		return nil
	}

	// Write back while still holding lock
	writeErr := writeAllRecipientsLocked(file, recipients)

	// Unlock before close (correct order)
	syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	file.Close()
	return writeErr
}

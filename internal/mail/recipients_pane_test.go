package mail

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// =============================================================================
// T005: Tests for recipients.go with pane addresses
// =============================================================================

func TestUpdateRecipientState_WithPaneAddress(t *testing.T) {
	tmpDir := t.TempDir()

	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	// Update recipient using pane address
	err := UpdateRecipientState(tmpDir, "mysession:editor.1", StatusReady, false)
	if err != nil {
		t.Fatalf("UpdateRecipientState failed: %v", err)
	}

	// Verify recipient was added with pane address
	recipients, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(recipients) != 1 {
		t.Fatalf("Expected 1 recipient, got %d", len(recipients))
	}

	if recipients[0].Recipient != "mysession:editor.1" {
		t.Errorf("Expected recipient 'mysession:editor.1', got '%s'", recipients[0].Recipient)
	}
	if recipients[0].Status != StatusReady {
		t.Errorf("Expected status ready, got %s", recipients[0].Status)
	}
}

func TestReadAllRecipients_MultiplePanes(t *testing.T) {
	tmpDir := t.TempDir()

	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	now := time.Now().Truncate(time.Second)
	states := []RecipientState{
		{
			Recipient:  "mysession:editor.0",
			Status:     StatusReady,
			UpdatedAt:  now,
			NotifiedAt: time.Time{},
		},
		{
			Recipient:  "mysession:editor.1",
			Status:     StatusWork,
			UpdatedAt:  now,
			NotifiedAt: time.Now(),
		},
		{
			Recipient:  "mysession:terminal.0",
			Status:     StatusOffline,
			UpdatedAt:  now,
			NotifiedAt: time.Time{},
		},
	}

	filePath := filepath.Join(agentmailDir, "recipients.jsonl")
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	for _, state := range states {
		data, _ := json.Marshal(state)
		file.Write(append(data, '\n'))
	}
	file.Close()

	recipients, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(recipients) != 3 {
		t.Fatalf("Expected 3 recipients, got %d", len(recipients))
	}

	// Verify each pane is tracked separately
	expected := map[string]string{
		"mysession:editor.0":   StatusReady,
		"mysession:editor.1":   StatusWork,
		"mysession:terminal.0": StatusOffline,
	}

	for _, r := range recipients {
		expectedStatus, ok := expected[r.Recipient]
		if !ok {
			t.Errorf("Unexpected recipient: %s", r.Recipient)
			continue
		}
		if r.Status != expectedStatus {
			t.Errorf("Recipient %s: expected status %s, got %s", r.Recipient, expectedStatus, r.Status)
		}
	}
}

func TestUpdateLastReadAt_WithPaneAddress(t *testing.T) {
	tmpDir := t.TempDir()

	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	// Create initial recipient with pane address
	now := time.Now().Truncate(time.Second)
	existing := RecipientState{
		Recipient:  "mysession:editor.1",
		Status:     StatusWork,
		UpdatedAt:  now,
		NotifiedAt: time.Now(),
		LastReadAt: 1000000000000,
	}
	filePath := filepath.Join(agentmailDir, "recipients.jsonl")
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	data, _ := json.Marshal(existing)
	file.Write(append(data, '\n'))
	file.Close()

	// Update LastReadAt using pane address
	newTimestamp := int64(1704067200000)
	err = UpdateLastReadAt(tmpDir, "mysession:editor.1", newTimestamp)
	if err != nil {
		t.Fatalf("UpdateLastReadAt failed: %v", err)
	}

	// Verify LastReadAt was updated
	recipients, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(recipients) != 1 {
		t.Fatalf("Expected 1 recipient, got %d", len(recipients))
	}

	if recipients[0].Recipient != "mysession:editor.1" {
		t.Errorf("Expected recipient 'mysession:editor.1', got '%s'", recipients[0].Recipient)
	}
	if recipients[0].LastReadAt != newTimestamp {
		t.Errorf("Expected LastReadAt %d, got %d", newTimestamp, recipients[0].LastReadAt)
	}
}

func TestCleanOfflineRecipients_WithPaneAddresses(t *testing.T) {
	tmpDir := t.TempDir()

	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	now := time.Now()
	recipients := []RecipientState{
		{Recipient: "mysession:editor.0", Status: StatusReady, UpdatedAt: now, NotifiedAt: time.Time{}},
		{Recipient: "mysession:editor.1", Status: StatusWork, UpdatedAt: now, NotifiedAt: time.Time{}},
		{Recipient: "mysession:stale.0", Status: StatusReady, UpdatedAt: now, NotifiedAt: time.Time{}},
	}
	if err := WriteAllRecipients(tmpDir, recipients); err != nil {
		t.Fatalf("WriteAllRecipients failed: %v", err)
	}

	// Clean offline recipients - keep only editor.0 and editor.1
	validPanes := []string{"mysession:editor.0", "mysession:editor.1"}
	_, err := CleanOfflineRecipients(tmpDir, validPanes)
	if err != nil {
		t.Fatalf("CleanOfflineRecipients failed: %v", err)
	}

	// Verify only valid panes remain
	readBack, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(readBack) != 2 {
		t.Fatalf("Expected 2 recipients after cleanup, got %d", len(readBack))
	}

	for _, r := range readBack {
		if r.Recipient != "mysession:editor.0" && r.Recipient != "mysession:editor.1" {
			t.Errorf("Unexpected recipient after cleanup: %s", r.Recipient)
		}
	}
}

func TestSetNotifiedAt_WithPaneAddress(t *testing.T) {
	tmpDir := t.TempDir()

	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	now := time.Now()
	recipients := []RecipientState{
		{Recipient: "mysession:editor.0", Status: StatusReady, UpdatedAt: now, NotifiedAt: time.Time{}},
		{Recipient: "mysession:editor.1", Status: StatusWork, UpdatedAt: now, NotifiedAt: time.Time{}},
	}
	if err := WriteAllRecipients(tmpDir, recipients); err != nil {
		t.Fatalf("WriteAllRecipients failed: %v", err)
	}

	// Set notified timestamp for one pane
	notifiedTime := time.Now()
	err := SetNotifiedAt(tmpDir, "mysession:editor.1", notifiedTime)
	if err != nil {
		t.Fatalf("SetNotifiedAt failed: %v", err)
	}

	// Verify only editor.1 has NotifiedAt set
	readBack, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	for _, r := range readBack {
		switch r.Recipient {
		case "mysession:editor.0":
			if !r.NotifiedAt.IsZero() {
				t.Error("editor.0 should not have NotifiedAt set")
			}
		case "mysession:editor.1":
			if r.NotifiedAt.IsZero() {
				t.Error("editor.1 should have NotifiedAt set")
			}
			if r.NotifiedAt.Unix() != notifiedTime.Unix() {
				t.Errorf("editor.1 NotifiedAt mismatch: expected %v, got %v", notifiedTime.Unix(), r.NotifiedAt.Unix())
			}
		}
	}
}

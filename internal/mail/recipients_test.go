package mail

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestReadAllRecipients_Empty - returns empty slice when file doesn't exist
func TestReadAllRecipients_Empty(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory to simulate git repo
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// File does not exist - should return empty slice, no error
	recipients, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients should not error for missing file: %v", err)
	}

	if recipients == nil {
		t.Error("Expected empty slice, got nil")
	}

	if len(recipients) != 0 {
		t.Errorf("Expected 0 recipients for missing file, got %d", len(recipients))
	}
}

// TestReadAllRecipients_ParsesJSONL - reads and parses valid JSONL
func TestReadAllRecipients_ParsesJSONL(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Create test data with multiple recipient states
	now := time.Now().Truncate(time.Second)
	state1 := RecipientState{
		Recipient: "agent-1",
		Status:    StatusReady,
		UpdatedAt: now,
		Notified:  false,
	}
	state2 := RecipientState{
		Recipient: "agent-2",
		Status:    StatusWork,
		UpdatedAt: now.Add(-time.Hour),
		Notified:  true,
	}
	state3 := RecipientState{
		Recipient: "agent-3",
		Status:    StatusOffline,
		UpdatedAt: now.Add(-2 * time.Hour),
		Notified:  false,
	}

	// Write JSONL content
	filePath := filepath.Join(gitDir, "mail-recipients.jsonl")
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	for _, state := range []RecipientState{state1, state2, state3} {
		data, _ := json.Marshal(state)
		file.Write(append(data, '\n'))
	}
	file.Close()

	// Read and verify
	recipients, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(recipients) != 3 {
		t.Fatalf("Expected 3 recipients, got %d", len(recipients))
	}

	// Verify first recipient
	if recipients[0].Recipient != "agent-1" {
		t.Errorf("First recipient should be agent-1, got %s", recipients[0].Recipient)
	}
	if recipients[0].Status != StatusReady {
		t.Errorf("First recipient status should be ready, got %s", recipients[0].Status)
	}
	if recipients[0].Notified != false {
		t.Error("First recipient should not be notified")
	}

	// Verify second recipient
	if recipients[1].Recipient != "agent-2" {
		t.Errorf("Second recipient should be agent-2, got %s", recipients[1].Recipient)
	}
	if recipients[1].Status != StatusWork {
		t.Errorf("Second recipient status should be work, got %s", recipients[1].Status)
	}
	if recipients[1].Notified != true {
		t.Error("Second recipient should be notified")
	}

	// Verify third recipient
	if recipients[2].Recipient != "agent-3" {
		t.Errorf("Third recipient should be agent-3, got %s", recipients[2].Recipient)
	}
	if recipients[2].Status != StatusOffline {
		t.Errorf("Third recipient status should be offline, got %s", recipients[2].Status)
	}
}

// TestReadAllRecipients_HandlesEmptyLines - handles JSONL with empty lines
func TestReadAllRecipients_HandlesEmptyLines(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Create JSONL with empty lines
	content := `{"recipient":"agent-1","status":"ready","updated_at":"2024-01-01T00:00:00Z","notified":false}

{"recipient":"agent-2","status":"work","updated_at":"2024-01-01T00:00:00Z","notified":true}
`
	filePath := filepath.Join(gitDir, "mail-recipients.jsonl")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	recipients, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(recipients) != 2 {
		t.Errorf("Expected 2 recipients (skipping empty lines), got %d", len(recipients))
	}
}

// TestWriteAllRecipients - writes recipients to file correctly
func TestWriteAllRecipients(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	now := time.Now().Truncate(time.Second)
	recipients := []RecipientState{
		{
			Recipient: "agent-1",
			Status:    StatusReady,
			UpdatedAt: now,
			Notified:  false,
		},
		{
			Recipient: "agent-2",
			Status:    StatusWork,
			UpdatedAt: now,
			Notified:  true,
		},
	}

	// Write recipients
	err := WriteAllRecipients(tmpDir, recipients)
	if err != nil {
		t.Fatalf("WriteAllRecipients failed: %v", err)
	}

	// Verify file exists and content is correct
	filePath := filepath.Join(gitDir, "mail-recipients.jsonl")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Read back and verify
	readBack, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(readBack) != 2 {
		t.Fatalf("Expected 2 recipients, got %d", len(readBack))
	}

	if readBack[0].Recipient != "agent-1" || readBack[0].Status != StatusReady {
		t.Errorf("First recipient mismatch: %+v", readBack[0])
	}
	if readBack[1].Recipient != "agent-2" || readBack[1].Status != StatusWork {
		t.Errorf("Second recipient mismatch: %+v", readBack[1])
	}

	// Verify content ends with newline
	if len(content) > 0 && content[len(content)-1] != '\n' {
		t.Error("File should end with newline")
	}
}

// TestWriteAllRecipients_CreatesParentDir - creates .git dir if missing
func TestWriteAllRecipients_CreatesParentDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Note: .git directory does NOT exist yet

	recipients := []RecipientState{
		{
			Recipient: "agent-1",
			Status:    StatusReady,
			UpdatedAt: time.Now(),
			Notified:  false,
		},
	}

	// Write recipients - should create .git directory
	err := WriteAllRecipients(tmpDir, recipients)
	if err != nil {
		t.Fatalf("WriteAllRecipients failed: %v", err)
	}

	// Verify file exists
	filePath := filepath.Join(tmpDir, ".git", "mail-recipients.jsonl")
	if _, err := os.Stat(filePath); err != nil {
		t.Errorf("File should exist: %v", err)
	}
}

// TestWriteAllRecipients_OverwritesExisting - overwrites existing file
func TestWriteAllRecipients_OverwritesExisting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Write initial content
	initial := []RecipientState{
		{Recipient: "old-agent", Status: StatusOffline, UpdatedAt: time.Now(), Notified: true},
	}
	if err := WriteAllRecipients(tmpDir, initial); err != nil {
		t.Fatalf("Initial WriteAllRecipients failed: %v", err)
	}

	// Write new content
	updated := []RecipientState{
		{Recipient: "new-agent", Status: StatusReady, UpdatedAt: time.Now(), Notified: false},
	}
	if err := WriteAllRecipients(tmpDir, updated); err != nil {
		t.Fatalf("Updated WriteAllRecipients failed: %v", err)
	}

	// Verify only new content exists
	readBack, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(readBack) != 1 {
		t.Fatalf("Expected 1 recipient after overwrite, got %d", len(readBack))
	}

	if readBack[0].Recipient != "new-agent" {
		t.Errorf("Expected new-agent, got %s", readBack[0].Recipient)
	}
}

// TestUpdateRecipientState_NewRecipient - adds new recipient
func TestUpdateRecipientState_NewRecipient(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Update a new recipient (file doesn't exist yet)
	err := UpdateRecipientState(tmpDir, "new-agent", StatusReady, false)
	if err != nil {
		t.Fatalf("UpdateRecipientState failed: %v", err)
	}

	// Verify recipient was added
	recipients, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(recipients) != 1 {
		t.Fatalf("Expected 1 recipient, got %d", len(recipients))
	}

	if recipients[0].Recipient != "new-agent" {
		t.Errorf("Expected recipient new-agent, got %s", recipients[0].Recipient)
	}
	if recipients[0].Status != StatusReady {
		t.Errorf("Expected status ready, got %s", recipients[0].Status)
	}
	if recipients[0].Notified != false {
		t.Error("New recipient should have Notified=false")
	}
	if recipients[0].UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

// TestUpdateRecipientState_UpdateExisting - updates existing recipient
func TestUpdateRecipientState_UpdateExisting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Create initial recipient
	initial := []RecipientState{
		{
			Recipient: "agent-1",
			Status:    StatusReady,
			UpdatedAt: time.Now().Add(-time.Hour),
			Notified:  true,
		},
		{
			Recipient: "agent-2",
			Status:    StatusOffline,
			UpdatedAt: time.Now().Add(-time.Hour),
			Notified:  false,
		},
	}
	if err := WriteAllRecipients(tmpDir, initial); err != nil {
		t.Fatalf("WriteAllRecipients failed: %v", err)
	}

	beforeUpdate := time.Now()

	// Update agent-1's status
	err := UpdateRecipientState(tmpDir, "agent-1", StatusWork, false)
	if err != nil {
		t.Fatalf("UpdateRecipientState failed: %v", err)
	}

	// Verify update
	recipients, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(recipients) != 2 {
		t.Fatalf("Expected 2 recipients, got %d", len(recipients))
	}

	// Find agent-1
	var agent1 *RecipientState
	for i := range recipients {
		if recipients[i].Recipient == "agent-1" {
			agent1 = &recipients[i]
			break
		}
	}

	if agent1 == nil {
		t.Fatal("agent-1 not found")
	}

	if agent1.Status != StatusWork {
		t.Errorf("Expected status work, got %s", agent1.Status)
	}
	if agent1.Notified != true {
		t.Error("Notified should remain true when resetNotified is false")
	}
	if agent1.UpdatedAt.Before(beforeUpdate) {
		t.Error("UpdatedAt should be updated to current time")
	}

	// Verify agent-2 is unchanged
	var agent2 *RecipientState
	for i := range recipients {
		if recipients[i].Recipient == "agent-2" {
			agent2 = &recipients[i]
			break
		}
	}

	if agent2 == nil {
		t.Fatal("agent-2 not found")
	}
	if agent2.Status != StatusOffline {
		t.Errorf("agent-2 should be unchanged, got status %s", agent2.Status)
	}
}

// TestUpdateRecipientState_ResetNotified - resets notified flag when requested
func TestUpdateRecipientState_ResetNotified(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Create initial recipient with Notified=true
	initial := []RecipientState{
		{
			Recipient: "agent-1",
			Status:    StatusWork,
			UpdatedAt: time.Now().Add(-time.Hour),
			Notified:  true,
		},
	}
	if err := WriteAllRecipients(tmpDir, initial); err != nil {
		t.Fatalf("WriteAllRecipients failed: %v", err)
	}

	// Update with resetNotified=true
	err := UpdateRecipientState(tmpDir, "agent-1", StatusReady, true)
	if err != nil {
		t.Fatalf("UpdateRecipientState failed: %v", err)
	}

	// Verify Notified was reset
	recipients, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(recipients) != 1 {
		t.Fatalf("Expected 1 recipient, got %d", len(recipients))
	}

	if recipients[0].Notified != false {
		t.Error("Notified should be reset to false when resetNotified is true")
	}
	if recipients[0].Status != StatusReady {
		t.Errorf("Status should be updated to ready, got %s", recipients[0].Status)
	}
}

// TestUpdateRecipientState_AddToExisting - adds new recipient to existing file
func TestUpdateRecipientState_AddToExisting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Create initial recipient
	initial := []RecipientState{
		{
			Recipient: "existing-agent",
			Status:    StatusReady,
			UpdatedAt: time.Now(),
			Notified:  false,
		},
	}
	if err := WriteAllRecipients(tmpDir, initial); err != nil {
		t.Fatalf("WriteAllRecipients failed: %v", err)
	}

	// Add new recipient
	err := UpdateRecipientState(tmpDir, "new-agent", StatusWork, false)
	if err != nil {
		t.Fatalf("UpdateRecipientState failed: %v", err)
	}

	// Verify both exist
	recipients, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(recipients) != 2 {
		t.Fatalf("Expected 2 recipients, got %d", len(recipients))
	}

	// Check both recipients are present
	foundExisting := false
	foundNew := false
	for _, r := range recipients {
		if r.Recipient == "existing-agent" {
			foundExisting = true
		}
		if r.Recipient == "new-agent" && r.Status == StatusWork {
			foundNew = true
		}
	}

	if !foundExisting {
		t.Error("existing-agent should still be present")
	}
	if !foundNew {
		t.Error("new-agent should be added")
	}
}

// TestStatusConstants - verify status constants are defined correctly
func TestStatusConstants(t *testing.T) {
	if StatusReady != "ready" {
		t.Errorf("StatusReady should be 'ready', got %s", StatusReady)
	}
	if StatusWork != "work" {
		t.Errorf("StatusWork should be 'work', got %s", StatusWork)
	}
	if StatusOffline != "offline" {
		t.Errorf("StatusOffline should be 'offline', got %s", StatusOffline)
	}
}

// TestRecipientsFile - verify file path constant
func TestRecipientsFile(t *testing.T) {
	expected := ".git/mail-recipients.jsonl"
	if RecipientsFile != expected {
		t.Errorf("RecipientsFile should be '%s', got %s", expected, RecipientsFile)
	}
}

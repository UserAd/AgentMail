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

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
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

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
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
	filePath := filepath.Join(agentmailDir, "recipients.jsonl")
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

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	// Create JSONL with empty lines
	content := `{"recipient":"agent-1","status":"ready","updated_at":"2024-01-01T00:00:00Z","notified":false}

{"recipient":"agent-2","status":"work","updated_at":"2024-01-01T00:00:00Z","notified":true}
`
	filePath := filepath.Join(agentmailDir, "recipients.jsonl")
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

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
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
	filePath := filepath.Join(agentmailDir, "recipients.jsonl")
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

// TestWriteAllRecipients_CreatesParentDir - creates .agentmail dir if missing
func TestWriteAllRecipients_CreatesParentDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Note: .agentmail directory does NOT exist yet

	recipients := []RecipientState{
		{
			Recipient: "agent-1",
			Status:    StatusReady,
			UpdatedAt: time.Now(),
			Notified:  false,
		},
	}

	// Write recipients - should create .agentmail directory
	err := WriteAllRecipients(tmpDir, recipients)
	if err != nil {
		t.Fatalf("WriteAllRecipients failed: %v", err)
	}

	// Verify file exists
	filePath := filepath.Join(tmpDir, ".agentmail", "recipients.jsonl")
	if _, err := os.Stat(filePath); err != nil {
		t.Errorf("File should exist: %v", err)
	}
}

// TestWriteAllRecipients_OverwritesExisting - overwrites existing file
func TestWriteAllRecipients_OverwritesExisting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
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

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
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

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
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

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
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

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
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
	expected := ".agentmail/recipients.jsonl"
	if RecipientsFile != expected {
		t.Errorf("RecipientsFile should be '%s', got %s", expected, RecipientsFile)
	}
}

// =============================================================================
// T048: Tests for ListMailboxRecipients
// =============================================================================

// TestListMailboxRecipients_EmptyDir - returns empty slice when no mailbox files exist
func TestListMailboxRecipients_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail/mailboxes directory
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mailboxes dir: %v", err)
	}

	recipients, err := ListMailboxRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ListMailboxRecipients failed: %v", err)
	}

	if len(recipients) != 0 {
		t.Errorf("Expected 0 recipients for empty dir, got %d", len(recipients))
	}
}

// TestListMailboxRecipients_FindsMailboxes - finds all .jsonl mailbox files
func TestListMailboxRecipients_FindsMailboxes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail/mailboxes directory
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mailboxes dir: %v", err)
	}

	// Create some mailbox files
	for _, name := range []string{"agent-1", "agent-2", "agent-3"} {
		file := filepath.Join(mailDir, name+".jsonl")
		if err := os.WriteFile(file, []byte{}, 0644); err != nil {
			t.Fatalf("Failed to create mailbox file: %v", err)
		}
	}

	recipients, err := ListMailboxRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ListMailboxRecipients failed: %v", err)
	}

	if len(recipients) != 3 {
		t.Errorf("Expected 3 recipients, got %d", len(recipients))
	}

	// Check all names are present
	expected := map[string]bool{"agent-1": false, "agent-2": false, "agent-3": false}
	for _, r := range recipients {
		if _, ok := expected[r]; !ok {
			t.Errorf("Unexpected recipient: %s", r)
		}
		expected[r] = true
	}
	for name, found := range expected {
		if !found {
			t.Errorf("Expected recipient %s not found", name)
		}
	}
}

// TestListMailboxRecipients_IgnoresNonJSONLFiles - only returns .jsonl files
func TestListMailboxRecipients_IgnoresNonJSONLFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail/mailboxes directory
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mailboxes dir: %v", err)
	}

	// Create a .jsonl mailbox file
	mailboxFile := filepath.Join(mailDir, "agent-1.jsonl")
	if err := os.WriteFile(mailboxFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create mailbox file: %v", err)
	}

	// Create non-.jsonl files (PID file, other files)
	pidFile := filepath.Join(mailDir, "mailman.pid")
	if err := os.WriteFile(pidFile, []byte("12345\n"), 0644); err != nil {
		t.Fatalf("Failed to create PID file: %v", err)
	}
	otherFile := filepath.Join(mailDir, "config.json")
	if err := os.WriteFile(otherFile, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create other file: %v", err)
	}

	recipients, err := ListMailboxRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ListMailboxRecipients failed: %v", err)
	}

	if len(recipients) != 1 {
		t.Errorf("Expected 1 recipient (ignoring non-.jsonl files), got %d", len(recipients))
	}

	if len(recipients) > 0 && recipients[0] != "agent-1" {
		t.Errorf("Expected agent-1, got %s", recipients[0])
	}
}

// TestListMailboxRecipients_NoMailDir - returns empty slice when .agentmail/mailboxes doesn't exist
func TestListMailboxRecipients_NoMailDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Don't create .agentmail/mailboxes directory

	recipients, err := ListMailboxRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ListMailboxRecipients should not error for missing dir: %v", err)
	}

	if len(recipients) != 0 {
		t.Errorf("Expected 0 recipients for missing dir, got %d", len(recipients))
	}
}

// =============================================================================
// T049: Tests for CleanStaleStates
// =============================================================================

// TestCleanStaleStates_RemovesOldStates - removes states older than threshold
func TestCleanStaleStates_RemovesOldStates(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	now := time.Now()

	// Create states with different ages
	recipients := []RecipientState{
		{Recipient: "fresh-agent", Status: StatusReady, UpdatedAt: now, Notified: false},
		{Recipient: "old-agent", Status: StatusReady, UpdatedAt: now.Add(-2 * time.Hour), Notified: false},
	}
	if err := WriteAllRecipients(tmpDir, recipients); err != nil {
		t.Fatalf("WriteAllRecipients failed: %v", err)
	}

	// Clean stale states (older than 1 hour)
	err := CleanStaleStates(tmpDir, time.Hour)
	if err != nil {
		t.Fatalf("CleanStaleStates failed: %v", err)
	}

	// Verify old agent was removed
	readBack, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(readBack) != 1 {
		t.Fatalf("Expected 1 recipient after cleanup, got %d", len(readBack))
	}

	if readBack[0].Recipient != "fresh-agent" {
		t.Errorf("Expected fresh-agent to remain, got %s", readBack[0].Recipient)
	}
}

// TestCleanStaleStates_KeepsRecentStates - keeps states newer than threshold
func TestCleanStaleStates_KeepsRecentStates(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	now := time.Now()

	// Create states that are less than 1 hour old
	recipients := []RecipientState{
		{Recipient: "agent-1", Status: StatusReady, UpdatedAt: now, Notified: false},
		{Recipient: "agent-2", Status: StatusWork, UpdatedAt: now.Add(-30 * time.Minute), Notified: false},
		{Recipient: "agent-3", Status: StatusOffline, UpdatedAt: now.Add(-59 * time.Minute), Notified: false},
	}
	if err := WriteAllRecipients(tmpDir, recipients); err != nil {
		t.Fatalf("WriteAllRecipients failed: %v", err)
	}

	// Clean stale states
	err := CleanStaleStates(tmpDir, time.Hour)
	if err != nil {
		t.Fatalf("CleanStaleStates failed: %v", err)
	}

	// Verify all agents remain
	readBack, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(readBack) != 3 {
		t.Errorf("Expected 3 recipients to remain, got %d", len(readBack))
	}
}

// TestCleanStaleStates_EmptyFile - handles empty/non-existent file gracefully
func TestCleanStaleStates_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail directory but no recipients file
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	// Clean stale states on empty/non-existent file
	err := CleanStaleStates(tmpDir, time.Hour)
	if err != nil {
		t.Fatalf("CleanStaleStates should not error on empty file: %v", err)
	}
}

// =============================================================================
// T054: Tests for SetNotifiedFlag
// =============================================================================

// TestSetNotifiedFlag_UpdatesExistingRecipient - updates notified flag for existing recipient
func TestSetNotifiedFlag_UpdatesExistingRecipient(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	now := time.Now()

	// Create a recipient with Notified=false
	recipients := []RecipientState{
		{Recipient: "agent-1", Status: StatusReady, UpdatedAt: now, Notified: false},
	}
	if err := WriteAllRecipients(tmpDir, recipients); err != nil {
		t.Fatalf("WriteAllRecipients failed: %v", err)
	}

	// Set notified flag to true
	err := SetNotifiedFlag(tmpDir, "agent-1", true)
	if err != nil {
		t.Fatalf("SetNotifiedFlag failed: %v", err)
	}

	// Verify
	readBack, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(readBack) != 1 {
		t.Fatalf("Expected 1 recipient, got %d", len(readBack))
	}

	if !readBack[0].Notified {
		t.Error("Expected Notified to be true")
	}
}

// TestSetNotifiedFlag_NoOpForNonExistent - does nothing for non-existent recipient
func TestSetNotifiedFlag_NoOpForNonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	now := time.Now()

	// Create a recipient
	recipients := []RecipientState{
		{Recipient: "agent-1", Status: StatusReady, UpdatedAt: now, Notified: false},
	}
	if err := WriteAllRecipients(tmpDir, recipients); err != nil {
		t.Fatalf("WriteAllRecipients failed: %v", err)
	}

	// Try to set notified flag for non-existent recipient
	err := SetNotifiedFlag(tmpDir, "non-existent", true)
	if err != nil {
		t.Fatalf("SetNotifiedFlag should not error for non-existent recipient: %v", err)
	}

	// Verify agent-1 is unchanged and no new agent was created
	readBack, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(readBack) != 1 {
		t.Errorf("Expected 1 recipient (non-existent not created), got %d", len(readBack))
	}

	if readBack[0].Recipient != "agent-1" {
		t.Errorf("Expected agent-1, got %s", readBack[0].Recipient)
	}
}

// TestSetNotifiedFlag_NoRecipientsFile - handles missing recipients file gracefully
func TestSetNotifiedFlag_NoRecipientsFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail directory but no recipients file
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	// Try to set notified flag when file doesn't exist
	err := SetNotifiedFlag(tmpDir, "agent-1", true)
	if err != nil {
		t.Fatalf("SetNotifiedFlag should not error for missing file: %v", err)
	}
}

// TestSetNotifiedFlag_SetToFalse - can reset notified flag to false
func TestSetNotifiedFlag_SetToFalse(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	now := time.Now()

	// Create a recipient with Notified=true
	recipients := []RecipientState{
		{Recipient: "agent-1", Status: StatusReady, UpdatedAt: now, Notified: true},
	}
	if err := WriteAllRecipients(tmpDir, recipients); err != nil {
		t.Fatalf("WriteAllRecipients failed: %v", err)
	}

	// Set notified flag to false
	err := SetNotifiedFlag(tmpDir, "agent-1", false)
	if err != nil {
		t.Fatalf("SetNotifiedFlag failed: %v", err)
	}

	// Verify
	readBack, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if readBack[0].Notified {
		t.Error("Expected Notified to be false")
	}
}

// =============================================================================
// T031: Unit tests for UpdateLastReadAt
// =============================================================================

func TestUpdateLastReadAt_CreatesFileAndDirectoryIfNotExist(t *testing.T) {
	tmpDir := t.TempDir()

	// Call UpdateLastReadAt - should create directories and file
	timestamp := int64(1704067200000) // 2024-01-01 00:00:00 UTC in milliseconds
	err := UpdateLastReadAt(tmpDir, "new-agent", timestamp)
	if err != nil {
		t.Fatalf("UpdateLastReadAt failed: %v", err)
	}

	// Verify file was created
	filePath := filepath.Join(tmpDir, RecipientsFile)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Recipients file was not created")
	}

	// Verify content
	recipients, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(recipients) != 1 {
		t.Fatalf("Expected 1 recipient, got %d", len(recipients))
	}

	if recipients[0].Recipient != "new-agent" {
		t.Errorf("Expected recipient 'new-agent', got %s", recipients[0].Recipient)
	}

	if recipients[0].LastReadAt != timestamp {
		t.Errorf("Expected LastReadAt %d, got %d", timestamp, recipients[0].LastReadAt)
	}

	// Verify defaults for new entry
	if recipients[0].Status != StatusReady {
		t.Errorf("Expected status 'ready', got %s", recipients[0].Status)
	}
}

func TestUpdateLastReadAt_UpdatesExistingRecipient(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail directory and initial recipient file
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	// Create existing recipient
	now := time.Now().Truncate(time.Second)
	existing := RecipientState{
		Recipient:  "existing-agent",
		Status:     StatusWork,
		UpdatedAt:  now,
		Notified:   true,
		LastReadAt: 1000000000000, // Old timestamp
	}
	filePath := filepath.Join(agentmailDir, "recipients.jsonl")
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	data, _ := json.Marshal(existing)
	file.Write(append(data, '\n'))
	file.Close()

	// Update LastReadAt
	newTimestamp := int64(1704067200000)
	err = UpdateLastReadAt(tmpDir, "existing-agent", newTimestamp)
	if err != nil {
		t.Fatalf("UpdateLastReadAt failed: %v", err)
	}

	// Verify only LastReadAt was updated
	recipients, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(recipients) != 1 {
		t.Fatalf("Expected 1 recipient, got %d", len(recipients))
	}

	if recipients[0].LastReadAt != newTimestamp {
		t.Errorf("Expected LastReadAt %d, got %d", newTimestamp, recipients[0].LastReadAt)
	}

	// Other fields should be preserved
	if recipients[0].Status != StatusWork {
		t.Errorf("Status should be preserved as 'work', got %s", recipients[0].Status)
	}
	if !recipients[0].Notified {
		t.Error("Notified flag should be preserved as true")
	}
}

func TestUpdateLastReadAt_CreatesNewEntryWhenRecipientNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail directory and initial recipient file
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	// Create existing recipient
	now := time.Now().Truncate(time.Second)
	existing := RecipientState{
		Recipient: "existing-agent",
		Status:    StatusReady,
		UpdatedAt: now,
		Notified:  false,
	}
	filePath := filepath.Join(agentmailDir, "recipients.jsonl")
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	data, _ := json.Marshal(existing)
	file.Write(append(data, '\n'))
	file.Close()

	// Update LastReadAt for a different recipient
	timestamp := int64(1704067200000)
	err = UpdateLastReadAt(tmpDir, "new-agent", timestamp)
	if err != nil {
		t.Fatalf("UpdateLastReadAt failed: %v", err)
	}

	// Verify both recipients exist
	recipients, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(recipients) != 2 {
		t.Fatalf("Expected 2 recipients, got %d", len(recipients))
	}

	// Find new recipient
	var newRecipient *RecipientState
	for i := range recipients {
		if recipients[i].Recipient == "new-agent" {
			newRecipient = &recipients[i]
			break
		}
	}

	if newRecipient == nil {
		t.Fatal("New recipient was not created")
	}

	if newRecipient.LastReadAt != timestamp {
		t.Errorf("Expected LastReadAt %d, got %d", timestamp, newRecipient.LastReadAt)
	}
}

func TestUpdateLastReadAt_PreservesOtherRecipients(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.Mkdir(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	// Create multiple recipients
	now := time.Now().Truncate(time.Second)
	filePath := filepath.Join(tmpDir, RecipientsFile)
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	recipients := []RecipientState{
		{Recipient: "agent-1", Status: StatusReady, UpdatedAt: now, Notified: false},
		{Recipient: "agent-2", Status: StatusWork, UpdatedAt: now, Notified: true},
		{Recipient: "agent-3", Status: StatusOffline, UpdatedAt: now, Notified: false},
	}

	for _, r := range recipients {
		data, _ := json.Marshal(r)
		file.Write(append(data, '\n'))
	}
	file.Close()

	// Update agent-2's LastReadAt
	timestamp := int64(1704067200000)
	err = UpdateLastReadAt(tmpDir, "agent-2", timestamp)
	if err != nil {
		t.Fatalf("UpdateLastReadAt failed: %v", err)
	}

	// Verify all recipients are preserved
	readBack, err := ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(readBack) != 3 {
		t.Fatalf("Expected 3 recipients, got %d", len(readBack))
	}

	// Check each recipient
	for _, r := range readBack {
		switch r.Recipient {
		case "agent-1":
			if r.Status != StatusReady {
				t.Errorf("agent-1 status should be ready, got %s", r.Status)
			}
		case "agent-2":
			if r.LastReadAt != timestamp {
				t.Errorf("agent-2 LastReadAt should be %d, got %d", timestamp, r.LastReadAt)
			}
			if r.Status != StatusWork {
				t.Errorf("agent-2 status should be work, got %s", r.Status)
			}
		case "agent-3":
			if r.Status != StatusOffline {
				t.Errorf("agent-3 status should be offline, got %s", r.Status)
			}
		}
	}
}

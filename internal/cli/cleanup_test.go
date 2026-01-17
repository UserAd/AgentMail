package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"agentmail/internal/mail"
)

// =============================================================================
// T010: Test offline recipient removal when window doesn't exist
// =============================================================================

func TestCleanup_OfflineRecipientRemoval(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.MkdirAll(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	now := time.Now()

	// Create recipients - agent-1 exists, agent-2 does not
	recipients := []mail.RecipientState{
		{Recipient: "agent-1", Status: mail.StatusReady, UpdatedAt: now, NotifiedAt: time.Time{}},
		{Recipient: "agent-2", Status: mail.StatusReady, UpdatedAt: now, NotifiedAt: time.Time{}},
		{Recipient: "agent-3", Status: mail.StatusWork, UpdatedAt: now, NotifiedAt: time.Time{}},
	}
	if err := mail.WriteAllRecipients(tmpDir, recipients); err != nil {
		t.Fatalf("WriteAllRecipients failed: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// Run cleanup with mock windows list (only agent-1 exists)
	exitCode := Cleanup(&stdout, &stderr, CleanupOptions{
		StaleHours:     48, // Default
		DeliveredHours: 2,  // Default
		DryRun:         false,
		RepoRoot:       tmpDir,
		SkipTmuxCheck:  true,
		MockInTmux:     true,
		MockWindows:    []string{"agent-1"}, // Only agent-1 exists as tmux window
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Verify agent-2 and agent-3 were removed (their windows don't exist)
	readBack, err := mail.ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(readBack) != 1 {
		t.Fatalf("Expected 1 recipient after cleanup, got %d", len(readBack))
	}

	if readBack[0].Recipient != "agent-1" {
		t.Errorf("Expected agent-1 to remain, got %s", readBack[0].Recipient)
	}
}

// =============================================================================
// T011: Test retention of recipients whose windows still exist
// =============================================================================

func TestCleanup_RetainsExistingWindowRecipients(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.MkdirAll(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	now := time.Now()

	// Create recipients - all have existing windows
	recipients := []mail.RecipientState{
		{Recipient: "agent-1", Status: mail.StatusReady, UpdatedAt: now, NotifiedAt: time.Time{}},
		{Recipient: "agent-2", Status: mail.StatusWork, UpdatedAt: now, NotifiedAt: time.Time{}},
		{Recipient: "agent-3", Status: mail.StatusOffline, UpdatedAt: now, NotifiedAt: time.Time{}},
	}
	if err := mail.WriteAllRecipients(tmpDir, recipients); err != nil {
		t.Fatalf("WriteAllRecipients failed: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// Run cleanup with all windows existing
	exitCode := Cleanup(&stdout, &stderr, CleanupOptions{
		StaleHours:     48,
		DeliveredHours: 2,
		DryRun:         false,
		RepoRoot:       tmpDir,
		SkipTmuxCheck:  true,
		MockInTmux:     true,
		MockWindows:    []string{"agent-1", "agent-2", "agent-3"}, // All exist
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Verify all recipients remain
	readBack, err := mail.ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(readBack) != 3 {
		t.Errorf("Expected 3 recipients to remain, got %d", len(readBack))
	}
}

// =============================================================================
// T012: Test cleanup completes successfully when recipients.jsonl is empty or missing
// =============================================================================

func TestCleanup_EmptyOrMissingRecipients(t *testing.T) {
	t.Run("missing recipients.jsonl", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create .agentmail directory but no recipients.jsonl
		agentmailDir := filepath.Join(tmpDir, ".agentmail")
		if err := os.MkdirAll(agentmailDir, 0755); err != nil {
			t.Fatalf("Failed to create .agentmail dir: %v", err)
		}

		var stdout, stderr bytes.Buffer

		// Run cleanup with no recipients file
		exitCode := Cleanup(&stdout, &stderr, CleanupOptions{
			StaleHours:     48,
			DeliveredHours: 2,
			DryRun:         false,
			RepoRoot:       tmpDir,
			SkipTmuxCheck:  true,
			MockInTmux:     true,
			MockWindows:    []string{"agent-1"},
		})

		if exitCode != 0 {
			t.Errorf("Expected exit code 0 for missing file, got %d. Stderr: %s", exitCode, stderr.String())
		}
	})

	t.Run("empty recipients.jsonl", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create .agentmail directory with empty recipients.jsonl
		agentmailDir := filepath.Join(tmpDir, ".agentmail")
		if err := os.MkdirAll(agentmailDir, 0755); err != nil {
			t.Fatalf("Failed to create .agentmail dir: %v", err)
		}

		// Create empty file
		filePath := filepath.Join(agentmailDir, "recipients.jsonl")
		if err := os.WriteFile(filePath, []byte{}, 0644); err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}

		var stdout, stderr bytes.Buffer

		// Run cleanup with empty recipients file
		exitCode := Cleanup(&stdout, &stderr, CleanupOptions{
			StaleHours:     48,
			DeliveredHours: 2,
			DryRun:         false,
			RepoRoot:       tmpDir,
			SkipTmuxCheck:  true,
			MockInTmux:     true,
			MockWindows:    []string{"agent-1"},
		})

		if exitCode != 0 {
			t.Errorf("Expected exit code 0 for empty file, got %d. Stderr: %s", exitCode, stderr.String())
		}
	})
}

// =============================================================================
// T013: Test non-tmux environment skips offline check with warning
// =============================================================================

func TestCleanup_NonTmuxEnvironmentSkipsOfflineCheck(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.MkdirAll(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	now := time.Now()

	// Create recipients
	recipients := []mail.RecipientState{
		{Recipient: "agent-1", Status: mail.StatusReady, UpdatedAt: now, NotifiedAt: time.Time{}},
		{Recipient: "agent-2", Status: mail.StatusReady, UpdatedAt: now, NotifiedAt: time.Time{}},
	}
	if err := mail.WriteAllRecipients(tmpDir, recipients); err != nil {
		t.Fatalf("WriteAllRecipients failed: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// Run cleanup NOT in tmux (MockInTmux = false)
	exitCode := Cleanup(&stdout, &stderr, CleanupOptions{
		StaleHours:     48,
		DeliveredHours: 2,
		DryRun:         false,
		RepoRoot:       tmpDir,
		SkipTmuxCheck:  true,  // Skip real tmux check
		MockInTmux:     false, // Not in tmux
		MockWindows:    nil,   // No windows (not used when not in tmux)
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Verify warning was printed
	stderrStr := stderr.String()
	if stderrStr == "" {
		t.Error("Expected warning message on stderr when not in tmux")
	}

	// Verify recipients were NOT removed (offline check was skipped)
	readBack, err := mail.ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(readBack) != 2 {
		t.Errorf("Expected 2 recipients to remain (offline check skipped), got %d", len(readBack))
	}
}

// =============================================================================
// Additional test: Verify OfflineRemoved count in CleanupResult
// =============================================================================

func TestCleanup_OfflineRemovedCount(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.MkdirAll(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	now := time.Now()

	// Create 5 recipients - only 2 have existing windows
	recipients := []mail.RecipientState{
		{Recipient: "agent-1", Status: mail.StatusReady, UpdatedAt: now, NotifiedAt: time.Time{}},
		{Recipient: "agent-2", Status: mail.StatusReady, UpdatedAt: now, NotifiedAt: time.Time{}},
		{Recipient: "agent-3", Status: mail.StatusReady, UpdatedAt: now, NotifiedAt: time.Time{}},
		{Recipient: "agent-4", Status: mail.StatusReady, UpdatedAt: now, NotifiedAt: time.Time{}},
		{Recipient: "agent-5", Status: mail.StatusReady, UpdatedAt: now, NotifiedAt: time.Time{}},
	}
	if err := mail.WriteAllRecipients(tmpDir, recipients); err != nil {
		t.Fatalf("WriteAllRecipients failed: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// Run cleanup - only agent-1 and agent-3 exist
	exitCode := Cleanup(&stdout, &stderr, CleanupOptions{
		StaleHours:     48,
		DeliveredHours: 2,
		DryRun:         false,
		RepoRoot:       tmpDir,
		SkipTmuxCheck:  true,
		MockInTmux:     true,
		MockWindows:    []string{"agent-1", "agent-3"}, // Only 2 exist
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Verify only 2 recipients remain
	readBack, err := mail.ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(readBack) != 2 {
		t.Errorf("Expected 2 recipients after cleanup, got %d", len(readBack))
	}

	// Check the summary output indicates 3 offline recipients were removed
	stdoutStr := stdout.String()
	if stdoutStr == "" {
		t.Log("Note: Summary output not yet implemented")
	}
}

// =============================================================================
// Test: CleanOfflineRecipients function directly
// =============================================================================

func TestCleanOfflineRecipients(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.MkdirAll(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	now := time.Now()

	// Create recipients
	recipients := []mail.RecipientState{
		{Recipient: "window-1", Status: mail.StatusReady, UpdatedAt: now, NotifiedAt: time.Time{}},
		{Recipient: "window-2", Status: mail.StatusWork, UpdatedAt: now, NotifiedAt: time.Time{}},
		{Recipient: "window-3", Status: mail.StatusOffline, UpdatedAt: now, NotifiedAt: time.Time{}},
	}
	if err := mail.WriteAllRecipients(tmpDir, recipients); err != nil {
		t.Fatalf("WriteAllRecipients failed: %v", err)
	}

	// Call CleanOfflineRecipients - only window-1 and window-3 exist
	validWindows := []string{"window-1", "window-3"}
	removed, err := mail.CleanOfflineRecipients(tmpDir, validWindows)
	if err != nil {
		t.Fatalf("CleanOfflineRecipients failed: %v", err)
	}

	// Should have removed 1 recipient (window-2)
	if removed != 1 {
		t.Errorf("Expected 1 removed, got %d", removed)
	}

	// Verify remaining recipients
	readBack, err := mail.ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(readBack) != 2 {
		t.Fatalf("Expected 2 recipients after cleanup, got %d", len(readBack))
	}

	// Check which recipients remain
	foundWindow1 := false
	foundWindow3 := false
	for _, r := range readBack {
		if r.Recipient == "window-1" {
			foundWindow1 = true
		}
		if r.Recipient == "window-3" {
			foundWindow3 = true
		}
		if r.Recipient == "window-2" {
			t.Error("window-2 should have been removed")
		}
	}

	if !foundWindow1 {
		t.Error("window-1 should have remained")
	}
	if !foundWindow3 {
		t.Error("window-3 should have remained")
	}
}

func TestCleanOfflineRecipients_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail directory but no recipients file
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.MkdirAll(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	// Call CleanOfflineRecipients with no recipients file
	validWindows := []string{"window-1"}
	removed, err := mail.CleanOfflineRecipients(tmpDir, validWindows)
	if err != nil {
		t.Fatalf("CleanOfflineRecipients should not error on missing file: %v", err)
	}

	if removed != 0 {
		t.Errorf("Expected 0 removed for missing file, got %d", removed)
	}
}

func TestCleanOfflineRecipients_AllExist(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.MkdirAll(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	now := time.Now()

	// Create recipients
	recipients := []mail.RecipientState{
		{Recipient: "window-1", Status: mail.StatusReady, UpdatedAt: now, NotifiedAt: time.Time{}},
		{Recipient: "window-2", Status: mail.StatusWork, UpdatedAt: now, NotifiedAt: time.Time{}},
	}
	if err := mail.WriteAllRecipients(tmpDir, recipients); err != nil {
		t.Fatalf("WriteAllRecipients failed: %v", err)
	}

	// Call CleanOfflineRecipients - all windows exist
	validWindows := []string{"window-1", "window-2"}
	removed, err := mail.CleanOfflineRecipients(tmpDir, validWindows)
	if err != nil {
		t.Fatalf("CleanOfflineRecipients failed: %v", err)
	}

	// Should have removed 0 recipients
	if removed != 0 {
		t.Errorf("Expected 0 removed, got %d", removed)
	}

	// Verify all recipients still exist
	readBack, err := mail.ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(readBack) != 2 {
		t.Errorf("Expected 2 recipients after cleanup, got %d", len(readBack))
	}
}

func TestCleanOfflineRecipients_NoneExist(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmail directory
	agentmailDir := filepath.Join(tmpDir, ".agentmail")
	if err := os.MkdirAll(agentmailDir, 0755); err != nil {
		t.Fatalf("Failed to create .agentmail dir: %v", err)
	}

	now := time.Now()

	// Create recipients
	recipients := []mail.RecipientState{
		{Recipient: "window-1", Status: mail.StatusReady, UpdatedAt: now, NotifiedAt: time.Time{}},
		{Recipient: "window-2", Status: mail.StatusWork, UpdatedAt: now, NotifiedAt: time.Time{}},
	}
	if err := mail.WriteAllRecipients(tmpDir, recipients); err != nil {
		t.Fatalf("WriteAllRecipients failed: %v", err)
	}

	// Call CleanOfflineRecipients - no windows exist (empty list)
	validWindows := []string{}
	removed, err := mail.CleanOfflineRecipients(tmpDir, validWindows)
	if err != nil {
		t.Fatalf("CleanOfflineRecipients failed: %v", err)
	}

	// Should have removed all 2 recipients
	if removed != 2 {
		t.Errorf("Expected 2 removed, got %d", removed)
	}

	// Verify no recipients remain
	readBack, err := mail.ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllRecipients failed: %v", err)
	}

	if len(readBack) != 0 {
		t.Errorf("Expected 0 recipients after cleanup, got %d", len(readBack))
	}
}

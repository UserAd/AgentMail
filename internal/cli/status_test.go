package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agentmail/internal/mail"
)

// T032: Test for status ready command
func TestStatusCommand_Ready(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create git dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	exitCode := Status([]string{"ready"}, &stdout, &stderr, StatusOptions{
		SkipTmuxCheck: true,
		MockWindow:    "agent-1",
		RepoRoot:      tmpDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Status command should be silent on success
	if stdout.Len() > 0 {
		t.Errorf("Expected empty stdout on success, got: %s", stdout.String())
	}
	if stderr.Len() > 0 {
		t.Errorf("Expected empty stderr on success, got: %s", stderr.String())
	}

	// Verify the state was written to recipients file
	recipients, err := mail.ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read recipients: %v", err)
	}

	if len(recipients) != 1 {
		t.Fatalf("Expected 1 recipient, got %d", len(recipients))
	}

	if recipients[0].Recipient != "agent-1" {
		t.Errorf("Expected recipient 'agent-1', got %s", recipients[0].Recipient)
	}
	if recipients[0].Status != mail.StatusReady {
		t.Errorf("Expected status 'ready', got %s", recipients[0].Status)
	}
}

// T033: Test for status work command
func TestStatusCommand_Work(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create git dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	exitCode := Status([]string{"work"}, &stdout, &stderr, StatusOptions{
		SkipTmuxCheck: true,
		MockWindow:    "agent-1",
		RepoRoot:      tmpDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Status command should be silent on success
	if stdout.Len() > 0 {
		t.Errorf("Expected empty stdout on success, got: %s", stdout.String())
	}
	if stderr.Len() > 0 {
		t.Errorf("Expected empty stderr on success, got: %s", stderr.String())
	}

	// Verify the state was written to recipients file
	recipients, err := mail.ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read recipients: %v", err)
	}

	if len(recipients) != 1 {
		t.Fatalf("Expected 1 recipient, got %d", len(recipients))
	}

	if recipients[0].Recipient != "agent-1" {
		t.Errorf("Expected recipient 'agent-1', got %s", recipients[0].Recipient)
	}
	if recipients[0].Status != mail.StatusWork {
		t.Errorf("Expected status 'work', got %s", recipients[0].Status)
	}
}

// T034: Test for status offline command
func TestStatusCommand_Offline(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create git dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	exitCode := Status([]string{"offline"}, &stdout, &stderr, StatusOptions{
		SkipTmuxCheck: true,
		MockWindow:    "agent-1",
		RepoRoot:      tmpDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Status command should be silent on success
	if stdout.Len() > 0 {
		t.Errorf("Expected empty stdout on success, got: %s", stdout.String())
	}
	if stderr.Len() > 0 {
		t.Errorf("Expected empty stderr on success, got: %s", stderr.String())
	}

	// Verify the state was written to recipients file
	recipients, err := mail.ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read recipients: %v", err)
	}

	if len(recipients) != 1 {
		t.Fatalf("Expected 1 recipient, got %d", len(recipients))
	}

	if recipients[0].Recipient != "agent-1" {
		t.Errorf("Expected recipient 'agent-1', got %s", recipients[0].Recipient)
	}
	if recipients[0].Status != mail.StatusOffline {
		t.Errorf("Expected status 'offline', got %s", recipients[0].Status)
	}
}

// T035: Test for non-tmux silent exit (exit 0)
func TestStatusCommand_NotInTmux(t *testing.T) {
	// Use t.Setenv for thread-safe environment variable manipulation
	t.Setenv("TMUX", "")

	var stdout, stderr bytes.Buffer

	exitCode := Status([]string{"ready"}, &stdout, &stderr, StatusOptions{
		SkipTmuxCheck: false, // Don't skip - we want to test the real check
	})

	// Should exit 0 silently (no-op)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 (silent no-op outside tmux), got %d", exitCode)
	}

	// Should produce no output
	if stdout.Len() > 0 {
		t.Errorf("Expected empty stdout when not in tmux, got: %s", stdout.String())
	}
	if stderr.Len() > 0 {
		t.Errorf("Expected empty stderr when not in tmux, got: %s", stderr.String())
	}
}

// T036: Test for invalid status name error (exit 1)
func TestStatusCommand_InvalidStatus(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create git dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	exitCode := Status([]string{"foo"}, &stdout, &stderr, StatusOptions{
		SkipTmuxCheck: true,
		MockWindow:    "agent-1",
		RepoRoot:      tmpDir,
	})

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for invalid status, got %d", exitCode)
	}

	// Should produce error on stderr
	expectedErr := "Invalid status: foo. Valid: ready, work, offline\n"
	if stderr.String() != expectedErr {
		t.Errorf("Expected stderr %q, got: %q", expectedErr, stderr.String())
	}

	// No output on stdout
	if stdout.Len() > 0 {
		t.Errorf("Expected empty stdout on error, got: %s", stdout.String())
	}
}

// Test ValidateStatus function directly
func TestValidateStatus(t *testing.T) {
	tests := []struct {
		status string
		valid  bool
	}{
		{"ready", true},
		{"work", true},
		{"offline", true},
		{"Ready", false},  // case sensitive
		{"WORK", false},   // case sensitive
		{"busy", false},   // invalid
		{"online", false}, // invalid
		{"", false},       // empty
		{"ready ", false}, // trailing space
		{" ready", false}, // leading space
	}

	for _, tt := range tests {
		result := ValidateStatus(tt.status)
		if result != tt.valid {
			t.Errorf("ValidateStatus(%q) = %v, want %v", tt.status, result, tt.valid)
		}
	}
}

// T041: Test notified flag reset on work/offline transition
func TestStatusCommand_NotifiedResetOnWorkOffline(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create git dir: %v", err)
	}

	// First, set status to ready with notified=true
	// We'll manually create the initial state
	initialState := []mail.RecipientState{
		{
			Recipient: "agent-1",
			Status:    mail.StatusReady,
			Notified:  true,
		},
	}
	if err := mail.WriteAllRecipients(tmpDir, initialState); err != nil {
		t.Fatalf("Failed to write initial state: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// Transition to work - should reset notified to false
	exitCode := Status([]string{"work"}, &stdout, &stderr, StatusOptions{
		SkipTmuxCheck: true,
		MockWindow:    "agent-1",
		RepoRoot:      tmpDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Verify notified was reset to false
	recipients, err := mail.ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read recipients: %v", err)
	}

	if len(recipients) != 1 {
		t.Fatalf("Expected 1 recipient, got %d", len(recipients))
	}

	if recipients[0].Notified != false {
		t.Errorf("Expected notified=false after work transition, got %v", recipients[0].Notified)
	}
	if recipients[0].Status != mail.StatusWork {
		t.Errorf("Expected status 'work', got %s", recipients[0].Status)
	}
}

// Test notified flag reset on offline transition
func TestStatusCommand_NotifiedResetOnOffline(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create git dir: %v", err)
	}

	// First, set status with notified=true
	initialState := []mail.RecipientState{
		{
			Recipient: "agent-1",
			Status:    mail.StatusReady,
			Notified:  true,
		},
	}
	if err := mail.WriteAllRecipients(tmpDir, initialState); err != nil {
		t.Fatalf("Failed to write initial state: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// Transition to offline - should reset notified to false
	exitCode := Status([]string{"offline"}, &stdout, &stderr, StatusOptions{
		SkipTmuxCheck: true,
		MockWindow:    "agent-1",
		RepoRoot:      tmpDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Verify notified was reset to false
	recipients, err := mail.ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read recipients: %v", err)
	}

	if len(recipients) != 1 {
		t.Fatalf("Expected 1 recipient, got %d", len(recipients))
	}

	if recipients[0].Notified != false {
		t.Errorf("Expected notified=false after offline transition, got %v", recipients[0].Notified)
	}
}

// Test notified flag NOT reset on ready transition
func TestStatusCommand_NotifiedPreservedOnReady(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create git dir: %v", err)
	}

	// First, set status with notified=true
	initialState := []mail.RecipientState{
		{
			Recipient: "agent-1",
			Status:    mail.StatusWork,
			Notified:  true,
		},
	}
	if err := mail.WriteAllRecipients(tmpDir, initialState); err != nil {
		t.Fatalf("Failed to write initial state: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// Transition to ready - should NOT reset notified
	exitCode := Status([]string{"ready"}, &stdout, &stderr, StatusOptions{
		SkipTmuxCheck: true,
		MockWindow:    "agent-1",
		RepoRoot:      tmpDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Verify notified is preserved (still true)
	recipients, err := mail.ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read recipients: %v", err)
	}

	if len(recipients) != 1 {
		t.Fatalf("Expected 1 recipient, got %d", len(recipients))
	}

	// Notified should be preserved when transitioning to ready
	if recipients[0].Notified != true {
		t.Errorf("Expected notified=true preserved after ready transition, got %v", recipients[0].Notified)
	}
	if recipients[0].Status != mail.StatusReady {
		t.Errorf("Expected status 'ready', got %s", recipients[0].Status)
	}
}

// Test missing status argument
func TestStatusCommand_MissingArgument(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Status([]string{}, &stdout, &stderr, StatusOptions{
		SkipTmuxCheck: true,
		MockWindow:    "agent-1",
	})

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for missing argument, got %d", exitCode)
	}

	// Should produce usage error
	if !strings.Contains(stderr.String(), "usage:") {
		t.Errorf("Expected usage error in stderr, got: %s", stderr.String())
	}
}

// T042: Integration test for status command
func TestStatusCommand_Integration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create git dir: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// Test full workflow: ready -> work -> ready -> offline

	// 1. Set ready
	exitCode := Status([]string{"ready"}, &stdout, &stderr, StatusOptions{
		SkipTmuxCheck: true,
		MockWindow:    "agent-1",
		RepoRoot:      tmpDir,
	})
	if exitCode != 0 {
		t.Errorf("Step 1 (ready): Expected exit code 0, got %d", exitCode)
	}

	recipients, _ := mail.ReadAllRecipients(tmpDir)
	if recipients[0].Status != mail.StatusReady {
		t.Errorf("Step 1: Expected status 'ready', got %s", recipients[0].Status)
	}

	// 2. Set work
	stdout.Reset()
	stderr.Reset()
	exitCode = Status([]string{"work"}, &stdout, &stderr, StatusOptions{
		SkipTmuxCheck: true,
		MockWindow:    "agent-1",
		RepoRoot:      tmpDir,
	})
	if exitCode != 0 {
		t.Errorf("Step 2 (work): Expected exit code 0, got %d", exitCode)
	}

	recipients, _ = mail.ReadAllRecipients(tmpDir)
	if recipients[0].Status != mail.StatusWork {
		t.Errorf("Step 2: Expected status 'work', got %s", recipients[0].Status)
	}

	// 3. Set ready again
	stdout.Reset()
	stderr.Reset()
	exitCode = Status([]string{"ready"}, &stdout, &stderr, StatusOptions{
		SkipTmuxCheck: true,
		MockWindow:    "agent-1",
		RepoRoot:      tmpDir,
	})
	if exitCode != 0 {
		t.Errorf("Step 3 (ready): Expected exit code 0, got %d", exitCode)
	}

	recipients, _ = mail.ReadAllRecipients(tmpDir)
	if recipients[0].Status != mail.StatusReady {
		t.Errorf("Step 3: Expected status 'ready', got %s", recipients[0].Status)
	}

	// 4. Set offline
	stdout.Reset()
	stderr.Reset()
	exitCode = Status([]string{"offline"}, &stdout, &stderr, StatusOptions{
		SkipTmuxCheck: true,
		MockWindow:    "agent-1",
		RepoRoot:      tmpDir,
	})
	if exitCode != 0 {
		t.Errorf("Step 4 (offline): Expected exit code 0, got %d", exitCode)
	}

	recipients, _ = mail.ReadAllRecipients(tmpDir)
	if recipients[0].Status != mail.StatusOffline {
		t.Errorf("Step 4: Expected status 'offline', got %s", recipients[0].Status)
	}
}

// Test multiple agents in recipients file
func TestStatusCommand_MultipleAgents(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentmail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create git dir: %v", err)
	}

	// Pre-populate with another agent
	initialState := []mail.RecipientState{
		{
			Recipient: "agent-2",
			Status:    mail.StatusReady,
			Notified:  false,
		},
	}
	if err := mail.WriteAllRecipients(tmpDir, initialState); err != nil {
		t.Fatalf("Failed to write initial state: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// Add agent-1 status
	exitCode := Status([]string{"work"}, &stdout, &stderr, StatusOptions{
		SkipTmuxCheck: true,
		MockWindow:    "agent-1",
		RepoRoot:      tmpDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Verify both agents exist
	recipients, err := mail.ReadAllRecipients(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read recipients: %v", err)
	}

	if len(recipients) != 2 {
		t.Fatalf("Expected 2 recipients, got %d", len(recipients))
	}

	// Check agent-2 is unchanged
	var agent1, agent2 *mail.RecipientState
	for i := range recipients {
		if recipients[i].Recipient == "agent-1" {
			agent1 = &recipients[i]
		} else if recipients[i].Recipient == "agent-2" {
			agent2 = &recipients[i]
		}
	}

	if agent1 == nil {
		t.Error("agent-1 not found in recipients")
	} else if agent1.Status != mail.StatusWork {
		t.Errorf("Expected agent-1 status 'work', got %s", agent1.Status)
	}

	if agent2 == nil {
		t.Error("agent-2 not found in recipients")
	} else if agent2.Status != mail.StatusReady {
		t.Errorf("Expected agent-2 status 'ready' (unchanged), got %s", agent2.Status)
	}
}

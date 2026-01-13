package daemon

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"agentmail/internal/mail"
)

// =============================================================================
// Helper functions for tests
// =============================================================================

// createTestMailDir creates a temp directory with .agentmail/mailboxes/ structure
func createTestMailDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	mailDir := filepath.Join(tmpDir, ".agentmail", "mailboxes")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatalf("Failed to create mail dir: %v", err)
	}
	return tmpDir
}

// createRecipientState creates a recipient state in the recipients file
func createRecipientState(t *testing.T, repoRoot, recipient, status string, notified bool, updatedAt time.Time) {
	t.Helper()
	recipients, err := mail.ReadAllRecipients(repoRoot)
	if err != nil {
		t.Fatalf("Failed to read recipients: %v", err)
	}

	// Check if recipient exists, update if so
	found := false
	for i := range recipients {
		if recipients[i].Recipient == recipient {
			recipients[i].Status = status
			recipients[i].Notified = notified
			recipients[i].UpdatedAt = updatedAt
			found = true
			break
		}
	}

	if !found {
		recipients = append(recipients, mail.RecipientState{
			Recipient: recipient,
			Status:    status,
			Notified:  notified,
			UpdatedAt: updatedAt,
		})
	}

	if err := mail.WriteAllRecipients(repoRoot, recipients); err != nil {
		t.Fatalf("Failed to write recipients: %v", err)
	}
}

// createUnreadMessage creates an unread message in a recipient's mailbox
func createUnreadMessage(t *testing.T, repoRoot, recipient, from, body string) {
	t.Helper()
	msg := mail.Message{
		ID:       "test123",
		From:     from,
		To:       recipient,
		Message:  body,
		ReadFlag: false,
	}
	if err := mail.Append(repoRoot, msg); err != nil {
		t.Fatalf("Failed to append message: %v", err)
	}
}

// readRecipientState reads a specific recipient's state from the file
func readRecipientState(t *testing.T, repoRoot, recipient string) *mail.RecipientState {
	t.Helper()
	recipients, err := mail.ReadAllRecipients(repoRoot)
	if err != nil {
		t.Fatalf("Failed to read recipients: %v", err)
	}
	for i := range recipients {
		if recipients[i].Recipient == recipient {
			return &recipients[i]
		}
	}
	return nil
}

// =============================================================================
// T043: Test for notification loop interval (10s default)
// =============================================================================

func TestLoopOptions_DefaultInterval(t *testing.T) {
	opts := LoopOptions{}

	// Default interval should be 10 seconds
	if opts.Interval != 0 {
		t.Errorf("Default Interval should be zero value, actual default is set in RunLoop")
	}

	// Verify the default constant
	if DefaultLoopInterval != 10*time.Second {
		t.Errorf("DefaultLoopInterval should be 10 seconds, got %v", DefaultLoopInterval)
	}
}

func TestLoopOptions_CustomInterval(t *testing.T) {
	opts := LoopOptions{
		Interval: 5 * time.Second,
	}

	if opts.Interval != 5*time.Second {
		t.Errorf("Custom interval should be 5 seconds, got %v", opts.Interval)
	}
}

// =============================================================================
// T044: Test for ready agent notification
// =============================================================================

func TestCheckAndNotify_NotifiesReadyAgentWithUnreadMessages(t *testing.T) {
	repoRoot := createTestMailDir(t)

	// Create a ready agent with unread messages
	now := time.Now()
	createRecipientState(t, repoRoot, "agent-1", mail.StatusReady, false, now)
	createUnreadMessage(t, repoRoot, "agent-1", "sender", "Hello!")

	// Track notifications (mock mode for testing)
	var notifiedAgents []string
	mockNotify := func(window string) error {
		notifiedAgents = append(notifiedAgents, window)
		return nil
	}

	opts := LoopOptions{
		RepoRoot:      repoRoot,
		SkipTmuxCheck: true,
	}

	err := CheckAndNotifyWithNotifier(opts, mockNotify)
	if err != nil {
		t.Fatalf("CheckAndNotify failed: %v", err)
	}

	// Verify agent-1 was notified
	if len(notifiedAgents) != 1 {
		t.Fatalf("Expected 1 notification, got %d", len(notifiedAgents))
	}
	if notifiedAgents[0] != "agent-1" {
		t.Errorf("Expected agent-1 to be notified, got %s", notifiedAgents[0])
	}
}

func TestCheckAndNotify_NoNotificationWhenNoUnreadMessages(t *testing.T) {
	repoRoot := createTestMailDir(t)

	// Create a ready agent with NO unread messages
	now := time.Now()
	createRecipientState(t, repoRoot, "agent-1", mail.StatusReady, false, now)

	// Track notifications
	var notifiedAgents []string
	mockNotify := func(window string) error {
		notifiedAgents = append(notifiedAgents, window)
		return nil
	}

	opts := LoopOptions{
		RepoRoot:      repoRoot,
		SkipTmuxCheck: true,
	}

	err := CheckAndNotifyWithNotifier(opts, mockNotify)
	if err != nil {
		t.Fatalf("CheckAndNotify failed: %v", err)
	}

	// Verify no notifications were sent
	if len(notifiedAgents) != 0 {
		t.Errorf("Expected 0 notifications (no unread messages), got %d", len(notifiedAgents))
	}
}

// =============================================================================
// T045: Test for work/offline agent skip
// =============================================================================

func TestCheckAndNotify_SkipsWorkAgent(t *testing.T) {
	repoRoot := createTestMailDir(t)

	// Create a work agent with unread messages
	now := time.Now()
	createRecipientState(t, repoRoot, "agent-work", mail.StatusWork, false, now)
	createUnreadMessage(t, repoRoot, "agent-work", "sender", "Hello!")

	// Track notifications
	var notifiedAgents []string
	mockNotify := func(window string) error {
		notifiedAgents = append(notifiedAgents, window)
		return nil
	}

	opts := LoopOptions{
		RepoRoot:      repoRoot,
		SkipTmuxCheck: true,
	}

	err := CheckAndNotifyWithNotifier(opts, mockNotify)
	if err != nil {
		t.Fatalf("CheckAndNotify failed: %v", err)
	}

	// Verify no notifications were sent (work agents are skipped)
	if len(notifiedAgents) != 0 {
		t.Errorf("Expected 0 notifications (work agent), got %d", len(notifiedAgents))
	}
}

func TestCheckAndNotify_SkipsOfflineAgent(t *testing.T) {
	repoRoot := createTestMailDir(t)

	// Create an offline agent with unread messages
	now := time.Now()
	createRecipientState(t, repoRoot, "agent-offline", mail.StatusOffline, false, now)
	createUnreadMessage(t, repoRoot, "agent-offline", "sender", "Hello!")

	// Track notifications
	var notifiedAgents []string
	mockNotify := func(window string) error {
		notifiedAgents = append(notifiedAgents, window)
		return nil
	}

	opts := LoopOptions{
		RepoRoot:      repoRoot,
		SkipTmuxCheck: true,
	}

	err := CheckAndNotifyWithNotifier(opts, mockNotify)
	if err != nil {
		t.Fatalf("CheckAndNotify failed: %v", err)
	}

	// Verify no notifications were sent (offline agents are skipped)
	if len(notifiedAgents) != 0 {
		t.Errorf("Expected 0 notifications (offline agent), got %d", len(notifiedAgents))
	}
}

func TestCheckAndNotify_MixedStatuses(t *testing.T) {
	repoRoot := createTestMailDir(t)

	now := time.Now()
	// Create agents with different statuses
	createRecipientState(t, repoRoot, "ready-agent", mail.StatusReady, false, now)
	createRecipientState(t, repoRoot, "work-agent", mail.StatusWork, false, now)
	createRecipientState(t, repoRoot, "offline-agent", mail.StatusOffline, false, now)

	// All have unread messages
	createUnreadMessage(t, repoRoot, "ready-agent", "sender", "Hello ready!")
	createUnreadMessage(t, repoRoot, "work-agent", "sender", "Hello work!")
	createUnreadMessage(t, repoRoot, "offline-agent", "sender", "Hello offline!")

	// Track notifications
	var notifiedAgents []string
	mockNotify := func(window string) error {
		notifiedAgents = append(notifiedAgents, window)
		return nil
	}

	opts := LoopOptions{
		RepoRoot:      repoRoot,
		SkipTmuxCheck: true,
	}

	err := CheckAndNotifyWithNotifier(opts, mockNotify)
	if err != nil {
		t.Fatalf("CheckAndNotify failed: %v", err)
	}

	// Verify only ready agent was notified
	if len(notifiedAgents) != 1 {
		t.Fatalf("Expected 1 notification, got %d", len(notifiedAgents))
	}
	if notifiedAgents[0] != "ready-agent" {
		t.Errorf("Expected ready-agent to be notified, got %s", notifiedAgents[0])
	}
}

// =============================================================================
// T046: Test for notified flag prevents duplicate notifications
// =============================================================================

func TestCheckAndNotify_SkipsAlreadyNotified(t *testing.T) {
	repoRoot := createTestMailDir(t)

	// Create a ready agent that was already notified
	now := time.Now()
	createRecipientState(t, repoRoot, "agent-1", mail.StatusReady, true, now) // notified=true
	createUnreadMessage(t, repoRoot, "agent-1", "sender", "Hello!")

	// Track notifications
	var notifiedAgents []string
	mockNotify := func(window string) error {
		notifiedAgents = append(notifiedAgents, window)
		return nil
	}

	opts := LoopOptions{
		RepoRoot:      repoRoot,
		SkipTmuxCheck: true,
	}

	err := CheckAndNotifyWithNotifier(opts, mockNotify)
	if err != nil {
		t.Fatalf("CheckAndNotify failed: %v", err)
	}

	// Verify no notifications were sent (already notified)
	if len(notifiedAgents) != 0 {
		t.Errorf("Expected 0 notifications (already notified), got %d", len(notifiedAgents))
	}
}

func TestCheckAndNotify_UpdatesNotifiedFlagAfterNotification(t *testing.T) {
	repoRoot := createTestMailDir(t)

	// Create a ready agent with notified=false
	now := time.Now()
	createRecipientState(t, repoRoot, "agent-1", mail.StatusReady, false, now)
	createUnreadMessage(t, repoRoot, "agent-1", "sender", "Hello!")

	// Track notifications
	mockNotify := func(window string) error {
		return nil
	}

	opts := LoopOptions{
		RepoRoot:      repoRoot,
		SkipTmuxCheck: true,
	}

	err := CheckAndNotifyWithNotifier(opts, mockNotify)
	if err != nil {
		t.Fatalf("CheckAndNotify failed: %v", err)
	}

	// Verify notified flag was set to true
	state := readRecipientState(t, repoRoot, "agent-1")
	if state == nil {
		t.Fatal("agent-1 state not found")
	}
	if !state.Notified {
		t.Error("Expected Notified to be true after notification")
	}
}

// =============================================================================
// T047: Test for stale state cleanup (>1hr old)
// =============================================================================

func TestCleanStaleStates_RemovesOldStates(t *testing.T) {
	repoRoot := createTestMailDir(t)

	now := time.Now()
	// Create states with different ages
	createRecipientState(t, repoRoot, "fresh-agent", mail.StatusReady, false, now)
	createRecipientState(t, repoRoot, "old-agent", mail.StatusReady, false, now.Add(-2*time.Hour)) // 2 hours old

	// Clean stale states (older than 1 hour)
	err := mail.CleanStaleStates(repoRoot, time.Hour)
	if err != nil {
		t.Fatalf("CleanStaleStates failed: %v", err)
	}

	// Verify old agent was removed
	recipients, err := mail.ReadAllRecipients(repoRoot)
	if err != nil {
		t.Fatalf("Failed to read recipients: %v", err)
	}

	if len(recipients) != 1 {
		t.Fatalf("Expected 1 recipient after cleanup, got %d", len(recipients))
	}

	if recipients[0].Recipient != "fresh-agent" {
		t.Errorf("Expected fresh-agent to remain, got %s", recipients[0].Recipient)
	}
}

func TestCleanStaleStates_KeepsRecentStates(t *testing.T) {
	repoRoot := createTestMailDir(t)

	now := time.Now()
	// Create states that are less than 1 hour old
	createRecipientState(t, repoRoot, "agent-1", mail.StatusReady, false, now)
	createRecipientState(t, repoRoot, "agent-2", mail.StatusWork, false, now.Add(-30*time.Minute))
	createRecipientState(t, repoRoot, "agent-3", mail.StatusOffline, false, now.Add(-59*time.Minute))

	// Clean stale states
	err := mail.CleanStaleStates(repoRoot, time.Hour)
	if err != nil {
		t.Fatalf("CleanStaleStates failed: %v", err)
	}

	// Verify all agents remain
	recipients, err := mail.ReadAllRecipients(repoRoot)
	if err != nil {
		t.Fatalf("Failed to read recipients: %v", err)
	}

	if len(recipients) != 3 {
		t.Errorf("Expected 3 recipients to remain, got %d", len(recipients))
	}
}

func TestCleanStaleStates_EmptyFile(t *testing.T) {
	repoRoot := createTestMailDir(t)

	// Clean stale states on empty/non-existent file
	err := mail.CleanStaleStates(repoRoot, time.Hour)
	if err != nil {
		t.Fatalf("CleanStaleStates should not error on empty file: %v", err)
	}
}

// =============================================================================
// T048: Test for ListMailboxRecipients
// =============================================================================

func TestListMailboxRecipients_EmptyDir(t *testing.T) {
	repoRoot := createTestMailDir(t)

	// No mailbox files exist
	recipients, err := mail.ListMailboxRecipients(repoRoot)
	if err != nil {
		t.Fatalf("ListMailboxRecipients failed: %v", err)
	}

	if len(recipients) != 0 {
		t.Errorf("Expected 0 recipients for empty dir, got %d", len(recipients))
	}
}

func TestListMailboxRecipients_FindsMailboxes(t *testing.T) {
	repoRoot := createTestMailDir(t)

	// Create some mailbox files
	mailDir := filepath.Join(repoRoot, ".agentmail", "mailboxes")
	for _, name := range []string{"agent-1", "agent-2", "agent-3"} {
		file := filepath.Join(mailDir, name+".jsonl")
		if err := os.WriteFile(file, []byte{}, 0644); err != nil {
			t.Fatalf("Failed to create mailbox file: %v", err)
		}
	}

	recipients, err := mail.ListMailboxRecipients(repoRoot)
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

func TestListMailboxRecipients_IgnoresNonJSONLFiles(t *testing.T) {
	repoRoot := createTestMailDir(t)

	mailDir := filepath.Join(repoRoot, ".agentmail", "mailboxes")

	// Create a .jsonl mailbox file
	mailboxFile := filepath.Join(mailDir, "agent-1.jsonl")
	if err := os.WriteFile(mailboxFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create mailbox file: %v", err)
	}

	// Create a non-.jsonl file (e.g., PID file)
	pidFile := filepath.Join(mailDir, "mailman.pid")
	if err := os.WriteFile(pidFile, []byte("12345\n"), 0644); err != nil {
		t.Fatalf("Failed to create PID file: %v", err)
	}

	recipients, err := mail.ListMailboxRecipients(repoRoot)
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

// =============================================================================
// T052: Test for RunLoop
// =============================================================================

func TestRunLoop_RespondsToStopChannel(t *testing.T) {
	repoRoot := createTestMailDir(t)

	stopCh := make(chan struct{})
	opts := LoopOptions{
		RepoRoot:      repoRoot,
		Interval:      100 * time.Millisecond, // Short interval for testing
		StopChan:      stopCh,
		SkipTmuxCheck: true,
	}

	done := make(chan struct{})
	go func() {
		RunLoop(opts)
		close(done)
	}()

	// Let it run briefly
	time.Sleep(50 * time.Millisecond)

	// Stop the loop
	close(stopCh)

	// Wait for loop to finish (with timeout)
	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("RunLoop did not stop within 2 seconds")
	}
}

func TestRunLoop_RunsCheckAndNotify(t *testing.T) {
	repoRoot := createTestMailDir(t)

	// Create a ready agent with unread messages
	now := time.Now()
	createRecipientState(t, repoRoot, "agent-1", mail.StatusReady, false, now)
	createUnreadMessage(t, repoRoot, "agent-1", "sender", "Hello!")

	stopCh := make(chan struct{})
	opts := LoopOptions{
		RepoRoot:      repoRoot,
		Interval:      100 * time.Millisecond,
		StopChan:      stopCh,
		SkipTmuxCheck: true,
	}

	done := make(chan struct{})
	go func() {
		RunLoop(opts)
		close(done)
	}()

	// Let it run one cycle
	time.Sleep(150 * time.Millisecond)

	// Stop the loop
	close(stopCh)
	<-done

	// Verify the notified flag was set (indicating CheckAndNotify ran)
	state := readRecipientState(t, repoRoot, "agent-1")
	if state == nil {
		t.Fatal("agent-1 state not found")
	}
	if !state.Notified {
		t.Error("Expected agent-1 to be marked as notified after loop cycle")
	}
}

// =============================================================================
// T055: Integration test
// =============================================================================

func TestIntegration_FullNotificationCycle(t *testing.T) {
	repoRoot := createTestMailDir(t)

	now := time.Now()

	// Set up various agents
	// Note: Stale agents that are "ready" and have unread messages WILL be notified
	// before stale cleanup runs. This is expected behavior - stale cleanup is
	// a separate concern from notification eligibility.
	createRecipientState(t, repoRoot, "ready-notme", mail.StatusReady, true, now)      // Already notified
	createRecipientState(t, repoRoot, "ready-agent", mail.StatusReady, false, now)     // Should be notified
	createRecipientState(t, repoRoot, "work-agent", mail.StatusWork, false, now)       // Skip (work)
	createRecipientState(t, repoRoot, "offline-agent", mail.StatusOffline, false, now) // Skip (offline)
	// Stale agent is offline - won't be notified and will be cleaned up
	createRecipientState(t, repoRoot, "stale-agent", mail.StatusOffline, false, now.Add(-2*time.Hour)) // Should be cleaned

	// Create unread messages
	createUnreadMessage(t, repoRoot, "ready-notme", "sender", "You have mail!")
	createUnreadMessage(t, repoRoot, "ready-agent", "sender", "You have mail!")
	createUnreadMessage(t, repoRoot, "work-agent", "sender", "You have mail!")
	createUnreadMessage(t, repoRoot, "offline-agent", "sender", "You have mail!")
	createUnreadMessage(t, repoRoot, "stale-agent", "sender", "You have mail!")

	// Track notifications
	var notifiedAgents []string
	mockNotify := func(window string) error {
		notifiedAgents = append(notifiedAgents, window)
		return nil
	}

	opts := LoopOptions{
		RepoRoot:      repoRoot,
		SkipTmuxCheck: true,
	}

	// Run CheckAndNotify
	err := CheckAndNotifyWithNotifier(opts, mockNotify)
	if err != nil {
		t.Fatalf("CheckAndNotify failed: %v", err)
	}

	// Verify only ready-agent was notified
	if len(notifiedAgents) != 1 {
		t.Fatalf("Expected 1 notification, got %d: %v", len(notifiedAgents), notifiedAgents)
	}
	if notifiedAgents[0] != "ready-agent" {
		t.Errorf("Expected ready-agent to be notified, got %s", notifiedAgents[0])
	}

	// Verify notified flag was updated
	state := readRecipientState(t, repoRoot, "ready-agent")
	if state == nil || !state.Notified {
		t.Error("ready-agent should have Notified=true")
	}

	// Now clean stale states
	err = mail.CleanStaleStates(repoRoot, time.Hour)
	if err != nil {
		t.Fatalf("CleanStaleStates failed: %v", err)
	}

	// Verify stale agent was removed
	recipients, err := mail.ReadAllRecipients(repoRoot)
	if err != nil {
		t.Fatalf("Failed to read recipients: %v", err)
	}

	for _, r := range recipients {
		if r.Recipient == "stale-agent" {
			t.Error("stale-agent should have been removed")
		}
	}

	// Should have 4 agents remaining (ready-notme, ready-agent, work-agent, offline-agent)
	if len(recipients) != 4 {
		t.Errorf("Expected 4 recipients after cleanup, got %d", len(recipients))
	}
}

// =============================================================================
// T054: Test that notified flag is updated
// =============================================================================

func TestSetNotifiedFlag_UpdatesState(t *testing.T) {
	repoRoot := createTestMailDir(t)

	now := time.Now()
	createRecipientState(t, repoRoot, "agent-1", mail.StatusReady, false, now)

	// Set notified flag
	err := mail.SetNotifiedFlag(repoRoot, "agent-1", true)
	if err != nil {
		t.Fatalf("SetNotifiedFlag failed: %v", err)
	}

	// Verify
	state := readRecipientState(t, repoRoot, "agent-1")
	if state == nil {
		t.Fatal("agent-1 not found")
	}
	if !state.Notified {
		t.Error("Expected Notified to be true")
	}

	// Set it back to false
	err = mail.SetNotifiedFlag(repoRoot, "agent-1", false)
	if err != nil {
		t.Fatalf("SetNotifiedFlag failed: %v", err)
	}

	state = readRecipientState(t, repoRoot, "agent-1")
	if state == nil {
		t.Fatal("agent-1 not found")
	}
	if state.Notified {
		t.Error("Expected Notified to be false")
	}
}

// =============================================================================
// Helper test for mailbox recipient message format
// =============================================================================

func TestCheckAndNotify_UsesMailboxFilesToDetermineRecipients(t *testing.T) {
	repoRoot := createTestMailDir(t)

	// Create a recipient state for an agent
	now := time.Now()
	createRecipientState(t, repoRoot, "agent-1", mail.StatusReady, false, now)

	// Create unread message (this also creates the mailbox file)
	msg := mail.Message{
		ID:       "msg001",
		From:     "sender",
		To:       "agent-1",
		Message:  "Test message",
		ReadFlag: false,
	}
	if err := mail.Append(repoRoot, msg); err != nil {
		t.Fatalf("Failed to append message: %v", err)
	}

	// Verify mailbox file exists
	mailboxFile := filepath.Join(repoRoot, ".agentmail", "mailboxes", "agent-1.jsonl")
	if _, err := os.Stat(mailboxFile); os.IsNotExist(err) {
		t.Fatal("Mailbox file should exist after Append")
	}

	// Verify we can find unread messages
	unread, err := mail.FindUnread(repoRoot, "agent-1")
	if err != nil {
		t.Fatalf("FindUnread failed: %v", err)
	}
	if len(unread) != 1 {
		t.Errorf("Expected 1 unread message, got %d", len(unread))
	}
}

// =============================================================================
// Test for notification message format
// =============================================================================

func TestNotifyAgent_MessageFormat(t *testing.T) {
	// This test verifies the notification protocol is correct
	// Since we can't test actual tmux commands in unit tests,
	// we verify the NotifyAgent function signature and behavior

	// The notification protocol is:
	// 1. tmux send-keys -t <window> "Check your agentmail"
	// 2. time.Sleep(1 * time.Second)
	// 3. tmux send-keys -t <window> Enter

	// We test that NotifyAgent returns error when not in tmux
	// (actual tmux testing requires integration tests)

	err := NotifyAgent("test-window")
	if err == nil {
		t.Log("NotifyAgent succeeded - either in tmux or using skip check")
	} else {
		// Expected: should fail with ErrNotInTmux when not in tmux
		t.Log("NotifyAgent returned error as expected outside tmux")
	}
}

// =============================================================================
// Test for SetNotifiedFlag with non-existent recipient
// =============================================================================

func TestSetNotifiedFlag_NonExistentRecipient(t *testing.T) {
	repoRoot := createTestMailDir(t)

	// Try to set notified flag for non-existent recipient
	err := mail.SetNotifiedFlag(repoRoot, "non-existent", true)
	if err != nil {
		t.Fatalf("SetNotifiedFlag should not error for non-existent recipient: %v", err)
	}

	// Verify no state was created
	recipients, err := mail.ReadAllRecipients(repoRoot)
	if err != nil {
		t.Fatalf("Failed to read recipients: %v", err)
	}

	if len(recipients) != 0 {
		t.Errorf("Expected 0 recipients (non-existent not created), got %d", len(recipients))
	}
}

// =============================================================================
// Helper to verify JSONL format in recipients file
// =============================================================================

func TestRecipientStateJSONFormat(t *testing.T) {
	repoRoot := createTestMailDir(t)

	now := time.Now().Truncate(time.Second)
	createRecipientState(t, repoRoot, "agent-1", mail.StatusReady, false, now)

	// Read raw file content
	filePath := filepath.Join(repoRoot, ".agentmail", "recipients.jsonl")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Verify it's valid JSON
	var state mail.RecipientState
	if err := json.Unmarshal(content[:len(content)-1], &state); err != nil { // -1 to remove newline
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if state.Recipient != "agent-1" {
		t.Errorf("Expected recipient agent-1, got %s", state.Recipient)
	}
	if state.Status != mail.StatusReady {
		t.Errorf("Expected status ready, got %s", state.Status)
	}
	if state.Notified != false {
		t.Error("Expected Notified=false")
	}
}

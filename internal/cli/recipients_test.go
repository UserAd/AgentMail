package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// T008: Unit tests for Recipients() in internal/cli/recipients_test.go
// T009: Test that lists all windows one per line
// T010: Test that marks current window with "[you]" suffix
// T011: Test that returns exit code 2 when not in tmux
//
// Expected function signature in recipients.go:
//
//	func Recipients(stdout, stderr io.Writer, opts RecipientsOptions) int
//
// Expected RecipientsOptions struct in recipients.go:
//
//	type RecipientsOptions struct {
//	    SkipTmuxCheck bool     // Skip tmux environment check
//	    MockWindows   []string // Mock list of tmux windows
//	    MockCurrent   string   // Mock current window name
//	}

// T009: Test that lists all windows one per line
func TestRecipientsCommand_ListsAllWindows(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Recipients(&stdout, &stderr, RecipientsOptions{
		SkipTmuxCheck:   true,
		MockPanes:       []string{"main", "agent1", "agent2", "worker"},
		MockPaneAddress: "main",
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	if stderr.String() != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr.String())
	}

	output := stdout.String()
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")

	if len(lines) != 4 {
		t.Errorf("Expected 4 lines (one per window), got %d: %v", len(lines), lines)
	}

	// Verify each window appears in output (order may vary, but should contain all windows)
	expectedWindows := map[string]bool{"main": false, "agent1": false, "agent2": false, "worker": false}
	for _, line := range lines {
		// Remove "[you]" suffix if present for comparison
		windowName := strings.TrimSuffix(line, " [you]")
		if _, ok := expectedWindows[windowName]; ok {
			expectedWindows[windowName] = true
		}
	}

	for window, found := range expectedWindows {
		if !found {
			t.Errorf("Expected window '%s' in output, got: %s", window, output)
		}
	}
}

// T010: Test that marks current window with "[you]" suffix
func TestRecipientsCommand_MarksCurrentWindow(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Recipients(&stdout, &stderr, RecipientsOptions{
		SkipTmuxCheck:   true,
		MockPanes:       []string{"main", "agent1", "agent2"},
		MockPaneAddress: "agent1",
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	output := stdout.String()

	// Check that current window has "[you]" suffix
	if !strings.Contains(output, "agent1 [you]") {
		t.Errorf("Expected current window 'agent1' to have '[you]' suffix, got: %s", output)
	}

	// Check that other windows do NOT have "[you]" suffix
	if strings.Contains(output, "main [you]") {
		t.Errorf("Window 'main' should not have '[you]' suffix, got: %s", output)
	}
	if strings.Contains(output, "agent2 [you]") {
		t.Errorf("Window 'agent2' should not have '[you]' suffix, got: %s", output)
	}
}

// T011: Test that returns exit code 2 when not in tmux
func TestRecipientsCommand_NotInTmux(t *testing.T) {
	// Use t.Setenv for thread-safe environment variable manipulation
	t.Setenv("TMUX", "")

	var stdout, stderr bytes.Buffer

	exitCode := Recipients(&stdout, &stderr, RecipientsOptions{
		// Don't skip tmux check - we want to test it
		SkipTmuxCheck: false,
	})

	if exitCode != 2 {
		t.Errorf("Expected exit code 2 (not in tmux), got %d", exitCode)
	}

	if stdout.String() != "" {
		t.Errorf("Expected empty stdout when not in tmux, got: %s", stdout.String())
	}

	expectedError := "error: not running inside a tmux session\n"
	if stderr.String() != expectedError {
		t.Errorf("Expected stderr '%s', got: '%s'", expectedError, stderr.String())
	}
}

// Additional test: Empty window list
func TestRecipientsCommand_EmptyWindowList(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Recipients(&stdout, &stderr, RecipientsOptions{
		SkipTmuxCheck:   true,
		MockPanes:       []string{},
		MockPaneAddress: "AgentMail:test.0",
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	if stderr.String() != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr.String())
	}

	output := stdout.String()
	if output != "" {
		t.Errorf("Expected empty output for no windows, got: %s", output)
	}
}

// Additional test: Single window (only current - should show "[you]")
func TestRecipientsCommand_SingleWindow(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Recipients(&stdout, &stderr, RecipientsOptions{
		SkipTmuxCheck:   true,
		MockPanes:       []string{"main"},
		MockPaneAddress: "main",
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	if stderr.String() != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr.String())
	}

	expectedOutput := "main [you]\n"
	if stdout.String() != expectedOutput {
		t.Errorf("Expected output '%s', got: '%s'", expectedOutput, stdout.String())
	}
}

// Additional test: Current window not in list (edge case)
func TestRecipientsCommand_CurrentWindowNotInList(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Recipients(&stdout, &stderr, RecipientsOptions{
		SkipTmuxCheck:   true,
		MockPanes:       []string{"agent1", "agent2"},
		MockPaneAddress: "orphan",
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	output := stdout.String()

	// Should list windows without "[you]" since current window is not in list
	if strings.Contains(output, "[you]") {
		t.Errorf("Expected no '[you]' marker since current window not in list, got: %s", output)
	}

	// Both windows should still appear
	if !strings.Contains(output, "agent1") {
		t.Errorf("Expected 'agent1' in output, got: %s", output)
	}
	if !strings.Contains(output, "agent2") {
		t.Errorf("Expected 'agent2' in output, got: %s", output)
	}
}

// Test: Output format verification (one window per line with newline)
func TestRecipientsCommand_OutputFormat(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Recipients(&stdout, &stderr, RecipientsOptions{
		SkipTmuxCheck:   true,
		MockPanes:       []string{"main", "agent1"},
		MockPaneAddress: "main",
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	output := stdout.String()

	// Verify output ends with newline
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("Expected output to end with newline, got: %q", output)
	}

	// Verify each line is a window name (with optional [you] suffix)
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSuffix(line, " [you]")
		if trimmed != "main" && trimmed != "agent1" {
			t.Errorf("Unexpected line in output: %s", line)
		}
	}
}

// T018: Test that excludes windows listed in .agentmailignore
func TestRecipientsCommand_ExcludesIgnoredWindows(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Recipients(&stdout, &stderr, RecipientsOptions{
		SkipTmuxCheck:   true,
		MockPanes:       []string{"main", "agent1", "agent2", "worker"},
		MockPaneAddress: "main",
		MockIgnoreList:  map[string]bool{"agent1": true, "worker": true},
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	output := stdout.String()

	// Check that ignored windows are NOT in output
	if strings.Contains(output, "agent1") {
		t.Errorf("Ignored window 'agent1' should not appear in output, got: %s", output)
	}
	if strings.Contains(output, "worker") {
		t.Errorf("Ignored window 'worker' should not appear in output, got: %s", output)
	}

	// Check that non-ignored windows ARE in output
	if !strings.Contains(output, "main [you]") {
		t.Errorf("Current window 'main' should appear with [you], got: %s", output)
	}
	if !strings.Contains(output, "agent2") {
		t.Errorf("Non-ignored window 'agent2' should appear in output, got: %s", output)
	}

	// Verify only 2 lines in output
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines (main and agent2), got %d: %v", len(lines), lines)
	}
}

// T019: Test that handles missing .agentmailignore gracefully (shows all windows)
func TestRecipientsCommand_HandlesMissingIgnoreFile(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// Use a mock git root that doesn't have an ignore file
	tempDir := t.TempDir()

	exitCode := Recipients(&stdout, &stderr, RecipientsOptions{
		SkipTmuxCheck:   true,
		MockPanes:       []string{"main", "agent1", "agent2"},
		MockPaneAddress: "main",
		MockGitRoot:     tempDir, // Directory without .agentmailignore
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	output := stdout.String()

	// All windows should be shown when ignore file is missing
	if !strings.Contains(output, "main [you]") {
		t.Errorf("Expected 'main [you]' in output, got: %s", output)
	}
	if !strings.Contains(output, "agent1") {
		t.Errorf("Expected 'agent1' in output, got: %s", output)
	}
	if !strings.Contains(output, "agent2") {
		t.Errorf("Expected 'agent2' in output, got: %s", output)
	}

	// Verify 3 lines in output
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d: %v", len(lines), lines)
	}
}

// T020: Test that ignores empty and whitespace-only lines in ignore file
func TestRecipientsCommand_IgnoresEmptyLinesInIgnoreFile(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// Create a temp directory with an ignore file containing empty lines
	tempDir := t.TempDir()
	ignoreContent := "agent1\n\n  \n\t\nworker\n"
	if err := os.WriteFile(tempDir+"/.agentmailignore", []byte(ignoreContent), 0o644); err != nil {
		t.Fatalf("Failed to create ignore file: %v", err)
	}

	exitCode := Recipients(&stdout, &stderr, RecipientsOptions{
		SkipTmuxCheck:   true,
		MockPanes:       []string{"main", "agent1", "agent2", "worker"},
		MockPaneAddress: "main",
		MockGitRoot:     tempDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	output := stdout.String()

	// Check that explicitly ignored windows are NOT in output
	if strings.Contains(output, "agent1\n") || strings.Contains(output, "agent1 [you]") {
		// agent1 could appear if it's the current window, but here main is current
		if !strings.Contains(output, "agent1 [you]") && strings.Contains(output, "agent1") {
			t.Errorf("Ignored window 'agent1' should not appear in output, got: %s", output)
		}
	}
	if strings.Contains(output, "worker") {
		t.Errorf("Ignored window 'worker' should not appear in output, got: %s", output)
	}

	// Check that non-ignored windows ARE in output
	if !strings.Contains(output, "main [you]") {
		t.Errorf("Current window 'main' should appear with [you], got: %s", output)
	}
	if !strings.Contains(output, "agent2") {
		t.Errorf("Non-ignored window 'agent2' should appear in output, got: %s", output)
	}
}

// T021: Test that handles unreadable ignore file (per FR-013)
func TestRecipientsCommand_HandlesUnreadableIgnoreFile(t *testing.T) {
	// Skip on systems where root can read any file (e.g., GitHub Actions containers)
	if os.Getuid() == 0 {
		t.Skip("Skipping unreadable file test when running as root")
	}

	var stdout, stderr bytes.Buffer

	// Create a temp directory with an unreadable ignore file
	tempDir := t.TempDir()
	ignorePath := tempDir + "/.agentmailignore"
	if err := os.WriteFile(ignorePath, []byte("agent1\n"), 0o000); err != nil {
		t.Fatalf("Failed to create ignore file: %v", err)
	}
	// Ensure cleanup even if test fails
	defer os.Chmod(ignorePath, 0o644)

	exitCode := Recipients(&stdout, &stderr, RecipientsOptions{
		SkipTmuxCheck:   true,
		MockPanes:       []string{"main", "agent1", "agent2"},
		MockPaneAddress: "main",
		MockGitRoot:     tempDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	output := stdout.String()

	// When ignore file is unreadable, all windows should be shown (per FR-013)
	if !strings.Contains(output, "main [you]") {
		t.Errorf("Expected 'main [you]' in output, got: %s", output)
	}
	if !strings.Contains(output, "agent1") {
		t.Errorf("Expected 'agent1' in output when ignore file unreadable, got: %s", output)
	}
	if !strings.Contains(output, "agent2") {
		t.Errorf("Expected 'agent2' in output, got: %s", output)
	}
}

// T024: Test that current window is shown with "[you]" even if in ignore list (per FR-004)
func TestRecipientsCommand_CurrentWindowShownEvenIfIgnored(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Recipients(&stdout, &stderr, RecipientsOptions{
		SkipTmuxCheck:   true,
		MockPanes:       []string{"main", "agent1", "agent2"},
		MockPaneAddress: "agent1", // Current window is in ignore list
		MockIgnoreList:  map[string]bool{"agent1": true, "agent2": true},
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	output := stdout.String()

	// Current window should ALWAYS be shown with [you], even if in ignore list
	if !strings.Contains(output, "agent1 [you]") {
		t.Errorf("Current window 'agent1' should appear with [you] even if in ignore list, got: %s", output)
	}

	// Other ignored windows should NOT be shown
	if strings.Contains(output, "agent2") {
		t.Errorf("Ignored window 'agent2' should not appear in output, got: %s", output)
	}

	// Non-ignored windows should be shown
	if !strings.Contains(output, "main") {
		t.Errorf("Non-ignored window 'main' should appear in output, got: %s", output)
	}

	// Verify only 2 lines in output (main and agent1 [you])
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines (main and agent1 [you]), got %d: %v", len(lines), lines)
	}
}

func TestRecipientsCommand_EmptyPaneList(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := Recipients(
		&stdout,
		&stderr,
		RecipientsOptions{
			SkipTmuxCheck:   true,
			MockPanes:       []string{}, // No panes
			MockPaneAddress: "mysession:agent.0",
			MockSession:     "mysession",
		},
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	output := stdout.String()
	// With an empty pane list, the loop has nothing to iterate over,
	// so no output is produced (the current pane only appears if it's in the pane list)
	if output != "" {
		t.Errorf("Expected empty output with no panes, got: %s", output)
	}
}

func TestRecipientsCommand_NoIgnoreList(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := Recipients(
		&stdout,
		&stderr,
		RecipientsOptions{
			SkipTmuxCheck:   true,
			MockPanes:       []string{"mysession:agent1.0", "mysession:agent2.0", "mysession:agent3.0"},
			MockPaneAddress: "mysession:agent1.0",
			MockSession:     "mysession",
			MockIgnoreList:  nil, // No ignore list
		},
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	output := stdout.String()

	// All panes should be shown
	if !strings.Contains(output, "agent2") {
		t.Errorf("Expected agent2 in output, got: %s", output)
	}
	if !strings.Contains(output, "agent3") {
		t.Errorf("Expected agent3 in output, got: %s", output)
	}
}

func TestRecipientsCommand_IgnoreListFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agentmailignore file
	ignoreContent := `# Test ignore file
mysession:agent2.0
:agent3.0
`
	gitRoot := tmpDir
	ignorePath := filepath.Join(gitRoot, ".agentmailignore")
	if err := os.WriteFile(ignorePath, []byte(ignoreContent), 0644); err != nil {
		t.Fatalf("Failed to create ignore file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	exitCode := Recipients(
		&stdout,
		&stderr,
		RecipientsOptions{
			SkipTmuxCheck:   true,
			MockPanes:       []string{"mysession:agent1.0", "mysession:agent2.0", "mysession:agent3.0"},
			MockPaneAddress: "mysession:agent1.0",
			MockSession:     "mysession",
			MockGitRoot:     gitRoot,
			MockIgnoreList:  nil, // Load from file
		},
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	output := stdout.String()

	// agent2 and agent3 should be ignored
	if strings.Contains(output, "agent2") {
		t.Errorf("agent2 should be ignored, got: %s", output)
	}
	if strings.Contains(output, "agent3") {
		t.Errorf("agent3 should be ignored, got: %s", output)
	}

	// agent1 (current) should still be shown
	if !strings.Contains(output, "agent1") {
		t.Errorf("Current agent should be shown, got: %s", output)
	}
}

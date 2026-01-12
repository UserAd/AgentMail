package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// Test basic onboarding output with multiple agents
func TestOnboardCommand_BasicOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Onboard(&stdout, &stderr, OnboardOptions{
		SkipTmuxCheck: true,
		MockWindows:   []string{"main", "agent1", "agent2"},
		MockCurrent:   "main",
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	if stderr.String() != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr.String())
	}

	output := stdout.String()

	// Check header
	if !strings.Contains(output, "## AgentMail") {
		t.Errorf("Expected '## AgentMail' header, got: %s", output)
	}

	// Check agent identity
	if !strings.Contains(output, "You are **main**") {
		t.Errorf("Expected agent identity 'You are **main**', got: %s", output)
	}

	// Check other agents
	if !strings.Contains(output, "Other agents:") {
		t.Errorf("Expected 'Other agents:' section, got: %s", output)
	}
	if !strings.Contains(output, "agent1") {
		t.Errorf("Expected 'agent1' in other agents, got: %s", output)
	}
	if !strings.Contains(output, "agent2") {
		t.Errorf("Expected 'agent2' in other agents, got: %s", output)
	}

	// Check command reference
	if !strings.Contains(output, "### Commands") {
		t.Errorf("Expected '### Commands' section, got: %s", output)
	}
	if !strings.Contains(output, "**send**") {
		t.Errorf("Expected '**send**' command reference, got: %s", output)
	}
	if !strings.Contains(output, "**receive**") {
		t.Errorf("Expected '**receive**' command reference, got: %s", output)
	}
	if !strings.Contains(output, "**recipients**") {
		t.Errorf("Expected '**recipients**' command reference, got: %s", output)
	}
	// Check examples are included
	if !strings.Contains(output, "Example:") {
		t.Errorf("Expected examples in output, got: %s", output)
	}
}

// Test that current window is excluded from other agents list
func TestOnboardCommand_ExcludesCurrentFromOthers(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Onboard(&stdout, &stderr, OnboardOptions{
		SkipTmuxCheck: true,
		MockWindows:   []string{"main", "agent1", "agent2"},
		MockCurrent:   "agent1",
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	output := stdout.String()

	// Current agent should be shown as identity
	if !strings.Contains(output, "You are **agent1**") {
		t.Errorf("Expected agent identity 'You are **agent1**', got: %s", output)
	}

	// Other agents should include main and agent2, but NOT agent1
	if !strings.Contains(output, "main") {
		t.Errorf("Expected 'main' in other agents, got: %s", output)
	}
	if !strings.Contains(output, "agent2") {
		t.Errorf("Expected 'agent2' in other agents, got: %s", output)
	}

	// The "Other agents:" line should not list agent1
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Other agents:") {
			if strings.Contains(line, "agent1") {
				t.Errorf("Current agent 'agent1' should not be in other agents list: %s", line)
			}
		}
	}
}

// Test silent exit when not in tmux (hook-friendly behavior)
func TestOnboardCommand_NotInTmux(t *testing.T) {
	t.Setenv("TMUX", "")

	var stdout, stderr bytes.Buffer

	exitCode := Onboard(&stdout, &stderr, OnboardOptions{
		SkipTmuxCheck: false, // Don't skip - test real check
	})

	// Should exit 0 silently (hook-friendly)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 (silent no-op outside tmux), got %d", exitCode)
	}

	if stdout.String() != "" {
		t.Errorf("Expected empty stdout when not in tmux, got: %s", stdout.String())
	}

	if stderr.String() != "" {
		t.Errorf("Expected empty stderr when not in tmux, got: %s", stderr.String())
	}
}

// Test no other agents available
func TestOnboardCommand_NoOtherAgents(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Onboard(&stdout, &stderr, OnboardOptions{
		SkipTmuxCheck: true,
		MockWindows:   []string{"main"},
		MockCurrent:   "main",
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	output := stdout.String()

	if !strings.Contains(output, "No other agents currently available") {
		t.Errorf("Expected 'No other agents currently available' message, got: %s", output)
	}
}

// Test that ignored windows are excluded from other agents
func TestOnboardCommand_ExcludesIgnoredWindows(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Onboard(&stdout, &stderr, OnboardOptions{
		SkipTmuxCheck:  true,
		MockWindows:    []string{"main", "agent1", "agent2", "worker"},
		MockCurrent:    "main",
		MockIgnoreList: map[string]bool{"agent1": true, "worker": true},
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	output := stdout.String()

	// Ignored windows should NOT appear in other agents
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Other agents:") {
			if strings.Contains(line, "agent1") {
				t.Errorf("Ignored window 'agent1' should not be in other agents: %s", line)
			}
			if strings.Contains(line, "worker") {
				t.Errorf("Ignored window 'worker' should not be in other agents: %s", line)
			}
			// Non-ignored window should be present
			if !strings.Contains(line, "agent2") {
				t.Errorf("Non-ignored window 'agent2' should be in other agents: %s", line)
			}
		}
	}
}

// Test with ignore file from mock git root
func TestOnboardCommand_LoadsIgnoreFile(t *testing.T) {
	tempDir := t.TempDir()
	ignoreContent := "agent1\nworker\n"
	if err := os.WriteFile(tempDir+"/.agentmailignore", []byte(ignoreContent), 0o644); err != nil {
		t.Fatalf("Failed to create ignore file: %v", err)
	}

	var stdout, stderr bytes.Buffer

	exitCode := Onboard(&stdout, &stderr, OnboardOptions{
		SkipTmuxCheck: true,
		MockWindows:   []string{"main", "agent1", "agent2", "worker"},
		MockCurrent:   "main",
		MockGitRoot:   tempDir,
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	output := stdout.String()

	// Check that only non-ignored agents appear in "Other agents:"
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Other agents:") {
			if strings.Contains(line, "agent1") {
				t.Errorf("Ignored window 'agent1' should not be in other agents: %s", line)
			}
			if strings.Contains(line, "worker") {
				t.Errorf("Ignored window 'worker' should not be in other agents: %s", line)
			}
			if !strings.Contains(line, "agent2") {
				t.Errorf("Non-ignored window 'agent2' should be in other agents: %s", line)
			}
		}
	}
}

// Test empty window list edge case
func TestOnboardCommand_EmptyWindowList(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Onboard(&stdout, &stderr, OnboardOptions{
		SkipTmuxCheck: true,
		MockWindows:   []string{},
		MockCurrent:   "",
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	output := stdout.String()

	// Should still output header and commands
	if !strings.Contains(output, "## AgentMail") {
		t.Errorf("Expected header even with empty window list, got: %s", output)
	}
	if !strings.Contains(output, "No other agents currently available") {
		t.Errorf("Expected 'No other agents' message, got: %s", output)
	}
}

// Test output format is markdown-friendly
func TestOnboardCommand_MarkdownFormat(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Onboard(&stdout, &stderr, OnboardOptions{
		SkipTmuxCheck: true,
		MockWindows:   []string{"main", "agent1"},
		MockCurrent:   "main",
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	output := stdout.String()

	// Check markdown formatting
	if !strings.Contains(output, "## ") {
		t.Errorf("Expected markdown h2 header (## ), got: %s", output)
	}
	if !strings.Contains(output, "### ") {
		t.Errorf("Expected markdown h3 header (### ), got: %s", output)
	}
	if !strings.Contains(output, "**") {
		t.Errorf("Expected markdown bold (**), got: %s", output)
	}
	// Check for code blocks
	if !strings.Contains(output, "```") {
		t.Errorf("Expected markdown code blocks (```), got: %s", output)
	}
}

// Test that send example uses first available agent
func TestOnboardCommand_DynamicSendExample(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Onboard(&stdout, &stderr, OnboardOptions{
		SkipTmuxCheck: true,
		MockWindows:   []string{"main", "worker", "agent1"},
		MockCurrent:   "main",
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	output := stdout.String()

	// The send example should use the first other agent (worker)
	if !strings.Contains(output, "agentmail send worker") {
		t.Errorf("Expected send example to use 'worker' (first other agent), got: %s", output)
	}
}

// Test all windows ignored except current
func TestOnboardCommand_AllOthersIgnored(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Onboard(&stdout, &stderr, OnboardOptions{
		SkipTmuxCheck:  true,
		MockWindows:    []string{"main", "agent1", "agent2"},
		MockCurrent:    "main",
		MockIgnoreList: map[string]bool{"agent1": true, "agent2": true},
	})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	output := stdout.String()

	// When all other agents are ignored, should show "No other agents" message
	if !strings.Contains(output, "No other agents currently available") {
		t.Errorf("Expected 'No other agents' when all others ignored, got: %s", output)
	}
}

package cli

import (
	"bytes"
	"strings"
	"testing"
)

// Test basic onboarding output
func TestOnboardCommand_BasicOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Onboard(&stdout, &stderr, OnboardOptions{})

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

	// Check description
	if !strings.Contains(output, "AgentMail enables inter-agent communication") {
		t.Errorf("Expected description, got: %s", output)
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

// Test exit code is always 0
func TestOnboardCommand_ExitCode(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Onboard(&stdout, &stderr, OnboardOptions{})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

// Test output format is markdown-friendly
func TestOnboardCommand_MarkdownFormat(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Onboard(&stdout, &stderr, OnboardOptions{})

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

// Test send example is present
func TestOnboardCommand_SendExample(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Onboard(&stdout, &stderr, OnboardOptions{})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	output := stdout.String()

	if !strings.Contains(output, "agentmail send") {
		t.Errorf("Expected send example, got: %s", output)
	}
}

// Test receive example is present
func TestOnboardCommand_ReceiveExample(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Onboard(&stdout, &stderr, OnboardOptions{})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	output := stdout.String()

	if !strings.Contains(output, "agentmail receive") {
		t.Errorf("Expected receive example, got: %s", output)
	}
}

// Test recipients example is present
func TestOnboardCommand_RecipientsExample(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Onboard(&stdout, &stderr, OnboardOptions{})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	output := stdout.String()

	if !strings.Contains(output, "agentmail recipients") {
		t.Errorf("Expected recipients example, got: %s", output)
	}
}

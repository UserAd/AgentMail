package cli

import (
	"bytes"
	"strings"
	"testing"
)

// T032: Unit tests for Help() in internal/cli/help_test.go
// T033: Test that help output includes send command with syntax
// T034: Test that help output includes receive command with syntax
// T035: Test that help output includes recipients command with syntax
// T036: Test that help output includes examples section
//
// Expected function signature in help.go:
//
//	func Help(stdout io.Writer) int

// T032: Basic unit test for Help() function
func TestHelpCommand_ReturnsExitCode0(t *testing.T) {
	var stdout bytes.Buffer

	exitCode := Help(&stdout)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

// T032: Test that Help() writes to stdout
func TestHelpCommand_WritesToStdout(t *testing.T) {
	var stdout bytes.Buffer

	Help(&stdout)

	if stdout.Len() == 0 {
		t.Error("Expected Help() to write to stdout, got empty output")
	}
}

// T033: Test that help output includes send command with syntax
func TestHelpCommand_IncludesSendCommand(t *testing.T) {
	var stdout bytes.Buffer

	Help(&stdout)

	output := stdout.String()

	// Check for send command presence
	if !strings.Contains(output, "send") {
		t.Errorf("Expected help output to include 'send' command, got: %s", output)
	}

	// Check for send syntax with recipient and message
	if !strings.Contains(output, "send <recipient>") {
		t.Errorf("Expected help output to include 'send <recipient>' syntax, got: %s", output)
	}

	// Check for send description
	if !strings.Contains(output, "Send a message") {
		t.Errorf("Expected help output to include send description, got: %s", output)
	}
}

// T034: Test that help output includes receive command with syntax
func TestHelpCommand_IncludesReceiveCommand(t *testing.T) {
	var stdout bytes.Buffer

	Help(&stdout)

	output := stdout.String()

	// Check for receive command presence
	if !strings.Contains(output, "receive") {
		t.Errorf("Expected help output to include 'receive' command, got: %s", output)
	}

	// Check for receive description
	if !strings.Contains(output, "Read the oldest unread message") {
		t.Errorf("Expected help output to include receive description, got: %s", output)
	}
}

// T035: Test that help output includes recipients command with syntax
func TestHelpCommand_IncludesRecipientsCommand(t *testing.T) {
	var stdout bytes.Buffer

	Help(&stdout)

	output := stdout.String()

	// Check for recipients command presence
	if !strings.Contains(output, "recipients") {
		t.Errorf("Expected help output to include 'recipients' command, got: %s", output)
	}

	// Check for recipients description
	if !strings.Contains(output, "List available message recipients") {
		t.Errorf("Expected help output to include recipients description, got: %s", output)
	}
}

// T036: Test that help output includes examples section
func TestHelpCommand_IncludesExamplesSection(t *testing.T) {
	var stdout bytes.Buffer

	Help(&stdout)

	output := stdout.String()

	// Check for EXAMPLES section
	if !strings.Contains(output, "EXAMPLES") {
		t.Errorf("Expected help output to include 'EXAMPLES' section, got: %s", output)
	}

	// Check for send example
	if !strings.Contains(output, `agentmail send agent2 "Hello"`) {
		t.Errorf("Expected help output to include send example, got: %s", output)
	}

	// Check for stdin example
	if !strings.Contains(output, `echo "Hello" | agentmail send agent2`) {
		t.Errorf("Expected help output to include stdin example, got: %s", output)
	}

	// Check for receive example
	if !strings.Contains(output, "agentmail receive") {
		t.Errorf("Expected help output to include receive example, got: %s", output)
	}

	// Check for recipients example
	if !strings.Contains(output, "agentmail recipients") {
		t.Errorf("Expected help output to include recipients example, got: %s", output)
	}
}

// Additional test: Verify help includes usage section
func TestHelpCommand_IncludesUsageSection(t *testing.T) {
	var stdout bytes.Buffer

	Help(&stdout)

	output := stdout.String()

	// Check for USAGE section
	if !strings.Contains(output, "USAGE") {
		t.Errorf("Expected help output to include 'USAGE' section, got: %s", output)
	}

	// Check for usage syntax
	if !strings.Contains(output, "agentmail <command> [arguments]") {
		t.Errorf("Expected help output to include usage syntax, got: %s", output)
	}

	// Check for --help in usage
	if !strings.Contains(output, "agentmail --help") {
		t.Errorf("Expected help output to include '--help' in usage, got: %s", output)
	}
}

// Additional test: Verify help includes commands section
func TestHelpCommand_IncludesCommandsSection(t *testing.T) {
	var stdout bytes.Buffer

	Help(&stdout)

	output := stdout.String()

	// Check for COMMANDS section
	if !strings.Contains(output, "COMMANDS") {
		t.Errorf("Expected help output to include 'COMMANDS' section, got: %s", output)
	}
}

// Additional test: Verify help includes header/description
func TestHelpCommand_IncludesHeader(t *testing.T) {
	var stdout bytes.Buffer

	Help(&stdout)

	output := stdout.String()

	// Check for header
	if !strings.Contains(output, "agentmail - Inter-agent communication for tmux sessions") {
		t.Errorf("Expected help output to include header description, got: %s", output)
	}
}

// Additional test: Verify stdin message capability is documented
func TestHelpCommand_DocumentsStdinCapability(t *testing.T) {
	var stdout bytes.Buffer

	Help(&stdout)

	output := stdout.String()

	// Check that stdin capability is documented
	if !strings.Contains(output, "stdin") {
		t.Errorf("Expected help output to document stdin capability, got: %s", output)
	}
}

// T039: Verify help returns exit code 0 (explicitly)
func TestHelpCommand_ExitCode0(t *testing.T) {
	var stdout bytes.Buffer

	exitCode := Help(&stdout)

	if exitCode != 0 {
		t.Errorf("Help should always return exit code 0, got %d", exitCode)
	}
}

// Test exact output format matches contract
func TestHelpCommand_ExactOutputFormat(t *testing.T) {
	var stdout bytes.Buffer

	Help(&stdout)

	output := stdout.String()

	// Verify the exact expected output from the contract
	expectedContent := `agentmail - Inter-agent communication for tmux sessions

USAGE:
    agentmail <command> [arguments]
    agentmail --help

COMMANDS:
    send <recipient> [message]    Send a message to a tmux window
                                  Message can be piped via stdin
    receive                       Read the oldest unread message
    recipients                    List available message recipients

EXAMPLES:
    agentmail send agent2 "Hello"
    echo "Hello" | agentmail send agent2
    agentmail receive
    agentmail recipients
`

	if output != expectedContent {
		t.Errorf("Help output does not match expected format.\nExpected:\n%s\nGot:\n%s", expectedContent, output)
	}
}

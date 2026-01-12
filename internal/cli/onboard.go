package cli

import (
	"fmt"
	"io"
)

// OnboardOptions configures the Onboard command behavior.
type OnboardOptions struct {
	// Reserved for future use
}

// Onboard implements the agentmail onboard command.
// It outputs AI-optimized context for agents about AgentMail usage.
//
// Contract:
// agentmail onboard
//
// Exit Codes:
// - 0: Success
//
// Stdout: Onboarding context for the agent
//
// Behavior:
// Output AgentMail context including:
//   - What AgentMail is
//   - Command quick reference with examples
func Onboard(stdout, stderr io.Writer, opts OnboardOptions) int {
	// Output onboarding context
	fmt.Fprintln(stdout, "## AgentMail")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "AgentMail enables inter-agent communication within tmux sessions.")
	fmt.Fprintln(stdout)

	// Command reference with descriptions and examples
	fmt.Fprintln(stdout, "### Commands")
	fmt.Fprintln(stdout)

	// Send command
	fmt.Fprintln(stdout, "**send** - Send a message to another agent")
	fmt.Fprintln(stdout, "```")
	fmt.Fprintln(stdout, "agentmail send <recipient> \"<message>\"")
	fmt.Fprintln(stdout, "```")
	fmt.Fprintln(stdout, "Example:")
	fmt.Fprintln(stdout, "```")
	fmt.Fprintln(stdout, "agentmail send agent2 \"Hello, are you available?\"")
	fmt.Fprintln(stdout, "```")
	fmt.Fprintln(stdout)

	// Receive command
	fmt.Fprintln(stdout, "**receive** - Read the oldest unread message from your mailbox")
	fmt.Fprintln(stdout, "```")
	fmt.Fprintln(stdout, "agentmail receive")
	fmt.Fprintln(stdout, "```")
	fmt.Fprintln(stdout, "Returns \"No unread messages\" if mailbox is empty.")
	fmt.Fprintln(stdout)

	// Recipients command
	fmt.Fprintln(stdout, "**recipients** - List all agents you can message")
	fmt.Fprintln(stdout, "```")
	fmt.Fprintln(stdout, "agentmail recipients")
	fmt.Fprintln(stdout, "```")
	fmt.Fprintln(stdout, "Shows all tmux windows. Your window is marked with [you].")

	return 0
}

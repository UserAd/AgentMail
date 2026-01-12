package cli

import (
	"fmt"
	"io"
	"strings"

	"agentmail/internal/mail"
	"agentmail/internal/tmux"
)

// OnboardOptions configures the Onboard command behavior.
type OnboardOptions struct {
	SkipTmuxCheck  bool            // Skip tmux environment check
	MockWindows    []string        // Mock list of tmux windows
	MockCurrent    string          // Mock current window name
	MockIgnoreList map[string]bool // Mock ignore list (nil = load from file)
	MockGitRoot    string          // Mock git root (for testing)
}

// Onboard implements the agentmail onboard command.
// It outputs AI-optimized context for agents about AgentMail usage.
//
// Contract:
// agentmail onboard
//
// Exit Codes:
// - 0: Success (or silent no-op outside tmux)
//
// Stdout: Onboarding context for the agent
//
// Behavior:
// 1. Check if running inside tmux ($TMUX env var)
// 2. If not in tmux: exit 0 silently (no-op for hook integration)
// 3. Output AgentMail context including:
//   - What AgentMail is
//   - Current tmux window name (agent identity)
//   - Available recipients
//   - Command quick reference
func Onboard(stdout, stderr io.Writer, opts OnboardOptions) int {
	// Silent exit if not in tmux (hook-friendly behavior)
	if !opts.SkipTmuxCheck {
		if !tmux.InTmux() {
			return 0
		}
	}

	// Get current window name (agent identity)
	var currentWindow string
	if opts.MockWindows != nil {
		currentWindow = opts.MockCurrent
	} else {
		var err error
		currentWindow, err = tmux.GetCurrentWindow()
		if err != nil {
			// Silent failure for hook integration
			return 0
		}
	}

	// Get list of windows
	var windows []string
	if opts.MockWindows != nil {
		windows = opts.MockWindows
	} else {
		var err error
		windows, err = tmux.ListWindows()
		if err != nil {
			// Silent failure for hook integration
			return 0
		}
	}

	// Load ignore list
	var ignoreList map[string]bool
	if opts.MockIgnoreList != nil {
		ignoreList = opts.MockIgnoreList
	} else {
		var gitRoot string
		if opts.MockGitRoot != "" {
			gitRoot = opts.MockGitRoot
		} else {
			gitRoot, _ = mail.FindGitRoot()
		}
		if gitRoot != "" {
			ignoreList, _ = mail.LoadIgnoreList(gitRoot)
		}
	}

	// Build list of other agents (excluding current window and ignored)
	var otherAgents []string
	for _, window := range windows {
		if window != currentWindow && (ignoreList == nil || !ignoreList[window]) {
			otherAgents = append(otherAgents, window)
		}
	}

	// Output onboarding context
	fmt.Fprintln(stdout, "## AgentMail")
	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "You are **%s**. ", currentWindow)
	fmt.Fprintln(stdout, "AgentMail enables inter-agent communication within this tmux session.")
	fmt.Fprintln(stdout)

	// Other agents section
	if len(otherAgents) > 0 {
		fmt.Fprintf(stdout, "Other agents: %s\n", strings.Join(otherAgents, ", "))
	} else {
		fmt.Fprintln(stdout, "No other agents currently available.")
	}
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
	if len(otherAgents) > 0 {
		fmt.Fprintf(stdout, "agentmail send %s \"Hello, are you available?\"\n", otherAgents[0])
	} else {
		fmt.Fprintln(stdout, "agentmail send agent2 \"Hello, are you available?\"")
	}
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

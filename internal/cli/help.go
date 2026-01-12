package cli

import (
	"fmt"
	"io"
)

const helpText = `agentmail - Inter-agent communication for tmux sessions

USAGE:
    agentmail <command> [arguments]
    agentmail --help

COMMANDS:
    send <recipient> [message]    Send a message to a tmux window
                                  Message can be piped via stdin
    receive                       Read the oldest unread message
    recipients                    List available message recipients
    mailman [--daemon]            Start the notification daemon
    status <ready|work|offline>   Set agent availability status

EXAMPLES:
    agentmail send agent2 "Hello"
    echo "Hello" | agentmail send agent2
    agentmail receive
    agentmail recipients
    agentmail mailman              # Start daemon in foreground
    agentmail mailman --daemon     # Start daemon in background
    agentmail status ready         # Mark agent as ready for notifications
`

// Help writes the help text to stdout and returns exit code 0.
func Help(stdout io.Writer) int {
	fmt.Fprint(stdout, helpText)
	return 0
}

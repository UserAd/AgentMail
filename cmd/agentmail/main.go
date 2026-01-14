package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"agentmail/internal/cli"
	"agentmail/internal/mcp"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func main() {
	// Root command
	rootFlagSet := flag.NewFlagSet("agentmail", flag.ContinueOnError)

	// Send command flags
	sendFlagSet := flag.NewFlagSet("agentmail send", flag.ContinueOnError)
	var (
		sendRecipient string
		sendMessage   string
	)
	// Long and short forms for recipient
	sendFlagSet.StringVar(&sendRecipient, "recipient", "", "recipient tmux window name")
	sendFlagSet.StringVar(&sendRecipient, "r", "", "recipient tmux window name (shorthand)")
	// Long and short forms for message
	sendFlagSet.StringVar(&sendMessage, "message", "", "message content")
	sendFlagSet.StringVar(&sendMessage, "m", "", "message content (shorthand)")

	sendCmd := &ffcli.Command{
		Name:       "send",
		ShortUsage: "agentmail send [flags] [<recipient>] [<message>]",
		ShortHelp:  "Send a message to a tmux window",
		LongHelp: `Send a message to another agent in a tmux window.

The recipient and message can be specified either as positional arguments
or using flags. Flags take precedence over positional arguments.

Message can also be piped via stdin.

Examples:
  agentmail send agent2 "Hello"
  agentmail send -r agent2 -m "Hello"
  agentmail send --recipient agent2 --message "Hello"
  echo "Hello" | agentmail send agent2
  echo "Hello" | agentmail send -r agent2`,
		FlagSet: sendFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			// Build final args: prefer flags, fall back to positional
			var finalArgs []string

			// Recipient: flag or first positional arg
			recipient := sendRecipient
			if recipient == "" && len(args) > 0 {
				recipient = args[0]
				args = args[1:]
			}
			if recipient != "" {
				finalArgs = append(finalArgs, recipient)
			}

			// Message: flag or second positional arg (or stdin)
			message := sendMessage
			if message == "" && len(args) > 0 {
				message = args[0]
			}
			if message != "" {
				finalArgs = append(finalArgs, message)
			}

			exitCode := cli.Send(finalArgs, os.Stdin, os.Stdout, os.Stderr, cli.SendOptions{})
			if exitCode != 0 {
				os.Exit(exitCode)
			}
			return nil
		},
	}

	// Receive command flags
	receiveFlagSet := flag.NewFlagSet("agentmail receive", flag.ContinueOnError)
	var hookMode bool
	receiveFlagSet.BoolVar(&hookMode, "hook", false, "enable hook mode for Claude Code integration")

	receiveCmd := &ffcli.Command{
		Name:       "receive",
		ShortUsage: "agentmail receive [--hook]",
		ShortHelp:  "Read the oldest unread message",
		LongHelp: `Read the oldest unread message from your mailbox.

Flags:
  --hook    Enable hook mode for Claude Code integration.
            In hook mode:
            - Output goes to STDERR (not STDOUT)
            - Exit code 2 indicates new message available
            - Exit code 0 for no messages, not in tmux, or errors
            - Silent operation (no output on exit code 0)

Examples:
  agentmail receive
  agentmail receive --hook`,
		FlagSet: receiveFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			exitCode := cli.Receive(os.Stdout, os.Stderr, cli.ReceiveOptions{
				HookMode: hookMode,
			})
			if exitCode != 0 {
				os.Exit(exitCode)
			}
			return nil
		},
	}

	// Recipients command (no flags)
	recipientsFlagSet := flag.NewFlagSet("agentmail recipients", flag.ContinueOnError)

	recipientsCmd := &ffcli.Command{
		Name:       "recipients",
		ShortUsage: "agentmail recipients",
		ShortHelp:  "List available message recipients",
		LongHelp: `List all tmux windows in the current session that can receive messages.

The current window is marked with [you].
Windows in .agentmailignore are excluded from the list.

Examples:
  agentmail recipients`,
		FlagSet: recipientsFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			exitCode := cli.Recipients(os.Stdout, os.Stderr, cli.RecipientsOptions{})
			if exitCode != 0 {
				os.Exit(exitCode)
			}
			return nil
		},
	}

	// Status command (no flags)
	statusFlagSet := flag.NewFlagSet("agentmail status", flag.ContinueOnError)

	statusCmd := &ffcli.Command{
		Name:       "status",
		ShortUsage: "agentmail status <ready|work|offline>",
		ShortHelp:  "Set agent availability status",
		LongHelp: `Set the agent's availability status for hooks integration.

Valid statuses:
  ready    Agent is ready to receive messages
  work     Agent is busy working (resets notification flag)
  offline  Agent is offline (resets notification flag)

The status is stored in .agentmail/recipients.jsonl and used by the
mailman daemon for notification decisions.

When transitioning to 'work' or 'offline', the notified flag is reset
to false, allowing future notifications when returning to 'ready'.

Outside of a tmux session, this command is a silent no-op (exit 0).

Examples:
  agentmail status ready
  agentmail status work
  agentmail status offline`,
		FlagSet: statusFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			exitCode := cli.Status(args, os.Stdout, os.Stderr, cli.StatusOptions{})
			if exitCode != 0 {
				os.Exit(exitCode)
			}
			return nil
		},
	}

	// Mailman command flags
	mailmanFlagSet := flag.NewFlagSet("agentmail mailman", flag.ContinueOnError)
	var daemonMode bool
	mailmanFlagSet.BoolVar(&daemonMode, "daemon", false, "run in background (daemonize)")

	mailmanCmd := &ffcli.Command{
		Name:       "mailman",
		ShortUsage: "agentmail mailman [--daemon]",
		ShortHelp:  "Start the mailman daemon",
		LongHelp: `Start the mailman daemon for message delivery notifications.

The mailman daemon monitors mailboxes and can notify agents when new
messages arrive.

Flags:
  --daemon    Run in background (daemonize)

Exit codes:
  0  Success
  2  Daemon already running

Examples:
  agentmail mailman           # Run in foreground
  agentmail mailman --daemon  # Run in background`,
		FlagSet: mailmanFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			exitCode := cli.Mailman(os.Stdout, os.Stderr, cli.MailmanOptions{
				Daemonize: daemonMode,
			})
			if exitCode != 0 {
				os.Exit(exitCode)
			}
			return nil
		},
	}

	// Onboard command (no flags)
	onboardFlagSet := flag.NewFlagSet("agentmail onboard", flag.ContinueOnError)

	onboardCmd := &ffcli.Command{
		Name:       "onboard",
		ShortUsage: "agentmail onboard",
		ShortHelp:  "Output agent onboarding context",
		LongHelp: `Output AI-optimized context for agent onboarding.

This command outputs information to help AI agents understand
AgentMail and their environment:
- Agent identity (current tmux window name)
- Available recipients (other agents)
- Command quick reference

Designed for use with Claude Code hooks (SessionStart) to
automatically onboard agents at the start of each session.

Outside of a tmux session, this command is a silent no-op (exit 0).

Examples:
  agentmail onboard`,
		FlagSet: onboardFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			exitCode := cli.Onboard(os.Stdout, os.Stderr, cli.OnboardOptions{})
			if exitCode != 0 {
				os.Exit(exitCode)
			}
			return nil
		},
	}

	// MCP command (no flags)
	mcpFlagSet := flag.NewFlagSet("agentmail mcp", flag.ContinueOnError)

	mcpCmd := &ffcli.Command{
		Name:       "mcp",
		ShortUsage: "agentmail mcp",
		ShortHelp:  "Start MCP server (STDIO transport)",
		LongHelp: `Start the Model Context Protocol (MCP) server for AI agent integration.

The MCP server exposes AgentMail functionality through four tools:
  send            Send a message to another agent
  receive         Receive the oldest unread message
  status          Set agent availability status
  list-recipients List available agents in the session

The server uses STDIO transport and communicates via JSON-RPC 2.0.
It must be run inside a tmux session.

Exit codes:
  0  Normal shutdown
  1  Not in tmux session or tmux context lost

Examples:
  agentmail mcp`,
		FlagSet: mcpFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			// Create and run MCP server
			server, err := mcp.NewServer(nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}

			// Run the server (blocks until shutdown)
			if err := server.Run(ctx, nil); err != nil {
				// Context cancellation is normal shutdown
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return nil
				}
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			return nil
		},
	}

	// Root command help text
	rootHelp := `agentmail - Inter-agent communication for tmux sessions

Agents running in different tmux windows can send and receive messages
through a simple file-based mail system stored in .agentmail/.

Commands:
  send        Send a message to a tmux window
  receive     Read the oldest unread message
  recipients  List available message recipients
  status      Set agent availability status
  mailman     Start the mailman daemon
  onboard     Output agent onboarding context
  mcp         Start MCP server (STDIO transport)

Use "agentmail <command> --help" for more information about a command.`

	// Root command
	root := &ffcli.Command{
		ShortUsage:  "agentmail <command> [flags] [arguments]",
		ShortHelp:   "Inter-agent communication for tmux sessions",
		LongHelp:    rootHelp,
		FlagSet:     rootFlagSet,
		Subcommands: []*ffcli.Command{sendCmd, receiveCmd, recipientsCmd, statusCmd, mailmanCmd, onboardCmd, mcpCmd},
		Exec: func(ctx context.Context, args []string) error {
			// No subcommand provided, show help
			fmt.Fprintln(os.Stderr, rootHelp)
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "Run 'agentmail <command> --help' for usage.")
			os.Exit(1)
			return nil
		},
	}

	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

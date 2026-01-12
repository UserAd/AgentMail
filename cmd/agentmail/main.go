package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"agentmail/internal/cli"

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

	// Root command help text
	rootHelp := `agentmail - Inter-agent communication for tmux sessions

Agents running in different tmux windows can send and receive messages
through a simple file-based mail system stored in .git/mail/.

Commands:
  send        Send a message to a tmux window
  receive     Read the oldest unread message
  recipients  List available message recipients
  mailman     Start the mailman daemon

Use "agentmail <command> --help" for more information about a command.`

	// Root command
	root := &ffcli.Command{
		ShortUsage:  "agentmail <command> [flags] [arguments]",
		ShortHelp:   "Inter-agent communication for tmux sessions",
		LongHelp:    rootHelp,
		FlagSet:     rootFlagSet,
		Subcommands: []*ffcli.Command{sendCmd, receiveCmd, recipientsCmd, mailmanCmd},
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

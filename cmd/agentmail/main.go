package main

import (
	"fmt"
	"os"

	"agentmail/internal/cli"
)

func main() {
	// Check for --help or -h flag BEFORE command parsing
	if len(os.Args) >= 2 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		exitCode := cli.Help(os.Stdout)
		os.Exit(exitCode)
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "error: missing command")
		fmt.Fprintln(os.Stderr)
		cli.Help(os.Stderr)
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "send":
		// T024: Wire up send subcommand
		// T049: Pass os.Stdin to Send()
		exitCode := cli.Send(os.Args[2:], os.Stdin, os.Stdout, os.Stderr, cli.SendOptions{})
		os.Exit(exitCode)
	case "receive":
		// T038: Wire up receive subcommand
		exitCode := cli.Receive(os.Stdout, os.Stderr, cli.ReceiveOptions{})
		os.Exit(exitCode)
	case "recipients":
		// T016: Wire up recipients subcommand
		exitCode := cli.Recipients(os.Stdout, os.Stderr, cli.RecipientsOptions{})
		os.Exit(exitCode)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command '%s'\n", command)
		fmt.Fprintln(os.Stderr)
		cli.Help(os.Stderr)
		os.Exit(1)
	}
}

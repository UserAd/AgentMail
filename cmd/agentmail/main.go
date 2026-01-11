package main

import (
	"fmt"
	"os"

	"agentmail/internal/cli"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "error: missing command")
		fmt.Fprintln(os.Stderr, "usage: agentmail <command> [arguments]")
		fmt.Fprintln(os.Stderr, "commands: send, receive")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "send":
		// T024: Wire up send subcommand
		exitCode := cli.Send(os.Args[2:], os.Stdout, os.Stderr, cli.SendOptions{})
		os.Exit(exitCode)
	case "receive":
		fmt.Fprintln(os.Stderr, "error: receive command not implemented")
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command '%s'\n", command)
		fmt.Fprintln(os.Stderr, "usage: agentmail <command> [arguments]")
		fmt.Fprintln(os.Stderr, "commands: send, receive")
		os.Exit(1)
	}
}

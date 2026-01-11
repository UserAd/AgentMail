package main

import (
	"fmt"
	"os"
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
		fmt.Fprintln(os.Stderr, "error: send command not implemented")
		os.Exit(1)
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

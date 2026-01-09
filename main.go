package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	version = "0.1.0"
)

func main() {
	// Define subcommands
	versionCmd := flag.NewFlagSet("version", flag.ExitOnError)
	helloCmd := flag.NewFlagSet("hello", flag.ExitOnError)

	// Define flags for hello subcommand
	helloName := helloCmd.String("name", "World", "Name to greet")

	// Show usage if no arguments
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Parse subcommands
	switch os.Args[1] {
	case "version":
		versionCmd.Parse(os.Args[2:])
		handleVersion()
	case "hello":
		helloCmd.Parse(os.Args[2:])
		handleHello(*helloName)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("AgentMail CLI")
	fmt.Println("\nUsage:")
	fmt.Println("  agentmail <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  version              Show version information")
	fmt.Println("  hello [--name NAME]  Say hello to someone")
	fmt.Println("  help                 Show this help message")
	fmt.Println("\nOptions:")
	fmt.Println("  -h, --help           Show help")
}

func handleVersion() {
	fmt.Printf("AgentMail version %s\n", version)
}

func handleHello(name string) {
	fmt.Printf("Hello, %s!\n", name)
}

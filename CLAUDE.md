# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AgentMail is a Golang CLI application built using the [Cobra](https://github.com/spf13/cobra) framework. The project follows the standard Cobra architectural pattern with a modular command structure.

## Development Commands

### Building
```bash
go build -o agentmail
```

### Testing
```bash
# Run all tests
go test ./cmd/... -v

# Run tests with coverage
go test ./cmd/... -cover

# Run a specific test
go test ./cmd -v -run TestHelloCommand

# Run a specific sub-test
go test ./cmd -v -run TestHelloCommand/custom_name_with_--name
```

### Running the CLI
```bash
./agentmail --help
./agentmail version
./agentmail hello --name Alice
```

## Architecture

### Command Structure

The application uses Cobra's command registration pattern:

1. **`main.go`** - Entry point that calls `cmd.Execute()`
2. **`cmd/root.go`** - Defines the root command and `Execute()` function
3. **`cmd/<command>.go`** - Individual command files that auto-register via `init()`

### Adding a New Command

To add a new command:

1. Create `cmd/<commandname>.go`
2. Define the command using `&cobra.Command{}`
3. Register it in `init()` with `rootCmd.AddCommand(<commandname>Cmd)`
4. Create corresponding `cmd/<commandname>_test.go` for tests

Example pattern:
```go
var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Brief description",
    Long:  "Detailed description",
    Run: func(cmd *cobra.Command, args []string) {
        // Command logic
    },
}

func init() {
    rootCmd.AddCommand(myCmd)
    // Add flags with: myCmd.Flags().StringVarP(...)
}
```

### Command Registration

All commands in `cmd/` package are automatically registered through `init()` functions. The `init()` functions execute before `main()`, registering commands with `rootCmd` via `rootCmd.AddCommand()`.

### Testing Pattern

Tests create isolated Cobra command instances to avoid state pollution between tests. Each test constructs its own command tree and uses buffers to capture output for assertions.

## Code Conventions

- **Version**: Defined as a constant in `cmd/version.go`
- **Flags**: Use `StringVarP()` for flags with both long (`--name`) and short (`-n`) forms
- **Package-level vars**: Used for flag values (e.g., `var helloName string`)
- **Module path**: `github.com/UserAd/AgentMail`

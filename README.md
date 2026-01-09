# AgentMail

A Golang CLI application built with [Cobra](https://github.com/spf13/cobra).

## Installation

```bash
go build -o agentmail
```

## Usage

```bash
# Show help
./agentmail help

# Show version
./agentmail version

# Say hello
./agentmail hello
./agentmail hello --name Alice
./agentmail hello -n Bob  # short flag
```

## Commands

- `version` - Display version information
- `hello` - Say hello with optional name flag (--name or -n)
- `help` - Show usage information
- `completion` - Generate shell autocompletion scripts

## Development

### Running Tests

```bash
# Run all tests
go test ./cmd/... -v

# Run with coverage
go test ./cmd/... -cover
```

## Technology Stack

- [Cobra](https://github.com/spf13/cobra) - Modern CLI framework for Go
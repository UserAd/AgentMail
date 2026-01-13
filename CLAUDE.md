# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AgentMail is a Go CLI tool for inter-agent communication within tmux sessions. Agents running in different tmux windows can send and receive messages through a simple mail system stored in `.agentmail/`.

## AgentMail Usage

**Requirements:** Must be running inside a tmux session.

Use `agentmail --help` to get information about agentmail usage.

### Sending Messages
```bash
agentmail send <recipient> "<message>"
```
- `<recipient>` must be a valid tmux window name in the current session
- Returns a message ID on success (e.g., `Message #ABC123 sent`)

### Receiving Messages
```bash
agentmail receive
```
- Reads the oldest unread message addressed to the current tmux window
- Marks the message as read after displaying
- Returns "No unread messages" if mailbox is empty
- Messages are delivered in FIFO order

### Message Storage
- Messages stored in `.agentmail/mailboxes/<recipient>.jsonl`
- Each recipient has their own mailbox file
- File locking ensures atomic operations

## Build Commands

Once Go source files are added:
- `go build ./...` - Build all packages
- `go test ./...` - Run all tests
- `go test -v ./path/to/package` - Run tests for a specific package with verbose output
- `go test -run TestName ./...` - Run a specific test by name
- `go vet ./...` - Run static analysis
- `go fmt ./...` - Format code

## Quality gates

- The build system shall produce agentmail binary
- The build system shall pass `go test -v -race ./...` without errors
- The build system shall pass `go vet ./...` without errors
- The build system shall pass `go fmt ./...` without errors
- The build system shall pass `govulncheck ./...` without errors
- The build system shall pass `gosec ./...` without errors

## Testing in CI Environment

To run tests in a container matching CI (Go 1.21, Linux):
```bash
docker run --rm -v $(pwd):/app -w /app golang:1.21 go test -v -race ./...
```

This helps catch issues that only manifest in the CI environment (e.g., running as root, different Go version).

## Specification Workflow

This project uses speckit for feature specification and planning. Available commands in `.claude/commands/`:
- `/speckit.specify` - Create or update feature specifications
- `/speckit.plan` - Generate implementation plans
- `/speckit.tasks` - Generate actionable task lists
- `/speckit.implement` - Execute implementation plans
- `/speckit.clarify` - Identify underspecified areas in specs
- `/speckit.analyze` - Cross-artifact consistency analysis
- `/speckit.checklist` - Generate custom checklists
- `/speckit.constitution` - Create/update project constitution
- `/speckit.taskstoissues` - Convert tasks to GitHub issues

Templates are stored in `.specify/templates/` and project constitution in `.specify/memory/constitution.md`.

## Active Technologies
- Go 1.21+ (per IC-001) + Standard library only (os/exec for tmux, encoding/json for JSONL) (001-agent-mail-structure)
- JSONL file in `.agentmail/` directory (001-agent-mail-structure)
- Go 1.21+ (project uses Go 1.25.3) + GitHub Actions (yaml workflows), PaulHatch/semantic-version action for version calculation (002-github-ci-cd)
- N/A (CI/CD configuration files only) (002-github-ci-cd)
- Go 1.21+ (per constitution IC-001, project uses Go 1.25.3) + Standard library only (os/exec, encoding/json, bufio, os) (003-recipients-help-stdin)
- JSONL files in `.agentmail/` directory (003-recipients-help-stdin)
- Ruby (Homebrew formula DSL), YAML (GitHub Actions), Go 1.21+ (existing) + Homebrew (user-side), GitHub Actions, gh CLI (for cross-repo updates) (004-homebrew-distribution)
- N/A (formula hosted in separate GitHub repo) (004-homebrew-distribution)
- Go 1.21+ (per constitution IC-001, project uses Go 1.25.3) + Standard library only (os, fmt, io - already used) (005-claude-hooks-integration)
- JSONL files in `.agentmail/` directory (005-claude-hooks-integration)
- Go 1.21+ (per constitution IC-001, project uses Go 1.25.3) + Standard library only (os/exec, encoding/json, syscall, time, os/signal) (006-mailman-daemon)
- JSONL files - `.agentmail/mailman.pid` (PID), `.agentmail/recipients.jsonl` (state) (006-mailman-daemon)
- Go 1.21+ (per IC-001) + Standard library only (os, filepath, syscall, encoding/json) (007-storage-restructure)
- JSONL files in `.agentmail/` directory hierarchy (007-storage-restructure)

## Recent Changes
- 001-agent-mail-structure: Added Go 1.21+ (per IC-001) + Standard library only (os/exec for tmux, encoding/json for JSONL)

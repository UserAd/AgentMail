# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AgentMail is a Go project (currently in early development).

## Build Commands

Once Go source files are added:
- `go build ./...` - Build all packages
- `go test ./...` - Run all tests
- `go test -v ./path/to/package` - Run tests for a specific package with verbose output
- `go test -run TestName ./...` - Run a specific test by name
- `go vet ./...` - Run static analysis
- `go fmt ./...` - Format code

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

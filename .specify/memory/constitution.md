<!--
Sync Impact Report
==================
Version change: 1.0.0 → 1.1.0 (MINOR - scope expansion beyond MVP)

Modified principles:
- I. CLI-First Design: Removed "No daemon processes in MVP scope" restriction
- II. Simplicity (YAGNI): Updated to reflect post-MVP stage, removed "MVP scope: send and receive commands only"

Added sections: None

Removed sections: None

Templates requiring updates:
- .specify/templates/plan-template.md: ✅ No changes needed (generic constitution check)
- .specify/templates/spec-template.md: ✅ No changes needed (compatible)
- .specify/templates/tasks-template.md: ✅ No changes needed (compatible)

Follow-up TODOs: None
-->

# AgentMail Constitution

## Core Principles

### I. CLI-First Design

AgentMail is a command-line tool. All functionality MUST be accessible via CLI commands with:

- Text-based input/output protocol: arguments → stdout, errors → stderr
- Deterministic exit codes: 0 (success), 1 (error), 2 (environment error)
- Human-readable output by default
- No GUI or web interface
- Daemon processes are permitted when they enhance CLI workflows (e.g., background notifications)

**Rationale**: CLI tools are composable, scriptable, and testable. Agent-to-agent communication requires predictable, automatable interfaces. Daemon processes extend CLI capabilities without replacing them.

### II. Simplicity (YAGNI)

Build only what is needed. Features MUST be justified by demonstrated need:

- Standard library dependencies preferred over external packages
- No premature abstractions or "future-proofing"
- Complexity MUST be explicitly justified in plan.md
- New commands/features require clear use cases

**Rationale**: AgentMail serves a focused purpose. Over-engineering creates maintenance burden and obscures core functionality.

### III. Test Coverage (NON-NEGOTIABLE)

All code MUST achieve minimum 80% test coverage as measured by `go test -cover`:

- Tests written before or alongside implementation (TDD encouraged)
- Unit tests for all public functions
- Integration tests for CLI command flows
- Coverage gate enforced before merge

**Rationale**: Inter-agent communication is infrastructure. Regressions break dependent agents silently.

### IV. Standard Library Preference

External dependencies MUST be justified. Prefer Go standard library:

- `os/exec` for tmux integration
- `encoding/json` for JSONL handling
- `crypto/rand` for ID generation
- `syscall` for file locking
- `os/signal` for daemon signal handling
- `time` for scheduling and timeouts

New dependencies require documented rationale in research.md with:
- Why standard library is insufficient
- Security/maintenance implications
- Alternative approaches considered

**Rationale**: Minimal dependencies reduce supply chain risk and simplify builds for a tool that may run in diverse agent environments.

## Technology Constraints

- **Language**: Go 1.21+ (per IC-001)
- **Storage**: JSONL files in `.git/mail/` directory (per-recipient files, state files)
- **Platform**: macOS and Linux with tmux installed
- **Build**: Standard `go build`, no CGO dependencies

## Quality Gates

Before any feature is considered complete:

1. **Coverage**: `go test -cover ./...` reports >= 80%
2. **Static Analysis**: `go vet ./...` passes with no errors
3. **Formatting**: `go fmt ./...` produces no changes
4. **Spec Compliance**: All acceptance scenarios from spec.md pass

## Governance

This constitution supersedes all other development practices for AgentMail.

**Amendment Process**:
1. Propose change with rationale in PR description
2. Update constitution version (MAJOR for principle changes, MINOR for additions, PATCH for clarifications)
3. Update dependent templates if affected
4. Document in Sync Impact Report

**Compliance**:
- All PRs MUST verify constitution compliance
- Violations require explicit justification or constitution amendment
- `/speckit.analyze` checks constitution alignment automatically

**Version**: 1.1.0 | **Ratified**: 2026-01-11 | **Last Amended**: 2026-01-12

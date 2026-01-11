<!--
Sync Impact Report
==================
Version change: 0.0.0 → 1.0.0 (MAJOR - initial ratification)

Modified principles: N/A (new constitution)

Added sections:
- Core Principles (4 principles)
- Technology Constraints
- Quality Gates
- Governance

Removed sections: N/A

Templates requiring updates:
- .specify/templates/plan-template.md: ✅ Already has Constitution Check section
- .specify/templates/spec-template.md: ✅ Compatible with principles
- .specify/templates/tasks-template.md: ✅ Compatible with test-first approach

Follow-up TODOs: None
-->

# AgentMail Constitution

## Core Principles

### I. CLI-First Design

AgentMail is a command-line tool. All functionality MUST be accessible via CLI commands with:

- Text-based input/output protocol: arguments → stdout, errors → stderr
- Deterministic exit codes: 0 (success), 1 (error), 2 (environment error)
- Human-readable output by default
- No GUI, web interface, or daemon processes in MVP scope

**Rationale**: CLI tools are composable, scriptable, and testable. Agent-to-agent communication requires predictable, automatable interfaces.

### II. Simplicity (YAGNI)

Start with the minimum viable implementation. Features MUST be justified by immediate need:

- MVP scope: `send` and `receive` commands only
- Standard library dependencies preferred over external packages
- No premature abstractions or "future-proofing"
- Complexity MUST be explicitly justified in plan.md

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

New dependencies require documented rationale in research.md with:
- Why standard library is insufficient
- Security/maintenance implications
- Alternative approaches considered

**Rationale**: Minimal dependencies reduce supply chain risk and simplify builds for a tool that may run in diverse agent environments.

## Technology Constraints

- **Language**: Go 1.21+ (per IC-001)
- **Storage**: JSONL files in `.git/mail/` directory (per-recipient files)
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

**Version**: 1.0.0 | **Ratified**: 2026-01-11 | **Last Amended**: 2026-01-11

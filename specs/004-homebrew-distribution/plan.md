# Implementation Plan: Homebrew Distribution

**Branch**: `004-homebrew-distribution` | **Date**: 2026-01-12 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/004-homebrew-distribution/spec.md`

## Summary

Enable macOS and Linux users to install AgentMail via Homebrew by creating a tap repository (`homebrew-agentmail`) with a formula that downloads pre-built binaries. The release workflow will be extended to automatically update the formula with new version numbers and SHA256 checksums on each release. (10 functional requirements, EARS-compliant)

## Technical Context

**Language/Version**: Ruby (Homebrew formula DSL), YAML (GitHub Actions), Go 1.21+ (existing)
**Primary Dependencies**: Homebrew (user-side), GitHub Actions, gh CLI (for cross-repo updates)
**Storage**: N/A (formula hosted in separate GitHub repo)
**Testing**: Manual `brew install` validation, `brew audit` for formula linting
**Target Platform**: macOS (Intel amd64 and Apple Silicon arm64), Linux (amd64 via Linuxbrew)
**Project Type**: Configuration/infrastructure (formula + CI workflow)
**Performance Goals**: Installation completes in under 30 seconds
**Constraints**: Must work without additional authentication for public tap
**Scale/Scope**: Single formula, single binary distribution

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Design Check (Phase 0)

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | ✅ PASS | Homebrew distribution supports CLI tool installation |
| II. Simplicity (YAGNI) | ✅ PASS | Minimal formula with pre-built binaries, no unnecessary complexity |
| III. Test Coverage | ✅ PASS | No Go code changes required; formula validated via `brew audit` |
| IV. Standard Library Preference | ✅ PASS | No new Go dependencies; CI uses standard GitHub Actions |

### Post-Design Check (Phase 1)

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | ✅ PASS | No changes to CLI interface |
| II. Simplicity (YAGNI) | ✅ PASS | Minimal infrastructure: 1 formula file, 1 workflow job, README update |
| III. Test Coverage | ✅ PASS | Formula tested via `brew audit`; no Go code changes |
| IV. Standard Library Preference | ✅ PASS | Uses only GitHub Actions built-ins (curl, shasum, sed) |

**Quality Gates Applicability**:
- Coverage: N/A (no Go code changes)
- Static Analysis: N/A (no Go code changes)
- Formatting: N/A (no Go code changes)
- Spec Compliance: Will verify all acceptance scenarios pass after implementation

## Project Structure

### Documentation (this feature)

```text
specs/004-homebrew-distribution/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output (minimal - no data model)
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
# Changes to AgentMail repo
.github/
└── workflows/
    └── release.yml      # Extended with formula update job

README.md                # Add Homebrew installation section

# New homebrew-agentmail repository (separate)
homebrew-agentmail/
├── Formula/
│   └── agentmail.rb     # Homebrew formula
└── README.md            # Tap documentation
```

**Structure Decision**: This feature creates infrastructure files (CI workflow modification, README update) in the main repo and requires a new separate repository for the Homebrew tap following Homebrew's `homebrew-<name>` convention.

## Complexity Tracking

> No constitution violations - table not needed.

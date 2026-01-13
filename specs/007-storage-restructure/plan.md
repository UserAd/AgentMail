# Implementation Plan: Storage Directory Restructure

**Branch**: `007-storage-restructure` | **Date**: 2026-01-13 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/007-storage-restructure/spec.md`

## Summary

Restructure AgentMail's storage from `.git/mail/` to `.agentmail/` with a dedicated `mailboxes/` subdirectory. This is a breaking change that moves all storage paths while maintaining existing JSONL formats and file locking mechanisms. No automatic migration or warnings for existing users.

## Technical Context

**Language/Version**: Go 1.21+ (per IC-001)
**Primary Dependencies**: Standard library only (os, filepath, syscall, encoding/json)
**Storage**: JSONL files in `.agentmail/` directory hierarchy
**Testing**: `go test -cover ./...` with minimum 80% coverage
**Target Platform**: macOS and Linux with tmux installed
**Project Type**: Single CLI application
**Performance Goals**: Directory/file creation within 100 milliseconds
**Constraints**: No external dependencies, git repository still required
**Scale/Scope**: Single-user CLI tool, local file operations only

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | PASS | No changes to CLI interface, only storage paths |
| II. Simplicity (YAGNI) | PASS | Simple constant changes, no new abstractions |
| III. Test Coverage (80%) | PASS | Update existing tests to use new paths |
| IV. Standard Library | PASS | No new dependencies required |

**Quality Gates Checklist**:
- [ ] Coverage: `go test -cover ./...` >= 80%
- [ ] Static Analysis: `go vet ./...` passes
- [ ] Formatting: `go fmt ./...` no changes
- [ ] Spec Compliance: All acceptance scenarios pass

**Constitution Conflict**: Storage constraint in constitution says `.git/mail/` but this feature changes to `.agentmail/`. This is intentional scope of the feature and will require constitution update after implementation.

## Project Structure

### Documentation (this feature)

```text
specs/007-storage-restructure/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (CLI contracts)
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
cmd/
└── agentmail/
    └── main.go              # Entry point (no changes expected)

internal/
├── cli/                     # CLI command implementations
│   ├── send.go              # Uses mail.EnsureMailDir
│   ├── receive.go           # Uses mail.GetMailboxPath
│   ├── mailman.go           # Uses daemon.PIDFilePath
│   ├── recipients.go        # Uses mail.RecipientsFile
│   └── status.go            # Uses mail paths
├── daemon/
│   └── daemon.go            # PIDFile constant, PIDFilePath function
├── mail/
│   ├── mailbox.go           # MailDir constant, EnsureMailDir, mailbox paths
│   └── recipients.go        # RecipientsFile constant
└── tmux/
    └── tmux.go              # No changes expected
```

**Structure Decision**: Existing Go project structure. Changes confined to path constants in `internal/mail/` and `internal/daemon/` packages.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Constitution storage path change | Feature requirement to decouple from .git | N/A - this is the feature's purpose |

## Files to Modify

### Primary Changes (Constants)

| File | Current Value | New Value |
|------|---------------|-----------|
| `internal/mail/mailbox.go` | `MailDir = ".git/mail"` | `MailDir = ".agentmail/mailboxes"` |
| `internal/mail/mailbox.go` | N/A | Add `RootDir = ".agentmail"` |
| `internal/mail/recipients.go` | `RecipientsFile = ".git/mail-recipients.jsonl"` | `RecipientsFile = ".agentmail/recipients.jsonl"` |
| `internal/daemon/daemon.go` | Uses `mail.MailDir + "/mailman.pid"` | Use `mail.RootDir + "/mailman.pid"` |

### Secondary Changes (Directory Creation)

| File | Change Required |
|------|-----------------|
| `internal/mail/mailbox.go` | `EnsureMailDir` must create both `.agentmail/` and `.agentmail/mailboxes/` |
| `internal/daemon/daemon.go` | `PIDFilePath` uses new root path |

### Test Updates

| Test File | Change Required |
|-----------|-----------------|
| `internal/mail/mailbox_test.go` | Update expected paths in assertions |
| `internal/mail/recipients_test.go` | Update expected paths in assertions |
| `internal/daemon/daemon_test.go` | Update expected paths in assertions |
| `internal/cli/*_test.go` | Update any hardcoded path expectations |

### Documentation Updates

| File | Change Required |
|------|-----------------|
| `README.md` | Update storage location documentation |
| `CLAUDE.md` | Update Message Storage section |
| `.specify/memory/constitution.md` | Update Technology Constraints storage path |

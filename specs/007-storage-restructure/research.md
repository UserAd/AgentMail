# Research: Storage Directory Restructure

**Feature**: 007-storage-restructure
**Date**: 2026-01-13

## Overview

This research documents the analysis and decisions for restructuring AgentMail's storage from `.git/mail/` to `.agentmail/`.

## Research Topics

### 1. Directory Structure Design

**Decision**: Use `.agentmail/` as root with `mailboxes/` subdirectory

**Rationale**:
- Separating mailboxes into a subdirectory allows other files (PID, recipients) to live at the root level
- Clear organizational hierarchy: root contains metadata, subdirectory contains user data
- Follows pattern of other tools (e.g., `.npm/`, `.cargo/`) that use subdirectories for different concerns

**Alternatives Considered**:
- Flat structure (`.agentmail/*.jsonl`): Rejected because it mixes mailboxes with metadata files
- Nested structure (`.agentmail/data/mailboxes/`): Rejected as over-engineering for current needs

### 2. Git Repository Requirement

**Decision**: Continue requiring git repository for operation

**Rationale**:
- Existing codebase uses git repository detection for finding project root
- Changing this would require new working directory detection logic
- tmux sessions typically operate within git repositories
- Simpler implementation path

**Alternatives Considered**:
- Allow any directory: Would require significant changes to root detection and scope expansion
- Create `.agentmail/` in current directory: Could lead to scattered mail directories

### 3. Migration Strategy

**Decision**: No automatic migration, no warnings

**Rationale**:
- AgentMail is a developer tool with technical users
- Manual migration is trivial: `mv .git/mail/* .agentmail/mailboxes/`
- No warnings reduces code complexity and output noise
- Clean break allows simpler codebase

**Alternatives Considered**:
- Automatic migration: Adds complexity, risk of data corruption
- One-time warning: Adds state tracking complexity (remembering if warned)
- Deprecation period with fallback: Over-engineering for tool's user base

### 4. Gitignore Handling

**Decision**: User manages `.gitignore` entries

**Rationale**:
- Some users may want to track mail data
- Automatic file modification can surprise users
- Previous `.git/mail/` was implicitly ignored; new location makes tracking explicit choice

**Alternatives Considered**:
- Auto-add to `.gitignore`: Could conflict with user intentions, requires file modification logic

### 5. File Permissions

**Decision**: Maintain existing permissions (0750 directories, 0640 files)

**Rationale**:
- Current implementation already uses these permissions
- No security model change required
- Appropriate for single-user CLI tool

**Alternatives Considered**: None - existing permissions are appropriate

## Implementation Findings

### Current Path Constants

```go
// internal/mail/mailbox.go
const MailDir = ".git/mail"

// internal/mail/recipients.go
const RecipientsFile = ".git/mail-recipients.jsonl"

// internal/daemon/daemon.go
const PIDFile = "mailman.pid"  // Combined with mail.MailDir
```

### New Path Constants

```go
// internal/mail/mailbox.go
const RootDir = ".agentmail"
const MailDir = ".agentmail/mailboxes"

// internal/mail/recipients.go
const RecipientsFile = ".agentmail/recipients.jsonl"

// internal/daemon/daemon.go
const PIDFile = "mailman.pid"  // Combined with mail.RootDir
```

### Functions Requiring Updates

| Function | File | Change |
|----------|------|--------|
| `EnsureMailDir` | `internal/mail/mailbox.go` | Create both RootDir and MailDir |
| `PIDFilePath` | `internal/daemon/daemon.go` | Use RootDir instead of MailDir |

### Files Referencing Old Paths

Grep results for `.git/mail` show references in:
- Source code: 4 files in `internal/`
- Tests: 8 test files
- Documentation: README.md, CLAUDE.md, constitution.md
- Specs: Previous feature specs (historical, no update needed)

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Existing users lose data | Low | Medium | Documentation clearly states manual migration needed |
| Tests fail after changes | High | Low | Update test assertions alongside code changes |
| Constitution conflict | Certain | Low | Update constitution as part of feature |

## Conclusion

This is a straightforward refactoring with no technical unknowns. The implementation requires:
1. Update 2 constant definitions
2. Update 1 function (`EnsureMailDir`)
3. Update 1 path calculation (`PIDFilePath`)
4. Update test assertions
5. Update documentation

No external research or dependency analysis needed.

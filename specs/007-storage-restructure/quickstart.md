# Quickstart: Storage Directory Restructure

**Feature**: 007-storage-restructure
**Date**: 2026-01-13

## What's Changing

AgentMail's storage location is moving from `.git/mail/` to `.agentmail/`:

| Component | Old Location | New Location |
|-----------|--------------|--------------|
| Mailboxes | `.git/mail/<recipient>.jsonl` | `.agentmail/mailboxes/<recipient>.jsonl` |
| Recipients | `.git/mail-recipients.jsonl` | `.agentmail/recipients.jsonl` |
| PID file | `.git/mail/mailman.pid` | `.agentmail/mailman.pid` |

## Implementation Checklist

### Step 1: Update Path Constants

**File**: `internal/mail/mailbox.go`

```go
// Add new constant
const RootDir = ".agentmail"

// Update existing constant
const MailDir = ".agentmail/mailboxes"  // was ".git/mail"
```

**File**: `internal/mail/recipients.go`

```go
// Update constant
const RecipientsFile = ".agentmail/recipients.jsonl"  // was ".git/mail-recipients.jsonl"
```

### Step 2: Update EnsureMailDir Function

**File**: `internal/mail/mailbox.go`

Update `EnsureMailDir` to create both directories:

```go
func EnsureMailDir(repoRoot string) error {
    // Create root directory first
    rootPath := filepath.Join(repoRoot, RootDir)
    if err := os.MkdirAll(rootPath, 0750); err != nil {
        return err
    }

    // Create mailboxes directory
    mailPath := filepath.Join(repoRoot, MailDir)
    return os.MkdirAll(mailPath, 0750)
}
```

### Step 3: Update PIDFilePath Function

**File**: `internal/daemon/daemon.go`

```go
func PIDFilePath(repoRoot string) string {
    return filepath.Join(repoRoot, mail.RootDir, PIDFile)  // was mail.MailDir
}
```

### Step 4: Update Tests

Update path assertions in:
- `internal/mail/mailbox_test.go`
- `internal/mail/recipients_test.go`
- `internal/daemon/daemon_test.go`
- `internal/cli/*_test.go`

### Step 5: Update Documentation

- `README.md`: Update storage location references
- `CLAUDE.md`: Update Message Storage section
- `.specify/memory/constitution.md`: Update Technology Constraints

## Verification

### Build and Test

```bash
# Build
go build ./...

# Run tests
go test -cover ./...

# Verify coverage >= 80%
go test -cover ./... | grep -E 'coverage:|ok'

# Static analysis
go vet ./...

# Format check
go fmt ./...
```

### Manual Testing

```bash
# Start in a tmux session within a git repository
tmux

# Test send (creates .agentmail/mailboxes/)
agentmail send test-window "Hello"

# Verify directory structure
ls -la .agentmail/
ls -la .agentmail/mailboxes/

# Test mailman
agentmail mailman start
ls -la .agentmail/mailman.pid
agentmail mailman status
agentmail mailman stop

# Verify no old paths
ls -la .git/mail 2>/dev/null || echo "Good: .git/mail does not exist"
```

## Migration (For Existing Users)

If you have existing data in `.git/mail/`:

```bash
# Stop daemon first
agentmail mailman stop

# Create new structure and move data
mkdir -p .agentmail/mailboxes
mv .git/mail/*.jsonl .agentmail/mailboxes/
mv .git/mail-recipients.jsonl .agentmail/recipients.jsonl

# Clean up (optional)
rm -rf .git/mail
```

## Common Issues

### Issue: "No such file or directory" for old paths

**Cause**: Code still references old path constants.
**Fix**: Ensure all path constants are updated and rebuild.

### Issue: Tests fail with path assertions

**Cause**: Test expectations use old paths.
**Fix**: Update expected paths in test assertions.

### Issue: Permission denied creating .agentmail

**Cause**: Repository root not writable.
**Fix**: Check permissions on repository root directory.

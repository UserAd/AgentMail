# Quickstart: Mailman Stop Command

**Feature**: 012-mailman-stop
**Date**: 2026-01-20

## Overview

Add `agentmail mailman stop` subcommand to stop the running mailman daemon using file-based signaling.

## Implementation Steps

### Step 1: Add Stop File Path Constant (internal/daemon/daemon.go)

Add constant for the stop file path:

```go
// StopFile is the filename for the mailman daemon stop signal within .agentmail/
const StopFile = ".stop"

// StopFilePath returns the full path to the stop signal file for a given repository root.
func StopFilePath(repoRoot string) string {
    return filepath.Join(repoRoot, mail.RootDir, StopFile)
}
```

### Step 2: Add Stop Command Handler (internal/cli/mailman_stop.go)

Create new file with the stop command implementation:

```go
// MailmanStopOptions configures the MailmanStop command behavior.
type MailmanStopOptions struct {
    RepoRoot string // Repository root (defaults to finding git root)
}

// MailmanStop implements the agentmail mailman stop command.
// Creates a .stop file to signal the daemon to shut down.
//
// Exit codes:
// - 0: Success (stop signal sent)
// - 1: Error (file exists or filesystem error)
func MailmanStop(stdout, stderr io.Writer, opts MailmanStopOptions) int {
    // Find repository root
    repoRoot := opts.RepoRoot
    if repoRoot == "" {
        var err error
        repoRoot, err = mail.FindGitRoot()
        if err != nil {
            repoRoot, _ = os.Getwd()
        }
    }

    stopPath := daemon.StopFilePath(repoRoot)

    // Atomic create - fails if file exists
    f, err := os.OpenFile(stopPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
    if err != nil {
        if os.IsExist(err) {
            fmt.Fprintln(stderr, "Stop already pending")
            return 1
        }
        fmt.Fprintf(stderr, "Failed to send stop signal: %v\n", err)
        return 1
    }
    f.Close()

    fmt.Fprintln(stdout, "Stop signal sent")
    return 0
}
```

### Step 3: Update Daemon File Watcher (internal/daemon/watcher.go)

Modify the file watcher to detect `.stop` file and signal shutdown:

```go
// In Run() method, add check for .stop file:
case event, ok := <-fw.watcher.Events:
    if !ok {
        return nil
    }

    // Check for stop signal file
    if event.Op&fsnotify.Create != 0 && filepath.Base(event.Name) == StopFile {
        // Signal shutdown via channel
        close(fw.stopChan)
        return nil
    }

    // Existing event handling...
```

### Step 4: Update Daemon Shutdown (internal/daemon/daemon.go)

Update `runForeground()` to handle file-based stop:

```go
// Wait for shutdown signal, test stop, or file-based stop
select {
case <-sigChan:
case <-stopChan:
case <-fileWatcher.StopChan():
}

// Remove stop file if it exists
_ = os.Remove(StopFilePath(repoRoot))

// Rest of existing shutdown sequence...
```

### Step 5: Register Subcommand (cmd/agentmail/main.go)

Add `stop` as subcommand to `mailmanCmd`:

```go
stopCmd := &ffcli.Command{
    Name:       "stop",
    ShortUsage: "agentmail mailman stop",
    ShortHelp:  "Stop the mailman daemon",
    Exec: func(ctx context.Context, args []string) error {
        exitCode := cli.MailmanStop(os.Stdout, os.Stderr, cli.MailmanStopOptions{})
        if exitCode != 0 {
            os.Exit(exitCode)
        }
        return nil
    },
}

mailmanCmd := &ffcli.Command{
    // ... existing config ...
    Subcommands: []*ffcli.Command{stopCmd},
}
```

### Step 6: Write Tests

#### internal/cli/mailman_stop_test.go

- Test success case (file created)
- Test "stop already pending" (file exists)
- Test filesystem error (permissions)

#### internal/daemon/watcher_test.go (extend existing)

- Test stop file detection triggers shutdown

## Key Files to Modify

| File | Change |
|------|--------|
| `internal/daemon/daemon.go` | Add StopFile constant, StopFilePath function |
| `internal/daemon/watcher.go` | Add stop file detection, StopChan() method |
| `internal/cli/mailman_stop.go` | NEW: Stop command handler |
| `internal/cli/mailman_stop_test.go` | NEW: Tests |
| `cmd/agentmail/main.go` | Add stop subcommand |

## Testing Commands

```bash
# Run all tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# Run only stop-related tests
go test -v -race ./internal/cli/... -run MailmanStop
go test -v -race ./internal/daemon/... -run Stop

# Check coverage percentage
go tool cover -func=coverage.out | grep total

# Verify quality gates
gofmt -l .
go vet ./...
govulncheck ./...
gosec ./...
```

## Manual Testing

```bash
# Build
go build -o agentmail ./cmd/agentmail

# Start daemon
./agentmail mailman --daemon

# Verify running
cat .agentmail/mailman.pid

# Stop daemon
./agentmail mailman stop
# Output: Stop signal sent

# Verify stopped (wait a moment)
cat .agentmail/mailman.pid  # Should not exist
cat .agentmail/.stop        # Should not exist (daemon removes it)

# Test "stop already pending"
touch .agentmail/.stop
./agentmail mailman stop
# Output: Stop already pending
rm .agentmail/.stop
```

# Research: Mailman Stop Command

**Feature**: 012-mailman-stop
**Date**: 2026-01-20

## Research Questions

### RQ-1: File-Based IPC Mechanism

**Question**: How to reliably signal the daemon to stop using a file?

**Decision**: Create an empty `.stop` file in `.agentmail/` directory. Daemon detects via existing fsnotify watcher.

**Rationale**:
- File creation is a simple, cross-platform IPC mechanism
- The daemon already has fsnotify watching `.agentmail/` for mailbox changes
- No need for Unix signals, process validation, or syscall dependencies
- Works on all platforms (Windows, macOS, Linux)

**Alternatives Considered**:

| Approach | Pros | Cons | Decision |
|----------|------|------|----------|
| SIGTERM signal | Standard Unix pattern | Unix-only, requires process validation | ❌ Rejected |
| Named pipe/FIFO | Bidirectional communication | Complex, Unix-only | ❌ Rejected |
| Unix socket | Reliable IPC | Complex setup, Unix-only | ❌ Rejected |
| File creation | Simple, cross-platform, uses existing watcher | One-way signal only | ✅ Selected |

### RQ-2: Atomic File Creation for Exclusivity

**Question**: How to ensure only one stop signal can be pending at a time?

**Decision**: Use `os.OpenFile` with `O_CREATE|O_EXCL` flags to atomically create the file only if it doesn't exist.

**Rationale**:
- `O_EXCL` flag causes the open to fail if file exists
- This is atomic at the filesystem level
- No race conditions between check and create
- Go's `os.OpenFile` supports this directly

**Implementation**:
```go
// Atomic create - fails if file exists
f, err := os.OpenFile(stopFilePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
if err != nil {
    if os.IsExist(err) {
        return "Stop already pending"
    }
    return fmt.Sprintf("Failed to send stop signal: %v", err)
}
f.Close()
return "Stop signal sent"
```

### RQ-3: Daemon Stop File Detection

**Question**: How should the daemon detect the `.stop` file?

**Decision**: Extend the existing `FileWatcher` to detect CREATE events for `.stop` file in `.agentmail/` directory.

**Rationale**:
- The daemon already watches `.agentmail/` and `mailboxes/` directories
- fsnotify provides CREATE events for new files
- No additional polling or infrastructure needed
- Detection is nearly instant (< 100ms typically)

**Implementation Notes**:
- The watcher's `Run()` function receives all events
- Filter for `fsnotify.Create` events where filename is `.stop`
- When detected, trigger graceful shutdown sequence

### RQ-4: Graceful Shutdown Sequence

**Question**: What's the correct order for daemon shutdown?

**Decision**: Follow this sequence:
1. Detect `.stop` file
2. Remove `.stop` file (acknowledge receipt)
3. Close file watcher (stops notification loop)
4. Wait for notification loop to finish
5. Remove PID file
6. Exit with code 0

**Rationale**:
- Removing `.stop` first prevents stale signal files
- Closing watcher before PID removal ensures clean state
- Matches existing signal-based shutdown sequence in `runForeground()`

**Existing Code Reference** (`internal/daemon/daemon.go:249-271`):
```go
// Wait for shutdown signal or test stop
<-sigChan

// Close file watcher to stop the notification loop
if fileWatcher != nil {
    _ = fileWatcher.Close()
}

<-loopDone // Wait for loop to finish

// Clean up PID file on shutdown
_ = DeletePID(repoRoot)

return 0
```

The new approach will add `.stop` file detection alongside signal handling.

## Dependencies Review

**No new dependencies required.** This approach simplifies dependencies:

| Before (SIGTERM) | After (File-based) |
|------------------|-------------------|
| os | os |
| os/exec | ~~removed~~ |
| syscall | ~~removed~~ |
| strings | ~~removed~~ |

The file-based approach uses only the `os` package for file operations.

## Conclusion

All research questions resolved. The file-based approach is simpler, more cross-platform, and leverages existing infrastructure (fsnotify watcher). Implementation requires:

1. **Stop command**: ~20 lines of code (check exists, create file, print message)
2. **Daemon changes**: ~10 lines (detect .stop file, add to shutdown sequence)

No process validation, signal handling, or platform-specific code needed.

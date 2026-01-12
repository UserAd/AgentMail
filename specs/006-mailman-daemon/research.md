# Research: Mailman Daemon

**Feature**: 006-mailman-daemon
**Date**: 2026-01-12

## Research Topics

### 1. Go Daemon Process Management

**Decision**: Use `os.StartProcess` with `syscall.Setsid()` for daemonization

**Rationale**:
- Go standard library provides all necessary primitives
- `os.StartProcess` allows re-executing the binary with detached process group
- `syscall.Setsid()` creates new session, detaching from controlling terminal
- No external dependencies required (per constitution IV)

**Alternatives Considered**:
- External daemon library (go-daemon): Rejected - adds dependency
- systemd integration: Rejected - platform-specific, over-engineered for use case
- Simple background with `&`: Rejected - doesn't properly detach, PID tracking unreliable

**Implementation Pattern**:
```go
// Foreground mode (default): just run the loop
// Background mode (--daemon): fork and exit parent
if daemonMode {
    // Re-exec with special env var to indicate child
    if os.Getenv("AGENTMAIL_DAEMON_CHILD") == "" {
        // Parent: fork child and exit
        cmd := exec.Command(os.Args[0], os.Args[1:]...)
        cmd.Env = append(os.Environ(), "AGENTMAIL_DAEMON_CHILD=1")
        cmd.Start()
        os.Exit(0)
    }
    // Child: detach from terminal
    syscall.Setsid()
}
```

### 2. PID File Management

**Decision**: Simple text file with PID, use `os.FindProcess` + signal 0 for liveness check

**Rationale**:
- Standard Unix pattern, well-understood
- No file locking needed for PID file itself (atomic write)
- Signal 0 is portable check for process existence
- Stale PID detection via process check, not file age

**Alternatives Considered**:
- File locking on PID file: Rejected - unnecessary complexity, process check sufficient
- Advisory lock files: Rejected - doesn't survive crashes well
- Socket-based locking: Rejected - over-engineered

**Implementation Pattern**:
```go
func isProcessRunning(pid int) bool {
    process, err := os.FindProcess(pid)
    if err != nil {
        return false
    }
    // Signal 0 checks if process exists without sending signal
    err = process.Signal(syscall.Signal(0))
    return err == nil
}
```

### 3. Signal Handling in Go

**Decision**: Use `os/signal.Notify` with buffered channel for SIGTERM/SIGINT

**Rationale**:
- Go's signal package is idiomatic and safe
- Buffered channel prevents signal loss
- Clean shutdown allows PID file cleanup

**Implementation Pattern**:
```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

go func() {
    <-sigChan
    cleanup()
    os.Exit(0)
}()
```

### 4. Tmux send-keys Command

**Decision**: Use `tmux send-keys -t <window>` for message delivery

**Rationale**:
- Existing `internal/tmux` package handles tmux detection
- `send-keys` is the standard way to inject input to tmux panes
- Can target specific windows by name

**Implementation Pattern**:
```go
func SendKeys(window, text string) error {
    cmd := exec.Command("tmux", "send-keys", "-t", window, text)
    return cmd.Run()
}

func SendEnter(window string) error {
    cmd := exec.Command("tmux", "send-keys", "-t", window, "Enter")
    return cmd.Run()
}
```

### 5. JSONL File Format for Recipients State

**Decision**: Single JSONL file with one line per recipient, rewrite on update

**Rationale**:
- Consistent with existing mailbox pattern
- Simple to parse and maintain
- File locking prevents race conditions
- Small file size (few recipients) makes rewrite efficient

**Schema**:
```json
{"recipient": "agent1", "status": "ready", "updated_at": "2026-01-12T10:00:00Z", "notified": false}
{"recipient": "agent2", "status": "work", "updated_at": "2026-01-12T10:01:00Z", "notified": true}
```

**Alternatives Considered**:
- Separate file per recipient: Rejected - adds complexity for small dataset
- SQLite: Rejected - adds dependency, over-engineered
- In-memory only: Rejected - state lost on restart

### 6. Notification Loop Timing

**Decision**: Simple `time.Sleep(10 * time.Second)` loop

**Rationale**:
- 10-second interval specified in requirements
- Simple sleep is sufficient for this use case
- No need for ticker complexity

**Implementation Pattern**:
```go
for {
    checkAndNotify()
    time.Sleep(10 * time.Second)
}
```

### 7. Detecting Unread Messages

**Decision**: Reuse existing `mail.FindUnread()` function, iterate over all recipients

**Rationale**:
- Existing function already handles JSONL parsing and filtering
- Need to check all recipients in state file
- List mailbox files in `.git/mail/` to find all recipients

**Implementation Pattern**:
```go
// Get all .jsonl files in .git/mail/ (excluding recipients.jsonl and mailman.pid)
// For each file, extract recipient name and check unread count
files, _ := filepath.Glob(filepath.Join(repoRoot, ".git/mail/*.jsonl"))
for _, f := range files {
    recipient := strings.TrimSuffix(filepath.Base(f), ".jsonl")
    if recipient == "recipients" {
        continue // Skip state file
    }
    unread, _ := mail.FindUnread(repoRoot, recipient)
    if len(unread) > 0 {
        // Check if recipient is ready and not notified
    }
}
```

## Resolved Clarifications

All technical context items resolved - no NEEDS CLARIFICATION markers remain.

| Item | Resolution |
|------|------------|
| Daemonization | Go stdlib with re-exec pattern |
| PID management | Signal 0 for liveness check |
| Signal handling | os/signal package |
| Tmux integration | Extend existing package |
| State storage | JSONL file with locking |
| Loop timing | Simple sleep loop |

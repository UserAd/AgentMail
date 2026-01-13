# Quickstart: File-Watching for Mailman with Timer Fallback

**Feature**: 009-watch-files
**Date**: 2026-01-13

## Overview

This feature enhances the mailman daemon with file-system watching for instant notification delivery, replacing the 10-second polling interval. It also adds last-read tracking to monitor when agents check their mail.

## What Changes

### User-Visible Changes

1. **Faster notifications**: Messages are delivered within 2 seconds instead of up to 10 seconds
2. **Automatic fallback**: If file watching fails, the system falls back to polling (no user action needed)
3. **Last-read tracking**: The system now tracks when agents last read their mail

### Developer Changes

1. **New dependency**: `github.com/fsnotify/fsnotify` v1.9.0
2. **Modified files**:
   - `internal/daemon/daemon.go` - Watcher integration
   - `internal/daemon/loop.go` - Event-driven mode
   - `internal/daemon/watcher.go` - NEW: File watcher abstraction
   - `internal/mail/recipients.go` - LastReadAt field
   - `internal/cli/receive.go` - Update last-read timestamp

## Development Setup

### 1. Add fsnotify dependency

```bash
go get github.com/fsnotify/fsnotify@v1.9.0
```

### 2. Run tests

```bash
go test -v -race ./...
```

### 3. Build

```bash
go build -o agentmail ./cmd/agentmail
```

## Testing the Feature

### Test 1: Instant Notifications (File Watching)

```bash
# Terminal 1: Start mailman
./agentmail mailman

# Look for "File watching enabled" message

# Terminal 2: Set agent to ready
./agentmail status ready

# Terminal 3: Send a message
./agentmail send agent1 "Test message"

# Verify: Notification appears in Terminal 2 within 2 seconds
```

### Test 2: Fallback Mode

```bash
# Simulate fallback by running on network filesystem or
# setting FSNOTIFY_DEBUG=1 to see watcher behavior

# When fallback activates, you'll see:
# "File watching unavailable, using polling"

# Notifications still work, but with 10-second delay
```

### Test 3: Last-Read Tracking

```bash
# Read mail
./agentmail receive

# Check recipients.jsonl for last_read_at field
cat .agentmail/recipients.jsonl | jq .
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Mailman Daemon                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   ┌─────────────┐     ┌─────────────┐     ┌─────────────┐  │
│   │ FileWatcher │────►│  Debouncer  │────►│ CheckAndNotify│ │
│   │ (fsnotify)  │     │  (500ms)    │     │             │  │
│   └──────┬──────┘     └─────────────┘     └─────────────┘  │
│          │                                                  │
│          │ watches                                          │
│          ▼                                                  │
│   ┌─────────────────────────────────────┐                  │
│   │ .agentmail/                         │                  │
│   │  ├── recipients.jsonl  ◄──── status changes           │
│   │  └── mailboxes/        ◄──── new/updated mail         │
│   └─────────────────────────────────────┘                  │
│                                                             │
│   ┌─────────────────────────────────────┐                  │
│   │ Fallback Timer (60s)               │◄──── safety net   │
│   └─────────────────────────────────────┘                  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Troubleshooting

### "File watching unavailable" message

**Cause**: OS doesn't support file watching (rare), network filesystem, or watcher resource limit

**Solution**: This is normal fallback behavior. The daemon continues with 10-second polling.

### High CPU usage

**Cause**: Burst of file events not being debounced

**Solution**: Check if another process is rapidly modifying `.agentmail/` files. The 500ms debounce should prevent this.

### "too many open files" error

**Cause**: macOS/BSD kqueue opens file descriptor per watched file

**Solution**: Increase ulimit (`ulimit -n 2048`) or reduce number of mailbox files

## Quality Gates

Before marking complete, verify:

- [ ] `go test -cover ./...` reports >= 80% coverage
- [ ] `go vet ./...` passes
- [ ] `go fmt ./...` produces no changes
- [ ] All acceptance scenarios from spec.md pass
- [ ] `go test -race ./...` passes

# Research: AgentMail Initial Project Structure

**Feature**: 001-agent-mail-structure
**Date**: 2026-01-11

## Research Tasks

### 1. Go CLI Best Practices

**Decision**: Use standard library with `cmd/` and `internal/` layout

**Rationale**:
- Standard Go project layout (`cmd/` for binaries, `internal/` for private packages) is well-documented and widely understood
- No external CLI framework needed for two simple commands (send/receive)
- Direct argument parsing via `os.Args` is sufficient for MVP
- Keeps dependencies minimal and binary size small

**Alternatives Considered**:
- **cobra/viper**: Overkill for 2 commands, adds unnecessary complexity
- **urfave/cli**: Good for larger CLIs, but MVP doesn't need subcommand infrastructure
- **flag package**: Could work, but manual arg handling is cleaner for positional args

### 2. JSONL File Handling in Go

**Decision**: Use `encoding/json` with line-by-line file reading/writing, one file per recipient

**Rationale**:
- Standard library `encoding/json` handles marshaling/unmarshaling
- JSONL is simply one JSON object per line - read line, unmarshal, process
- Per-recipient files (`.git/mail/<recipient>.jsonl`) improve concurrency:
  - No filtering needed when reading (all messages in file are for recipient)
  - Sends to different recipients don't block each other
  - Smaller files, faster reads
- File locking needed for concurrent access (use `flock` via syscall)
- Append-only for writes, full read for queries

**Alternatives Considered**:
- **Single shared file**: More contention, requires recipient filtering
- **BoltDB/BadgerDB**: Embedded DB is overkill for simple message storage
- **SQLite**: Too heavy for MVP, adds CGO dependency
- **Plain JSON array**: Harder to append without reading entire file

### 3. tmux Integration

**Decision**: Use `os/exec` to call tmux commands directly

**Rationale**:
- `tmux display-message -p '#W'` returns current window name
- `tmux list-windows -F '#{window_name}'` returns all window names
- Simple exec calls, parse stdout, handle errors
- No need for tmux library/bindings

**Alternatives Considered**:
- **tmux control mode**: More complex, not needed for simple queries
- **Environment variables**: `$TMUX` indicates tmux session but not window name

### 4. Unique ID Generation

**Decision**: Use `crypto/rand` with base62 encoding for short IDs

**Rationale**:
- 8-character base62 ID provides ~47 bits of entropy (sufficient for local use)
- Human-readable and copy-pasteable
- No external dependency needed
- Format: `[a-zA-Z0-9]{8}` (e.g., "xK7mN2pQ")

**Alternatives Considered**:
- **UUID**: Too long for display (36 chars)
- **Timestamp-based**: Collision risk with concurrent sends
- **Sequential**: Requires state management

### 5. File Locking Strategy

**Decision**: Use advisory file locking via `syscall.Flock` on per-recipient files

**Rationale**:
- Prevents concurrent write corruption
- Standard POSIX approach works on macOS/Linux
- Lock only the recipient's file during write operations
- Per-recipient files reduce contention (agents writing to different recipients don't block)
- Read operations can proceed without locks (eventual consistency acceptable for MVP)

**Alternatives Considered**:
- **Separate lock file**: Extra complexity
- **Atomic rename**: Works for full rewrites but not appends
- **No locking**: Risk of data corruption with concurrent agents
- **Global lock**: Would serialize all operations, poor concurrency

### 6. Error Exit Codes

**Decision**: Follow Unix conventions with specific exit codes

**Rationale**:
- Exit 0: Success (including "no messages" case per spec)
- Exit 1: General error (file I/O, invalid arguments)
- Exit 2: Not running in tmux (per FR-005b)
- Standard convention makes scripting easier

**Alternatives Considered**:
- **Single non-zero code**: Less informative for automation
- **Negative codes**: Not portable

## Summary

All technical decisions favor simplicity and standard library usage. No external dependencies required beyond Go standard library. The architecture follows Go conventions and should be straightforward to implement and maintain.

| Area | Decision | Dependency |
|------|----------|------------|
| CLI Framework | os.Args + manual parsing | stdlib |
| JSON Handling | encoding/json, per-recipient files | stdlib |
| tmux Integration | os/exec | stdlib |
| ID Generation | crypto/rand + base62 | stdlib |
| File Locking | syscall.Flock (per-recipient) | stdlib |
| Storage | `.git/mail/<recipient>.jsonl` | n/a |
| Exit Codes | 0/1/2 convention | n/a |

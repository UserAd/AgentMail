# CLI Contract: Mailman Stop Command

**Feature**: 012-mailman-stop
**Date**: 2026-01-20

## Command Signature

```text
agentmail mailman stop
```

**No flags or arguments required.**

## Exit Codes

| Code | Meaning | Condition |
|------|---------|-----------|
| 0 | Success | Stop signal file created |
| 1 | Error | File already exists or filesystem error |

## Output Specification

### Success (Exit 0)

**Stream**: stdout
**Format**: `Stop signal sent\n`

```text
Stop signal sent
```

### Error Conditions (Exit 1)

All error messages are written to stderr.

#### Stop Already Pending

```text
Stop already pending
```

**Condition**: `.agentmail/.stop` file already exists.

#### Filesystem Error

```text
Failed to send stop signal: <error message>
```

**Condition**: Unable to create `.stop` file due to permissions, disk space, etc.

## Behavior Specification

### Fire-and-Forget Pattern

The stop command creates a signal file and immediately exits with code 0. It does NOT:
- Wait for the daemon to terminate
- Verify the daemon is running
- Verify the daemon received the signal
- Delete any files

The daemon is responsible for:
- Detecting the `.stop` file via file watcher
- Removing the `.stop` file
- Removing the `.pid` file
- Graceful shutdown

### File-Based IPC

The stop mechanism uses file creation as inter-process communication:

1. Stop command checks if `.agentmail/.stop` exists
2. If exists → error "Stop already pending"
3. If not exists → create the file
4. Daemon's file watcher detects CREATE event
5. Daemon initiates shutdown and cleans up files

This approach:
- Requires no process validation
- Works cross-platform
- Uses existing file watcher infrastructure
- Simplifies error handling

## Integration with Existing Commands

### Relationship to `agentmail mailman`

```text
agentmail mailman          # Start daemon (foreground)
agentmail mailman --daemon # Start daemon (background)
agentmail mailman stop     # Stop daemon (NEW)
```

The `stop` subcommand is a peer to the existing start behavior, not a flag.

### CLI Help Output (Expected)

```text
$ agentmail mailman --help
Start or stop the mailman daemon

Usage:
  agentmail mailman [--daemon]
  agentmail mailman stop

Commands:
  stop    Stop the running mailman daemon

Flags:
  --daemon    Run in background (daemonize)

Exit codes:
  0  Success
  1  Error
  2  Daemon already running (start only)
```

## Test Scenarios

### Happy Path

```bash
# Start daemon
$ agentmail mailman --daemon
Mailman daemon started in background (PID: 12345)

# Stop daemon
$ agentmail mailman stop
Stop signal sent
$ echo $?
0

# Verify daemon stopped (after brief delay)
$ cat .agentmail/mailman.pid
cat: .agentmail/mailman.pid: No such file or directory
$ cat .agentmail/.stop
cat: .agentmail/.stop: No such file or directory
```

### Stop Already Pending

```bash
# Simulate pending stop
$ touch .agentmail/.stop

$ agentmail mailman stop
Stop already pending
$ echo $?
1
```

### No Daemon Running (Still Succeeds)

```bash
# No daemon running, but command still creates file
$ agentmail mailman stop
Stop signal sent
$ echo $?
0

# File exists until something removes it
$ ls .agentmail/.stop
.agentmail/.stop
```

**Note**: The stop command doesn't verify if a daemon is running. It simply creates the signal file. If no daemon is running, the file will remain until manually deleted or a daemon starts and detects it.

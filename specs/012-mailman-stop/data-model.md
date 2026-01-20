# Data Model: Mailman Stop Command

**Feature**: 012-mailman-stop
**Date**: 2026-01-20

## Overview

This feature introduces a file-based signaling mechanism for stopping the daemon. The stop command creates a signal file that the daemon detects via its file watcher.

## New Entity

### Stop Signal File

**Location**: `.agentmail/.stop`
**Format**: Empty file (presence indicates stop request)
**Permissions**: `0600` (owner read/write only)
**Lifecycle**: Created by stop command, removed by daemon during shutdown

**Purpose**: Acts as an inter-process communication (IPC) mechanism between the CLI stop command and the running daemon.

**Operations**:

| Operation | Actor | Trigger |
|-----------|-------|---------|
| Create | CLI (stop command) | User runs `agentmail mailman stop` |
| Detect | Daemon (file watcher) | fsnotify CREATE event |
| Delete | Daemon | During graceful shutdown |

## Existing Entity (Read-Only by Stop Command)

### Mailman PID File

**Location**: `.agentmail/mailman.pid`
**Format**: Plain text, single line containing numeric PID followed by newline
**Permissions**: `0600` (owner read/write only)

**Note**: The stop command does NOT read or modify the PID file. The daemon removes this file during shutdown.

## State Transitions

```text
Stop Command Flow (Simplified):

[Start: agentmail mailman stop]
    │
    ▼
┌─────────────────────┐
│ Check .stop exists? │
└──────────┬──────────┘
           │
      Yes ─┼─ No
           │    │
           ▼    ▼
      [Exit 1]  ┌─────────────────┐
      "Stop     │ Create .stop    │
      already   └────────┬────────┘
      pending"           │
                    ┌────┴────┐
                    │ Success?│
                    └────┬────┘
                         │
                    No ──┼── Yes
                         │    │
                         ▼    ▼
                    [Exit 1] [Exit 0]
                    "Failed" "Stop signal
                             sent"
```

```text
Daemon Shutdown Flow:

[Daemon Running]
    │
    ▼
┌─────────────────────────┐
│ File watcher monitoring │
│ .agentmail/ directory   │
└──────────┬──────────────┘
           │
           │ CREATE event for .stop
           ▼
┌─────────────────────────┐
│ Initiate graceful       │
│ shutdown                │
└──────────┬──────────────┘
           │
           ▼
┌─────────────────────────┐
│ Remove .stop file       │
└──────────┬──────────────┘
           │
           ▼
┌─────────────────────────┐
│ Close file watcher      │
│ Wait for loop done      │
└──────────┬──────────────┘
           │
           ▼
┌─────────────────────────┐
│ Remove mailman.pid      │
└──────────┬──────────────┘
           │
           ▼
      [Exit 0]
```

## File System Layout

```text
.agentmail/
├── mailman.pid      # Daemon PID (existing, created by daemon)
├── .stop            # Stop signal (NEW, created by stop command)
├── recipients.jsonl # Agent states (existing)
└── mailboxes/       # Message storage (existing)
    └── *.jsonl
```

## Comparison: Old vs New Approach

| Aspect | Old (SIGTERM) | New (File-based) |
|--------|---------------|------------------|
| IPC Mechanism | Unix signals | File creation |
| Validation | Process name check | None (fire-and-forget) |
| Dependencies | syscall, os/exec | os (file ops only) |
| Cross-platform | Unix only | All platforms |
| Daemon detection | Instant (signal) | Instant (fsnotify) |
| Error scenarios | Permission denied, wrong process | File exists, filesystem error |

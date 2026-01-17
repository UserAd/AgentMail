# Quickstart: Cleanup Command

**Feature**: 011-cleanup
**Date**: 2026-01-15

## Overview

The `agentmail cleanup` command removes stale data from the AgentMail system:
- Recipients no longer present as tmux windows (offline)
- Recipients inactive for extended periods (stale)
- Read messages older than a threshold
- Empty mailbox files

## Usage

### Basic Cleanup

```bash
agentmail cleanup
```

Runs cleanup with default thresholds:
- Stale recipients: 48 hours
- Delivered messages: 2 hours

### Custom Thresholds

```bash
# Remove recipients inactive for 24 hours
agentmail cleanup --stale-hours 24

# Remove read messages older than 1 hour
agentmail cleanup --delivered-hours 1

# Both options together
agentmail cleanup --stale-hours 24 --delivered-hours 1
```

### Dry Run (Preview)

```bash
agentmail cleanup --dry-run
```

Shows what would be cleaned without making changes.

## Output

### Normal Execution

```
Cleanup complete:
  Recipients removed: 3 (2 offline, 1 stale)
  Messages removed: 15
  Mailboxes removed: 2
```

### Dry Run

```
Cleanup preview (dry-run):
  Recipients to remove: 3 (2 offline, 1 stale)
  Messages to remove: 15
  Mailboxes to remove: 2
```

### With Warnings

```
Warning: Skipped 1 locked file(s)
Cleanup complete:
  Recipients removed: 2 (1 offline, 1 stale)
  Messages removed: 10
  Mailboxes removed: 1
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success (cleanup completed) |
| 1 | Error (invalid arguments, file system error) |

## Flags Reference

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--stale-hours` | int | 48 | Remove recipients not updated within N hours |
| `--delivered-hours` | int | 2 | Remove read messages older than N hours |
| `--dry-run` | bool | false | Preview changes without making them |

## Behavior Notes

1. **Non-tmux environments**: The offline recipient check is skipped when not running inside tmux. Other cleanup operations proceed normally.

2. **Locked files**: If a file cannot be locked within 1 second, it is skipped with a warning. Re-run cleanup later to process skipped files.

3. **Message timestamps**: Only messages with a `created_at` timestamp are eligible for age-based cleanup. Messages without this field (legacy) are skipped.

4. **Unread messages**: Messages with `read_flag: false` are NEVER deleted regardless of age.

## Examples

### Scheduled Cleanup (cron)

```bash
# Run daily at 3 AM
0 3 * * * cd /path/to/project && agentmail cleanup >> /var/log/agentmail-cleanup.log 2>&1
```

### Conservative Cleanup (keep more data)

```bash
agentmail cleanup --stale-hours 168 --delivered-hours 24
```
Keeps recipients for 7 days, messages for 24 hours.

### Aggressive Cleanup (remove more data)

```bash
agentmail cleanup --stale-hours 1 --delivered-hours 0
```
Removes recipients inactive for 1 hour, removes all read messages immediately.

## Integration

### With Mailman Daemon

Cleanup can run while the mailman daemon is active. Locked files are skipped, and cleanup can be retried.

### In Scripts

```bash
#!/bin/bash
# Check if cleanup needed before running
if [ -d ".agentmail/mailboxes" ]; then
    agentmail cleanup --dry-run
    read -p "Proceed with cleanup? (y/n) " confirm
    if [ "$confirm" = "y" ]; then
        agentmail cleanup
    fi
fi
```

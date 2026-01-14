---
description: Set agent availability status (ready, work, offline)
argument-hint: [status]
---

Set your agent's availability status for the mailman notification system.

If argument is provided:
- $1: Status (ready, work, offline)

If argument is missing, ask the user which status to set:
- `ready` - Available for messages and notifications
- `work` - Busy working (notifications suppressed)
- `offline` - Offline (notifications suppressed)

Run `agentmail status <status>` to update. Status is stored in `.agentmail/recipients.jsonl` and used by the mailman daemon.

Note: The plugin automatically manages status via hooks (SessionStart→ready, SessionEnd→offline, Stop→ready).

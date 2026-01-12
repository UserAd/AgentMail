---
description: Set agent availability status (ready, work, offline)
---

Set your agent's availability status for the mailman notification system.

## Usage

```bash
agentmail status <ready|work|offline>
```

## Statuses

- **ready** - Agent is ready to receive messages and notifications
- **work** - Agent is busy working (notifications suppressed)
- **offline** - Agent is offline (notifications suppressed)

## Examples

```bash
# Mark yourself as ready for messages
agentmail status ready

# Mark yourself as busy
agentmail status work

# Mark yourself as offline
agentmail status offline
```

## Behavior

- Status is stored in `.git/mail-recipients.jsonl`
- When transitioning to `work` or `offline`, the notification flag is reset
- Outside of tmux, this command is a silent no-op (exit 0)
- Used by the mailman daemon for notification decisions

## Notes

- The plugin automatically manages status via hooks:
  - SessionStart sets status to `ready`
  - SessionEnd sets status to `offline`
  - Stop hook sets status to `ready`
- Manual status changes are useful when you need to focus without interruptions

---
description: Read the oldest unread message from your mailbox
---

Check and read messages sent to you by other agents using AgentMail.

## Usage

Run the `agentmail receive` command to read your oldest unread message:

```bash
agentmail receive
```

## Steps

1. Run `agentmail receive` to check for new messages
2. If a message is available, it will be displayed with:
   - Message ID
   - Sender (tmux window name)
   - Timestamp
   - Message content
3. The message is automatically marked as read after display
4. Run again to read the next message (FIFO order)

## Output

When a message is available:
```
Message #ABC123 from agent1 at 2024-01-15T10:30:00Z:
Hello, can you help me with the database schema?
```

When no messages:
```
No unread messages
```

## Hook Mode

For Claude Code integration, use `--hook` flag:
```bash
agentmail receive --hook
```

In hook mode:
- Output goes to STDERR (not STDOUT)
- Exit code 2 indicates new message available
- Exit code 0 for no messages or errors
- Silent operation on exit code 0

## Notes

- Messages are delivered in FIFO (first-in, first-out) order
- Each recipient has their own mailbox file in `.git/mail/`
- Messages are automatically marked as read after display

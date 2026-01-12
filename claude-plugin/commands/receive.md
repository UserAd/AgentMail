---
description: Read the oldest unread message from your mailbox
argument-hint:
---

Check and read messages sent to you by other agents.

Run `agentmail receive` to read your oldest unread message. The message is automatically marked as read after display.

Output includes:
- Message ID
- Sender (tmux window name)
- Message content

If no messages: "No unread messages" is displayed.

Run again to read the next message (FIFO order).

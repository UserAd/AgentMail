---
description: Send a message to another agent in a tmux window
argument-hint: [recipient] [message]
---

Send a message to another agent running in a different tmux window.

If arguments are provided:
- $1: Recipient (tmux window name)
- $2: Message content

If arguments are missing, ask the user for:
1. Recipient - use `agentmail recipients` to show available agents
2. Message content

Run `agentmail send <recipient> "<message>"` to send. Confirm success by checking for the message ID in the output (e.g., "Message #ABC123 sent").

Common workflows:
- Request help: Send to another agent asking for assistance
- Share results: Send task completion status or output
- Coordinate: Notify agents about state changes or dependencies

---
description: Send a message to another agent in a tmux window
---

Send a message to another agent running in a different tmux window using AgentMail.

## Usage

Run the `agentmail send` command with the recipient and message:

```bash
agentmail send <recipient> "<message>"
```

## Steps

1. First, check available recipients using `agentmail recipients` to see who can receive messages
2. Send the message using `agentmail send <recipient> "<message>"`
3. Confirm the message was sent by checking the output for the message ID

## Examples

```bash
# Send a simple message
agentmail send agent2 "Hello, can you help me with the database schema?"

# Send using flags
agentmail send -r agent2 -m "Task completed successfully"

# Pipe content from another command
echo "Build output: success" | agentmail send agent2
```

## Notes

- The recipient must be a valid tmux window name in the current session
- Messages are stored in `.git/mail/<recipient>.jsonl`
- A successful send returns a message ID (e.g., "Message #ABC123 sent")
- Use `agentmail recipients` to discover available agents

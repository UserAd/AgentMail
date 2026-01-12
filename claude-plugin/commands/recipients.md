---
description: List available message recipients (other agents in tmux)
---

List all tmux windows in the current session that can receive messages.

## Usage

```bash
agentmail recipients
```

## Output

Shows all available tmux windows with your current window marked:

```
agent1 [you]
agent2
worker
orchestrator
```

## Notes

- The current window is marked with `[you]`
- Windows listed in `.agentmailignore` are excluded
- Only windows in the current tmux session are shown
- Use this to discover which agents are available before sending messages

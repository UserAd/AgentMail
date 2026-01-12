# Quickstart: Claude Code Hooks Integration

**Feature**: 005-claude-hooks-integration
**Date**: 2026-01-12

## Prerequisites

- AgentMail installed (`agentmail` in PATH)
- Claude Code CLI installed
- Running inside a tmux session

## Quick Test

```bash
# Test hook mode with no messages (should exit silently with code 0)
agentmail receive --hook
echo $?  # Should output: 0

# Send a test message from another tmux window, then:
agentmail receive --hook
echo $?  # Should output: 2 (and message shown on STDERR)
```

## Claude Code Configuration

### Option 1: Project-level settings

Create or edit `.claude/settings.json` in your project root:

```json
{
  "hooks": {
    "user-prompt-submit": {
      "command": "agentmail receive --hook"
    }
  }
}
```

### Option 2: User-level settings

Edit `~/.claude/settings.json`:

```json
{
  "hooks": {
    "user-prompt-submit": {
      "command": "agentmail receive --hook"
    }
  }
}
```

## How It Works

1. Before each prompt submission, Claude Code runs `agentmail receive --hook`
2. If you have mail from another agent:
   - Message appears in Claude Code's output (via STDERR)
   - Exit code 2 triggers Claude Code's notification behavior
3. If no mail:
   - Silent exit (no output, exit code 0)
   - Your workflow continues uninterrupted

## Exit Code Reference

| Exit Code | Meaning | Output |
|-----------|---------|--------|
| 0 | No messages / not in tmux / error | Silent (no output) |
| 2 | New message received | Message on STDERR |

## Troubleshooting

### Hook not triggering?

1. Verify you're in a tmux session: `echo $TMUX`
2. Verify agentmail is in PATH: `which agentmail`
3. Check Claude Code settings are valid JSON

### Not seeing messages?

1. Verify sender used correct window name: `agentmail recipients`
2. Check for messages directly: `agentmail receive` (without --hook)

### Getting errors instead of silent exit?

Make sure you're using `--hook` flag. Without it, errors are displayed normally.

## Development Testing

Run the test suite to verify hook behavior:

```bash
go test -v ./internal/cli -run TestReceive
```

Expected test coverage for hook mode:
- Hook mode with messages → exit 2, STDERR output
- Hook mode without messages → exit 0, no output
- Hook mode not in tmux → exit 0, no output
- Hook mode with errors → exit 0, no output

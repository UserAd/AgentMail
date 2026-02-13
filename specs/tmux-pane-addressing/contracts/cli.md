# CLI Contract: Tmux Pane Addressing

**Date**: 2026-02-13

## Changed Commands

### `agentmail send <recipient> <message>`

**Before**: `<recipient>` is a tmux window name.
**After**: `<recipient>` is one of:
- Full pane address: `session:window.pane` (e.g., `AgentMail:editor.1`)
- Medium address: `:window.pane` (e.g., `:editor.1`) — colon prefix, session inferred from current
- Short address: `window` (e.g., `editor`) — backward compatible, even for names with dots (e.g., `my.app`)

**Behavior changes**:
- Short address with single-pane window: resolves to full address, delivers message. No visible change to user.
- Short address with multi-pane window: returns error with suggestion:
  ```
  Error: Ambiguous recipient: window 'editor' has 3 panes. Use AgentMail:editor.0, AgentMail:editor.1, or AgentMail:editor.2
  ```
  Exit code: 1
- The `From` field in stored messages always uses full pane address format.
- Recipient validation uses `tmux display-message -t <address>` instead of `tmux list-windows`.

**Exit codes**: No change (0 success, 1 error, 2 environment error).

**Output format**: No change (`Message #<ID> sent`).

### `agentmail receive`

**Before**: Looks up mailbox by current window name.
**After**: Looks up mailbox by current full pane address.

**Behavior changes**:
- The receiver identity is determined by `GetCurrentPaneAddress()` instead of `GetCurrentWindow()`.
- Messages from old window-name-based mailboxes are not visible (orphaned by design).
- Output `From:` line shows full pane address of sender.

**Hook mode**: Same behavior, but polls the pane-specific mailbox.

**Exit codes**: No change.

### `agentmail recipients`

**Before**: Lists tmux window names, marks current with `[you]`.
**After**: Lists full pane addresses, marks current pane with `[you]`.

**Output example**:
```
AgentMail:editor.0
AgentMail:editor.1
AgentMail:code.0 [you]
AgentMail:tests.0
```

### `agentmail status <status>`

**Before**: Updates recipient state keyed by window name.
**After**: Updates recipient state keyed by full pane address.

**No visible output change.**

### `agentmail cleanup`

**Before**: Cleans up window-name-based mailboxes and recipient states.
**After**: Cleans up pane-address-based mailboxes and recipient states. Also removes orphaned legacy mailbox files (files whose names do not match the percent-encoded pane address pattern `<name>%3A<name>%2E<digits>.jsonl`).

### `agentmail mailman start`

**Before**: Daemon tracks windows and notifies by window name.
**After**: Daemon tracks panes and notifies by pane address using `tmux send-keys -t <pane-address>`.

## New Tmux Functions

### `GetCurrentPaneAddress() (string, error)`

Returns the full `session:window.pane` address of the calling process's pane.

### `GetCurrentSession() (string, error)`

Returns the current tmux session name.

### `ListPanes() ([]string, error)`

Returns all pane addresses in the current session as `session:window.pane` strings. This is the default discovery scope for recipients, MCP list-recipients, and daemon.

### `ListAllPanes() ([]string, error)`

Returns all pane addresses across all sessions. Not exposed in any command or MCP tool; available for potential future use.

### `PaneExists(address string) (bool, error)`

Checks if a specific pane exists by targeting it with `tmux display-message`.

## MCP Tool Changes

### `send` tool

**Schema change**: `recipient` parameter description updated to accept pane addresses.
**Behavior**: Same resolution as CLI send command.

### `receive` tool

**Behavior**: Uses current pane address instead of window name.

### `status` tool

**Behavior**: Tracks status at pane level.

### `list-recipients` tool

**Response change**: Returns pane addresses instead of window names.

```json
{
  "recipients": [
    {"address": "AgentMail:editor.0", "is_current": false},
    {"address": "AgentMail:editor.1", "is_current": false},
    {"address": "AgentMail:code.0", "is_current": true}
  ]
}
```

**Breaking change**: The `window` field is replaced by `address` (full pane address). No backward-compatible `window` field is included.

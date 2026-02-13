# Data Model: Tmux Pane Addressing

**Feature Branch**: `tmux-pane-addressing`
**Date**: 2026-02-13

## Entity Changes

### PaneAddress (New)

Represents a parsed tmux pane address with its components.

| Field   | Type   | Description                        | Example        |
|---------|--------|------------------------------------|----------------|
| Session | string | Tmux session name                  | `AgentMail`    |
| Window  | string | Tmux window name                   | `editor`       |
| Pane    | int    | Pane index within the window       | `1`            |
| Full    | string | Canonical `session:window.pane`    | `AgentMail:editor.1` |

**Validation rules**:
- Session: non-empty string
- Window: non-empty string
- Pane: non-negative integer
- Full: must match pattern `<session>:<window>.<pane>` where pane is a non-negative integer

**Construction**:
- `ParseAddress(input, currentSession)` → PaneAddress or error
- `FormatAddress(session, window, pane)` → string (`session:window.pane`)

### Message (Modified)

The `From` and `To` fields change from window names to full pane addresses.

| Field     | Type      | Before              | After                          |
|-----------|-----------|---------------------|--------------------------------|
| ID        | string    | 8-char base62       | No change                      |
| From      | string    | Window name         | Full pane address (`session:window.pane`) |
| To        | string    | Window name         | Full pane address (`session:window.pane`) |
| Message   | string    | Body text           | No change                      |
| ReadFlag  | bool      | Read status         | No change                      |
| CreatedAt | time.Time | Timestamp           | No change                      |

**JSON wire format**: Unchanged structure, only the content of `from`/`to` changes.

**Backward compatibility**: Old messages with window-name-only `from`/`to` are stored in orphaned mailbox files that are not queried by the new code. They are effectively inaccessible via normal commands and should be cleaned up with `agentmail cleanup`.

### RecipientState (Modified)

The `Recipient` field changes from window name to full pane address.

| Field      | Type      | Before        | After                          |
|------------|-----------|---------------|--------------------------------|
| Recipient  | string    | Window name   | Full pane address (`session:window.pane`) |
| Status     | string    | No change     | No change                      |
| UpdatedAt  | time.Time | No change     | No change                      |
| NotifiedAt | time.Time | No change     | No change                      |
| LastReadAt | int64     | No change     | No change                      |

## Storage Changes

### Mailbox Files

| Aspect   | Before                              | After                                       |
|----------|-------------------------------------|---------------------------------------------|
| Path     | `.agentmail/mailboxes/<window>.jsonl` | `.agentmail/mailboxes/<sanitized-address>.jsonl` |
| Example  | `.agentmail/mailboxes/editor.jsonl`  | `.agentmail/mailboxes/AgentMail%3Aeditor%2E0.jsonl` |
| Key      | Window name                         | Percent-encoded full pane address           |

**Sanitization**: Percent-encode structural characters: `%` → `%25`, `:` → `%3A`, `.` (pane separator) → `%2E`. This encoding is collision-safe (injective) and reversible. `ListMailboxRecipients()` returns canonical pane addresses by decoding filenames — all public APIs use canonical addresses, not raw filenames.

### Recipients File

| Aspect   | Before                              | After                                       |
|----------|-------------------------------------|---------------------------------------------|
| Path     | `.agentmail/recipients.jsonl`       | No change                                   |
| Key      | Window name in `recipient` field    | Full pane address in `recipient` field      |
| Example  | `{"recipient":"editor",...}`        | `{"recipient":"AgentMail:editor.0",...}`    |

### Ignore List

| Aspect   | Before                              | After                                       |
|----------|-------------------------------------|---------------------------------------------|
| Path     | `.agentmailignore`                  | No change                                   |
| Format   | One window name per line            | One address per line (full, medium `:window.pane`, or short window name) |
| Matching | Exact window name match             | Match against full pane address; short names match all panes of that window; medium `:window.pane` matches that specific pane in the current session |

## Address Resolution Flow

```
Input Address          Resolution Steps                           Output
─────────────────────────────────────────────────────────────────────────
"editor"               1. No ":" → short form                    mysession:editor.0
(single pane)          2. List panes for window "editor"           (if 1 pane)
                       3. Single pane → resolve to full address

"editor"               1. No ":" → short form                    ERROR: ambiguous
(multi pane)           2. List panes for window "editor"
                       3. Multiple panes → ambiguity error

"my.app"               1. No ":" → short form                    mysession:my.app.0
(single pane)          2. Window name = "my.app" (dots are literal)  (if 1 pane)
                       3. Single pane → resolve to full address

"logs.1"               1. No ":" → short form                    mysession:logs.1.0
(single pane)          2. Window name = "logs.1" (NOT medium form)  (if 1 pane)
                       3. Unambiguous: no colon prefix

":editor.1"            1. Starts with ":" → medium form          mysession:editor.1
                       2. Strip ":", split on last "."
                       3. Pane = 1, prepend current session

"mysession:editor.1"   1. Has ":" (not at pos 0) → full form     mysession:editor.1
                       2. Parse session, window, pane
                       3. Validate pane exists
```

## Tmux Query Commands

| Purpose                  | Command                                                                     |
|--------------------------|-----------------------------------------------------------------------------|
| Get current pane address | `tmux display-message -t $TMUX_PANE -p '#{session_name}:#{window_name}.#{pane_index}'` |
| List panes in session    | `tmux list-panes -s -F '#{session_name}:#{window_name}.#{pane_index}'`     |
| List all panes           | `tmux list-panes -a -F '#{session_name}:#{window_name}.#{pane_index}'`     |
| Check pane exists        | `tmux display-message -t '<address>' -p '#{pane_id}'` (exit code 0 = exists) |
| Get current session      | `tmux display-message -t $TMUX_PANE -p '#{session_name}'`                  |

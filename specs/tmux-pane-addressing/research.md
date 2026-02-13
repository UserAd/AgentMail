# Research: Tmux Pane Addressing

**Feature Branch**: `tmux-pane-addressing`
**Date**: 2026-02-13

## R1: Tmux Pane Query Commands

**Decision**: Use `tmux display-message -t $TMUX_PANE -p '#{session_name}:#{window_name}.#{pane_index}'` to get the current agent's full pane address.

**Rationale**: Live testing confirms this command returns the exact `session:window.pane` format needed (e.g., `AgentMail:main.0`). The `$TMUX_PANE` variable (e.g., `%0`) is immutable per process lifetime and already used by AgentMail for `GetCurrentWindow()`. Extending the format string to include session and pane index is a minimal change.

**Alternatives considered**:
- Querying session, window, and pane separately with three commands — rejected (3x exec overhead, race condition risk between queries)
- Using `tmux list-panes` to find current pane — rejected (requires filtering, more complex than direct query)

## R2: Listing All Panes (for Recipients and Validation)

**Decision**: Use `tmux list-panes -s -F '#{session_name}:#{window_name}.#{pane_index}'` for listing panes in the current session, and `tmux list-panes -a -F '...'` for cross-session listing.

**Rationale**: Live testing confirms these commands return one `session:window.pane` entry per line. The `-s` flag lists all panes across all windows in the current session. The `-a` flag lists all panes across all sessions.

**Alternatives considered**:
- Iterating `list-windows` then `list-panes` per window — rejected (N+1 exec calls, slower)
- Using window count to detect multi-pane windows — rejected (less direct than listing panes)

## R3: Pane Existence Validation

**Decision**: Use `tmux display-message -t 'session:window.pane' -p '#{pane_id}'` to check if a pane exists. Non-zero exit code means the pane does not exist.

**Rationale**: This leverages tmux's own target resolution. If the pane doesn't exist, tmux returns an error (exit code 1). This is simpler and more reliable than listing all panes and searching.

**Alternatives considered**:
- Listing all panes and checking membership — rejected (more expensive, especially for cross-session)
- Using `tmux has-session` — rejected (only checks session, not pane)

## R4: Address Parsing Strategy

**Decision**: Implement a single `ParseAddress` function that handles three forms:
1. Full: `session:window.pane` → contains `:` with non-empty session prefix, used as-is
2. Medium: `:window.pane` → starts with `:`, prepend current session
3. Short: `window` → no `:` prefix, resolve to full address (single-pane) or error (multi-pane)

**Rationale**: The address format mirrors tmux's native `-t` target syntax. The leading `:` for medium form matches tmux's own shorthand (`:window.pane` means "current session"). This eliminates the ambiguity with dot-containing window names identified in code review.

**Parsing rules**:
- Starts with `:` → medium format (strip leading `:`, split on last `.` to extract integer pane index)
- Contains `:` (not at position 0) → full format (split on first `:`, then split remainder on last `.`)
- No `:` → short format (entire string is window name, even if it contains dots)

**Edge case resolution**: Window names like `my.app` or `logs.1` are unambiguously short form because they lack a `:` prefix. To target pane 1 of `logs`, use `:logs.1` (medium) or `mysession:logs.1` (full). This was changed from the original design (which used `window.pane` without `:`) after code review identified the collision between medium-form parsing and dotted window names.

**Alternatives considered**:
- Using `window.pane` without colon prefix for medium form — rejected (ambiguous with dotted window names like `logs.1`)
- Using a regex — rejected (harder to maintain, less clear than structured parsing)
- Requiring full addresses always — rejected (breaks backward compatibility)

## R5: Filename Sanitization

**Decision**: Use percent-encoding for structural characters in mailbox filenames: `%` → `%25`, `:` → `%3A`, `.` (pane separator only, i.e., the last `.` before the pane index) → `%2E`.

**Rationale**: The original design (`:`→`_`, `.`→`-`) was identified in code review as non-injective: session/window names containing `_` or `-` could produce collisions (e.g., `s_a:w.0` and `s:a_w.0` both map to `s_a_w-0.jsonl`). Percent-encoding is collision-safe by definition since the escape character `%` is itself escaped first.

**Mapping**: `mysession:editor.0` → `mysession%3Aeditor%2E0.jsonl`

**Reversibility**: Fully reversible by decoding `%XX` sequences. `ListMailboxRecipients()` MUST return canonical pane addresses (by decoding filenames), not raw sanitized filenames, to prevent double-encoding and ensure a single type boundary: all public APIs use canonical addresses, percent-encoding is internal to the storage layer only.

**Alternatives considered**:
- Simple character replacement (`:`→`_`, `.`→`-`) — rejected after review (non-injective, causes mailbox aliasing/cross-delivery)
- Hash-based filenames — rejected (not human-readable, requires index file)

## R6: Backward Compatibility and Migration

**Decision**: No migration of existing mailbox files. Old window-name-based files are orphaned. New messages exclusively use pane-based filenames.

**Rationale**: Aligns with YAGNI principle (Constitution II). Migration adds complexity for a transition that happens once. Old files can be cleaned up with `agentmail cleanup`. The `cleanup` command already handles removing empty and orphaned mailboxes.

**Alternatives considered**:
- Auto-migration on first run — rejected (complex, error-prone for multi-pane windows)
- Dual-read (check both old and new paths) — rejected (ongoing complexity for temporary benefit)

## R7: tmux Control Mode (-CC) Impact

**Decision**: No special handling needed for tmux control mode.

**Rationale**: Control mode clients are standard tmux clients. The same `display-message`, `list-panes`, and target syntax work identically. `$TMUX_PANE` is set for control mode sessions. AgentMail's approach of querying tmux via `os/exec` is transparent to client mode.

## R8: Discovery Scope

**Decision**: Recipient discovery (`ListPanes()`, `agentmail recipients`, MCP `list-recipients`) is scoped to the current tmux session only. Cross-session send is supported when the user provides a full `session:window.pane` address.

**Rationale**: The current codebase uses `tmux list-windows` (current session only). Extending to `list-panes -s` (current session) is the natural equivalent. Cross-session discovery via `list-panes -a` would expose panes from unrelated projects, creating noise and potential security concerns. Cross-session send is still possible with explicit addressing, which is the right trade-off: you must know the target to send cross-session.

**Alternatives considered**:
- Full cross-session discovery via `ListAllPanes()` — rejected (noisy, security concern, breaks current session-scoped model)
- `ListAllPanes()` is still implemented for potential future use but not exposed in any command or MCP tool

## R9: Breaking Change — No Storage or MCP Backward Compatibility

**Decision**: This is a breaking change for storage format and MCP response schema. No data migration, no MCP backward-compatible fields, no upgrade procedure. The MCP `list-recipients` response replaces `window` with `address`. Old mailbox files are ignored. Note: CLI addressing backward compatibility IS maintained — short-form window names (FR-003) continue to work for single-pane windows.

**Rationale**: AgentMail is pre-1.0 infrastructure used by agents, not end users. Breaking changes to storage and wire format are acceptable when they maintain code simplicity (Constitution II — YAGNI). Adding dual-read logic or deprecated fields adds ongoing maintenance burden for a one-time transition. CLI backward compat for short window names is preserved because it's zero-cost (handled by the parser) and prevents breaking existing scripts.

## R11: No New External Dependencies

**Decision**: All changes use Go standard library only. No new dependencies required.

**Rationale**: The feature extends existing `os/exec`-based tmux queries with different format strings and adds string parsing logic (including `net/url` or manual percent-encoding). Both are well-served by the standard library. Constitution IV compliance maintained.

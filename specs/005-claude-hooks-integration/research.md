# Research: Claude Code Hooks Integration

**Feature**: 005-claude-hooks-integration
**Date**: 2026-01-12

## Research Questions

### RQ-001: How should the `--hook` flag be parsed?

**Decision**: Parse `--hook` flag in `main.go` before calling `Receive()`, pass as `HookMode bool` in `ReceiveOptions`

**Rationale**:
- Consistent with existing pattern where `ReceiveOptions` struct controls behavior
- Flag parsing in `main.go` matches existing command routing pattern
- Simple boolean flag avoids complexity of flag package for single option

**Alternatives Considered**:
1. Use Go `flag` package - Rejected: Overkill for single boolean flag, would change argument parsing style
2. Parse flag inside `Receive()` function - Rejected: Violates separation of concerns (CLI parsing vs business logic)

### RQ-002: What exit code should indicate "message available"?

**Decision**: Exit code 2

**Rationale**:
- Exit code 0 = success/no action (standard)
- Exit code 1 = error (already used by agentmail)
- Exit code 2 = currently used for "not in tmux" but spec changes this for hook mode
- In hook mode, exit code 2 means "notification available" - distinct from error

**Alternatives Considered**:
1. Exit code 1 for notification - Rejected: Confuses error with notification state
2. Custom exit code (e.g., 42) - Rejected: Non-standard, harder to remember

### RQ-003: How to handle errors silently in hook mode?

**Decision**: Wrap error conditions in hook mode check, return 0 instead of error message + exit code 1

**Rationale**:
- Hooks should be non-disruptive; errors should not interrupt Claude Code workflow
- Silent failure aligns with spec assumption: "Silent failure (exit 0) is preferable to noisy errors for hook integrations"
- Existing `Receive()` function already returns different exit codes for different conditions

**Implementation Pattern**:
```go
if opts.HookMode {
    // Silent exit on any error
    return 0
}
// Normal error handling
fmt.Fprintf(stderr, "error: %v\n", err)
return 1
```

### RQ-004: What should the STDERR output format be?

**Decision**: "You got new mail\n" prefix followed by standard message format

**Rationale**:
- Clear notification header for Claude Code to display
- Reuse existing message formatting (From, ID, content)
- Single newline between header and message for readability

**Output Format**:
```
You got new mail
From: sender-window
ID: ABC12345

Message content here
```

### RQ-005: How to test hook mode behavior?

**Decision**: Extend existing table-driven tests in `receive_test.go` with hook mode variants

**Rationale**:
- Existing test structure uses `ReceiveOptions` for mocking
- Hook mode is just another option in the struct
- Table-driven tests allow comprehensive coverage of all scenarios

**Test Cases Required** (mapped to requirements):

| Test Case | Expected Behavior | Requirements |
|-----------|-------------------|--------------|
| Hook mode + messages exist | STDERR output with "You got new mail\n" + message, exit 2, message marked read | FR-001a, FR-001b, FR-001c |
| Hook mode + no messages | No output, exit 0 | FR-002 |
| Hook mode + not in tmux | No output, exit 0 | FR-003 |
| Hook mode + file read error | No output, exit 0 | FR-004a |
| Hook mode + lock timeout | No output, exit 0 | FR-004b |
| Hook mode + corrupted mailbox | No output, exit 0 | FR-004c |
| Hook mode output stream | All output to STDERR (not STDOUT) | FR-005 |
| Non-hook mode unchanged | Regression: normal behavior preserved | N/A |

## Technology Best Practices

### Go CLI Flag Parsing

For a single boolean flag with no value, manual parsing is idiomatic:
```go
hookMode := len(os.Args) >= 3 && os.Args[2] == "--hook"
```

This avoids:
- Dependency on `flag` package complexity
- Need to handle positional vs flag argument ordering
- Subcommand-specific flag scopes

### STDERR vs STDOUT in CLI Tools

Best practice for hooks/notifications:
- STDOUT: Primary output (data, results)
- STDERR: Diagnostic output (errors, notifications, progress)

Hook mode uses STDERR because:
- Claude Code captures STDERR for display
- STDOUT remains clean for potential piping
- Matches Unix convention for notification output

## Dependencies

### New Dependency: github.com/peterbourgon/ff/v3

**Justification** (per Constitution Principle IV - Standard Library Preference):

**Why standard library is insufficient:**
- Go's `flag` package doesn't support subcommands natively
- Manual argument parsing with positional flags (e.g., `--hook` as `os.Args[2]`) is fragile
- Adding `-r`/`--recipient` and `-m`/`--message` to send command requires robust flag parsing
- Mixing positional arguments with flags requires careful handling that's error-prone manually

**Security/maintenance implications:**
- ff/v3 is maintained by Peter Bourgon (well-known Go community member)
- Minimal dependency tree (no transitive dependencies)
- Stable API (v3.4.0, released 2023)
- Used by production Go tools

**Alternatives considered:**
1. **Manual parsing** - Rejected: Fragile for `--hook` position, complex for `-r`/`-m` flags
2. **spf13/cobra** - Rejected: Heavy dependency, overkill for simple CLI
3. **urfave/cli** - Rejected: Larger API surface than needed
4. **Standard `flag` package** - Rejected: No subcommand support, no short flag aliases

**Decision:** ff/v3 provides the minimal necessary functionality (subcommands + flag parsing) with:
- Small binary size impact (~1.2 MB)
- Zero transitive dependencies
- Flag-first philosophy aligned with CLI-first constitution principle

### Existing Dependencies (unchanged):
- `os` - for exit codes, args
- `fmt` - for output formatting
- `io` - for Writer interface

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Exit code 2 meaning changes in hook mode | Low | Document clearly; non-hook mode unchanged |
| Silent failures hide real issues | Low | Only in hook mode; normal mode shows errors |
| Flag parsing edge cases | Low | Simple `--hook` check; no complex parsing |

## Conclusion

This feature is straightforward to implement using existing patterns:
1. Add `HookMode bool` to `ReceiveOptions`
2. Modify `Receive()` to check `HookMode` at each exit point
3. Write output to STDERR when in hook mode
4. Add test cases for all hook mode scenarios
5. Update README with Claude Code hooks section

No external research or dependencies required. Implementation can proceed directly to Phase 1 design artifacts.

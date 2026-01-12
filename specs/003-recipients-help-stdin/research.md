# Research: Recipients Command, Help Flag, and Stdin Message Input

**Feature**: 003-recipients-help-stdin
**Date**: 2026-01-12

## Research Topics

### 1. Stdin Detection in Go

**Decision**: Use `os.Stdin.Stat()` to check if stdin is a pipe/redirect

**Rationale**: The Go standard library provides `os.File.Stat()` which returns file mode information. When stdin is a terminal (interactive), the mode includes `os.ModeCharDevice`. When stdin is a pipe or redirect, this flag is absent.

**Implementation Pattern**:
```go
func isStdinPipe() bool {
    stat, err := os.Stdin.Stat()
    if err != nil {
        return false
    }
    return (stat.Mode() & os.ModeCharDevice) == 0
}
```

**Alternatives Considered**:
- `term.IsTerminal(int(os.Stdin.Fd()))` from golang.org/x/term - Rejected: external dependency violates constitution
- Timeout-based read - Rejected: unreliable, poor UX

---

### 2. Git Repository Root Detection

**Decision**: Walk up directory tree looking for `.git/` directory

**Rationale**: The `.agentmailignore` file must be in the git repository root (per clarification). Standard approach is to walk up from current directory until `.git/` is found.

**Implementation Pattern**:
```go
func findGitRoot() (string, error) {
    dir, err := os.Getwd()
    if err != nil {
        return "", err
    }
    for {
        if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
            return dir, nil
        }
        parent := filepath.Dir(dir)
        if parent == dir {
            return "", errors.New("not in a git repository")
        }
        dir = parent
    }
}
```

**Alternatives Considered**:
- `git rev-parse --show-toplevel` via exec - Rejected: slower, adds process overhead
- Environment variable - Rejected: unreliable, user must set manually

---

### 3. Ignore File Parsing

**Decision**: Simple line-by-line parsing with whitespace trimming

**Rationale**: Per spec FR-004/FR-005, the ignore file is a simple list of window names, one per line. Empty and whitespace-only lines are ignored.

**Implementation Pattern**:
```go
func parseIgnoreFile(path string) (map[string]bool, error) {
    file, err := os.Open(path)
    if err != nil {
        if os.IsNotExist(err) || os.IsPermission(err) {
            return nil, nil // Per FR-016: treat as if file doesn't exist
        }
        return nil, err
    }
    defer file.Close()

    ignored := make(map[string]bool)
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line != "" {
            ignored[line] = true
        }
    }
    return ignored, scanner.Err()
}
```

**Alternatives Considered**:
- Glob/regex patterns - Rejected: spec requires exact matching (per assumptions)
- JSON format - Rejected: over-engineered for simple list

---

### 4. Help Text Organization

**Decision**: Centralized help text in `internal/cli/help.go` with structured format

**Rationale**: Per FR-011/FR-012, help must show all commands with descriptions and syntax examples. A single source of truth ensures consistency.

**Implementation Pattern**:
```go
const helpText = `agentmail - Inter-agent communication for tmux sessions

USAGE:
    agentmail <command> [arguments]
    agentmail --help

COMMANDS:
    send <recipient> [message]    Send a message to a tmux window
                                  Message can be piped via stdin
    receive                       Read the oldest unread message
    recipients                    List available message recipients

EXAMPLES:
    agentmail send agent2 "Hello"
    echo "Hello" | agentmail send agent2
    agentmail receive
    agentmail recipients
`
```

**Alternatives Considered**:
- Per-command help files - Rejected: over-engineered for 3 commands
- Flag package auto-generation - Rejected: less control over format

---

### 5. Stdin Precedence Behavior

**Decision**: Stdin takes precedence when both stdin and argument are provided (per FR-010)

**Rationale**: This allows maximum flexibility. Users can pipe content and the argument serves as documentation or fallback.

**Implementation Pattern**:
```go
func getMessageContent(args []string, stdin io.Reader) (string, error) {
    if isStdinPipe() {
        content, err := io.ReadAll(stdin)
        if err != nil {
            return "", err
        }
        if len(content) > 0 {
            return strings.TrimSuffix(string(content), "\n"), nil
        }
    }
    // Fall back to argument
    if len(args) >= 2 {
        return args[1], nil
    }
    return "", errors.New("no message provided")
}
```

**Alternatives Considered**:
- Argument takes precedence - Rejected: less useful for scripting
- Error on both provided - Rejected: reduces usability

---

## Dependencies Audit

| Package | Source | Purpose | Constitution Compliance |
|---------|--------|---------|------------------------|
| os | stdlib | File operations, stdin stat | PASS |
| bufio | stdlib | Line-by-line file reading | PASS |
| strings | stdlib | String manipulation | PASS |
| io | stdlib | Reader interface for stdin | PASS |
| path/filepath | stdlib | Path manipulation | PASS |

**Result**: All dependencies are Go standard library. Constitution Principle IV satisfied.

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Stdin detection fails on exotic terminals | Low | Low | Fall back to argument if stat fails |
| Git root not found | Low | Medium | Clear error message, graceful degradation |
| Large stdin input | Low | Low | No artificial limit per clarification |

---

## Open Questions

None. All technical decisions resolved using standard library approaches.

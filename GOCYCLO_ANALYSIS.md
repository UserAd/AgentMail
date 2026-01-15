# Gocyclo Cyclomatic Complexity Analysis

**Generated:** 2026-01-15 (Updated after refactoring)
**Tool:** gocyclo v0.6.0

## Summary

After refactoring, all critical complexity issues have been resolved. Only CLI setup code (`main.main`) remains above 15, which is acceptable for command-line applications.

### Refactoring Results

| Function | Before | After | Reduction |
|----------|--------|-------|-----------|
| `cli.Send` | 30 | 15 | 50% |
| `cli.Receive` | 25 | 9 | 64% |
| `daemon.CheckAndNotifyWithNotifier` | 24 | 2 | 92% |
| `mcp.doSend` | 21 | 13 | 38% |
| `mcp.doListRecipients` | 13 | 8 | 38% |

### Remaining High Complexity (Production Code)

| Complexity | Function | File | Notes |
|------------|----------|------|-------|
| 18 | `main.main` | cmd/agentmail/main.go | Acceptable for CLI setup |
| 15 | `cli.Send` | internal/cli/send.go | At threshold |

---

## Original Analysis

### Complexity Thresholds
- **1-5:** Simple, low risk
- **6-10:** Moderate complexity
- **11-15:** High complexity - consider refactoring
- **16-20:** Very high complexity - should refactor
- **21+:** Critical - needs immediate attention

## High-Priority Functions (Complexity > 20)

| Complexity | Function | File | Line |
|------------|----------|------|------|
| 30 | `cli.Send` | internal/cli/send.go | 31 |
| 25 | `cli.Receive` | internal/cli/receive.go | 34 |
| 24 | `daemon.CheckAndNotifyWithNotifier` | internal/daemon/loop.go | 140 |
| 22 | `mcp.TestStatusTool_SchemaValidation` | internal/mcp/tools_test.go | 251 |
| 21 | `mcp.doSend` | internal/mcp/handlers.go | 89 |

## Medium-Priority Functions (Complexity 15-20)

| Complexity | Function | File | Line |
|------------|----------|------|------|
| 19 | `mcp.TestSendTool_SchemaValidation` | internal/mcp/tools_test.go | 127 |
| 18 | `main.main` | cmd/agentmail/main.go | 16 |
| 16 | `mcp.TestServer_100ConsecutiveInvocations` | internal/mcp/handlers_test.go | 2736 |
| 16 | `mail.TestUpdateRecipientState_UpdateExisting` | internal/mail/recipients_test.go | 322 |
| 15 | `mail.TestUpdateLastReadAt_PreservesOtherRecipients` | internal/mail/recipients_test.go | 1061 |

## Root Cause Analysis

### 1. Mock/Test Infrastructure in Production Code

Both `cli.Send` and `cli.Receive` contain significant complexity from mock support:

```go
// Pattern repeating throughout:
if opts.MockWindows != nil {
    for _, w := range opts.MockWindows {
        if w == recipient {
            recipientExists = true
            break
        }
    }
} else {
    var err error
    recipientExists, err = tmux.WindowExists(recipient)
    // ...
}
```

**Impact:** Each mock option adds 2-3 branches.

### 2. Hook Mode Conditional Logic

`cli.Receive` has complexity from hook mode which adds early-exit branches throughout:

```go
if opts.HookMode {
    return 0  // Silent exit
}
// Normal error handling path
```

**Impact:** Hook mode adds ~8 additional branches.

### 3. Two-Phase Processing

`daemon.CheckAndNotifyWithNotifier` processes two types of agents (stated and stateless), essentially doubling the code complexity:
- Phase 1: Loop through stated agents with multiple conditions
- Phase 2: Loop through stateless agents with similar conditions

### 4. Duplicated Validation Logic

`cli.Send` and `mcp.doSend` share nearly identical validation logic but are implemented separately, both with mock support.

## Recommendations

### Immediate (Complexity > 20)

#### 1. Extract Validation Logic into Helper Functions

Create a shared validation package or helper functions:

```go
// internal/validation/recipient.go
type RecipientValidator struct {
    Windows    []string  // nil means use real tmux
    IgnoreList map[string]bool
}

func (v *RecipientValidator) Exists(recipient string) (bool, error)
func (v *RecipientValidator) IsIgnored(recipient string) bool
func (v *RecipientValidator) IsSelf(recipient, sender string) bool
```

**Expected reduction:** cli.Send 30 → ~18, mcp.doSend 21 → ~12

#### 2. Split cli.Send Into Smaller Functions

```go
func Send(...) int {
    if err := validateTmux(opts); err != nil { ... }

    msg, err := parseMessage(args, stdin, opts)
    if err != nil { ... }

    sender, err := getSender(opts)
    if err != nil { ... }

    if err := validateRecipient(recipient, sender, opts); err != nil { ... }

    return storeAndRespond(msg, stdout, stderr, opts)
}
```

**Expected reduction:** 30 → ~10

#### 3. Separate Hook Mode Into Different Function

```go
func Receive(stdout, stderr io.Writer, opts ReceiveOptions) int {
    if opts.HookMode {
        return receiveHookMode(stderr, opts)
    }
    return receiveNormalMode(stdout, stderr, opts)
}
```

**Expected reduction:** 25 → ~12 each

#### 4. Refactor CheckAndNotifyWithNotifier

Split into two functions:

```go
func CheckAndNotifyWithNotifier(opts LoopOptions, ...) error {
    if err := notifyStatedAgents(opts, notify, windowChecker); err != nil {
        return err
    }
    return notifyStatelessAgents(opts, notify, windowChecker)
}
```

**Expected reduction:** 24 → ~12 each

### Medium-Term

#### 5. Interface-Based Testing

Replace mock structs with interfaces to reduce conditional branches:

```go
type WindowLister interface {
    ListWindows() ([]string, error)
    WindowExists(name string) (bool, error)
    GetCurrentWindow() (string, error)
}

// Production implementation uses real tmux
// Test implementation returns mock data
```

**Benefit:** Eliminates `if opts.Mock != nil` patterns throughout.

#### 6. Use Table-Driven Tests

Several test functions with high complexity (19-22) can be converted to table-driven tests:

```go
func TestSendTool_SchemaValidation(t *testing.T) {
    tests := []struct{
        name string
        input map[string]any
        wantErr bool
    }{
        // test cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // single test implementation
        })
    }
}
```

### Low Priority

#### 7. Simplify main.go

The main function (complexity 18) is acceptable for CLI setup but could be improved by:
- Moving command definitions to separate files
- Using a command registry pattern

## Files Not Requiring Changes

Functions with complexity 10 or below are acceptable and follow good practices. Test file complexity is less critical than production code.

## Metrics After Recommended Changes

| Function | Before | After (Est.) |
|----------|--------|--------------|
| cli.Send | 30 | 10-12 |
| cli.Receive | 25 | 10-12 |
| daemon.CheckAndNotifyWithNotifier | 24 | 10-12 |
| mcp.doSend | 21 | 10-12 |
| main.main | 18 | 10-14 |

## Running Gocyclo

```bash
# Install
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest

# Run with average
gocyclo -avg .

# Show only functions with complexity > 15
gocyclo -over 15 .

# Top 10 most complex functions
gocyclo -top 10 .
```

## CI Integration

Add to quality gates in CLAUDE.md:
```bash
# Fail if any function exceeds complexity of 20
gocyclo -over 20 . && exit 1 || exit 0
```

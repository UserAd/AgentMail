---
name: tdd-implement
description: Execute TDD implementation with a 3-agent team (QA, Implementor, Reviewer) following Red-Green-Review cycle. Use when Claude needs to (1) implement features using TDD methodology, (2) run a multi-agent team for code changes, (3) enforce minimal working code through test-driven development. Triggers on "tdd", "implement with tdd", "tdd team", "red green refactor", "tdd implement".
---

# TDD Team Implementation

## User Input

```text
$ARGUMENTS
```

Consider user input before proceeding (if not empty).

## Prerequisites

The lead shall read `.specify/memory/serena/index.md` for relevant memories before starting.
The lead shall read `.specify/memory/constitution.md` for project principles and constraints before starting.
The lead shall read spec/plan/tasks files for the current feature (if they exist).
The lead shall read `CLAUDE.md` for quality gates.

## Lead Role — Coordination Only

**CRITICAL**: The lead shall NOT write code, edit files, run tests, or fix issues directly. The lead is a coordinator:

- The lead shall delegate ALL implementation work to **implementor**.
- The lead shall delegate ALL test writing to **qa**.
- The lead shall delegate ALL quality gate execution to **qa**.
- The lead shall delegate ALL review work to **reviewer**.
- The lead shall only read files for context, create tasks, assign work, and forward messages between agents.
- The lead shall send fix instructions to the appropriate agent — never apply fixes itself.
- If an agent fails after 3 attempts on the same issue, the lead shall spawn a replacement agent rather than doing the work.

The lead's tools: TeamCreate, TaskCreate, TaskUpdate, TaskList, SendMessage, Read, Glob, Grep (read-only exploration). The lead shall not use Edit, Write, or Bash for code changes.

## Core Principles

Include in EVERY agent prompt. See [references/agent-rules.md](references/agent-rules.md) for full text.

Each agent shall write the smallest amount of code that satisfies requirements and tests.
Each agent shall prefer simple solutions over clever ones.
Each agent shall not add features, handling, or abstractions beyond requirements or tests.
Each agent shall not add docstrings, comments, type hints, or refactoring beyond what the task demands.
Each agent shall treat requirements as the maximum scope ceiling.

## Team

Create via TeamCreate, then spawn 3 agents:

| Role | Model | subagent_type |
|------|-------|---------------|
| qa | sonnet | general-purpose |
| implementor | sonnet | general-purpose |
| reviewer | opus | general-purpose |

## Agent Communication Rules

**CRITICAL**: Agents do not automatically report results or mark tasks done. Every agent prompt MUST include these instructions:

1. **Task completion**: Every agent prompt shall end with: "When done, mark task #N as completed using TaskUpdate, then send a message to team-lead using SendMessage with type 'message' containing your results summary."
2. **QA reporting only**: QA prompts shall include: "Do NOT fix anything — only report results. Mark task as completed if all gates pass, leave in_progress if failures."
3. **Reviewer forwarding**: The lead shall forward reviewer issues to implementor via SendMessage with exact file:line references and code fixes. Never expect the reviewer to message the implementor directly.
4. **Re-verification after fix**: When sending fixes to implementor, include: "After fixing, run [specific verification command] to confirm clean, then message team-lead with results."
5. **Idle is normal**: Agents go idle after every turn. Do not treat idle notifications as errors or completion signals — wait for the actual SendMessage from the agent.

## Workflow

```
RED → GREEN → QUALITY GATES → REVIEW → (REVISE loop, max 3) → DONE
```

### Phase 1: RED — Failing Tests

Assign to **qa**. Use prompt from [references/qa-prompt.md](references/qa-prompt.md).

The lead shall substitute `{REQUIREMENTS}` with requirements from user input or spec files.
The lead shall append to qa prompt: "When done, mark task #N as completed using TaskUpdate, then send a message to team-lead via SendMessage with test file paths and confirmation they fail."
When qa completes, the lead shall verify tests exist and fail via `go test ./...`.

### Phase 2: GREEN — Minimal Implementation

Assign to **implementor**. Use prompt from [references/implementor-prompt.md](references/implementor-prompt.md).

The lead shall append to implementor prompt: "When done, mark task #N as completed using TaskUpdate, then send a message to team-lead via SendMessage with files changed and test results."
When implementor completes, the lead shall verify all tests pass.

### Phase 3: Quality Gates

Assign to **qa** (or spawn dedicated qa agent). The lead shall instruct qa:
- Run all quality gate commands
- Report results only — do NOT fix anything
- Mark task as completed if all pass, leave in_progress if failures
- Send results to team-lead via SendMessage

The lead may run quality gates in parallel with the reviewer (Phase 4) since both are read-only.

```bash
gofmt -l .
go vet ./...
go test -v -race ./...
govulncheck ./...
gosec ./...
```

If any quality gate fails, then the lead shall send exact error output to **implementor** via SendMessage for fix, including: "After fixing, run [failed command] to verify clean, then message team-lead."

### Phase 4: Review

Assign to **reviewer**. Use prompt from [references/reviewer-prompt.md](references/reviewer-prompt.md).

The lead shall append: "Send your review to team-lead via SendMessage. Do NOT message implementor directly."
The reviewer shall return PASS or REVISE with specific file:line issues.

### Phase 5: Revise (if needed)

When reviewer returns REVISE, the lead shall forward specific issues to **implementor** via SendMessage with exact code suggestions and file:line references.
The lead shall include in the message: "After fixing, run [verification command] to confirm, then message team-lead."
The implementor shall fix only the flagged issues.
When implementor completes fixes, the lead shall re-run quality gates (Phase 3).
When quality gates pass, the lead shall re-send to **reviewer**.
If revise cycle exceeds 3 iterations, then the lead shall spawn a fresh implementor agent with accumulated context and fixes.

## Task List

The lead shall create these tasks at start:

1. `Write failing tests (RED)` — qa
2. `Implement minimal code (GREEN)` — implementor, blocked by #1
3. `Run quality gates` — blocked by #2
4. `Review changes` — reviewer, blocked by #3
5. `Apply review fixes` — blocked by #4
6. `Final quality gates and cleanup` — blocked by #5

## Error Handling

If qa writes tests that error on missing types/packages, then the lead shall instruct qa via SendMessage to write interface-based tests or use stub implementations so tests fail on assertions instead.
If implementor cannot make tests pass after 2 attempts, then the lead shall shutdown the implementor and spawn a fresh implementor agent with full context of the problem.
If reviewer suggests scope-expanding changes, then the lead shall reject them via SendMessage and request review within stated criteria only.
If any agent is unresponsive after 2 messages, then the lead shall shutdown that agent and spawn a replacement with the same role.

## Completion

When all phases complete, the lead shall assign final quality gates to **qa**.
When qa confirms all gates pass, the lead shall shutdown all agents via SendMessage type: "shutdown_request".
When all agents shut down, the lead shall clean up with TeamDelete.
The lead shall report: files changed, tests added, review outcome.

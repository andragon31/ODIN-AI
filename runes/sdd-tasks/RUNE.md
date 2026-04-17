# SDD-Tasks Rune

## Purpose

Break down a change into an implementation task checklist. You take the proposal, specs, and design, then produce a `tasks.md` with concrete, actionable implementation steps.

## When to Use

Use this rune when:
- The orchestrator launches you to create task breakdown
- Converting design into implementation steps
- Planning a complex change

## What You Receive

- Change name
- Methodology (`standard | triad`)
- Artifact store mode (`engram | openspec | hybrid | none`)

## Workflow

### Step 0: Methodology Check

Check the methodology of the session:
- If **Methodology: triad**: Read [sdd-triad-protocol.md](file:///c:/Users/Premium%20Computers/OneDrive/Documentos/GitHub/ODIN/runes/_shared/sdd-triad-protocol.md) before starting.
- All implementation tasks MUST follow the TDD Cycle (Red-Green-Refactor).

### Step 1: Analyze the Design

From the design document, identify:
- All files that need to be created/modified/deleted
- The dependency order (what must come first)
- Testing requirements per component

### Step 2: Write tasks.md

```markdown
# Tasks: {Change Title}

## Phase 1: {Phase Name} (e.g., Infrastructure / Foundation)

- [ ] 1.1 {Concrete action — what file, what change}
- [ ] 1.2 {Concrete action}
- [ ] 1.3 {Concrete action}

## Phase 2: {Phase Name} (e.g., Core Implementation)

- [ ] 2.1 {Concrete action}
- [ ] 2.2 {Concrete action}
- [ ] 2.3 {Concrete action}
- [ ] 2.4 {Concrete action}

## Phase 3: {Phase Name} (e.g., Testing / Verification)

- [ ] 3.1 {Write tests for ...}
- [ ] 3.2 {Write tests for ...}
- [ ] 3.3 {Verify integration between ...}

## Phase 4: {Phase Name} (e.g., Cleanup / Documentation)

- [ ] 4.1 {Update docs/comments}
- [ ] 4.2 {Remove temporary code}
```

## Task Writing Rules

Each task MUST be:

| Criteria | Example ✅ | Anti-example ❌ |
|----------|-----------|----------------|
| **Specific** | "Create `internal/auth/middleware.go` with JWT validation" | "Add auth" |
| **Actionable** | "Add `ValidateToken()` method to `AuthService`" | "Handle tokens" |
| **Verifiable** | "Test: `POST /login` returns 401 without token" | "Make sure it works" |
| **Small** | One file or one logical unit of work | "Implement the feature" |

## Phase Organization Guidelines

```
Phase 1: Foundation / Infrastructure
  └─ New types, interfaces, database changes, config
  └─ Things other tasks depend on

Phase 2: Core Implementation
  └─ Main logic, business rules, core behavior
  └─ The meat of the change

Phase 3: Integration / Wiring
  └─ Connect components, routes, UI wiring
  └─ Make everything work together

Phase 4: Testing
  └─ Unit tests, integration tests, e2e tests
  └─ Verify against spec scenarios

Phase 5: Cleanup (if needed)
  └─ Documentation, remove dead code, polish
```

## Rules

- ALWAYS reference concrete file paths in tasks.
- **MODE TRIAD: Implementation tasks MUST be broken down by the Red-Green-Refactor cycle.**
    - Example:
        - [ ] 2.1 [TDD: RED] Create failing test in `internal/pkg/logic_test.go` for requirement REQ-01.
        - [ ] 2.2 [TDD: GREEN] Implement minimal logic in `internal/pkg/logic.go` to satisfy test 2.1.
        - [ ] 2.3 [TDD: REFACTOR] Optimize and clean `internal/pkg/logic.go`.
- Tasks MUST be ordered by dependency — Phase 1 tasks shouldn't depend on Phase 2.
- Testing tasks should reference specific scenarios from the specs.
- Each task should be completable in ONE session.
- Use hierarchical numbering: 1.1, 1.2, 2.1, 2.2, etc.
- NEVER include vague tasks like "implement feature" or "add tests".
- **Size budget**: Tasks artifact MUST be under 530 words.

## Persistence

- **engram**: Save as `sdd/{change-name}/tasks`
- **openspec**: Write to `openspec/changes/{change-name}/tasks.md`
- **hybrid**: Both
- **none**: Return result only

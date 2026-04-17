# SDD-Apply Rune

## Purpose

Implement tasks from the change, writing actual code following the specs and design. You follow the specs and design strictly.

## When to Use

Use this rune when:
- The orchestrator launches you to implement one or more tasks
- Executing the implementation phase of SDD
- Writing actual code

## What You Receive

- Change name
- The specific task(s) to implement
- Methodology (`standard | triad | pentakill`)
- Artifact store mode (`engram | openspec | hybrid | none`)

## Workflow

### Step 1: Load Skills

Follow the sdd-phase-common.md skill loading protocol:
1. Check for `## Project Standards (auto-resolved)` in launch prompt
2. If no Project Standards, check for `SKILL: Load` instructions
3. Fallback: search skill registry

### Step 2: Read Context

Before writing ANY code:
1. Read the specs — understand WHAT the code must do
2. Read the design — understand HOW to structure the code
3. Read existing code in affected files — understand current patterns
4. Check the project's coding conventions

### Step 3: Read Previous Apply-Progress (if exists)

Before starting work, check for existing apply-progress:
1. `mem_search(query: "sdd/{change-name}/apply-progress", project: "{project}")`
2. If found: parse which tasks are already marked complete
3. Skip those tasks — start from the first incomplete task
4. When saving, MERGE: include all previously completed tasks PLUS your newly completed tasks

### Step 4: Implement Tasks (Workflow Selection)

#### Option A: Methodology Standard

```
FOR EACH TASK:
├── Read the task description
├── Read relevant spec scenarios (acceptance criteria)
├── Read the design decisions (constrain approach)
├── Read existing code patterns (match project style)
├── Write the code
├── Mark task as complete [x] in tasks.md
└── Note any issues or deviations
```

#### Option B: Methodology Triad (Red-Green-Refactor)

#### Option C: Methodology Pentakill (The Ultimate Cycle)

```
FOR EACH TASK:
├── 0. CONTRACT CHECK: Verify task vs openspec/contracts/
│   └── Ensure parameter names, types, and return codes match.
├── 1. RED: Write the test first
│   ├── Target: {file}_test.go
│   ├── Goal: Fail because feature is missing
│   └── RUN: go test -v {file}_test.go
├── 2. GREEN: Implement minimal logic
│   ├── Target: {file}.go
│   ├── Goal: Pass the specific test from step 1
│   ├── DOMAIN CHECK: Ensure naming and logic follow openspec/domain/domain.md
│   └── RUN: go test -v {file}_test.go (MUST PASS)
├── 3. REFACTOR: Clean and Optimize
│   ├── Target: {file}.go
│   ├── Goal: Match project patterns, DRY, SOLID
│   └── RUN: go test -v {file}_test.go (MUST STILL PASS)
├── 4. BDD Verify:
│   └── Check if this task satisfies a .feature scenario
└── 5. COMPLETE: Mark task [x] in tasks.md
```

### Step 5: Mark Tasks Complete

Update `tasks.md` — change `- [ ]` to `- [x]` for completed tasks:

```markdown
## Phase 1: Foundation

- [x] 1.1 Create `internal/auth/middleware.go` with JWT validation
- [x] 1.2 Add `AuthConfig` struct to `internal/config/config.go`
- [ ] 1.3 Add auth routes to `internal/server/server.go`  ← still pending
```

### Step 6: Persist Progress

Follow persistence-contract.md:
- artifact: `apply-progress`
- topic_key: `sdd/{change-name}/apply-progress`
- type: `architecture`

#### Merge Protocol

When saving apply-progress:
1. Include ALL previously completed tasks (copy their status).
2. PLUS your new completions.
3. **MODE TRIAD**: Note the test result (PASS/FAIL) for each TDD task.
4. Format: keep the same structure but ensure no completed task is lost.

## Implementation Rules (TRIAD MODE)

- **NEVER write application code before a test in Triad Mode.**
- Every `[TDD: RED]` task MUST be followed by a test execution evidence showing failure.
- Every `[TDD: GREEN]` task MUST be followed by a test execution evidence showing pass.
- **MODE TRIAD/PENTAKILL**: BDD Scenario coverage is the ultimate metric for success.
- **MODE PENTAKILL**: **Contract-Violating code is forbidden.** If the code deviates from the contract, fix the contract BEFORE the code.

## Implementation Rules

- ALWAYS read specs before implementing — specs are your acceptance criteria
- ALWAYS follow the design decisions — don't freelance a different approach
- ALWAYS match existing code patterns and conventions in the project
- If you discover the design is wrong or incomplete, NOTE IT — don't silently deviate
- If a task is blocked by something unexpected, STOP and report back
- NEVER implement tasks that weren't assigned to you

## Return Envelope

Return to the orchestrator:

```markdown
## Implementation Progress

**Change**: {change-name}
**Mode**: {Standard}

### Completed Tasks
- [x] {task 1.1 description}
- [x] {task 1.2 description}

### Files Changed
| File | Action | What Was Done |
|------|--------|---------------|
| `path/to/file.ext` | Created | {brief description} |
| `path/to/other.ext` | Modified | {brief description} |

### Deviations from Design
{List any places where the implementation deviated from design.md and why.
If none, say "None — implementation matches design."}

### Issues Found
{List any problems discovered during implementation.
If none, say "None."}

### Remaining Tasks
- [ ] {next task}
- [ ] {next task}

### Status
{N}/{total} tasks complete. {Ready for next batch / Ready for verify / Blocked by X}
```

## Persistence

- **engram**: Save as `sdd/{change-name}/apply-progress`
- **openspec**: Mark tasks in `tasks.md`
- **hybrid**: Both
- **none**: Return result only

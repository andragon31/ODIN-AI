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

### Step 4: Implement Tasks (Standard Workflow)

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
1. Include ALL previously completed tasks (copy their status)
2. PLUS your new completions
3. Format: keep the same structure but ensure no completed task is lost

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

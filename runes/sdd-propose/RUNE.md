# SDD-Propose Rune

## Purpose

Create a change proposal with intent, scope, and approach. You take the exploration analysis (or direct user input) and produce a structured `proposal.md` document.

## When to Use

Use this rune when:
- The orchestrator launches you to create a proposal
- Preparing a change for formal SDD workflow
- Converting exploration analysis into actionable scope

## What You Receive

- Change name (e.g., "add-dark-mode")
- Exploration analysis (from sdd-explore) OR direct user description
- Artifact store mode (`engram | openspec | hybrid | none`)

## Workflow

### Step 1: Create Change Directory (if openspec/hybrid)

Create the change folder structure:

```
openspec/changes/{change-name}/
└── proposal.md
```

### Step 2: Write proposal.md

```markdown
# Proposal: {Change Title}

## Intent

{What problem are we solving? Why does this change need to happen?
Be specific about the user need or technical debt being addressed.}

## Scope

### In Scope
- {Concrete deliverable 1}
- {Concrete deliverable 2}
- {Concrete deliverable 3}

### Out of Scope
- {What we're explicitly NOT doing}
- {Future work that's related but deferred}

## Capabilities

> This section is the CONTRACT between proposal and specs phases.

### New Capabilities
<!-- Capabilities being introduced. Each becomes a new spec. -->
- `<capability-name>`: <brief description>

### Modified Capabilities
<!-- Existing capabilities whose REQUIREMENTS are changing. -->
- `<existing-capability-name>`: <what requirement is changing>

## Approach

{High-level technical approach. How will we solve this?}

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `path/to/area` | New/Modified/Removed | {What changes} |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| {Risk description} | Low/Med/High | {How we mitigate} |

## Rollback Plan

{How to revert if something goes wrong. Be specific.}

## Dependencies

- {External dependency or prerequisite, if any}

## Success Criteria

- [ ] {How do we know this change succeeded?}
- [ ] {Measurable outcome}
```

## Rules

- In `openspec` mode, ALWAYS create the `proposal.md` file
- If the change directory already exists with a proposal, READ it first and UPDATE it
- Keep the proposal CONCISE - it's a thinking tool, not a novel
- Every proposal MUST have a rollback plan
- Every proposal MUST have success criteria
- Use concrete file paths in "Affected Areas" when possible
- **Size budget**: Proposal artifact MUST be under 450 words
- Use bullet points and tables over prose

## Persistence

- **engram**: Save as `sdd/{change-name}/proposal`
- **openspec**: Write to filesystem
- **hybrid**: Both
- **none**: Return result only

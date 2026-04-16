# SDD-Archive Rune

## Purpose

Sync delta specs to main specs and archive a completed change. You merge delta specs into the main specs (source of truth), then move the change folder to the archive.

## When to Use

Use this rune when:
- The orchestrator launches you to archive after verification
- Completing the SDD cycle for a change
- Merging specs and cleaning up

## What You Receive

- Change name
- Artifact store mode (`engram | openspec | hybrid | none`)

## Workflow

### Step 1: Load Skills

Follow the sdd-phase-common.md skill loading protocol.

### Step 2: Sync Delta Specs to Main Specs

**IF mode is `openspec` or `hybrid`:**

For each delta spec in `openspec/changes/{change-name}/specs/`:

#### If Main Spec Exists (`openspec/specs/{domain}/spec.md`)

Read the existing main spec and apply the delta:

```
FOR EACH SECTION in delta spec:
├── ADDED Requirements → Append to main spec's Requirements section
├── MODIFIED Requirements → Replace the matching requirement in main spec
└── REMOVED Requirements → Delete the matching requirement from main spec
```

**Merge carefully:**
- Match requirements by name
- Preserve all OTHER requirements that aren't in the delta
- Maintain proper Markdown formatting

#### If Main Spec Does NOT Exist

The delta spec IS a full spec. Copy it directly:

```bash
openspec/changes/{change-name}/specs/{domain}/spec.md
  → openspec/specs/{domain}/spec.md
```

### Step 3: Move to Archive

Move the entire change folder to archive with date prefix:

```
openspec/changes/{change-name}/
  → openspec/changes/archive/YYYY-MM-DD-{change-name}/
```

Use today's date in ISO format.

### Step 4: Verify Archive

**IF mode is `openspec` or `hybrid`:**

Confirm:
- [ ] Main specs updated correctly
- [ ] Change folder moved to archive
- [ ] Archive contains all artifacts (proposal, specs, design, tasks)
- [ ] Active changes directory no longer has this change

**IF mode is `engram`:**

Confirm all artifact observation IDs are recorded in the archive report.

### Step 5: Persist Archive Report

Follow persistence-contract.md:
- artifact: `archive-report`
- topic_key: `sdd/{change-name}/archive-report`
- type: `architecture`

## Archive Report Format

```markdown
## Change Archived

**Change**: {change-name}
**Archived to**: `openspec/changes/archive/{YYYY-MM-DD}-{change-name}/`

### Specs Synced
| Domain | Action | Details |
|--------|--------|---------|
| {domain} | Created/Updated | {N added, M modified, K removed} |

### Archive Contents
- proposal.md ✅
- specs/ ✅
- design.md ✅
- tasks.md ✅ ({N}/{N} tasks complete)

### Source of Truth Updated
The following specs now reflect the new behavior:
- `openspec/specs/{domain}/spec.md`

### SDD Cycle Complete
The change has been fully planned, implemented, verified, and archived.
Ready for the next change.
```

## Rules

- NEVER archive a change that has CRITICAL issues in its verification report
- ALWAYS sync delta specs BEFORE moving to archive
- When merging into existing specs, PRESERVE requirements not mentioned in the delta
- Use ISO date format (YYYY-MM-DD) for archive folder prefix
- If the merge would be destructive, WARN the orchestrator and ask for confirmation
- The archive is an AUDIT TRAIL — never delete or modify archived changes
- If `openspec/changes/archive/` doesn't exist, create it

## Persistence

- **engram**: Save as `sdd/{change-name}/archive-report`
- **openspec**: Move folder, write report
- **hybrid**: Both
- **none**: Return result only

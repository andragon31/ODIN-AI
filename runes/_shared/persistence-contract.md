# Persistence Contract (shared across all SDD phases)

## Mode Resolution

The orchestrator passes `artifact_store.mode` with one of: `engram | openspec | hybrid | none`.

Default (if user doesn't specify): if Engram is available → `engram`. Otherwise → `none`.

## Mode Roles

- **`engram`**: Working memory between sessions. Upserts overwrite — no iteration history. Local only.
- **`openspec`**: Source of truth. Files in repo, git history, team-shareable, full audit trail.
- **`hybrid`**: Both — files for team + engram for recovery.
- **`none`**: Ephemeral. Lost when conversation ends.

## Behavior Per Mode

| Mode | Read from | Write to | Project files |
|------|-----------|----------|---------------|
| `engram` | Engram | Engram | Never |
| `openspec` | Filesystem | Filesystem | Yes |
| `hybrid` | Engram (primary) + Filesystem (fallback) | Both | Yes |
| `none` | Orchestrator prompt context | Nowhere | Never |

## Common Rules

- `none` → do NOT create or modify any project files; return results inline only
- `engram` → do NOT write any project files; persist to Engram
- `openspec` → write files ONLY to paths defined in `openspec-convention.md`
- `hybrid` → persist to BOTH Engram AND filesystem
- NEVER force `openspec/` creation unless orchestrator explicitly passed `openspec` or `hybrid`

## Sub-Agent Context Rules

Who reads, who writes:
- Non-SDD (general task): orchestrator searches engram, passes summary in prompt; sub-agent saves discoveries via `mem_save`
- SDD (phase with dependencies): sub-agent reads artifacts directly from backend; sub-agent saves its artifact
- SDD (phase without dependencies, e.g. explore): nobody reads; sub-agent saves its artifact

## Skill Registry

The orchestrator pre-resolves compact rules from the skill registry and injects them as `## Project Standards (auto-resolved)` in the launch prompt.
# Engram Artifact Convention

NOTE: Critical engram calls (`mem_search`, `mem_save`, `mem_get_observation`) are inlined directly in each skill's RUNE.md. This document is supplementary reference.

## Naming Rules

ALL SDD artifacts persisted to Engram MUST follow this deterministic naming:

```
title:     sdd/{change-name}/{artifact-type}
topic_key: sdd/{change-name}/{artifact-type}
type:      architecture
project:   {detected or current project name}
scope:     project
```

### Artifact Types

| Artifact Type | Produced By | Description |
|---------------|-------------|-------------|
| `explore` | sdd-explore | Exploration analysis |
| `proposal` | sdd-propose | Change proposal |
| `spec` | sdd-spec | Delta specifications |
| `design` | sdd-design | Technical design |
| `tasks` | sdd-tasks | Task breakdown |
| `apply-progress` | sdd-apply | Implementation progress |
| `verify-report` | sdd-verify | Verification report |
| `archive-report` | sdd-archive | Archive closure with lineage |

### Recovery Protocol

```
Step 1: mem_search(query: "sdd/{change-name}/{artifact-type}", project: "{project}") → truncated preview + ID
Step 2: mem_get_observation(id: {observation-id}) → complete content
```

## Upsert Behavior

Same `topic_key` + `project` + `scope` → UPDATE (overwrite), not INSERT. Previous content is lost on re-run.
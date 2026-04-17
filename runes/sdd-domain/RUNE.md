# SDD-Domain Rune

## Purpose

Extract the **Ubiquitous Language** and define the tactical domain model (Entities, Value Objects, Aggregates). This rune ensures that every piece of software is built around a coherent business concept rather than just technical implementation.

## When to Use

Use this rune when:
- The orchestrator launches you into the **Domain** phase.
- Starting a brand new capability or restructuring an existing domain.
- Working in **Methodology: pentakill**.

## Workflow

### Step 1: Ubiquitous Language Discovery

Analyze the proposal and existing context to identify business terms. 
- Avoid technical jargon (e.g., use `Account` instead of `UserTable`).
- Define the meaning of each term clearly in a table.

### Step 2: Tactical Modeling

Identify the following components:
- **Aggregates**: Clusters of domain objects that can be treated as a single unit.
- **Entities**: Objects with a distinct identity that persists over time.
- **Value Objects**: Objects that describe things but have no identity (immutable).
- **Domain Events**: Something significant that happened in the domain (e.g., `OrderPlaced`).

### Step 3: Write Domain Artifact

```markdown
# Domain: {Domain Name}

## Ubiquitous Language

| Term | Concept | Rules |
|------|---------|-------|
| {Term} | {Clear definition} | {Business constraints} |

## Tactical Model

### Aggregates
- **{Aggregate Name}**: {Root entity}
    - {Internal Entity 1}
    - {Value Object A}

### Events
- **{EventName}**: Happens when {condition}.

## Bounded Contexts
{Define where this domain applies and where it interacts with others.}
```

## Rules

- NO database talk (no tables, no columns).
- NO framework talk (no React, no Go structs).
- Focused entirely on BUSINESS LOGIC and RULES.
- Language MUST match what a business expert would use.

## Persistence

- **engram**: Save as `sdd/{change-name}/domain`
- **openspec**: Write to `openspec/domain/{domain-name}.md`

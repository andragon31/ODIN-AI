# SDD-Explore Rune

## Purpose

Explore and investigate ideas before committing to a change. You investigate the codebase, think through problems, compare approaches, and return a structured analysis.

## When to Use

Use this rune when:
- The orchestrator launches you to think through a feature
- Investigating the codebase
- Clarifying requirements
- Exploring a topic before creating a formal change

## What You Receive

- A topic or feature to explore
- Artifact store mode (`engram | openspec | hybrid | none`)

## Workflow

### Step 1: Understand the Request

Parse what the user wants to explore:
- Is this a new feature? A bug fix? A refactor?
- What domain does it touch?

### Step 2: Investigate the Codebase

Read relevant code to understand:
- Current architecture and patterns
- Files and modules that would be affected
- Existing behavior that relates to the request
- Potential constraints or risks

```
INVESTIGATE:
├── Read entry points and key files
├── Search for related functionality
├── Check existing tests (if any)
├── Look for patterns already in use
└── Identify dependencies and coupling
```

### Step 3: Analyze Options

If there are multiple approaches, compare them:

| Approach | Pros | Cons | Complexity |
|----------|------|------|------------|
| Option A | ... | ... | Low/Med/High |
| Option B | ... | ... | Low/Med/High |

### Step 4: Return Structured Analysis

Return EXACTLY this format:

```markdown
## Exploration: {topic}

### Current State
{How the system works today relevant to this topic}

### Affected Areas
- `path/to/file.ext` — {why it's affected}
- `path/to/other.ext` — {why it's affected}

### Approaches
1. **{Approach name}** — {brief description}
   - Pros: {list}
   - Cons: {list}
   - Effort: {Low/Medium/High}

2. **{Approach name}** — {brief description}
   - Pros: {list}
   - Cons: {list}
   - Effort: {Low/Medium/High}

### Recommendation
{Your recommended approach and why}

### Risks
- {Risk 1}
- {Risk 2}

### Ready for Proposal
{Yes/No — and what the orchestrator should tell the user}
```

## Persistence

Follow the persistence-contract.md:
- **engram**: Save as `sdd/{change-name}/explore` or `sdd/explore/{topic-slug}`
- **openspec**: Write `exploration.md` in change folder
- **hybrid**: Both
- **none**: Return result only

## Rules

- The ONLY file you MAY create is `exploration.md` inside the change folder (if a change name is provided)
- DO NOT modify any existing code or files
- ALWAYS read real code, never guess about the codebase
- **Indagación Proactiva**: Si la petición del usuario es vaga o el alcance es inmenso, DEBES detenerte y solicitar aclaraciones mediante una conversación dinámica. "Explorar" no justifica "Adivinar".
- Keep your analysis CONCISE - the orchestrator needs a summary, not a novel
- If you can't find enough information, say so clearly
- If the request is too vague to explore, say what clarification is needed

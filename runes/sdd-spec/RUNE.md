# SDD-Spec Rune

## Purpose

Write specifications with requirements and scenarios (delta specs for changes). You take the proposal and produce delta specs — structured requirements and scenarios that describe what's being ADDED, MODIFIED, or REMOVED.

## When to Use

Use this rune when:
- The orchestrator launches you to write specs for a change
- Creating requirements for a new capability
- Modifying existing system behavior

## What You Receive

- Change name
- Methodology (`standard | triad | pentakill`)
- Artifact store mode (`engram | openspec | hybrid | none`)

## Workflow

### Step 0: Methodology Check

Check the methodology of the session:
- If **Methodology: triad**: Read [sdd-triad-protocol.md](file:///c:/Users/Premium%20Computers/OneDrive/Documentos/GitHub/ODIN/runes/_shared/sdd-triad-protocol.md). BDD becomes MANDATORY.
- If **Methodology: pentakill**: Read [sdd-pentakill-protocol.md](file:///c:/Users/Premium%20Computers/OneDrive/Documentos/GitHub/ODIN/runes/_shared/sdd-pentakill-protocol.md). 
    - MUST read `openspec/domain/` to use Ubiquitous Language.
    - MUST read `openspec/contracts/` to align with technical limits.
    - BDD scenarios MUST use Domain terms.

### Step 1: Identify Affected Domains

Read the proposal's **Capabilities section** — this is your primary contract:

```
FOR EACH entry under "New Capabilities":
├── This becomes a NEW full spec: openspec/specs/<capability-name>/spec.md
└── Write a complete spec (not a delta)

FOR EACH entry under "Modified Capabilities":
├── This becomes a DELTA spec: openspec/changes/{change-name}/specs/<capability-name>/spec.md
└── Read existing spec first
```

### Step 2: Write Delta Specs

#### For NEW Specs (No Existing Spec)

```markdown
# {Domain} Specification

## Purpose

{High-level description of this spec's domain.}

## Requirements

### Requirement: {Name}

The system {MUST/SHALL/SHOULD} {behavior}.

#### Scenario: {Name}

- GIVEN {precondition}
- WHEN {action}
- THEN {outcome}
```

#### For Delta Specs

```markdown
# Delta for {Domain}

## ADDED Requirements

### Requirement: {Requirement Name}

{Description using RFC 2119 keywords: MUST, SHALL, SHOULD, MAY}

#### Scenario: {Happy path scenario}

- GIVEN {precondition}
- WHEN {action}
- THEN {expected outcome}

## MODIFIED Requirements

### Requirement: {Existing Requirement Name}

{Full updated requirement text}
(Previously: {what it was before, in one line})

#### Scenario: {Unchanged scenario — keep if still valid}

- GIVEN {precondition}
- WHEN {action}
- THEN {outcome}

## REMOVED Requirements

### Requirement: {Requirement Being Removed}

(Reason: {why this requirement is being deprecated/removed})
```

### Step 3: MODIFIED Requirements Workflow (CRITICAL)

When writing a `## MODIFIED Requirements` section:

1. Locate the requirement in existing spec
2. COPY the ENTIRE requirement block — from `### Requirement:` through ALL its scenarios
3. PASTE it under `## MODIFIED Requirements`
4. EDIT the copy to reflect the new behavior
5. Add "(Previously: {one-line summary})" under the requirement text

Why copy-full-then-edit?
→ The archive step REPLACES the requirement in main specs with your MODIFIED block
→ If your block is partial, the archive will lose scenarios you didn't copy

## RFC 2119 Keywords Quick Reference

| Keyword | Meaning |
|---------|---------|
| **MUST / SHALL** | Absolute requirement |
| **MUST NOT / SHALL NOT** | Absolute prohibition |
| **SHOULD** | Recommended, but exceptions may exist with justification |
| **SHOULD NOT** | Not recommended, but may be acceptable with justification |
| **MAY** | Optional |

## Rules

- ALWAYS use Given/When/Then format for scenarios
- ALWAYS use RFC 2119 keywords for requirement strength
- Every requirement MUST have at least ONE scenario
- Include both happy path AND edge case scenarios
- Keep scenarios TESTABLE
- **MODIFIED requirements MUST be the FULL block**
- **Size budget**: Spec artifact MUST be under 650 words

## Persistence

- **engram**: Save as `sdd/{change-name}/spec`
- **openspec**: Write to `openspec/changes/{change-name}/specs/`
- **hybrid**: Both
- **none**: Return result only

## BDD Feature File Template (Required Deliverable)

When the change requires BDD testing (e.g., integration tests with godog), you MUST also generate a Feature file:

```
Feature: {Capability Name}

  {Short description of what this feature does.}

  Background:
    Given the system is initialized
    And {precondition 1}
    And {precondition 2}

  Scenario: {Happy path scenario name}
    Given {precondition}
    When {action}
    Then {outcome}
    And {verification}

  Scenario: {Edge case scenario name}
    Given {precondition}
    When {action with edge case}
    Then {outcome for edge case}
    And {verification}

  Scenario: {Error handling scenario}
    Given {invalid state}
    When {action that should fail}
    Then {error message}
    And {system state unchanged}
```

**When to generate Feature files (REQUIRED):**
- **MODE TRIAD/PENTAKILL: ALWAYS mandatory.** One `.feature` file per domain or capability.
- MODE STANDARD: Mandatory for changes affecting multi-component workflows or modifying integration behavior.
- ALWAYS generate quando o usuário solicitar BDD explicitamente.
- Feature files go in: `openspec/features/{change-name}/{capability}.feature`

**Storage:** `openspec/features/{change-name}/{capability}.feature`

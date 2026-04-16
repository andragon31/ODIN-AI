# SDD-Verify Rune

## Purpose

Validate that implementation matches specs, design, and tasks. You are the quality gate — prove with real execution evidence that the implementation is complete, correct, and behaviorally compliant.

## When to Use

Use this rune when:
- The orchestrator launches you to verify a completed change
- Running the quality gate before archive
- Validating partial implementation

## What You Receive

- Change name
- Artifact store mode (`engram | openspec | hybrid | none`)

## Workflow

### Step 1: Load Skills

Follow the sdd-phase-common.md skill loading protocol.

### Step 2: Read Testing Capabilities

Read cached testing capabilities to determine TDD mode:
- From engram: `mem_search("sdd/{project}/testing-capabilities")`
- From openspec: `openspec/config.yaml`
- Fallback: check project files directly

### Step 3: Check Completeness

Verify ALL tasks are done:

```
Read tasks.md
├── Count total tasks
├── Count completed tasks [x]
├── List incomplete tasks [ ]
└── Flag: CRITICAL if core tasks incomplete
```

### Step 4: Check Correctness (Static Specs Match)

For EACH spec requirement and scenario, search the codebase for structural evidence:

```
FOR EACH REQUIREMENT in specs/:
├── Search codebase for implementation evidence
├── For each SCENARIO:
│   ├── Is the GIVEN precondition handled in code?
│   ├── Is the WHEN action implemented?
│   ├── Is the THEN outcome produced?
│   └── Are edge cases covered?
└── Flag: CRITICAL if requirement missing
```

### Step 5: Check Coherence (Design Match)

Verify design decisions were followed:

```
FOR EACH DECISION in design.md:
├── Was the chosen approach actually used?
├── Were rejected alternatives accidentally implemented?
├── Do file changes match the "File Changes" table?
└── Flag: WARNING if deviation found
```

### Step 6: Run Tests (Real Execution)

```
Detect test runner from:
├── Cached testing capabilities
├── openspec/config.yaml → rules.verify.test_command
├── package.json, pyproject.toml, Makefile
└── Fallback: ask orchestrator

Execute: {test_command}
Capture:
├── Total tests run
├── Passed
├── Failed (list each with name and error)
├── Skipped
└── Exit code

Flag: CRITICAL if exit code != 0
```

### Step 7: Build & Type Check

```
Execute: {build_command}
Capture:
├── Exit code
├── Errors (if any)

Flag: CRITICAL if build fails
```

### Step 8: Spec Compliance Matrix (Behavioral Validation)

Cross-reference EVERY spec scenario against actual test run results:

```
FOR EACH REQUIREMENT in specs/:
  FOR EACH SCENARIO:
  ├── Find tests that cover this scenario
  ├── Look up test result
  ├── Assign compliance status:
  │   ├── ✅ COMPLIANT   → test exists AND passed
  │   ├── ❌ FAILING     → test exists BUT failed
  │   ├── ❌ UNTESTED    → no test found
  │   └── ⚠️ PARTIAL    → test exists, covers only part
  └── Record: requirement, scenario, test, result
```

A spec scenario is only COMPLIANT when there is a test that PASSED proving the behavior.

## Verification Report Format

```markdown
## Verification Report

**Change**: {change-name}
**Mode**: {Standard}

---

### Completeness
| Metric | Value |
|--------|-------|
| Tasks total | {N} |
| Tasks complete | {N} |
| Tasks incomplete | {N} |

---

### Build & Tests Execution
**Build**: ✅ Passed / ❌ Failed

**Tests**: ✅ {N} passed / ❌ {N} failed / ⚠️ {N} skipped

---

### Spec Compliance Matrix
| Requirement | Scenario | Test | Result |
|-------------|----------|------|--------|
| {REQ-01} | {Scenario} | `{test file} > {test name}` | ✅ COMPLIANT |

**Compliance summary**: {N}/{total} scenarios compliant

---

### Issues Found
**CRITICAL** (must fix before archive): {List or "None"}
**WARNING** (should fix): {List or "None"}

---

### Verdict
{PASS / PASS WITH WARNINGS / FAIL}
```

## Rules

- ALWAYS read the actual source code — don't trust summaries
- ALWAYS execute tests — static analysis alone is NOT verification
- A spec scenario is only COMPLIANT when a test that covers it has PASSED
- Compare against SPECS first (behavioral), DESIGN second (structural)
- Be objective — report what IS, not what should be
- CRITICAL issues = must fix before archive
- WARNINGS = should fix but won't block
- DO NOT fix any issues — only report them

## Persistence

- **engram**: Save as `sdd/{change-name}/verify-report`
- **openspec**: Write to `openspec/changes/{change-name}/verify-report.md`
- **hybrid**: Both
- **none**: Return result only

package runeforge

import (
	"fmt"
	"strings"

	"github.com/odin-ai/odin/internal/skills"
)

// buildPrompt builds a generation prompt for a rune
func (f *RuneForge) buildPrompt(req ForgeRequest) string {
	tags := ""
	if len(req.Tags) > 0 {
		tags = "\nTags: " + strings.Join(req.Tags, ", ")
	}

	description := req.Description
	if description == "" {
		description = "A useful skill for " + req.Name
	}

	// Use separate variables for code blocks to avoid backtick issues in raw strings
	codeBlockStart := "```yaml"
	codeBlockEnd := "```"

	prompt := fmt.Sprintf(`You are a skill designer for ODIN, a local-first AI ecosystem.

Generate a new Rune (skill) with the following specification:

## Rune Name
%s

## Description
%s
%s

## Requirements

Generate ONLY a valid YAML Rune skill with these exact fields:

%s
name: %s
version: "1.0.0"
description: <clear description of what this skill does>
author: "ODIN"
tags: [<comma-separated relevant tags>]
triggers:
  filePatterns: [<file patterns that activate this skill, e.g., "*.go", "*.md">]
  commands: [<CLI commands that activate this skill, e.g., "build", "test">]
  context: [<context keywords that activate this skill, e.g., "CI", "debug">]
execution:
  type: "prompt" | "script" | "wasm"  # choose the appropriate type
  prompt: <if type is prompt, include the prompt template>
  script: <if type is script, include the script>
  sandbox: true
outputs:
  files: [<expected output files>]
  console: <expected console output>
  errors: [<expected error conditions>]
%s

## Guidelines

1. **name**: Use lowercase with hyphens (e.g., "branch-pr", "issue-create")
2. **version**: Always use semantic versioning (e.g., "1.0.0")
3. **description**: Be concise but informative (1-2 sentences)
4. **tags**: Include 3-5 relevant tags for discoverability
5. **triggers**: Be specific - only include patterns/commands/context that truly activate this skill
6. **execution.type**: Choose "prompt" for LLM-based skills, "script" for shell/code execution
7. **execution.prompt**: Use {{.Variable}} for templating variables
8. **sandbox**: Always true unless the skill has security implications

## Example

%s
name: branch-pr
version: "1.0.0"
description: Creates feature branches and PRs following the issue-first workflow
author: "ODIN"
tags:
  - git
  - pr
  - workflow
  - github
triggers:
  filePatterns:
    - "*.go"
  commands:
    - pr
    - branch
  context:
    - github
execution:
  type: prompt
  prompt: |
    Create a feature branch following the issue-first workflow.
    Issue: {{.Issue}}
    Branch name: {{.BranchName}}
  sandbox: true
outputs:
  console: "Branch created: {{.BranchName}}\nPR opened: {{.PRURL}}"
%s

Generate the YAML for the requested skill now. Return ONLY the YAML, no explanations.`,
		req.Name, description, tags,
		codeBlockStart, req.Name,
		codeBlockEnd,
		codeBlockStart, codeBlockEnd)

	return prompt
}

// buildAdaptPrompt builds a prompt for adapting an existing rune
func (f *RuneForge) buildAdaptPrompt(exampleRune *skills.Rune, target string) string {
	codeBlockStart := "```yaml"
	codeBlockEnd := "```"

	return fmt.Sprintf(`You are a skill designer adapting runes for different platforms.

## Original Rune
Name: %s
Version: %s
Description: %s
Tags: %s
Triggers: %v
Execution Type: %s
Execution: %s %s
Outputs: %v

## Target Platform
%s

## Task
Adapt the original rune to work with the target platform.
Maintain the same functionality but adjust:
1. Commands and file patterns for the new platform
2. Execution script/prompt if needed
3. Tags to reflect the new context

## Requirements

Return ONLY valid YAML with these fields:

%s
name: <adapted name>
version: "1.0.0"
description: <clear description>
author: "ODIN"
tags:
  - <relevant tags>
triggers:
  filePatterns: [<file patterns>]
  commands: [<commands>]
  context: [<context>]
execution:
  type: "prompt" | "script" | "wasm"
  prompt: <prompt template if type is prompt>
  script: <script if type is script>
  sandbox: true
outputs:
  files: [<expected outputs>]
  console: <expected console output>
%s

Return ONLY the YAML, no explanations.`,
		exampleRune.Name, exampleRune.Version, exampleRune.Description,
		strings.Join(exampleRune.Tags, ", "),
		exampleRune.Triggers,
		exampleRune.Execution.Type,
		exampleRune.Execution.Prompt,
		exampleRune.Execution.Script,
		exampleRune.Outputs,
		target,
		codeBlockStart, codeBlockEnd)
}

// buildValidationPrompt builds a prompt for validating a rune
func (f *RuneForge) buildValidationPrompt(r *skills.Rune) string {
	return fmt.Sprintf(`Validate the following Rune skill:

name: %s
version: %s
description: %s
tags: %s
triggers:
  filePatterns: %v
  commands: %v
  context: %v
execution:
  type: %s
  prompt: %s
  script: %s
  sandbox: %v
outputs:
  files: %v
  console: %s
  errors: %v

Check for:
1. Required fields present (name, version, description, execution)
2. Valid name format (lowercase with hyphens)
3. Valid version format (semver)
4. Valid execution.type (prompt, script, or wasm)
5. Triggers are specific and useful
6. Outputs are realistic

Return JSON with validation result:
{"valid": true/false, "errors": ["list of errors"], "warnings": ["list of warnings"]}`, r.Name, r.Version, r.Description,
		strings.Join(r.Tags, ", "),
		r.Triggers.FilePatterns,
		r.Triggers.Commands,
		r.Triggers.Context,
		r.Execution.Type,
		r.Execution.Prompt,
		r.Execution.Script,
		r.Execution.Sandbox,
		r.Outputs.Files,
		r.Outputs.Console,
		r.Outputs.Errors)
}

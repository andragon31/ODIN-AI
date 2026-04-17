# security.rego - ODIN Rune Security Policies
# Package: odin.security.runes

package odin.security.runes

default allow := false

# Allow rune execution by default
allow {
    input.rune.execution.sandbox == true
}

# Deny script execution without sandbox
deny[msg] {
    input.rune.execution.type == "script"
    input.rune.execution.sandbox == false
    msg := "Script execution without sandbox is not allowed"
}

# Deny WASM execution without sandbox
deny[msg] {
    input.rune.execution.type == "wasm"
    input.rune.execution.sandbox == false
    msg := "WASM execution without sandbox is not allowed"
}

# Deny execution of unknown types
deny[msg] {
    input.rune.execution.type != "prompt"
    input.rune.execution.type != "script"
    input.rune.execution.type != "wasm"
    msg := "Unknown execution type"
}

# Check if rune has required fields
deny[msg] {
    input.rune.name == ""
    msg := "Rune must have a name"
}

deny[msg] {
    input.rune.version == ""
    msg := "Rune must have a version"
}

deny[msg] {
    input.rune.triggers.commands == []
    input.rune.triggers.filePatterns == []
    msg := "Rune must have at least one trigger (command or file pattern)"
}

# Validate version format (semver)
deny[msg] {
    input.rune.version != ""
    not valid_version(input.rune.version)
    msg := "Rune version must be valid semver"
}

valid_version(v) {
    parts := split(v, ".")
    count(parts) >= 2
}

# Rune execution audit
allow[msg] {
    input.rune.execution.type == "prompt"
    input.operation == "execute"
    msg := "Prompt-based execution is allowed"
}

# Validate triggers are not empty
deny[msg] {
    input.rune.triggers.commands == []
    input.rune.triggers.filePatterns == []
    msg := "Rune must define at least one trigger"
}

# Check for dangerous patterns in prompts
deny[msg] {
    input.rune.execution.type == "prompt"
    contains(input.rune.execution.prompt, "rm -rf")
    msg := "Prompt contains dangerous command pattern"
}

deny[msg] {
    input.rune.execution.type == "prompt"
    contains(input.rune.execution.prompt, "sudo")
    msg := "Prompt contains privilege escalation pattern"
}
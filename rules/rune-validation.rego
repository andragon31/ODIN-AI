# rune-validation.rego - Rune Content Validation Policies
# Package: odin.runes

package odin.runes

default valid_rune := false

# Valid rune must have required files
valid_rune {
    has_required_files
    has_valid_yaml
    has_required_fields
}

# Check for required RUNE.md file
has_required_files {
    input.files[_] == "RUNE.md"
}

# Check for rune.yaml validity
has_valid_yaml {
    input.yaml.name != ""
    input.yaml.version != ""
}

# Check for required fields in rune.yaml
has_required_fields {
    input.yaml.name != ""
    input.yaml.version != ""
    input.yaml.trigger != ""
}

# Validate trigger matches file location
deny["Trigger mismatch: rune location doesn't match trigger"] {
    input.yaml.trigger != ""
    not contains(input.path, replace(input.yaml.trigger, "-", "_"))
}

# Require description for rune
deny["Rune must have a description"] {
    input.yaml.description == ""
}

# Version must follow semver
deny["Invalid version format - must be semver"] {
    input.yaml.version != ""
    not matches_version(input.yaml.version)
}

matches_version(v) {
    count(trim_prefix(v, "v")) >= 2
}
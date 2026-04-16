package runeforge

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/odin-ai/odin/internal/skills"
)

// Parser parses model output into Rune structs
type Parser struct{}

// NewParser creates a new parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseRune parses model output into a Rune struct
func (p *Parser) ParseRune(content string) (*skills.Rune, error) {
	// Try to extract YAML from the content
	yamlContent := p.extractYAML(content)
	if yamlContent == "" {
		return nil, fmt.Errorf("no YAML found in model output")
	}

	// Parse YAML
	rune := &skills.Rune{}
	decoder := yaml.NewDecoder(strings.NewReader(yamlContent))
	decoder.KnownFields(true)

	if err := decoder.Decode(rune); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Set defaults
	if rune.Execution.Sandbox {
		rune.Execution.Sandbox = true
	}

	return rune, nil
}

// ParseRuneJSON parses JSON model output into a Rune
func (p *Parser) ParseRuneJSON(content string) (*skills.Rune, error) {
	// Try to extract JSON from content
	jsonContent := p.extractJSON(content)
	if jsonContent == "" {
		return nil, fmt.Errorf("no JSON found in model output")
	}

	rune := &skills.Rune{}
	if err := json.Unmarshal([]byte(jsonContent), rune); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return rune, nil
}

// ParseRuneFromMarkdown parses a rune from markdown code blocks
func (p *Parser) ParseRuneFromMarkdown(content string) (*skills.Rune, error) {
	// Look for yaml or json code blocks
	patterns := []struct {
		name     string
		startPat *regexp.Regexp
		endPat   *regexp.Regexp
		parser   func(string) (*skills.Rune, error)
	}{
		{
			name:     "yaml",
			startPat: regexp.MustCompile("(?i)```yaml\\s*$"),
			endPat:   regexp.MustCompile("\\s*```\\s*$"),
			parser:   p.ParseRune,
		},
		{
			name:     "json",
			startPat: regexp.MustCompile("(?i)```json\\s*$"),
			endPat:   regexp.MustCompile("\\s*```\\s*$"),
			parser:   p.ParseRuneJSON,
		},
	}

	for _, pattern := range patterns {
		lines := strings.Split(content, "\n")
		var yamlLines []string
		inBlock := false

		for _, line := range lines {
			if pattern.startPat.MatchString(line) {
				inBlock = true
				continue
			}
			if inBlock && pattern.endPat.MatchString(line) {
				break
			}
			if inBlock {
				yamlLines = append(yamlLines, line)
			}
		}

		if len(yamlLines) > 0 {
			extracted := strings.Join(yamlLines, "\n")
			rune, err := pattern.parser(extracted)
			if err == nil {
				return rune, nil
			}
		}
	}

	return nil, fmt.Errorf("no valid rune found in markdown")
}

// extractYAML extracts YAML content from model output
func (p *Parser) extractYAML(content string) string {
	// Look for YAML code blocks first
	yamlBlockPat := regexp.MustCompile("(?is)```yaml\\s*\\n(.+?)\\n```|```yaml\\s*\\n(.+?)\\n```|\\n---\\n(.+?)\\n---")
	matches := yamlBlockPat.FindStringSubmatch(content)
	if len(matches) > 0 {
		for i := 1; i < len(matches); i++ {
			if matches[i] != "" {
				return strings.TrimSpace(matches[i])
			}
		}
	}

	// Try to find YAML-like content (starts with name: or has common fields)
	lines := strings.Split(content, "\n")
	var yamlLines []string
	inYaml := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines at the start
		if !inYaml && trimmed == "" {
			continue
		}

		// Check if this looks like YAML
		if !inYaml {
			if strings.HasPrefix(trimmed, "name:") ||
				strings.HasPrefix(trimmed, "#") ||
				strings.HasPrefix(trimmed, "---") {
				inYaml = true
			} else {
				continue
			}
		}

		// End of YAML block
		if strings.HasPrefix(trimmed, "---") && len(yamlLines) > 0 {
			break
		}

		yamlLines = append(yamlLines, line)
	}

	if len(yamlLines) > 0 {
		return strings.Join(yamlLines, "\n")
	}

	return ""
}

// extractJSON extracts JSON content from model output
func (p *Parser) extractJSON(content string) string {
	// Look for JSON code blocks
	jsonBlockPat := regexp.MustCompile("(?is)```json\\s*\\n(.+?)\\n````")
	matches := jsonBlockPat.FindStringSubmatch(content)
	if len(matches) > 0 {
		return strings.TrimSpace(matches[1])
	}

	// Try to find raw JSON
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start != -1 && end != -1 && end > start {
		return content[start : end+1]
	}

	return ""
}

// ParsePartialRune parses a rune with partial/incomplete fields
func (p *Parser) ParsePartialRune(content string) (*skills.Rune, []string, error) {
	rune, err := p.ParseRune(content)
	if err != nil {
		return nil, []string{err.Error()}, err
	}

	warnings := []string{}

	// Check for common issues
	if rune.Name == "" {
		warnings = append(warnings, "name is empty")
	}
	if rune.Version == "" {
		rune.Version = "0.1.0" // Default version
		warnings = append(warnings, "version was empty, defaulting to 0.1.0")
	}
	if rune.Execution.Type == "" {
		warnings = append(warnings, "execution.type is empty, defaulting to prompt")
		rune.Execution.Type = "prompt"
	}

	return rune, warnings, nil
}

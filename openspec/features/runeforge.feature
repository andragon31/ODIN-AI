Feature: RuneForge - Rune Generation

  RuneForge generates new runes using ODIN's local Router
  (Ollama, OpenRouter, or Anthropic). It supports direct generation,
  generation from examples, and iterative refinement.

  Background:
    Given the RuneForge engine is initialized
    And a Router is configured with a mock provider

  Scenario: Generate rune from prompt
    When I generate a rune with description "authentication rune"
    Then a new rune should be created
    And the rune should have a valid name and description
    And the rune should pass validation

  Scenario: Generate rune from example
    Given an existing rune at path "runes/branch-pr"
    When I generate a new rune adapted from the example for "deployment"
    Then the new rune should be created
    And it should be based on the example structure
    And the description should mention the adaptation

  Scenario: Validate generated rune
    Given a rune with missing required fields
    When I validate the rune
    Then I should receive validation errors
    And the errors should list missing fields

  Scenario: Parse rune from markdown
    Given a markdown document containing a yaml code block
    When I parse the rune from the markdown
    Then I should receive a valid Rune object
    And all fields should be extracted correctly

  Scenario: Generate async with timeout
    When I generate a rune with a 5 second timeout
    Then the operation should complete within the timeout
    Or it should return a timeout error

  Scenario: Partial rune parsing with defaults
    Given a rune with only name and description
    When I parse the rune with partial fields
    Then the rune should have default values for missing fields
    And warnings should indicate which fields were defaulted
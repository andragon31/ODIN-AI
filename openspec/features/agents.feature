Feature: Multi-Agent Configuration

  ODIN supports multi-agent configurations where different AI agents
  (Cursor, Claude Code, etc.) can be installed and configured. The system
  should detect existing agents and provide seamless integration.

  Background:
    Given ODIN is initialized in a workspace
    And the agents registry is configured

  Scenario: Detect Cursor AI installation
    When the system checks for Cursor AI
    Then it should detect if Cursor is installed
    And it should identify the Cursor model in use
    And the model should be logged for routing decisions

  Scenario: Detect Claude Code installation
    When the system checks for Claude Code
    Then it should detect if Claude Code is installed
    And it should identify the Claude model being used

  Scenario: List configured agents
    When I request the list of configured agents
    Then I should see all detected agents
    And each agent should show its status (active/inactive)
    And the default agent should be marked

  Scenario: Set default agent preference
    Given multiple agents are installed
    When I set "cursor" as the default agent
    Then future routing decisions should prefer Cursor
    And the preference should be persisted

  Scenario: Agent fallback chain
    Given multiple agents are configured in fallback order
    When the primary agent fails
    Then the system should automatically try the next agent
    And the fallback should be transparent to the user

  Scenario: Agent-specific configuration
    Given an agent requires specific environment variables
    When the agent configuration is loaded
    Then the environment should be properly set up
    And the agent should receive its required config
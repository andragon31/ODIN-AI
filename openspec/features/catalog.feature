Feature: Catalog Management

  ODIN's catalog system manages available components, agents, and runes.
  Users should be able to browse, search, and install items from the catalog.

  Background:
    Given the catalog manager is initialized
    And the catalog contains sample components

  Scenario: List available components
    When I request the list of available components
    Then I should receive a list of component names
    And each component should have a name and description

  Scenario: Search components by tag
    Given the catalog has components with various tags
    When I search for components tagged with "security"
    Then I should receive only security-related components
    And no non-security components should be included

  Scenario: Get component details
    Given a component named "heimdall" exists in the catalog
    When I request details for the "heimdall" component
    Then I should receive the component's full description
    And the component's available runes should be listed

  Scenario: Detect installed agents
    When the system detects installed agents
    Then I should see a list of agent IDs
    And each agent should have a name and version

  Scenario: Install component from catalog
    Given the catalog shows component "mimir" is available
    When I install the "mimir" component
    Then the installation should complete successfully
    And the component should appear in the installed components list

  Scenario: Catalog update check
    Given the catalog is connected to the remote index
    When I check for catalog updates
    Then I should receive the current catalog version
    And any new components should be listed
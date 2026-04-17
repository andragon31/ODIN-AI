Feature: Pipeline Orchestration

  ODIN's pipeline orchestrates multi-stage installation with rollback
  support. Each stage (detect, backup, install, verify, commit) can
  fail gracefully with automatic rollback.

  Background:
    Given a pipeline is created for component "test-component"
    And system detection has been initialized

  Scenario: Execute full pipeline successfully
    When the pipeline runs all stages in order
    Then stage "detect" should complete successfully
    And stage "backup" should complete successfully
    And stage "install" should complete successfully
    And stage "verify" should complete successfully
    And stage "commit" should complete successfully
    And the pipeline should report success

  Scenario: Pipeline detects system environment
    When system detection runs
    Then the detected OS should be identified
    And the architecture should be detected
    And the user home directory should be found
    And installed agents should be listed

  Scenario: Pipeline creates backup before install
    Given the system has existing ODIN data at ~/.odin
    When the backup stage runs
    Then a backup archive should be created
    And the backup path should be stored for rollback

  Scenario: Pipeline rolls back on install failure
    Given the pipeline is running
    And the install stage fails
    When the pipeline performs rollback
    Then the component directory should be removed
    And the backup should be restored
    And an error should be returned

  Scenario: Pipeline handles interruption gracefully
    Given the pipeline is running
    When the user cancels the pipeline (Ctrl+C)
    Then the pipeline should stop gracefully
    And completed stages should be rolled back
    And an interruption error should be returned

  Scenario: Pipeline runs install stage
    Given a component exists in the catalog
    When the install stage executes
    Then the component directory should be created at ~/.odin/{component}
    And the component's runes should be installed
    And the stage should report success

  Scenario: Pipeline verifies installation
    Given a component has been installed
    When the verify stage runs
    Then the component directory should exist
    And all required files should be present
    And the verification should pass
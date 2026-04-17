Feature: Mimir Memory Engine

  Mimir is ODIN's memory engine providing persistent storage with
  semantic vector search, full-text search, and knowledge graph edges.
  It uses SQLite with WAL mode for reliability.

  Background:
    Given the memory store is initialized
    And the embedder is configured with Ollama

  Scenario: Store a new memory
    Given I have a memory with content "Meeting notes from standup"
    And the memory is tagged with "work"
    When I store the memory
    Then the memory should be saved with a unique ID
    And the memory should have a created timestamp
    And the vector embedding should be stored

  Scenario: Search memories by vector similarity
    Given I have stored several memories with embeddings
    When I search for memories similar to "standup meeting"
    Then I should receive the most similar memories
    And each result should have a similarity score
    And results should be ordered by relevance

  Scenario: Full-text search using FTS5
    Given I have memories with content about "authentication"
    When I search for "auth" using FTS5
    Then I should receive memories containing "auth"
    And the search should be case-insensitive
    And results should be ranked by relevance

  Scenario: Retrieve memory by ID
    Given a memory exists with ID "memory-123"
    When I retrieve the memory by ID
    Then I should receive the full memory content
    And all tags should be included
    And metadata should be preserved

  Scenario: Update memory tags
    Given a memory with tags "work, important"
    When I add tag "urgent" to the memory
    Then the memory should now have three tags
    And existing tags should be preserved

  Scenario: Delete memory
    Given a memory exists with ID "memory-to-delete"
    When I delete the memory
    Then the memory should be removed from storage
    And its vector should also be deleted
    And subsequent retrieval should return nil

  Scenario: Add edge between memories
    Given memory "A" and memory "B" exist
    When I add an edge from "A" to "B" with relation "relates_to"
    Then the edge should be stored in the graph
    And I should be able to query edges for memory "A"

  Scenario: Query memory edges
    Given memory "A" has edges to memories "B" and "C"
    When I query edges for memory "A"
    Then I should receive both connected memories
    And each edge should have its relation type

  Scenario: List all tags
    Given memories with tags "work", "personal", "work"
    When I list all unique tags
    Then I should receive "work" and "personal"
    And duplicates should be removed

  Scenario: Prune memories by tag
    Given I have memories with tags "keep-me" and "discard"
    When I prune memories keeping only "keep-me"
    Then memories with "discard" should be deleted
    And memories with "keep-me" should be preserved
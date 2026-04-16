// Package memory provides the Mimir memory engine for ODIN
package memory

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/odin-ai/odin/pkg/logger"
)

// SearchOptions contains search configuration options
type SearchOptions struct {
	Limit     int
	Offset    int
	Project   string
	Tags      []string
	Since     time.Time
	Until     time.Time
	Encrypted *bool // nil means ignore, true means only encrypted, false means only unencrypted
	MinScore  float64
	SortBy    string // "relevance", "date", "access"
	SortDesc  bool
}

// DefaultSearchOptions returns the default search options
func DefaultSearchOptions() *SearchOptions {
	return &SearchOptions{
		Limit:    10,
		Offset:   0,
		SortBy:   "relevance",
		SortDesc: true,
		MinScore: 0.0,
	}
}

// AdvancedSearch performs an advanced search with multiple filters
func (s *Store) AdvancedSearch(query string, opts *SearchOptions) ([]SearchResult, error) {
	if opts == nil {
		opts = DefaultSearchOptions()
	}
	if opts.Limit <= 0 {
		opts.Limit = 10
	}

	// Get base search results
	var baseResults []SearchResult
	var err error

	if query != "" {
		baseResults, err = s.Search(query, opts.Limit+opts.Offset)
		if err != nil {
			return nil, fmt.Errorf("search failed: %w", err)
		}
	} else {
		// No query, just filter all memories
		memories, err := s.db.ListMemories(opts.Project)
		if err != nil {
			return nil, fmt.Errorf("failed to list memories: %w", err)
		}
		baseResults = make([]SearchResult, len(memories))
		for i, m := range memories {
			baseResults[i] = SearchResult{Memory: m, Score: 1.0}
		}
	}

	// Apply filters
	filtered := make([]SearchResult, 0, len(baseResults))
	for _, result := range baseResults {
		m := result.Memory

		// Project filter
		if opts.Project != "" && m.Project != opts.Project {
			continue
		}

		// Tags filter (must have ALL specified tags)
		if len(opts.Tags) > 0 {
			hasAllTags := true
			for _, tag := range opts.Tags {
				found := false
				for _, mtag := range m.Tags {
					if strings.EqualFold(strings.TrimSpace(mtag), strings.TrimSpace(tag)) {
						found = true
						break
					}
				}
				if !found {
					hasAllTags = false
					break
				}
			}
			if !hasAllTags {
				continue
			}
		}

		// Date filters
		if !opts.Since.IsZero() && m.CreatedAt.Before(opts.Since) {
			continue
		}
		if !opts.Until.IsZero() && m.CreatedAt.After(opts.Until) {
			continue
		}

		// Encryption filter
		if opts.Encrypted != nil {
			if m.Encrypted != *opts.Encrypted {
				continue
			}
		}

		// Score filter
		if result.Score < opts.MinScore {
			continue
		}

		filtered = append(filtered, result)
	}

	// Apply offset
	if opts.Offset > 0 && opts.Offset < len(filtered) {
		filtered = filtered[opts.Offset:]
	}

	// Apply limit
	if opts.Limit < len(filtered) {
		filtered = filtered[:opts.Limit]
	}

	// Sort results
	s.sortResults(filtered, opts.SortBy, opts.SortDesc)

	return filtered, nil
}

// sortResults sorts search results by the specified criteria
func (s *Store) sortResults(results []SearchResult, sortBy string, descending bool) {
	switch strings.ToLower(sortBy) {
	case "date", "created", "created_at":
		sort.Slice(results, func(i, j int) bool {
			cmp := results[i].Memory.CreatedAt.Before(results[j].Memory.CreatedAt)
			if descending {
				return !cmp
			}
			return cmp
		})
	case "access", "accessed", "accessed_at":
		sort.Slice(results, func(i, j int) bool {
			cmp := results[i].Memory.AccessedAt.Before(results[j].Memory.AccessedAt)
			if descending {
				return !cmp
			}
			return cmp
		})
	case "update", "updated", "updated_at":
		sort.Slice(results, func(i, j int) bool {
			cmp := results[i].Memory.UpdatedAt.Before(results[j].Memory.UpdatedAt)
			if descending {
				return !cmp
			}
			return cmp
		})
	case "score", "relevance":
		sort.Slice(results, func(i, j int) bool {
			cmp := results[i].Score < results[j].Score
			if descending {
				return !cmp
			}
			return cmp
		})
	default:
		// Default: sort by relevance descending
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})
	}
}

// SearchByTags searches for memories with specific tags
func (s *Store) SearchByTags(tags []string, matchAll bool, limit int) ([]*Memory, error) {
	if limit <= 0 {
		limit = 10
	}

	memories, err := s.db.ListMemories("")
	if err != nil {
		return nil, fmt.Errorf("failed to list memories: %w", err)
	}

	var results []*Memory
	for _, m := range memories {
		if matchAll {
			// Must have ALL tags
			hasAll := true
			for _, tag := range tags {
				found := false
				for _, mtag := range m.Tags {
					if strings.EqualFold(strings.TrimSpace(mtag), strings.TrimSpace(tag)) {
						found = true
						break
					}
				}
				if !found {
					hasAll = false
					break
				}
			}
			if hasAll {
				results = append(results, m)
			}
		} else {
			// Must have ANY of the tags
			for _, tag := range tags {
				for _, mtag := range m.Tags {
					if strings.EqualFold(strings.TrimSpace(mtag), strings.TrimSpace(tag)) {
						results = append(results, m)
						break
					}
				}
			}
		}

		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

// SearchByProject searches for memories in a specific project
func (s *Store) SearchByProject(project string, limit int) ([]*Memory, error) {
	if limit <= 0 {
		limit = 10
	}

	memories, err := s.db.ListMemories(project)
	if err != nil {
		return nil, fmt.Errorf("failed to list memories: %w", err)
	}

	if limit < len(memories) {
		return memories[:limit], nil
	}
	return memories, nil
}

// SearchByDateRange searches for memories within a date range
func (s *Store) SearchByDateRange(since, until time.Time, limit int) ([]*Memory, error) {
	if limit <= 0 {
		limit = 10
	}

	memories, err := s.db.ListMemories("")
	if err != nil {
		return nil, fmt.Errorf("failed to list memories: %w", err)
	}

	var results []*Memory
	for _, m := range memories {
		if !since.IsZero() && m.CreatedAt.Before(since) {
			continue
		}
		if !until.IsZero() && m.CreatedAt.After(until) {
			continue
		}
		results = append(results, m)
		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

// SearchSimilar finds memories similar to a given memory
func (s *Store) SearchSimilar(memoryID string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 5
	}

	m, err := s.Recall(memoryID)
	if err != nil || m == nil {
		return nil, fmt.Errorf("memory not found: %s", memoryID)
	}

	// Use content-based search
	return s.Search(m.Content, limit)
}

// AggregateSearch performs multiple searches and combines results
func (s *Store) AggregateSearch(queries []string, strategy string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if len(queries) == 0 {
		return nil, nil
	}

	var allResults []SearchResult
	seen := make(map[string]bool)

	for _, query := range queries {
		results, err := s.Search(query, limit)
		if err != nil {
			continue
		}

		for _, r := range results {
			if !seen[r.Memory.ID] {
				seen[r.Memory.ID] = true
				allResults = append(allResults, r)
			}
		}
	}

	// Sort and limit
	switch strings.ToLower(strategy) {
	case "merge":
		// Sort by score descending
		sort.Slice(allResults, func(i, j int) bool {
			return allResults[i].Score > allResults[j].Score
		})
	case "intersect":
		// Keep only memories that appear in all searches
		filtered := make([]SearchResult, 0)
		for _, r := range allResults {
			count := 0
			for _, query := range queries {
				results, _ := s.Search(query, 100)
				for _, sr := range results {
					if sr.Memory.ID == r.Memory.ID {
						count++
						break
					}
				}
			}
			if count == len(queries) {
				filtered = append(filtered, r)
			}
		}
		allResults = filtered
	default:
		// Default to merge
		sort.Slice(allResults, func(i, j int) bool {
			return allResults[i].Score > allResults[j].Score
		})
	}

	if limit < len(allResults) {
		return allResults[:limit], nil
	}
	return allResults, nil
}

// SearchResultGroup represents a group of search results
type SearchResultGroup struct {
	Label   string
	Results []SearchResult
	Total   int
}

// GroupedSearch performs multiple searches and groups results
func (s *Store) GroupedSearch(queries map[string]string) ([]SearchResultGroup, error) {
	groups := make([]SearchResultGroup, 0, len(queries))

	for label, query := range queries {
		results, err := s.Search(query, 10)
		if err != nil {
			logger.Warn("Search failed for group", "label", label, "error", err)
			continue
		}

		groups = append(groups, SearchResultGroup{
			Label:   label,
			Results: results,
			Total:   len(results),
		})
	}

	return groups, nil
}

// SearchSuggestion represents a search suggestion
type SearchSuggestion struct {
	Text  string `json:"text"`
	Type  string `json:"type"` // "tag", "project", "keyword"
	Count int    `json:"count"`
}

// GetSuggestions returns search suggestions based on partial input
func (s *Store) GetSuggestions(partial string, limit int) ([]SearchSuggestion, error) {
	if limit <= 0 {
		limit = 5
	}

	partial = strings.ToLower(strings.TrimSpace(partial))
	if partial == "" {
		return nil, nil
	}

	var suggestions []SearchSuggestion

	// Get matching tags
	_, err := s.db.ListTags()
	if err == nil {
		tagCounts := make(map[string]int)
		memories, _ := s.db.ListMemories("")
		for _, m := range memories {
			for _, tag := range m.Tags {
				tag = strings.TrimSpace(tag)
				if strings.Contains(strings.ToLower(tag), partial) {
					tagCounts[tag]++
				}
			}
		}

		for tag, count := range tagCounts {
			suggestions = append(suggestions, SearchSuggestion{
				Text:  tag,
				Type:  "tag",
				Count: count,
			})
		}
	}

	// Get matching projects
	memories, _ := s.db.ListMemories("")
	projectCounts := make(map[string]int)
	for _, m := range memories {
		if m.Project != "" && strings.Contains(strings.ToLower(m.Project), partial) {
			projectCounts[m.Project]++
		}
	}
	for project, count := range projectCounts {
		suggestions = append(suggestions, SearchSuggestion{
			Text:  project,
			Type:  "project",
			Count: count,
		})
	}

	// Limit results
	if limit < len(suggestions) {
		suggestions = suggestions[:limit]
	}

	return suggestions, nil
}

// SearchAnalytics contains analytics about search usage
type SearchAnalytics struct {
	TotalSearches      int
	AverageResultCount float64
	TopQueries         []QueryStats
	TopTags            []TagStats
}

// QueryStats contains statistics about a specific query
type QueryStats struct {
	Query      string
	Count      int
	AvgResults float64
}

// TagStats contains statistics about a tag
type TagStats struct {
	Tag   string
	Count int
}

// GetAnalytics returns search analytics
func (s *Store) GetAnalytics() (*SearchAnalytics, error) {
	// This would typically track search history
	// For now, return basic stats
	memories, err := s.db.ListMemories("")
	if err != nil {
		return nil, err
	}

	_, _ = s.db.ListTags() // Check available tags exist

	analytics := &SearchAnalytics{
		TotalSearches:      0, // Would be tracked
		AverageResultCount: 0,
		TopTags:            make([]TagStats, 0),
	}

	// Count tag usage
	tagCounts := make(map[string]int)
	for _, m := range memories {
		for _, tag := range m.Tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tagCounts[tag]++
			}
		}
	}

	for tag, count := range tagCounts {
		analytics.TopTags = append(analytics.TopTags, TagStats{
			Tag:   tag,
			Count: count,
		})
	}

	// Sort tags by count
	sort.Slice(analytics.TopTags, func(i, j int) bool {
		return analytics.TopTags[i].Count > analytics.TopTags[j].Count
	})

	// Limit to top 10
	if len(analytics.TopTags) > 10 {
		analytics.TopTags = analytics.TopTags[:10]
	}

	return analytics, nil
}

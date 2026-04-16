// Package memory provides the Mimir memory engine for ODIN
package memory

import (
	"fmt"
	"sort"
)

// GraphQuery represents a query on the knowledge graph
type GraphQuery struct {
	FromID       string
	Depth        int
	RelationType string // e.g., "depends_on", "inspired_by"
}

// GraphResult represents the result of a graph query
type GraphResult struct {
	Memory   *Memory
	Path     []string // Path of memory IDs from source
	Distance int      // Depth from source
}

// GraphEngine handles knowledge graph operations
type GraphEngine struct {
	db *DB
}

// NewGraphEngine creates a new GraphEngine
func NewGraphEngine(db *DB) *GraphEngine {
	return &GraphEngine{db: db}
}

// Query traverses the knowledge graph from a starting memory
func (g *GraphEngine) Query(query GraphQuery) ([]GraphResult, error) {
	if query.Depth <= 0 {
		query.Depth = 3 // Default depth
	}

	visited := make(map[string]bool)
	results := make([]GraphResult, 0)

	// BFS traversal
	queue := []graphNode{
		{memoryID: query.FromID, path: []string{query.FromID}, distance: 0},
	}

	for len(queue) > 0 && len(results) < 100 { // Limit results
		current := queue[0]
		queue = queue[1:]

		if visited[current.memoryID] {
			continue
		}
		visited[current.memoryID] = true

		// Get the memory
		m, _, err := g.db.GetMemory(current.memoryID)
		if err != nil || m == nil {
			continue
		}

		// Add to results (except the source)
		if current.distance > 0 {
			results = append(results, GraphResult{
				Memory:   m,
				Path:     current.path,
				Distance: current.distance,
			})
		}

		// Stop if we've reached max depth
		if current.distance >= query.Depth {
			continue
		}

		// Get connected memories via edges
		edges, err := g.db.GetEdges(current.memoryID)
		if err != nil {
			continue
		}

		for _, edge := range edges {
			// Filter by relation type if specified
			if query.RelationType != "" && edge.Relation != query.RelationType {
				continue
			}

			// Add neighbor (to_id) if not visited
			if !visited[edge.ToID] {
				newPath := make([]string, len(current.path)+1)
				copy(newPath, current.path)
				newPath[len(current.path)] = edge.ToID
				queue = append(queue, graphNode{
					memoryID: edge.ToID,
					path:     newPath,
					distance: current.distance + 1,
				})
			}

			// Add neighbor (from_id) if not visited
			if !visited[edge.FromID] && edge.FromID != current.memoryID {
				newPath := make([]string, len(current.path)+1)
				copy(newPath, current.path)
				newPath[len(current.path)] = edge.FromID
				queue = append(queue, graphNode{
					memoryID: edge.FromID,
					path:     newPath,
					distance: current.distance + 1,
				})
			}
		}
	}

	return results, nil
}

type graphNode struct {
	memoryID string
	path     []string
	distance int
}

// FindPath finds a path between two memories
func (g *GraphEngine) FindPath(fromID, toID string, maxDepth int) ([]string, error) {
	if maxDepth <= 0 {
		maxDepth = 5
	}

	visited := make(map[string]bool)
	parent := make(map[string]string)

	queue := []string{fromID}
	visited[fromID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == toID {
			// Reconstruct path
			path := make([]string, 0)
			for id := toID; id != ""; id = parent[id] {
				path = append(path, id)
			}
			// Reverse to get from -> to
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}
			return path, nil
		}

		edges, err := g.db.GetEdges(current)
		if err != nil {
			continue
		}

		for _, edge := range edges {
			neighbor := edge.ToID
			if neighbor == current {
				neighbor = edge.FromID
			}

			if !visited[neighbor] {
				visited[neighbor] = true
				parent[neighbor] = current
				queue = append(queue, neighbor)
			}
		}
	}

	return nil, fmt.Errorf("no path found between %s and %s", fromID, toID)
}

// GetConnectedMemories returns all memories connected to a given memory
func (g *GraphEngine) GetConnectedMemories(memoryID string, relationType string) ([]*Memory, error) {
	edges, err := g.db.GetEdges(memoryID)
	if err != nil {
		return nil, err
	}

	memorySet := make(map[string]*Memory)
	for _, edge := range edges {
		if relationType != "" && edge.Relation != relationType {
			continue
		}

		// Get the other memory in the edge
		otherID := edge.ToID
		if otherID == memoryID {
			otherID = edge.FromID
		}

		m, _, err := g.db.GetMemory(otherID)
		if err != nil || m == nil {
			continue
		}
		memorySet[otherID] = m
	}

	memories := make([]*Memory, 0, len(memorySet))
	for _, m := range memorySet {
		memories = append(memories, m)
	}

	// Sort by updated time
	sort.Slice(memories, func(i, j int) bool {
		return memories[i].UpdatedAt.After(memories[j].UpdatedAt)
	})

	return memories, nil
}

// GetRelationCounts returns counts of each relation type
func (g *GraphEngine) GetRelationCounts(memoryID string) (map[string]int, error) {
	edges, err := g.db.GetEdges(memoryID)
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	for _, edge := range edges {
		counts[edge.Relation]++
	}

	return counts, nil
}

// SuggestConnections suggests potential connections based on content similarity
func (g *GraphEngine) SuggestConnections(memoryID string, limit int) ([]*Memory, []string, error) {
	if limit <= 0 {
		limit = 5
	}

	m, _, err := g.db.GetMemory(memoryID)
	if err != nil || m == nil {
		return nil, nil, fmt.Errorf("memory not found: %s", memoryID)
	}

	// Get all memories
	allMemories, err := g.db.ListMemories("")
	if err != nil {
		return nil, nil, err
	}

	// Get connected memories to exclude
	connected, err := g.GetConnectedMemories(memoryID, "")
	if err != nil {
		return nil, nil, err
	}

	connectedSet := make(map[string]bool)
	for _, c := range connected {
		connectedSet[c.ID] = true
	}

	// Calculate similarity scores
	type candidate struct {
		memory     *Memory
		similarity int // Simple count of common words
	}

	var candidates []candidate
	for _, other := range allMemories {
		if other.ID == memoryID {
			continue
		}
		if connectedSet[other.ID] {
			continue
		}

		// Simple similarity: count common tags
		commonTags := 0
		tagSet := make(map[string]bool)
		for _, tag := range m.Tags {
			tagSet[tag] = true
		}
		for _, tag := range other.Tags {
			if tagSet[tag] {
				commonTags++
			}
		}

		if commonTags > 0 {
			candidates = append(candidates, candidate{other, commonTags})
		}
	}

	// Sort by similarity
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].similarity > candidates[j].similarity
	})

	// Take top k
	if limit > len(candidates) {
		limit = len(candidates)
	}

	memories := make([]*Memory, limit)
	reasons := make([]string, limit)
	for i := 0; i < limit; i++ {
		memories[i] = candidates[i].memory
		reasons[i] = fmt.Sprintf("shares %d tag(s)", candidates[i].similarity)
	}

	return memories, reasons, nil
}

// ClusterMemories finds clusters of related memories
func (g *GraphEngine) ClusterMemories(minConnections int) ([][]*Memory, error) {
	if minConnections <= 0 {
		minConnections = 2
	}

	// Get all edges
	allEdges, err := g.db.GetAllEdges()
	if err != nil {
		return nil, err
	}

	// Build adjacency list
	adjacency := make(map[string]map[string]bool)
	for _, edge := range allEdges {
		if adjacency[edge.FromID] == nil {
			adjacency[edge.FromID] = make(map[string]bool)
		}
		if adjacency[edge.ToID] == nil {
			adjacency[edge.ToID] = make(map[string]bool)
		}
		adjacency[edge.FromID][edge.ToID] = true
		adjacency[edge.ToID][edge.FromID] = true
	}

	// Find connected components using BFS
	visited := make(map[string]bool)
	var clusters [][]*Memory

	for startID := range adjacency {
		if visited[startID] {
			continue
		}

		// BFS from this node
		var cluster []*Memory
		queue := []string{startID}

		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			if visited[current] {
				continue
			}
			visited[current] = true

			m, _, err := g.db.GetMemory(current)
			if err != nil || m == nil {
				continue
			}
			cluster = append(cluster, m)

			for neighbor := range adjacency[current] {
				if !visited[neighbor] {
					queue = append(queue, neighbor)
				}
			}
		}

		// Only add clusters with enough connections
		if len(cluster) >= minConnections {
			clusters = append(clusters, cluster)
		}
	}

	return clusters, nil
}

// GetMemoryStats returns statistics about the knowledge graph
func (g *GraphEngine) GetMemoryStats() (*GraphStats, error) {
	memories, err := g.db.ListMemories("")
	if err != nil {
		return nil, err
	}

	edges, err := g.db.GetAllEdges()
	if err != nil {
		return nil, err
	}

	stats := &GraphStats{
		TotalMemories: len(memories),
		TotalEdges:    len(edges),
		RelationTypes: make(map[string]int),
		AverageDegree: 0,
	}

	// Count relation types and calculate average degree
	degreeSum := 0
	nodeDegrees := make(map[string]int)
	for _, edge := range edges {
		stats.RelationTypes[edge.Relation]++
		nodeDegrees[edge.FromID]++
		nodeDegrees[edge.ToID]++
	}

	for _, d := range nodeDegrees {
		degreeSum += d
	}
	if len(nodeDegrees) > 0 {
		stats.AverageDegree = float64(degreeSum) / float64(len(nodeDegrees))
	}

	// Find most connected memory
	maxDegree := 0
	var hubMemory *Memory
	for id, degree := range nodeDegrees {
		if degree > maxDegree {
			maxDegree = degree
			m, _, _ := g.db.GetMemory(id)
			hubMemory = m
		}
	}
	stats.MostConnectedMemory = hubMemory
	stats.MostConnectedDegree = maxDegree

	return stats, nil
}

// GraphStats holds statistics about the knowledge graph
type GraphStats struct {
	TotalMemories       int
	TotalEdges          int
	RelationTypes       map[string]int
	AverageDegree       float64
	MostConnectedMemory *Memory
	MostConnectedDegree int
}

// FindShortestPath is an alias for FindPath for API consistency
func (g *GraphEngine) FindShortestPath(fromID, toID string) ([]string, error) {
	return g.FindPath(fromID, toID, 5)
}

// GetMemoryGraph returns the subgraph centered on a memory
func (g *GraphEngine) GetMemoryGraph(memoryID string, depth int) (*MemoryGraph, error) {
	query := GraphQuery{
		FromID: memoryID,
		Depth:  depth,
	}

	results, err := g.Query(query)
	if err != nil {
		return nil, err
	}

	graph := &MemoryGraph{
		Nodes: make([]*Memory, 0, len(results)+1),
		Edges: make([]*MemoryEdge, 0),
	}

	// Add center memory
	center, _, err := g.db.GetMemory(memoryID)
	if err == nil && center != nil {
		graph.Nodes = append(graph.Nodes, center)
	}

	// Add result memories
	nodeSet := make(map[string]bool)
	for _, r := range results {
		if !nodeSet[r.Memory.ID] {
			graph.Nodes = append(graph.Nodes, r.Memory)
			nodeSet[r.Memory.ID] = true
		}

		// Reconstruct edges from path
		for i := 0; i < len(r.Path)-1; i++ {
			edges, _ := g.db.GetEdges(r.Path[i])
			for _, edge := range edges {
				if (edge.FromID == r.Path[i] && edge.ToID == r.Path[i+1]) ||
					(edge.ToID == r.Path[i] && edge.FromID == r.Path[i+1]) {
					graph.Edges = append(graph.Edges, edge)
				}
			}
		}
	}

	return graph, nil
}

// MemoryGraph represents a subgraph of memories and edges
type MemoryGraph struct {
	Nodes []*Memory
	Edges []*MemoryEdge
}

package graph

// ExportData exports graph data for serialization
type ExportData struct {
	Nodes         map[int64]*Node
	Edges         map[int64][]Edge
	ReverseEdges  map[int64][]Edge
	Restrictions  map[int64][]TurnRestriction
}

// Export exports the graph data
func (g *Graph) Export() *ExportData {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	return &ExportData{
		Nodes:         g.nodes,
		Edges:         g.edges,
		ReverseEdges:  g.reverseEdges,
		Restrictions:  g.restrictions,
	}
}

// Import imports graph data
func (g *Graph) Import(data *ExportData) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	g.nodes = data.Nodes
	g.edges = data.Edges
	
	if data.ReverseEdges != nil {
		g.reverseEdges = data.ReverseEdges
	} else {
		// Rebuild reverse edges if not present in data
		g.reverseEdges = make(map[int64][]Edge)
		for _, edgeList := range data.Edges {
			for _, edge := range edgeList {
				g.reverseEdges[edge.To] = append(g.reverseEdges[edge.To], edge)
			}
		}
	}
	
	if data.Restrictions != nil {
		g.restrictions = data.Restrictions
	} else {
		g.restrictions = make(map[int64][]TurnRestriction)
	}
}


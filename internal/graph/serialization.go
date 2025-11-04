package graph

// ExportData exports graph data for serialization
type ExportData struct {
	Nodes        map[int64]*Node
	Edges        map[int64][]Edge
	Restrictions map[int64][]TurnRestriction
}

// Export exports the graph data
func (g *Graph) Export() *ExportData {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	return &ExportData{
		Nodes:        g.nodes,
		Edges:        g.edges,
		Restrictions: g.restrictions,
	}
}

// Import imports graph data
func (g *Graph) Import(data *ExportData) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	g.nodes = data.Nodes
	g.edges = data.Edges
	if data.Restrictions != nil {
		g.restrictions = data.Restrictions
	} else {
		g.restrictions = make(map[int64][]TurnRestriction)
	}
}


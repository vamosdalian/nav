package graph

// ExportData exports graph data for serialization
type ExportData struct {
	Nodes map[int64]*Node
	Edges map[int64][]Edge
}

// Export exports the graph data
func (g *Graph) Export() *ExportData {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	return &ExportData{
		Nodes: g.nodes,
		Edges: g.edges,
	}
}

// Import imports graph data
func (g *Graph) Import(data *ExportData) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	g.nodes = data.Nodes
	g.edges = data.Edges
}


package graph

import (
	"fmt"
	"math"
	"sync"
)

// Node represents a geographic point in the road network
type Node struct {
	ID  int64
	Lat float64
	Lon float64
}

// Edge represents a road segment between two nodes
type Edge struct {
	From   int64
	To     int64
	Weight float64 // Cost (distance, time, etc.)
	OSMWayID int64
	MaxSpeed float64
	Tags     map[string]string
}

// Graph represents the road network
type Graph struct {
	nodes     map[int64]*Node
	edges     map[int64][]Edge // adjacency list: nodeID -> outgoing edges
	mutex     sync.RWMutex
}

// NewGraph creates a new empty graph
func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[int64]*Node),
		edges: make(map[int64][]Edge),
	}
}

// AddNode adds a node to the graph
func (g *Graph) AddNode(node *Node) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.nodes[node.ID] = node
}

// AddEdge adds an edge to the graph
func (g *Graph) AddEdge(edge Edge) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.edges[edge.From] = append(g.edges[edge.From], edge)
}

// GetNode returns a node by ID
func (g *Graph) GetNode(id int64) (*Node, error) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	node, exists := g.nodes[id]
	if !exists {
		return nil, fmt.Errorf("node %d not found", id)
	}
	return node, nil
}

// GetEdges returns all outgoing edges from a node
func (g *Graph) GetEdges(nodeID int64) []Edge {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.edges[nodeID]
}

// UpdateEdgeWeight updates the weight of a specific edge
func (g *Graph) UpdateEdgeWeight(from, to int64, newWeight float64) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	edges, exists := g.edges[from]
	if !exists {
		return fmt.Errorf("no edges from node %d", from)
	}
	
	found := false
	for i := range edges {
		if edges[i].To == to {
			edges[i].Weight = newWeight
			found = true
		}
	}
	
	if !found {
		return fmt.Errorf("edge from %d to %d not found", from, to)
	}
	
	return nil
}

// UpdateEdgeWeightByWay updates all edges belonging to a specific OSM way
func (g *Graph) UpdateEdgeWeightByWay(osmWayID int64, multiplier float64) int {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	count := 0
	for _, edgeList := range g.edges {
		for i := range edgeList {
			if edgeList[i].OSMWayID == osmWayID {
				edgeList[i].Weight *= multiplier
				count++
			}
		}
	}
	return count
}

// NodeCount returns the total number of nodes
func (g *Graph) NodeCount() int {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return len(g.nodes)
}

// EdgeCount returns the total number of edges
func (g *Graph) EdgeCount() int {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	count := 0
	for _, edges := range g.edges {
		count += len(edges)
	}
	return count
}

// FindNearestNode finds the closest node to given coordinates
func (g *Graph) FindNearestNode(lat, lon float64) (*Node, error) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	if len(g.nodes) == 0 {
		return nil, fmt.Errorf("graph is empty")
	}
	
	var nearest *Node
	minDist := math.MaxFloat64
	
	for _, node := range g.nodes {
		dist := HaversineDistance(lat, lon, node.Lat, node.Lon)
		if dist < minDist {
			minDist = dist
			nearest = node
		}
	}
	
	return nearest, nil
}

// HaversineDistance calculates the great-circle distance between two points (in meters)
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // meters
	
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180
	
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	
	return earthRadius * c
}


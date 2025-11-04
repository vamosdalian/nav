package osm

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
	"github.com/vamosdalian/nav/internal/graph"
)

// Parser handles OSM data parsing
type Parser struct {
	graph *graph.Graph
}

// NewParser creates a new OSM parser
func NewParser(g *graph.Graph) *Parser {
	return &Parser{graph: g}
}

// ParseFile parses an OSM PBF file and populates the graph
func (p *Parser) ParseFile(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()
	
	return p.Parse(f)
}

// Parse parses OSM data from a reader
func (p *Parser) Parse(r io.Reader) error {
	// Use scanner with 3 concurrent decoders for better compatibility
	scanner := osmpbf.New(context.Background(), r, 3)
	defer scanner.Close()
	
	// First pass: collect all nodes
	allNodes := make(map[int64]*graph.Node)
	
	// Collect ways
	ways := make([]*osm.Way, 0)
	
	for scanner.Scan() {
		obj := scanner.Object()
		
		switch v := obj.(type) {
		case *osm.Node:
			allNodes[int64(v.ID)] = &graph.Node{
				ID:  int64(v.ID),
				Lat: v.Lat,
				Lon: v.Lon,
			}
			
		case *osm.Way:
			if p.isRoutableWay(v) {
				ways = append(ways, v)
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}
	
	// Find which nodes are actually used in routable ways
	usedNodes := make(map[int64]bool)
	for _, way := range ways {
		for _, node := range way.Nodes {
			usedNodes[int64(node.ID)] = true
		}
	}
	
	// Add only used nodes to graph
	nodeCount := 0
	for nodeID := range usedNodes {
		if node, exists := allNodes[nodeID]; exists {
			p.graph.AddNode(node)
			nodeCount++
		}
	}
	
	// Process ways and create edges
	edgeCount := 0
	for _, way := range ways {
		edgesBefore := p.graph.EdgeCount()
		p.processWay(way, allNodes)
		edgeCount += p.graph.EdgeCount() - edgesBefore
	}
	
	fmt.Printf("Loaded %d nodes (from %d total) and %d routable ways\n", nodeCount, len(allNodes), len(ways))
	
	return nil
}

// isRoutableWay checks if a way is routable (road)
func (p *Parser) isRoutableWay(way *osm.Way) bool {
	highway := way.Tags.Find("highway")
	if highway == "" {
		return false
	}
	
	// Filter out non-routable highways
	nonRoutable := map[string]bool{
		"footway":     true,
		"path":        true,
		"steps":       true,
		"cycleway":    true,
		"pedestrian":  true,
		"construction": true,
		"proposed":    true,
	}
	
	return !nonRoutable[highway]
}

// processWay creates edges from a way
func (p *Parser) processWay(way *osm.Way, nodes map[int64]*graph.Node) {
	if len(way.Nodes) < 2 {
		return
	}
	
	oneway := false
	if v := way.Tags.Find("oneway"); v == "yes" {
		oneway = true
	}
	
	maxSpeed := p.getMaxSpeed(way)
	tags := p.extractTags(way)
	
	for i := 0; i < len(way.Nodes)-1; i++ {
		fromID := int64(way.Nodes[i].ID)
		toID := int64(way.Nodes[i+1].ID)
		
		fromNode, fromExists := nodes[fromID]
		toNode, toExists := nodes[toID]
		
		if !fromExists || !toExists {
			continue
		}
		
		distance := graph.HaversineDistance(
			fromNode.Lat, fromNode.Lon,
			toNode.Lat, toNode.Lon,
		)
		
		// Create forward edge
		p.graph.AddEdge(graph.Edge{
			From:     fromID,
			To:       toID,
			Weight:   distance,
			OSMWayID: int64(way.ID),
			MaxSpeed: maxSpeed,
			Tags:     tags,
		})
		
		// Create backward edge if not oneway
		if !oneway {
			p.graph.AddEdge(graph.Edge{
				From:     toID,
				To:       fromID,
				Weight:   distance,
				OSMWayID: int64(way.ID),
				MaxSpeed: maxSpeed,
				Tags:     tags,
			})
		}
	}
}

// getMaxSpeed extracts maximum speed from way tags (in m/s)
func (p *Parser) getMaxSpeed(way *osm.Way) float64 {
	if maxspeed := way.Tags.Find("maxspeed"); maxspeed != "" {
		// Simple parsing - would need more robust parsing in production
		var speed float64
		fmt.Sscanf(maxspeed, "%f", &speed)
		return speed / 3.6 // Convert km/h to m/s
	}
	
	// Default speeds based on highway type (m/s)
	highway := way.Tags.Find("highway")
	defaults := map[string]float64{
		"motorway":       33.33, // 120 km/h
		"trunk":          27.78, // 100 km/h
		"primary":        22.22, // 80 km/h
		"secondary":      19.44, // 70 km/h
		"tertiary":       13.89, // 50 km/h
		"residential":    8.33,  // 30 km/h
		"service":        5.56,  // 20 km/h
		"unclassified":   13.89, // 50 km/h
	}
	
	if speed, exists := defaults[highway]; exists {
		return speed
	}
	
	return 13.89 // Default 50 km/h
}

// extractTags extracts relevant tags from way
func (p *Parser) extractTags(way *osm.Way) map[string]string {
	tags := make(map[string]string)
	
	relevantKeys := []string{"highway", "name", "surface", "lanes"}
	for _, key := range relevantKeys {
		if value := way.Tags.Find(key); value != "" {
			tags[key] = value
		}
	}
	
	return tags
}


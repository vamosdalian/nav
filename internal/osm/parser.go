package osm

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

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

	// Collect relations (for turn restrictions)
	relations := make([]*osm.Relation, 0)

	// Progress tracking
	var nodeCount, wayCount, relationCount int64
	lastProgressTime := time.Now()
	progressInterval := 5 * time.Second

	log.Println("Phase 1/5: Scanning OSM data...")

	for scanner.Scan() {
		obj := scanner.Object()

		switch v := obj.(type) {
		case *osm.Node:
			allNodes[int64(v.ID)] = &graph.Node{
				ID:  int64(v.ID),
				Lat: v.Lat,
				Lon: v.Lon,
			}
			nodeCount++

		case *osm.Way:
			if p.isRoutableWay(v) {
				ways = append(ways, v)
			}
			wayCount++

		case *osm.Relation:
			if p.isRestrictionRelation(v) {
				relations = append(relations, v)
			}
			relationCount++
		}

		// Show progress every 2 seconds
		if time.Since(lastProgressTime) >= progressInterval {
			log.Printf("  Progress: %d nodes, %d ways (%d routable), %d relations scanned...",
				nodeCount, wayCount, len(ways), relationCount)
			lastProgressTime = time.Now()
		}
	}

	log.Printf("Phase 1/5: Complete - Scanned %d nodes, %d ways, %d relations",
		nodeCount, wayCount, relationCount)

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	// Find which nodes are actually used in routable ways
	log.Println("Phase 2/5: Filtering used nodes...")
	usedNodes := make(map[int64]bool)
	for _, way := range ways {
		for _, node := range way.Nodes {
			usedNodes[int64(node.ID)] = true
		}
	}
	log.Printf("Phase 2/5: Complete - Found %d used nodes (from %d total)",
		len(usedNodes), len(allNodes))

	// Add only used nodes to graph
	log.Println("Phase 3/5: Adding nodes to graph...")
	graphNodeCount := 0
	for nodeID := range usedNodes {
		if node, exists := allNodes[nodeID]; exists {
			p.graph.AddNode(node)
			graphNodeCount++
		}
	}
	log.Printf("Phase 3/5: Complete - Added %d nodes to graph", graphNodeCount)

	// Process ways and create edges
	log.Println("Phase 4/5: Processing ways and creating edges...")
	lastProgressTime = time.Now()
	processedWays := 0
	for _, way := range ways {
		p.processWay(way, allNodes)
		processedWays++

		// Show progress every 2 seconds
		if time.Since(lastProgressTime) >= progressInterval {
			progress := float64(processedWays) / float64(len(ways)) * 100
			log.Printf("  Progress: %d/%d ways processed (%.1f%%), %d edges created...",
				processedWays, len(ways), progress, p.graph.EdgeCount())
			lastProgressTime = time.Now()
		}
	}
	log.Printf("Phase 4/5: Complete - Processed %d ways, created %d edges",
		len(ways), p.graph.EdgeCount())

	// Process turn restrictions
	log.Println("Phase 5/5: Processing turn restrictions...")
	restrictionCount := 0
	for _, relation := range relations {
		if p.processRestriction(relation) {
			restrictionCount++
		}
	}
	log.Printf("Phase 5/5: Complete - Processed %d turn restrictions", restrictionCount)

	log.Printf("✓ OSM parsing complete: %d nodes, %d edges, %d restrictions",
		graphNodeCount, p.graph.EdgeCount(), restrictionCount)

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
		"footway":      true,
		"path":         true,
		"steps":        true,
		"cycleway":     true,
		"pedestrian":   true,
		"construction": true,
		"proposed":     true,
	}

	return !nonRoutable[highway]
}

// processWay creates edges from a way
func (p *Parser) processWay(way *osm.Way, nodes map[int64]*graph.Node) {
	if len(way.Nodes) < 2 {
		return
	}

	// Check for oneway restrictions
	oneway := false
	onewayTag := way.Tags.Find("oneway")
	if onewayTag == "yes" || onewayTag == "1" || onewayTag == "true" {
		oneway = true
	}
	// Special case: oneway=-1 means reverse direction only
	reverseOneway := (onewayTag == "-1" || onewayTag == "reverse")

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

		// Create forward edge (unless reverse oneway)
		if !reverseOneway {
			p.graph.AddEdge(graph.Edge{
				From:     fromID,
				To:       toID,
				Weight:   distance,
				OSMWayID: int64(way.ID),
				MaxSpeed: maxSpeed,
				Tags:     tags,
			})
		}

		// Create backward edge if not oneway (or if reverse oneway)
		if !oneway || reverseOneway {
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
		"motorway":     33.33, // 120 km/h
		"trunk":        27.78, // 100 km/h
		"primary":      22.22, // 80 km/h
		"secondary":    19.44, // 70 km/h
		"tertiary":     13.89, // 50 km/h
		"residential":  8.33,  // 30 km/h
		"service":      5.56,  // 20 km/h
		"unclassified": 13.89, // 50 km/h
	}

	if speed, exists := defaults[highway]; exists {
		return speed
	}

	return 13.89 // Default 50 km/h
}

// extractTags extracts relevant tags from way
func (p *Parser) extractTags(way *osm.Way) map[string]string {
	tags := make(map[string]string)

	relevantKeys := []string{"highway", "name", "surface", "lanes", "oneway"}
	for _, key := range relevantKeys {
		if value := way.Tags.Find(key); value != "" {
			tags[key] = value
		}
	}

	return tags
}

// isRestrictionRelation checks if a relation is a turn restriction
func (p *Parser) isRestrictionRelation(relation *osm.Relation) bool {
	relType := relation.Tags.Find("type")
	return relType == "restriction" || relType == "restriction:conditional"
}

// processRestriction processes a turn restriction relation
func (p *Parser) processRestriction(relation *osm.Relation) bool {
	restriction := relation.Tags.Find("restriction")
	if restriction == "" {
		return false
	}

	var fromWay, toWay int64
	var viaNode int64

	// Parse relation members
	for _, member := range relation.Members {
		switch member.Role {
		case "from":
			if member.Type == osm.TypeWay {
				fromWay = int64(member.Ref)
			}
		case "to":
			if member.Type == osm.TypeWay {
				toWay = int64(member.Ref)
			}
		case "via":
			if member.Type == osm.TypeNode {
				viaNode = int64(member.Ref)
			}
			// Note: Some restrictions use via way, we'll skip those for now
		}
	}

	// We only handle simple node-based restrictions
	if fromWay == 0 || toWay == 0 || viaNode == 0 {
		return false
	}

	// Add restriction to graph
	p.graph.AddRestriction(graph.TurnRestriction{
		FromWay: fromWay,
		ViaNode: viaNode,
		ToWay:   toWay,
		Type:    restriction,
	})

	return true
}

package storage

import (
	"os"
	"testing"

	"github.com/vamosdalian/nav/internal/graph"
)

func TestSaveAndLoad(t *testing.T) {
	// Create a test graph
	g := createTestGraph()

	// Create temporary file
	tmpFile := "test_graph.bin.snappy"
	defer os.Remove(tmpFile)

	// Save graph
	store := NewStorage(tmpFile)
	if err := store.Save(g); err != nil {
		t.Fatalf("Failed to save graph: %v", err)
	}

	// Load graph
	loadedGraph, err := store.Load()
	if err != nil {
		t.Fatalf("Failed to load graph: %v", err)
	}

	// Verify loaded graph matches original
	verifyGraphsEqual(t, g, loadedGraph)
}

func TestSaveAndLoadEmptyGraph(t *testing.T) {
	// Create empty graph
	g := graph.NewGraph()

	tmpFile := "test_empty_graph.bin.snappy"
	defer os.Remove(tmpFile)

	// Save empty graph
	store := NewStorage(tmpFile)
	if err := store.Save(g); err != nil {
		t.Fatalf("Failed to save empty graph: %v", err)
	}

	// Load empty graph
	loadedGraph, err := store.Load()
	if err != nil {
		t.Fatalf("Failed to load empty graph: %v", err)
	}

	// Verify counts
	if loadedGraph.NodeCount() != 0 {
		t.Errorf("Expected 0 nodes, got %d", loadedGraph.NodeCount())
	}
	if loadedGraph.EdgeCount() != 0 {
		t.Errorf("Expected 0 edges, got %d", loadedGraph.EdgeCount())
	}
}

func TestSaveAndLoadLargeGraph(t *testing.T) {
	// Create a larger test graph
	g := graph.NewGraph()

	// Add 10000 nodes
	for i := int64(1); i <= 10000; i++ {
		g.AddNode(&graph.Node{
			ID:  i,
			Lat: 13.0 + float64(i)*0.0001,
			Lon: 100.0 + float64(i)*0.0001,
		})
	}

	// Add 20000 edges
	for i := int64(1); i < 10000; i++ {
		g.AddEdge(graph.Edge{
			From:     i,
			To:       i + 1,
			Weight:   100.0 + float64(i),
			OSMWayID: 1000 + i,
			MaxSpeed: 30.0,
			Tags: map[string]string{
				"highway": "primary",
				"name":    "Test Road",
			},
		})
		g.AddEdge(graph.Edge{
			From:     i + 1,
			To:       i,
			Weight:   100.0 + float64(i),
			OSMWayID: 1000 + i,
			MaxSpeed: 30.0,
			Tags: map[string]string{
				"highway": "primary",
			},
		})
	}

	tmpFile := "test_large_graph.bin.snappy"
	defer os.Remove(tmpFile)

	// Save
	store := NewStorage(tmpFile)
	if err := store.Save(g); err != nil {
		t.Fatalf("Failed to save large graph: %v", err)
	}

	// Load
	loadedGraph, err := store.Load()
	if err != nil {
		t.Fatalf("Failed to load large graph: %v", err)
	}

	// Verify counts
	if loadedGraph.NodeCount() != g.NodeCount() {
		t.Errorf("Node count mismatch: expected %d, got %d", g.NodeCount(), loadedGraph.NodeCount())
	}
	if loadedGraph.EdgeCount() != g.EdgeCount() {
		t.Errorf("Edge count mismatch: expected %d, got %d", g.EdgeCount(), loadedGraph.EdgeCount())
	}
}

func TestSaveAndLoadWithRestrictions(t *testing.T) {
	g := graph.NewGraph()

	// Add nodes
	for i := int64(1); i <= 5; i++ {
		g.AddNode(&graph.Node{
			ID:  i,
			Lat: 13.0 + float64(i)*0.01,
			Lon: 100.0 + float64(i)*0.01,
		})
	}

	// Add edges
	g.AddEdge(graph.Edge{From: 1, To: 2, Weight: 100, OSMWayID: 101, MaxSpeed: 30})
	g.AddEdge(graph.Edge{From: 2, To: 3, Weight: 150, OSMWayID: 102, MaxSpeed: 30})
	g.AddEdge(graph.Edge{From: 3, To: 4, Weight: 200, OSMWayID: 103, MaxSpeed: 30})

	// Add turn restrictions
	g.AddRestriction(graph.TurnRestriction{
		FromWay: 101,
		ViaNode: 2,
		ToWay:   102,
		Type:    "no_left_turn",
	})
	g.AddRestriction(graph.TurnRestriction{
		FromWay: 102,
		ViaNode: 3,
		ToWay:   103,
		Type:    "no_right_turn",
	})

	tmpFile := "test_restrictions.bin.snappy"
	defer os.Remove(tmpFile)

	// Save and load
	store := NewStorage(tmpFile)
	if err := store.Save(g); err != nil {
		t.Fatalf("Failed to save graph with restrictions: %v", err)
	}

	loadedGraph, err := store.Load()
	if err != nil {
		t.Fatalf("Failed to load graph with restrictions: %v", err)
	}

	// Verify restrictions
	data := loadedGraph.Export()
	totalRestrictions := 0
	for _, resList := range data.Restrictions {
		totalRestrictions += len(resList)
	}
	if totalRestrictions != 2 {
		t.Errorf("Expected 2 restrictions, got %d", totalRestrictions)
	}
}

func TestSaveAndLoadWithComplexTags(t *testing.T) {
	g := graph.NewGraph()

	g.AddNode(&graph.Node{ID: 1, Lat: 13.7563, Lon: 100.5018})
	g.AddNode(&graph.Node{ID: 2, Lat: 13.7663, Lon: 100.5118})

	// Add edge with multiple tags
	g.AddEdge(graph.Edge{
		From:     1,
		To:       2,
		Weight:   500.5,
		OSMWayID: 12345,
		MaxSpeed: 25.0,
		Tags: map[string]string{
			"highway":  "residential",
			"name":     "ถนนสุขุมวิท",
			"surface":  "asphalt",
			"lanes":    "2",
			"oneway":   "yes",
			"maxspeed": "90",
		},
	})

	tmpFile := "test_tags.bin.snappy"
	defer os.Remove(tmpFile)

	// Save and load
	store := NewStorage(tmpFile)
	if err := store.Save(g); err != nil {
		t.Fatalf("Failed to save graph: %v", err)
	}

	loadedGraph, err := store.Load()
	if err != nil {
		t.Fatalf("Failed to load graph: %v", err)
	}

	// Verify tags
	edges := loadedGraph.GetEdges(1)
	if len(edges) != 1 {
		t.Fatalf("Expected 1 edge, got %d", len(edges))
	}

	edge := edges[0]
	if len(edge.Tags) != 6 {
		t.Errorf("Expected 6 tags, got %d", len(edge.Tags))
	}

	expectedTags := map[string]string{
		"highway":  "residential",
		"name":     "ถนนสุขุมวิท",
		"surface":  "asphalt",
		"lanes":    "2",
		"oneway":   "yes",
		"maxspeed": "90",
	}

	for key, expectedValue := range expectedTags {
		if actualValue, exists := edge.Tags[key]; !exists {
			t.Errorf("Tag '%s' not found", key)
		} else if actualValue != expectedValue {
			t.Errorf("Tag '%s': expected '%s', got '%s'", key, expectedValue, actualValue)
		}
	}
}

func TestInvalidFileFormat(t *testing.T) {
	// Create file with invalid magic number
	tmpFile := "test_invalid.bin.snappy"
	defer os.Remove(tmpFile)

	// Write invalid data
	file, _ := os.Create(tmpFile)
	file.Write([]byte{0x00, 0x00, 0x00, 0x00}) // Wrong magic
	file.Close()

	// Try to load
	store := NewStorage(tmpFile)
	_, err := store.Load()
	if err == nil {
		t.Error("Expected error for invalid file format, got nil")
	}
}

func TestNonExistentFile(t *testing.T) {
	store := NewStorage("non_existent_file.bin.snappy")
	_, err := store.Load()
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestReverseEdges(t *testing.T) {
	g := graph.NewGraph()

	// Add nodes
	g.AddNode(&graph.Node{ID: 1, Lat: 13.0, Lon: 100.0})
	g.AddNode(&graph.Node{ID: 2, Lat: 13.1, Lon: 100.1})
	g.AddNode(&graph.Node{ID: 3, Lat: 13.2, Lon: 100.2})

	// Add edges (Graph.AddEdge automatically creates reverse edges)
	g.AddEdge(graph.Edge{From: 1, To: 2, Weight: 100, OSMWayID: 1, MaxSpeed: 30})
	g.AddEdge(graph.Edge{From: 2, To: 3, Weight: 200, OSMWayID: 2, MaxSpeed: 30})

	tmpFile := "test_reverse_edges.bin.snappy"
	defer os.Remove(tmpFile)

	// Save and load
	store := NewStorage(tmpFile)
	if err := store.Save(g); err != nil {
		t.Fatalf("Failed to save graph: %v", err)
	}

	loadedGraph, err := store.Load()
	if err != nil {
		t.Fatalf("Failed to load graph: %v", err)
	}

	// Verify reverse edges
	reverseEdges1 := loadedGraph.GetReverseEdges(2)
	if len(reverseEdges1) != 1 {
		t.Errorf("Expected 1 reverse edge to node 2, got %d", len(reverseEdges1))
	}

	reverseEdges2 := loadedGraph.GetReverseEdges(3)
	if len(reverseEdges2) != 1 {
		t.Errorf("Expected 1 reverse edge to node 3, got %d", len(reverseEdges2))
	}
}

// Helper function to create a test graph
func createTestGraph() *graph.Graph {
	g := graph.NewGraph()

	// Add nodes
	nodes := []*graph.Node{
		{ID: 1, Lat: 13.7563, Lon: 100.5018},
		{ID: 2, Lat: 13.7663, Lon: 100.5118},
		{ID: 3, Lat: 13.7763, Lon: 100.5218},
		{ID: 4, Lat: 13.7863, Lon: 100.5318},
	}

	for _, node := range nodes {
		g.AddNode(node)
	}

	// Add edges with tags
	edges := []graph.Edge{
		{
			From:     1,
			To:       2,
			Weight:   1234.56,
			OSMWayID: 100,
			MaxSpeed: 30.0,
			Tags: map[string]string{
				"highway": "primary",
				"name":    "Test Street",
			},
		},
		{
			From:     2,
			To:       3,
			Weight:   2345.67,
			OSMWayID: 101,
			MaxSpeed: 25.0,
			Tags: map[string]string{
				"highway": "secondary",
				"surface": "asphalt",
			},
		},
		{
			From:     3,
			To:       4,
			Weight:   3456.78,
			OSMWayID: 102,
			MaxSpeed: 20.0,
			Tags: map[string]string{
				"highway": "residential",
			},
		},
	}

	for _, edge := range edges {
		g.AddEdge(edge)
	}

	// Add restrictions
	g.AddRestriction(graph.TurnRestriction{
		FromWay: 100,
		ViaNode: 2,
		ToWay:   101,
		Type:    "no_left_turn",
	})

	return g
}

// Helper function to verify two graphs are equal
func verifyGraphsEqual(t *testing.T, original, loaded *graph.Graph) {
	// Check node count
	if original.NodeCount() != loaded.NodeCount() {
		t.Errorf("Node count mismatch: expected %d, got %d", original.NodeCount(), loaded.NodeCount())
	}

	// Check edge count
	if original.EdgeCount() != loaded.EdgeCount() {
		t.Errorf("Edge count mismatch: expected %d, got %d", original.EdgeCount(), loaded.EdgeCount())
	}

	// Export both graphs for detailed comparison
	origData := original.Export()
	loadedData := loaded.Export()

	// Verify nodes
	for id, origNode := range origData.Nodes {
		loadedNode, exists := loadedData.Nodes[id]
		if !exists {
			t.Errorf("Node %d missing in loaded graph", id)
			continue
		}
		if origNode.Lat != loadedNode.Lat || origNode.Lon != loadedNode.Lon {
			t.Errorf("Node %d mismatch: expected (%.6f, %.6f), got (%.6f, %.6f)",
				id, origNode.Lat, origNode.Lon, loadedNode.Lat, loadedNode.Lon)
		}
	}

	// Verify edges
	for from, origEdges := range origData.Edges {
		loadedEdges, exists := loadedData.Edges[from]
		if !exists {
			t.Errorf("Edges from node %d missing in loaded graph", from)
			continue
		}
		if len(origEdges) != len(loadedEdges) {
			t.Errorf("Edge count from node %d mismatch: expected %d, got %d",
				from, len(origEdges), len(loadedEdges))
			continue
		}

		// Create map for easier comparison
		loadedEdgeMap := make(map[int64]graph.Edge)
		for _, e := range loadedEdges {
			loadedEdgeMap[e.To] = e
		}

		for _, origEdge := range origEdges {
			loadedEdge, exists := loadedEdgeMap[origEdge.To]
			if !exists {
				t.Errorf("Edge %d->%d missing in loaded graph", origEdge.From, origEdge.To)
				continue
			}

			if origEdge.Weight != loadedEdge.Weight {
				t.Errorf("Edge %d->%d weight mismatch: expected %.2f, got %.2f",
					origEdge.From, origEdge.To, origEdge.Weight, loadedEdge.Weight)
			}
			if origEdge.OSMWayID != loadedEdge.OSMWayID {
				t.Errorf("Edge %d->%d OSMWayID mismatch: expected %d, got %d",
					origEdge.From, origEdge.To, origEdge.OSMWayID, loadedEdge.OSMWayID)
			}
			if origEdge.MaxSpeed != loadedEdge.MaxSpeed {
				t.Errorf("Edge %d->%d MaxSpeed mismatch: expected %.2f, got %.2f",
					origEdge.From, origEdge.To, origEdge.MaxSpeed, loadedEdge.MaxSpeed)
			}

			// Verify tags
			if len(origEdge.Tags) != len(loadedEdge.Tags) {
				t.Errorf("Edge %d->%d tag count mismatch: expected %d, got %d",
					origEdge.From, origEdge.To, len(origEdge.Tags), len(loadedEdge.Tags))
			}
			for key, origValue := range origEdge.Tags {
				loadedValue, exists := loadedEdge.Tags[key]
				if !exists {
					t.Errorf("Edge %d->%d tag '%s' missing", origEdge.From, origEdge.To, key)
				} else if origValue != loadedValue {
					t.Errorf("Edge %d->%d tag '%s' mismatch: expected '%s', got '%s'",
						origEdge.From, origEdge.To, key, origValue, loadedValue)
				}
			}
		}
	}
}

// Benchmark tests
func BenchmarkSave(b *testing.B) {
	g := createLargeBenchmarkGraph()
	tmpFile := "bench_save.bin.snappy"
	defer os.Remove(tmpFile)

	store := NewStorage(tmpFile)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := store.Save(g); err != nil {
			b.Fatalf("Save failed: %v", err)
		}
	}
}

func BenchmarkLoad(b *testing.B) {
	g := createLargeBenchmarkGraph()
	tmpFile := "bench_load.bin.snappy"
	defer os.Remove(tmpFile)

	store := NewStorage(tmpFile)
	if err := store.Save(g); err != nil {
		b.Fatalf("Save failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := store.Load(); err != nil {
			b.Fatalf("Load failed: %v", err)
		}
	}
}

func createLargeBenchmarkGraph() *graph.Graph {
	g := graph.NewGraph()

	// Create 1000 nodes
	for i := int64(1); i <= 1000; i++ {
		g.AddNode(&graph.Node{
			ID:  i,
			Lat: 13.0 + float64(i)*0.001,
			Lon: 100.0 + float64(i)*0.001,
		})
	}

	// Create ~2000 edges
	for i := int64(1); i < 1000; i++ {
		g.AddEdge(graph.Edge{
			From:     i,
			To:       i + 1,
			Weight:   100.0 + float64(i),
			OSMWayID: i,
			MaxSpeed: 30.0,
			Tags: map[string]string{
				"highway": "primary",
				"name":    "Benchmark Road",
			},
		})
		g.AddEdge(graph.Edge{
			From:     i + 1,
			To:       i,
			Weight:   100.0 + float64(i),
			OSMWayID: i,
			MaxSpeed: 30.0,
			Tags: map[string]string{
				"highway": "primary",
			},
		})
	}

	return g
}


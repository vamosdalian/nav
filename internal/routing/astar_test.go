package routing

import (
	"testing"

	"github.com/vamosdalian/nav/internal/graph"
)

// BenchmarkAStar benchmarks the A* algorithm
func BenchmarkAStar(b *testing.B) {
	g := createTestGraph()
	router := NewRouter(g)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = router.FindRoute(43.73, 7.42, 43.74, 7.43)
	}
}

// BenchmarkAStarWithProfile benchmarks routing with different profiles
func BenchmarkAStarWithProfile(b *testing.B) {
	g := createTestGraph()
	router := NewRouter(g)
	
	profiles := []struct {
		name    string
		profile RoutingProfile
	}{
		{"car", CarProfile},
		{"bike", BikeProfile},
		{"foot", FootProfile},
	}
	
	for _, p := range profiles {
		b.Run(p.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = router.FindRouteWithProfile(43.73, 7.42, 43.74, 7.43, p.profile)
			}
		})
	}
}

// BenchmarkMultipleRoutes benchmarks alternative route finding
func BenchmarkMultipleRoutes(b *testing.B) {
	g := createTestGraph()
	router := NewRouter(g)
	
	alternatives := []int{1, 2, 3}
	
	for _, alt := range alternatives {
		b.Run(string(rune('0'+alt))+"_routes", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = router.FindMultipleRoutes(43.73, 7.42, 43.74, 7.43, alt)
			}
		})
	}
}

// createTestGraph creates a simple test graph for benchmarking
func createTestGraph() *graph.Graph {
	g := graph.NewGraph()
	
	// Create a small test network
	// This is a simplified version - in real benchmarks, use actual data
	nodes := []struct {
		id  int64
		lat float64
		lon float64
	}{
		{1, 43.73, 7.42},
		{2, 43.735, 7.425},
		{3, 43.74, 7.43},
	}
	
	for _, n := range nodes {
		g.AddNode(&graph.Node{ID: n.id, Lat: n.lat, Lon: n.lon})
	}
	
	// Add edges
	g.AddEdge(graph.Edge{
		From:     1,
		To:       2,
		Weight:   1000,
		OSMWayID: 100,
		Tags:     map[string]string{"highway": "primary"},
	})
	
	g.AddEdge(graph.Edge{
		From:     2,
		To:       3,
		Weight:   1000,
		OSMWayID: 101,
		Tags:     map[string]string{"highway": "primary"},
	})
	
	return g
}


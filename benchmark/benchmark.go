package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/vamosdalian/nav/internal/graph"
	"github.com/vamosdalian/nav/internal/osm"
	"github.com/vamosdalian/nav/internal/routing"
	"github.com/vamosdalian/nav/internal/storage"
)

// BenchmarkResult stores benchmark results
type BenchmarkResult struct {
	Name           string
	Iterations     int
	TotalTime      time.Duration
	AvgTime        time.Duration
	MinTime        time.Duration
	MaxTime        time.Duration
	SuccessCount   int
	FailureCount   int
	NodesExplored  int
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  Navigation Service - Performance Benchmark")
	fmt.Println("========================================\n")
	
	// Load or parse graph
	g := loadGraph()
	if g == nil {
		log.Fatal("Failed to load graph")
	}
	
	fmt.Printf("Graph loaded: %d nodes, %d edges\n\n", g.NodeCount(), g.EdgeCount())
	
	// Run benchmarks
	runAllBenchmarks(g)
}

func loadGraph() *graph.Graph {
	// Try to load from cache first
	if _, err := os.Stat("graph.bin.gz"); err == nil {
		fmt.Println("Loading graph from cache...")
		store := storage.NewStorage("graph.bin.gz")
		g, err := store.Load()
		if err == nil {
			return g
		}
		fmt.Printf("Failed to load cache: %v\n", err)
	}
	
	// Parse from OSM if cache not available
	osmPath := os.Getenv("OSM_DATA_PATH")
	if osmPath == "" {
		osmPath = "monaco-latest.osm.pbf"
	}
	
	fmt.Printf("Parsing OSM data from %s...\n", osmPath)
	g := graph.NewGraph()
	parser := osm.NewParser(g)
	
	if err := parser.ParseFile(osmPath); err != nil {
		log.Printf("Failed to parse OSM: %v", err)
		return nil
	}
	
	return g
}

func runAllBenchmarks(g *graph.Graph) {
	testCases := []struct {
		name                  string
		fromLat, fromLon      float64
		toLat, toLon          float64
		profile               string
	}{
		{"Short Distance - Car", 43.73, 7.42, 43.735, 7.425, "car"},
		{"Medium Distance - Car", 43.73, 7.42, 43.74, 7.43, "car"},
		{"Short Distance - Bike", 43.73, 7.42, 43.735, 7.425, "bike"},
		{"Medium Distance - Bike", 43.73, 7.42, 43.74, 7.43, "bike"},
		{"Short Distance - Foot", 43.73, 7.42, 43.735, 7.425, "foot"},
		{"Medium Distance - Foot", 43.73, 7.42, 43.74, 7.43, "foot"},
	}
	
	fmt.Println("Running benchmarks...\n")
	fmt.Println("Test Case                        | Iterations | Avg Time | Min Time | Max Time | Success")
	fmt.Println("--------------------------------|------------|----------|----------|----------|--------")
	
	for _, tc := range testCases {
		result := benchmarkRoute(g, tc.name, tc.fromLat, tc.fromLon, tc.toLat, tc.toLon, tc.profile, 100)
		printResult(result)
	}
	
	// Benchmark multiple routes
	fmt.Println("\nAlternative Routes Benchmarks:\n")
	fmt.Println("Test Case                        | Iterations | Avg Time | Success")
	fmt.Println("--------------------------------|------------|----------|--------")
	
	for _, alt := range []int{1, 2, 3} {
		name := fmt.Sprintf("Car - %d alternatives", alt)
		result := benchmarkMultipleRoutes(g, name, 43.73, 7.42, 43.74, 7.43, alt, 50)
		fmt.Printf("%-32s | %10d | %8.2fms | %d/%d\n",
			result.Name, result.Iterations, 
			float64(result.AvgTime.Microseconds())/1000.0,
			result.SuccessCount, result.Iterations)
	}
	
	fmt.Println("\n========================================")
	fmt.Println("  Benchmark Complete!")
	fmt.Println("========================================")
}

func benchmarkRoute(g *graph.Graph, name string, fromLat, fromLon, toLat, toLon float64, profileName string, iterations int) BenchmarkResult {
	router := NewRouter(g)
	profile := routing.GetProfile(profileName)
	
	result := BenchmarkResult{
		Name:       name,
		Iterations: iterations,
		MinTime:    time.Hour,
		MaxTime:    0,
	}
	
	var totalTime time.Duration
	
	for i := 0; i < iterations; i++ {
		start := time.Now()
		_, err := router.FindRouteWithProfile(fromLat, fromLon, toLat, toLon, profile)
		elapsed := time.Since(start)
		
		totalTime += elapsed
		
		if elapsed < result.MinTime {
			result.MinTime = elapsed
		}
		if elapsed > result.MaxTime {
			result.MaxTime = elapsed
		}
		
		if err == nil {
			result.SuccessCount++
		} else {
			result.FailureCount++
		}
	}
	
	result.TotalTime = totalTime
	result.AvgTime = totalTime / time.Duration(iterations)
	
	return result
}

func benchmarkMultipleRoutes(g *graph.Graph, name string, fromLat, fromLon, toLat, toLon float64, numRoutes, iterations int) BenchmarkResult {
	router := NewRouter(g)
	
	result := BenchmarkResult{
		Name:       name,
		Iterations: iterations,
	}
	
	var totalTime time.Duration
	
	for i := 0; i < iterations; i++ {
		start := time.Now()
		_, err := router.FindMultipleRoutes(fromLat, fromLon, toLat, toLon, numRoutes)
		elapsed := time.Since(start)
		
		totalTime += elapsed
		
		if err == nil {
			result.SuccessCount++
		} else {
			result.FailureCount++
		}
	}
	
	result.TotalTime = totalTime
	result.AvgTime = totalTime / time.Duration(iterations)
	
	return result
}

func printResult(r BenchmarkResult) {
	avgMs := float64(r.AvgTime.Microseconds()) / 1000.0
	minMs := float64(r.MinTime.Microseconds()) / 1000.0
	maxMs := float64(r.MaxTime.Microseconds()) / 1000.0
	
	fmt.Printf("%-32s | %10d | %7.2fms | %7.2fms | %7.2fms | %d/%d\n",
		r.Name, r.Iterations, avgMs, minMs, maxMs, r.SuccessCount, r.Iterations)
}


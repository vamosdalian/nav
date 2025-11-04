package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/vamosdalian/nav/internal/api"
	"github.com/vamosdalian/nav/internal/config"
	"github.com/vamosdalian/nav/internal/graph"
	"github.com/vamosdalian/nav/internal/osm"
	"github.com/vamosdalian/nav/internal/routing"
	"github.com/vamosdalian/nav/internal/storage"
)

func main() {
	parseOnly := flag.Bool("parse-only", false, "Parse OSM data, save graph, and exit without starting server")
	flag.Parse()

	log.Println("Starting Navigation Service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize graph
	var g *graph.Graph

	// Try to load existing graph data first
	if cfg.GraphDataPath != "" {
		if _, err := os.Stat(cfg.GraphDataPath); err == nil {
			log.Printf("Loading graph from %s...", cfg.GraphDataPath)
			store := storage.NewStorage(cfg.GraphDataPath)
			g, err = store.Load()
			if err != nil {
				log.Printf("Failed to load graph: %v, will parse OSM data", err)
			} else {
				log.Printf("Graph loaded: %d nodes, %d edges", g.NodeCount(), g.EdgeCount())
			}
		}
	}

	// If graph not loaded and OSM data available, parse it
	if g == nil && cfg.OSMDataPath != "" {
		log.Printf("Parsing OSM data from %s...", cfg.OSMDataPath)
		g = graph.NewGraph()
		parser := osm.NewParser(g)

		if err := parser.ParseFile(cfg.OSMDataPath); err != nil {
			log.Fatalf("Failed to parse OSM data: %v", err)
		}

		log.Printf("Graph built: %d nodes, %d edges", g.NodeCount(), g.EdgeCount())

		// Save parsed graph for future use
		if cfg.GraphDataPath != "" {
			log.Printf("Saving graph to %s...", cfg.GraphDataPath)
			store := storage.NewStorage(cfg.GraphDataPath)
			if err := store.Save(g); err != nil {
				log.Printf("Warning: Failed to save graph: %v", err)
			} else {
				log.Printf("Graph saved successfully")
				if *parseOnly {
					log.Println("Parse-only mode: exiting without starting server")
					return
				}
			}
		} else if *parseOnly {
			log.Fatal("Parse-only mode requires GRAPH_DATA_PATH to be set")
		}
	}

	if g == nil {
		log.Fatal("No graph data available. Set OSM_DATA_PATH or GRAPH_DATA_PATH")
	}

	if *parseOnly {
		log.Println("Parse-only mode: graph already loaded, exiting without starting server")
		return
	}

	// Initialize router
	router := routing.NewRouter(g)

	// Initialize API server
	apiServer := api.NewServer(router, g)
	handler := apiServer.SetupRoutes()

	// Start HTTP server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Server listening on %s", addr)
	log.Printf("API Endpoints:")
	log.Printf("  POST /route - Find route between two points")
	log.Printf("  GET  /route/get - Find route (query params)")
	log.Printf("  POST /weight/update - Update edge weights")
	log.Printf("  GET  /health - Health check")

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

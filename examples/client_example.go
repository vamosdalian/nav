package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Example Go client for the Navigation Service API

const baseURL = "http://localhost:8080"

type RouteRequest struct {
	FromLat      float64 `json:"from_lat"`
	FromLon      float64 `json:"from_lon"`
	ToLat        float64 `json:"to_lat"`
	ToLon        float64 `json:"to_lon"`
	Alternatives int     `json:"alternatives,omitempty"`
}

type RouteResponse struct {
	Routes []RouteInfo `json:"routes"`
	Code   string      `json:"code"`
}

type RouteInfo struct {
	Distance float64     `json:"distance"`
	Duration float64     `json:"duration"`
	Geometry [][2]float64 `json:"geometry"`
}

type UpdateWeightRequest struct {
	OSMWayID   int64   `json:"osm_way_id"`
	Multiplier float64 `json:"multiplier"`
}

type UpdateWeightResponse struct {
	Code         string `json:"code"`
	EdgesUpdated int    `json:"edges_updated"`
}

func main() {
	// Example 1: Find a route
	route, err := findRoute(43.73, 7.42, 43.74, 7.43, 0)
	if err != nil {
		fmt.Printf("Error finding route: %v\n", err)
	} else {
		fmt.Printf("Route found:\n")
		fmt.Printf("  Distance: %.2f meters\n", route.Routes[0].Distance)
		fmt.Printf("  Duration: %.2f seconds\n", route.Routes[0].Duration)
		fmt.Printf("  Points: %d\n", len(route.Routes[0].Geometry))
	}

	// Example 2: Find alternative routes
	routes, err := findRoute(43.73, 7.42, 43.74, 7.43, 2)
	if err != nil {
		fmt.Printf("Error finding routes: %v\n", err)
	} else {
		fmt.Printf("\nFound %d alternative routes:\n", len(routes.Routes))
		for i, r := range routes.Routes {
			fmt.Printf("  Route %d: %.2fm, %.2fs\n", i+1, r.Distance, r.Duration)
		}
	}

	// Example 3: Update road weights
	updated, err := updateWeight(123456789, 2.0)
	if err != nil {
		fmt.Printf("Error updating weights: %v\n", err)
	} else {
		fmt.Printf("\nUpdated %d edges\n", updated.EdgesUpdated)
	}
}

func findRoute(fromLat, fromLon, toLat, toLon float64, alternatives int) (*RouteResponse, error) {
	req := RouteRequest{
		FromLat:      fromLat,
		FromLon:      fromLon,
		ToLat:        toLat,
		ToLon:        toLon,
		Alternatives: alternatives,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(baseURL+"/route", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result RouteResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func updateWeight(osmWayID int64, multiplier float64) (*UpdateWeightResponse, error) {
	req := UpdateWeightRequest{
		OSMWayID:   osmWayID,
		Multiplier: multiplier,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(baseURL+"/weight/update", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result UpdateWeightResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}


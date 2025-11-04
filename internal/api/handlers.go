package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/vamosdalian/nav/internal/encoding"
	"github.com/vamosdalian/nav/internal/graph"
	"github.com/vamosdalian/nav/internal/routing"
)

// Server holds the HTTP server dependencies
type Server struct {
	router *routing.Router
	graph  *graph.Graph
}

// NewServer creates a new API server
func NewServer(r *routing.Router, g *graph.Graph) *Server {
	return &Server{
		router: r,
		graph:  g,
	}
}

// RouteRequest represents a routing request
type RouteRequest struct {
	FromLat        float64 `json:"from_lat"`
	FromLon        float64 `json:"from_lon"`
	ToLat          float64 `json:"to_lat"`
	ToLon          float64 `json:"to_lon"`
	Alternatives   int     `json:"alternatives,omitempty"`
	Format         string  `json:"format,omitempty"`         // "geojson" (default) or "polyline"
	Profile        string  `json:"profile,omitempty"`        // "car" (default), "bike", or "foot"
	Unidirectional bool    `json:"unidirectional,omitempty"` // Force unidirectional A* (default: false, uses bidirectional)
}

// RouteResponse represents a routing response
type RouteResponse struct {
	Routes []RouteInfo `json:"routes"`
	Code   string      `json:"code"`
	Format string      `json:"format,omitempty"` // Format used for geometry
}

// RouteInfo contains route details
type RouteInfo struct {
	Distance float64     `json:"distance"`
	Duration float64     `json:"duration"`
	Geometry interface{} `json:"geometry"` // Can be [][2]float64, string (polyline), or GeoJSON
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// HandleRoute handles route requests
func (s *Server) HandleRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only POST method is allowed")
		return
	}

	var req RouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON request")
		return
	}

	// Validate coordinates
	if !s.validateCoordinates(req.FromLat, req.FromLon) || !s.validateCoordinates(req.ToLat, req.ToLon) {
		s.sendError(w, http.StatusBadRequest, "invalid_coordinates", "Invalid coordinates")
		return
	}

	// Get routing profile
	profile := routing.GetProfile(req.Profile)

	// Find routes with the specified profile
	var routes []*routing.Route
	var err error

	if req.Alternatives > 0 {
		// Temporarily set profile for multiple routes
		s.router.SetProfile(profile)
		routes, err = s.router.FindMultipleRoutes(req.FromLat, req.FromLon, req.ToLat, req.ToLon, req.Alternatives)
		s.router.SetProfile(routing.CarProfile) // Reset to default
	} else {
		var route *routing.Route
		var routeErr error

		// Default to bidirectional A* (11x faster), unless explicitly disabled
		if req.Unidirectional {
			route, routeErr = s.router.FindRouteWithProfile(req.FromLat, req.FromLon, req.ToLat, req.ToLon, profile)
		} else {
			route, routeErr = s.router.FindRouteBidirectionalWithProfile(req.FromLat, req.FromLon, req.ToLat, req.ToLon, profile)
		}

		if routeErr == nil {
			routes = []*routing.Route{route}
		}
		err = routeErr
	}

	if err != nil {
		s.sendError(w, http.StatusNotFound, "no_route", err.Error())
		return
	}

	// Determine output format (default: geojson)
	format := req.Format
	if format == "" {
		format = "geojson"
	}

	// Build response
	response := RouteResponse{
		Code:   "Ok",
		Format: format,
		Routes: make([]RouteInfo, len(routes)),
	}

	for i, route := range routes {
		coordinates := make([][2]float64, len(route.Nodes))
		for j, nodeID := range route.Nodes {
			node, _ := s.graph.GetNode(nodeID)
			coordinates[j] = [2]float64{node.Lon, node.Lat}
		}

		var geometry interface{}
		switch format {
		case "polyline":
			geometry = encoding.EncodePolyline(coordinates)
		default: // "geojson" or empty
			geometry = encoding.NewLineStringGeometry(coordinates)
		}

		response.Routes[i] = RouteInfo{
			Distance: route.Distance,
			Duration: route.Duration,
			Geometry: geometry,
		}
	}

	s.sendJSON(w, http.StatusOK, response)
}

// HandleRouteGet handles GET-based route requests (OSRM-compatible)
func (s *Server) HandleRouteGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET method is allowed")
		return
	}

	// Parse query parameters
	fromLat, err1 := strconv.ParseFloat(r.URL.Query().Get("from_lat"), 64)
	fromLon, err2 := strconv.ParseFloat(r.URL.Query().Get("from_lon"), 64)
	toLat, err3 := strconv.ParseFloat(r.URL.Query().Get("to_lat"), 64)
	toLon, err4 := strconv.ParseFloat(r.URL.Query().Get("to_lon"), 64)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		s.sendError(w, http.StatusBadRequest, "invalid_parameters", "Invalid coordinate parameters")
		return
	}

	alternatives := 0
	if alt := r.URL.Query().Get("alternatives"); alt != "" {
		alternatives, _ = strconv.Atoi(alt)
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "geojson"
	}

	profile := r.URL.Query().Get("profile")

	// Use the same logic as POST handler
	req := RouteRequest{
		FromLat:      fromLat,
		FromLon:      fromLon,
		ToLat:        toLat,
		ToLon:        toLon,
		Alternatives: alternatives,
		Format:       format,
		Profile:      profile,
	}

	// Reuse route finding logic
	if !s.validateCoordinates(req.FromLat, req.FromLon) || !s.validateCoordinates(req.ToLat, req.ToLon) {
		s.sendError(w, http.StatusBadRequest, "invalid_coordinates", "Invalid coordinates")
		return
	}

	// Get routing profile
	profileObj := routing.GetProfile(req.Profile)

	var routes []*routing.Route
	var err error

	if req.Alternatives > 0 {
		s.router.SetProfile(profileObj)
		routes, err = s.router.FindMultipleRoutes(req.FromLat, req.FromLon, req.ToLat, req.ToLon, req.Alternatives)
		s.router.SetProfile(routing.CarProfile) // Reset to default
	} else {
		route, routeErr := s.router.FindRouteWithProfile(req.FromLat, req.FromLon, req.ToLat, req.ToLon, profileObj)
		if routeErr == nil {
			routes = []*routing.Route{route}
		}
		err = routeErr
	}

	if err != nil {
		s.sendError(w, http.StatusNotFound, "no_route", err.Error())
		return
	}

	response := RouteResponse{
		Code:   "Ok",
		Format: req.Format,
		Routes: make([]RouteInfo, len(routes)),
	}

	for i, route := range routes {
		coordinates := make([][2]float64, len(route.Nodes))
		for j, nodeID := range route.Nodes {
			node, _ := s.graph.GetNode(nodeID)
			coordinates[j] = [2]float64{node.Lon, node.Lat}
		}

		var geometry interface{}
		switch req.Format {
		case "polyline":
			geometry = encoding.EncodePolyline(coordinates)
		default: // "geojson" or empty
			geometry = encoding.NewLineStringGeometry(coordinates)
		}

		response.Routes[i] = RouteInfo{
			Distance: route.Distance,
			Duration: route.Duration,
			Geometry: geometry,
		}
	}

	s.sendJSON(w, http.StatusOK, response)
}

// UpdateWeightRequest represents a weight update request
type UpdateWeightRequest struct {
	OSMWayID   int64   `json:"osm_way_id"`
	Multiplier float64 `json:"multiplier"`
}

// UpdateWeightResponse represents update response
type UpdateWeightResponse struct {
	Code         string `json:"code"`
	EdgesUpdated int    `json:"edges_updated"`
}

// HandleUpdateWeight handles edge weight updates
func (s *Server) HandleUpdateWeight(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only POST method is allowed")
		return
	}

	var req UpdateWeightRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON request")
		return
	}

	if req.Multiplier <= 0 {
		s.sendError(w, http.StatusBadRequest, "invalid_multiplier", "Multiplier must be positive")
		return
	}

	count := s.graph.UpdateEdgeWeightByWay(req.OSMWayID, req.Multiplier)

	s.sendJSON(w, http.StatusOK, UpdateWeightResponse{
		Code:         "Ok",
		EdgesUpdated: count,
	})
}

// HandleHealth handles health check requests
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status": "healthy",
		"nodes":  s.graph.NodeCount(),
		"edges":  s.graph.EdgeCount(),
	}
	s.sendJSON(w, http.StatusOK, health)
}

func (s *Server) validateCoordinates(lat, lon float64) bool {
	return lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180
}

func (s *Server) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (s *Server) sendError(w http.ResponseWriter, status int, code, message string) {
	s.sendJSON(w, status, ErrorResponse{
		Code:    code,
		Message: message,
	})
}

// SetupRoutes configures all API routes
func (s *Server) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/route", s.HandleRoute)
	mux.HandleFunc("/route/get", s.HandleRouteGet)
	mux.HandleFunc("/weight/update", s.HandleUpdateWeight)
	mux.HandleFunc("/health", s.HandleHealth)

	// Add CORS and logging middleware
	return s.loggingMiddleware(s.corsMiddleware(mux))
}

// CORS middleware
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logging middleware
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

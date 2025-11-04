package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/vamosdalian/nav/internal/encoding"
	"github.com/vamosdalian/nav/internal/graph"
	"github.com/vamosdalian/nav/internal/routing"
)

// Server holds the HTTP server dependencies
type Server struct {
	router         *routing.Router
	graph          *graph.Graph
	profileManager *routing.ProfileManager
}

// NewServer creates a new API server
func NewServer(r *routing.Router, g *graph.Graph, pm *routing.ProfileManager) *Server {
	return &Server{
		router:         r,
		graph:          g,
		profileManager: pm,
	}
}

// RouteRequest represents a routing request (flat structure for GET/POST compatibility)
type RouteRequest struct {
	FromLat        float64 `json:"from_lat"`
	FromLon        float64 `json:"from_lon"`
	ToLat          float64 `json:"to_lat"`
	ToLon          float64 `json:"to_lon"`
	Alternatives   int     `json:"alternatives,omitempty"`
	Format         string  `json:"format,omitempty"`         // "geojson" (default) or "polyline"
	Profile        string  `json:"profile,omitempty"`        // Profile name (e.g., "car")
	Unidirectional bool    `json:"unidirectional,omitempty"` // Force unidirectional A* (default: false)

	// Runtime overrides (flat structure for GET query params)
	AvoidTolls    *bool    `json:"avoid_tolls,omitempty"`
	AvoidHighways *bool    `json:"avoid_highways,omitempty"`
	AvoidFerries  *bool    `json:"avoid_ferries,omitempty"`
	AvoidTunnels  *bool    `json:"avoid_tunnels,omitempty"`
	AllowUturns   *bool    `json:"allow_uturns,omitempty"`
	MaxSpeed      *float64 `json:"max_speed,omitempty"` // km/h
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

// HandleRoute handles route requests (supports both GET and POST)
func (s *Server) HandleRoute(w http.ResponseWriter, r *http.Request) {
	var req RouteRequest
	var err error

	switch r.Method {
	case http.MethodPost:
		// Parse JSON body
		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.sendError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON request")
			return
		}

	case http.MethodGet:
		// Parse query parameters
		req, err = s.parseRouteQueryParams(r)
		if err != nil {
			s.sendError(w, http.StatusBadRequest, "invalid_parameters", err.Error())
			return
		}

	default:
		s.sendError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET and POST methods are allowed")
		return
	}

	// Validate coordinates
	if !s.validateCoordinates(req.FromLat, req.FromLon) || !s.validateCoordinates(req.ToLat, req.ToLon) {
		s.sendError(w, http.StatusBadRequest, "invalid_coordinates", "Invalid coordinates")
		return
	}

	// Get effective routing profile
	effectiveProfile, err := s.getEffectiveProfile(&req)
	if err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid_profile", err.Error())
		return
	}

	// Find routes with the specified profile
	routes, err := s.findRoutes(req, effectiveProfile)
	if err != nil {
		s.sendError(w, http.StatusNotFound, "no_route", err.Error())
		return
	}

	// Build and send response
	s.sendRouteResponse(w, routes, req.Format)
}

// HandleListProfiles handles listing all available profiles
func (s *Server) HandleListProfiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET method is allowed")
		return
	}

	profiles := s.profileManager.ListProfiles()
	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"code":     "Ok",
		"profiles": profiles,
		"count":    len(profiles),
	})
}

// HandleGetProfile handles getting a specific profile details
func (s *Server) HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET method is allowed")
		return
	}

	// Extract profile name from URL path
	// Expecting /profiles/{name}
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 2 {
		s.sendError(w, http.StatusBadRequest, "invalid_path", "Profile name is required")
		return
	}

	profileName := pathParts[1]
	profile, err := s.profileManager.GetProfile(profileName)
	if err != nil {
		s.sendError(w, http.StatusNotFound, "profile_not_found", err.Error())
		return
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"code":    "Ok",
		"profile": profile,
	})
}

// HandleReloadProfiles handles reloading all profiles from disk
func (s *Server) HandleReloadProfiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only POST method is allowed")
		return
	}

	if err := s.profileManager.Reload(); err != nil {
		s.sendError(w, http.StatusInternalServerError, "reload_failed", err.Error())
		return
	}

	profiles := s.profileManager.ListProfiles()
	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"code":     "Ok",
		"message":  "Profiles reloaded successfully",
		"profiles": profiles,
		"count":    len(profiles),
	})
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

// parseRouteQueryParams parses GET request query parameters into RouteRequest
func (s *Server) parseRouteQueryParams(r *http.Request) (RouteRequest, error) {
	q := r.URL.Query()
	req := RouteRequest{}

	// Required parameters
	var err error
	req.FromLat, err = strconv.ParseFloat(q.Get("from_lat"), 64)
	if err != nil {
		return req, fmt.Errorf("invalid from_lat")
	}

	req.FromLon, err = strconv.ParseFloat(q.Get("from_lon"), 64)
	if err != nil {
		return req, fmt.Errorf("invalid from_lon")
	}

	req.ToLat, err = strconv.ParseFloat(q.Get("to_lat"), 64)
	if err != nil {
		return req, fmt.Errorf("invalid to_lat")
	}

	req.ToLon, err = strconv.ParseFloat(q.Get("to_lon"), 64)
	if err != nil {
		return req, fmt.Errorf("invalid to_lon")
	}

	// Optional parameters
	if alt := q.Get("alternatives"); alt != "" {
		req.Alternatives, _ = strconv.Atoi(alt)
	}

	req.Format = q.Get("format")
	req.Profile = q.Get("profile")

	if uni := q.Get("unidirectional"); uni != "" {
		req.Unidirectional, _ = strconv.ParseBool(uni)
	}

	// Runtime overrides
	if val := q.Get("avoid_tolls"); val != "" {
		b, _ := strconv.ParseBool(val)
		req.AvoidTolls = &b
	}

	if val := q.Get("avoid_highways"); val != "" {
		b, _ := strconv.ParseBool(val)
		req.AvoidHighways = &b
	}

	if val := q.Get("avoid_ferries"); val != "" {
		b, _ := strconv.ParseBool(val)
		req.AvoidFerries = &b
	}

	if val := q.Get("avoid_tunnels"); val != "" {
		b, _ := strconv.ParseBool(val)
		req.AvoidTunnels = &b
	}

	if val := q.Get("allow_uturns"); val != "" {
		b, _ := strconv.ParseBool(val)
		req.AllowUturns = &b
	}

	if val := q.Get("max_speed"); val != "" {
		f, _ := strconv.ParseFloat(val, 64)
		req.MaxSpeed = &f
	}

	return req, nil
}

// getEffectiveProfile loads a profile and applies runtime options
func (s *Server) getEffectiveProfile(req *RouteRequest) (*routing.ProfileConfig, error) {
	profileName := req.Profile

	// Use first available profile if not specified
	if profileName == "" {
		profiles := s.profileManager.ListProfiles()
		if len(profiles) == 0 {
			return nil, fmt.Errorf("no profiles available")
		}
		profileName = profiles[0]
	}

	// Get base profile
	baseProfile, err := s.profileManager.GetProfile(profileName)
	if err != nil {
		return nil, fmt.Errorf("profile '%s' not found: %w", profileName, err)
	}

	// Build RouteOptions from request
	options := &routing.RouteOptions{
		AvoidTolls:    req.AvoidTolls,
		AvoidHighways: req.AvoidHighways,
		AvoidFerries:  req.AvoidFerries,
		AvoidTunnels:  req.AvoidTunnels,
		AllowUturns:   req.AllowUturns,
		MaxSpeed:      req.MaxSpeed,
	}

	// Apply runtime options if any are set
	return routing.GetEffectiveProfile(baseProfile, options), nil
}

// findRoutes finds routes using the effective profile
func (s *Server) findRoutes(req RouteRequest, profile *routing.ProfileConfig) ([]*routing.Route, error) {
	// Temporary bridge: Convert new ProfileConfig to old RoutingProfile
	// This allows us to use the existing Router implementation
	// TODO: Update Router to work directly with ProfileConfig
	oldProfile := s.convertToOldProfile(profile)

	var routes []*routing.Route
	var err error

	if req.Alternatives > 0 {
		// Set profile temporarily for multiple routes
		s.router.SetProfile(oldProfile)
		routes, err = s.router.FindMultipleRoutes(req.FromLat, req.FromLon, req.ToLat, req.ToLon, req.Alternatives)
	} else {
		var route *routing.Route
		var routeErr error

		// Default to bidirectional A* (faster), unless explicitly disabled
		if req.Unidirectional {
			route, routeErr = s.router.FindRouteWithProfile(req.FromLat, req.FromLon, req.ToLat, req.ToLon, oldProfile)
		} else {
			route, routeErr = s.router.FindRouteBidirectionalWithProfile(req.FromLat, req.FromLon, req.ToLat, req.ToLon, oldProfile)
		}

		if routeErr == nil {
			routes = []*routing.Route{route}
		}
		err = routeErr
	}

	return routes, err
}

// convertToOldProfile converts new ProfileConfig to old RoutingProfile (temporary bridge)
func (s *Server) convertToOldProfile(config *routing.ProfileConfig) routing.RoutingProfile {
	// Build allowed highways map
	allowedHighways := make(map[string]bool)
	speedFactors := make(map[string]float64)

	for hwType, hwConfig := range config.Highways {
		allowedHighways[hwType] = hwConfig.Allowed
		speedFactors[hwType] = hwConfig.SpeedFactor
	}

	// Build avoid surfaces map
	avoidSurfaces := make(map[string]bool)
	for surfaceType, surfaceConfig := range config.Surfaces {
		// Consider surfaces with penalty > 2.0 as "avoided"
		if surfaceConfig.Penalty > 2.0 {
			avoidSurfaces[surfaceType] = true
		}
	}

	return routing.RoutingProfile{
		Name:            config.Name,
		AllowedHighways: allowedHighways,
		SpeedFactors:    speedFactors,
		AvoidSurfaces:   avoidSurfaces,
		MaxSpeed:        config.Settings.MaxSpeedKmh / 3.6, // Convert km/h to m/s
	}
}

// sendRouteResponse builds and sends the route response
func (s *Server) sendRouteResponse(w http.ResponseWriter, routes []*routing.Route, format string) {
	// Determine output format (default: geojson)
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

func (s *Server) validateCoordinates(lat, lon float64) bool {
	return lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180
}

// splitPath splits URL path into parts
func splitPath(path string) []string {
	parts := []string{}
	for _, p := range strings.Split(path, "/") {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
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

	// Route endpoints
	mux.HandleFunc("/route", s.HandleRoute) // Supports both GET and POST

	// Profile endpoints
	mux.HandleFunc("/profiles", s.profileHandler)              // GET list, or specific profile
	mux.HandleFunc("/profiles/reload", s.HandleReloadProfiles) // POST reload

	// Utility endpoints
	mux.HandleFunc("/weight/update", s.HandleUpdateWeight)
	mux.HandleFunc("/health", s.HandleHealth)

	// Add CORS and logging middleware
	return s.loggingMiddleware(s.corsMiddleware(mux))
}

// profileHandler routes profile requests to appropriate handler
func (s *Server) profileHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := splitPath(r.URL.Path)

	if len(pathParts) == 1 {
		// /profiles - list all profiles
		s.HandleListProfiles(w, r)
	} else if len(pathParts) == 2 {
		// /profiles/{name} - get specific profile
		s.HandleGetProfile(w, r)
	} else {
		s.sendError(w, http.StatusNotFound, "not_found", "Invalid profile endpoint")
	}
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

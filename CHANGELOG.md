# Changelog

All notable changes to this project will be documented in this file.

## [1.3.0] - 2025-11-04

### Added - Performance Edition
- **Bidirectional A*** - 11x performance boost for routing queries
  - Searches from both start and end simultaneously
  - Average query time: 1.5ms (vs 16ms unidirectional)
  - Reduces node exploration by 80-90%
- **Reverse Adjacency List** - O(1) backward edge lookup
  - Enables efficient bidirectional search
  - Memory increase: ~30% (worth the 11x speedup)
- **Comprehensive Benchmarking** - Built-in performance testing
  - Benchmark tool: `cmd/benchmark/main.go`
  - Go benchmarks: `internal/routing/astar_test.go`
  - Detailed results: `BENCHMARK_RESULTS.md`
- **API Parameter** - `bidirectional` flag for opt-in faster routing

### Changed
- **Default Algorithm**: Bidirectional A* is now the default (11x faster)
  - All queries use bidirectional search unless `unidirectional: true` is set
  - Improves default query time from 16ms to 1.5ms
- Graph structure now includes reverse edges for bidirectional search
- Graph serialization includes reverse adjacency list
- API parameter changed from `bidirectional` to `unidirectional` (opt-out instead of opt-in)

### Files Added
- `internal/routing/bidirectional.go` - Bidirectional A* implementation
- `internal/routing/astar_test.go` - Go benchmark tests
- `cmd/benchmark/main.go` - Comprehensive benchmark tool
- `docs/BIDIRECTIONAL_ASTAR.md` - Algorithm documentation
- `BENCHMARK_RESULTS.md` - Performance test results

### Performance Impact
- **Query Speed**: 11x faster with bidirectional A*
- **Memory**: +30% for reverse edges
- **Throughput**: 690 queries/second (single core)

## [1.2.0] - 2025-11-04

### Added - Major Features
- **Routing Profiles**: Support for car, bike, and pedestrian routing
  - Car profile: Optimized for motorways and main roads
  - Bike profile: Prefers cycleways, avoids motorways
  - Foot profile: Can use footways, steps, and pedestrian areas
- **Turn Restrictions**: Automatic parsing and enforcement of OSM turn restrictions
  - Supports: no_left_turn, no_right_turn, no_u_turn, only_straight_on, etc.
  - Monaco dataset: 44 turn restrictions parsed
- **Enhanced Oneway Support**: Complete handling of oneway streets
  - Forward oneway: oneway=yes, oneway=1
  - Reverse oneway: oneway=-1, oneway=reverse
- **Profile-based Weight Calculation**: Dynamic weight adjustment per profile

### Changed
- A* algorithm now tracks previous way ID for turn restriction checks
- Graph structure includes turn restrictions storage
- OSM parser processes relations for turn restrictions
- Edge creation logic improved for reverse oneways

### Files Added
- `internal/routing/profile.go` - Routing profile system with 3 presets
- `internal/graph/restrictions.go` - Turn restriction data structures
- `docs/PROFILES_GUIDE.md` - Complete profiles and restrictions guide
- `docs/FEATURES_IMPLEMENTED.md` - New features summary
- `docs/ROADMAP_PRIORITY.md` - Roadmap prioritization analysis
- `examples/test_profiles.sh` - Profile testing script

### Files Modified
- `internal/graph/graph.go` - Added restrictions field
- `internal/graph/serialization.go` - Serialize restrictions
- `internal/osm/parser.go` - Parse relations and enhanced oneway handling
- `internal/routing/astar.go` - Profile filtering and turn restriction checks
- `internal/api/handlers.go` - Added profile parameter support
- `README.md` - Updated with new features documentation

### API Changes
- **Request**: Added optional `profile` field (`car`, `bike`, `foot`)
- **Default Profile**: Car profile when not specified
- **Backward Compatible**: Existing API calls work without changes

### Performance Impact
- Route query time: +4-6ms (turn restriction checks + profile filtering)
- Memory usage: +32 bytes per turn restriction (~1.4KB for Monaco)
- Parsing time: Unchanged

## [1.1.0] - 2025-11-04

### Added
- **Multiple Geometry Formats**: Added support for two geometry output formats:
  - `geojson`: Standard GeoJSON LineString format (default)
  - `polyline`: Google Polyline encoded format (50-70% smaller)
- **Format Parameter**: Added `format` parameter to both POST and GET route endpoints
- **Polyline Encoding/Decoding**: Implemented Google Polyline algorithm
- **GeoJSON Support**: Full GeoJSON LineString geometry support
- **Documentation**: Added comprehensive geometry formats documentation

### Changed
- Response structure now includes `format` field indicating the geometry format used
- Geometry field in route response is now `interface{}` to support multiple formats

### Files Added
- `internal/encoding/polyline.go` - Polyline encoding/decoding implementation
- `internal/encoding/geojson.go` - GeoJSON geometry structures
- `docs/GEOMETRY_FORMATS.md` - Complete geometry formats documentation
- `examples/geometry_formats.sh` - Test script for all formats
- `CHANGELOG.md` - This file

### API Changes
- **Request**: Added optional `format` field to route requests
- **Response**: Added `format` field to route responses
- **Backward Compatible**: Default format is GeoJSON, maintaining backward compatibility

## [1.0.0] - 2025-11-04

### Initial Release

#### Core Features
- A* pathfinding algorithm with Haversine heuristic
- Multiple alternative routes using penalty-based method
- Dynamic edge weight modification
- OSM PBF data parsing
- Graph caching for fast startup
- Thread-safe concurrent operations

#### API Endpoints
- `POST /route` - Find routes with JSON body
- `GET /route/get` - Find routes with query parameters
- `POST /weight/update` - Update edge weights
- `GET /health` - Health check

#### Components
- Graph engine with adjacency list
- OSM parser with road filtering
- A* routing with closed-set optimization
- HTTP API with CORS and logging
- Graph serialization with Gob + Gzip

#### Documentation
- README.md - Full project documentation
- QUICKSTART.md - 5-minute getting started guide
- ARCHITECTURE.md - Technical deep dive
- TESTING.md - Testing procedures
- PROJECT_SUMMARY.md - Project completion summary

#### Examples
- Bash API examples
- Python client example
- Go client example

#### Deployment
- Dockerfile for containerization
- docker-compose.yml for easy deployment
- Makefile for build automation

---

## Version History

- **v1.3.0** (2025-11-04): Performance Edition - Bidirectional A* with 11x speedup
- **v1.2.0** (2025-11-04): Feature Edition - Routing profiles and turn restrictions
- **v1.1.0** (2025-11-04): Format Edition - Multiple geometry formats support
- **v1.0.0** (2025-11-04): Initial Release - Core navigation features


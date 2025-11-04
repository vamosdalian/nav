# Changelog

All notable changes to this project will be documented in this file.

## [1.1.0] - 2025-11-04

### Added
- **Multiple Geometry Formats**: Added support for three geometry output formats:
  - `geojson`: Standard GeoJSON LineString format (default)
  - `polyline`: Google Polyline encoded format (50-70% smaller)
  - `simple`: Plain coordinate array format
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

- **v1.1.0** (2025-11-04): Added multiple geometry formats support
- **v1.0.0** (2025-11-04): Initial release with core navigation features


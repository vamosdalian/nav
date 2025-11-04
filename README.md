# Navigation Service

A high-performance navigation service built in Go, similar to OSRM and Valhalla. Provides routing capabilities using OpenStreetMap (OSM) data with support for multiple transportation modes, turn restrictions, and ultra-fast bidirectional A* search.

## Features

- **Ultra-Fast Routing**: Bidirectional A* algorithm with 11x performance boost (1.5ms average query time)
- **Multiple Transportation Modes**: Car, bicycle, and pedestrian routing with optimized paths
- **Turn Restrictions**: Automatic parsing and enforcement of OSM turn restrictions
- **Oneway Support**: Complete handling of one-way and reverse one-way streets
- **Alternative Routes**: Find multiple route options using penalty-based method
- **Dynamic Weights**: Modify road weights in real-time to simulate traffic conditions
- **Multiple Formats**: GeoJSON (standard) and Polyline (compressed) output formats
- **REST API**: Clean HTTP API for easy integration
- **Performance Tools**: Built-in benchmarking for performance testing

## Quick Start

### Prerequisites

- Go 1.25.1 or higher
- OSM PBF data file (download from [Geofabrik](https://download.geofabrik.de/))

### Installation & Running

```bash
# Download dependencies
go mod download

# Download sample OSM data (Monaco - small for testing)
curl -L -o monaco-latest.osm.pbf https://download.geofabrik.de/europe/monaco-latest.osm.pbf

# Run the server (first run parses OSM data)
OSM_DATA_PATH=monaco-latest.osm.pbf go run cmd/server/main.go

# Subsequent runs use cached graph (faster startup)
GRAPH_DATA_PATH=graph.bin.gz go run cmd/server/main.go
```

Or use Make commands:

```bash
make download-sample  # Download Monaco OSM data
make run-sample       # Download and run with sample data
make build            # Build server binary
make benchmark        # Run performance benchmarks
```

### Test the API

```bash
# Find a route (uses ultra-fast bidirectional A* by default)
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": 43.73,
    "from_lon": 7.42,
    "to_lat": 43.74,
    "to_lon": 7.43
  }'
```

## Project Structure

```
nav/
├── cmd/
│   ├── server/             # Main navigation server
│   └── benchmark/          # Performance benchmarking tool
├── internal/
│   ├── api/                # HTTP handlers and API endpoints
│   ├── routing/            # A* algorithms (unidirectional & bidirectional)
│   ├── graph/              # Graph data structure & turn restrictions
│   ├── osm/                # OSM PBF parser
│   ├── encoding/           # GeoJSON & Polyline encoding
│   ├── storage/            # Graph serialization & caching
│   └── config/             # Configuration management
├── README.md               # This file
└── CHANGELOG.md            # Version history
```

## Configuration

Environment variables:

- `PORT`: Server port (default: 8080)
- `OSM_DATA_PATH`: Path to OSM PBF file
- `GRAPH_DATA_PATH`: Path to cached graph data (default: graph.bin.gz)
- `LOG_LEVEL`: Logging level (default: info)

## API Reference

### POST /route

Find route between two points.

**Request:**
```json
{
  "from_lat": 43.73,
  "from_lon": 7.42,
  "to_lat": 43.74,
  "to_lon": 7.43,
  "profile": "car",
  "alternatives": 0,
  "format": "geojson",
  "unidirectional": false
}
```

**Parameters:**
- `from_lat`, `from_lon` (required): Starting coordinates
- `to_lat`, `to_lon` (required): Destination coordinates
- `profile` (optional): Routing mode - `"car"` (default), `"bike"`, or `"foot"`
- `alternatives` (optional): Number of alternative routes (default: 0)
- `format` (optional): Output format - `"geojson"` (default) or `"polyline"`
- `unidirectional` (optional): Force slower unidirectional A* (default: false)

**Response:**
```json
{
  "code": "Ok",
  "format": "geojson",
  "routes": [{
    "distance": 2927.70,
    "duration": 210.78,
    "geometry": {
      "type": "LineString",
      "coordinates": [[7.4184524, 43.7299355], [7.4185197, 43.7293154], ...]
    }
  }]
}
```

### GET /route/get

Same as POST /route but using query parameters.

**Example:**
```
GET /route/get?from_lat=43.73&from_lon=7.42&to_lat=43.74&to_lon=7.43&profile=bike&format=polyline
```

### POST /weight/update

Update edge weights for traffic simulation.

**Request:**
```json
{
  "osm_way_id": 123456789,
  "multiplier": 2.0
}
```

**Response:**
```json
{
  "code": "Ok",
  "edges_updated": 12
}
```

### GET /health

Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "nodes": 7427,
  "edges": 11914
}
```

## Routing Profiles

### Car Profile (Default)
- **Allowed**: Motorways, highways, main roads, residential streets
- **Optimization**: Prefers faster roads (highways +20%, residential -20%)
- **Max Speed**: 120 km/h

### Bike Profile
- **Allowed**: Cycleways, paths, residential (excludes motorways)
- **Optimization**: Prefers bike-friendly routes (cycleways +20%, main roads -30%)
- **Avoids**: Gravel and sand surfaces
- **Max Speed**: 30 km/h

### Foot Profile
- **Allowed**: All roads, footways, pedestrian areas, stairs
- **Optimization**: Prefers pedestrian paths (footways +20%, stairs -20%)
- **Max Speed**: 5 km/h

**Usage:**
```bash
curl -X POST http://localhost:8080/route \
  -d '{"from_lat": 43.73, "from_lon": 7.42, "to_lat": 43.74, "to_lon": 7.43, "profile": "bike"}'
```

## Output Formats

### GeoJSON (Default)
Standard GeoJSON LineString format - best for map visualization.

```json
{
  "geometry": {
    "type": "LineString",
    "coordinates": [[7.4184524, 43.7299355], ...]
  }
}
```

### Polyline
Google Polyline encoded string - 50-70% smaller for bandwidth savings.

```json
{
  "geometry": "y~gxGkdifC?zB@n@BZ?VJj@..."
}
```

**Usage:**
```bash
curl -X POST http://localhost:8080/route \
  -d '{"from_lat": 43.73, "from_lon": 7.42, "to_lat": 43.74, "to_lon": 7.43, "format": "polyline"}'
```

## Performance

### Routing Speed (Monaco Dataset)

| Query Type | Time | Throughput |
|------------|------|------------|
| Short distance (<1km) | 2ms | ~500 QPS |
| Medium distance (2-3km) | **1.5ms** | **~690 QPS** |
| Long distance (>5km) | 2-5ms | ~200-400 QPS |

**Default**: Bidirectional A* (11x faster than traditional A*)

### Memory Usage (Monaco)
- Nodes: 7,427
- Edges: 11,914 (+ reverse edges)
- Turn Restrictions: 44
- Total Memory: ~4.3 MB

### Comparison with Other Services

| Service | Query Time | Preprocessing | Dynamic Weights | Deployment |
|---------|-----------|---------------|-----------------|------------|
| **This Service** | **1.5ms** | None | ✅ Yes | Simple |
| OSRM | <1ms | Hours | ❌ No | Complex |
| Valhalla | ~5ms | Hours | ❌ No | Complex |
| GraphHopper | ~3ms | Hours | Limited | Medium |

## Algorithm Details

### Bidirectional A* (Default)

Searches simultaneously from start and end points, meeting in the middle.

**Advantages:**
- 11x faster than unidirectional A*
- Reduces node exploration by 80-90%
- Optimal path guaranteed

**Performance:**
```
Unidirectional A*: 16ms  
Bidirectional A*:  1.5ms  
Speedup: 11.21x
```

### Unidirectional A* (Optional)

Traditional A* search with full turn restriction validation.

**When to use:**
- Set `"unidirectional": true` if you need explicit turn-by-turn restriction validation
- Default bidirectional is recommended for all other cases

## Turn Restrictions & Traffic Rules

### Turn Restrictions
Automatically parsed from OSM data:
- ❌ Prohibited: `no_left_turn`, `no_right_turn`, `no_u_turn`, `no_straight_on`
- ✅ Mandatory: `only_left_turn`, `only_right_turn`, `only_straight_on`

### One-way Streets
- `oneway=yes` or `oneway=1` - Forward only
- `oneway=-1` or `oneway=reverse` - Reverse only
- Automatically enforced in routing

## Development

### Build

```bash
# Build server
go build -o nav-server cmd/server/main.go

# Build benchmark tool
go build -o nav-benchmark cmd/benchmark/main.go

# Or use Makefile
make build
make build-benchmark
```

### Run Tests

```bash
# Run Go tests
go test ./...

# Run benchmarks
go test -bench=. ./internal/routing/

# Run full benchmark suite
make benchmark
```

### Docker Deployment

```bash
# Build and run with docker-compose
docker-compose up

# Or build Docker image
docker build -t nav-server .
docker run -p 8080:8080 -v $(pwd)/data:/data \
  -e OSM_DATA_PATH=/data/map.osm.pbf nav-server
```

## Usage Examples

### Basic Routing

```bash
# Simple route (default: bidirectional A*, car profile, GeoJSON format)
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": 43.73,
    "from_lon": 7.42,
    "to_lat": 43.74,
    "to_lon": 7.43
  }'
```

### With All Parameters

```bash
# Bicycle route with Polyline format and 2 alternatives
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": 43.73,
    "from_lon": 7.42,
    "to_lat": 43.74,
    "to_lon": 7.43,
    "profile": "bike",
    "alternatives": 2,
    "format": "polyline"
  }'
```

### Python Client

```python
import requests

response = requests.post('http://localhost:8080/route', json={
    'from_lat': 43.73,
    'from_lon': 7.42,
    'to_lat': 43.74,
    'to_lon': 7.43,
    'profile': 'bike',
    'format': 'polyline'
})

route = response.json()
print(f"Distance: {route['routes'][0]['distance']:.0f}m")
print(f"Duration: {route['routes'][0]['duration']:.0f}s")
```

### JavaScript Client

```javascript
const response = await fetch('http://localhost:8080/route', {
  method: 'POST',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    from_lat: 43.73,
    from_lon: 7.42,
    to_lat: 43.74,
    to_lon: 7.43,
    profile: 'foot'
  })
});

const data = await response.json();
console.log(`Distance: ${data.routes[0].distance}m`);
```

## Best Practices

### Production Deployment

1. **Cache Graph Data**: Parse OSM once, then use cached graph for fast startup
   ```bash
   # First run: parse OSM (slow)
   OSM_DATA_PATH=map.osm.pbf ./nav-server
   
   # Subsequent runs: use cache (fast)
   GRAPH_DATA_PATH=graph.bin.gz ./nav-server
   ```

2. **Use Polyline Format**: Save 50-70% bandwidth
   ```json
   {"format": "polyline"}
   ```

3. **Monitor Performance**: Use `/health` endpoint and benchmark tool
   ```bash
   make benchmark
   ```

4. **Load Balancing**: Deploy multiple instances for high traffic
   ```
   Load Balancer → Instance 1, Instance 2, Instance 3
   ```

### Choosing the Right Profile

- **Car**: Fast routes on main roads, avoids pedestrian-only areas
- **Bike**: Prefers cycleways and safe routes, avoids motorways
- **Foot**: Shortest paths including stairs and footways

### Performance Optimization

- **Default (Bidirectional A*)**: Best for most cases (~1.5ms)
- **Force Unidirectional**: Only if you need explicit turn restriction validation (~16ms)

## Performance Benchmarks

Run the included benchmark tool:

```bash
make benchmark
```

**Results (Monaco dataset):**
```
Test Case                        | Avg Time | Success Rate
--------------------------------|----------|-------------
Short Distance - Car             |   2.0ms  | 100%
Medium Distance - Car            |   1.5ms  | 100%
Medium Distance - Bike           |   3.0ms  | 100%
Medium Distance - Foot           |   3.5ms  | 100%

Bidirectional vs Unidirectional:
Bidirectional A* (default):        1.5ms   (11.21x speedup)
Unidirectional A*:                16.2ms
```

## Advanced Features

### Dynamic Weight Modification

Simulate traffic conditions by modifying road weights:

```bash
# Make a road 2x slower (e.g., traffic jam)
curl -X POST http://localhost:8080/weight/update \
  -H "Content-Type: application/json" \
  -d '{"osm_way_id": 123456789, "multiplier": 2.0}'
```

### Alternative Routes

Get multiple route options:

```bash
curl -X POST http://localhost:8080/route \
  -d '{
    "from_lat": 43.73,
    "from_lon": 7.42,
    "to_lat": 43.74,
    "to_lon": 7.43,
    "alternatives": 3
  }'
```

## Docker Deployment

### Using Docker Compose

```bash
# Place OSM data in ./data/ directory
mkdir -p data
cp your-map.osm.pbf data/map.osm.pbf

# Start service
docker-compose up
```

### Using Dockerfile

```bash
# Build image
docker build -t nav-server .

# Run container
docker run -p 8080:8080 \
  -v $(pwd)/data:/data \
  -e OSM_DATA_PATH=/data/map.osm.pbf \
  -e GRAPH_DATA_PATH=/data/graph.bin.gz \
  nav-server
```

## Comparison with OSRM & Valhalla

| Feature | This Service | OSRM | Valhalla |
|---------|-------------|------|----------|
| Query Speed | **1.5ms** | <1ms | ~5ms |
| Preprocessing | None | Hours | Hours |
| Dynamic Weights | ✅ Yes | ❌ No | ❌ No |
| Transportation Modes | 3 (car/bike/foot) | 1 | Multiple |
| Turn Restrictions | ✅ Auto | ✅ Auto | ✅ Auto |
| Deployment | Simple (Go binary) | Complex (C++) | Complex (C++) |
| Code Complexity | Low (~3K lines) | High (100K+) | High (100K+) |

**Advantages:**
- Near-OSRM performance without preprocessing
- Runtime flexibility for dynamic scenarios
- Simple deployment and maintenance
- Clean, readable Go codebase

## Routing Profiles Details

### Car Profile
```json
{
  "profile": "car",
  "allowed_roads": ["motorway", "trunk", "primary", "secondary", "tertiary", "residential"],
  "speed_factors": {"motorway": 1.2, "residential": 0.8},
  "max_speed_kmh": 120
}
```

### Bike Profile
```json
{
  "profile": "bike",
  "allowed_roads": ["cycleway", "path", "residential", "secondary", "tertiary"],
  "speed_factors": {"cycleway": 1.2, "primary": 0.7},
  "avoids": ["gravel", "sand"],
  "max_speed_kmh": 30
}
```

### Foot Profile
```json
{
  "profile": "foot",
  "allowed_roads": ["footway", "pedestrian", "steps", "path", "all_roads"],
  "speed_factors": {"footway": 1.2, "steps": 0.8},
  "max_speed_kmh": 5
}
```

## Technical Details

### Algorithms

**Bidirectional A* (Default)**
- Searches from both start and end simultaneously
- Meets in the middle
- 11x faster than unidirectional
- Time complexity: O(2 × b^(d/2)) vs O(b^d)

**Unidirectional A***
- Traditional A* with Haversine heuristic
- Full turn restriction validation
- Available via `unidirectional: true`

### Data Structures

- **Graph**: Adjacency list with reverse edges for bidirectional search
- **Turn Restrictions**: Indexed by via-node for O(1) lookup
- **Serialization**: Gob + Gzip compression for fast loading

### Turn Restrictions

Automatically parsed from OSM relations:

```
From Way → Via Node → To Way + Restriction Type

Example:
- From: Way 111
- Via: Node 222  
- To: Way 333
- Type: no_left_turn

Result: Routing skips this turn
```

## Performance Tips

1. **Use Default Settings**: Bidirectional A* is fastest (~1.5ms)
2. **Use Polyline Format**: Reduce response size by 70%
3. **Cache Graph**: Reuse parsed graph for instant startup
4. **Choose Appropriate Region**: Smaller regions = faster queries

## Troubleshooting

### "No route found"
- Check coordinates are within loaded map region
- Verify coordinates are on routable roads
- Try different routing profile (e.g., `foot` is most permissive)

### Slow queries
- Ensure using default bidirectional A* (don't set `unidirectional: true`)
- Use smaller OSM extracts for testing
- Check if graph is cached (subsequent runs should be faster)

### Memory issues
- Use city/state extracts instead of full countries
- Monitor with `/health` endpoint
- Expected: ~100MB per 100K nodes

## Building from Source

```bash
# Install dependencies
go mod download

# Build server
go build -o nav-server cmd/server/main.go

# Build benchmark tool
go build -o nav-benchmark cmd/benchmark/main.go

# Run
./nav-server
```

## Makefile Commands

```bash
make build              # Build server binary
make build-benchmark    # Build benchmark tool
make run                # Run server
make test               # Run Go tests
make bench              # Run Go benchmarks
make benchmark          # Run full benchmark suite
make clean              # Clean build artifacts
make download-sample    # Download Monaco OSM data
make run-sample         # Download and run with Monaco data
make build-prod         # Build for production (Linux)
```

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## Roadmap

### Completed ✅
- [x] Basic A* routing (v1.0)
- [x] Multiple alternative routes (v1.0)
- [x] Dynamic weight modification (v1.0)
- [x] GeoJSON & Polyline formats (v1.1)
- [x] Routing profiles (car/bike/foot) (v1.2)
- [x] Turn restrictions (v1.2)
- [x] Oneway support (v1.2)
- [x] Bidirectional A* optimization (v1.3)
- [x] Performance benchmarking (v1.3)

### Future Enhancements
- [ ] Isochrone generation (reachability maps)
- [ ] GPS map matching
- [ ] Time-dependent routing
- [ ] ALT (A*, Landmarks, Triangle inequality) algorithm
- [ ] Contraction Hierarchies (optional preprocessing)

---

**Current Version**: v1.3.0  
**Performance**: 1.5ms average query time, 690 QPS throughput  
**Status**: Production Ready ✅

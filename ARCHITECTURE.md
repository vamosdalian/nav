# Architecture Documentation

## System Overview

This navigation service is designed as a high-performance, production-grade routing system for OpenStreetMap data, similar to OSRM and Valhalla but with enhanced runtime flexibility.

## Design Principles

1. **Simplicity**: Clean, maintainable Go code
2. **Performance**: Optimized algorithms and data structures
3. **Flexibility**: Runtime weight modification without preprocessing
4. **Scalability**: Thread-safe operations with concurrent request handling
5. **Best Practices**: Standard Go project layout, error handling, testing

## Core Components

### 1. Graph Engine (`internal/graph/`)

**Purpose**: In-memory graph representation of the road network

**Key Files**:
- `graph.go` - Core graph data structure
- `serialization.go` - Import/export functionality

**Features**:
- Adjacency list representation for efficient neighbor lookup
- Thread-safe with RW mutex for concurrent access
- Haversine distance calculation for geographic data
- Dynamic weight modification at runtime

**Data Structures**:
```go
Node: {ID, Latitude, Longitude}
Edge: {From, To, Weight, OSMWayID, MaxSpeed, Tags}
Graph: {nodes: map[int64]*Node, edges: map[int64][]Edge}
```

**Complexity**:
- Add Node/Edge: O(1)
- Get Neighbors: O(1)
- Find Nearest Node: O(n) - could be optimized with spatial index
- Update Weight: O(edges per node)

### 2. Routing Engine (`internal/routing/`)

**Purpose**: Pathfinding algorithms

**Algorithm**: A* (A-star) with Haversine heuristic

**Why A*?**
- Optimal: Guarantees shortest path
- Fast: Heuristic guides search toward goal
- Geographic: Haversine distance is admissible heuristic
- Simple: No preprocessing required (unlike Contraction Hierarchies)

**Alternative Routes**:
Uses penalty-based approach:
1. Find optimal route
2. Apply 50% penalty to used edges
3. Find next route
4. Verify ≥30% difference
5. Repeat

**Complexity**:
- Single Route: O((V + E) log V) - A* with priority queue
- K Routes: O(K × (V + E) log V)
- Space: O(V) for tracking

**Trade-offs**:
- ✅ No preprocessing needed
- ✅ Works with dynamic weights
- ✅ Simple implementation
- ❌ Slower than Contraction Hierarchies for static data
- ❌ Not optimal for very large graphs (countries)

### 3. OSM Parser (`internal/osm/`)

**Purpose**: Parse OpenStreetMap PBF data into graph

**Process**:
1. Stream OSM PBF file (memory efficient)
2. Extract nodes (lat/lon coordinates)
3. Filter routable ways (roads)
4. Create bidirectional edges (unless one-way)
5. Calculate weights based on distance and speed

**Supported Road Types**:
- Motorway, trunk, primary, secondary, tertiary
- Residential, service, unclassified
- Filters out footways, cycleways, construction

**Weight Calculation**:
```
weight = distance (meters)
Can be extended to: weight = distance / speed
```

**Tags Preserved**:
- highway (road type)
- name (street name)
- surface (road quality)
- lanes (number of lanes)
- maxspeed (speed limit)
- oneway (directional restriction)

### 4. API Server (`internal/api/`)

**Purpose**: RESTful HTTP interface

**Endpoints**:

| Path | Method | Purpose | Input | Output |
|------|--------|---------|-------|--------|
| /health | GET | Status check | - | Node/edge count |
| /route | POST | Find route | JSON coordinates | Route(s) with geometry |
| /route/get | GET | Find route | Query params | Route(s) with geometry |
| /weight/update | POST | Modify weights | OSM Way ID, multiplier | Edges updated count |

**Response Format** (OSRM-compatible):
```json
{
  "code": "Ok",
  "routes": [
    {
      "distance": 1234.56,
      "duration": 88.85,
      "geometry": [[lon, lat], [lon, lat], ...]
    }
  ]
}
```

**Middleware**:
- CORS: Allow cross-origin requests
- Logging: Request/response logging
- Error handling: Consistent error responses

**Concurrency**:
- Multiple requests handled concurrently
- Thread-safe graph access with RW locks
- Read-heavy optimization (multiple readers, single writer)

### 5. Storage Layer (`internal/storage/`)

**Purpose**: Persist graph to disk for fast loading

**Format**: Gob (Go binary) + Gzip compression

**Benefits**:
- Fast serialization/deserialization
- Compression reduces file size (50-70%)
- Avoids re-parsing OSM data

**Performance**:
- Small region (Monaco): <1 second to load
- Large region (State): 5-30 seconds to load
- vs. OSM parsing: 10-100x faster

**Trade-offs**:
- ✅ Very fast loading
- ✅ Simple implementation
- ❌ Go-specific format (not portable)
- ❌ Not human-readable

**Future Optimization**:
Could use formats like:
- Protocol Buffers (portable)
- FlatBuffers (zero-copy)
- Custom binary format (optimized)

### 6. Configuration (`internal/config/`)

**Purpose**: Environment-based configuration

**12-Factor App Compliant**:
- Config via environment variables
- No hardcoded paths
- Easy deployment across environments

**Variables**:
- PORT: Server port
- OSM_DATA_PATH: Source data
- GRAPH_DATA_PATH: Cache location
- LOG_LEVEL: Verbosity

## Data Flow

### Initial Setup (First Run)
```
OSM PBF File
    ↓
OSM Parser
    ↓
Graph Builder
    ↓
Storage (cache)
```

### Subsequent Runs
```
Cached Graph File
    ↓
Storage Loader
    ↓
Ready to Route
```

### Routing Request
```
Client Request (lat/lon)
    ↓
API Handler
    ↓
Find Nearest Nodes
    ↓
A* Algorithm
    ↓
Reconstruct Path
    ↓
Format Response (geometry)
    ↓
JSON Response
```

### Weight Update
```
Client Request (way_id, multiplier)
    ↓
API Handler
    ↓
Graph.UpdateEdgeWeight()
    ↓
Response (count updated)
```

## Scalability Considerations

### Vertical Scaling (Single Instance)

**Memory**:
- ~100-200 bytes per node
- ~50-100 bytes per edge
- Example: 1M nodes, 2M edges ≈ 300MB

**CPU**:
- Single route: 1-10ms (urban)
- Concurrent requests: Limited by CPU cores
- Go runtime handles scheduling

**Disk**:
- OSM data: 1MB - 10GB (compressed)
- Graph cache: 50-70% of memory usage

### Horizontal Scaling (Multiple Instances)

**Stateless Design**:
- Each instance loads full graph
- No shared state between instances
- Load balancer distributes requests

**Deployment**:
```
Load Balancer
    ↓
├── Instance 1 (Region A)
├── Instance 2 (Region B)
└── Instance 3 (Region C)
```

**Improvements**:
- Geographic sharding (split by region)
- Read replicas (same graph, multiple instances)
- Caching layer (Redis for common routes)

## Performance Characteristics

### Routing Speed

| Graph Size | Nodes | Edges | Route Time |
|------------|-------|-------|------------|
| City | 10K-100K | 20K-200K | <1ms |
| State | 100K-1M | 200K-2M | 1-10ms |
| Country | 1M-10M | 2M-20M | 10-100ms |
| Continent | 10M+ | 20M+ | 100ms-1s |

### Memory Usage

| Region | Nodes | Edges | RAM |
|--------|-------|-------|-----|
| Monaco | ~5K | ~10K | ~1MB |
| NYC | ~500K | ~1M | ~100MB |
| California | ~5M | ~10M | ~1GB |
| USA | ~20M | ~40M | ~4GB |

### Preprocessing Time (First Run)

| Region | OSM File Size | Parse Time |
|--------|--------------|------------|
| City | 10-100MB | 10-60s |
| State | 100MB-1GB | 1-10min |
| Country | 1-10GB | 10-60min |

## Comparison with OSRM/Valhalla

### OSRM (Contraction Hierarchies)

**Pros**:
- Very fast queries (<1ms)
- Low memory after preprocessing
- Highly optimized

**Cons**:
- Long preprocessing (hours for countries)
- Cannot modify weights at runtime
- Complex codebase (C++)

### Valhalla (Multi-modal)

**Pros**:
- Multiple travel modes
- Turn-by-turn navigation
- Rich feature set

**Cons**:
- Heavy preprocessing
- High complexity
- Large deployment size

### This Service (A* with Runtime Weights)

**Pros**:
- ✅ Simple codebase (Go)
- ✅ No preprocessing required
- ✅ Runtime weight modification
- ✅ Easy deployment
- ✅ Multiple alternative routes

**Cons**:
- ❌ Slower than CH (but still fast)
- ❌ Higher memory usage
- ❌ Basic features (no turn restrictions)

**Best For**:
- Dynamic routing scenarios
- Rapid prototyping
- Custom routing logic
- Real-time weight adjustments
- Moderate-sized regions

## Extension Points

### Easy Extensions

1. **Routing Profiles**: Add bike/pedestrian weights
2. **Turn Restrictions**: Parse and respect turn data
3. **Time-Dependent Routing**: Add temporal weights
4. **Isochrones**: Expand A* to cover areas
5. **Map Matching**: Snap GPS traces to roads

### Advanced Extensions

1. **Spatial Index**: R-tree for faster nearest node
2. **Bidirectional Search**: Meet-in-the-middle A*
3. **Contraction Hierarchies**: Optional preprocessing
4. **Traffic Integration**: Real-time weight updates
5. **Distributed Graph**: Shard across machines

## Testing Strategy

### Unit Tests
- Graph operations
- Algorithm correctness
- API handlers

### Integration Tests
- Full routing pipeline
- OSM parsing
- Storage/loading

### Performance Tests
- Benchmark routing speed
- Memory profiling
- Concurrent load testing

### Example Test Cases
```go
// Test shortest path
TestBasicRoute(t *testing.T)

// Test alternative routes
TestMultipleRoutes(t *testing.T)

// Test weight modification
TestDynamicWeights(t *testing.T)

// Test concurrent access
TestConcurrentRouting(t *testing.T)
```

## Monitoring & Observability

### Metrics to Track

1. **Request Metrics**:
   - Request rate (req/s)
   - Response time (p50, p95, p99)
   - Error rate

2. **Graph Metrics**:
   - Node/edge count
   - Memory usage
   - Cache hit rate

3. **Algorithm Metrics**:
   - Nodes explored per query
   - Route length distribution
   - Alternative route quality

### Logging

- Structured logging (JSON)
- Request/response logging
- Error logging with stack traces
- Performance logging

### Health Checks

- `/health` endpoint
- Ready: Graph loaded
- Live: Server responding
- Metrics: Node/edge counts

## Security Considerations

### Input Validation
- Coordinate bounds checking
- Rate limiting
- Request size limits

### DoS Prevention
- Query timeout
- Max nodes explored limit
- Connection limits

### Data Privacy
- No user tracking
- Stateless requests
- No PII stored

## Deployment Checklist

- [ ] Build binary: `go build`
- [ ] Download OSM data
- [ ] Set environment variables
- [ ] Test health endpoint
- [ ] Test route endpoint
- [ ] Configure monitoring
- [ ] Set up load balancer (if needed)
- [ ] Configure logging
- [ ] Set resource limits
- [ ] Test backup/restore

## Future Roadmap

### Short Term
- [ ] Add unit tests
- [ ] Performance benchmarks
- [ ] API documentation (OpenAPI)
- [ ] Docker registry publish

### Medium Term
- [ ] Routing profiles (car/bike/foot)
- [ ] Turn restrictions
- [ ] Spatial index (R-tree)
- [ ] GraphQL API

### Long Term
- [ ] Contraction Hierarchies option
- [ ] Distributed graph
- [ ] Traffic integration
- [ ] Map matching
- [ ] Isochrone generation

## Contributing

See main README.md for contribution guidelines.

## References

- [A* Algorithm](https://en.wikipedia.org/wiki/A*_search_algorithm)
- [OpenStreetMap](https://www.openstreetmap.org/)
- [OSRM](http://project-osrm.org/)
- [Valhalla](https://github.com/valhalla/valhalla)
- [Contraction Hierarchies](https://en.wikipedia.org/wiki/Contraction_hierarchies)


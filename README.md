# Navigation Service

A high-performance navigation service built in Go, similar to OSRM and Valhalla. This service provides routing capabilities using OpenStreetMap (OSM) data with support for multiple alternative routes and dynamic edge weight modification.

## Features

- **Fast Routing**: A* algorithm with Haversine heuristic for optimal geographic routing
- **Multiple Routes**: Find alternative routes using penalty-based method
- **Dynamic Weights**: Modify road weights in real-time to simulate traffic conditions
- **OSM Support**: Parse and process OpenStreetMap PBF data
- **REST API**: Clean HTTP API for easy integration
- **Efficient Storage**: Graph serialization for fast startup times

## Architecture

```
nav/
├── cmd/
│   └── server/          # Main server application
├── internal/
│   ├── api/            # HTTP handlers and routes
│   ├── config/         # Configuration management
│   ├── graph/          # Graph data structure
│   ├── osm/            # OSM data parser
│   ├── routing/        # Routing algorithms (A*)
│   └── storage/        # Graph persistence
└── go.mod
```

## Quick Start

### Prerequisites

- Go 1.25.1 or higher
- OSM PBF data file (download from [Geofabrik](https://download.geofabrik.de/))

### Installation

```bash
# Clone the repository
cd /Users/lmc10232/project/nav

# Download dependencies
go mod download

# Download OSM data (example: Monaco - small dataset for testing)
wget https://download.geofabrik.de/europe/monaco-latest.osm.pbf
```

### Running the Service

```bash
# Parse OSM data on first run
OSM_DATA_PATH=monaco-latest.osm.pbf go run cmd/server/main.go

# Subsequent runs can use cached graph data
GRAPH_DATA_PATH=graph.bin.gz go run cmd/server/main.go

# Or specify both (will try to load graph, fallback to parsing OSM)
OSM_DATA_PATH=monaco-latest.osm.pbf GRAPH_DATA_PATH=graph.bin.gz go run cmd/server/main.go
```

### Configuration

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
  "from_lat": 43.7384,
  "from_lon": 7.4246,
  "to_lat": 43.7312,
  "to_lon": 7.4197,
  "alternatives": 2
}
```

**Response:**
```json
{
  "code": "Ok",
  "routes": [
    {
      "distance": 1234.56,
      "duration": 88.85,
      "geometry": [
        [7.4246, 43.7384],
        [7.4240, 43.7380],
        [7.4197, 43.7312]
      ]
    }
  ]
}
```

### GET /route/get

Find route using query parameters.

**Example:**
```
GET /route/get?from_lat=43.7384&from_lon=7.4246&to_lat=43.7312&to_lon=7.4197&alternatives=1
```

### POST /weight/update

Update edge weights for a specific OSM way (to simulate traffic).

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
  "nodes": 15234,
  "edges": 32156
}
```

## Algorithm Details

### A* Routing

The service uses the A* algorithm with the following characteristics:

- **Heuristic**: Haversine distance (great-circle distance)
- **Edge Weights**: Based on road distance and speed limits
- **Optimality**: Guarantees shortest path when heuristic is admissible

### Alternative Routes

Alternative routes are found using a penalty-based method:

1. Find the optimal route using A*
2. Apply 50% penalty to edges used in the route
3. Find next route with penalized edges
4. Ensure sufficient difference (< 70% overlap)
5. Repeat for requested number of alternatives

### Weight Modification

Road weights can be modified in real-time:

- **By Edge**: Update specific from/to node pair
- **By OSM Way**: Update all edges belonging to a way
- **Use Cases**: Traffic simulation, road closures, temporary restrictions

## Performance Considerations

- **Graph Loading**: First run parses OSM data (slow), subsequent runs load cached graph (fast)
- **Memory Usage**: Approximately 100-200 bytes per node, 50-100 bytes per edge
- **Routing Speed**: ~1-10ms for typical urban routes (depends on graph size)
- **Concurrency**: Thread-safe operations with read/write locks

## Best Practices

1. **Cache Graph Data**: Always save parsed graph for faster startup
2. **Use Appropriate OSM Extracts**: Download specific regions from Geofabrik
3. **Monitor Memory**: Large regions (e.g., entire countries) require significant RAM
4. **Load Balancing**: Deploy multiple instances behind a load balancer for high traffic
5. **Health Monitoring**: Use `/health` endpoint for monitoring and readiness checks

## Development

### Build

```bash
go build -o nav-server cmd/server/main.go
```

### Run Tests

```bash
go test ./...
```

### Docker (optional)

```dockerfile
FROM golang:1.25.1-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o nav-server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/nav-server .
EXPOSE 8080
CMD ["./nav-server"]
```

## Example Usage

```bash
# Find a route
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": 43.7384,
    "from_lon": 7.4246,
    "to_lat": 43.7312,
    "to_lon": 7.4197,
    "alternatives": 2
  }'

# Update weights (simulate heavy traffic)
curl -X POST http://localhost:8080/weight/update \
  -H "Content-Type: application/json" \
  -d '{
    "osm_way_id": 123456789,
    "multiplier": 3.0
  }'

# Health check
curl http://localhost:8080/health
```

## Comparison with OSRM/Valhalla

| Feature | This Service | OSRM | Valhalla |
|---------|-------------|------|----------|
| Algorithm | A* | Contraction Hierarchies | Multi-modal |
| Language | Go | C++ | C++ |
| Setup | Simple | Complex | Complex |
| Weight Modification | Runtime | Preprocessing required | Preprocessing required |
| Memory Usage | Moderate | Low (after CH) | Moderate |
| Query Speed | Fast | Very Fast | Fast |

## License

MIT

## Contributing

Contributions welcome! Please open an issue or submit a pull request.

## Roadmap

- [ ] Add support for turn restrictions
- [ ] Implement Dijkstra rank for bi-directional search
- [ ] Add routing profiles (car, bike, pedestrian)
- [ ] Support for time-dependent routing
- [ ] Add isochrone generation
- [ ] Implement map matching
- [ ] Add GraphQL API
- [ ] Performance benchmarks


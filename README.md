# Navigation Service

A high-performance navigation service built in Go, similar to OSRM and Valhalla. This service provides routing capabilities using OpenStreetMap (OSM) data with support for multiple alternative routes and dynamic edge weight modification.

## Features

- **Fast Routing**: A* algorithm with Haversine heuristic for optimal geographic routing
- **Multiple Routes**: Find alternative routes using penalty-based method
- **Routing Profiles**: Car, bike, and pedestrian routing with optimized weights
- **Turn Restrictions**: Automatic parsing and enforcement of OSM turn restrictions
- **Oneway Support**: Complete oneway and reverse oneway handling
- **Dynamic Weights**: Modify road weights in real-time to simulate traffic conditions
- **Multiple Formats**: GeoJSON and Polyline formats
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
  "from_lat": 43.73,
  "from_lon": 7.42,
  "to_lat": 43.74,
  "to_lon": 7.43,
  "alternatives": 2,
  "format": "geojson"
}
```

**Parameters:**
- `from_lat`, `from_lon`: Starting coordinates
- `to_lat`, `to_lon`: Destination coordinates  
- `alternatives` (optional): Number of alternative routes
- `format` (optional): Geometry format - `"geojson"` (default) or `"polyline"`
- `profile` (optional): Routing profile - `"car"` (default), `"bike"`, or `"foot"`

**Response (GeoJSON format - default):**
```json
{
  "code": "Ok",
  "format": "geojson",
  "routes": [
    {
      "distance": 2927.70,
      "duration": 210.78,
      "geometry": {
        "type": "LineString",
        "coordinates": [
          [7.4184524, 43.7299355],
          [7.4185197, 43.7293154],
          [7.4185385, 43.7291224]
        ]
      }
    }
  ]
}
```

**Response (Polyline format):**
```json
{
  "code": "Ok",
  "format": "polyline",
  "routes": [
    {
      "distance": 2927.70,
      "duration": 210.78,
      "geometry": "y~gxGkdifC?zB@n@BZ?VJj@Lh@Pr@..."
    }
  ]
}
```

**Geometry Formats:**
- **GeoJSON** (default): Standard GeoJSON LineString - best for map visualization
- **Polyline**: Google Polyline encoded string - 50-70% smaller, saves bandwidth

See [docs/GEOMETRY_FORMATS.md](docs/GEOMETRY_FORMATS.md) for detailed format documentation.

### GET /route/get

Find route using query parameters.

**Example:**
```
GET /route/get?from_lat=43.73&from_lon=7.42&to_lat=43.74&to_lon=7.43&alternatives=1&format=geojson
```

**Query Parameters:**
- `from_lat`, `from_lon`: Starting coordinates
- `to_lat`, `to_lon`: Destination coordinates  
- `alternatives` (optional): Number of alternative routes (default: 0)
- `format` (optional): Geometry format - `geojson` or `polyline` (default: `geojson`)
- `profile` (optional): Routing profile - `car`, `bike`, or `foot` (default: `car`)

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

### Routing Profiles

Three pre-configured routing profiles for different transportation modes:

#### Car Profile (Default)
- **Allowed Roads**: Motorways, trunk roads, primary/secondary/tertiary roads, residential
- **Speed Optimization**: Highways +20% faster, residential -20% slower
- **Max Speed**: 120 km/h
- **Use Cases**: Car navigation, driving directions

#### Bike Profile
- **Allowed Roads**: Cycleways, paths, residential, secondary roads (avoids motorways)
- **Speed Optimization**: Cycleways +20% preferred, primary roads -30% (less safe)
- **Avoids**: Gravel and sand surfaces
- **Max Speed**: 30 km/h
- **Use Cases**: Bicycle navigation, cycling routes

#### Foot Profile
- **Allowed Roads**: Footways, pedestrian areas, steps, paths, residential
- **Speed Optimization**: Footways +20% preferred, stairs -20% slower
- **Max Speed**: 5 km/h
- **Use Cases**: Walking directions, pedestrian navigation

### Turn Restrictions

Automatically parses and enforces OSM turn restrictions:

- ❌ **No-turn restrictions**: `no_left_turn`, `no_right_turn`, `no_u_turn`, `no_straight_on`
- ✅ **Only-turn restrictions**: `only_left_turn`, `only_right_turn`, `only_straight_on`
- Monaco dataset includes **44 turn restrictions** automatically enforced

### Oneway Support

Complete handling of one-way streets:

- ✅ `oneway=yes` or `oneway=1` - Forward direction only
- ✅ `oneway=-1` or `oneway=reverse` - Reverse direction only
- ✅ Prevents routing against traffic flow

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
# Find a route (car, default)
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": 43.73,
    "from_lon": 7.42,
    "to_lat": 43.74,
    "to_lon": 7.43
  }'

# Find a bicycle route
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": 43.73,
    "from_lon": 7.42,
    "to_lat": 43.74,
    "to_lon": 7.43,
    "profile": "bike"
  }'

# Find a walking route with Polyline format
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": 43.73,
    "from_lon": 7.42,
    "to_lat": 43.74,
    "to_lon": 7.43,
    "profile": "foot",
    "format": "polyline"
  }'

# Find alternative car routes
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": 43.73,
    "from_lon": 7.42,
    "to_lat": 43.74,
    "to_lon": 7.43,
    "alternatives": 2,
    "profile": "car"
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

### High Priority (重要且常用)

- [ ] **Routing Profiles** (car, bike, pedestrian) - 不同交通方式的路由配置
  - 影响: 扩展应用场景，满足多种出行需求
  - 难度: 中等
  - 价值: ⭐⭐⭐⭐⭐

- [ ] **Turn Restrictions** - 转弯限制支持
  - 影响: 提高路线准确性，避免非法转弯
  - 难度: 中等
  - 价值: ⭐⭐⭐⭐⭐

- [ ] **Performance Benchmarks** - 性能基准测试
  - 影响: 了解系统性能瓶颈，优化方向
  - 难度: 低
  - 价值: ⭐⭐⭐⭐

### Medium Priority (有用但不紧急)

- [ ] **Bidirectional A*** - 双向搜索优化
  - 影响: 提升长距离路线查询速度 2-3倍
  - 难度: 中等
  - 价值: ⭐⭐⭐⭐

- [ ] **Isochrone Generation** - 等时圈生成
  - 影响: 新功能，可视化可达范围
  - 难度: 中等
  - 价值: ⭐⭐⭐

- [ ] **Map Matching** - GPS轨迹匹配
  - 影响: 支持轨迹分析和导航纠偏
  - 难度: 高
  - 价值: ⭐⭐⭐

### Low Priority (可选功能)

- [ ] **Time-Dependent Routing** - 时间相关路由
  - 影响: 支持实时交通和高峰期路由
  - 难度: 高
  - 价值: ⭐⭐

- [ ] **GraphQL API** - GraphQL 接口
  - 影响: 提供更灵活的 API 查询方式
  - 难度: 低
  - 价值: ⭐⭐


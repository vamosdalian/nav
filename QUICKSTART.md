# Quick Start Guide

Get your navigation service running in 5 minutes!

## Step 1: Build the Server

```bash
cd /Users/lmc10232/project/nav
go build -o nav-server cmd/server/main.go
```

## Step 2: Download Sample Map Data

Download a small OSM dataset (Monaco - ~1MB) for testing:

```bash
make download-sample
# Or manually:
# curl -o monaco-latest.osm.pbf https://download.geofabrik.de/europe/monaco-latest.osm.pbf
```

**For other regions**, visit [Geofabrik Downloads](https://download.geofabrik.de/):
- Cities: https://download.geofabrik.de/europe.html (select your city)
- States/Provinces: Download state-level extracts
- Countries: Full country extracts available

## Step 3: Start the Server

```bash
# First run - parse OSM data (takes a few seconds)
OSM_DATA_PATH=monaco-latest.osm.pbf ./nav-server

# Or use the makefile
make run-sample
```

You should see:
```
Starting Navigation Service...
Parsing OSM data from monaco-latest.osm.pbf...
Graph built: XXXX nodes, YYYY edges
Server listening on :8080
```

## Step 4: Test the API

### Health Check
```bash
curl http://localhost:8080/health
```

### Find a Route

**Monaco Coordinates** (for testing - verified working):
- Port Area: 43.73, 7.42
- Monte Carlo: 43.74, 7.43
- Center: 43.7384, 7.4246

```bash
# Find route in Monaco (verified working)
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": 43.73,
    "from_lon": 7.42,
    "to_lat": 43.74,
    "to_lon": 7.43
  }'
```

Response:
```json
{
  "code": "Ok",
  "routes": [
    {
      "distance": 1234.56,
      "duration": 88.85,
      "geometry": [[7.4279, 43.7396], [7.4197, 43.7311]]
    }
  ]
}
```

### Find Alternative Routes

```bash
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": 43.73,
    "from_lon": 7.42,
    "to_lat": 43.74,
    "to_lon": 7.43,
    "alternatives": 2
  }'
```

### Update Road Weights (Simulate Traffic)

```bash
curl -X POST http://localhost:8080/weight/update \
  -H "Content-Type: application/json" \
  -d '{
    "osm_way_id": 123456789,
    "multiplier": 2.0
  }'
```

## Step 5: Faster Subsequent Runs

After the first run, the graph is cached to `graph.bin.gz`. Next time:

```bash
# Much faster - loads from cache
GRAPH_DATA_PATH=graph.bin.gz ./nav-server
```

Or specify both (will try cache first):
```bash
OSM_DATA_PATH=monaco-latest.osm.pbf GRAPH_DATA_PATH=graph.bin.gz ./nav-server
```

## Advanced Usage

### Use a Different Region

```bash
# Download your region (example: Paris)
wget https://download.geofabrik.de/europe/france/ile-de-france-latest.osm.pbf

# Parse and run
OSM_DATA_PATH=ile-de-france-latest.osm.pbf GRAPH_DATA_PATH=paris-graph.bin.gz ./nav-server
```

### Docker Deployment

```bash
# Build image
docker build -t nav-server .

# Run with your OSM data
docker run -p 8080:8080 \
  -v $(pwd)/data:/data \
  -e OSM_DATA_PATH=/data/monaco-latest.osm.pbf \
  -e GRAPH_DATA_PATH=/data/graph.bin.gz \
  nav-server
```

### Docker Compose

```bash
# Place your OSM file in ./data/ directory
mkdir -p data
cp monaco-latest.osm.pbf data/map.osm.pbf

# Start
docker-compose up
```

## Configuration Options

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| PORT | 8080 | Server port |
| OSM_DATA_PATH | - | Path to OSM PBF file |
| GRAPH_DATA_PATH | graph.bin.gz | Path to cached graph |
| LOG_LEVEL | info | Logging level |

## Troubleshooting

### "No graph data available"
Make sure to set either `OSM_DATA_PATH` or `GRAPH_DATA_PATH`.

### Parsing takes too long
Large regions (countries) can take 10-30 minutes to parse. Use smaller extracts (cities/states) or load cached graphs.

### Out of memory
Large countries need 4-16GB RAM. Use smaller regions or increase memory.

### No route found
Make sure coordinates are within your loaded map region.

## Next Steps

- Read the full [README.md](README.md) for architecture details
- Check [examples/](examples/) for API client examples
- Explore the API: `http://localhost:8080/health`

## API Endpoints Summary

| Endpoint | Method | Description |
|----------|--------|-------------|
| /health | GET | Health check and stats |
| /route | POST | Find route (JSON body) |
| /route/get | GET | Find route (query params) |
| /weight/update | POST | Update edge weights |

## Example Use Cases

1. **Basic Navigation**: Single shortest path
2. **Alternative Routes**: Multiple route options
3. **Traffic Simulation**: Modify weights in real-time
4. **Custom Routing**: Different vehicle profiles
5. **Service Integration**: Embed in your application

Happy routing! 🗺️


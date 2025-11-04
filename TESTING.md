# Testing Guide

## Quick Test Commands

### Start the Server

```bash
cd /Users/lmc10232/project/nav
make run-sample
```

Or manually:
```bash
OSM_DATA_PATH=monaco-latest.osm.pbf go run cmd/server/main.go
```

### Test Endpoints

#### 1. Health Check
```bash
curl http://localhost:8080/health
```

Expected output:
```json
{
  "status": "healthy",
  "nodes": 7427,
  "edges": 11921
}
```

#### 2. Find a Route (Working Example)
```bash
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": 43.73,
    "from_lon": 7.42,
    "to_lat": 43.74,
    "to_lon": 7.43
  }'
```

Expected output:
```json
{
  "code": "Ok",
  "routes": [
    {
      "distance": 2927.7,
      "duration": 210.8,
      "geometry": [[7.4184524, 43.7299355], ...]
    }
  ]
}
```

#### 3. Find Alternative Routes
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

#### 4. Using GET Method
```bash
curl "http://localhost:8080/route/get?from_lat=43.73&from_lon=7.42&to_lat=43.74&to_lon=7.43"
```

#### 5. Update Road Weights
```bash
# First, find a route and note an OSM way ID from the response
# Then update its weight (e.g., simulate heavy traffic)
curl -X POST http://localhost:8080/weight/update \
  -H "Content-Type: application/json" \
  -d '{
    "osm_way_id": 123456789,
    "multiplier": 2.0
  }'
```

## Testing with Python

```python
import requests

# Find a route
response = requests.post('http://localhost:8080/route', json={
    'from_lat': 43.73,
    'from_lon': 7.42,
    'to_lat': 43.74,
    'to_lon': 7.43
})

result = response.json()
print(f"Distance: {result['routes'][0]['distance']:.2f} meters")
print(f"Duration: {result['routes'][0]['duration']:.2f} seconds")
print(f"Points: {len(result['routes'][0]['geometry'])}")
```

## Important Notes

### Coordinates Must Be On Connected Roads

- Not all coordinates will have routes between them
- Monaco has some disconnected road segments
- Make sure coordinates are on actual drivable roads

### Good Test Coordinates for Monaco

These coordinates are verified to work:

| Location | Latitude | Longitude |
|----------|----------|-----------|
| Port Area | 43.73 | 7.42 |
| Monte Carlo | 43.74 | 7.43 |
| North | 43.745 | 7.425 |
| South | 43.728 | 7.420 |

### Testing with Different Regions

When using a different OSM file:

1. Download from [Geofabrik](https://download.geofabrik.de/)
2. Use coordinates within that region
3. Check OpenStreetMap to find actual road coordinates

```bash
# Example: Testing with Berlin
wget https://download.geofabrik.de/europe/germany/berlin-latest.osm.pbf
OSM_DATA_PATH=berlin-latest.osm.pbf go run cmd/server/main.go

# Test coordinates in Berlin
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": 52.52,
    "from_lon": 13.405,
    "to_lat": 52.51,
    "to_lon": 13.42
  }'
```

## Performance Testing

### Single Route Performance
```bash
time curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": 43.73,
    "from_lon": 7.42,
    "to_lat": 43.74,
    "to_lon": 7.43
  }'
```

Expected: <100ms for Monaco-sized regions

### Concurrent Requests
```bash
# Install Apache Bench
# sudo apt-get install apache2-utils

# Test with 100 requests, 10 concurrent
ab -n 100 -c 10 -p request.json -T application/json http://localhost:8080/route
```

request.json:
```json
{
  "from_lat": 43.73,
  "from_lon": 7.42,
  "to_lat": 43.74,
  "to_lon": 7.43
}
```

## Troubleshooting Tests

### "No route found"

**Possible causes:**
1. Coordinates are not on connected road segments
2. Coordinates are outside the loaded map region
3. Road network has disconnected components

**Solutions:**
- Use verified test coordinates above
- Check coordinates on openstreetmap.org
- Try coordinates closer together

### Server won't start

**Check:**
```bash
# Is port 8080 in use?
lsof -i :8080

# Kill existing process
lsof -ti:8080 | xargs kill -9
```

### Slow parsing

**Normal for large regions:**
- Monaco: ~1 second
- City: 10-60 seconds
- State: 1-10 minutes
- Country: 10-60 minutes

**Use cached graph:**
```bash
GRAPH_DATA_PATH=graph.bin.gz go run cmd/server/main.go
```

## Automated Testing

Run all example tests:
```bash
cd examples
chmod +x api_examples.sh
./api_examples.sh
```

## Expected Performance Metrics

| Metric | Monaco | City | State |
|--------|--------|------|-------|
| Graph Load | <1s | 1-5s | 5-30s |
| Route Query | <10ms | 10-50ms | 50-200ms |
| Memory Usage | ~10MB | ~100MB | ~1GB |

## CI/CD Testing

Example GitHub Actions workflow:

```yaml
name: Test Navigation Service
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.25'
      - name: Download test data
        run: make download-sample
      - name: Build
        run: go build cmd/server/main.go
      - name: Start server
        run: OSM_DATA_PATH=monaco-latest.osm.pbf ./main &
      - name: Wait for server
        run: sleep 5
      - name: Test health
        run: curl -f http://localhost:8080/health
      - name: Test routing
        run: |
          curl -f -X POST http://localhost:8080/route \
            -H "Content-Type: application/json" \
            -d '{"from_lat": 43.73, "from_lon": 7.42, "to_lat": 43.74, "to_lon": 7.43}'
```

## Success Criteria

✅ Server starts without errors  
✅ Health endpoint returns 200  
✅ Node and edge counts > 0  
✅ Route query returns valid geometry  
✅ Distance and duration are positive numbers  
✅ Alternative routes differ from each other  
✅ Weight updates return success  

Happy testing! 🧪


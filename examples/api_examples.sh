#!/bin/bash

# API Examples for Navigation Service
# Make sure the server is running before executing these examples

BASE_URL="http://localhost:8080"

echo "=== Navigation Service API Examples ==="
echo

# 1. Health Check
echo "1. Health Check"
echo "GET $BASE_URL/health"
curl -s $BASE_URL/health | jq .
echo
echo "---"
echo

# 2. Find a single route (POST)
echo "2. Find a single route (POST /route)"
echo "Finding route in Monaco..."
curl -s -X POST $BASE_URL/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": 43.73,
    "from_lon": 7.42,
    "to_lat": 43.74,
    "to_lon": 7.43
  }' | jq .
echo
echo "---"
echo

# 3. Find multiple alternative routes
echo "3. Find alternative routes"
curl -s -X POST $BASE_URL/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": 43.73,
    "from_lon": 7.42,
    "to_lat": 43.74,
    "to_lon": 7.43,
    "alternatives": 2
  }' | jq .
echo
echo "---"
echo

# 4. Find route using GET method
echo "4. Find route using GET method"
curl -s "$BASE_URL/route/get?from_lat=43.73&from_lon=7.42&to_lat=43.74&to_lon=7.43" | jq .
echo
echo "---"
echo

# 5. Update edge weights (simulate traffic)
echo "5. Update edge weights (simulate heavy traffic)"
curl -s -X POST $BASE_URL/weight/update \
  -H "Content-Type: application/json" \
  -d '{
    "osm_way_id": 123456789,
    "multiplier": 2.5
  }' | jq .
echo
echo "---"
echo

echo "=== Examples Complete ==="


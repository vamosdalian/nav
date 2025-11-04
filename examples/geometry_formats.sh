#!/bin/bash

# Geometry Formats Test Script
# 测试不同的几何格式

BASE_URL="http://localhost:8080"

echo "========================================="
echo "  导航服务 - 几何格式测试"
echo "========================================="
echo

# Test coordinates
FROM_LAT=43.73
FROM_LON=7.42
TO_LAT=43.74
TO_LON=7.43

# 1. GeoJSON Format (默认)
echo "1. GeoJSON 格式 (默认)"
echo "----------------------------------------"
curl -s -X POST $BASE_URL/route \
  -H "Content-Type: application/json" \
  -d "{
    \"from_lat\": $FROM_LAT,
    \"from_lon\": $FROM_LON,
    \"to_lat\": $TO_LAT,
    \"to_lon\": $TO_LON
  }" | jq '{
    format: .format,
    geometry_type: .routes[0].geometry.type,
    total_points: (.routes[0].geometry.coordinates | length),
    first_3_coords: .routes[0].geometry.coordinates[:3],
    distance: .routes[0].distance,
    duration: .routes[0].duration
  }'
echo
echo

# 2. Polyline Format
echo "2. Polyline 格式（编码）"
echo "----------------------------------------"
curl -s -X POST $BASE_URL/route \
  -H "Content-Type: application/json" \
  -d "{
    \"from_lat\": $FROM_LAT,
    \"from_lon\": $FROM_LON,
    \"to_lat\": $TO_LAT,
    \"to_lon\": $TO_LON,
    \"format\": \"polyline\"
  }" | jq '{
    format: .format,
    geometry_preview: (.routes[0].geometry[:60] + "..."),
    geometry_length: (.routes[0].geometry | length),
    distance: .routes[0].distance,
    duration: .routes[0].duration
  }'
echo
echo

# 3. 使用 GET 方法测试 Polyline
echo "3. GET 请求 + Polyline 格式"
echo "----------------------------------------"
curl -s "$BASE_URL/route/get?from_lat=$FROM_LAT&from_lon=$FROM_LON&to_lat=$TO_LAT&to_lon=$TO_LON&format=polyline" \
  | jq '{
    format: .format,
    geometry_preview: (.routes[0].geometry[:50] + "...")
  }'
echo
echo

# 4. 多条路线 + GeoJSON
echo "4. 多条备选路线 + GeoJSON"
echo "----------------------------------------"
curl -s -X POST $BASE_URL/route \
  -H "Content-Type: application/json" \
  -d "{
    \"from_lat\": $FROM_LAT,
    \"from_lon\": $FROM_LON,
    \"to_lat\": $TO_LAT,
    \"to_lon\": $TO_LON,
    \"alternatives\": 2,
    \"format\": \"geojson\"
  }" | jq '{
    format: .format,
    total_routes: (.routes | length),
    routes: [.routes[] | {
      distance: .distance,
      duration: .duration,
      points: (.geometry.coordinates | length)
    }]
  }'
echo

echo "========================================="
echo "  测试完成！"
echo "========================================="


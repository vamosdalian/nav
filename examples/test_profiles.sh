#!/bin/bash

# 测试路由配置功能
# Test Routing Profiles

BASE_URL="http://localhost:8080"

echo "========================================="
echo "  路由配置测试"
echo "  Testing Routing Profiles"
echo "========================================="
echo

# Test coordinates
FROM_LAT=43.73
FROM_LON=7.42
TO_LAT=43.74
TO_LON=7.43

echo "测试坐标: ($FROM_LAT, $FROM_LON) -> ($TO_LAT, $TO_LON)"
echo "Test coordinates: ($FROM_LAT, $FROM_LON) -> ($TO_LAT, $TO_LON)"
echo
echo "========================================="
echo

# 1. 汽车路由
echo "1. 汽车路由 (Car Routing)"
echo "----------------------------------------"
curl -s -X POST $BASE_URL/route \
  -H "Content-Type: application/json" \
  -d "{
    \"from_lat\": $FROM_LAT,
    \"from_lon\": $FROM_LON,
    \"to_lat\": $TO_LAT,
    \"to_lon\": $TO_LON,
    \"profile\": \"car\"
  }" | jq '{
    code: .code,
    distance_m: .routes[0].distance,
    duration_s: .routes[0].duration,
    points: (.routes[0].geometry.coordinates | length)
  }' || echo "ERROR: 路由失败"
echo
echo

# 2. 自行车路由
echo "2. 自行车路由 (Bike Routing)"
echo "----------------------------------------"
curl -s -X POST $BASE_URL/route \
  -H "Content-Type: application/json" \
  -d "{
    \"from_lat\": $FROM_LAT,
    \"from_lon\": $FROM_LON,
    \"to_lat\": $TO_LAT,
    \"to_lon\": $TO_LON,
    \"profile\": \"bike\"
  }" | jq '{
    code: .code,
    distance_m: .routes[0].distance,
    duration_s: .routes[0].duration,
    points: (.routes[0].geometry.coordinates | length)
  }' || echo "ERROR: 路由失败"
echo
echo

# 3. 步行路由
echo "3. 步行路由 (Foot Routing)"
echo "----------------------------------------"
curl -s -X POST $BASE_URL/route \
  -H "Content-Type: application/json" \
  -d "{
    \"from_lat\": $FROM_LAT,
    \"from_lon\": $FROM_LON,
    \"to_lat\": $TO_LAT,
    \"to_lon\": $TO_LON,
    \"profile\": \"foot\"
  }" | jq '{
    code: .code,
    distance_m: .routes[0].distance,
    duration_s: .routes[0].duration,
    points: (.routes[0].geometry.coordinates | length)
  }' || echo "ERROR: 路由失败"
echo
echo

# 4. 对比测试 - 不同 profile 的结果
echo "4. 对比不同 Profile 的路线差异"
echo "----------------------------------------"
echo "汽车:"
curl -s -X POST $BASE_URL/route \
  -H "Content-Type: application/json" \
  -d "{
    \"from_lat\": $FROM_LAT,
    \"from_lon\": $FROM_LON,
    \"to_lat\": $TO_LAT,
    \"to_lon\": $TO_LON,
    \"profile\": \"car\",
    \"format\": \"polyline\"
  }" | jq -r '.routes[0].geometry' | wc -c

echo "自行车:"
curl -s -X POST $BASE_URL/route \
  -H "Content-Type: application/json" \
  -d "{
    \"from_lat\": $FROM_LAT,
    \"from_lon\": $FROM_LON,
    \"to_lat\": $TO_LAT,
    \"to_lon\": $TO_LON,
    \"profile\": \"bike\",
    \"format\": \"polyline\"
  }" | jq -r '.routes[0].geometry' | wc -c

echo "步行:"
curl -s -X POST $BASE_URL/route \
  -H "Content-Type: application/json" \
  -d "{
    \"from_lat\": $FROM_LAT,
    \"from_lon\": $FROM_LON,
    \"to_lat\": $TO_LAT,
    \"to_lon\": $TO_LON,
    \"profile\": \"foot\",
    \"format\": \"polyline\"
  }" | jq -r '.routes[0].geometry' | wc -c
echo

echo "========================================="
echo "  测试完成！"
echo "  Tests Complete!"
echo "========================================="


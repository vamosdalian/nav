# 几何格式选项 / Geometry Format Options

导航服务支持两种几何格式返回路线数据。

## 支持的格式

### 1. GeoJSON (默认)

标准 GeoJSON LineString 格式，最适合地图可视化。

**请求示例:**
```bash
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": 43.73,
    "from_lon": 7.42,
    "to_lat": 43.74,
    "to_lon": 7.43,
    "format": "geojson"
  }'
```

或者省略 `format` 参数（默认使用 GeoJSON）:
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

**响应示例:**
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
          [7.4185385, 43.7291224],
          ...
        ]
      }
    }
  ]
}
```

**适用场景:**
- 直接在地图上显示（Leaflet, Mapbox, OpenLayers）
- 符合 GeoJSON 标准
- 易于可视化和调试

---

### 2. Polyline (编码格式)

Google Polyline 编码格式，压缩的字符串表示，节省带宽。

**请求示例:**
```bash
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": 43.73,
    "from_lon": 7.42,
    "to_lat": 43.74,
    "to_lon": 7.43,
    "format": "polyline"
  }'
```

**GET 请求示例:**
```bash
curl "http://localhost:8080/route/get?from_lat=43.73&from_lon=7.42&to_lat=43.74&to_lon=7.43&format=polyline"
```

**响应示例:**
```json
{
  "code": "Ok",
  "format": "polyline",
  "routes": [
    {
      "distance": 2927.70,
      "duration": 210.78,
      "geometry": "y~gxGkdifC?zB@n@BZ?VJj@Lh@Pr@Nj@LXJPHJH..."
    }
  ]
}
```

**适用场景:**
- 减少数据传输量（比坐标数组小 50-70%）
- 与 Google Maps API 兼容
- 移动应用（节省流量）
- 需要存储大量路线

**解码示例 (JavaScript):**
```javascript
// 使用 @mapbox/polyline 库
const polyline = require('@mapbox/polyline');
const decoded = polyline.decode('y~gxGkdifC?zB@n@...');
// decoded = [[43.72994, 7.41845], [43.72932, 7.41852], ...]
```

**解码示例 (Python):**
```python
# 使用 polyline 库
import polyline
coords = polyline.decode('y~gxGkdifC?zB@n@...')
# coords = [(43.72994, 7.41845), (43.72932, 7.41852), ...]
```

---

## 格式对比

| 格式 | 数据大小 | 易用性 | 标准化 | 适用场景 |
|------|---------|--------|-------|---------|
| **GeoJSON** | 大 | ⭐⭐⭐⭐⭐ | GeoJSON 标准 | 地图可视化、Web 应用 |
| **Polyline** | 小（-50~70%） | ⭐⭐⭐ | Google 标准 | 移动应用、存储、API 集成 |

## 性能对比

以一条 231 个点的路线为例：

| 格式 | 大小 (bytes) | 压缩比 |
|------|-------------|--------|
| GeoJSON | ~15,500 | 100% |
| Polyline | ~4,200 | 27% |

---

## 使用示例

### Python 客户端

```python
import requests

# GeoJSON 格式（默认）
response = requests.post('http://localhost:8080/route', json={
    'from_lat': 43.73,
    'from_lon': 7.42,
    'to_lat': 43.74,
    'to_lon': 7.43
}).json()

geojson = response['routes'][0]['geometry']
print(f"Type: {geojson['type']}")  # LineString
print(f"Points: {len(geojson['coordinates'])}")

# Polyline 格式
response = requests.post('http://localhost:8080/route', json={
    'from_lat': 43.73,
    'from_lon': 7.42,
    'to_lat': 43.74,
    'to_lon': 7.43,
    'format': 'polyline'
}).json()

polyline_str = response['routes'][0]['geometry']
print(f"Polyline: {polyline_str[:50]}...")

# 解码 polyline
import polyline
coords = polyline.decode(polyline_str)
print(f"Decoded points: {len(coords)}")
```

### JavaScript 客户端

```javascript
// GeoJSON 格式
const response = await fetch('http://localhost:8080/route', {
  method: 'POST',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    from_lat: 43.73,
    from_lon: 7.42,
    to_lat: 43.74,
    to_lon: 7.43,
    format: 'geojson'
  })
});

const data = await response.json();
const geojson = data.routes[0].geometry;

// 直接在 Leaflet 中使用
L.geoJSON(geojson).addTo(map);

// Polyline 格式
const polylineResponse = await fetch('http://localhost:8080/route', {
  method: 'POST',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    from_lat: 43.73,
    from_lon: 7.42,
    to_lat: 43.74,
    to_lon: 7.43,
    format: 'polyline'
  })
});

const polylineData = await polylineResponse.json();
const encoded = polylineData.routes[0].geometry;

// 使用 @mapbox/polyline 解码
const polyline = require('@mapbox/polyline');
const coords = polyline.decode(encoded);
```

### Go 客户端

```go
package main

import (
    "encoding/json"
    "bytes"
    "net/http"
)

type RouteRequest struct {
    FromLat float64 `json:"from_lat"`
    FromLon float64 `json:"from_lon"`
    ToLat   float64 `json:"to_lat"`
    ToLon   float64 `json:"to_lon"`
    Format  string  `json:"format"`
}

func main() {
    // GeoJSON 格式
    req := RouteRequest{
        FromLat: 43.73,
        FromLon: 7.42,
        ToLat:   43.74,
        ToLon:   7.43,
        Format:  "geojson",
    }
    
    jsonData, _ := json.Marshal(req)
    resp, _ := http.Post("http://localhost:8080/route", 
                         "application/json", 
                         bytes.NewBuffer(jsonData))
    
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    
    // Polyline 格式
    req.Format = "polyline"
    jsonData, _ = json.Marshal(req)
    resp, _ = http.Post("http://localhost:8080/route",
                       "application/json",
                       bytes.NewBuffer(jsonData))
}
```

---

## 选择建议

**使用 GeoJSON 如果:**
- ✅ 在 Web 地图上显示路线
- ✅ 需要符合标准的 GeoJSON 格式
- ✅ 与 Leaflet、Mapbox、OpenLayers 等库集成
- ✅ 数据大小不是主要考虑因素

**使用 Polyline 如果:**
- ✅ 需要最小的数据传输量
- ✅ 开发移动应用（节省流量）
- ✅ 需要存储大量路线
- ✅ 与 Google Maps API 集成
- ✅ 需要传输数百条路线

---

## API 参考

### POST /route

**请求参数:**
```json
{
  "from_lat": 43.73,
  "from_lon": 7.42,
  "to_lat": 43.74,
  "to_lon": 7.43,
  "alternatives": 2,
  "format": "geojson"  // 可选: "geojson" 或 "polyline"
}
```

### GET /route/get

**查询参数:**
```
?from_lat=43.73
&from_lon=7.42
&to_lat=43.74
&to_lon=7.43
&alternatives=2
&format=polyline
```

---

## Polyline 算法说明

我们使用 Google 的 Polyline 编码算法：
- **标准**: [Google Maps Polyline Algorithm](https://developers.google.com/maps/documentation/utilities/polylinealgorithm)
- **精度**: 5 位小数（~1 米精度）
- **压缩**: 差分编码 + 变长整数
- **兼容**: 可用 Google 和其他解码库解码

**编码过程:**
1. 将经纬度乘以 1e5 并四舍五入为整数
2. 计算与前一个点的差值
3. 差值编码为变长整数
4. 转换为 ASCII 字符

**解码库:**
- Python: `pip install polyline`
- JavaScript: `npm install @mapbox/polyline`
- Go: `github.com/twpayne/go-polyline`

---

## 注意事项

1. **默认格式**: 如果不指定 `format`，默认使用 GeoJSON
2. **坐标顺序**: 所有格式都使用 `[longitude, latitude]` 顺序（GeoJSON 标准）
3. **精度**: Polyline 格式精度为 5 位小数（~1 米），其他格式保留原始精度
4. **兼容性**: Polyline 格式与 Google Maps API 完全兼容

---

## 示例脚本

完整的示例脚本请参见：
- `examples/geometry_formats.sh` - Bash 测试脚本
- `examples/geometry_formats.py` - Python 示例
- `examples/geometry_formats.js` - JavaScript 示例

Happy routing! 🗺️


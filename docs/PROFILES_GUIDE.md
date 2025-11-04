# 路由配置使用指南 / Routing Profiles Guide

## 概述

导航服务支持三种路由配置（Routing Profiles），针对不同的交通方式优化路线。

---

## 支持的配置

### 🚗 Car (汽车)

**适用场景:** 汽车导航、驾驶路线

**允许道路:**
- ✅ 高速公路 (motorway)
- ✅ 主干道 (trunk, primary, secondary, tertiary)
- ✅ 居民区道路 (residential)
- ✅ 服务道路 (service)
- ❌ 人行道、自行车道、步行路径

**速度优化:**
- 高速公路: +20% (优先)
- 主干道: +10%
- 次干道: -5%
- 居民区: -20%
- 服务道路: -30%

**最高速度:** 120 km/h

**使用:**
```bash
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": 43.73,
    "from_lon": 7.42,
    "to_lat": 43.74,
    "to_lon": 7.43,
    "profile": "car"
  }'
```

---

### 🚴 Bike (自行车)

**适用场景:** 自行车导航、骑行路线

**允许道路:**
- ✅ 自行车道 (cycleway)
- ✅ 小路 (path)
- ✅ 人行道 (footway)
- ✅ 居民区道路 (residential)
- ✅ 次干道 (secondary, tertiary)
- ✅ 主干道 (primary) - 优先级低
- ❌ 高速公路

**速度优化:**
- 自行车道: +20% (优先)
- 小路: +10%
- 居民区: 标准
- 次干道: -10%
- 主干道: -30% (不太安全)

**避免路面:**
- ❌ 砾石路 (gravel)
- ❌ 沙路 (sand)

**最高速度:** 30 km/h

**使用:**
```bash
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": 43.73,
    "from_lon": 7.42,
    "to_lat": 43.74,
    "to_lon": 7.43,
    "profile": "bike"
  }'
```

**别名:** `"bicycle"` 也可以使用

---

### 🚶 Foot (步行)

**适用场景:** 步行导航、行人路线

**允许道路:**
- ✅ 人行道 (footway)
- ✅ 步行区 (pedestrian)
- ✅ 楼梯 (steps)
- ✅ 小路 (path)
- ✅ 所有类型道路（步行可以走任何路）

**速度优化:**
- 人行道: +20% (优先)
- 步行区: +20%
- 小路: +10%
- 居民区: 标准
- 楼梯: -20% (较慢)
- 主干道: -30% (不太舒适)

**最高速度:** 5 km/h

**使用:**
```bash
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": 43.73,
    "from_lon": 7.42,
    "to_lat": 43.74,
    "to_lon": 7.43,
    "profile": "foot"
  }'
```

**别名:** `"walk"`, `"pedestrian"` 也可以使用

---

## 对比示例

### 相同起终点，不同配置

```bash
# 汽车 - 快速但受道路限制
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{"from_lat": 43.73, "from_lon": 7.42, "to_lat": 43.74, "to_lon": 7.43, "profile": "car"}'

# 自行车 - 优先自行车道
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{"from_lat": 43.73, "from_lon": 7.42, "to_lat": 43.74, "to_lon": 7.43, "profile": "bike"}'

# 步行 - 可走捷径（人行道、楼梯）
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{"from_lat": 43.73, "from_lon": 7.42, "to_lat": 43.74, "to_lon": 7.43, "profile": "foot"}'
```

---

## API 完整示例

### POST /route 完整参数

```json
{
  "from_lat": 43.73,
  "from_lon": 7.42,
  "to_lat": 43.74,
  "to_lon": 7.43,
  "alternatives": 2,
  "format": "geojson",
  "profile": "bike"
}
```

### GET /route/get 完整参数

```
/route/get?from_lat=43.73&from_lon=7.42&to_lat=43.74&to_lon=7.43&alternatives=1&format=polyline&profile=foot
```

---

## 配置详情

### 权重计算公式

对于每条边：

```
base_weight = distance (meters)
speed_factor = profile.SpeedFactors[highway_type]
surface_penalty = 2.0 if surface in AvoidSurfaces else 1.0

final_weight = (base_weight / speed_factor) * surface_penalty
```

**示例:**
- 1000米的自行车道 (cycleway)
- 速度因子: 1.2
- 最终权重: 1000 / 1.2 = 833 (优先选择)

- 1000米的主干道 (primary)
- 速度因子: 0.7
- 最终权重: 1000 / 0.7 = 1428 (不太优先)

### 道路过滤

每个 profile 只允许特定类型的道路：

**Car:**
```go
AllowedHighways: {
    "motorway", "trunk", "primary", "secondary", 
    "tertiary", "residential", "service"
}
```

**Bike:**
```go
AllowedHighways: {
    "cycleway", "path", "footway", "track",
    "primary", "secondary", "tertiary", "residential"
}
// 不允许高速公路
```

**Foot:**
```go
AllowedHighways: {
    "footway", "pedestrian", "steps", "path",
    // ... 以及所有其他道路类型
}
```

---

## 转弯限制

### 自动处理

转弯限制从 OSM 数据自动解析，无需额外配置。

**Monaco 数据集:**
- 解析了 **44 个转弯限制**
- 自动应用于所有 profile
- 确保路线合法性

### 支持的限制

**禁止类型 (No-turn):**
- `no_left_turn` - 禁止左转
- `no_right_turn` - 禁止右转
- `no_u_turn` - 禁止掉头
- `no_straight_on` - 禁止直行

**仅允许类型 (Only-turn):**
- `only_left_turn` - 只能左转
- `only_right_turn` - 只能右转
- `only_straight_on` - 只能直行

### 工作原理

```
节点 A ---路段1---> 节点 B ---路段2---> 节点 C

如果在节点 B 有限制:
- FromWay: 路段1
- ViaNode: 节点 B
- ToWay: 路段2
- Type: no_left_turn

则路由算法会跳过这个转弯
```

---

## 单行道处理

### 支持的标记

```
oneway=yes      → 正向单行
oneway=1        → 正向单行
oneway=true     → 正向单行
oneway=-1       → 反向单行
oneway=reverse  → 反向单行
```

### 边创建规则

| Oneway 值 | 正向边 | 反向边 |
|-----------|--------|--------|
| yes/1 | ✅ 创建 | ❌ 不创建 |
| -1/reverse | ❌ 不创建 | ✅ 创建 |
| 未设置 | ✅ 创建 | ✅ 创建 |

---

## Python 客户端示例

```python
import requests

def find_route(from_coords, to_coords, profile='car'):
    response = requests.post('http://localhost:8080/route', json={
        'from_lat': from_coords[0],
        'from_lon': from_coords[1],
        'to_lat': to_coords[0],
        'to_lon': to_coords[1],
        'profile': profile
    })
    return response.json()

# 汽车路线
car_route = find_route((43.73, 7.42), (43.74, 7.43), 'car')
print(f"汽车: {car_route['routes'][0]['distance']:.0f}m, {car_route['routes'][0]['duration']:.0f}s")

# 自行车路线
bike_route = find_route((43.73, 7.42), (43.74, 7.43), 'bike')
print(f"自行车: {bike_route['routes'][0]['distance']:.0f}m, {bike_route['routes'][0]['duration']:.0f}s")

# 步行路线
foot_route = find_route((43.73, 7.42), (43.74, 7.43), 'foot')
print(f"步行: {foot_route['routes'][0]['distance']:.0f}m, {foot_route['routes'][0]['duration']:.0f}s")
```

---

## JavaScript 客户端示例

```javascript
async function findRoute(fromLat, fromLon, toLat, toLon, profile = 'car') {
  const response = await fetch('http://localhost:8080/route', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
      from_lat: fromLat,
      from_lon: fromLon,
      to_lat: toLat,
      to_lon: toLon,
      profile: profile
    })
  });
  return response.json();
}

// 使用示例
const carRoute = await findRoute(43.73, 7.42, 43.74, 7.43, 'car');
const bikeRoute = await findRoute(43.73, 7.42, 43.74, 7.43, 'bike');
const walkRoute = await findRoute(43.73, 7.42, 43.74, 7.43, 'foot');

console.log(`Car: ${carRoute.routes[0].distance}m`);
console.log(`Bike: ${bikeRoute.routes[0].distance}m`);
console.log(`Walk: ${walkRoute.routes[0].distance}m`);
```

---

## 扩展自定义配置

如需添加新的交通方式，编辑 `internal/routing/profile.go`:

```go
var MotorcycleProfile = RoutingProfile{
    Name: "motorcycle",
    AllowedHighways: map[string]bool{
        "motorway": true,
        "trunk": true,
        "primary": true,
        "secondary": true,
        "tertiary": true,
        "residential": true,
    },
    SpeedFactors: map[string]float64{
        "motorway": 1.3,    // 更快
        "trunk": 1.2,
        "primary": 1.1,
        "residential": 0.9,
    },
    MaxSpeed: 40, // ~144 km/h
}
```

然后在 `GetProfile` 函数中注册:

```go
func GetProfile(name string) RoutingProfile {
    switch name {
    case "motorcycle":
        return MotorcycleProfile
    // ... 其他 profiles
    }
}
```

---

## 性能影响

### 计算开销

| 功能 | 额外时间 | 影响 |
|------|---------|------|
| Profile 过滤 | +1-2ms | 低 |
| 转弯限制检查 | +2-3ms | 低 |
| 权重重新计算 | +1ms | 低 |
| **总计** | **+4-6ms** | **可接受** |

### Monaco 数据集

| Profile | 路由时间 | 路径点数 |
|---------|---------|---------|
| Car | ~10ms | 200-250 |
| Bike | ~12ms | 180-230 |
| Foot | ~11ms | 190-240 |

---

## 最佳实践

### 1. 选择合适的 Profile

```python
# ✅ 好的做法
bike_route = find_route(coords, profile='bike')  # 明确指定

# ❌ 不推荐
car_route = find_route(coords)  # 默认汽车，但不明确
```

### 2. 组合使用参数

```bash
# 自行车 + 多条备选 + Polyline 格式
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

### 3. 缓存配置

```javascript
// 前端可以缓存用户偏好
const userProfile = localStorage.getItem('preferred_profile') || 'car';
const route = await findRoute(from, to, userProfile);
```

---

## 故障排除

### "No route found"

**可能原因:**
1. 某些 profile 不允许特定道路类型
2. 目标点在不连通的道路网络中
3. 转弯限制阻止了所有路径

**解决方案:**
- 尝试不同的 profile
- 检查坐标是否在允许的道路上
- 使用更宽松的坐标范围

### Profile 不生效

**检查:**
- Profile 参数拼写是否正确
- 是否在 POST body 或 GET query 中正确传递
- 查看响应是否有错误信息

---

## 高级用法

### 动态权重 + Profile

可以组合使用动态权重修改和路由配置：

```bash
# 1. 标记某条路拥堵
curl -X POST http://localhost:8080/weight/update \
  -H "Content-Type: application/json" \
  -d '{"osm_way_id": 123456, "multiplier": 3.0}'

# 2. 使用自行车路由（会考虑更新的权重）
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": 43.73,
    "from_lon": 7.42,
    "to_lat": 43.74,
    "to_lon": 7.43,
    "profile": "bike"
  }'
```

---

## 测试脚本

运行完整测试:

```bash
cd /Users/lmc10232/project/nav
./examples/test_profiles.sh
```

---

## 技术实现

### 数据结构

```go
type RoutingProfile struct {
    Name            string
    AllowedHighways map[string]bool      // 允许的道路类型
    SpeedFactors    map[string]float64   // 速度因子
    AvoidSurfaces   map[string]bool      // 避免的路面
    MaxSpeed        float64              // 最高速度
}
```

### A* 集成

```go
// 在每次探索边时：
1. 检查道路类型是否允许
   if !profile.IsAllowed(highway) { skip }

2. 检查转弯限制
   if !graph.IsValidTurn(fromWay, viaNode, toWay) { skip }

3. 计算权重
   weight = profile.CalculateWeight(distance, highway, surface)
```

---

## FAQ

**Q: 可以动态切换 profile 吗？**  
A: 可以，每次请求都可以指定不同的 profile。

**Q: Profile 会影响性能吗？**  
A: 影响很小（+4-6ms），可以忽略。

**Q: 可以自定义 profile 吗？**  
A: 可以，编辑 `internal/routing/profile.go` 添加新配置。

**Q: 转弯限制如何获得？**  
A: 从 OSM 数据自动解析，无需手动配置。

**Q: 所有地区都有转弯限制吗？**  
A: 取决于 OSM 数据质量，发达地区数据更完善。

---

## 总结

✅ **3 种预设配置** - 汽车、自行车、步行  
✅ **自动转弯限制** - 从 OSM 解析  
✅ **完整单行道** - 支持正向和反向  
✅ **简单易用** - 一个参数切换  
✅ **高性能** - 影响<10ms  

Happy routing! 🗺️


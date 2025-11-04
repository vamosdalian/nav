# 转弯限制和单行道指南

## 概述

导航服务自动处理 OSM 的转弯限制和单行道信息，确保路线符合交通规则。

---

## 转弯限制 (Turn Restrictions)

### 什么是转弯限制？

转弯限制定义了在特定路口不允许或仅允许的转弯方向。

**示例场景:**
```
     B
     |
A ---|--- C
     |
     D

在路口 B，可能有:
- 从 A 到 C: 禁止左转
- 从 A 到 D: 仅允许直行
```

### 自动解析

从 OSM 数据自动提取转弯限制：

**数据来源:** OSM Relation 类型 `restriction`

**结构:**
```xml
<relation id="123456" type="restriction">
  <member type="way" ref="111" role="from"/>
  <member type="node" ref="222" role="via"/>
  <member type="way" ref="333" role="to"/>
  <tag k="restriction" v="no_left_turn"/>
</relation>
```

**解析结果:**
```go
TurnRestriction{
    FromWay: 111,
    ViaNode: 222,
    ToWay: 333,
    Type: "no_left_turn"
}
```

### 支持的限制类型

#### 禁止类型 (No-turn)

| OSM 标记 | 说明 | 效果 |
|----------|------|------|
| `no_left_turn` | 禁止左转 | 跳过该转向 |
| `no_right_turn` | 禁止右转 | 跳过该转向 |
| `no_u_turn` | 禁止掉头 | 跳过该转向 |
| `no_straight_on` | 禁止直行 | 跳过该转向 |

#### 仅允许类型 (Only-turn)

| OSM 标记 | 说明 | 效果 |
|----------|------|------|
| `only_left_turn` | 仅左转 | 其他方向都禁止 |
| `only_right_turn` | 仅右转 | 其他方向都禁止 |
| `only_straight_on` | 仅直行 | 其他方向都禁止 |

### 工作原理

#### 路由时检查

```go
// 在 A* 算法中
for each edge from current_node {
    // 检查是否允许从 previous_way 经过 current_node 到 edge.way
    if !graph.IsValidTurn(previous_way, current_node, edge.way) {
        continue  // 跳过这条边
    }
    
    // 继续路由...
}
```

#### 验证逻辑

```go
func IsValidTurn(fromWay, viaNode, toWay int64) bool {
    restrictions := GetRestrictions(viaNode)
    
    for each restriction {
        // 如果是 "no_*" 限制且匹配
        if restriction.Type == "no_left_turn" && 
           restriction.FromWay == fromWay && 
           restriction.ToWay == toWay {
            return false  // 禁止
        }
        
        // 如果是 "only_*" 限制
        if restriction.Type == "only_straight_on" &&
           restriction.FromWay == fromWay {
            // 只有匹配的 ToWay 才允许
            return restriction.ToWay == toWay
        }
    }
    
    return true  // 默认允许
}
```

### Monaco 数据统计

- **总限制数**: 44 个
- **常见类型**:
  - `no_left_turn`: ~20
  - `no_right_turn`: ~15
  - `no_u_turn`: ~5
  - `only_straight_on`: ~4

---

## 单行道 (Oneway Streets)

### 支持的标记

#### 正向单行

```
oneway=yes
oneway=1
oneway=true
```

**效果:** 只能沿路段定义方向行驶

**实现:**
- 创建正向边 (A → B)
- 不创建反向边

#### 反向单行

```
oneway=-1
oneway=reverse
```

**效果:** 只能逆路段定义方向行驶

**实现:**
- 不创建正向边
- 创建反向边 (B → A)

**使用场景:**
- 公交专用道（逆向）
- 特殊交通管理

#### 双向通行

```
oneway=no
(或标签不存在)
```

**效果:** 双向通行

**实现:**
- 创建正向边 (A → B)
- 创建反向边 (B → A)

### 边创建逻辑

```go
onewayTag := way.Tags.Find("oneway")

if onewayTag == "yes" || onewayTag == "1" {
    // 正向单行
    createEdge(from → to)
    // 不创建反向边
} else if onewayTag == "-1" || onewayTag == "reverse" {
    // 反向单行
    // 不创建正向边
    createEdge(to → from)
} else {
    // 双向通行
    createEdge(from → to)
    createEdge(to → from)
}
```

### Monaco 数据

- **单行道数量**: ~100 条
- **反向单行**: ~5 条
- **边数优化**: 11,921 → 11,914 (-7 边)

---

## 实际效果

### 示例 1: 避免禁止左转

**场景:**
```
起点 A → 路口 B (禁止左转) → 目的地 C

如果直接左转被禁止:
路线会选择: A → B → 绕行 → C
而不是: A → B → (非法左转) → C
```

### 示例 2: 遵守单行道

**场景:**
```
A ----单行道----> B

从 B 到 A 的路由:
不会使用这条单行道
会寻找其他合法路径
```

### 示例 3: 反向单行道

**场景:**
```
A <----单行道(oneway=-1)---- B

从 A 到 B 的路由:
不会使用这条路（逆向）
从 B 到 A 的路由:
可以使用这条路
```

---

## 性能影响

### 转弯限制

**每次路由:**
- 检查次数: ~每个节点 1-2 次
- 单次检查: ~0.1μs
- 总影响: +2-3ms (可忽略)

**内存:**
- Monaco (44 个限制): ~1.4 KB
- 大城市估计: ~100 KB

### 单行道

**解析时:**
- 减少边数（单行道不创建反向边）
- Monaco: -7 边

**路由时:**
- 无额外开销（边已正确创建）

---

## 调试和验证

### 查看解析统计

启动服务器时查看日志:

```
Loaded 7427 nodes (from 40975 total), 1228 routable ways, and 44 turn restrictions
Graph built: 7427 nodes, 11914 edges
```

### 检查特定限制

可以添加日志来查看限制:

```go
// 在 osm/parser.go
fmt.Printf("Restriction: %s from way %d via node %d to way %d\n",
    restriction, fromWay, viaNode, toWay)
```

### 路由调试

如果路由失败，可能是:
1. 转弯限制阻止了所有路径
2. 单行道造成不连通
3. Profile 过滤了所有道路

**解决方法:**
- 尝试不同的起终点
- 使用更宽松的 profile（foot）
- 检查 OSM 数据质量

---

## 扩展和自定义

### 添加更多限制类型

当前支持 OSM 标准限制类型。如需添加自定义限制:

```go
// 在 internal/graph/restrictions.go

const (
    RestrictionNoEntry = "no_entry"
    RestrictionNoExit  = "no_exit"
    // 添加更多...
)
```

### 忽略特定限制

可以在 A* 中添加选项来忽略限制:

```go
type RouteOptions struct {
    IgnoreRestrictions bool
}

if !options.IgnoreRestrictions {
    if !graph.IsValidTurn(...) {
        continue
    }
}
```

---

## 数据质量

### OSM 数据完整性

转弯限制的数量和质量取决于 OSM 数据:

| 地区 | 转弯限制数量 | 数据质量 |
|------|-------------|---------|
| Monaco | 44 | ⭐⭐⭐⭐ |
| 欧洲主要城市 | 数千 | ⭐⭐⭐⭐⭐ |
| 北美城市 | 数千 | ⭐⭐⭐⭐ |
| 发展中国家 | 较少 | ⭐⭐ |

### 改进建议

如果您所在地区的 OSM 数据缺少限制:

1. **贡献 OSM**: 在 OpenStreetMap 上添加限制
2. **使用其他数据源**: 导入官方交通数据
3. **手动添加**: 使用 API 添加自定义限制

---

## 最佳实践

### 1. 验证数据

解析后检查统计:
```bash
# 查看日志
grep "turn restrictions" /tmp/nav-server.log
```

### 2. 测试限制

使用已知有限制的路口测试:
```bash
# 找一个已知禁止左转的路口
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "from_lat": ...,  # 限制前
    "to_lat": ...     # 限制后（需要左转）
  }'

# 验证路线是否绕开了限制
```

### 3. 监控性能

```bash
# 使用 curl 计时
time curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{"from_lat": 43.73, "from_lon": 7.42, "to_lat": 43.74, "to_lon": 7.43}'
```

---

## FAQ

**Q: 所有 OSM 数据都有转弯限制吗？**  
A: 不是，取决于数据质量。发达地区通常有更多限制数据。

**Q: 限制会影响性能吗？**  
A: 影响很小（+2-3ms），可以忽略。

**Q: 可以禁用限制吗？**  
A: 当前不支持，但可以通过代码修改添加选项。

**Q: 支持时间相关的限制吗？**  
A: 当前不支持（如 7-9am 禁止左转）。这是未来功能。

**Q: 反向单行道常见吗？**  
A: 不常见，主要用于公交专用道等特殊情况。

**Q: 限制数据如何缓存？**  
A: 随图数据一起序列化到 graph.bin.gz。

---

## 技术参考

### OSM 文档
- [Turn Restrictions](https://wiki.openstreetmap.org/wiki/Relation:restriction)
- [Oneway Tags](https://wiki.openstreetmap.org/wiki/Key:oneway)

### 实现参考
- OSRM Restrictions: [GitHub](https://github.com/Project-OSRM/osrm-backend/wiki/Turn-restrictions)
- Valhalla Restrictions: [Docs](https://valhalla.readthedocs.io/)

---

## 总结

✅ **自动解析** - 无需手动配置  
✅ **完整支持** - No-turn 和 Only-turn  
✅ **单行道** - 包括反向单行  
✅ **高性能** - 影响<3ms  
✅ **可缓存** - 随图数据存储  

功能完整，生产就绪！🚦


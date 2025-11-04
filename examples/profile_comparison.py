#!/usr/bin/env python3
"""
对比不同路由配置的结果
Compare routing profiles
"""

import requests
import json

BASE_URL = "http://localhost:8080"

def find_route(from_lat, from_lon, to_lat, to_lon, profile='car', format='geojson'):
    """查找路线"""
    url = f"{BASE_URL}/route"
    payload = {
        "from_lat": from_lat,
        "from_lon": from_lon,
        "to_lat": to_lat,
        "to_lon": to_lon,
        "profile": profile,
        "format": format
    }
    
    response = requests.post(url, json=payload)
    response.raise_for_status()
    return response.json()

def compare_profiles(from_coords, to_coords):
    """对比三种配置"""
    print("=" * 60)
    print("  路由配置对比 / Routing Profile Comparison")
    print("=" * 60)
    print(f"\n起点: {from_coords}")
    print(f"终点: {to_coords}\n")
    print("-" * 60)
    
    profiles = ['car', 'bike', 'foot']
    results = {}
    
    for profile in profiles:
        try:
            route = find_route(
                from_coords[0], from_coords[1],
                to_coords[0], to_coords[1],
                profile=profile
            )
            
            if route['routes']:
                r = route['routes'][0]
                results[profile] = {
                    'distance': r['distance'],
                    'duration': r['duration'],
                    'points': len(r['geometry']['coordinates'])
                }
                
                print(f"\n{profile.upper():6} Profile:")
                print(f"  距离: {r['distance']:,.2f} 米")
                print(f"  时间: {r['duration']:,.2f} 秒 ({r['duration']/60:.1f} 分钟)")
                print(f"  路径点: {len(r['geometry']['coordinates'])} 个")
                print(f"  平均速度: {r['distance']/r['duration']*3.6:.1f} km/h")
        except Exception as e:
            print(f"\n{profile.upper():6} Profile: ERROR - {e}")
    
    print("\n" + "=" * 60)
    print("  对比总结")
    print("=" * 60)
    
    if len(results) >= 2:
        print("\n距离对比:")
        for profile in profiles:
            if profile in results:
                print(f"  {profile:6}: {results[profile]['distance']:,.0f}m")
        
        print("\n时间对比:")
        for profile in profiles:
            if profile in results:
                mins = results[profile]['duration'] / 60
                print(f"  {profile:6}: {mins:.1f} 分钟")
    
    print("\n" + "=" * 60)

if __name__ == "__main__":
    # Monaco 测试坐标
    from_coords = (43.73, 7.42)
    to_coords = (43.74, 7.43)
    
    compare_profiles(from_coords, to_coords)
    
    print("\n提示: 不同配置可能选择不同的路径")
    print("- 汽车: 优先快速道路")
    print("- 自行车: 优先自行车道，避免大路")
    print("- 步行: 可以走捷径（人行道、楼梯）")


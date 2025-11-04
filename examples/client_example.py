#!/usr/bin/env python3
"""
Example Python client for the Navigation Service API
"""

import requests
import json

BASE_URL = "http://localhost:8080"

def find_route(from_lat, from_lon, to_lat, to_lon, alternatives=0):
    """Find a route between two points"""
    url = f"{BASE_URL}/route"
    payload = {
        "from_lat": from_lat,
        "from_lon": from_lon,
        "to_lat": to_lat,
        "to_lon": to_lon,
    }
    
    if alternatives > 0:
        payload["alternatives"] = alternatives
    
    response = requests.post(url, json=payload)
    response.raise_for_status()
    return response.json()

def update_weight(osm_way_id, multiplier):
    """Update edge weights for a specific OSM way"""
    url = f"{BASE_URL}/weight/update"
    payload = {
        "osm_way_id": osm_way_id,
        "multiplier": multiplier,
    }
    
    response = requests.post(url, json=payload)
    response.raise_for_status()
    return response.json()

def health_check():
    """Check service health"""
    url = f"{BASE_URL}/health"
    response = requests.get(url)
    response.raise_for_status()
    return response.json()

if __name__ == "__main__":
    # Example 1: Health check
    print("=== Health Check ===")
    health = health_check()
    print(json.dumps(health, indent=2))
    print()
    
    # Example 2: Find a single route
    print("=== Find Single Route ===")
    route = find_route(43.73, 7.42, 43.74, 7.43)
    if route["routes"]:
        r = route["routes"][0]
        print(f"Distance: {r['distance']:.2f} meters")
        print(f"Duration: {r['duration']:.2f} seconds")
        print(f"Points: {len(r['geometry'])}")
    print()
    
    # Example 3: Find alternative routes
    print("=== Find Alternative Routes ===")
    routes = find_route(43.73, 7.42, 43.74, 7.43, alternatives=2)
    print(f"Found {len(routes['routes'])} routes:")
    for i, r in enumerate(routes["routes"], 1):
        print(f"  Route {i}: {r['distance']:.2f}m, {r['duration']:.2f}s")
    print()
    
    # Example 4: Update weights (simulate traffic)
    print("=== Update Weights ===")
    result = update_weight(123456789, 2.5)
    print(f"Updated {result['edges_updated']} edges")


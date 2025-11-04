package encoding

// GeoJSONGeometry represents a GeoJSON geometry object
type GeoJSONGeometry struct {
	Type        string        `json:"type"`
	Coordinates [][2]float64  `json:"coordinates"`
}

// GeoJSONFeature represents a GeoJSON feature
type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	Geometry   GeoJSONGeometry        `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

// GeoJSONFeatureCollection represents a GeoJSON feature collection
type GeoJSONFeatureCollection struct {
	Type     string            `json:"type"`
	Features []GeoJSONFeature  `json:"features"`
}

// NewLineStringGeometry creates a GeoJSON LineString geometry
func NewLineStringGeometry(coordinates [][2]float64) GeoJSONGeometry {
	return GeoJSONGeometry{
		Type:        "LineString",
		Coordinates: coordinates,
	}
}

// NewRouteFeature creates a GeoJSON feature for a route
func NewRouteFeature(coordinates [][2]float64, distance, duration float64) GeoJSONFeature {
	return GeoJSONFeature{
		Type:     "Feature",
		Geometry: NewLineStringGeometry(coordinates),
		Properties: map[string]interface{}{
			"distance": distance,
			"duration": duration,
		},
	}
}

// NewFeatureCollection creates a GeoJSON feature collection
func NewFeatureCollection(features []GeoJSONFeature) GeoJSONFeatureCollection {
	return GeoJSONFeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}
}


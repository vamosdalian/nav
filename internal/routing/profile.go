package routing

// RoutingProfile defines routing behavior for different transportation modes
type RoutingProfile struct {
	Name            string
	AllowedHighways map[string]bool  // Which road types are allowed
	SpeedFactors    map[string]float64 // Speed multipliers for different road types
	AvoidSurfaces   map[string]bool    // Surfaces to avoid
	MaxSpeed        float64            // Maximum speed (m/s)
}

// Predefined routing profiles
var (
	// CarProfile - Standard car routing
	CarProfile = RoutingProfile{
		Name: "car",
		AllowedHighways: map[string]bool{
			"motorway":    true,
			"trunk":       true,
			"primary":     true,
			"secondary":   true,
			"tertiary":    true,
			"unclassified": true,
			"residential": true,
			"service":     true,
			"motorway_link": true,
			"trunk_link":    true,
			"primary_link":  true,
			"secondary_link": true,
		},
		SpeedFactors: map[string]float64{
			"motorway":    1.2,  // 20% faster on highways
			"trunk":       1.1,
			"primary":     1.0,
			"secondary":   0.95,
			"tertiary":    0.9,
			"residential": 0.8,
			"service":     0.7,
		},
		AvoidSurfaces: map[string]bool{},
		MaxSpeed:      33.33, // ~120 km/h
	}

	// BikeProfile - Bicycle routing
	BikeProfile = RoutingProfile{
		Name: "bike",
		AllowedHighways: map[string]bool{
			"cycleway":    true,
			"path":        true,
			"footway":     true,
			"track":       true,
			"primary":     true,
			"secondary":   true,
			"tertiary":    true,
			"residential": true,
			"service":     true,
			"unclassified": true,
		},
		SpeedFactors: map[string]float64{
			"cycleway":    1.2,  // Prefer dedicated bike paths
			"path":        1.1,
			"residential": 1.0,
			"secondary":   0.9,
			"primary":     0.7,  // Less desirable
			"service":     0.95,
		},
		AvoidSurfaces: map[string]bool{
			"gravel": true,
			"sand":   true,
		},
		MaxSpeed: 8.33, // ~30 km/h
	}

	// FootProfile - Pedestrian routing
	FootProfile = RoutingProfile{
		Name: "foot",
		AllowedHighways: map[string]bool{
			"footway":     true,
			"path":        true,
			"steps":       true,
			"pedestrian":  true,
			"residential": true,
			"service":     true,
			"track":       true,
			"cycleway":    true,
			"primary":     true,
			"secondary":   true,
			"tertiary":    true,
			"unclassified": true,
		},
		SpeedFactors: map[string]float64{
			"footway":     1.2,
			"pedestrian":  1.2,
			"path":        1.1,
			"residential": 1.0,
			"service":     1.0,
			"steps":       0.8,  // Slower on stairs
			"primary":     0.7,  // Less comfortable
		},
		AvoidSurfaces: map[string]bool{},
		MaxSpeed:      1.4, // ~5 km/h walking speed
	}
)

// GetProfile returns a routing profile by name
func GetProfile(name string) RoutingProfile {
	switch name {
	case "bike", "bicycle":
		return BikeProfile
	case "foot", "walk", "pedestrian":
		return FootProfile
	case "car", "driving":
		return CarProfile
	default:
		return CarProfile // Default to car
	}
}

// IsAllowed checks if a highway type is allowed for this profile
func (p *RoutingProfile) IsAllowed(highway string) bool {
	return p.AllowedHighways[highway]
}

// GetSpeedFactor returns the speed factor for a highway type
func (p *RoutingProfile) GetSpeedFactor(highway string) float64 {
	if factor, exists := p.SpeedFactors[highway]; exists {
		return factor
	}
	return 1.0 // Default factor
}

// ShouldAvoidSurface checks if a surface should be avoided
func (p *RoutingProfile) ShouldAvoidSurface(surface string) bool {
	return p.AvoidSurfaces[surface]
}

// CalculateWeight calculates edge weight based on profile
func (p *RoutingProfile) CalculateWeight(distance float64, highway string, surface string) float64 {
	// Base weight is distance
	weight := distance
	
	// Apply speed factor
	speedFactor := p.GetSpeedFactor(highway)
	weight = weight / speedFactor
	
	// Penalize avoided surfaces
	if p.ShouldAvoidSurface(surface) {
		weight *= 2.0 // Double the cost
	}
	
	return weight
}


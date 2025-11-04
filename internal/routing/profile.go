package routing

// ProfileConfig represents a complete routing profile configuration
type ProfileConfig struct {
	Name          string                   `yaml:"name" json:"name"`
	Description   string                   `yaml:"description" json:"description"`
	Version       string                   `yaml:"version" json:"version"`
	Settings      Settings                 `yaml:"settings" json:"settings"`
	Highways      map[string]HighwayConfig `yaml:"highways" json:"highways"`
	Surfaces      map[string]SurfaceConfig `yaml:"surfaces" json:"surfaces"`
	Features      Features                 `yaml:"features" json:"features"`
	WeightFormula WeightFormula            `yaml:"weight_formula" json:"weight_formula"`
}

// Settings contains basic routing settings
type Settings struct {
	MaxSpeedKmh     float64 `yaml:"max_speed_kmh" json:"max_speed_kmh"`
	DefaultSpeedKmh float64 `yaml:"default_speed_kmh" json:"default_speed_kmh"`
}

// HighwayConfig defines configuration for a highway type
type HighwayConfig struct {
	Allowed     bool    `yaml:"allowed" json:"allowed"`
	SpeedFactor float64 `yaml:"speed_factor" json:"speed_factor"`
	Preference  float64 `yaml:"preference" json:"preference"`
}

// SurfaceConfig defines configuration for a surface type
type SurfaceConfig struct {
	Penalty float64 `yaml:"penalty" json:"penalty"`
}

// Features contains routing feature flags
type Features struct {
	AvoidTolls    bool `yaml:"avoid_tolls" json:"avoid_tolls"`
	AvoidHighways bool `yaml:"avoid_highways" json:"avoid_highways"`
	AvoidFerries  bool `yaml:"avoid_ferries" json:"avoid_ferries"`
	AvoidTunnels  bool `yaml:"avoid_tunnels" json:"avoid_tunnels"`
	AllowUturns   bool `yaml:"allow_uturns" json:"allow_uturns"`
}

// WeightFormula defines how edge weights are calculated
type WeightFormula struct {
	UseTime        bool    `yaml:"use_time" json:"use_time"`
	DistanceWeight float64 `yaml:"distance_weight" json:"distance_weight"`
	TimeWeight     float64 `yaml:"time_weight" json:"time_time_weight"`
}

// Legacy RoutingProfile for backward compatibility
type RoutingProfile struct {
	Name            string
	AllowedHighways map[string]bool
	SpeedFactors    map[string]float64
	AvoidSurfaces   map[string]bool
	MaxSpeed        float64
}

// Predefined routing profiles
var (
	// CarProfile - Standard car routing
	CarProfile = RoutingProfile{
		Name: "car",
		AllowedHighways: map[string]bool{
			"motorway":       true,
			"trunk":          true,
			"primary":        true,
			"secondary":      true,
			"tertiary":       true,
			"unclassified":   true,
			"residential":    true,
			"service":        true,
			"motorway_link":  true,
			"trunk_link":     true,
			"primary_link":   true,
			"secondary_link": true,
		},
		SpeedFactors: map[string]float64{
			"motorway":    1.2, // 20% faster on highways
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
			"cycleway":     true,
			"path":         true,
			"footway":      true,
			"track":        true,
			"primary":      true,
			"secondary":    true,
			"tertiary":     true,
			"residential":  true,
			"service":      true,
			"unclassified": true,
		},
		SpeedFactors: map[string]float64{
			"cycleway":    1.2, // Prefer dedicated bike paths
			"path":        1.1,
			"residential": 1.0,
			"secondary":   0.9,
			"primary":     0.7, // Less desirable
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
			"footway":      true,
			"path":         true,
			"steps":        true,
			"pedestrian":   true,
			"residential":  true,
			"service":      true,
			"track":        true,
			"cycleway":     true,
			"primary":      true,
			"secondary":    true,
			"tertiary":     true,
			"unclassified": true,
		},
		SpeedFactors: map[string]float64{
			"footway":     1.2,
			"pedestrian":  1.2,
			"path":        1.1,
			"residential": 1.0,
			"service":     1.0,
			"steps":       0.8, // Slower on stairs
			"primary":     0.7, // Less comfortable
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

// ProfileConfig methods

// Clone creates a deep copy of the profile
func (p *ProfileConfig) Clone() *ProfileConfig {
	clone := &ProfileConfig{
		Name:          p.Name,
		Description:   p.Description,
		Version:       p.Version,
		Settings:      p.Settings,
		Features:      p.Features,
		WeightFormula: p.WeightFormula,
	}

	// Deep copy maps
	clone.Highways = make(map[string]HighwayConfig)
	for k, v := range p.Highways {
		clone.Highways[k] = v
	}

	clone.Surfaces = make(map[string]SurfaceConfig)
	for k, v := range p.Surfaces {
		clone.Surfaces[k] = v
	}

	return clone
}

// IsHighwayAllowed checks if a highway type is allowed
func (p *ProfileConfig) IsHighwayAllowed(highway string) bool {
	if config, exists := p.Highways[highway]; exists {
		return config.Allowed
	}
	return false
}

// GetHighwayConfig returns the configuration for a highway type
func (p *ProfileConfig) GetHighwayConfig(highway string) (HighwayConfig, bool) {
	config, exists := p.Highways[highway]
	return config, exists
}

// GetSurfaceConfig returns the configuration for a surface type
func (p *ProfileConfig) GetSurfaceConfig(surface string) (SurfaceConfig, bool) {
	config, exists := p.Surfaces[surface]
	return config, exists
}

// GetEffectiveSpeed calculates the effective speed for a road segment
func (p *ProfileConfig) GetEffectiveSpeed(maxSpeed float64, highway string) float64 {
	// Start with default speed
	speed := p.Settings.DefaultSpeedKmh

	// Use edge's max speed if available
	if maxSpeed > 0 {
		speed = maxSpeed * 3.6 // Convert m/s to km/h
	}

	// Apply highway speed factor
	if hwConfig, exists := p.Highways[highway]; exists {
		speed *= hwConfig.SpeedFactor
	}

	// Cap at profile max speed
	if speed > p.Settings.MaxSpeedKmh {
		speed = p.Settings.MaxSpeedKmh
	}

	return speed / 3.6 // Convert back to m/s
}

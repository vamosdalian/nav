package routing

// RouteOptions contains optional parameters that can override profile defaults
type RouteOptions struct {
	// Feature overrides
	AvoidTolls    *bool `json:"avoid_tolls,omitempty"`
	AvoidHighways *bool `json:"avoid_highways,omitempty"`
	AvoidFerries  *bool `json:"avoid_ferries,omitempty"`
	AvoidTunnels  *bool `json:"avoid_tunnels,omitempty"`
	AllowUturns   *bool `json:"allow_uturns,omitempty"`

	// Speed overrides
	MaxSpeed *float64 `json:"max_speed,omitempty"` // km/h
}

// ApplyOptions applies route options to a profile (modifies the profile)
func (p *ProfileConfig) ApplyOptions(opts *RouteOptions) {
	if opts == nil {
		return
	}

	// Apply feature overrides
	if opts.AvoidTolls != nil {
		p.Features.AvoidTolls = *opts.AvoidTolls
	}
	if opts.AvoidHighways != nil {
		p.Features.AvoidHighways = *opts.AvoidHighways
	}
	if opts.AvoidFerries != nil {
		p.Features.AvoidFerries = *opts.AvoidFerries
	}
	if opts.AvoidTunnels != nil {
		p.Features.AvoidTunnels = *opts.AvoidTunnels
	}
	if opts.AllowUturns != nil {
		p.Features.AllowUturns = *opts.AllowUturns
	}

	// Apply speed override
	if opts.MaxSpeed != nil && *opts.MaxSpeed > 0 {
		p.Settings.MaxSpeedKmh = *opts.MaxSpeed
	}
}

// GetEffectiveProfile returns a profile with options applied
// This creates a copy of the base profile to avoid modifying the original
func GetEffectiveProfile(baseProfile *ProfileConfig, opts *RouteOptions) *ProfileConfig {
	if opts == nil {
		return baseProfile
	}

	// Clone profile to avoid modifying the original
	effective := baseProfile.Clone()

	// Apply options
	effective.ApplyOptions(opts)

	return effective
}

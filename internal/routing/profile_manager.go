package routing

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// ProfileManager manages routing profiles
type ProfileManager struct {
	profiles  map[string]*ProfileConfig
	mutex     sync.RWMutex
	configDir string
}

// NewProfileManager creates a new profile manager
func NewProfileManager(configDir string) *ProfileManager {
	return &ProfileManager{
		profiles:  make(map[string]*ProfileConfig),
		configDir: configDir,
	}
}

// LoadProfiles loads all profiles from config directory
func (pm *ProfileManager) LoadProfiles() error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Require profile directory to be set
	if pm.configDir == "" {
		return fmt.Errorf("profile directory is required, please set a valid profile directory")
	}

	// Check if directory exists
	if _, err := os.Stat(pm.configDir); os.IsNotExist(err) {
		return fmt.Errorf("profile directory '%s' does not exist", pm.configDir)
	}

	// Load from config files
	if err := pm.loadFromDirectory(); err != nil {
		return fmt.Errorf("failed to load profiles from directory: %w", err)
	}

	// Check if at least one profile was loaded
	if len(pm.profiles) == 0 {
		return fmt.Errorf("no valid profiles found in directory '%s'", pm.configDir)
	}

	log.Printf("Loaded %d profile(s): %v", len(pm.profiles), pm.listProfileNames())
	return nil
}

// loadFromDirectory loads all YAML files from the config directory
func (pm *ProfileManager) loadFromDirectory() error {
	files, err := filepath.Glob(filepath.Join(pm.configDir, "*.yaml"))
	if err != nil {
		return fmt.Errorf("failed to list profile files: %w", err)
	}

	loadedCount := 0
	for _, file := range files {
		if err := pm.loadProfileFromFile(file); err != nil {
			log.Printf("Warning: failed to load profile %s: %v", file, err)
		} else {
			log.Printf("Loaded profile from: %s", file)
			loadedCount++
		}
	}

	if loadedCount == 0 {
		return fmt.Errorf("no valid YAML files found in %s", pm.configDir)
	}

	return nil
}

// loadProfileFromFile loads a single profile from a YAML file
func (pm *ProfileManager) loadProfileFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var profile ProfileConfig
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate profile
	if err := pm.validateProfile(&profile); err != nil {
		return fmt.Errorf("invalid profile: %w", err)
	}

	pm.profiles[profile.Name] = &profile
	return nil
}

// validateProfile validates profile configuration
func (pm *ProfileManager) validateProfile(p *ProfileConfig) error {
	if p.Name == "" {
		return fmt.Errorf("profile name is required")
	}

	if p.Settings.MaxSpeedKmh <= 0 {
		return fmt.Errorf("max_speed_kmh must be positive")
	}

	if p.Settings.DefaultSpeedKmh <= 0 {
		return fmt.Errorf("default_speed_kmh must be positive")
	}

	// Validate weight formula
	if p.WeightFormula.UseTime {
		total := p.WeightFormula.DistanceWeight + p.WeightFormula.TimeWeight
		if total < 0.99 || total > 1.01 {
			return fmt.Errorf("distance_weight + time_weight must equal 1.0 (got %.2f)", total)
		}
	}

	return nil
}

// GetProfile retrieves a profile by name
func (pm *ProfileManager) GetProfile(name string) (*ProfileConfig, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	profile, exists := pm.profiles[name]
	if !exists {
		return nil, fmt.Errorf("profile '%s' not found", name)
	}

	return profile, nil
}

// ListProfiles returns all available profile names
func (pm *ProfileManager) ListProfiles() []string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	return pm.listProfileNames()
}

// listProfileNames returns profile names (must be called with lock held)
func (pm *ProfileManager) listProfileNames() []string {
	names := make([]string, 0, len(pm.profiles))
	for name := range pm.profiles {
		names = append(names, name)
	}
	return names
}

// Reload reloads all profiles from disk
func (pm *ProfileManager) Reload() error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Clear existing profiles
	pm.profiles = make(map[string]*ProfileConfig)

	// Require profile directory
	if pm.configDir == "" {
		return fmt.Errorf("profile directory is required")
	}

	// Reload from files
	if err := pm.loadFromDirectory(); err != nil {
		return fmt.Errorf("failed to reload profiles: %w", err)
	}

	// Check if at least one profile was loaded
	if len(pm.profiles) == 0 {
		return fmt.Errorf("no valid profiles found after reload")
	}

	log.Printf("Reloaded %d profile(s)", len(pm.profiles))
	return nil
}

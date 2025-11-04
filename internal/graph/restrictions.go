package graph

// TurnRestriction represents a turn restriction in the road network
type TurnRestriction struct {
	FromWay int64  // OSM way ID where the turn starts
	ViaNode int64  // Node where the turn happens
	ToWay   int64  // OSM way ID where the turn ends
	Type    string // Type of restriction: "no_left_turn", "no_right_turn", "no_u_turn", "only_straight_on", etc.
}

// RestrictionType constants
const (
	RestrictionNoLeftTurn    = "no_left_turn"
	RestrictionNoRightTurn   = "no_right_turn"
	RestrictionNoUTurn       = "no_u_turn"
	RestrictionNoStraightOn  = "no_straight_on"
	RestrictionOnlyLeftTurn  = "only_left_turn"
	RestrictionOnlyRightTurn = "only_right_turn"
	RestrictionOnlyStraightOn = "only_straight_on"
)

// AddRestriction adds a turn restriction to the graph
func (g *Graph) AddRestriction(restriction TurnRestriction) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	if g.restrictions == nil {
		g.restrictions = make(map[int64][]TurnRestriction)
	}
	
	// Index by via node for fast lookup during routing
	g.restrictions[restriction.ViaNode] = append(g.restrictions[restriction.ViaNode], restriction)
}

// GetRestrictions returns all turn restrictions at a node
func (g *Graph) GetRestrictions(nodeID int64) []TurnRestriction {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	if g.restrictions == nil {
		return nil
	}
	
	return g.restrictions[nodeID]
}

// IsValidTurn checks if a turn from one way to another is allowed
func (g *Graph) IsValidTurn(fromWayID, viaNodeID, toWayID int64) bool {
	restrictions := g.GetRestrictions(viaNodeID)
	
	if len(restrictions) == 0 {
		return true // No restrictions, turn is allowed
	}
	
	hasOnlyRestriction := false
	isExplicitlyAllowed := false
	
	for _, r := range restrictions {
		// Check if this restriction applies to our turn
		if r.FromWay == fromWayID {
			// Handle "only_*" restrictions
			if r.Type == RestrictionOnlyLeftTurn || r.Type == RestrictionOnlyRightTurn || r.Type == RestrictionOnlyStraightOn {
				hasOnlyRestriction = true
				if r.ToWay == toWayID {
					isExplicitlyAllowed = true
				}
			}
			
			// Handle "no_*" restrictions
			if (r.Type == RestrictionNoLeftTurn || r.Type == RestrictionNoRightTurn || 
			    r.Type == RestrictionNoUTurn || r.Type == RestrictionNoStraightOn) && r.ToWay == toWayID {
				return false // Explicitly forbidden
			}
		}
	}
	
	// If there's an "only" restriction, the turn must be explicitly allowed
	if hasOnlyRestriction {
		return isExplicitlyAllowed
	}
	
	return true // No applicable restrictions, turn is allowed
}


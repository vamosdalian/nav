package routing

import (
	"container/heap"
	"fmt"

	"github.com/vamosdalian/nav/internal/graph"
)

// FindRouteBidirectional finds a route using bidirectional A* search
// This searches from both start and end simultaneously, meeting in the middle
// Typically 2-3x faster than unidirectional A* for long distances
func (r *Router) FindRouteBidirectional(fromLat, fromLon, toLat, toLon float64) (*Route, error) {
	return r.FindRouteBidirectionalWithProfile(fromLat, fromLon, toLat, toLon, r.profile)
}

// FindRouteBidirectionalWithProfile finds a route using bidirectional search with a specific profile
func (r *Router) FindRouteBidirectionalWithProfile(fromLat, fromLon, toLat, toLon float64, profile RoutingProfile) (*Route, error) {
	// Find nearest nodes to start and end coordinates
	startNode, err := r.graph.FindNearestNode(fromLat, fromLon)
	if err != nil {
		return nil, fmt.Errorf("cannot find start node: %w", err)
	}

	endNode, err := r.graph.FindNearestNode(toLat, toLon)
	if err != nil {
		return nil, fmt.Errorf("cannot find end node: %w", err)
	}

	if startNode.ID == endNode.ID {
		return &Route{
			Nodes:    []int64{startNode.ID},
			Distance: 0,
			Duration: 0,
		}, nil
	}

	// Temporarily set profile
	oldProfile := r.profile
	r.profile = profile
	defer func() { r.profile = oldProfile }()

	return r.bidirectionalAStar(startNode.ID, endNode.ID)
}

func (r *Router) bidirectionalAStar(start, end int64) (*Route, error) {
	startNode, _ := r.graph.GetNode(start)
	endNode, _ := r.graph.GetNode(end)

	// Simplified bidirectional search (without turn restrictions for performance)
	// Forward search structures
	forwardOpenSet := &priorityQueue{}
	heap.Init(forwardOpenSet)
	forwardCameFrom := make(map[int64]int64)
	forwardGScore := make(map[int64]float64)
	forwardClosed := make(map[int64]bool)

	// Backward search structures
	backwardOpenSet := &priorityQueue{}
	heap.Init(backwardOpenSet)
	backwardCameFrom := make(map[int64]int64)
	backwardGScore := make(map[int64]float64)
	backwardClosed := make(map[int64]bool)

	// Initialize
	forwardGScore[start] = 0
	backwardGScore[end] = 0

	hStart := graph.HaversineDistance(startNode.Lat, startNode.Lon, endNode.Lat, endNode.Lon)

	heap.Push(forwardOpenSet, &item{
		nodeID:   start,
		priority: hStart,
		gScore:   0,
	})

	heap.Push(backwardOpenSet, &item{
		nodeID:   end,
		priority: hStart,
		gScore:   0,
	})

	// Track best meeting point
	bestDistance := float64(1e9)
	var meetingNode int64

	maxIterations := 100000
	iterations := 0

	for forwardOpenSet.Len() > 0 && backwardOpenSet.Len() > 0 && iterations < maxIterations {
		iterations++

		// Alternate between forward and backward search
		if iterations%2 == 0 {
			// Forward step
			if forwardOpenSet.Len() > 0 {
				current := heap.Pop(forwardOpenSet).(*item)

				if forwardClosed[current.nodeID] {
					continue
				}
				forwardClosed[current.nodeID] = true

				// Check if backward search has reached this node
				if backDist, exists := backwardGScore[current.nodeID]; exists {
					totalDist := forwardGScore[current.nodeID] + backDist
					if totalDist < bestDistance {
						bestDistance = totalDist
						meetingNode = current.nodeID
					}
				}

				// Expand forward
				edges := r.graph.GetEdges(current.nodeID)
				for _, edge := range edges {
					if forwardClosed[edge.To] {
						continue
					}

					// Check profile
					highway := edge.Tags["highway"]
					if !r.profile.IsAllowed(highway) {
						continue
					}

					surface := edge.Tags["surface"]
					weight := r.profile.CalculateWeight(edge.Weight, highway, surface)

					tentativeGScore := forwardGScore[current.nodeID] + weight

					if currentGScore, exists := forwardGScore[edge.To]; !exists || tentativeGScore < currentGScore {
						forwardCameFrom[edge.To] = current.nodeID
						forwardGScore[edge.To] = tentativeGScore

						neighbor, _ := r.graph.GetNode(edge.To)
						h := graph.HaversineDistance(neighbor.Lat, neighbor.Lon, endNode.Lat, endNode.Lon)
						fScore := tentativeGScore + h

						heap.Push(forwardOpenSet, &item{
							nodeID:   edge.To,
							priority: fScore,
							gScore:   tentativeGScore,
						})
					}
				}
			}
		} else {
			// Backward step
			if backwardOpenSet.Len() > 0 {
				current := heap.Pop(backwardOpenSet).(*item)

				if backwardClosed[current.nodeID] {
					continue
				}
				backwardClosed[current.nodeID] = true

				// Check if forward search has reached this node
				if fwdDist, exists := forwardGScore[current.nodeID]; exists {
					totalDist := fwdDist + backwardGScore[current.nodeID]
					if totalDist < bestDistance {
						bestDistance = totalDist
						meetingNode = current.nodeID
					}
				}

				// Expand backward (find incoming edges)
				r.expandBackward(current.nodeID, startNode, backwardOpenSet, backwardCameFrom, backwardGScore, backwardClosed)
			}
		}

		// Early termination if we found a path and searches have progressed
		if meetingNode != 0 && iterations > 50 {
			break
		}
	}

	if meetingNode == 0 {
		return nil, fmt.Errorf("no route found from %d to %d", start, end)
	}

	// Reconstruct path from both directions
	return r.reconstructBidirectionalPath(
		forwardCameFrom, backwardCameFrom,
		start, end, meetingNode,
		bestDistance,
	), nil
}

// expandBackward expands backward search using reverse adjacency list
func (r *Router) expandBackward(nodeID int64, targetNode *graph.Node,
	openSet *priorityQueue, cameFrom map[int64]int64,
	gScore map[int64]float64, closed map[int64]bool) {

	// Get incoming edges using reverse adjacency list (much faster!)
	reverseEdges := r.graph.GetReverseEdges(nodeID)

	for _, edge := range reverseEdges {
		fromNodeID := edge.From

		if closed[fromNodeID] {
			continue
		}

		// Check profile
		highway := edge.Tags["highway"]
		if !r.profile.IsAllowed(highway) {
			continue
		}

		surface := edge.Tags["surface"]
		weight := r.profile.CalculateWeight(edge.Weight, highway, surface)

		tentativeGScore := gScore[nodeID] + weight

		if currentGScore, exists := gScore[fromNodeID]; !exists || tentativeGScore < currentGScore {
			cameFrom[fromNodeID] = nodeID
			gScore[fromNodeID] = tentativeGScore

			neighbor, _ := r.graph.GetNode(fromNodeID)
			h := graph.HaversineDistance(neighbor.Lat, neighbor.Lon, targetNode.Lat, targetNode.Lon)
			fScore := tentativeGScore + h

			heap.Push(openSet, &item{
				nodeID:   fromNodeID,
				priority: fScore,
				gScore:   tentativeGScore,
			})
		}
	}
}

func (r *Router) reconstructBidirectionalPath(
	forwardCameFrom, backwardCameFrom map[int64]int64,
	start, end, meeting int64,
	distance float64) *Route {

	// Build forward path: start -> meeting
	forwardPath := []int64{meeting}
	curr := meeting
	for curr != start {
		curr = forwardCameFrom[curr]
		forwardPath = append([]int64{curr}, forwardPath...)
	}

	// Build backward path: meeting -> end
	backwardPath := []int64{}
	curr = meeting
	for curr != end {
		next := int64(0)
		for node, parent := range backwardCameFrom {
			if parent == curr {
				next = node
				break
			}
		}
		if next == 0 {
			break
		}
		backwardPath = append(backwardPath, next)
		curr = next
	}

	// Combine paths
	fullPath := append(forwardPath, backwardPath...)

	return &Route{
		Nodes:    fullPath,
		Distance: distance,
		Duration: distance / 13.89,
	}
}

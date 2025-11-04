package routing

import (
	"container/heap"
	"fmt"

	"github.com/vamosdalian/nav/internal/graph"
)

// Route represents a path from source to destination
type Route struct {
	Nodes    []int64
	Distance float64
	Duration float64
}

// Router provides routing functionality
type Router struct {
	graph *graph.Graph
}

// NewRouter creates a new router
func NewRouter(g *graph.Graph) *Router {
	return &Router{graph: g}
}

// FindRoute finds the shortest path using A* algorithm
func (r *Router) FindRoute(fromLat, fromLon, toLat, toLon float64) (*Route, error) {
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
	
	return r.astar(startNode.ID, endNode.ID)
}

// FindMultipleRoutes finds alternative routes using penalty method
func (r *Router) FindMultipleRoutes(fromLat, fromLon, toLat, toLon float64, numRoutes int) ([]*Route, error) {
	if numRoutes < 1 {
		numRoutes = 1
	}
	
	routes := make([]*Route, 0, numRoutes)
	penalizedEdges := make(map[edgeKey]float64)
	
	for i := 0; i < numRoutes; i++ {
		route, err := r.findRouteWithPenalty(fromLat, fromLon, toLat, toLon, penalizedEdges)
		if err != nil {
			if i == 0 {
				return nil, err
			}
			break // No more alternative routes found
		}
		
		// Check if route is sufficiently different
		if i > 0 && !r.isSufficientlyDifferent(route, routes) {
			break
		}
		
		routes = append(routes, route)
		
		// Penalize edges used in this route for next iteration
		for j := 0; j < len(route.Nodes)-1; j++ {
			key := edgeKey{from: route.Nodes[j], to: route.Nodes[j+1]}
			penalizedEdges[key] = 1.5 // 50% penalty
		}
	}
	
	return routes, nil
}

type edgeKey struct {
	from, to int64
}

func (r *Router) findRouteWithPenalty(fromLat, fromLon, toLat, toLon float64, penalties map[edgeKey]float64) (*Route, error) {
	startNode, err := r.graph.FindNearestNode(fromLat, fromLon)
	if err != nil {
		return nil, err
	}
	
	endNode, err := r.graph.FindNearestNode(toLat, toLon)
	if err != nil {
		return nil, err
	}
	
	return r.astarWithPenalty(startNode.ID, endNode.ID, penalties)
}

func (r *Router) astar(start, end int64) (*Route, error) {
	return r.astarWithPenalty(start, end, nil)
}

func (r *Router) astarWithPenalty(start, end int64, penalties map[edgeKey]float64) (*Route, error) {
	endNode, err := r.graph.GetNode(end)
	if err != nil {
		return nil, err
	}
	
	startNode, err := r.graph.GetNode(start)
	if err != nil {
		return nil, err
	}
	
	// Priority queue for open set
	openSet := &priorityQueue{}
	heap.Init(openSet)
	
	// Track visited nodes
	cameFrom := make(map[int64]int64)
	gScore := make(map[int64]float64)
	gScore[start] = 0
	
	// Track closed set to avoid revisiting
	closedSet := make(map[int64]bool)
	
	h := graph.HaversineDistance(startNode.Lat, startNode.Lon, endNode.Lat, endNode.Lon)
	
	heap.Push(openSet, &item{
		nodeID:   start,
		priority: h,
		gScore:   0,
	})
	
	nodesExplored := 0
	maxNodesToExplore := 100000 // Prevent infinite loops
	
	for openSet.Len() > 0 && nodesExplored < maxNodesToExplore {
		current := heap.Pop(openSet).(*item)
		
		// Skip if already processed
		if closedSet[current.nodeID] {
			continue
		}
		closedSet[current.nodeID] = true
		nodesExplored++
		
		if current.nodeID == end {
			return r.reconstructPath(cameFrom, start, end, gScore[end]), nil
		}
		
		// Explore neighbors
		edges := r.graph.GetEdges(current.nodeID)
		for _, edge := range edges {
			if closedSet[edge.To] {
				continue
			}
			
			weight := edge.Weight
			
			// Apply penalty if exists
			if penalties != nil {
				if penalty, exists := penalties[edgeKey{from: edge.From, to: edge.To}]; exists {
					weight *= penalty
				}
			}
			
			tentativeGScore := gScore[current.nodeID] + weight
			
			if currentGScore, exists := gScore[edge.To]; !exists || tentativeGScore < currentGScore {
				cameFrom[edge.To] = current.nodeID
				gScore[edge.To] = tentativeGScore
				
				neighbor, _ := r.graph.GetNode(edge.To)
				h := graph.HaversineDistance(neighbor.Lat, neighbor.Lon, endNode.Lat, endNode.Lon)
				fScore := tentativeGScore + h
				
				heap.Push(openSet, &item{
					nodeID:   edge.To,
					priority: fScore,
					gScore:   tentativeGScore,
				})
			}
		}
	}
	
	return nil, fmt.Errorf("no route found from %d to %d (explored %d nodes)", start, end, nodesExplored)
}

func (r *Router) reconstructPath(cameFrom map[int64]int64, start, end int64, distance float64) *Route {
	path := []int64{end}
	current := end
	
	for current != start {
		current = cameFrom[current]
		path = append([]int64{current}, path...)
	}
	
	return &Route{
		Nodes:    path,
		Distance: distance,
		Duration: distance / 13.89, // Assume average speed ~50 km/h (13.89 m/s)
	}
}

func (r *Router) isSufficientlyDifferent(newRoute *Route, existingRoutes []*Route) bool {
	threshold := 0.3 // Routes should share less than 30% of nodes
	
	newSet := make(map[int64]bool)
	for _, nodeID := range newRoute.Nodes {
		newSet[nodeID] = true
	}
	
	for _, existing := range existingRoutes {
		overlap := 0
		for _, nodeID := range existing.Nodes {
			if newSet[nodeID] {
				overlap++
			}
		}
		
		similarity := float64(overlap) / float64(len(existing.Nodes))
		if similarity > (1 - threshold) {
			return false
		}
	}
	
	return true
}

// Priority queue implementation for A*
type item struct {
	nodeID   int64
	priority float64
	gScore   float64
	index    int
}

type priorityQueue []*item

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].priority < pq[j].priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}


package onetomany

import (
	"github.com/ttpr0/go-routing/attr"
	"github.com/ttpr0/go-routing/geo"
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
)

func NewAvoidDijkstra(g graph.IGraph, max_range int32, att attr.IAttributes, avoid_roads Optional[[]attr.RoadType], avoid_areas Optional[geo.Feature]) *AvoidDijkstra {
	return &AvoidDijkstra{
		g:           g,
		max_range:   max_range,
		att:         att,
		avoid_roads: avoid_roads,
		avoid_areas: avoid_areas,
	}
}

type AvoidDijkstra struct {
	g           graph.IGraph
	max_range   int32
	att         attr.IAttributes
	avoid_roads Optional[[]attr.RoadType]
	avoid_areas Optional[geo.Feature]
}

func (self *AvoidDijkstra) CreateSolver() ISolver {
	node_flags := NewFlags[DistFlag](int32(self.g.NodeCount()), DistFlag{1000000})
	return &AvoidDijkstraSolver{
		g:           self.g,
		node_flags:  node_flags,
		max_range:   self.max_range,
		att:         self.att,
		avoid_roads: self.avoid_roads,
		avoid_areas: self.avoid_areas,
	}
}

type AvoidDijkstraSolver struct {
	g           graph.IGraph
	node_flags  Flags[DistFlag]
	max_range   int32
	att         attr.IAttributes
	avoid_roads Optional[[]attr.RoadType]
	avoid_areas Optional[geo.Feature]
}

// CalcDiatanceFromStarts implements ISolver.
func (self *AvoidDijkstraSolver) CalcDistanceFromStart(starts Array[Tuple[int32, int32]]) error {
	self.node_flags.Reset()
	var avoid_geom Optional[geo.Geometry]
	if self.avoid_areas.HasValue() {
		avoid_geom = Some(self.avoid_areas.Value.Geometry())
	} else {
		avoid_geom = None[geo.Geometry]()
	}
	_CalcAvoidDijkstra(self.g, starts, self.node_flags, self.max_range, self.att, self.avoid_roads, avoid_geom)
	return nil
}

// GetDistance implements ISolver.
func (self *AvoidDijkstraSolver) GetDistance(node int32) int32 {
	return self.node_flags.Get(node).Dist
}

func _CalcAvoidDijkstra(g graph.IGraph, starts Array[Tuple[int32, int32]], node_flags Flags[DistFlag], max_range int32, att attr.IAttributes, avoid_roads Optional[[]attr.RoadType], avoid_geom Optional[geo.Geometry]) {
	heap := NewPriorityQueue[PQItem, int32](100)
	explorer := g.GetGraphExplorer()

	temp_point := geo.NewPoint(geo.Coord{0, 0})

	for _, item := range starts {
		start := item.A
		if avoid_geom.HasValue() {
			coord := g.GetNodeGeom(start)
			temp_point.SetCoordinates(coord)
			if avoid_geom.Value.Contains(&temp_point) {
				continue
			}
		}
		dist := item.B
		start_flag := node_flags.Get(start)
		start_flag.Dist = dist
		heap.Enqueue(PQItem{start, dist}, dist)
	}

	for {
		curr_item, ok := heap.Dequeue()
		if !ok {
			break
		}
		curr_id := curr_item.item
		curr_dist := curr_item.dist
		curr_flag := node_flags.Get(curr_id)
		if curr_flag.Dist < curr_dist {
			continue
		}
		explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_EDGES, func(ref graph.EdgeRef) {
			other_id := ref.OtherID
			if avoid_geom.HasValue() {
				coord := g.GetNodeGeom(other_id)
				temp_point.SetCoordinates(coord)
				if avoid_geom.Value.Contains(&temp_point) {
					return
				}
			}
			if avoid_roads.HasValue() {
				edge_attr := att.GetEdgeAttribs(ref.EdgeID)
				if Contains(avoid_roads.Value, edge_attr.Type) {
					return
				}
			}
			other_flag := node_flags.Get(other_id)
			new_length := curr_flag.Dist + explorer.GetEdgeWeight(ref)
			if new_length > max_range {
				return
			}
			if other_flag.Dist > new_length {
				other_flag.Dist = new_length
				heap.Enqueue(PQItem{other_id, new_length}, new_length)
			}
		})
	}
}

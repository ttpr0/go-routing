package onetomany

import (
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
)

func NewRangeDijkstra(g graph.IGraph, max_range int32) *RangeDijkstra {
	return &RangeDijkstra{g: g, max_range: max_range}
}

type RangeDijkstra struct {
	g         graph.IGraph
	max_range int32
}

func (self *RangeDijkstra) CreateSolver() ISolver {
	node_flags := NewFlags[DistFlag](int32(self.g.NodeCount()), DistFlag{1000000})
	return &RangeDijkstraSolver{
		g:          self.g,
		node_flags: node_flags,
		max_range:  self.max_range,
	}
}

type RangeDijkstraSolver struct {
	g          graph.IGraph
	node_flags Flags[DistFlag]
	max_range  int32
}

// CalcDiatanceFromStarts implements ISolver.
func (self *RangeDijkstraSolver) CalcDistanceFromStart(starts Array[Tuple[int32, int32]]) error {
	self.node_flags.Reset()
	_CalcRangeDijkstra(self.g, starts, self.node_flags, self.max_range)
	return nil
}

// GetDistance implements ISolver.
func (self *RangeDijkstraSolver) GetDistance(node int32) int32 {
	return self.node_flags.Get(node).Dist
}

func _CalcRangeDijkstra(g graph.IGraph, starts Array[Tuple[int32, int32]], node_flags Flags[DistFlag], max_range int32) {
	heap := NewPriorityQueue[PQItem, int32](100)
	explorer := g.GetGraphExplorer()

	for _, item := range starts {
		start := item.A
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

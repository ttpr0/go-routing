package onetomany

import (
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
)

func NewRangeDijkstraTC(g graph.IGraph, max_range int32) *RangeDijkstraTC {
	return &RangeDijkstraTC{g: g, max_range: max_range}
}

type RangeDijkstraTC struct {
	g         graph.IGraph
	max_range int32
}

func (self *RangeDijkstraTC) CreateSolver() ISolver {
	node_flags := NewFlags[DistFlag](int32(self.g.NodeCount()), DistFlag{1000000})
	edge_flags := NewFlags[DistFlag](int32(self.g.EdgeCount()), DistFlag{1000000})
	return &RangeDijkstraTCSolver{
		g:          self.g,
		node_flags: node_flags,
		edge_flags: edge_flags,
		max_range:  self.max_range,
	}
}

type RangeDijkstraTCSolver struct {
	g          graph.IGraph
	node_flags Flags[DistFlag]
	edge_flags Flags[DistFlag]
	max_range  int32
}

// CalcDiatanceFromStarts implements ISolver.
func (self *RangeDijkstraTCSolver) CalcDistanceFromStarts(starts Array[Tuple[int32, int32]]) error {
	self.node_flags.Reset()
	self.edge_flags.Reset()
	_CalcRangeDijkstraTC(self.g, starts, self.node_flags, self.edge_flags, self.max_range)
	return nil
}

// CalcDistanceFromStart implements ISolver.
func (self *RangeDijkstraTCSolver) CalcDistanceFromStart(start int32) error {
	self.node_flags.Reset()
	self.edge_flags.Reset()
	starts := [1]Tuple[int32, int32]{MakeTuple(start, int32(0))}
	_CalcRangeDijkstraTC(self.g, starts[:], self.node_flags, self.edge_flags, self.max_range)
	return nil
}

// GetDistance implements ISolver.
func (self *RangeDijkstraTCSolver) GetDistance(node int32) int32 {
	return self.node_flags.Get(node).Dist
}

func _CalcRangeDijkstraTC(g graph.IGraph, starts Array[Tuple[int32, int32]], node_flags Flags[DistFlag], edge_flags Flags[DistFlag], max_range int32) {
	heap := NewPriorityQueue[PQItem, int32](100)
	explorer := g.GetGraphExplorer()

	for _, item := range starts {
		start := item.A
		dist := item.B
		start_flag := node_flags.Get(start)
		start_flag.Dist = dist

		explorer.ForAdjacentEdges(start, graph.FORWARD, graph.ADJACENT_EDGES, func(ref graph.EdgeRef) {
			edge_id := ref.EdgeID
			next_node_id := ref.OtherID
			edge_dist := explorer.GetEdgeWeight(ref) + dist
			if edge_dist > max_range {
				return
			}
			edge_flag := edge_flags.Get(edge_id)
			edge_flag.Dist = edge_dist
			heap.Enqueue(PQItem{edge_id, edge_dist}, edge_dist)
			node_flag := node_flags.Get(next_node_id)
			if edge_dist < node_flag.Dist {
				node_flag.Dist = edge_dist
			}
		})
	}

	for {
		curr_item, ok := heap.Dequeue()
		if !ok {
			break
		}
		curr_id := curr_item.item
		curr_dist := curr_item.dist
		curr_flag := edge_flags.Get(curr_id)
		if curr_flag.Dist < curr_dist {
			continue
		}
		curr_edge := g.GetEdge(curr_id)
		curr_ref := graph.EdgeRef{EdgeID: curr_id, Type: 0, OtherID: curr_edge.NodeB}
		explorer.ForAdjacentEdges(curr_edge.NodeB, graph.FORWARD, graph.ADJACENT_EDGES, func(ref graph.EdgeRef) {
			other_id := ref.EdgeID
			other_node_b := ref.OtherID
			other_flag := edge_flags.Get(other_id)
			new_length := curr_flag.Dist + explorer.GetEdgeWeight(ref) + explorer.GetTurnCost(curr_ref, curr_edge.NodeB, ref)
			if new_length > max_range {
				return
			}
			if other_flag.Dist > new_length {
				edge_flag := edge_flags.Get(other_id)
				edge_flag.Dist = new_length
				heap.Enqueue(PQItem{other_id, new_length}, new_length)
				node_flag := node_flags.Get(other_node_b)
				if new_length < node_flag.Dist {
					node_flag.Dist = new_length
				}
			}
		})
	}
}

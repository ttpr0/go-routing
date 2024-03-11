package onetomany

import (
	"github.com/ttpr0/go-routing/graph"
	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

func NewRangeRPHAST(g graph.ICHGraph, target_nodes Array[int32], max_range int32) *RangeRPHAST {
	return &RangeRPHAST{
		g:                 g,
		max_range:         max_range,
		down_edges_subset: _TargetSelection(g, target_nodes),
	}
}

// with restricted target selection
func NewRangeRPHAST2(g graph.ICHGraph, target_nodes Array[int32], max_range int32) *RangeRPHAST {
	return &RangeRPHAST{
		g:                 g,
		max_range:         max_range,
		down_edges_subset: _RestrictedTargetSelection(g, target_nodes, max_range),
	}
}

type RangeRPHAST struct {
	g                 graph.ICHGraph
	max_range         int32
	down_edges_subset List[structs.Shortcut]
}

func (self *RangeRPHAST) CreateSolver() ISolver {
	node_flags := NewFlags[DistFlag](int32(self.g.NodeCount()), DistFlag{1000000})
	return &RangeRPHASTSolver{
		g:                 self.g,
		max_range:         self.max_range,
		node_flags:        node_flags,
		down_edges_subset: self.down_edges_subset,
	}
}

type RangeRPHASTSolver struct {
	g                 graph.ICHGraph
	max_range         int32
	down_edges_subset List[structs.Shortcut]
	node_flags        Flags[DistFlag]
}

// CalcDiatanceFromStarts implements ISolver.
func (self *RangeRPHASTSolver) CalcDistanceFromStart(starts Array[Tuple[int32, int32]]) error {
	self.node_flags.Reset()
	_CalcRangeRPHAST(self.g, starts, self.max_range, self.node_flags, self.down_edges_subset)
	return nil
}

// GetDistance implements ISolver.
func (self *RangeRPHASTSolver) GetDistance(node int32) int32 {
	return self.node_flags.Get(node).Dist
}

func _CalcRangeRPHAST(g graph.ICHGraph, starts Array[Tuple[int32, int32]], max_range int32, node_flags Flags[DistFlag], down_edges_subset List[structs.Shortcut]) {
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
		explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_UPWARDS, func(ref graph.EdgeRef) {
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
	// downwards sweep
	for i := 0; i < len(down_edges_subset); i++ {
		edge := down_edges_subset[i]
		curr_flag := node_flags.Get(edge.From)
		curr_len := curr_flag.Dist
		new_len := curr_len + edge.Weight
		if new_len > max_range {
			continue
		}
		other_flag := node_flags.Get(edge.To)
		if other_flag.Dist > new_len {
			other_flag.Dist = new_len
		}
	}
}

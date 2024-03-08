package onetomany

import (
	"github.com/ttpr0/go-routing/graph"
	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

func NewRPHAST(g graph.ICHGraph, target_nodes Array[int32]) *RPHAST {
	return &RPHAST{
		g:                 g,
		down_edges_subset: _TargetSelection(g, target_nodes),
	}
}

type RPHAST struct {
	g                 graph.ICHGraph
	down_edges_subset List[structs.Shortcut]
}

func (self *RPHAST) CreateSolver() ISolver {
	node_flags := NewFlags[DistFlag](int32(self.g.NodeCount()), DistFlag{1000000})
	return &RPHASTSolver{
		g:                 self.g,
		node_flags:        node_flags,
		down_edges_subset: self.down_edges_subset,
	}
}

type RPHASTSolver struct {
	g                 graph.ICHGraph
	down_edges_subset List[structs.Shortcut]
	node_flags        Flags[DistFlag]
}

// CalcDiatanceFromStarts implements ISolver.
func (self *RPHASTSolver) CalcDistanceFromStarts(starts Array[Tuple[int32, int32]]) error {
	self.node_flags.Reset()
	_CalcRPHAST(self.g, starts, self.node_flags, self.down_edges_subset)
	return nil
}

// CalcDistanceFromStart implements ISolver.
func (self *RPHASTSolver) CalcDistanceFromStart(start int32) error {
	self.node_flags.Reset()
	starts := [1]Tuple[int32, int32]{MakeTuple(start, int32(0))}
	_CalcRPHAST(self.g, starts[:], self.node_flags, self.down_edges_subset)
	return nil
}

// GetDistance implements ISolver.
func (self *RPHASTSolver) GetDistance(node int32) int32 {
	return self.node_flags.Get(node).Dist
}

func _TargetSelection(g graph.ICHGraph, target_nodes Array[int32]) List[structs.Shortcut] {
	node_queue := NewQueue[int32]()
	for i := 0; i < target_nodes.Length(); i++ {
		node_queue.Push(target_nodes[i])
	}

	// select graph subset by marking visited nodes
	explorer := g.GetGraphExplorer()
	graph_subset := NewArray[bool](int(g.NodeCount()))
	for {
		if node_queue.Size() == 0 {
			break
		}
		node, _ := node_queue.Pop()
		if graph_subset[node] {
			continue
		}
		graph_subset[node] = true
		node_level := g.GetNodeLevel(node)
		explorer.ForAdjacentEdges(node, graph.BACKWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			if graph_subset[ref.OtherID] {
				return
			}
			if node_level >= g.GetNodeLevel(ref.OtherID) {
				return
			}
			node_queue.Push(ref.OtherID)
		})
	}
	// selecting subset of downward edges for linear sweep
	down_edges_subset := NewList[structs.Shortcut](target_nodes.Length())
	down_edges, _ := g.GetDownEdges(graph.FORWARD)
	for i := 0; i < len(down_edges); i++ {
		edge := down_edges[i]
		if !graph_subset[edge.From] {
			continue
		}
		down_edges_subset.Add(edge)
	}

	return down_edges_subset
}

func _CalcRPHAST(g graph.ICHGraph, starts Array[Tuple[int32, int32]], node_flags Flags[DistFlag], down_edges_subset List[structs.Shortcut]) {
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
		other_flag := node_flags.Get(edge.To)
		if other_flag.Dist > new_len {
			other_flag.Dist = new_len
		}
	}
}

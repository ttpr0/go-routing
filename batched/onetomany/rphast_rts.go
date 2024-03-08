package onetomany

import (
	"github.com/ttpr0/go-routing/graph"
	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

// creates RPHAST with restricted target selection
func NewRPHAST2(g graph.ICHGraph, target_nodes Array[int32], max_range int32) *RPHAST {
	return &RPHAST{
		g:                 g,
		down_edges_subset: _RestrictedTargetSelection(g, target_nodes, max_range),
	}
}

func _RestrictedTargetSelection(g graph.ICHGraph, target_nodes Array[int32], max_range int32) List[structs.Shortcut] {
	node_queue := NewPriorityQueue[int32, int32](10000)

	for i := 0; i < target_nodes.Length(); i++ {
		node_queue.Enqueue(target_nodes[i], 0)
	}

	// select graph subset by marking visited nodes
	explorer := g.GetGraphExplorer()
	lengths := NewArray[int32](g.NodeCount())
	graph_subset := NewArray[bool](g.NodeCount())
	for {
		node, ok := node_queue.Dequeue()
		if !ok {
			break
		}
		if graph_subset[node] {
			continue
		}
		graph_subset[node] = true
		node_level := g.GetNodeLevel(node)
		node_len := lengths[node]
		explorer.ForAdjacentEdges(node, graph.BACKWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			if graph_subset[ref.OtherID] {
				return
			}
			if node_level >= g.GetNodeLevel(ref.OtherID) {
				return
			}
			new_len := node_len + explorer.GetEdgeWeight(ref)
			if new_len > max_range {
				return
			}
			if new_len < lengths[ref.OtherID] {
				lengths[ref.OtherID] = new_len
				node_queue.Enqueue(ref.OtherID, new_len)
			}
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

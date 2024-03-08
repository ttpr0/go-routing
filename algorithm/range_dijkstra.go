package algorithm

import (
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
)

type DistFlag struct {
	Dist int32
}

func (self *DistFlag) GetDist() int32 {
	return self.Dist
}

type PQItem struct {
	item int32
	dist int32
}

func CalcRangeDijkstra(g graph.IGraph, starts Array[Tuple[int32, int32]], node_flags Flags[DistFlag], max_range int32) {
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

func CalcRangeDijkstraTC(g graph.IGraph, starts Array[Tuple[int32, int32]], node_flags Flags[DistFlag], edge_flags Flags[DistFlag], max_range int32) {
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

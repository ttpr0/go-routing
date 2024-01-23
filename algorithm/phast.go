package algorithm

import (
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
)

type _FlagPH struct {
	path_length int32
	prev_edge   int32
	visited     bool
}

func CalcPHAST(g graph.ICHGraph, start int32) Array[int32] {
	dist := NewArray[int32](g.NodeCount())
	visited := NewArray[bool](g.NodeCount())
	for i := 0; i < g.NodeCount(); i++ {
		dist[i] = 1000000000
	}
	dist[start] = 0

	heap := NewPriorityQueue[int32, int32](100)
	heap.Enqueue(start, 0)

	explorer := g.GetGraphExplorer()

	for {
		curr_id, ok := heap.Dequeue()
		if !ok {
			break
		}
		if visited[curr_id] {
			continue
		}
		visited[curr_id] = true
		explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			other_id := ref.OtherID
			if g.GetNodeLevel(other_id) <= g.GetNodeLevel(curr_id) {
				return
			}
			if visited[other_id] {
				return
			}
			new_length := dist[curr_id] + explorer.GetEdgeWeight(ref)
			if dist[other_id] > new_length {
				dist[other_id] = new_length
				heap.Enqueue(other_id, new_length)
			}
		})
	}

	for i := 0; i < len(dist); i++ {
		curr_len := dist[i]
		explorer.ForAdjacentEdges(int32(i), graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			other_id := ref.OtherID
			if g.GetNodeLevel(other_id) >= g.GetNodeLevel(int32(i)) {
				return
			}
			if dist[other_id] > (curr_len + explorer.GetEdgeWeight(ref)) {
				dist[other_id] = curr_len + explorer.GetEdgeWeight(ref)
			}
		})
	}

	return dist
}

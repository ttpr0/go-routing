package algorithm

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func CalcPHAST3(g graph.ICHGraph, start int32, max_range int32) Array[int32] {
	visited := NewArray[bool](g.NodeCount())
	dist := NewArray[int32](g.NodeCount())
	for i := 0; i < len(dist); i++ {
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
		//curr := (*d.graph).GetNode(curr_id)
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
			if new_length > max_range {
				return
			}
			if dist[other_id] > new_length {
				dist[other_id] = new_length
				heap.Enqueue(other_id, new_length)
			}
		})
	}

	down_edges, _ := g.GetDownEdges(graph.FORWARD)
	num_edges := int32(len(down_edges))
	i := int32(0)
	for i < num_edges {
		edge := down_edges[i]
		curr_len := dist[edge.From]
		if curr_len > max_range {
			count := graph.Shortcut_get_payload[int32](&edge, 0)
			i += count
			continue
		}
		new_len := curr_len + edge.Weight
		if dist[edge.To] > new_len {
			dist[edge.To] = new_len
		}
		i += 1
	}

	return dist
}

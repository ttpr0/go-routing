package algorithm

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func CalcAllDijkstra(g graph.IGraph, start int32, max_range int32) Array[int32] {
	dist := NewArray[int32](g.NodeCount())
	visited := NewArray[bool](g.NodeCount())
	for i := 0; i < g.NodeCount(); i++ {
		dist[i] = 1000000000
	}
	dist[start] = 0

	heap := NewPriorityQueue[int32, int32](100)
	heap.Enqueue(start, 0)

	explorer := g.GetDefaultExplorer()

	for {
		curr_id, ok := heap.Dequeue()
		if !ok {
			break
		}
		if visited[curr_id] {
			continue
		}
		visited[curr_id] = true
		edges := explorer.GetAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_ALL)
		for {
			ref, ok := edges.Next()
			if !ok {
				break
			}
			if ref.IsShortcut() {
				continue
			}
			other_id := ref.OtherID
			if visited[other_id] {
				continue
			}
			new_length := dist[curr_id] + explorer.GetEdgeWeight(ref)
			if new_length > max_range {
				continue
			}
			if dist[other_id] > new_length {
				dist[other_id] = new_length
				heap.Enqueue(other_id, new_length)
			}
		}
	}

	return dist
}

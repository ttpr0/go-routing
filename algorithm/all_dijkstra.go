package algorithm

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type _FlagD struct {
	path_length int32
	prev_edge   int32
	visited     bool
}

func CalcAllDijkstra(g graph.IGraph, start int32) {
	flags := make([]_FlagD, g.NodeCount())
	for i := 0; i < len(flags); i++ {
		flags[i].path_length = 1000000000
	}
	flags[start].path_length = 0

	heap := NewPriorityQueue[int32, int32](100)
	heap.Enqueue(start, 0)

	explorer := g.GetDefaultExplorer()

	for {
		curr_id, ok := heap.Dequeue()
		if !ok {
			return
		}
		//curr := (*d.graph).GetNode(curr_id)
		curr_flag := flags[curr_id]
		if curr_flag.visited {
			continue
		}
		curr_flag.visited = true
		edges := explorer.GetAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_ALL)
		for {
			ref, ok := edges.Next()
			if !ok {
				break
			}
			if ref.IsShortcut() {
				continue
			}
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			//other := (*d.graph).GetNode(other_id)
			other_flag := flags[other_id]
			if other_flag.visited {
				continue
			}
			new_length := curr_flag.path_length + explorer.GetEdgeWeight(ref)
			if other_flag.path_length > new_length {
				other_flag.prev_edge = edge_id
				other_flag.path_length = new_length
				heap.Enqueue(other_id, new_length)
			}
			flags[other_id] = other_flag
		}
		flags[curr_id] = curr_flag
	}
}

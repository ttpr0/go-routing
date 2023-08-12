package algorithm

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type _FlagPH struct {
	path_length int32
	prev_edge   int32
	visited     bool
}

func CalcPHAST(g graph.ICHGraph, start int32) {
	weight := g.GetWeighting()

	flags := make([]_FlagPH, g.NodeCount())
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
			break
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
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			if g.GetNodeLevel(other_id) <= g.GetNodeLevel(curr_id) {
				continue
			}
			other_flag := flags[other_id]
			if other_flag.visited {
				continue
			}
			new_length := curr_flag.path_length + weight.GetEdgeWeight(ref.EdgeID)
			if other_flag.path_length > new_length {
				other_flag.prev_edge = edge_id
				other_flag.path_length = new_length
				heap.Enqueue(other_id, new_length)
			}
			flags[other_id] = other_flag
		}
		flags[curr_id] = curr_flag
	}

	for i := 0; i < len(flags); i++ {
		curr_len := flags[i].path_length
		edges := explorer.GetAdjacentEdges(int32(i), graph.FORWARD, graph.ADJACENT_ALL)
		for {
			ref, ok := edges.Next()
			if !ok {
				break
			}
			other_id := ref.OtherID
			if g.GetNodeLevel(other_id) >= g.GetNodeLevel(int32(i)) {
				continue
			}
			other_flag := flags[other_id]
			if other_flag.path_length > (curr_len + weight.GetEdgeWeight(ref.EdgeID)) {
				other_flag.path_length = curr_len + weight.GetEdgeWeight(ref.EdgeID)
				flags[other_id] = other_flag
			}
		}
	}
}

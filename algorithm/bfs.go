package algorithm

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type _FlagB struct {
	path_length float64
	prev_edge   int32
	visited     bool
}

func CalcBreathFirstSearch(g graph.IGraph, start int32) {
	flags := make([]_FlagD, g.NodeCount())
	for i := 0; i < len(flags); i++ {
		flags[i].path_length = 1000000000
	}
	flags[start].path_length = 0

	queue := util.NewQueue[int32]()
	queue.Push(start)

	for {
		curr_id, ok := queue.Pop()
		if !ok {
			return
		}
		//curr := (*d.graph).GetNode(curr_id)
		curr_flag := flags[curr_id]
		if curr_flag.visited {
			continue
		}
		curr_flag.visited = true
		edges := g.GetAdjacentEdges(curr_id, graph.FORWARD)
		for {
			ref, ok := edges.Next()
			if !ok {
				break
			}
			if ref.IsShortcut() {
				continue
			}
			other_id := ref.NodeID
			//other := (*d.graph).GetNode(other_id)
			other_flag := flags[other_id]
			if other_flag.visited {
				continue
			}
			queue.Push(other_id)
		}
		flags[curr_id] = curr_flag
	}
}

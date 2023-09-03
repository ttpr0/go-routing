package algorithm

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func CalcBreathFirstSearch(g graph.IGraph, start int32) Array[bool] {
	visited := NewArray[bool](g.NodeCount())

	queue := NewQueue[int32]()
	queue.Push(start)

	explorer := g.GetDefaultExplorer()

	for {
		curr_id, ok := queue.Pop()
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
			queue.Push(other_id)
		}
	}

	return visited
}

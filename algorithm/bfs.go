package algorithm

import (
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
)

func CalcBreathFirstSearch(g graph.IGraph, start int32) Array[bool] {
	visited := NewArray[bool](g.NodeCount())

	queue := NewQueue[int32]()
	queue.Push(start)

	explorer := g.GetGraphExplorer()

	for {
		curr_id, ok := queue.Pop()
		if !ok {
			break
		}
		if visited[curr_id] {
			continue
		}
		visited[curr_id] = true
		explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			if ref.IsShortcut() {
				return
			}
			other_id := ref.OtherID
			if visited[other_id] {
				return
			}
			queue.Push(other_id)
		})
	}

	return visited
}

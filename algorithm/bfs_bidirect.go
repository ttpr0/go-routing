package algorithm

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

// Computes breath-first-search using both forward and backward search.
// Only marks nodes visited by both searches
func CalcBidirectBFS(g graph.IGraph, start int32) Array[bool] {
	explorer := g.GetDefaultExplorer()

	// fwd search
	visited_fwd := NewArray[bool](g.NodeCount())
	queue := NewQueue[int32]()
	queue.Push(start)
	for {
		curr_id, ok := queue.Pop()
		if !ok {
			break
		}
		if visited_fwd[curr_id] {
			continue
		}
		visited_fwd[curr_id] = true
		edges := explorer.GetAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_EDGES)
		for {
			ref, ok := edges.Next()
			if !ok {
				break
			}
			if ref.IsShortcut() {
				continue
			}
			other_id := ref.OtherID
			if visited_fwd[other_id] {
				continue
			}
			queue.Push(other_id)
		}
	}

	// bwd search
	visited_bwd := NewArray[bool](g.NodeCount())
	queue = NewQueue[int32]()
	queue.Push(start)
	for {
		curr_id, ok := queue.Pop()
		if !ok {
			break
		}
		if visited_bwd[curr_id] {
			continue
		}
		visited_bwd[curr_id] = true
		edges := explorer.GetAdjacentEdges(curr_id, graph.BACKWARD, graph.ADJACENT_EDGES)
		for {
			ref, ok := edges.Next()
			if !ok {
				break
			}
			if ref.IsShortcut() {
				continue
			}
			other_id := ref.OtherID
			if visited_bwd[other_id] {
				continue
			}
			queue.Push(other_id)
		}
	}

	// combine results
	visited := NewArray[bool](g.NodeCount())
	for i := 0; i < g.NodeCount(); i++ {
		if visited_fwd[i] && visited_bwd[i] {
			visited[i] = true
		}
	}

	return visited
}

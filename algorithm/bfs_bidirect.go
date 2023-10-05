package algorithm

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

// Computes breath-first-search using both forward and backward search.
// Only marks nodes visited by both searches
func CalcBidirectBFS(g graph.IGraph, start int32) Array[bool] {
	explorer := g.GetDefaultExplorer()

	// bidirectional search
	visited_fwd := NewArray[bool](g.NodeCount())
	visited_bwd := NewArray[bool](g.NodeCount())
	queue_fwd := NewQueue[int32]()
	queue_fwd.Push(start)
	queue_bwd := NewQueue[int32]()
	queue_bwd.Push(start)
	fwd_finished := false
	bwd_finished := false
	for {
		if fwd_finished && bwd_finished {
			break
		}

		if !fwd_finished {
			s := true
			curr_id, ok := queue_fwd.Pop()
			if !ok {
				fwd_finished = true
				s = false
			}
			if visited_fwd[curr_id] {
				s = false
			}
			if s {
				visited_fwd[curr_id] = true
				explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_EDGES, func(ref graph.EdgeRef) {
					if ref.IsShortcut() {
						return
					}
					other_id := ref.OtherID
					if visited_fwd[other_id] {
						return
					}
					if !visited_bwd[other_id] && bwd_finished {
						return
					}
					queue_fwd.Push(other_id)
				})
			}
		}

		if !bwd_finished {
			s := true
			curr_id, ok := queue_bwd.Pop()
			if !ok {
				bwd_finished = true
				s = false
			}
			if visited_bwd[curr_id] {
				s = false
			}
			if s {
				visited_bwd[curr_id] = true
				explorer.ForAdjacentEdges(curr_id, graph.BACKWARD, graph.ADJACENT_EDGES, func(ref graph.EdgeRef) {
					if ref.IsShortcut() {
						return
					}
					other_id := ref.OtherID
					if visited_bwd[other_id] {
						return
					}
					if !visited_fwd[other_id] && fwd_finished {
						return
					}
					queue_bwd.Push(other_id)
				})
			}
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

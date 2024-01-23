package algorithm

import (
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
)

// Computes max flow using edmonds-karp algorithm.
//
// All Edges are interpreted as having equal capacity 1.
func ComputeMaxFlow(g graph.IGraph, source, sink int32) int {
	max_flow := 0

	type _Flag struct {
		prev_node  int32
		prev_edge  int32
		is_reverse bool
	}

	edge_flow := NewArray[byte](int(g.EdgeCount()))
	flags := NewArray[_Flag](int(g.NodeCount()))
	visited := NewArray[bool](int(g.NodeCount()))
	for {
		queue := NewQueue[int32]()

		for i := 0; i < visited.Length(); i++ {
			visited[i] = false
		}
		visited[source] = true
		queue.Push(source)

		end := int32(-1)

		explorer := g.GetGraphExplorer()
		for {
			curr, ok := queue.Pop()
			if !ok {
				break
			}
			if curr == sink {
				end = curr
				break
			}

			explorer.ForAdjacentEdges(curr, graph.FORWARD, graph.ADJACENT_EDGES, func(ref graph.EdgeRef) {
				// check if edge should stil be traversed
				if visited[ref.OtherID] || edge_flow[ref.EdgeID] == 1 {
					return
				}

				other_flag := flags[ref.OtherID]
				other_flag.is_reverse = false
				other_flag.prev_edge = ref.EdgeID
				other_flag.prev_node = curr
				flags[ref.OtherID] = other_flag
				queue.Push(ref.OtherID)
				visited[ref.OtherID] = true
			})
			explorer.ForAdjacentEdges(curr, graph.BACKWARD, graph.ADJACENT_EDGES, func(ref graph.EdgeRef) {
				// check if edge should stil be traversed
				if visited[ref.OtherID] || edge_flow[ref.EdgeID] == 0 {
					return
				}

				other_flag := flags[ref.OtherID]
				other_flag.is_reverse = true
				other_flag.prev_edge = ref.EdgeID
				other_flag.prev_node = curr
				flags[ref.OtherID] = other_flag
				queue.Push(ref.OtherID)
				visited[ref.OtherID] = true
			})
		}

		if end == -1 {
			break
		}
		for {
			if end == source {
				break
			}
			curr_flag := flags[end]
			if curr_flag.is_reverse {
				edge_flow[curr_flag.prev_edge] = 0
			} else {
				edge_flow[curr_flag.prev_edge] = 1
			}
			end = curr_flag.prev_node
		}
		max_flow += 1
	}
	return max_flow
}

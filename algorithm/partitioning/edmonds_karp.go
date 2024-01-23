package partitioning

import (
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
)

type EdmondsKarp struct {
	g          graph.IGraph
	node_tiles Array[int16]

	// edge flows stored as 0 or 1;
	//
	// To solve issue with reverse oneway edges, each directed edge of the original graph is viewed as two edges with two flows:
	// first is for the actual edge,
	// second for a "virtual" edge used during BFS reversed to the orignal edge;
	edge_flow   Array[Tuple[byte, byte]]
	base_tile   int16
	source_tile int16
	sink_tile   int16
	max_flow    int

	source_queue ArrayQueue[int32]
	bfs_flags    Array[_Flag]
	visited      Array[bool]
}

func NewEdmondsKarp(g graph.IGraph, sources List[int32], source_tile int16, sinks List[int32], sink_tile int16, base_nodes List[int32], base_tile int16) *EdmondsKarp {
	source_queue := NewArrayQueue[int32](100)
	node_tiles := NewArray[int16](int(g.NodeCount()))
	bfs_flags := NewArray[_Flag](int(g.NodeCount()))
	visited := NewArray[bool](int(g.NodeCount()))
	edge_flow := NewArray[Tuple[byte, byte]](int(g.EdgeCount()))

	for _, node := range base_nodes {
		node_tiles[node] = base_tile
	}
	for _, source := range sources {
		source_queue.Push(source)
		node_tiles[source] = source_tile
	}
	for _, sink := range sinks {
		node_tiles[sink] = sink_tile
	}

	return &EdmondsKarp{
		g:           g,
		node_tiles:  node_tiles,
		edge_flow:   edge_flow,
		base_tile:   base_tile,
		source_tile: source_tile,
		sink_tile:   sink_tile,
		max_flow:    0,

		source_queue: source_queue,
		bfs_flags:    bfs_flags,
		visited:      visited,
	}
}

func (self *EdmondsKarp) ComputeMaxFlow() int {
	for {
		flow := self.BFS()
		if flow == 0 {
			break
		}
		self.max_flow += flow
	}
	return self.max_flow
}

func (self *EdmondsKarp) ComputeMinCut() {
	queue := self.source_queue.Copy()
	visited := NewArray[bool](int(self.g.NodeCount()))

	// clear visited
	for i := 0; i < int(self.g.NodeCount()); i++ {
		if self.node_tiles[i] == self.source_tile {
			visited[i] = true
		}
	}

	explorer := self.g.GetGraphExplorer()
	for {
		curr, ok := queue.Pop()
		if !ok {
			break
		}
		if self.node_tiles[curr] == self.sink_tile {
			panic("this should not happen")
		}
		if self.node_tiles[curr] == self.base_tile {
			self.node_tiles[curr] = self.source_tile
		}

		explorer.ForAdjacentEdges(curr, graph.FORWARD, graph.ADJACENT_EDGES, func(ref graph.EdgeRef) {
			if visited[ref.OtherID] || self.node_tiles[ref.OtherID] != self.base_tile {
				return
			}

			flow := self.edge_flow[ref.EdgeID]
			if flow.A == 0 || flow.B == 1 {
				queue.Push(ref.OtherID)
				visited[ref.OtherID] = true
			}
		})
		explorer.ForAdjacentEdges(curr, graph.BACKWARD, graph.ADJACENT_EDGES, func(ref graph.EdgeRef) {
			if visited[ref.OtherID] || self.node_tiles[ref.OtherID] != self.base_tile {
				return
			}

			flow := self.edge_flow[ref.EdgeID]
			if flow.A == 1 || flow.B == 0 {
				queue.Push(ref.OtherID)
				visited[ref.OtherID] = true
			}
		})
	}

	for i := 0; i < int(self.g.NodeCount()); i++ {
		if self.node_tiles[i] == self.base_tile {
			self.node_tiles[i] = self.sink_tile
		}
	}
}

func (self *EdmondsKarp) GetNodeTiles() Array[int16] {
	return self.node_tiles
}

type _Flag struct {
	prev_node  int32
	prev_edge  int32
	is_reverse bool
	is_virtual bool
}

// computed bfs on residual graph and returns new flow
func (self *EdmondsKarp) BFS() int {
	flags := self.bfs_flags
	visited := self.visited
	queue := self.source_queue.Copy()

	// clear visited
	for i := 0; i < visited.Length(); i++ {
		if self.node_tiles[i] == self.source_tile {
			visited[i] = true
		} else {
			visited[i] = false
		}
	}

	explorer := self.g.GetGraphExplorer()

	end := int32(-1)
	for {
		curr, ok := queue.Pop()
		if !ok {
			break
		}
		if self.node_tiles[curr] == self.sink_tile {
			end = curr
			break
		}

		explorer.ForAdjacentEdges(curr, graph.FORWARD, graph.ADJACENT_EDGES, func(ref graph.EdgeRef) {
			// check if edge should stil be traversed
			if visited[ref.OtherID] {
				return
			}
			// check if node is part of subgraph
			tile := self.node_tiles[ref.OtherID]
			if tile != self.base_tile && tile != self.source_tile && tile != self.sink_tile {
				return
			}

			flow := self.edge_flow[ref.EdgeID]
			if flow.A == 0 {
				other_flag := flags[ref.OtherID]
				other_flag.is_reverse = false
				other_flag.is_virtual = false
				other_flag.prev_edge = ref.EdgeID
				other_flag.prev_node = curr
				flags[ref.OtherID] = other_flag
				queue.Push(ref.OtherID)
				visited[ref.OtherID] = true
			} else if flow.B == 1 {
				other_flag := flags[ref.OtherID]
				other_flag.is_reverse = true
				other_flag.is_virtual = true
				other_flag.prev_edge = ref.EdgeID
				other_flag.prev_node = curr
				flags[ref.OtherID] = other_flag
				queue.Push(ref.OtherID)
				visited[ref.OtherID] = true
			}
		})
		explorer.ForAdjacentEdges(curr, graph.BACKWARD, graph.ADJACENT_EDGES, func(ref graph.EdgeRef) {
			// check if edge should stil be traversed
			if visited[ref.OtherID] {
				return
			}
			// check if node is part of subgraph
			tile := self.node_tiles[ref.OtherID]
			if tile != self.base_tile && tile != self.source_tile && tile != self.sink_tile {
				return
			}

			flow := self.edge_flow[ref.EdgeID]
			if flow.A == 1 {
				other_flag := flags[ref.OtherID]
				other_flag.is_reverse = true
				other_flag.is_virtual = false
				other_flag.prev_edge = ref.EdgeID
				other_flag.prev_node = curr
				flags[ref.OtherID] = other_flag
				queue.Push(ref.OtherID)
				visited[ref.OtherID] = true
			} else if flow.B == 0 {
				other_flag := flags[ref.OtherID]
				other_flag.is_reverse = false
				other_flag.is_virtual = true
				other_flag.prev_edge = ref.EdgeID
				other_flag.prev_node = curr
				flags[ref.OtherID] = other_flag
				queue.Push(ref.OtherID)
				visited[ref.OtherID] = true
			}
		})
	}

	if end == -1 {
		return 0
	}
	for {
		if self.node_tiles[end] == self.source_tile {
			break
		}
		curr_flag := flags[end]
		if curr_flag.is_reverse {
			flow := self.edge_flow[curr_flag.prev_edge]
			if curr_flag.is_virtual {
				flow.B = 0
			} else {
				flow.A = 0
			}
			self.edge_flow[curr_flag.prev_edge] = flow
		} else {
			flow := self.edge_flow[curr_flag.prev_edge]
			if curr_flag.is_virtual {
				flow.B = 1
			} else {
				flow.A = 1
			}
			self.edge_flow[curr_flag.prev_edge] = flow
		}
		end = curr_flag.prev_node
	}
	return 1
}

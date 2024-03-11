package routing

import (
	"fmt"

	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
	"golang.org/x/exp/slog"
)

type _FlagBOD struct {
	curr_node   int32
	path_length float64
	prev_edge   int32
	is_shortcut bool
	visited     bool
	skip        bool
}

type BODijkstra struct {
	heap     PriorityQueue[_FlagBOD, float64]
	start_id int32
	end_id   int32
	graph    graph.ITiledGraph
	flags    Dict[int32, _FlagBOD]
}

func NewBODijkstra(graph graph.ITiledGraph, start, end int32) *BODijkstra {
	d := BODijkstra{graph: graph, start_id: start, end_id: end}

	flags := NewDict[int32, _FlagBOD](100)
	d.flags = flags

	heap := NewPriorityQueue[_FlagBOD, float64](100)
	heap.Enqueue(_FlagBOD{curr_node: start, prev_edge: -1, path_length: 0, skip: false}, 0)
	d.heap = heap

	return &d
}

func (self *BODijkstra) CalcShortestPath() bool {
	explorer := self.graph.GetGraphExplorer()

	for {
		curr_flag, ok := self.heap.Dequeue()
		if !ok {
			return false
		}
		curr_id := curr_flag.curr_node
		if curr_id == self.end_id {
			return true
		}
		//curr := (*d.graph).GetNode(curr_id)
		if self.flags.ContainsKey(curr_id) {
			temp_flag := self.flags.Get(curr_id)
			if temp_flag.path_length < curr_flag.path_length {
				continue
			}
		}
		curr_flag.visited = true
		self.flags.Set(curr_id, curr_flag)
		handler := func(ref graph.EdgeRef) {
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			var other_flag _FlagBOD
			if self.flags.ContainsKey(other_id) {
				other_flag = self.flags.Get(other_id)
			} else {
				other_flag = _FlagBOD{curr_node: other_id, path_length: 1000000, prev_edge: -1}
			}
			if other_flag.visited {
				return
			}
			new_length := curr_flag.path_length + float64(explorer.GetEdgeWeight(ref))
			if other_flag.path_length > new_length {
				if ref.IsCrossBorder() {
					tile_id := self.graph.GetNodeTile(other_id)
					if self.graph.GetNodeTile(self.end_id) == tile_id || self.graph.GetNodeTile(self.start_id) == tile_id {
						other_flag.skip = false
					} else {
						other_flag.skip = true
					}
				} else {
					other_flag.skip = curr_flag.skip
				}
				other_flag.curr_node = other_id
				other_flag.prev_edge = edge_id
				other_flag.is_shortcut = ref.IsShortcut()
				other_flag.path_length = new_length
				self.heap.Enqueue(other_flag, new_length)
			}
			self.flags[other_id] = other_flag
		}
		if curr_flag.skip {
			explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_SKIP, handler)
		} else {
			explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_ALL, handler)
		}
	}
}

func (self *BODijkstra) Steps(count int, handler func(int32)) bool {
	explorer := self.graph.GetGraphExplorer()

	for c := 0; c < count; c++ {
		curr_flag, ok := self.heap.Dequeue()
		if !ok {
			return false
		}
		curr_id := curr_flag.curr_node
		if curr_id == self.end_id {
			return false
		}
		if self.flags.ContainsKey(curr_id) {
			temp_flag := self.flags.Get(curr_id)
			if temp_flag.path_length < curr_flag.path_length {
				continue
			}
		}
		curr_flag.visited = true
		self.flags.Set(curr_id, curr_flag)
		handler := func(ref graph.EdgeRef) {
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			var other_flag _FlagBOD
			if self.flags.ContainsKey(other_id) {
				other_flag = self.flags.Get(other_id)
			} else {
				other_flag = _FlagBOD{curr_node: other_id, path_length: 1000000, prev_edge: -1}
			}
			if other_flag.visited {
				return
			}
			if ref.IsShortcut() {
				self.graph.GetEdgesFromShortcut(edge_id, false, handler)
			} else {
				handler(edge_id)
			}
			new_length := curr_flag.path_length + float64(explorer.GetEdgeWeight(ref))
			if other_flag.path_length > new_length {
				if ref.IsCrossBorder() {
					tile_id := self.graph.GetNodeTile(other_id)
					if self.graph.GetNodeTile(self.end_id) == tile_id || self.graph.GetNodeTile(self.start_id) == tile_id {
						other_flag.skip = false
					} else {
						other_flag.skip = true
					}
				} else {
					other_flag.skip = curr_flag.skip
				}
				other_flag.curr_node = other_id
				other_flag.prev_edge = edge_id
				other_flag.is_shortcut = ref.IsShortcut()
				other_flag.path_length = new_length
				self.heap.Enqueue(other_flag, new_length)
			}
			self.flags[other_id] = other_flag
		}
		if curr_flag.skip {
			explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_SKIP, handler)
		} else {
			explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_ALL, handler)
		}
	}
	return true
}

func (self *BODijkstra) GetShortestPath() Path {
	explorer := self.graph.GetGraphExplorer()

	path := NewList[int32](10)
	length := int32(self.flags[self.end_id].path_length)
	curr_id := self.end_id
	var edge int32
	for {
		if curr_id == self.start_id {
			break
		}
		curr_flag := self.flags[curr_id]
		edge = curr_flag.prev_edge
		if curr_flag.is_shortcut {
			self.graph.GetEdgesFromShortcut(edge, false, func(e int32) {
				path.Add(e)
			})
			curr_id = explorer.GetOtherNode(graph.CreateCHShortcutRef(edge), curr_id)
		} else {
			path.Add(edge)
			curr_id = explorer.GetOtherNode(graph.CreateEdgeRef(edge), curr_id)
		}
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	slog.Debug(fmt.Sprintf("length: %v", length))
	return NewPath(self.graph, path)
}

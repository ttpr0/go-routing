package routing

import (
	"fmt"

	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
	"golang.org/x/exp/slog"
)

type CH2 struct {
	startheap   PriorityQueue[int32, float64]
	endheap     PriorityQueue[int32, float64]
	mid_id      int32
	start_id    int32
	end_id      int32
	path_length float64
	graph       graph.ICHGraph
	flags       Dict[int32, flag_ch]
}

func NewCH2(graph graph.ICHGraph, start, end int32) *CH2 {
	startheap := NewPriorityQueue[int32, float64](10)
	endheap := NewPriorityQueue[int32, float64](10)
	flags := NewDict[int32, flag_ch](100)

	flags[start] = flag_ch{path_length1: 0, visited1: false, prev_edge1: -1, is_shortcut1: false, path_length2: 1000000, visited2: false, prev_edge2: -1, is_shortcut2: false}
	startheap.Enqueue(start, 0)
	flags[end] = flag_ch{path_length1: 1000000, visited1: false, prev_edge1: -1, is_shortcut1: false, path_length2: 0, visited2: false, prev_edge2: -1, is_shortcut2: false}
	endheap.Enqueue(end, 0)

	ch := CH2{
		startheap:   startheap,
		endheap:     endheap,
		mid_id:      -1,
		start_id:    start,
		end_id:      end,
		path_length: 100000000,
		graph:       graph,
		flags:       flags,
	}

	return &ch
}

func (self *CH2) CalcShortestPath() bool {
	explorer := self.graph.GetGraphExplorer()

	for {
		if self.startheap.Len() == 0 && self.endheap.Len() == 0 {
			break
		}
		if self.mid_id != -1 {
			s_id, _ := self.startheap.Peek()
			e_id, _ := self.endheap.Peek()
			s_flag := self.flags[s_id]
			e_flag := self.flags[e_id]
			if s_flag.path_length1 >= self.path_length && e_flag.path_length2 >= self.path_length {
				break
			}
		}

		// from start
		if self.startheap.Len() != 0 {
			curr_id, _ := self.startheap.Dequeue()
			curr_flag := self.flags[curr_id]
			if curr_flag.visited1 {
				continue
			}
			curr_flag.visited1 = true
			if curr_flag.visited2 && self.path_length > (curr_flag.path_length1+curr_flag.path_length2) {
				self.mid_id = curr_id
				self.path_length = curr_flag.path_length1 + curr_flag.path_length2
			}
			explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_UPWARDS, func(ref graph.EdgeRef) {
				edge_id := ref.EdgeID
				other_id := ref.OtherID
				var other_flag flag_ch
				if self.flags.ContainsKey(other_id) {
					other_flag = self.flags[other_id]
				} else {
					other_flag = flag_ch{path_length1: 1000000, visited1: false, prev_edge1: -1, is_shortcut1: false, path_length2: 1000000, visited2: false, prev_edge2: -1, is_shortcut2: false}
				}
				weight := explorer.GetEdgeWeight(ref)
				new_length := curr_flag.path_length1 + float64(weight)
				if new_length < other_flag.path_length1 {
					other_flag.path_length1 = new_length
					other_flag.prev_edge1 = edge_id
					other_flag.is_shortcut1 = ref.IsShortcut()
					self.startheap.Enqueue(other_id, new_length)
				}
				self.flags[other_id] = other_flag
			})
			self.flags[curr_id] = curr_flag
		}

		// from end
		if self.endheap.Len() != 0 {
			curr_id, _ := self.endheap.Dequeue()
			curr_flag := self.flags[curr_id]
			if curr_flag.visited2 {
				continue
			}
			curr_flag.visited2 = true
			if curr_flag.visited1 && self.path_length > (curr_flag.path_length1+curr_flag.path_length2) {
				self.mid_id = curr_id
				self.path_length = curr_flag.path_length1 + curr_flag.path_length2
			}
			explorer.ForAdjacentEdges(curr_id, graph.BACKWARD, graph.ADJACENT_UPWARDS, func(ref graph.EdgeRef) {
				edge_id := ref.EdgeID
				other_id := ref.OtherID
				var other_flag flag_ch
				if self.flags.ContainsKey(other_id) {
					other_flag = self.flags[other_id]
				} else {
					other_flag = flag_ch{path_length1: 1000000, visited1: false, prev_edge1: -1, is_shortcut1: false, path_length2: 1000000, visited2: false, prev_edge2: -1, is_shortcut2: false}
				}
				weight := explorer.GetEdgeWeight(ref)
				new_length := curr_flag.path_length2 + float64(weight)
				if new_length < other_flag.path_length2 {
					other_flag.path_length2 = new_length
					other_flag.prev_edge2 = edge_id
					other_flag.is_shortcut2 = ref.IsShortcut()
					self.endheap.Enqueue(other_id, new_length)
				}
				self.flags[other_id] = other_flag
			})
			self.flags[curr_id] = curr_flag
		}
	}
	if self.mid_id == -1 {
		return false
	}
	return true
}

func (self *CH2) Steps(count int, handler func(int32)) bool {
	explorer := self.graph.GetGraphExplorer()
	for c := 0; c < count; c++ {
		if self.startheap.Len() == 0 && self.endheap.Len() == 0 {
			return false
		}
		if self.mid_id != -1 {
			s_id, _ := self.startheap.Peek()
			e_id, _ := self.endheap.Peek()
			s_flag := self.flags[s_id]
			e_flag := self.flags[e_id]
			if s_flag.path_length1 >= self.path_length && e_flag.path_length2 >= self.path_length {
				return false
			}
		}
		// from start
		if self.startheap.Len() != 0 {
			curr_id, _ := self.startheap.Dequeue()
			curr_flag := self.flags[curr_id]
			if curr_flag.visited1 {
				continue
			}
			curr_flag.visited1 = true
			if curr_flag.visited2 && self.path_length > (curr_flag.path_length1+curr_flag.path_length2) {
				self.mid_id = curr_id
				self.path_length = curr_flag.path_length1 + curr_flag.path_length2
			}
			explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
				edge_id := ref.EdgeID
				other_id := ref.OtherID
				if self.graph.GetNodeLevel(other_id) <= self.graph.GetNodeLevel(curr_id) {
					return
				}
				var other_flag flag_ch
				if self.flags.ContainsKey(other_id) {
					other_flag = self.flags[other_id]
				} else {
					other_flag = flag_ch{path_length1: 1000000, visited1: false, prev_edge1: -1, is_shortcut1: false, path_length2: 1000000, visited2: false, prev_edge2: -1, is_shortcut2: false}
				}
				weight := explorer.GetEdgeWeight(ref)

				if ref.IsShortcut() {
					self.graph.GetEdgesFromShortcut(edge_id, false, handler)
				} else {
					handler(edge_id)
				}

				new_length := curr_flag.path_length1 + float64(weight)
				if new_length < other_flag.path_length1 {
					other_flag.path_length1 = new_length
					other_flag.prev_edge1 = edge_id
					other_flag.is_shortcut1 = ref.IsShortcut()
					self.startheap.Enqueue(other_id, new_length)
				}
				self.flags[other_id] = other_flag
			})
			self.flags[curr_id] = curr_flag
		}

		if self.mid_id != -1 {
			break
		}
		// from end
		if self.endheap.Len() != 0 {
			curr_id, _ := self.endheap.Dequeue()
			curr_flag := self.flags[curr_id]
			if curr_flag.visited2 {
				continue
			}
			curr_flag.visited2 = true
			if curr_flag.visited1 && self.path_length > (curr_flag.path_length1+curr_flag.path_length2) {
				self.mid_id = curr_id
				self.path_length = curr_flag.path_length1 + curr_flag.path_length2
			}
			explorer.ForAdjacentEdges(curr_id, graph.BACKWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
				edge_id := ref.EdgeID
				other_id := ref.OtherID
				if self.graph.GetNodeLevel(other_id) <= self.graph.GetNodeLevel(curr_id) {
					return
				}
				var other_flag flag_ch
				if self.flags.ContainsKey(other_id) {
					other_flag = self.flags[other_id]
				} else {
					other_flag = flag_ch{path_length1: 1000000, visited1: false, prev_edge1: -1, is_shortcut1: false, path_length2: 1000000, visited2: false, prev_edge2: -1, is_shortcut2: false}
				}
				weight := explorer.GetEdgeWeight(ref)
				if ref.IsShortcut() {
					self.graph.GetEdgesFromShortcut(edge_id, false, handler)
				} else {
					handler(edge_id)
				}
				new_length := curr_flag.path_length2 + float64(weight)
				if new_length < other_flag.path_length2 {
					other_flag.path_length2 = new_length
					other_flag.prev_edge2 = edge_id
					other_flag.is_shortcut2 = ref.IsShortcut()
					self.endheap.Enqueue(other_id, new_length)
				}
				self.flags[other_id] = other_flag
			})
			self.flags[curr_id] = curr_flag
		}
	}
	return true
}

func (self *CH2) GetShortestPath() Path {
	explorer := self.graph.GetGraphExplorer()

	path := NewList[int32](10)
	length := int32(self.flags[self.mid_id].path_length1 + self.flags[self.mid_id].path_length2)
	curr_id := self.mid_id
	for {
		if curr_id == self.start_id {
			break
		}
		curr_flag := self.flags[curr_id]
		if curr_flag.is_shortcut1 {
			self.graph.GetEdgesFromShortcut(curr_flag.prev_edge1, true, func(edge int32) {
				path.Add(edge)
			})
			curr_id = explorer.GetOtherNode(graph.CreateCHShortcutRef(curr_flag.prev_edge1), curr_id)
		} else {
			path.Add(curr_flag.prev_edge1)
			curr_id = explorer.GetOtherNode(graph.CreateEdgeRef(curr_flag.prev_edge1), curr_id)
		}
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	curr_id = self.mid_id
	for {
		if curr_id == self.end_id {
			break
		}
		curr_flag := self.flags[curr_id]
		if curr_flag.is_shortcut2 {
			self.graph.GetEdgesFromShortcut(curr_flag.prev_edge2, false, func(edge int32) {
				path.Add(edge)
			})
			curr_id = explorer.GetOtherNode(graph.CreateCHShortcutRef(curr_flag.prev_edge2), curr_id)
		} else {
			path.Add(curr_flag.prev_edge2)
			curr_id = explorer.GetOtherNode(graph.CreateEdgeRef(curr_flag.prev_edge2), curr_id)
		}
	}
	slog.Debug(fmt.Sprintf("length: %v", length))
	return NewPath(self.graph, path)
}

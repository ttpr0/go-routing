package routing

import (
	"fmt"

	"github.com/ttpr0/go-routing/geo"
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
	"golang.org/x/exp/slog"
)

type flag_a struct {
	path_length float64
	prev_edge   int32
	visited     bool
}

type AStar struct {
	heap      PriorityQueue[int32, float64]
	start_id  int32
	end_id    int32
	end_point geo.Coord
	graph     graph.IGraph
	flags     []flag_a
}

func NewAStar(graph graph.IGraph, start, end int32) *AStar {
	d := AStar{graph: graph, start_id: start, end_id: end}

	d.end_point = d.graph.GetNodeGeom(end)

	flags := make([]flag_a, graph.NodeCount())
	for i := 0; i < len(flags); i++ {
		flags[i].path_length = 1000000000
	}
	flags[start].path_length = 0
	d.flags = flags

	heap := NewPriorityQueue[int32, float64](100)
	heap.Enqueue(d.start_id, 0)
	d.heap = heap

	return &d
}

func (self *AStar) CalcShortestPath() bool {
	explorer := self.graph.GetGraphExplorer()
	for {
		curr_id, ok := self.heap.Dequeue()
		if !ok {
			return false
		}
		if curr_id == self.end_id {
			return true
		}
		//curr := (*d.graph).GetNode(curr_id)
		curr_flag := self.flags[curr_id]
		if curr_flag.visited {
			continue
		}
		curr_flag.visited = true
		explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			if !ref.IsEdge() {
				return
			}
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			//other := (*d.graph).GetNode(other_id)
			other_flag := self.flags[other_id]
			if other_flag.visited {
				return
			}
			lambda := geo.HaversineDistance(geo.Coord(self.graph.GetNodeGeom(other_id)), geo.Coord(self.end_point)) * 3.6 / 130
			new_length := curr_flag.path_length + float64(explorer.GetEdgeWeight(ref))
			if other_flag.path_length > new_length {
				other_flag.prev_edge = edge_id
				other_flag.path_length = new_length
				self.heap.Enqueue(other_id, new_length+lambda)
			}
			self.flags[other_id] = other_flag
		})
		self.flags[curr_id] = curr_flag
	}
}

func (self *AStar) Steps(count int, handler func(int32)) bool {
	explorer := self.graph.GetGraphExplorer()
	for c := 0; c < count; c++ {
		curr_id, ok := self.heap.Dequeue()
		if !ok {
			return false
		}
		if curr_id == self.end_id {
			return false
		}
		//curr := (*d.graph).GetNode(curr_id)
		curr_flag := self.flags[curr_id]
		if curr_flag.visited {
			continue
		}
		curr_flag.visited = true
		explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			if !ref.IsEdge() {
				return
			}
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			//other := (*d.graph).GetNode(other_id)
			other_flag := self.flags[other_id]
			if other_flag.visited {
				return
			}
			handler(edge_id)
			lambda := geo.HaversineDistance(geo.Coord(self.graph.GetNodeGeom(other_id)), geo.Coord(self.end_point)) * 3.6 / 130
			new_length := curr_flag.path_length + float64(explorer.GetEdgeWeight(ref))
			if other_flag.path_length > new_length {
				other_flag.prev_edge = edge_id
				other_flag.path_length = new_length
				self.heap.Enqueue(other_id, new_length+lambda)
			}
			self.flags[other_id] = other_flag
		})
		self.flags[curr_id] = curr_flag
	}
	return true
}

func (self *AStar) GetShortestPath() Path {
	path := make([]int32, 0, 10)
	explorer := self.graph.GetGraphExplorer()
	curr_id := self.end_id
	var edge int32
	for {
		if curr_id == self.start_id {
			break
		}
		edge = self.flags[curr_id].prev_edge
		path = append(path, edge)
		curr_id = explorer.GetOtherNode(graph.CreateEdgeRef(edge), curr_id)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	slog.Debug(fmt.Sprintf("count: %v", len(path)))
	return NewPath(self.graph, path)
}

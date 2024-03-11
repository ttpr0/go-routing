package routing

import (
	"fmt"

	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
	"golang.org/x/exp/slog"
)

type flag_td struct {
	path_length float64
	ref         graph.EdgeRef
	visited     bool
}

type TransitDijkstra struct {
	heap      PriorityQueue[int32, float64]
	start_id  int32
	end_id    int32
	departure int32
	graph     *graph.TransitGraph
	flags     []flag_td
}

func NewTransitDijkstra(g *graph.TransitGraph, start, end int32) *TransitDijkstra {
	d := TransitDijkstra{graph: g, start_id: start, end_id: end, departure: 36000}

	flags := make([]flag_td, g.NodeCount())
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

func (self *TransitDijkstra) CalcShortestPath() bool {
	explorer := self.graph.GetGraphExplorer()
	transit_explorer := self.graph.GetTransitExplorer()

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
		arival := self.departure + int32(curr_flag.path_length)
		if self.graph.IsStop(curr_id) {
			transit_explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
				if ref.IsShortcut() {
					return
				}
				other_id := ref.OtherID
				other_flag := self.flags[other_id]
				if other_flag.visited {
					return
				}
				conn_weight := transit_explorer.GetConnectionWeight(ref, arival)
				if !conn_weight.HasValue() {
					return
				}
				new_length := float64(conn_weight.Value.Arrival - self.departure)
				if other_flag.path_length > new_length {
					other_flag.ref = ref
					other_flag.path_length = new_length
					self.heap.Enqueue(other_id, new_length)
				}
				self.flags[other_id] = other_flag
			})
		}
		explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			other_id := ref.OtherID
			other_flag := self.flags[other_id]
			if other_flag.visited {
				return
			}
			edge_weight := float64(explorer.GetEdgeWeight(ref))
			new_length := curr_flag.path_length + edge_weight
			if other_flag.path_length > new_length {
				other_flag.ref = ref
				other_flag.path_length = new_length
				self.heap.Enqueue(other_id, new_length)
			}
			self.flags[other_id] = other_flag
		})
		self.flags[curr_id] = curr_flag
	}
}

func (self *TransitDijkstra) Steps(count int, handler func(int32)) bool {
	return false
}

func (self *TransitDijkstra) GetShortestPath() Path {
	explorer := self.graph.GetGraphExplorer()

	path := make([]int32, 0, 10)
	length := int32(self.flags[self.end_id].path_length)
	curr_id := self.end_id
	for {
		if curr_id == self.start_id {
			break
		}
		ref := self.flags[curr_id].ref
		if ref.IsEdge() {
			path = append(path, ref.EdgeID)
		}
		curr_id = explorer.GetOtherNode(ref, curr_id)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	slog.Debug(fmt.Sprintf("length: %v", length))
	return NewPath(self.graph, path)
}

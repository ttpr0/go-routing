package routing

import (
	"fmt"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type _FlagBOD struct {
	curr_node   int32
	path_length float64
	prev_edge   int32
	visited     bool
	skip        bool
}

type BODijkstra struct {
	heap     PriorityQueue[_FlagBOD, float64]
	start_id int32
	end_id   int32
	graph    graph.ITiledGraph
	geom     graph.IGeometry
	weight   graph.IWeighting
	flags    Dict[int32, _FlagBOD]
}

func NewBODijkstra(graph graph.ITiledGraph, start, end int32) *BODijkstra {
	d := BODijkstra{graph: graph, start_id: start, end_id: end, geom: graph.GetGeometry(), weight: graph.GetWeighting()}

	flags := NewDict[int32, _FlagBOD](100)
	d.flags = flags

	heap := NewPriorityQueue[_FlagBOD, float64](100)
	heap.Enqueue(_FlagBOD{curr_node: start, prev_edge: -1, path_length: 0, skip: false}, 0)
	d.heap = heap

	return &d
}

func (self *BODijkstra) CalcShortestPath() bool {
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
		edges := self.graph.GetAdjacentEdges(curr_id, graph.FORWARD)
		for {
			ref, ok := edges.Next()
			if !ok {
				break
			}
			if !ref.IsEdge() {
				continue
			}
			if curr_flag.skip && !ref.IsCrossBorder() && !ref.IsSkip() {
				continue
			}
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			var other_flag _FlagBOD
			if self.flags.ContainsKey(other_id) {
				other_flag = self.flags.Get(other_id)
			} else {
				other_flag = _FlagBOD{curr_node: other_id, path_length: 1000000, prev_edge: -1}
			}
			if other_flag.visited {
				continue
			}
			new_length := curr_flag.path_length + float64(self.weight.GetEdgeWeight(edge_id))
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
				other_flag.path_length = new_length
				self.heap.Enqueue(other_flag, new_length)
			}
			self.flags[other_id] = other_flag
		}
	}
}

func (self *BODijkstra) Steps(count int, visitededges *List[geo.CoordArray]) bool {
	for c := 0; c < count; c++ {
		curr_flag, ok := self.heap.Dequeue()
		if !ok {
			return false
		}
		curr_id := curr_flag.curr_node
		if curr_id == self.end_id {
			return false
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
		edges := self.graph.GetAdjacentEdges(curr_id, graph.FORWARD)
		for {
			ref, ok := edges.Next()
			if !ok {
				break
			}
			if !ref.IsEdge() {
				continue
			}
			if curr_flag.skip && !ref.IsCrossBorder() && !ref.IsSkip() {
				continue
			}
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			var other_flag _FlagBOD
			if self.flags.ContainsKey(other_id) {
				other_flag = self.flags.Get(other_id)
			} else {
				other_flag = _FlagBOD{curr_node: other_id, path_length: 1000000, prev_edge: -1}
			}
			if other_flag.visited {
				continue
			}
			visitededges.Add(self.geom.GetEdge(edge_id))
			new_length := curr_flag.path_length + float64(self.weight.GetEdgeWeight(edge_id))
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
				other_flag.path_length = new_length
				self.heap.Enqueue(other_flag, new_length)
			}
			self.flags[other_id] = other_flag
		}
	}
	return true
}

func (self *BODijkstra) GetShortestPath() Path {
	path := make([]int32, 0, 10)
	curr_id := self.end_id
	var edge int32
	for {
		if curr_id == self.start_id {
			break
		}
		edge = self.flags[curr_id].prev_edge
		path = append(path, edge)
		curr_id, _ = self.graph.GetOtherNode(edge, curr_id)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	fmt.Println("count:", len(path))
	return NewPath(self.graph, path)
}

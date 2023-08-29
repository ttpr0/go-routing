package routing

import (
	"fmt"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type flag_d struct {
	path_length float64
	prev_edge   int32
	visited     bool
}

type Dijkstra struct {
	heap     PriorityQueue[int32, float64]
	start_id int32
	end_id   int32
	graph    graph.IGraph
	flags    []flag_d
}

func NewDijkstra(graph graph.IGraph, start, end int32) *Dijkstra {
	d := Dijkstra{graph: graph, start_id: start, end_id: end}

	flags := make([]flag_d, graph.NodeCount())
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

func (self *Dijkstra) CalcShortestPath() bool {
	explorer := self.graph.GetDefaultExplorer()

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
		edges := explorer.GetAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_ALL)
		for {
			ref, ok := edges.Next()
			if !ok {
				break
			}
			if !ref.IsEdge() {
				continue
			}
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			//other := (*d.graph).GetNode(other_id)
			other_flag := self.flags[other_id]
			if other_flag.visited {
				continue
			}
			new_length := curr_flag.path_length + float64(explorer.GetEdgeWeight(ref))
			if other_flag.path_length > new_length {
				other_flag.prev_edge = edge_id
				other_flag.path_length = new_length
				self.heap.Enqueue(other_id, new_length)
			}
			self.flags[other_id] = other_flag
		}
		self.flags[curr_id] = curr_flag
	}
}

func (self *Dijkstra) Steps(count int, visitededges *List[geo.CoordArray]) bool {
	explorer := self.graph.GetDefaultExplorer()

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
		edges := explorer.GetAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_ALL)
		for {
			ref, ok := edges.Next()
			if !ok {
				break
			}
			if !ref.IsEdge() {
				continue
			}
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			//other := (*d.graph).GetNode(other_id)
			other_flag := self.flags[other_id]
			if other_flag.visited {
				continue
			}
			visitededges.Add(self.graph.GetEdgeGeom(edge_id))
			new_length := curr_flag.path_length + float64(explorer.GetEdgeWeight(ref))
			if other_flag.path_length > new_length {
				other_flag.prev_edge = edge_id
				other_flag.path_length = new_length
				self.heap.Enqueue(other_id, new_length)
			}
			self.flags[other_id] = other_flag
		}
		self.flags[curr_id] = curr_flag
	}
	return true
}

func (self *Dijkstra) GetShortestPath() Path {
	explorer := self.graph.GetDefaultExplorer()

	path := make([]int32, 0, 10)
	length := int32(self.flags[self.end_id].path_length)
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
	fmt.Println("length:", length)
	return NewPath(self.graph, path)
}

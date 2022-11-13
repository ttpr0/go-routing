package graph

import (
	"fmt"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type flag_a struct {
	path_length float64
	prev_edge   int32
	distance    float64
	visited     bool
}

type AStar struct {
	heap      util.PriorityQueue[int32]
	start_id  int32
	end_id    int32
	end_point Coord
	graph     IGraph
	geom      IGeometry
	weight    IWeighting
	flags     []flag_a
}

func NewAStar(graph IGraph, start, end int32) *AStar {
	d := AStar{graph: graph, start_id: start, end_id: end, geom: graph.GetGeometry(), weight: graph.GetWeighting()}

	d.end_point = d.geom.GetNode(end)

	flags := make([]flag_a, graph.NodeCount())
	for i := 0; i < len(flags); i++ {
		flags[i].path_length = 1000000000
	}
	flags[start].path_length = 0
	d.flags = flags

	heap := util.NewPriorityQueue[int32](100)
	heap.Enqueue(d.start_id, 0)
	d.heap = heap

	return &d
}

func (self *AStar) CalcShortestPath() bool {
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
		edges := self.graph.GetAdjacentEdges(curr_id)
		for _, edge_id := range edges {
			edge := self.graph.GetEdge(edge_id)
			other_id, dir := self.graph.GetOtherNode(edge_id, curr_id)
			//other := (*d.graph).GetNode(other_id)
			other_flag := self.flags[other_id]
			if other_flag.visited || (edge.Oneway && dir == BACKWARD) {
				continue
			}
			other_flag.distance = geo.HaversineDistance(geo.Coord(self.geom.GetNode(other_id)), geo.Coord(self.end_point)) * 3.6 / 130
			new_length := curr_flag.path_length - curr_flag.distance + float64(self.weight.GetEdgeWeight(edge_id)) + other_flag.distance
			if other_flag.path_length > new_length {
				other_flag.prev_edge = edge_id
				other_flag.path_length = new_length
				self.heap.Enqueue(other_id, float32(new_length))
			}
			self.flags[other_id] = other_flag
		}
		self.flags[curr_id] = curr_flag
	}
}

func (self *AStar) GetShortestPath() Path {
	path := make([]int32, 0, 10)
	curr_id := self.end_id
	var edge int32
	for {
		path = append(path, curr_id)
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

package graph

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type ISPTConsumer interface {
	ConsumePoint(point Coord, value int)
}

type flag_spt struct {
	path_length float64
	prev_edge   int32
	visited     bool
}

type ShortestPathTree struct {
	heap     util.PriorityQueue[int32, float64]
	start_id int32
	max_val  int32
	graph    IGraph
	geom     IGeometry
	weight   IWeighting
	flags    []flag_spt
	consumer ISPTConsumer
}

func NewShortestPathTree(graph IGraph, start, max_val int32, consumer ISPTConsumer) *ShortestPathTree {
	d := ShortestPathTree{
		graph:    graph,
		start_id: start,
		max_val:  max_val,
		geom:     graph.GetGeometry(),
		weight:   graph.GetWeighting(),
	}

	flags := make([]flag_spt, graph.NodeCount())
	for i := 0; i < len(flags); i++ {
		flags[i].path_length = 1000000000
	}
	flags[start].path_length = 0
	d.flags = flags

	heap := util.NewPriorityQueue[int32, float64](100)
	heap.Enqueue(d.start_id, 0)
	d.heap = heap

	d.consumer = consumer

	return &d
}

func (self *ShortestPathTree) CalcShortestPathTree() {
	for {
		curr_id, _ := self.heap.Dequeue()
		//curr := (*d.graph).GetNode(curr_id)
		curr_flag := self.flags[curr_id]
		if curr_flag.path_length > float64(self.max_val) {
			return
		}
		if curr_flag.visited {
			continue
		}
		self.consumer.ConsumePoint(self.geom.GetNode(curr_id), int(curr_flag.path_length))
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
			new_length := curr_flag.path_length + float64(self.weight.GetEdgeWeight(edge_id))
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

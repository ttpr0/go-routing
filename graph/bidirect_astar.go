package graph

import (
	"sync"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type flag_ba struct {
	path_length1 float64
	path_length2 float64
	prev_edge1   int32
	prev_edge2   int32
	visited1     bool
	visited2     bool
	distance1    float64
	distance2    float64
}

type BidirectAStar struct {
	startheap   util.PriorityQueue[int32]
	endheap     util.PriorityQueue[int32]
	mid_id      int32
	start_id    int32
	end_id      int32
	start_point Coord
	end_point   Coord
	graph       IGraph
	geom        IGeometry
	weight      IWeighting
	flags       []flag_ba
}

func NewBidirectAStar(graph IGraph, start, end int32) *BidirectAStar {
	d := BidirectAStar{graph: graph, start_id: start, end_id: end, geom: graph.GetGeometry(), weight: graph.GetWeighting()}

	d.end_point = d.geom.GetNode(end)
	d.start_point = d.geom.GetNode(start)

	flags := make([]flag_ba, graph.NodeCount())
	for i := 0; i < len(flags); i++ {
		flags[i].path_length1 = 1000000000
		flags[i].path_length2 = 1000000000
	}
	flags[start].path_length1 = 0
	flags[end].path_length2 = 0
	d.flags = flags

	startheap := util.NewPriorityQueue[int32](100)
	startheap.Enqueue(d.start_id, 0)
	d.startheap = startheap
	endheap := util.NewPriorityQueue[int32](100)
	endheap.Enqueue(d.end_id, 0)
	d.endheap = endheap

	return &d
}

func (self *BidirectAStar) CalcShortestPath() bool {
	finished := false
	fromStart := func() {
		for !finished {
			curr_id, _ := self.startheap.Dequeue()
			//curr := (*d.graph).GetNode(curr_id)
			if self.flags[curr_id].visited1 {
				continue
			}
			if self.flags[curr_id].visited2 {
				self.mid_id = curr_id
				finished = true
				return
			}
			self.flags[curr_id].visited1 = true
			edges := self.graph.GetAdjacentEdges(curr_id)
			for _, edge_id := range edges {
				edge := self.graph.GetEdge(edge_id)
				other_id, dir := self.graph.GetOtherNode(edge_id, curr_id)
				//other := (*d.graph).GetNode(other_id)
				if self.flags[other_id].visited1 || (edge.Oneway && dir == BACKWARD) {
					continue
				}
				self.flags[other_id].distance1 = geo.HaversineDistance(geo.Coord(self.geom.GetNode(other_id)), geo.Coord(self.end_point)) * 3.6 / 130
				new_length := self.flags[curr_id].path_length1 - self.flags[curr_id].distance1 + float64(self.weight.GetEdgeWeight(edge_id)) + self.flags[other_id].distance1
				if self.flags[other_id].path_length1 > new_length {
					self.flags[other_id].prev_edge1 = edge_id
					self.flags[other_id].path_length1 = new_length
					self.startheap.Enqueue(other_id, float32(new_length))
				}
			}
		}
	}
	fromEnd := func() {
		for !finished {
			curr_id, _ := self.endheap.Dequeue()
			//curr := (*d.graph).GetNode(curr_id)
			if self.flags[curr_id].visited2 {
				continue
			}
			if self.flags[curr_id].visited1 {
				self.mid_id = curr_id
				finished = true
				return
			}
			self.flags[curr_id].visited2 = true
			edges := self.graph.GetAdjacentEdges(curr_id)
			for _, edge_id := range edges {
				edge := self.graph.GetEdge(edge_id)
				other_id, dir := self.graph.GetOtherNode(edge_id, curr_id)
				//other := (*d.graph).GetNode(other_id)
				if self.flags[other_id].visited2 || (edge.Oneway && dir == BACKWARD) {
					continue
				}
				self.flags[other_id].distance2 = geo.HaversineDistance(geo.Coord(self.geom.GetNode(other_id)), geo.Coord(self.start_point)) * 3.6 / 130
				new_length := self.flags[curr_id].path_length2 - self.flags[curr_id].distance2 + float64(self.weight.GetEdgeWeight(edge_id)) + self.flags[other_id].distance2
				if self.flags[other_id].path_length2 > new_length {
					self.flags[other_id].prev_edge2 = edge_id
					self.flags[other_id].path_length2 = new_length
					self.endheap.Enqueue(other_id, float32(new_length))
				}
			}
		}
	}
	wg := sync.WaitGroup{}
	failure := false
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			err := recover()
			if err != nil {
				failure = true
			}
		}()
		fromStart()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			err := recover()
			if err != nil {
				failure = true
			}
		}()
		fromEnd()
	}()
	wg.Wait()
	if failure {
		return false
	}
	return true
}

func (self *BidirectAStar) GetShortestPath() Path {
	path := make([]int32, 0, 10)
	curr_id := self.mid_id
	var edge int32
	for {
		path = append(path, curr_id)
		if curr_id == self.start_id {
			break
		}
		edge = self.flags[curr_id].prev_edge1
		path = append(path, edge)
		curr_id, _ = self.graph.GetOtherNode(edge, curr_id)
	}
	path = path[1:]
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	curr_id = self.mid_id
	for {
		path = append(path, curr_id)
		if curr_id == self.end_id {
			break
		}
		edge = self.flags[curr_id].prev_edge2
		path = append(path, edge)
		curr_id, _ = self.graph.GetOtherNode(edge, curr_id)
	}
	return NewPath(self.graph, path)
}

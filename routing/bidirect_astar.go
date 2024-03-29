package routing

import (
	"github.com/ttpr0/go-routing/geo"
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
)

type flag_ba struct {
	path_length1 float64
	path_length2 float64
	prev_edge1   int32
	prev_edge2   int32
	visited1     bool
	visited2     bool
	lambda1      float64
	lambda2      float64
}

type BidirectAStar struct {
	startheap   PriorityQueue[int32, float64]
	endheap     PriorityQueue[int32, float64]
	mid_id      int32
	start_id    int32
	end_id      int32
	start_point geo.Coord
	end_point   geo.Coord
	graph       graph.IGraph
	flags       []flag_ba
}

func NewBidirectAStar(graph graph.IGraph, start, end int32) *BidirectAStar {
	d := BidirectAStar{graph: graph, start_id: start, end_id: end}

	d.end_point = d.graph.GetNodeGeom(end)
	d.start_point = d.graph.GetNodeGeom(start)

	flags := make([]flag_ba, graph.NodeCount())
	for i := 0; i < len(flags); i++ {
		flags[i].path_length1 = 1000000000
		flags[i].path_length2 = 1000000000
	}
	flags[start].path_length1 = 0
	flags[end].path_length2 = 0
	d.flags = flags

	startheap := NewPriorityQueue[int32, float64](100)
	startheap.Enqueue(d.start_id, 0)
	d.startheap = startheap
	endheap := NewPriorityQueue[int32, float64](100)
	endheap.Enqueue(d.end_id, 0)
	d.endheap = endheap

	return &d
}

func (self *BidirectAStar) CalcShortestPath() bool {
	explorer := self.graph.GetGraphExplorer()

	lambda_route := geo.HaversineDistance(geo.Coord(self.end_point), geo.Coord(self.start_point))
	finished := false
	for !finished {
		// from start
		curr_id, _ := self.startheap.Dequeue()
		//curr := (*d.graph).GetNode(curr_id)
		curr_flag := self.flags[curr_id]
		if curr_flag.visited1 {
			continue
		}
		curr_flag.visited1 = true
		explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			if !ref.IsEdge() {
				return
			}
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			//other := (*d.graph).GetNode(other_id)
			other_flag := self.flags[other_id]
			if other_flag.visited1 {
				return
			}
			other_flag.lambda1 = geo.HaversineDistance(geo.Coord(self.graph.GetNodeGeom(other_id)), geo.Coord(self.end_point)) * 3.6 / 130
			other_flag.lambda2 = geo.HaversineDistance(geo.Coord(self.graph.GetNodeGeom(other_id)), geo.Coord(self.start_point)) * 3.6 / 130
			lambda := (other_flag.lambda1-other_flag.lambda2)/2 + lambda_route/2
			new_length := curr_flag.path_length1 + float64(explorer.GetEdgeWeight(ref))
			if other_flag.visited2 {
				shortest := new_length + other_flag.path_length2
				top1, ok1 := self.startheap.Peek()
				top2, ok2 := self.endheap.Peek()
				if ok1 && ok2 && float64(top1+top2)+lambda_route >= shortest {
					other_flag.prev_edge1 = edge_id
					other_flag.path_length1 = new_length
					self.flags[other_id] = other_flag
					self.mid_id = other_id
					finished = true
				}
			}
			if other_flag.path_length1 > new_length {
				other_flag.prev_edge1 = edge_id
				other_flag.path_length1 = new_length
				self.startheap.Enqueue(other_id, new_length+lambda)
			}
			self.flags[other_id] = other_flag
		})
		self.flags[curr_id] = curr_flag

		if finished {
			break
		}
		// from end
		curr_id, _ = self.endheap.Dequeue()
		//curr := (*d.graph).GetNode(curr_id)
		curr_flag = self.flags[curr_id]
		if curr_flag.visited2 {
			continue
		}
		curr_flag.visited2 = true
		explorer.ForAdjacentEdges(curr_id, graph.BACKWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			if !ref.IsEdge() {
				return
			}
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			//other := (*d.graph).GetNode(other_id)
			other_flag := self.flags[other_id]
			if other_flag.visited2 {
				return
			}
			other_flag.lambda1 = geo.HaversineDistance(geo.Coord(self.graph.GetNodeGeom(other_id)), geo.Coord(self.end_point)) * 3.6 / 130
			other_flag.lambda2 = geo.HaversineDistance(geo.Coord(self.graph.GetNodeGeom(other_id)), geo.Coord(self.start_point)) * 3.6 / 130
			lambda := (other_flag.lambda2-other_flag.lambda1)/2 + lambda_route/2
			new_length := curr_flag.path_length2 + float64(explorer.GetEdgeWeight(ref))
			if other_flag.visited1 {
				shortest := new_length + other_flag.path_length1
				top1, ok1 := self.startheap.Peek()
				top2, ok2 := self.endheap.Peek()
				if ok1 && ok2 && float64(top1+top2)+lambda_route >= shortest {
					other_flag.prev_edge2 = edge_id
					other_flag.path_length2 = new_length
					self.flags[other_id] = other_flag
					self.mid_id = other_id
					finished = true
				}
			}
			if other_flag.path_length2 > new_length {
				other_flag.prev_edge2 = edge_id
				other_flag.path_length2 = new_length
				self.endheap.Enqueue(other_id, new_length+lambda)
			}
			self.flags[other_id] = other_flag
		})
		self.flags[curr_id] = curr_flag
	}

	return true
}

func (self *BidirectAStar) Steps(count int, handler func(int32)) bool {
	explorer := self.graph.GetGraphExplorer()

	lambda_route := geo.HaversineDistance(geo.Coord(self.end_point), geo.Coord(self.start_point))
	is_finished := false
	for c := 0; c < count; c++ {
		curr_id, _ := self.startheap.Dequeue()
		//curr := (*d.graph).GetNode(curr_id)
		curr_flag := self.flags[curr_id]
		if curr_flag.visited1 {
			continue
		}
		if curr_flag.visited2 {
			self.mid_id = curr_id
			is_finished = true
		}
		curr_flag.visited1 = true
		explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			if !ref.IsEdge() {
				return
			}
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			//other := (*d.graph).GetNode(other_id)
			other_flag := self.flags[other_id]
			if other_flag.visited1 {
				return
			}
			handler(edge_id)
			other_flag.lambda1 = geo.HaversineDistance(geo.Coord(self.graph.GetNodeGeom(other_id)), geo.Coord(self.end_point)) * 3.6 / 130
			other_flag.lambda2 = geo.HaversineDistance(geo.Coord(self.graph.GetNodeGeom(other_id)), geo.Coord(self.start_point)) * 3.6 / 130
			lambda := (other_flag.lambda1-other_flag.lambda2)/2 + lambda_route/2
			new_length := curr_flag.path_length1 + float64(explorer.GetEdgeWeight(ref))
			if other_flag.visited2 {
				shortest := new_length + other_flag.path_length2
				top1, ok1 := self.startheap.Peek()
				top2, ok2 := self.endheap.Peek()
				if ok1 && ok2 && float64(top1+top2)+lambda_route >= shortest {
					other_flag.prev_edge1 = edge_id
					other_flag.path_length1 = new_length
					self.flags[other_id] = other_flag
					self.mid_id = other_id
					is_finished = true
				}
			}
			if other_flag.path_length1 > new_length {
				other_flag.prev_edge1 = edge_id
				other_flag.path_length1 = new_length
				self.startheap.Enqueue(other_id, new_length+lambda)
			}
			self.flags[other_id] = other_flag
		})
		self.flags[curr_id] = curr_flag
		if is_finished {
			break
		}

		curr_id, _ = self.endheap.Dequeue()
		//curr := (*d.graph).GetNode(curr_id)
		curr_flag = self.flags[curr_id]
		if curr_flag.visited2 {
			continue
		}
		if curr_flag.visited1 {
			self.mid_id = curr_id
			return false
		}
		curr_flag.visited2 = true
		explorer.ForAdjacentEdges(curr_id, graph.BACKWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			if !ref.IsEdge() {
				return
			}
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			//other := (*d.graph).GetNode(other_id)
			other_flag := self.flags[other_id]
			if other_flag.visited2 {
				return
			}
			handler(edge_id)
			other_flag.lambda1 = geo.HaversineDistance(geo.Coord(self.graph.GetNodeGeom(other_id)), geo.Coord(self.end_point)) * 3.6 / 130
			other_flag.lambda2 = geo.HaversineDistance(geo.Coord(self.graph.GetNodeGeom(other_id)), geo.Coord(self.start_point)) * 3.6 / 130
			lambda := (other_flag.lambda2-other_flag.lambda1)/2 + lambda_route/2
			new_length := curr_flag.path_length2 + float64(explorer.GetEdgeWeight(ref))
			if other_flag.visited1 {
				shortest := new_length + other_flag.path_length1
				top1, ok1 := self.startheap.Peek()
				top2, ok2 := self.endheap.Peek()
				if ok1 && ok2 && float64(top1+top2)+lambda_route >= shortest {
					other_flag.prev_edge2 = edge_id
					other_flag.path_length2 = new_length
					self.flags[other_id] = other_flag
					self.mid_id = other_id
					is_finished = true
				}
			}
			if other_flag.path_length2 > new_length {
				other_flag.prev_edge2 = edge_id
				other_flag.path_length2 = new_length
				self.endheap.Enqueue(other_id, new_length+lambda)
			}
			self.flags[other_id] = other_flag
		})
		self.flags[curr_id] = curr_flag
		if is_finished {
			break
		}
	}
	if is_finished {
		return false
	}
	return true
}

func (self *BidirectAStar) GetShortestPath() Path {
	explorer := self.graph.GetGraphExplorer()

	path := make([]int32, 0, 10)
	curr_id := self.mid_id
	var edge int32
	for {
		if curr_id == self.start_id {
			break
		}
		edge = self.flags[curr_id].prev_edge1
		path = append(path, edge)
		curr_id = explorer.GetOtherNode(graph.CreateEdgeRef(edge), curr_id)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	curr_id = self.mid_id
	for {
		if curr_id == self.end_id {
			break
		}
		edge = self.flags[curr_id].prev_edge2
		path = append(path, edge)
		curr_id = explorer.GetOtherNode(graph.CreateEdgeRef(edge), curr_id)
	}
	return NewPath(self.graph, path)
}

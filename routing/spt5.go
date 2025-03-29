package routing

import (
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
)

type ShortestPathTree5 struct {
	heap  PriorityQueue[int32, float64]
	graph graph.IGraph
	flags []flag_spt
}

func NewShortestPathTree5(graph graph.IGraph) *ShortestPathTree5 {
	d := ShortestPathTree5{
		graph: graph,
	}

	flags := make([]flag_spt, graph.NodeCount())
	for i := 0; i < len(flags); i++ {
		flags[i].path_length = 1000000000
	}
	d.flags = flags

	heap := NewPriorityQueue[int32, float64](100)
	d.heap = heap

	return &d
}

func (self *ShortestPathTree5) CalcShortestPathTree(start int32, max_val int32, consumer ISPTConsumer) {
	self.heap.Enqueue(start, 0)
	self.flags[start].path_length = 0
	explorer := self.graph.GetGraphExplorer()

	for {
		curr_id, _ := self.heap.Dequeue()
		//curr := (*d.graph).GetNode(curr_id)
		curr_flag := self.flags[curr_id]
		if curr_flag.path_length > float64(max_val) {
			return
		}
		if curr_flag.visited {
			continue
		}
		consumer.ConsumePoint(self.graph.GetNodeGeom(curr_id), int(curr_flag.path_length))
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
			new_length := curr_flag.path_length + float64(explorer.GetEdgeWeight(ref))
			if other_flag.path_length > new_length {
				other_flag.prev_edge = edge_id
				other_flag.path_length = new_length
				self.heap.Enqueue(other_id, new_length)
				consumer.ConsumeEdge(edge_id, int(curr_flag.path_length), int(new_length))
			}
			self.flags[other_id] = other_flag
		})
		self.flags[curr_id] = curr_flag
	}
}

package routing

import (
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
)

type FlagSPT struct {
	PathLength float64
	PrevEdge   int32
	Visited    bool
}

type ShortestPathTree2 struct {
	heap      PriorityQueue[int32, float64]
	start_id  int32
	max_range float64
	graph     graph.IGraph
	flags     []FlagSPT
}

func NewSPT2(graph graph.IGraph) *ShortestPathTree2 {
	d := ShortestPathTree2{graph: graph}

	flags := make([]FlagSPT, graph.NodeCount())
	d.flags = flags

	heap := NewPriorityQueue[int32, float64](100)
	d.heap = heap

	return &d
}

func (self *ShortestPathTree2) Init(start int32, max_range float64) {
	self.max_range = max_range
	self.start_id = start
	self.heap.Clear()
	self.heap.Enqueue(start, 0)
	for i := 0; i < len(self.flags); i++ {
		self.flags[i] = FlagSPT{1000000000, -1, false}
	}
	self.flags[start].PathLength = 0
}
func (self *ShortestPathTree2) CalcSPT() {
	explorer := self.graph.GetGraphExplorer()

	for {
		curr_id, ok := self.heap.Dequeue()
		if !ok {
			return
		}
		//curr := (*d.graph).GetNode(curr_id)
		curr_flag := self.flags[curr_id]
		if curr_flag.Visited {
			continue
		}
		curr_flag.Visited = true
		self.flags[curr_id] = curr_flag
		if curr_flag.PathLength > self.max_range {
			return
		}
		explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			if !ref.IsEdge() {
				return
			}
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			//other := (*d.graph).GetNode(other_id)
			other_flag := self.flags[other_id]
			if other_flag.Visited {
				return
			}
			new_length := curr_flag.PathLength + float64(explorer.GetEdgeWeight(ref))
			if other_flag.PathLength > new_length {
				other_flag.PrevEdge = edge_id
				other_flag.PathLength = new_length
				self.heap.Enqueue(other_id, new_length)
			}
			self.flags[other_id] = other_flag
		})
	}
}

func (self *ShortestPathTree2) GetSPT() []FlagSPT {
	return self.flags
}

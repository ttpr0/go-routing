package algorithm

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type FlagSPT struct {
	PathLength float64
	PrevEdge   int32
	Visited    bool
}

type RPHAST struct {
	heap      PriorityQueue[int32, float64]
	start_id  int32
	subset    Array[bool]
	max_range float64
	graph     graph.ICHGraph
	geom      graph.IGeometry
	weight    graph.IWeighting
	flags     []FlagSPT
}

func NewRPHAST(graph graph.ICHGraph, graph_subset Array[bool]) *RPHAST {
	d := RPHAST{graph: graph, geom: graph.GetGeometry(), weight: graph.GetWeighting()}

	d.subset = graph_subset

	flags := make([]FlagSPT, graph.NodeCount())
	d.flags = flags

	heap := NewPriorityQueue[int32, float64](100)
	d.heap = heap

	return &d
}

func (self *RPHAST) Init(start int32, max_range float64) {
	self.start_id = start
	self.max_range = max_range
	self.heap.Clear()
	self.heap.Enqueue(start, 0)
	for i := 0; i < len(self.flags); i++ {
		self.flags[i] = FlagSPT{1000000000, -1, false}
	}
	self.flags[start].PathLength = 0
}
func (self *RPHAST) CalcSPT() {
	for {
		curr_id, ok := self.heap.Dequeue()
		if !ok {
			break
		}
		//curr := (*d.graph).GetNode(curr_id)
		curr_flag := self.flags[curr_id]
		if curr_flag.Visited {
			continue
		}
		curr_flag.Visited = true
		edges := self.graph.GetAdjacentEdges(curr_id, graph.FORWARD)
		for {
			ref, ok := edges.Next()
			if !ok {
				break
			}
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			if self.graph.GetNodeLevel(other_id) <= self.graph.GetNodeLevel(curr_id) {
				continue
			}
			other_flag := self.flags[other_id]
			if other_flag.Visited {
				continue
			}
			new_length := curr_flag.PathLength + float64(ref.Weight)
			if other_flag.PathLength > new_length {
				other_flag.PrevEdge = edge_id
				other_flag.PathLength = new_length
				self.heap.Enqueue(other_id, new_length)
			}
			self.flags[other_id] = other_flag
		}
		self.flags[curr_id] = curr_flag
	}

	for i := 0; i < len(self.flags); i++ {
		curr_len := self.flags[i].PathLength
		if curr_len > self.max_range {
			continue
		}
		edges := self.graph.GetAdjacentEdges(int32(i), graph.FORWARD)
		for {
			ref, ok := edges.Next()
			if !ok {
				break
			}
			other_id := ref.OtherID
			if !self.subset[other_id] {
				continue
			}
			if self.graph.GetNodeLevel(other_id) >= self.graph.GetNodeLevel(int32(i)) {
				continue
			}
			other_flag := self.flags[other_id]
			if other_flag.PathLength > (curr_len + float64(ref.Weight)) {
				other_flag.PathLength = curr_len + float64(ref.Weight)
				self.flags[other_id] = other_flag
			}
		}
	}
}

func (self *RPHAST) GetSPT() []FlagSPT {
	return self.flags
}

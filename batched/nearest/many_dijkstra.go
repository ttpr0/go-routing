package nearest

import (
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
)

func NewManyDijkstra(g graph.IGraph, max_range int32) *ManyDijkstra {
	return &ManyDijkstra{g: g, max_range: max_range}
}

type ManyDijkstra struct {
	g         graph.IGraph
	max_range int32
}

func (self *ManyDijkstra) CreateSolver() ISolver {
	node_flags := NewFlags[DistFlag](int32(self.g.NodeCount()), DistFlag{1000000, -1})
	return &ManyDijkstraSolver{
		g:          self.g,
		node_flags: node_flags,
		max_range:  self.max_range,
	}
}

type ManyDijkstraSolver struct {
	g          graph.IGraph
	node_flags Flags[DistFlag]
	max_range  int32
}

func (self *ManyDijkstraSolver) CalcNearestNeighbours(sources List[Array[Tuple[int32, int32]]]) error {
	self.node_flags.Reset()
	_CalcManyDijkstra(self.g, sources, self.node_flags, self.max_range)
	return nil
}

func (self *ManyDijkstraSolver) GetNeighbour(node int32) int32 {
	flag := self.node_flags.Get(node)
	return flag.Source
}
func (self *ManyDijkstraSolver) GetDistance(node int32) int32 {
	flag := self.node_flags.Get(node)
	return flag.Dist
}

func _CalcManyDijkstra(g graph.IGraph, sources List[Array[Tuple[int32, int32]]], node_flags Flags[DistFlag], max_range int32) {
	heap := NewPriorityQueue[PQItem, int32](100)
	explorer := g.GetGraphExplorer()

	for source_id, source := range sources {
		for _, item := range source {
			start := item.A
			dist := item.B
			start_flag := node_flags.Get(start)
			if start_flag.Dist > dist {
				start_flag.Dist = dist
				start_flag.Source = int32(source_id)
				heap.Enqueue(PQItem{start, dist}, dist)
			}
		}
	}

	for {
		curr_item, ok := heap.Dequeue()
		if !ok {
			break
		}
		curr_id := curr_item.item
		curr_dist := curr_item.dist
		curr_flag := node_flags.Get(curr_id)
		if curr_flag.Dist < curr_dist {
			continue
		}
		explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_EDGES, func(ref graph.EdgeRef) {
			other_id := ref.OtherID
			other_flag := node_flags.Get(other_id)
			new_length := curr_flag.Dist + explorer.GetEdgeWeight(ref)
			if new_length > max_range {
				return
			}
			if other_flag.Dist > new_length {
				other_flag.Dist = new_length
				other_flag.Source = curr_flag.Source
				heap.Enqueue(PQItem{other_id, new_length}, new_length)
			}
		})
	}
}

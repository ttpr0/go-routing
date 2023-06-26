package routing

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type FlagSPT3 struct {
	PathLength float64
	PrevEdge   int32
	Visited    bool
	skip       bool
}

type ShortestPathTree3 struct {
	heap         PriorityQueue[int32, float64]
	start_id     int32
	max_range    float64
	graph        graph.ITiledGraph
	geom         graph.IGeometry
	weight       graph.IWeighting
	flags        []FlagSPT3
	active_tiles Dict[int16, bool]
}

func NewSPT3(graph graph.ITiledGraph) *ShortestPathTree3 {
	d := ShortestPathTree3{graph: graph, geom: graph.GetGeometry(), weight: graph.GetWeighting()}

	flags := make([]FlagSPT3, graph.NodeCount())
	d.flags = flags

	heap := NewPriorityQueue[int32, float64](100)
	d.heap = heap

	d.active_tiles = NewDict[int16, bool](10)

	return &d
}

func (self *ShortestPathTree3) Init(start int32, max_range float64) {
	self.max_range = max_range
	self.start_id = start
	self.heap.Clear()
	self.heap.Enqueue(start, 0)
	for i := 0; i < len(self.flags); i++ {
		self.flags[i] = FlagSPT3{1000000000, -1, false, false}
	}
	self.flags[start].PathLength = 0
	self.active_tiles.Clear()
}
func (self *ShortestPathTree3) CalcSPT() {
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
			//other := (*d.graph).GetNode(other_id)
			other_flag := self.flags[other_id]
			if other_flag.Visited {
				continue
			}
			new_length := curr_flag.PathLength + float64(self.weight.GetEdgeWeight(edge_id))
			if other_flag.PathLength > new_length {
				if ref.IsCrossBorder() {
					tile_id := self.graph.GetNodeTile(other_id)
					self.active_tiles[tile_id] = true
					if self.graph.GetNodeTile(self.start_id) == tile_id {
						other_flag.skip = false
					} else {
						other_flag.skip = true
					}
				} else {
					other_flag.skip = curr_flag.skip
				}
				other_flag.PrevEdge = edge_id
				other_flag.PathLength = new_length
				self.heap.Enqueue(other_id, new_length)
			}
			self.flags[other_id] = other_flag
		}
	}
}
func (self *ShortestPathTree3) GetSPT() []FlagSPT3 {
	return self.flags
}
func (self *ShortestPathTree3) GetActiveTiles() Dict[int16, bool] {
	return self.active_tiles
}

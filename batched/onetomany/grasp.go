package onetomany

import (
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
)

// dont use this for now
func NewGRASP(g graph.ITiledGraph, target_nodes Array[int32], max_range int32) *GRASP {
	tilecount := g.TileCount() + 2
	active_tiles := NewArray[bool](int(tilecount))
	for i := 0; i < target_nodes.Length(); i++ {
		tile := g.GetNodeTile(target_nodes[i])
		active_tiles[tile] = true
	}

	return &GRASP{
		g:            g,
		max_range:    max_range,
		active_tiles: active_tiles,
	}
}

type GRASP struct {
	g            graph.ITiledGraph
	max_range    int32
	active_tiles Array[bool]
}

func (self *GRASP) CreateSolver() ISolver {
	node_flags := NewFlags[DistFlag](int32(self.g.NodeCount()), DistFlag{1000000})
	found_tiles := NewArray[bool](self.active_tiles.Length())
	return &GRASPSolver{
		g:            self.g,
		max_range:    self.max_range,
		active_tiles: self.active_tiles,
		found_tiles:  found_tiles,
		node_flags:   node_flags,
	}
}

type GRASPSolver struct {
	g            graph.ITiledGraph
	max_range    int32
	active_tiles Array[bool]
	found_tiles  Array[bool]
	node_flags   Flags[DistFlag]
}

// CalcDiatanceFromStarts implements ISolver.
func (self *GRASPSolver) CalcDistanceFromStarts(starts Array[Tuple[int32, int32]]) error {
	self.node_flags.Reset()
	for i := 0; i < self.found_tiles.Length(); i++ {
		self.found_tiles[i] = false
	}
	_CalcGRASP(self.g, starts, self.max_range, self.node_flags, self.active_tiles, self.found_tiles)
	return nil
}

// CalcDistanceFromStart implements ISolver.
func (self *GRASPSolver) CalcDistanceFromStart(start int32) error {
	self.node_flags.Reset()
	for i := 0; i < self.found_tiles.Length(); i++ {
		self.found_tiles[i] = false
	}
	starts := [1]Tuple[int32, int32]{MakeTuple(start, int32(0))}
	_CalcGRASP(self.g, starts[:], self.max_range, self.node_flags, self.active_tiles, self.found_tiles)
	return nil
}

// GetDistance implements ISolver.
func (self *GRASPSolver) GetDistance(node int32) int32 {
	return self.node_flags.Get(node).Dist
}

func _CalcGRASP(g graph.ITiledGraph, starts Array[Tuple[int32, int32]], max_range int32, node_flags Flags[DistFlag], active_tiles, found_tiles Array[bool]) {
	heap := NewPriorityQueue[PQItem, int32](100)
	explorer := g.GetGraphExplorer()

	// TODO: fix start tiles
	s_tile := g.GetNodeTile(starts[0].A)
	for _, item := range starts {
		start := item.A
		dist := item.B
		start_flag := node_flags.Get(start)
		start_flag.Dist = dist
		heap.Enqueue(PQItem{start, dist}, dist)
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
		explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_UPWARDS, func(ref graph.EdgeRef) {
			other_id := ref.OtherID
			other_flag := node_flags.Get(other_id)
			new_length := curr_flag.Dist + explorer.GetEdgeWeight(ref)
			if other_flag.Dist > new_length {
				other_flag.Dist = new_length
				heap.Enqueue(PQItem{other_id, new_length}, new_length)
			}
		})

		curr_tile := g.GetNodeTile(curr_id)
		handler := func(ref graph.EdgeRef) {
			other_id := ref.OtherID
			other_flag := node_flags.Get(other_id)
			new_length := curr_flag.Dist + explorer.GetEdgeWeight(ref)
			if new_length > max_range {
				return
			}
			if other_flag.Dist > new_length {
				other_flag.Dist = new_length
				heap.Enqueue(PQItem{other_id, new_length}, new_length)
			}
		}
		if curr_tile == s_tile {
			explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_EDGES, handler)
		} else {
			found_tiles[curr_tile] = true
			explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_SKIP, handler)
		}

	}
	for i := 0; i < active_tiles.Length(); i++ {
		if !active_tiles[i] || !found_tiles[i] {
			continue
		}
		down_edges, _ := g.GetIndexEdges(int16(i), graph.FORWARD)
		for j := 0; j < down_edges.Length(); j++ {
			edge := down_edges[i]
			curr_flag := node_flags.Get(edge.From)
			curr_len := curr_flag.Dist
			new_len := curr_len + edge.Weight
			if new_len > max_range {
				continue
			}
			other_flag := node_flags.Get(edge.To)
			if other_flag.Dist > new_len {
				other_flag.Dist = new_len
			}
		}
	}
}

package graph

import (
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

// Reorders nodes of graph g inplace.
// Contraction Hierarchy has to be built with tiles.
func CreateCHGraph4(g *CHGraph, node_tiles Array[int16]) *CHGraph4 {
	// Reorder nodes
	order := ComputeTileLevelOrdering(g, node_tiles)
	mapping := NewArray[int32](g.NodeCount())
	reordered_node_tiles := NewArray[int16](g.NodeCount())
	for i := 0; i < g.NodeCount(); i++ {
		id := order[i]
		new_id := int32(i)
		mapping[id] = new_id
		reordered_node_tiles[new_id] = node_tiles[id]
	}
	ReorderCHGraph(g, mapping)
	node_tiles = reordered_node_tiles
	is_border := _IsBorderNode2(g, node_tiles)

	// initialize down edges lists
	fwd_down_edges := NewList[CHEdge4](g.NodeCount())
	bwd_down_edges := NewList[CHEdge4](g.NodeCount())

	explorer := g.GetDefaultExplorer()
	border_count := 0

	// add overlay down-edges
	fwd_down_edges.Add(CHEdge4{
		IsDummy: true,
	})
	bwd_down_edges.Add(CHEdge4{
		IsDummy: true,
	})
	fwd_other_edges := NewDict[int16, List[CHEdge4]](100)
	bwd_other_edges := NewDict[int16, List[CHEdge4]](100)
	for i := 0; i < g.NodeCount(); i++ {
		this_id := int32(i)
		this_tile := node_tiles[this_id]
		if !is_border[this_id] {
			border_count = i + 1
			break
		}
		explorer.ForAdjacentEdges(this_id, FORWARD, ADJACENT_DOWNWARDS, func(ref EdgeRef) {
			other_id := ref.OtherID
			other_tile := node_tiles[other_id]
			edge := CHEdge4{
				From:   this_id,
				To:     other_id,
				Weight: explorer.GetEdgeWeight(ref),
				ToTile: other_tile,
			}
			if !is_border[other_id] {
				var edges List[CHEdge4]
				if fwd_other_edges.ContainsKey(this_tile) {
					edges = fwd_other_edges[this_tile]
				} else {
					edges = NewList[CHEdge4](10)
				}
				edges.Add(edge)
				fwd_other_edges[this_tile] = edges
			} else {
				fwd_down_edges.Add(edge)
			}
		})
		explorer.ForAdjacentEdges(this_id, BACKWARD, ADJACENT_DOWNWARDS, func(ref EdgeRef) {
			other_id := ref.OtherID
			other_tile := node_tiles[other_id]
			edge := CHEdge4{
				From:   this_id,
				To:     other_id,
				Weight: explorer.GetEdgeWeight(ref),
				ToTile: other_tile,
			}
			if !is_border[other_id] {
				var edges List[CHEdge4]
				if bwd_other_edges.ContainsKey(this_tile) {
					edges = bwd_other_edges[this_tile]
				} else {
					edges = NewList[CHEdge4](10)
				}
				edges.Add(edge)
				bwd_other_edges[this_tile] = edges
			} else {
				bwd_down_edges.Add(edge)
			}
		})
	}
	// add other down edges
	curr_tile := int16(-1)
	for i := border_count; i < g.NodeCount(); i++ {
		this_id := int32(i)
		this_tile := node_tiles[this_id]
		if this_tile != curr_tile {
			fwd_down_edges.Add(CHEdge4{
				ToTile:  this_tile,
				IsDummy: true,
			})
			bwd_down_edges.Add(CHEdge4{
				ToTile:  this_tile,
				IsDummy: true,
			})
			curr_tile = this_tile
			if fwd_other_edges.ContainsKey(this_tile) {
				edges := fwd_other_edges[this_tile]
				for _, edge := range edges {
					fwd_down_edges.Add(edge)
				}
			}
			if bwd_other_edges.ContainsKey(this_tile) {
				edges := bwd_other_edges[this_tile]
				for _, edge := range edges {
					bwd_down_edges.Add(edge)
				}
			}
		}
		explorer.ForAdjacentEdges(this_id, FORWARD, ADJACENT_DOWNWARDS, func(ref EdgeRef) {
			other_id := ref.OtherID
			fwd_down_edges.Add(CHEdge4{
				From:   this_id,
				To:     other_id,
				Weight: explorer.GetEdgeWeight(ref),
			})
		})
		explorer.ForAdjacentEdges(this_id, BACKWARD, ADJACENT_DOWNWARDS, func(ref EdgeRef) {
			other_id := ref.OtherID
			bwd_down_edges.Add(CHEdge4{
				From:   this_id,
				To:     other_id,
				Weight: explorer.GetEdgeWeight(ref),
			})
		})
	}

	// set count in dummy edges
	fwd_id := 0
	fwd_count := 0
	for i := 0; i < fwd_down_edges.Length(); i++ {
		edge := fwd_down_edges[i]
		if edge.IsDummy {
			// set count in previous dummy
			fwd_down_edges[fwd_id].To = int32(fwd_count)
			// reset count
			fwd_id = i
			fwd_count = 0
			continue
		}
		fwd_count += 1
	}
	fwd_down_edges[fwd_id].To = int32(fwd_count)
	bwd_id := 0
	bwd_count := 0
	for i := 0; i < bwd_down_edges.Length(); i++ {
		edge := bwd_down_edges[i]
		if edge.IsDummy {
			// set count in previous dummy
			bwd_down_edges[bwd_id].To = int32(bwd_count)
			// reset count
			bwd_id = i
			bwd_count = 0
			continue
		}
		bwd_count += 1
	}
	bwd_down_edges[bwd_id].To = int32(bwd_count)

	return &CHGraph4{
		CHGraph: *g,

		node_tiles:     node_tiles,
		fwd_down_edges: Array[CHEdge4](fwd_down_edges),
		bwd_down_edges: Array[CHEdge4](bwd_down_edges),
	}
}

type CHGraph4 struct {
	CHGraph

	// tile of every node
	node_tiles Array[int16]
	// stores all fowwards-down edges
	fwd_down_edges Array[CHEdge4]
	// stores all backwards-down edges
	bwd_down_edges Array[CHEdge4]
}

func (self *CHGraph4) GetNodeTile(node int32) int16 {
	return self.node_tiles[node]
}
func (self *CHGraph4) TileCount() int {
	max := int16(0)
	for i := 0; i < len(self.node_tiles); i++ {
		tile := self.node_tiles[i]
		if tile > max {
			max = tile
		}
	}
	return int(max + 1)
}
func (self *CHGraph4) GetDownEdges(dir Direction) Array[CHEdge4] {
	if dir == FORWARD {
		return self.fwd_down_edges
	} else {
		return self.bwd_down_edges
	}
}

type CHEdge4 struct {
	From    int32
	To      int32
	Weight  int32
	ToTile  int16
	IsDummy bool
}

func _IsBorderNode2(graph ICHGraph, node_tiles Array[int16]) Array[bool] {
	is_border := NewArray[bool](graph.NodeCount())

	explorer := graph.GetDefaultExplorer()
	for i := 0; i < graph.NodeCount(); i++ {
		explorer.ForAdjacentEdges(int32(i), FORWARD, ADJACENT_ALL, func(ref EdgeRef) {
			if node_tiles[i] != node_tiles[ref.OtherID] {
				is_border[i] = true
			}
		})
		explorer.ForAdjacentEdges(int32(i), BACKWARD, ADJACENT_ALL, func(ref EdgeRef) {
			if node_tiles[i] != node_tiles[ref.OtherID] {
				is_border[i] = true
			}
		})
	}

	return is_border
}

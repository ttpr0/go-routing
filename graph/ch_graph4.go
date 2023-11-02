package graph

import (
	"errors"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

// // Reorders nodes of graph g inplace.
// // Contraction Hierarchy has to be built with tiles.
// func PrepareGSPHASTIndex(g *CHGraph) error {
// 	if !g._build_with_tiles {
// 		return errors.New("graph has to be build with tiles")
// 	}
// 	// Reorder nodes
// 	order := ComputeTileLevelOrdering(g, g.node_tiles.Value, g.node_levels)
// 	mapping := NewArray[int32](g.NodeCount())
// 	for i := 0; i < g.NodeCount(); i++ {
// 		id := order[i]
// 		new_id := int32(i)
// 		mapping[id] = new_id
// 	}
// 	ReorderCHGraph(g, mapping)
// 	node_tiles := g.node_tiles.Value
// 	is_border := _IsBorderNode2(g, node_tiles)

// 	// initialize down edges lists
// 	fwd_down_edges := NewList[Shortcut](g.NodeCount())
// 	bwd_down_edges := NewList[Shortcut](g.NodeCount())

// 	explorer := g.GetGraphExplorer()
// 	border_count := 0

// 	// add overlay down-edges
// 	fwd_down_edges.Add(_CreateDummyShortcut(-1))
// 	bwd_down_edges.Add(_CreateDummyShortcut(-1))
// 	fwd_other_edges := NewDict[int16, List[Shortcut]](100)
// 	bwd_other_edges := NewDict[int16, List[Shortcut]](100)
// 	for i := 0; i < g.NodeCount(); i++ {
// 		this_id := int32(i)
// 		this_tile := node_tiles[this_id]
// 		if !is_border[this_id] {
// 			border_count = i + 1
// 			break
// 		}
// 		explorer.ForAdjacentEdges(this_id, FORWARD, ADJACENT_DOWNWARDS, func(ref EdgeRef) {
// 			other_id := ref.OtherID
// 			other_tile := node_tiles[other_id]
// 			edge := Shortcut{
// 				From:   this_id,
// 				To:     other_id,
// 				Weight: explorer.GetEdgeWeight(ref),
// 			}
// 			Shortcut_set_payload(&edge, other_tile, 0)
// 			if !is_border[other_id] {
// 				var edges List[Shortcut]
// 				if fwd_other_edges.ContainsKey(this_tile) {
// 					edges = fwd_other_edges[this_tile]
// 				} else {
// 					edges = NewList[Shortcut](10)
// 				}
// 				edges.Add(edge)
// 				fwd_other_edges[this_tile] = edges
// 			} else {
// 				fwd_down_edges.Add(edge)
// 			}
// 		})
// 		explorer.ForAdjacentEdges(this_id, BACKWARD, ADJACENT_DOWNWARDS, func(ref EdgeRef) {
// 			other_id := ref.OtherID
// 			other_tile := node_tiles[other_id]
// 			edge := Shortcut{
// 				From:   this_id,
// 				To:     other_id,
// 				Weight: explorer.GetEdgeWeight(ref),
// 			}
// 			Shortcut_set_payload(&edge, other_tile, 0)
// 			if !is_border[other_id] {
// 				var edges List[Shortcut]
// 				if bwd_other_edges.ContainsKey(this_tile) {
// 					edges = bwd_other_edges[this_tile]
// 				} else {
// 					edges = NewList[Shortcut](10)
// 				}
// 				edges.Add(edge)
// 				bwd_other_edges[this_tile] = edges
// 			} else {
// 				bwd_down_edges.Add(edge)
// 			}
// 		})
// 	}
// 	// add other down edges
// 	curr_tile := int16(-1)
// 	for i := border_count; i < g.NodeCount(); i++ {
// 		this_id := int32(i)
// 		this_tile := node_tiles[this_id]
// 		if this_tile != curr_tile {
// 			fwd_down_edges.Add(_CreateDummyShortcut(this_tile))
// 			bwd_down_edges.Add(_CreateDummyShortcut(this_tile))
// 			curr_tile = this_tile
// 			if fwd_other_edges.ContainsKey(this_tile) {
// 				edges := fwd_other_edges[this_tile]
// 				for _, edge := range edges {
// 					fwd_down_edges.Add(edge)
// 				}
// 			}
// 			if bwd_other_edges.ContainsKey(this_tile) {
// 				edges := bwd_other_edges[this_tile]
// 				for _, edge := range edges {
// 					bwd_down_edges.Add(edge)
// 				}
// 			}
// 		}
// 		explorer.ForAdjacentEdges(this_id, FORWARD, ADJACENT_DOWNWARDS, func(ref EdgeRef) {
// 			other_id := ref.OtherID
// 			fwd_down_edges.Add(Shortcut{
// 				From:   this_id,
// 				To:     other_id,
// 				Weight: explorer.GetEdgeWeight(ref),
// 			})
// 		})
// 		explorer.ForAdjacentEdges(this_id, BACKWARD, ADJACENT_DOWNWARDS, func(ref EdgeRef) {
// 			other_id := ref.OtherID
// 			bwd_down_edges.Add(Shortcut{
// 				From:   this_id,
// 				To:     other_id,
// 				Weight: explorer.GetEdgeWeight(ref),
// 			})
// 		})
// 	}

// 	// set count in dummy edges
// 	fwd_id := 0
// 	fwd_count := 0
// 	for i := 0; i < fwd_down_edges.Length(); i++ {
// 		edge := fwd_down_edges[i]
// 		is_dummy := Shortcut_get_payload[bool](&edge, 2)
// 		if is_dummy {
// 			// set count in previous dummy
// 			fwd_down_edges[fwd_id].To = int32(fwd_count)
// 			// reset count
// 			fwd_id = i
// 			fwd_count = 0
// 			continue
// 		}
// 		fwd_count += 1
// 	}
// 	fwd_down_edges[fwd_id].To = int32(fwd_count)
// 	bwd_id := 0
// 	bwd_count := 0
// 	for i := 0; i < bwd_down_edges.Length(); i++ {
// 		edge := bwd_down_edges[i]
// 		is_dummy := Shortcut_get_payload[bool](&edge, 2)
// 		if is_dummy {
// 			// set count in previous dummy
// 			bwd_down_edges[bwd_id].To = int32(bwd_count)
// 			// reset count
// 			bwd_id = i
// 			bwd_count = 0
// 			continue
// 		}
// 		bwd_count += 1
// 	}
// 	bwd_down_edges[bwd_id].To = int32(bwd_count)

// 	g._contains_dummies = true
// 	g.fwd_down_edges = Some(Array[Shortcut](fwd_down_edges))
// 	g.bwd_down_edges = Some(Array[Shortcut](bwd_down_edges))
// 	return nil
// }

// Modifies ch-data inplace.
func PrepareGSPHASTIndex2(graph *Graph, data ISpeedUpData) error {
	ch_data := data.(*_CHData)
	if !ch_data._build_with_tiles {
		return errors.New("graph has to be build with tiles")
	}

	// Reorder nodes
	order := ComputeTileLevelOrdering(graph, ch_data.node_tiles.Value, ch_data.node_levels)
	mapping := NodeOrderToNodeMapping(order)
	ch_data._ReorderNodes(mapping)
	node_tiles := ch_data.node_tiles.Value
	is_border := _IsBorderNode3(graph, node_tiles)

	temp_graph := CHGraph2{
		base:   graph.base,
		weight: graph.weight,

		id_mapping: ch_data.id_mapping,

		ch_shortcuts: ch_data.shortcuts,
		ch_topology:  ch_data.topology,
		node_levels:  ch_data.node_levels,
		node_tiles:   ch_data.node_tiles,
	}
	explorer := temp_graph.GetGraphExplorer()
	border_count := 0

	// initialize down edges lists
	fwd_down_edges := NewList[Shortcut](temp_graph.NodeCount())
	bwd_down_edges := NewList[Shortcut](temp_graph.NodeCount())

	// add overlay down-edges
	fwd_down_edges.Add(_CreateDummyShortcut(-1))
	bwd_down_edges.Add(_CreateDummyShortcut(-1))
	fwd_other_edges := NewDict[int16, List[Shortcut]](100)
	bwd_other_edges := NewDict[int16, List[Shortcut]](100)
	for i := 0; i < temp_graph.NodeCount(); i++ {
		this_id := int32(i)
		this_tile := node_tiles[this_id]
		if !is_border[this_id] {
			border_count = i + 1
			break
		}
		explorer.ForAdjacentEdges(this_id, FORWARD, ADJACENT_DOWNWARDS, func(ref EdgeRef) {
			other_id := ref.OtherID
			other_tile := node_tiles[other_id]
			edge := Shortcut{
				From:   this_id,
				To:     other_id,
				Weight: explorer.GetEdgeWeight(ref),
			}
			Shortcut_set_payload(&edge, other_tile, 0)
			if !is_border[other_id] {
				var edges List[Shortcut]
				if fwd_other_edges.ContainsKey(this_tile) {
					edges = fwd_other_edges[this_tile]
				} else {
					edges = NewList[Shortcut](10)
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
			edge := Shortcut{
				From:   this_id,
				To:     other_id,
				Weight: explorer.GetEdgeWeight(ref),
			}
			Shortcut_set_payload(&edge, other_tile, 0)
			if !is_border[other_id] {
				var edges List[Shortcut]
				if bwd_other_edges.ContainsKey(this_tile) {
					edges = bwd_other_edges[this_tile]
				} else {
					edges = NewList[Shortcut](10)
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
	for i := border_count; i < temp_graph.NodeCount(); i++ {
		this_id := int32(i)
		this_tile := node_tiles[this_id]
		if this_tile != curr_tile {
			fwd_down_edges.Add(_CreateDummyShortcut(this_tile))
			bwd_down_edges.Add(_CreateDummyShortcut(this_tile))
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
			fwd_down_edges.Add(Shortcut{
				From:   this_id,
				To:     other_id,
				Weight: explorer.GetEdgeWeight(ref),
			})
		})
		explorer.ForAdjacentEdges(this_id, BACKWARD, ADJACENT_DOWNWARDS, func(ref EdgeRef) {
			other_id := ref.OtherID
			bwd_down_edges.Add(Shortcut{
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
		is_dummy := Shortcut_get_payload[bool](&edge, 2)
		if is_dummy {
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
		is_dummy := Shortcut_get_payload[bool](&edge, 2)
		if is_dummy {
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

	ch_data._contains_dummies = true
	ch_data.fwd_down_edges = Some(Array[Shortcut](fwd_down_edges))
	ch_data.bwd_down_edges = Some(Array[Shortcut](bwd_down_edges))
	return nil
}

func _CreateDummyShortcut(to_tile int16) Shortcut {
	dummy := Shortcut{}
	Shortcut_set_payload(&dummy, to_tile, 0)
	Shortcut_set_payload(&dummy, true, 2)
	return dummy
}

func _IsBorderNode2(graph ICHGraph, node_tiles Array[int16]) Array[bool] {
	is_border := NewArray[bool](graph.NodeCount())

	explorer := graph.GetGraphExplorer()
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

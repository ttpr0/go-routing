package graph

import (
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func PreparePHASTIndex(g *CHGraph) {
	fwd_down_edges := NewList[Shortcut](g.NodeCount())
	bwd_down_edges := NewList[Shortcut](g.NodeCount())

	explorer := g.GetGraphExplorer()
	for i := 0; i < g.NodeCount(); i++ {
		this_id := int32(i)
		count := 0
		explorer.ForAdjacentEdges(this_id, FORWARD, ADJACENT_DOWNWARDS, func(ref EdgeRef) {
			other_id := ref.OtherID
			fwd_down_edges.Add(Shortcut{
				From:   this_id,
				To:     other_id,
				Weight: explorer.GetEdgeWeight(ref),
			})
			count += 1
		})
		for j := fwd_down_edges.Length() - count; j < fwd_down_edges.Length(); j++ {
			ch_edge := fwd_down_edges[j]
			Shortcut_set_payload(&ch_edge, int32(count), 0)
			fwd_down_edges[j] = ch_edge
		}

		count = 0
		explorer.ForAdjacentEdges(this_id, BACKWARD, ADJACENT_DOWNWARDS, func(ref EdgeRef) {
			other_id := ref.OtherID
			bwd_down_edges.Add(Shortcut{
				From:   this_id,
				To:     other_id,
				Weight: explorer.GetEdgeWeight(ref),
			})
			count += 1
		})
		for j := bwd_down_edges.Length() - count; j < bwd_down_edges.Length(); j++ {
			ch_edge := bwd_down_edges[j]
			Shortcut_set_payload(&ch_edge, int32(count), 0)
			bwd_down_edges[j] = ch_edge
		}
	}

	g.fwd_down_edges = Some(Array[Shortcut](fwd_down_edges))
	g.bwd_down_edges = Some(Array[Shortcut](bwd_down_edges))
}

// Modifies ch-data inplace.
func PreparePHASTIndex2(graph *Graph, ch_data *_CHData) {
	order := ComputeLevelOrdering(graph, ch_data.node_levels)
	mapping := NodeOrderToNodeMapping(order)
	ch_data._ReorderNodes(mapping)

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

	fwd_down_edges := NewList[Shortcut](temp_graph.NodeCount())
	bwd_down_edges := NewList[Shortcut](temp_graph.NodeCount())

	for i := 0; i < temp_graph.NodeCount(); i++ {
		this_id := int32(i)
		count := 0
		explorer.ForAdjacentEdges(this_id, FORWARD, ADJACENT_DOWNWARDS, func(ref EdgeRef) {
			other_id := ref.OtherID
			fwd_down_edges.Add(Shortcut{
				From:   this_id,
				To:     other_id,
				Weight: explorer.GetEdgeWeight(ref),
			})
			count += 1
		})
		for j := fwd_down_edges.Length() - count; j < fwd_down_edges.Length(); j++ {
			ch_edge := fwd_down_edges[j]
			Shortcut_set_payload(&ch_edge, int32(count), 0)
			fwd_down_edges[j] = ch_edge
		}

		count = 0
		explorer.ForAdjacentEdges(this_id, BACKWARD, ADJACENT_DOWNWARDS, func(ref EdgeRef) {
			other_id := ref.OtherID
			bwd_down_edges.Add(Shortcut{
				From:   this_id,
				To:     other_id,
				Weight: explorer.GetEdgeWeight(ref),
			})
			count += 1
		})
		for j := bwd_down_edges.Length() - count; j < bwd_down_edges.Length(); j++ {
			ch_edge := bwd_down_edges[j]
			Shortcut_set_payload(&ch_edge, int32(count), 0)
			bwd_down_edges[j] = ch_edge
		}
	}

	ch_data.fwd_down_edges = Some(Array[Shortcut](fwd_down_edges))
	ch_data.bwd_down_edges = Some(Array[Shortcut](bwd_down_edges))
}

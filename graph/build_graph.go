package graph

import (
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// build graphs
//*******************************************

func BuildBaseGraph(base *GraphBase, weight IWeighting) *Graph {
	return &Graph{
		base:   *base,
		weight: weight,
	}
}

func BuildCHGraph(base *GraphBase, weight IWeighting, ch_data ISpeedUpData) *CHGraph {
	data := ch_data.(*_CHData)

	return &CHGraph{
		base:   *base,
		weight: weight,

		id_mapping: data.id_mapping,

		_build_with_tiles: data._build_with_tiles,

		ch_shortcuts: data.shortcuts,
		ch_topology:  data.topology,
		node_levels:  data.node_levels,
	}
}

func BuildCHGraph2(base *GraphBase, weight IWeighting, ch_data ISpeedUpData) *CHGraph2 {
	data := ch_data.(*_CHData)

	return &CHGraph2{
		base:   *base,
		weight: weight,

		id_mapping: data.id_mapping,

		_build_with_tiles: data._build_with_tiles,

		ch_shortcuts: data.shortcuts,
		ch_topology:  data.topology,
		node_levels:  data.node_levels,

		node_tiles: data.node_tiles,

		fwd_down_edges: data.fwd_down_edges,
		bwd_down_edges: data.bwd_down_edges,
	}
}

func BuildTiledGraph(base *GraphBase, weight IWeighting, tiled_data ISpeedUpData) *TiledGraph {
	data := tiled_data.(*_TiledData)

	return &TiledGraph{
		base:   *base,
		weight: weight,

		skip_shortcuts: data.skip_shortcuts,
		skip_topology:  data.skip_topology,
		node_tiles:     data.node_tiles,
		edge_types:     data.edge_types,
		cell_index:     data.cell_index,
	}
}

func BuildTiledGraph2(base *GraphBase, weight IWeighting, tiled_data ISpeedUpData) *TiledGraph2 {
	data := tiled_data.(*_TiledData)

	return &TiledGraph2{
		base:   *base,
		weight: weight,

		id_mapping: data.id_mapping,

		skip_shortcuts: data.skip_shortcuts,
		skip_topology:  data.skip_topology,
		node_tiles:     data.node_tiles,
		edge_types:     data.edge_types,
		cell_index:     data.cell_index,
	}
}

//*******************************************
// build graph components
//*******************************************

func _BuildTopology(store GraphStore) AdjacencyArray {
	nodes := store.nodes
	edges := store.edges

	dyn := NewAdjacencyList(nodes.Length())
	for id, edge := range edges {
		dyn.AddFWDEntry(edge.NodeA, edge.NodeB, int32(id), 0)
		dyn.AddBWDEntry(edge.NodeA, edge.NodeB, int32(id), 0)
	}

	return *AdjacencyListToArray(&dyn)
}

func _BuildKDTreeIndex(store GraphStore) KDTree[int32] {
	node_geoms := store.node_geoms

	tree := NewKDTree[int32](2)
	for i := 0; i < len(node_geoms); i++ {
		geom := node_geoms[i]
		tree.Insert(geom[:], int32(i))
	}
	return tree
}

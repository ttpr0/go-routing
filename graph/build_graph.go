package graph

import (
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// build graphs
//*******************************************

func BuildGraphBase(nodes Array[Node], edges Array[Edge]) *GraphBase {
	topology := _BuildTopology(nodes, edges)
	return &GraphBase{
		nodes:    nodes,
		edges:    edges,
		topology: topology,
	}
}

func BuildGraph(base IGraphBase, weight IWeighting) *Graph {
	return &Graph{
		base:   base,
		weight: weight,
	}
}

func BuildCHGraph(base IGraphBase, weight IWeighting, ch_data *CH, ch_index Optional[*CHIndex]) *CHGraph {
	return &CHGraph{
		base:   base,
		weight: weight,

		ch_shortcuts: ch_data.shortcuts,
		ch_topology:  ch_data.topology,
		node_levels:  ch_data.node_levels,

		partition: None[*Partition](),

		ch_index: ch_index,
	}
}

func BuildPartitionedCHGraph(base IGraphBase, weight IWeighting, ch_data *CH, partition Optional[*Partition], ch_index Optional[*CHIndex]) *CHGraph {
	return &CHGraph{
		base:   base,
		weight: weight,

		ch_shortcuts: ch_data.shortcuts,
		ch_topology:  ch_data.topology,
		node_levels:  ch_data.node_levels,

		partition: partition,

		ch_index: ch_index,
	}
}

func BuildTiledGraph(base IGraphBase, weight IWeighting, partition *Partition, overlay *Overlay, cell_index Optional[*CellIndex]) *TiledGraph {
	return &TiledGraph{
		base:   base,
		weight: weight,

		partition:      partition,
		skip_shortcuts: overlay.skip_shortcuts,
		skip_topology:  overlay.skip_topology,
		edge_types:     overlay.edge_types,
		cell_index:     cell_index,
	}
}

func BuildGraphIndex(base IGraphBase) IGraphIndex {
	index := _BuildKDTreeIndex(base)
	return &BaseGraphIndex{
		index: index,
	}
}

//*******************************************
// build graph components
//*******************************************

func _BuildTopology(nodes Array[Node], edges Array[Edge]) _AdjacencyArray {
	dyn := _NewAdjacencyList(nodes.Length())
	for id, edge := range edges {
		dyn.AddFWDEntry(edge.NodeA, edge.NodeB, int32(id), 0)
		dyn.AddBWDEntry(edge.NodeA, edge.NodeB, int32(id), 0)
	}

	return *_AdjacencyListToArray(&dyn)
}

func _BuildKDTreeIndex(base IGraphBase) KDTree[int32] {
	tree := NewKDTree[int32](2)
	for i := 0; i < base.NodeCount(); i++ {
		if !base.IsNode(int32(i)) {
			continue
		}
		node := base.GetNode(int32(i))
		geom := node.Loc
		tree.Insert(geom[:], int32(i))
	}
	return tree
}

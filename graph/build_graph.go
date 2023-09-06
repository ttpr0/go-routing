package graph

import (
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// utility methods
//*******************************************

func _BuildTopology(store GraphStore) TopologyStore {
	nodes := store.nodes
	edges := store.edges

	dyn := NewDynamicTopology(nodes.Length())
	for id, edge := range edges {
		dyn.AddFWDEntry(edge.NodeA, edge.NodeB, int32(id), 0)
		dyn.AddBWDEntry(edge.NodeA, edge.NodeB, int32(id), 0)
	}

	return *DynamicToTopology(&dyn)
}

func _BuildWeighting(store GraphStore) DefaultWeighting {
	edges := store.edges

	weights := NewArray[int32](edges.Length())
	for id, edge := range edges {
		w := edge.Length * 3.6 / float32(edge.Maxspeed)
		if w < 1 {
			w = 1
		}
		weights[id] = int32(w)
	}

	return DefaultWeighting{
		edge_weights: weights,
	}
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

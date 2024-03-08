package comps

import (
	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

// Reorders Array based on mapping (old-id -> new-id).
func Reorder[T any](arr Array[T], mapping Array[int32]) Array[T] {
	new_arr := NewArray[T](arr.Length())
	for i, id := range mapping {
		new_arr[id] = arr[i]
	}
	return new_arr
}

//*******************************************
// build graph components
//*******************************************

func _BuildTopology(nodes Array[structs.Node], edges Array[structs.Edge]) structs.AdjacencyArray {
	dyn := structs.NewAdjacencyList(nodes.Length())
	for id, edge := range edges {
		dyn.AddFWDEntry(edge.NodeA, edge.NodeB, int32(id), 0)
		dyn.AddBWDEntry(edge.NodeA, edge.NodeB, int32(id), 0)
	}

	return *structs.AdjacencyListToArray(&dyn)
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

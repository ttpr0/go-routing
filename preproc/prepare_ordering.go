package preproc

import (
	"sort"

	"github.com/ttpr0/go-routing/comps"
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// compute orderings
//*******************************************

// Orders nodes by CH-level.
func ComputeLevelOrdering(g graph.IGraph, ch *comps.CH) Array[int32] {
	indices := NewList[Tuple[int32, int16]](int(g.NodeCount()))
	for i := 0; i < int(g.NodeCount()); i++ {
		indices.Add(MakeTuple(int32(i), ch.GetNodeLevel(int32(i))))
	}
	sort.SliceStable(indices, func(i, j int) bool {
		return indices[i].B > indices[j].B
	})
	order := NewArray[int32](len(indices))
	for i, index := range indices {
		order[i] = index.A
	}
	return _NodeOrderToNodeMapping(order)
}

// Orders nodes by tiles and levels.
// Border nodes are pushed to front of all nodes.
// Within their tiles nodes are ordered by level.
func ComputeTileLevelOrdering(g graph.IGraph, partition *comps.Partition, ch *comps.CH) Array[int32] {
	// sort by level
	indices := NewList[Tuple[int32, int16]](int(g.NodeCount()))
	for i := 0; i < int(g.NodeCount()); i++ {
		indices.Add(MakeTuple(int32(i), ch.GetNodeLevel(int32(i))))
	}
	sort.SliceStable(indices, func(i, j int) bool {
		return indices[i].B > indices[j].B
	})
	// sort by tile
	is_border := _IsBorderNode3(g, partition)
	for i := 0; i < int(g.NodeCount()); i++ {
		index := indices[i]
		tile := partition.GetNodeTile(index.A)
		if is_border[index.A] {
			tile = -10000
		}
		index.B = tile
		indices[i] = index
	}
	sort.SliceStable(indices, func(i, j int) bool {
		return indices[i].B < indices[j].B
	})
	order := NewArray[int32](len(indices))
	for i, index := range indices {
		order[i] = index.A
	}
	return _NodeOrderToNodeMapping(order)
}

// Orders nodes by tiles.
// Border nodes are pushed to front of all nodes.
func ComputeTileOrdering(g graph.IGraph, partition *comps.Partition) Array[int32] {
	is_border := _IsBorderNode3(g, partition)
	indices := NewList[Tuple[int32, int16]](int(g.NodeCount()))
	for i := 0; i < int(g.NodeCount()); i++ {
		tile := partition.GetNodeTile(int32(i))
		if is_border[i] {
			tile = -10000
		}
		indices.Add(MakeTuple(int32(i), tile))
	}
	sort.SliceStable(indices, func(i, j int) bool {
		return indices[i].B < indices[j].B
	})
	order := NewArray[int32](len(indices))
	for i, index := range indices {
		order[i] = index.A
	}
	return _NodeOrderToNodeMapping(order)
}

// Convert node ordering to node mapping (for graph.ReorderNodes functions).
// order contains id's of nodes in their new order.
func _NodeOrderToNodeMapping(order Array[int32]) Array[int32] {
	mapping := NewArray[int32](len(order))
	for new_id, id := range order {
		mapping[int(id)] = int32(new_id)
	}
	return mapping
}

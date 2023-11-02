package graph

import (
	"sort"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// modification methods
//*******************************************

// Removes all weightings and speed-ups.
func RemoveBaseNodes(base *GraphBase, nodes List[int32]) {
	base._RemoveNodes(nodes)
}

//*******************************************
// reordering methods
//*******************************************

// reorders node information of base-graph,
// mapping: old id -> new id
func ReorderGraphBaseNodes(base *GraphBase, mapping Array[int32]) {
	base._ReorderNodes(mapping)
}

// reorders node information of base-graph,
// mapping: old id -> new id
func ReorderWeightingNodes(weight IWeighting, mapping Array[int32]) {
	w_h := WEIGHTING_HANDLERS[weight.Type()]
	w_h._ReorderNodes(weight, mapping)
}

type ReorderType byte

const (
	ALL_NODES         ReorderType = 0
	ONLY_SOURCE_NODES ReorderType = 2
	ONLY_TARGET_NODES ReorderType = 4
)

// reorders node information of base-graph,
// mapping: old id -> new id
func ReorderSpeedUpNodes(speed ISpeedUpData, mapping Array[int32], typ ReorderType) {
	s_h := SPEEDUP_HANDLERS[speed.Type()]
	s_h._ReorderNodesInplace(speed, mapping, typ)
}

// reorders node information of base-graph,
// mapping: old id -> new id
func ReorderStoredSpeedUpNodes(dir, name string, typ1 SpeedUpType, mapping Array[int32], typ2 ReorderType) {
	s_h := SPEEDUP_HANDLERS[typ1]
	s_h._ReorderNodes(dir, name, mapping, typ2)
}

// func SortNodesByLevel(g *CHGraph) {
// 	indices := NewList[Tuple[int32, int16]](int(g.NodeCount()))
// 	for i := 0; i < int(g.NodeCount()); i++ {
// 		indices.Add(MakeTuple(int32(i), g.GetNodeLevel(int32(i))))
// 	}
// 	sort.SliceStable(indices, func(i, j int) bool {
// 		return indices[i].B > indices[j].B
// 	})
// 	order := NewArray[int32](len(indices))
// 	for i, index := range indices {
// 		order[i] = index.A
// 	}

// 	mapping := NewArray[int32](len(order))
// 	for new_id, id := range order {
// 		mapping[int(id)] = int32(new_id)
// 	}

// 	ReorderCHGraph(g, mapping)
// }

// // Reorders nodes in graph inplace.
// // "node_mapping" maps old id -> new id.
// func ReorderGraph(g *Graph, node_mapping Array[int32]) {
// 	g.base._ReorderNodes(node_mapping)
// 	panic("dont use this")
// 	// g.weight._ReorderNodes(node_mapping)
// }

// // Reorders nodes in graph inplace.
// // "node_mapping" maps old id -> new id.
// func ReorderCHGraph(g *CHGraph, node_mapping Array[int32]) {
// 	g.base._ReorderNodes(node_mapping)
// 	panic("dont use this")
// 	// g.weight._ReorderNodes(node_mapping)

// 	g.ch_shortcuts._ReorderNodes(node_mapping)
// 	g.ch_topology._ReorderNodes(node_mapping)
// 	Reorder[int16](g.node_levels, node_mapping)

// 	if g._build_with_tiles {
// 		Reorder[int16](g.node_tiles.Value, node_mapping)
// 	}

// 	if g.HasDownEdges(FORWARD) || g.HasDownEdges(BACKWARD) {
// 		panic("not implemented")
// 	}
// }

// // Reorders nodes in graph inplace.
// // "node_mapping" maps old id -> new id.
// func ReorderTiledGraph(g *TiledGraph, node_mapping Array[int32]) {
// 	g.base._ReorderNodes(node_mapping)
// 	panic("dont use this")
// 	// g.weight._ReorderNodes(node_mapping)

// 	g.skip_shortcuts._ReorderNodes(node_mapping)
// 	g.skip_topology._ReorderNodes(node_mapping)
// 	Reorder[int16](g.node_tiles, node_mapping)
// 	if g.HasCellIndex() {
// 		g.cell_index.Value._ReorderNodes(node_mapping)
// 	}
// }

//*******************************************
// compute orderings
//*******************************************

// Orders nodes by CH-level.
func ComputeLevelOrdering(g IGraph, node_levels Array[int16]) Array[int32] {
	indices := NewList[Tuple[int32, int16]](int(g.NodeCount()))
	for i := 0; i < int(g.NodeCount()); i++ {
		indices.Add(MakeTuple(int32(i), node_levels[i]))
	}
	sort.SliceStable(indices, func(i, j int) bool {
		return indices[i].B > indices[j].B
	})
	order := NewArray[int32](len(indices))
	for i, index := range indices {
		order[i] = index.A
	}
	return order
}

// Orders nodes by tiles and levels.
// Border nodes are pushed to front of all nodes.
// Within their tiles nodes are ordered by level.
func ComputeTileLevelOrdering(g IGraph, node_tiles Array[int16], node_levels Array[int16]) Array[int32] {
	// sort by level
	indices := NewList[Tuple[int32, int16]](int(g.NodeCount()))
	for i := 0; i < int(g.NodeCount()); i++ {
		indices.Add(MakeTuple(int32(i), node_levels[i]))
	}
	sort.SliceStable(indices, func(i, j int) bool {
		return indices[i].B > indices[j].B
	})
	// sort by tile
	is_border := _IsBorderNode3(g, node_tiles)
	for i := 0; i < int(g.NodeCount()); i++ {
		index := indices[i]
		tile := node_tiles[index.A]
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
	return order
}

// Orders nodes by tiles.
// Border nodes are pushed to front of all nodes.
func ComputeTileOrdering(g IGraph, node_tiles Array[int16]) Array[int32] {
	is_border := _IsBorderNode3(g, node_tiles)
	indices := NewList[Tuple[int32, int16]](int(g.NodeCount()))
	for i := 0; i < int(g.NodeCount()); i++ {
		tile := node_tiles[i]
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
	return order
}

// Convert node ordering to node mapping (for ReorderXXGraph functions).
// order contains id's of nodes in their new order.
func NodeOrderToNodeMapping(order Array[int32]) Array[int32] {
	mapping := NewArray[int32](len(order))
	for new_id, id := range order {
		mapping[int(id)] = int32(new_id)
	}
	return mapping
}

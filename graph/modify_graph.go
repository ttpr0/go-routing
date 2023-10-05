package graph

import (
	"sort"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// modifikation methods
//*******************************************

func RemoveNodes(graph *Graph, nodes List[int32]) *Graph {
	store := graph.store

	remove := NewArray[bool](store.NodeCount())
	for _, n := range nodes {
		remove[n] = true
	}

	new_nodes := NewList[Node](100)
	new_node_geoms := NewList[geo.Coord](100)
	mapping := NewArray[int32](store.NodeCount())
	id := int32(0)
	for i := 0; i < store.NodeCount(); i++ {
		if remove[i] {
			mapping[i] = -1
			continue
		}
		new_nodes.Add(store.GetNode(int32(i)))
		new_node_geoms.Add(store.GetNodeGeom(int32(i)))
		mapping[i] = id
		id += 1
	}
	new_edges := NewList[Edge](100)
	new_edge_geoms := NewList[geo.CoordArray](100)
	for i := 0; i < store.EdgeCount(); i++ {
		edge := store.GetEdge(int32(i))
		if remove[edge.NodeA] || remove[edge.NodeB] {
			continue
		}
		new_edges.Add(Edge{
			NodeA:    mapping[edge.NodeA],
			NodeB:    mapping[edge.NodeB],
			Type:     edge.Type,
			Length:   edge.Length,
			Maxspeed: edge.Maxspeed,
			Oneway:   edge.Oneway,
		})
		new_edge_geoms.Add(store.GetEdgeGeom(int32(i)))
	}

	new_store := GraphStore{
		nodes:      Array[Node](new_nodes),
		edges:      Array[Edge](new_edges),
		node_geoms: new_node_geoms,
		edge_geoms: new_edge_geoms,
	}
	return &Graph{
		store:    new_store,
		topology: _BuildTopology(new_store),
		weight:   _BuildWeighting(new_store),
		index:    _BuildKDTreeIndex(new_store),
	}
}

//*******************************************
// reordering methods
//*******************************************

func SortNodesByLevel(g *CHGraph) {
	indices := NewList[Tuple[int32, int16]](int(g.NodeCount()))
	for i := 0; i < int(g.NodeCount()); i++ {
		indices.Add(MakeTuple(int32(i), g.GetNodeLevel(int32(i))))
	}
	sort.SliceStable(indices, func(i, j int) bool {
		return indices[i].B > indices[j].B
	})
	order := NewArray[int32](len(indices))
	for i, index := range indices {
		order[i] = index.A
	}

	mapping := NewArray[int32](len(order))
	for new_id, id := range order {
		mapping[int(id)] = int32(new_id)
	}

	ReorderCHGraph(g, mapping)
}

// Reorders nodes in graph inplace.
// "node_mapping" maps old id -> new id.
func ReorderGraph(g *Graph, node_mapping Array[int32]) {
	g.store._ReorderNodes(node_mapping)
	g.topology._ReorderNodes(node_mapping)
	g.weight._ReorderNodes(node_mapping)
	g.index = _BuildKDTreeIndex(g.store)
}

// Reorders nodes in graph inplace.
// "node_mapping" maps old id -> new id.
func ReorderCHGraph(g *CHGraph, node_mapping Array[int32]) {
	g.store._ReorderNodes(node_mapping)
	g.topology._ReorderNodes(node_mapping)
	g.weight._ReorderNodes(node_mapping)
	g.index = _BuildKDTreeIndex(g.store)

	g.ch_store._ReorderNodes(node_mapping)
	g.ch_topology._ReorderNodes(node_mapping)
}

// Reorders nodes in graph inplace.
// "node_mapping" maps old id -> new id.
func ReorderTiledGraph(g *TiledGraph, node_mapping Array[int32]) {
	g.store._ReorderNodes(node_mapping)
	g.topology._ReorderNodes(node_mapping)
	g.weight._ReorderNodes(node_mapping)
	g.index = _BuildKDTreeIndex(g.store)

	g.skip_store._ReorderNodes(node_mapping)
	g.skip_topology._ReorderNodes(node_mapping)
}

// Orders nodes by CH-level.
func ComputeLevelOrdering(g ICHGraph) Array[int32] {
	indices := NewList[Tuple[int32, int16]](int(g.NodeCount()))
	for i := 0; i < int(g.NodeCount()); i++ {
		indices.Add(MakeTuple(int32(i), g.GetNodeLevel(int32(i))))
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
func ComputeTileLevelOrdering(g ICHGraph, node_tiles Array[int16]) Array[int32] {
	// sort by level
	indices := NewList[Tuple[int32, int16]](int(g.NodeCount()))
	for i := 0; i < int(g.NodeCount()); i++ {
		indices.Add(MakeTuple(int32(i), g.GetNodeLevel(int32(i))))
	}
	sort.SliceStable(indices, func(i, j int) bool {
		return indices[i].B > indices[j].B
	})
	// sort by tile
	is_border := _IsBorderNode2(g, node_tiles)
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

// Orders nodes by tiles and levels.
// Border nodes are pushed to front of all nodes.
// Within their tiles nodes are ordered by level.
func ComputeTileLevelOrdering2(g ITiledGraph, node_levels Array[int16]) Array[int32] {
	// sort by level
	indices := NewList[Tuple[int32, int16]](int(g.NodeCount()))
	for i := 0; i < int(g.NodeCount()); i++ {
		indices.Add(MakeTuple(int32(i), node_levels[i]))
	}
	sort.SliceStable(indices, func(i, j int) bool {
		return indices[i].B > indices[j].B
	})
	// sort by tile
	is_border := _IsBorderNode3(g)
	for i := 0; i < int(g.NodeCount()); i++ {
		index := indices[i]
		tile := g.GetNodeTile(index.A)
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
func ComputeTileOrdering(g ITiledGraph) Array[int32] {
	is_border := _IsBorderNode3(g)
	indices := NewList[Tuple[int32, int16]](int(g.NodeCount()))
	for i := 0; i < int(g.NodeCount()); i++ {
		tile := g.GetNodeTile(int32(i))
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

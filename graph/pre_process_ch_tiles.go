package graph

import (
	"fmt"
	"sort"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// preprocess tiled graph 4
//*******************************************

// func PreprocessTiledGraph4(graph *Graph, node_tiles Array[int16]) *TiledGraph {
// 	fmt.Println("Compute node ordering:")
// 	// contraction_order, border_nodes := TiledNodeOrdering(graph)

// 	fmt.Println("Compute subset contraction:")
// 	dg := TransformToCHPreprocGraph(graph)
// 	CalcPartialContraction5(dg, node_tiles)
// 	// CalcContraction2(dg, contraction_order)

// 	fmt.Println("Set border nodes to maxlevel:")
// 	border_nodes := _IsBorderNode3(graph, node_tiles)
// 	max_level := int16(0)
// 	node_levels := dg.node_levels
// 	for i := 0; i < node_levels.Length(); i++ {
// 		if node_levels[i] > max_level {
// 			max_level = node_levels[i]
// 		}
// 	}
// 	for i := 0; i < node_levels.Length(); i++ {
// 		if border_nodes[i] {
// 			node_levels[i] = max_level + 1
// 		}
// 	}

// 	fmt.Println("Create topology from shortcuts:")
// 	edge_types := NewArray[byte](graph.EdgeCount())
// 	_UpdateCrossBorder(graph, node_tiles, edge_types)
// 	skip_topology, skip_shortcuts := CreateCHSkipTopology(dg, border_nodes, node_tiles)
// 	new_graph := &TiledGraph{
// 		store:    graph.store,
// 		topology: graph.topology,
// 		weight:   graph.weight,
// 		index:    graph.index,

// 		skip_shortcuts: skip_shortcuts,
// 		skip_topology:  *skip_topology,
// 		node_tiles:     node_tiles,
// 		edge_types:     edge_types,
// 	}

// 	// reorder graph
// 	order := ComputeTileLevelOrdering(graph, node_tiles, node_levels)
// 	mapping := NewArray[int32](graph.NodeCount())
// 	for i := 0; i < graph.NodeCount(); i++ {
// 		id := order[i]
// 		new_id := int32(i)
// 		mapping[id] = new_id
// 	}
// 	ReorderTiledGraph(new_graph, mapping)
// 	node_levels = Reorder[int16](node_levels, mapping)
// 	node_tiles = new_graph.node_tiles

// 	// remap shortcuts
// 	dg.shortcuts._ReorderNodes(mapping)
// 	ch_shortcuts := dg.shortcuts

// 	fmt.Println("Create downwards edge lists:")
// 	edges := new_graph.store.edges
// 	edge_weigths := new_graph.weight.edge_weights
// 	tiles := _GetTiles(new_graph.node_tiles)
// 	cell_index := _NewCellIndex()
// 	for index, tile := range tiles {
// 		fmt.Println("Process Tile:", index+1, "/", len(tiles))
// 		// get all down edges or shortcuts
// 		edge_list := NewList[Shortcut](100)
// 		for i := 0; i < ch_shortcuts.ShortcutCount(); i++ {
// 			shc := ch_shortcuts.GetShortcut(int32(i))
// 			if node_tiles[shc.From] != tile || node_tiles[shc.To] != tile {
// 				continue
// 			}
// 			if node_levels[shc.From] > node_levels[shc.To] {
// 				edge_list.Add(Shortcut{
// 					From:   shc.From,
// 					To:     shc.From,
// 					Weight: shc.Weight,
// 				})
// 			}
// 		}
// 		for i := 0; i < edges.Length(); i++ {
// 			edge := edges[i]
// 			if node_tiles[edge.NodeA] != tile || node_tiles[edge.NodeB] != tile {
// 				continue
// 			}
// 			if node_levels[edge.NodeA] > node_levels[edge.NodeB] {
// 				edge_list.Add(Shortcut{
// 					From:   edge.NodeA,
// 					To:     edge.NodeB,
// 					Weight: edge_weigths[i],
// 				})
// 			}
// 		}

// 		// sort down edges by node level
// 		sort.Slice(edge_list, func(i, j int) bool {
// 			e_i := edge_list[i]
// 			level_i := node_levels[e_i.From]
// 			e_j := edge_list[j]
// 			level_j := node_levels[e_j.From]
// 			return level_i > level_j
// 		})

// 		// add edges to index_edges
// 		cell_index.SetFWDIndexEdges(tile, Array[Shortcut](edge_list))
// 	}
// 	new_graph.cell_index = Some(cell_index)

// 	return new_graph
// }

func PreprocessTiledGraph5(graph *Graph, node_tiles Array[int16]) *_TiledData {
	fmt.Println("Compute node ordering:")
	// contraction_order, border_nodes := TiledNodeOrdering(graph)

	fmt.Println("Compute subset contraction:")
	ch_data := CalcPartialContraction5(graph, node_tiles)
	// CalcContraction2(dg, contraction_order)

	fmt.Println("Set border nodes to maxlevel:")
	border_nodes := _IsBorderNode3(graph, node_tiles)
	max_level := int16(0)
	node_levels := ch_data.node_levels
	for i := 0; i < node_levels.Length(); i++ {
		if node_levels[i] > max_level {
			max_level = node_levels[i]
		}
	}
	for i := 0; i < node_levels.Length(); i++ {
		if border_nodes[i] {
			node_levels[i] = max_level + 1
		}
	}

	fmt.Println("Create topology from shortcuts:")
	edge_types := NewArray[byte](graph.EdgeCount())
	_UpdateCrossBorder(graph, node_tiles, edge_types)
	skip_topology, skip_shortcuts := CreateCHSkipTopology2(graph, ch_data, border_nodes, node_tiles)
	tiled_data := &_TiledData{
		skip_shortcuts: skip_shortcuts,
		skip_topology:  *skip_topology,
		node_tiles:     node_tiles,
		edge_types:     edge_types,

		_base_weighting: graph._weight_name,
	}

	// reorder graph
	order := ComputeTileLevelOrdering(graph, node_tiles, node_levels)
	mapping := NodeOrderToNodeMapping(order)
	tiled_data._ReorderNodes(mapping)
	node_tiles = tiled_data.node_tiles

	// remap shortcuts
	ch_data._ReorderNodes(mapping)
	ch_shortcuts := ch_data.shortcuts
	node_levels = ch_data.node_levels

	// create temporary graph
	temp_graph := &TiledGraph2{
		base:   graph.base,
		weight: graph.weight,

		id_mapping: tiled_data.id_mapping,

		skip_shortcuts: tiled_data.skip_shortcuts,
		skip_topology:  tiled_data.skip_topology,
		node_tiles:     tiled_data.node_tiles,
		edge_types:     tiled_data.edge_types,
		cell_index:     tiled_data.cell_index,
	}

	fmt.Println("Create downwards edge lists:")
	// edges := temp_graph.store.edges
	// edge_weigths := temp_graph.weight.edge_weights
	explorer := temp_graph.GetGraphExplorer()
	tiles := _GetTiles(temp_graph.node_tiles)
	cell_index := _NewCellIndex()
	for index, tile := range tiles {
		fmt.Println("Process Tile:", index+1, "/", len(tiles))
		// get all down edges or shortcuts
		edge_list := NewList[Shortcut](100)
		for i := 0; i < ch_shortcuts.ShortcutCount(); i++ {
			shc := ch_shortcuts.GetShortcut(int32(i))
			if node_tiles[shc.From] != tile || node_tiles[shc.To] != tile {
				continue
			}
			if node_levels[shc.From] > node_levels[shc.To] {
				edge_list.Add(Shortcut{
					From:   shc.From,
					To:     shc.From,
					Weight: shc.Weight,
				})
			}
		}
		for i := 0; i < temp_graph.EdgeCount(); i++ {
			edge := temp_graph.GetEdge(int32(i))
			if node_tiles[edge.NodeA] != tile || node_tiles[edge.NodeB] != tile {
				continue
			}
			if node_levels[edge.NodeA] > node_levels[edge.NodeB] {
				edge_list.Add(Shortcut{
					From:   edge.NodeA,
					To:     edge.NodeB,
					Weight: explorer.GetEdgeWeight(CreateEdgeRef(int32(i))),
				})
			}
		}

		// sort down edges by node level
		sort.Slice(edge_list, func(i, j int) bool {
			e_i := edge_list[i]
			level_i := node_levels[e_i.From]
			e_j := edge_list[j]
			level_j := node_levels[e_j.From]
			return level_i > level_j
		})

		// add edges to index_edges
		cell_index.SetFWDIndexEdges(tile, Array[Shortcut](edge_list))
	}

	tiled_data.cell_index = Some(cell_index)
	return tiled_data
}

//*******************************************
// compute node ordering
//*******************************************

func TiledNodeOrdering(graph *TiledGraph) (Array[int32], Array[bool]) {
	tiles := _GetTiles(graph.node_tiles)
	border_nodes := NewArray[bool](graph.NodeCount())
	sp_counts := NewArray[int](graph.NodeCount())
	explorer := graph.GetGraphExplorer()

	// compute border_nodes and sp_counts for every tile
	fmt.Println("start computing tiles...")
	for index, tile := range tiles {
		fmt.Println("    - process tile:", index+1, "/", len(tiles))
		b_nodes, i_nodes := GetBorderNodes2(graph, tile)
		flags := NewDict[int32, _Flag](100)
		for _, b_node := range b_nodes {
			border_nodes[b_node] = true
			flags.Clear()
			_CalcFullSPT(graph, b_node, flags)
			for _, i_node := range i_nodes {
				if !flags.ContainsKey(i_node) {
					continue
				}
				curr_id := i_node
				var edge int32
				for {
					sp_counts[curr_id] += 1
					if curr_id == b_node {
						break
					}
					edge = flags[curr_id].prevEdge
					curr_id = explorer.GetOtherNode(CreateEdgeRef(edge), curr_id)
				}
			}
		}
	}

	// remove border nodes from order
	fmt.Println("start creating interior-nodes list...")
	nodes := NewList[int32](int(graph.NodeCount()))
	for i := 0; i < int(graph.NodeCount()); i++ {
		if border_nodes[i] {
			continue
		}
		nodes.Add(int32(i))
	}

	// sort nodes by number of shortest path they are on
	fmt.Println("start ordering nodes...")
	sort.Slice(nodes, func(i, j int) bool {
		a := nodes[i]
		count_a := sp_counts[a]
		b := nodes[j]
		count_b := sp_counts[b]
		return count_a < count_b
	})

	fmt.Println("finished!")

	return Array[int32](nodes), border_nodes
}

//*******************************************
// create topology store
//*******************************************

// creates topology with cross-border edges (type 10), skip-edges (type 20) and shortcuts (type 100)
func CreateCHSkipTopology(dg *CHPreprocGraph, border_nodes Array[bool], node_tiles Array[int16]) (*AdjacencyArray, ShortcutStore) {
	dyn_top := NewAdjacencyList(dg.NodeCount())
	shortcuts := NewShortcutStore(100, true)

	explorer := dg.GetExplorer()

	for i := 0; i < dg.NodeCount(); i++ {
		if !border_nodes[i] {
			continue
		}
		explorer.ForAdjacentEdges(int32(i), FORWARD, ADJACENT_ALL, func(ref EdgeRef) {
			if !border_nodes[ref.OtherID] {
				return
			}
			if ref.IsShortcut() {
				shc := dg.GetShortcut(ref.EdgeID)
				shc_n := NewShortcut(shc.From, shc.To, explorer.GetEdgeWeight(ref))
				edges_n := [2]Tuple[int32, byte]{}
				shc_id, _ := shortcuts.AddCHShortcut(shc_n, edges_n)
				dyn_top.AddEdgeEntries(shc.From, shc.To, shc_id, 100)
			} else {
				edge := dg.GetEdge(ref.EdgeID)
				if node_tiles[edge.NodeA] != node_tiles[edge.NodeB] {
					dyn_top.AddEdgeEntries(edge.NodeA, edge.NodeB, ref.EdgeID, 10)
				} else {
					dyn_top.AddEdgeEntries(edge.NodeA, edge.NodeB, ref.EdgeID, 20)
				}
			}
		})
	}

	return AdjacencyListToArray(&dyn_top), shortcuts
}
func CreateCHSkipTopology2(graph *Graph, ch_data *_CHData, border_nodes Array[bool], node_tiles Array[int16]) (*AdjacencyArray, ShortcutStore) {
	dyn_top := NewAdjacencyList(graph.NodeCount())
	shortcuts := NewShortcutStore(100, true)

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

	for i := 0; i < temp_graph.NodeCount(); i++ {
		if !border_nodes[i] {
			continue
		}
		explorer.ForAdjacentEdges(int32(i), FORWARD, ADJACENT_ALL, func(ref EdgeRef) {
			if !border_nodes[ref.OtherID] {
				return
			}
			if ref.IsShortcut() {
				shc := temp_graph.GetShortcut(ref.EdgeID)
				shc_n := NewShortcut(shc.From, shc.To, explorer.GetEdgeWeight(ref))
				edges_n := [2]Tuple[int32, byte]{}
				shc_id, _ := shortcuts.AddCHShortcut(shc_n, edges_n)
				dyn_top.AddEdgeEntries(shc.From, shc.To, shc_id, 100)
			} else {
				edge := temp_graph.GetEdge(ref.EdgeID)
				if node_tiles[edge.NodeA] != node_tiles[edge.NodeB] {
					dyn_top.AddEdgeEntries(edge.NodeA, edge.NodeB, ref.EdgeID, 10)
				} else {
					dyn_top.AddEdgeEntries(edge.NodeA, edge.NodeB, ref.EdgeID, 20)
				}
			}
		})
	}

	return AdjacencyListToArray(&dyn_top), shortcuts
}

// computes border and interior nodes of graph tile
func GetBorderNodes2(graph ITiledGraph, tile_id int16) (Array[int32], Array[int32]) {
	border := NewList[int32](100)
	interior := NewList[int32](100)

	explorer := graph.GetGraphExplorer()
	for i := 0; i < graph.NodeCount(); i++ {
		curr_tile := graph.GetNodeTile(int32(i))
		if curr_tile != tile_id {
			continue
		}
		is_border := false
		explorer.ForAdjacentEdges(int32(i), FORWARD, ADJACENT_EDGES, func(ref EdgeRef) {
			if ref.IsCrossBorder() {
				is_border = true
			}
		})
		explorer.ForAdjacentEdges(int32(i), BACKWARD, ADJACENT_EDGES, func(ref EdgeRef) {
			if ref.IsCrossBorder() {
				is_border = true
			}
		})
		if is_border {
			border.Add(int32(i))
		} else {
			interior.Add(int32(i))
		}
	}
	return Array[int32](border), Array[int32](interior)
}

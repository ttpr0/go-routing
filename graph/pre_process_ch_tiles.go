package graph

import (
	"fmt"
	"sort"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// preprocess tiled graph 4
//*******************************************

func TransformToTiled4(graph *TiledGraph) *TiledGraph3 {
	fmt.Println("Compute node ordering:")
	contraction_order, border_nodes := TiledNodeOrdering(graph)

	fmt.Println("Compute subset contraction:")
	g := &Graph{
		store:    graph.store,
		topology: graph.topology,
		weight:   graph.weight,
		index:    graph.index,
	}
	dg := TransformToCHPreprocGraph(g)
	CalcContraction2(dg, contraction_order)

	fmt.Println("Set border nodes to maxlevel:")
	max_level := int16(0)
	node_levels := dg.node_levels
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
	node_tiles := graph.skip_store.node_tiles
	skip_topology, shortcuts, skip_weights := CreateCHSkipTopology2(dg, border_nodes, node_tiles)
	new_graph := &TiledGraph{
		store:    graph.store,
		topology: graph.topology,
		weight:   graph.weight,
		index:    graph.index,

		skip_store: TiledStore{
			node_tiles:   graph.skip_store.node_tiles,
			shortcuts:    shortcuts,
			edge_refs:    NewList[Tuple[int32, byte]](0),
			skip_weights: skip_weights,
			edge_types:   graph.skip_store.edge_types,
		},
		skip_topology: *skip_topology,
	}

	fmt.Println("Create downwards edge lists:")
	edges := graph.store.edges
	edge_weigths := graph.weight.edge_weights
	ch_shortcuts := dg.shortcuts
	ch_weights := dg.sh_weight
	tiles := GetTiles(graph)
	index_edges := NewList[TiledSHEdge](100)
	tile_ranges := NewDict[int16, Tuple[int32, int32]](tiles.Length())
	for index, tile := range tiles {
		fmt.Println("Process Tile:", index+1, "/", len(tiles))
		// get all down edges or shortcuts
		edge_list := NewList[TiledSHEdge](100)
		for i := 0; i < ch_shortcuts.Length(); i++ {
			shc := ch_shortcuts[i]
			if node_tiles[shc.NodeA] != tile || node_tiles[shc.NodeB] != tile {
				continue
			}
			if node_levels[shc.NodeA] > node_levels[shc.NodeB] {
				edge_list.Add(TiledSHEdge{
					From:   shc.NodeA,
					To:     shc.NodeB,
					Weight: ch_weights[i],
				})
			}
		}
		for i := 0; i < edges.Length(); i++ {
			edge := edges[i]
			if node_tiles[edge.NodeA] != tile || node_tiles[edge.NodeB] != tile {
				continue
			}
			if node_levels[edge.NodeA] > node_levels[edge.NodeB] {
				edge_list.Add(TiledSHEdge{
					From:   edge.NodeA,
					To:     edge.NodeB,
					Weight: edge_weigths[i],
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
		start := index_edges.Length()
		count := 0
		for _, e := range edge_list {
			index_edges.Add(e)
			count += 1
		}
		tile_ranges[tile] = MakeTuple(int32(start), int32(count))
	}

	return &TiledGraph3{
		TiledGraph:  *new_graph,
		tile_ranges: tile_ranges,
		index_edges: Array[TiledSHEdge](index_edges),
	}
}

//*******************************************
// compute node ordering
//*******************************************

func TiledNodeOrdering(graph *TiledGraph) (Array[int32], Array[bool]) {
	tiles := GetTiles(graph)
	border_nodes := NewArray[bool](graph.NodeCount())
	sp_counts := NewArray[int](graph.NodeCount())
	explorer := graph.GetDefaultExplorer()

	// compute border_nodes and sp_counts for every tile
	fmt.Println("start computing tiles...")
	for index, tile := range tiles {
		fmt.Println("    - process tile:", index+1, "/", len(tiles))
		b_nodes, i_nodes := GetBorderNodes2(graph, tile)
		flags := NewDict[int32, _Flag](100)
		for _, b_node := range b_nodes {
			border_nodes[b_node] = true
			flags.Clear()
			CalcFullSPT(graph, b_node, flags)
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
func CreateCHSkipTopology(graph *TiledGraph, ch_shortcuts List[CHShortcut], ch_weights List[int32], border_nodes Array[bool]) (*TypedTopologyStore, List[Shortcut], List[int32]) {
	dyn_top := NewDynamicTopology(graph.NodeCount())

	edge_types := graph.skip_store.edge_types
	for i := 0; i < graph.EdgeCount(); i++ {
		edge_id := int32(i)
		edge := graph.GetEdge(edge_id)
		if edge_types[edge_id] == 10 {
			dyn_top.AddEdgeEntries(edge.NodeA, edge.NodeB, edge_id, 10)
		} else if border_nodes[edge.NodeA] && border_nodes[edge.NodeB] {
			dyn_top.AddEdgeEntries(edge.NodeA, edge.NodeB, edge_id, 20)
		}
	}

	shortcuts := NewList[Shortcut](100)
	shortcut_weights := NewList[int32](100)
	for i := 0; i < ch_shortcuts.Length(); i++ {
		shc := ch_shortcuts[i]
		if border_nodes[shc.NodeA] && border_nodes[shc.NodeB] {
			shc_id := int32(shortcuts.Length())
			shortcuts.Add(Shortcut{
				NodeA: shc.NodeA,
				NodeB: shc.NodeB,
			})
			shortcut_weights.Add(ch_weights[i])
			dyn_top.AddEdgeEntries(shc.NodeA, shc.NodeB, shc_id, 100)
		}
	}

	return DynamicToTypedTopology(&dyn_top), shortcuts, shortcut_weights
}

// creates topology with cross-border edges (type 10), skip-edges (type 20) and shortcuts (type 100)
func CreateCHSkipTopology2(dg *CHPreprocGraph, border_nodes Array[bool], node_tiles Array[int16]) (*TypedTopologyStore, List[Shortcut], List[int32]) {
	dyn_top := NewDynamicTopology(dg.NodeCount())
	shortcuts := NewList[Shortcut](100)
	shortcut_weights := NewList[int32](100)

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
				shc_id := int32(shortcuts.Length())
				shortcuts.Add(Shortcut{
					NodeA: shc.NodeA,
					NodeB: shc.NodeB,
				})
				shortcut_weights.Add(explorer.GetEdgeWeight(ref))
				dyn_top.AddEdgeEntries(shc.NodeA, shc.NodeB, shc_id, 100)
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

	return DynamicToTypedTopology(&dyn_top), shortcuts, shortcut_weights
}

// computes border and interior nodes of graph tile
func GetBorderNodes2(graph ITiledGraph, tile_id int16) (Array[int32], Array[int32]) {
	border := NewList[int32](100)
	interior := NewList[int32](100)

	explorer := graph.GetDefaultExplorer()
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

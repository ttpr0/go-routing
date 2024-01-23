package graph

import (
	"fmt"
	"sort"

	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// preprocess isophast overlay
//*******************************************

func PrepareIsoPHAST(graph *Graph, partition *Partition) (*Overlay, *CellIndex) {
	fmt.Println("Compute subset contraction:")
	ch_data := CalcPartialContraction5(graph, partition)

	fmt.Println("Set border nodes to maxlevel:")
	border_nodes := _IsBorderNode3(graph, partition)
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
	_UpdateCrossBorder(graph, partition, edge_types)
	skip_topology, skip_shortcuts := CreateCHSkipTopology(graph, ch_data, border_nodes, partition)
	tiled_data := &Overlay{
		skip_shortcuts: skip_shortcuts,
		skip_topology:  *skip_topology,
		edge_types:     edge_types,
	}

	// create cell-index
	fmt.Println("Create downwards edge lists:")
	explorer := graph.GetGraphExplorer()
	tiles := partition.GetTiles()
	ch_shortcuts := ch_data.shortcuts
	node_levels = ch_data.node_levels
	cell_index := _NewCellIndex()
	for index, tile := range tiles {
		fmt.Println("Process Tile:", index+1, "/", len(tiles))
		// get all down edges or shortcuts
		edge_list := NewList[Shortcut](100)
		for i := 0; i < ch_shortcuts.ShortcutCount(); i++ {
			shc := ch_shortcuts.GetShortcut(int32(i))
			if partition.GetNodeTile(shc.From) != tile || partition.GetNodeTile(shc.To) != tile {
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
		for i := 0; i < graph.EdgeCount(); i++ {
			edge := graph.GetEdge(int32(i))
			if partition.GetNodeTile(edge.NodeA) != tile || partition.GetNodeTile(edge.NodeB) != tile {
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
		sort.SliceStable(edge_list, func(i, j int) bool {
			e_i := edge_list[i]
			level_i := node_levels[e_i.From]
			e_j := edge_list[j]
			level_j := node_levels[e_j.From]
			return level_i > level_j
		})

		// add edges to index_edges
		cell_index.SetFWDIndexEdges(tile, Array[Shortcut](edge_list))
	}

	return tiled_data, &cell_index
}

//*******************************************
// create topology store
//*******************************************

// creates topology with cross-border edges (type 10), skip-edges (type 20) and shortcuts (type 100)
func CreateCHSkipTopology(graph *Graph, ch_data *CH, border_nodes Array[bool], partition *Partition) (*_AdjacencyArray, _ShortcutStore) {
	dyn_top := _NewAdjacencyList(graph.NodeCount())
	shortcuts := _NewShortcutStore(100, true)

	temp_graph := CHGraph{
		base:   graph.base,
		weight: graph.weight,

		ch_shortcuts: ch_data.shortcuts,
		ch_topology:  ch_data.topology,
		node_levels:  ch_data.node_levels,
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
				if partition.GetNodeTile(edge.NodeA) != partition.GetNodeTile(edge.NodeB) {
					dyn_top.AddEdgeEntries(edge.NodeA, edge.NodeB, ref.EdgeID, 10)
				} else {
					dyn_top.AddEdgeEntries(edge.NodeA, edge.NodeB, ref.EdgeID, 20)
				}
			}
		})
	}

	return _AdjacencyListToArray(&dyn_top), shortcuts
}

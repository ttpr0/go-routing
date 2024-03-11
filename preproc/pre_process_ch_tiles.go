package preproc

import (
	"fmt"
	"sort"

	"github.com/ttpr0/go-routing/comps"
	"github.com/ttpr0/go-routing/graph"
	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
	"golang.org/x/exp/slog"
)

//*******************************************
// preprocess isophast overlay
//*******************************************

func PrepareIsoPHAST(base comps.IGraphBase, weight comps.IWeighting, partition *comps.Partition) (*comps.Overlay, *comps.CellIndex) {
	slog.Debug("Compute subset contraction:")
	ch_data := CalcPartialContraction5(base, weight, partition)

	g := graph.BuildGraph(base, weight, None[comps.IGraphIndex]())

	slog.Debug("Set border nodes to maxlevel:")
	border_nodes := _IsBorderNode3(g, partition)
	max_level := int16(0)
	node_levels := NewArray[int16](base.NodeCount())
	for i := 0; i < node_levels.Length(); i++ {
		node_levels[i] = ch_data.GetNodeLevel(int32(i))
		if node_levels[i] > max_level {
			max_level = node_levels[i]
		}
	}
	for i := 0; i < node_levels.Length(); i++ {
		if border_nodes[i] {
			node_levels[i] = max_level + 1
		}
	}

	slog.Debug("Create topology from shortcuts:")
	edge_types := NewArray[byte](g.EdgeCount())
	_UpdateCrossBorder(g, partition, edge_types)
	skip_topology, skip_shortcuts := CreateCHSkipTopology(base, weight, ch_data, border_nodes, partition)
	tiled_data := comps.NewOverlay(skip_shortcuts, *skip_topology, edge_types)

	// create cell-index
	slog.Debug("Create downwards edge lists:")
	explorer := g.GetGraphExplorer()
	tiles := partition.GetTiles()
	cell_index := comps.NewCellIndex()
	for index, tile := range tiles {
		slog.Debug(fmt.Sprintf("Process Tile: %v/%v", index+1, len(tiles)))
		// get all down edges or shortcuts
		edge_list := NewList[structs.Shortcut](100)
		for i := 0; i < ch_data.ShortcutCount(); i++ {
			shc := ch_data.GetShortcut(int32(i))
			if partition.GetNodeTile(shc.From) != tile || partition.GetNodeTile(shc.To) != tile {
				continue
			}
			if node_levels[shc.From] > node_levels[shc.To] {
				edge_list.Add(structs.Shortcut{
					From:   shc.From,
					To:     shc.From,
					Weight: shc.Weight,
				})
			}
		}
		for i := 0; i < g.EdgeCount(); i++ {
			edge := g.GetEdge(int32(i))
			if partition.GetNodeTile(edge.NodeA) != tile || partition.GetNodeTile(edge.NodeB) != tile {
				continue
			}
			if node_levels[edge.NodeA] > node_levels[edge.NodeB] {
				edge_list.Add(structs.Shortcut{
					From:   edge.NodeA,
					To:     edge.NodeB,
					Weight: explorer.GetEdgeWeight(graph.CreateEdgeRef(int32(i))),
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
		cell_index.SetFWDIndexEdges(tile, Array[structs.Shortcut](edge_list))
	}

	return tiled_data, &cell_index
}

//*******************************************
// create topology store
//*******************************************

// creates topology with cross-border edges (type 10), skip-edges (type 20) and shortcuts (type 100)
func CreateCHSkipTopology(base comps.IGraphBase, weight comps.IWeighting, ch_data *comps.CH, border_nodes Array[bool], partition *comps.Partition) (*structs.AdjacencyArray, structs.ShortcutStore) {
	dyn_top := structs.NewAdjacencyList(base.NodeCount())
	shortcuts := structs.NewShortcutStore(100, true)

	temp_graph := graph.BuildCHGraph(base, weight, None[comps.IGraphIndex](), ch_data, None[*comps.CHIndex]())
	explorer := temp_graph.GetGraphExplorer()

	for i := 0; i < temp_graph.NodeCount(); i++ {
		if !border_nodes[i] {
			continue
		}
		explorer.ForAdjacentEdges(int32(i), graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			if !border_nodes[ref.OtherID] {
				return
			}
			if ref.IsShortcut() {
				shc := temp_graph.GetShortcut(ref.EdgeID)
				shc_n := structs.NewShortcut(shc.From, shc.To, explorer.GetEdgeWeight(ref))
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

	return structs.AdjacencyListToArray(&dyn_top), shortcuts
}

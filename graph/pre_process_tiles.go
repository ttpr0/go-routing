package graph

import (
	"fmt"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// preprocess tiled-graph
//*******************************************

// Creates tiled-graph with skeleton cliques.
func PreprocessTiledGraph(graph *Graph, node_tiles Array[int16]) *_TiledData {
	skip_shortcuts := NewShortcutStore(100, false)
	edge_types := NewArray[byte](graph.EdgeCount())

	_UpdateCrossBorder(graph, node_tiles, edge_types)

	tiles := NewDict[int16, bool](100)
	for _, tile_id := range node_tiles {
		if tiles.ContainsKey(tile_id) {
			continue
		}
		tiles[tile_id] = true
	}

	tile_count := tiles.Length()
	c := 1
	for tile_id, _ := range tiles {
		fmt.Printf("tile %v: %v / %v \n", tile_id, c, tile_count)
		fmt.Printf("tile %v: getting start nodes \n", tile_id)
		start_nodes, end_nodes := _GetInOutNodes(graph, tile_id, node_tiles, edge_types)
		fmt.Printf("tile %v: calculating skip edges \n", tile_id)
		_CalcSkipEdges(graph, start_nodes, end_nodes, edge_types)
		fmt.Printf("tile %v: finished \n", tile_id)
		c += 1
	}

	skip_topology := _CreateSkipTopology(graph, &skip_shortcuts, edge_types)

	return &_TiledData{
		skip_shortcuts: skip_shortcuts,
		skip_topology:  *skip_topology,
		node_tiles:     node_tiles,
		edge_types:     edge_types,

		_base_weighting: graph._weight_name,
	}
}

// Creates tiled-graph with full-shortcut cliques.
func PreprocessTiledGraph3(graph *Graph, node_tiles Array[int16]) *_TiledData {
	skip_shortcuts := NewShortcutStore(100, false)
	edge_types := NewArray[byte](graph.EdgeCount())

	_UpdateCrossBorder(graph, node_tiles, edge_types)

	tiles := NewDict[int16, bool](100)
	for _, tile_id := range node_tiles {
		if tiles.ContainsKey(tile_id) {
			continue
		}
		tiles[tile_id] = true
	}

	tile_count := tiles.Length()
	c := 1
	for tile_id, _ := range tiles {
		fmt.Printf("tile %v: %v / %v \n", tile_id, c, tile_count)
		fmt.Printf("tile %v: getting start nodes \n", tile_id)
		start_nodes, end_nodes := _GetInOutNodes(graph, tile_id, node_tiles, edge_types)
		fmt.Printf("tile %v: calculating skip edges \n", tile_id)
		_CalcShortcutEdges(graph, start_nodes, end_nodes, edge_types, &skip_shortcuts)
		fmt.Printf("tile %v: finished \n", tile_id)
		c += 1
	}

	skip_topology := _CreateSkipTopology(graph, &skip_shortcuts, edge_types)

	return &_TiledData{
		skip_shortcuts: skip_shortcuts,
		skip_topology:  *skip_topology,
		node_tiles:     node_tiles,
		edge_types:     edge_types,

		_base_weighting: graph._weight_name,
	}
}

//*******************************************
// preprocessing utility methods
//*******************************************

// return list of nodes that have at least one cross-border edge
//
// returns in_nodes, out_nodes
func _GetInOutNodes(graph *Graph, tile_id int16, node_tiles Array[int16], edge_types Array[byte]) (List[int32], List[int32]) {
	in_list := NewList[int32](100)
	out_list := NewList[int32](100)

	explorer := graph.GetGraphExplorer()
	for i := 0; i < graph.NodeCount(); i++ {
		id := i
		tile := node_tiles[id]
		if tile != tile_id {
			continue
		}
		is_added := false
		explorer.ForAdjacentEdges(int32(id), BACKWARD, ADJACENT_ALL, func(ref EdgeRef) {
			if is_added {
				return
			}
			if edge_types[ref.EdgeID] == 10 {
				in_list.Add(int32(id))
				is_added = true
			}
		})

		is_added = false
		explorer.ForAdjacentEdges(int32(id), FORWARD, ADJACENT_ALL, func(ref EdgeRef) {
			if is_added {
				return
			}
			if edge_types[ref.EdgeID] == 10 {
				out_list.Add(int32(id))
				is_added = true
			}
		})
	}
	return in_list, out_list
}

// sets edge type of cross border edges to 10
func _UpdateCrossBorder(graph *Graph, node_tiles Array[int16], edge_types Array[byte]) {
	for i := 0; i < graph.EdgeCount(); i++ {
		edge := graph.GetEdge(int32(i))
		if node_tiles[edge.NodeA] != node_tiles[edge.NodeB] {
			edge_types[i] = 10
		}
	}
}

//*******************************************
// compute clique
//*******************************************

type _Flag struct {
	pathlength int32
	prevEdge   int32
	visited    bool
}

// marks every edge as that lies on a shortest path between border nodes with edge_type 20
func _CalcSkipEdges(graph *Graph, start_nodes, end_nodes List[int32], edge_types Array[byte]) {
	explorer := graph.GetGraphExplorer()
	for _, start := range start_nodes {
		heap := NewPriorityQueue[int32, int32](10)
		flags := NewDict[int32, _Flag](10)

		flags[start] = _Flag{pathlength: 0, visited: false, prevEdge: -1}
		heap.Enqueue(start, 0)

		for {
			curr_id, ok := heap.Dequeue()
			if !ok {
				break
			}
			curr_flag := flags[curr_id]
			if curr_flag.visited {
				continue
			}
			curr_flag.visited = true
			explorer.ForAdjacentEdges(curr_id, FORWARD, ADJACENT_ALL, func(ref EdgeRef) {
				if !ref.IsEdge() {
					return
				}
				edge_id := ref.EdgeID
				if edge_types[edge_id] == 10 {
					return
				}
				other_id := explorer.GetOtherNode(ref, curr_id)
				var other_flag _Flag
				if flags.ContainsKey(other_id) {
					other_flag = flags[other_id]
				} else {
					other_flag = _Flag{pathlength: 10000000, visited: false, prevEdge: -1}
				}
				if other_flag.visited {
					return
				}
				weight := explorer.GetEdgeWeight(ref)
				newlength := curr_flag.pathlength + weight
				if newlength < other_flag.pathlength {
					other_flag.pathlength = newlength
					other_flag.prevEdge = edge_id
					heap.Enqueue(other_id, newlength)
				}
				flags[other_id] = other_flag
			})
			flags[curr_id] = curr_flag
		}

		for _, end := range end_nodes {
			if !flags.ContainsKey(end) {
				continue
			}
			curr_id := end
			for {
				if curr_id == start {
					break
				}
				edge_id := flags[curr_id].prevEdge
				edge_types[edge_id] = 20
				curr_id = explorer.GetOtherNode(EdgeRef{EdgeID: edge_id}, curr_id)
			}
		}
	}
}

// computes shortest paths from every start to end node and adds shortcuts
func _CalcShortcutEdges(graph *Graph, start_nodes, end_nodes List[int32], edge_types Array[byte], shortcuts *ShortcutStore) {
	explorer := graph.GetGraphExplorer()
	for _, start := range start_nodes {
		heap := NewPriorityQueue[int32, int32](10)
		flags := NewDict[int32, _Flag](10)

		flags[start] = _Flag{pathlength: 0, visited: false, prevEdge: -1}
		heap.Enqueue(start, 0)

		for {
			curr_id, ok := heap.Dequeue()
			if !ok {
				break
			}
			curr_flag := flags[curr_id]
			if curr_flag.visited {
				continue
			}
			curr_flag.visited = true
			explorer.ForAdjacentEdges(curr_id, FORWARD, ADJACENT_ALL, func(ref EdgeRef) {
				if !ref.IsEdge() {
					return
				}
				edge_id := ref.EdgeID
				if edge_types[edge_id] == 10 {
					return
				}
				other_id := ref.OtherID
				var other_flag _Flag
				if flags.ContainsKey(other_id) {
					other_flag = flags[other_id]
				} else {
					other_flag = _Flag{pathlength: 10000000, visited: false, prevEdge: -1}
				}
				if other_flag.visited {
					return
				}
				weight := explorer.GetEdgeWeight(ref)
				newlength := curr_flag.pathlength + weight
				if newlength < other_flag.pathlength {
					other_flag.pathlength = newlength
					other_flag.prevEdge = edge_id
					heap.Enqueue(other_id, newlength)
				}
				flags[other_id] = other_flag
			})
			flags[curr_id] = curr_flag
		}

		for _, end := range end_nodes {
			if !flags.ContainsKey(end) {
				continue
			}
			path := make([]int32, 0)
			length := int32(flags[end].pathlength)
			curr_id := end
			var edge int32
			for {
				if curr_id == start {
					break
				}
				edge = flags[curr_id].prevEdge
				path = append(path, edge)
				curr_id = explorer.GetOtherNode(EdgeRef{EdgeID: edge}, curr_id)
			}
			shc := NewShortcut(start, end, length)
			shortcuts.AddShortcut(shc, path)
		}
	}
}

//*******************************************
// create topology store
//*******************************************

// creates topology with cross-border edges (type 10), skip edges (type 20) and shortcuts (type 100)
func _CreateSkipTopology(graph *Graph, shortcuts *ShortcutStore, edge_types Array[byte]) *AdjacencyArray {
	dyn_top := NewAdjacencyList(graph.NodeCount())

	for i := 0; i < graph.EdgeCount(); i++ {
		edge_id := int32(i)
		edge_typ := edge_types[edge_id]
		if edge_typ != 10 && edge_typ != 20 {
			continue
		}
		edge := graph.GetEdge(edge_id)
		dyn_top.AddEdgeEntries(edge.NodeA, edge.NodeB, edge_id, edge_typ)
	}

	for i := 0; i < shortcuts.ShortcutCount(); i++ {
		shc_id := int32(i)
		shc := shortcuts.GetShortcut(shc_id)
		dyn_top.AddEdgeEntries(shc.From, shc.To, shc_id, 100)
	}

	return AdjacencyListToArray(&dyn_top)
}

//*******************************************
// preprocess tiled-graph index
//*******************************************

func PrepareGRASPCellIndex(graph *TiledGraph) {
	tiles := _GetTiles(graph.node_tiles)
	cell_index := _NewCellIndex()
	for index, tile := range tiles {
		fmt.Println("Process Tile:", index, "/", len(tiles))
		index_edges := NewList[Shortcut](4)
		b_nodes, i_nodes := _GetBorderNodes(graph, tile)
		flags := NewDict[int32, _Flag](100)
		for _, b_node := range b_nodes {
			flags.Clear()
			_CalcFullSPT(graph, b_node, flags)
			for _, i_node := range i_nodes {
				if flags.ContainsKey(i_node) {
					flag := flags[i_node]
					index_edges.Add(Shortcut{
						From:   b_node,
						To:     i_node,
						Weight: flag.pathlength,
					})
				}
			}
		}
		cell_index.SetFWDIndexEdges(tile, Array[Shortcut](index_edges))
	}
	graph.cell_index = Some(cell_index)
}

// Modifies tiled-data inplace.
func PrepareGRASPCellIndex2(graph *Graph, data *_TiledData) {
	temp_graph := &TiledGraph{
		base:   graph.base,
		weight: graph.weight,

		skip_shortcuts: data.skip_shortcuts,
		skip_topology:  data.skip_topology,
		node_tiles:     data.node_tiles,
		edge_types:     data.edge_types,
		cell_index:     data.cell_index,
	}

	tiles := _GetTiles(temp_graph.node_tiles)
	cell_index := _NewCellIndex()
	for index, tile := range tiles {
		fmt.Println("Process Tile:", index, "/", len(tiles))
		index_edges := NewList[Shortcut](4)
		b_nodes, i_nodes := _GetBorderNodes(temp_graph, tile)
		flags := NewDict[int32, _Flag](100)
		for _, b_node := range b_nodes {
			flags.Clear()
			_CalcFullSPT(temp_graph, b_node, flags)
			for _, i_node := range i_nodes {
				if flags.ContainsKey(i_node) {
					flag := flags[i_node]
					index_edges.Add(Shortcut{
						From:   b_node,
						To:     i_node,
						Weight: flag.pathlength,
					})
				}
			}
		}
		cell_index.SetFWDIndexEdges(tile, Array[Shortcut](index_edges))
	}
	data.cell_index = Some(cell_index)
}

func _GetTiles(tiles Array[int16]) List[int16] {
	tile_dict := NewDict[int16, bool](100)
	for i := 0; i < tiles.Length(); i++ {
		tile_id := tiles[i]
		if tile_dict.ContainsKey(tile_id) {
			continue
		}
		tile_dict[tile_id] = true
	}
	tile_list := NewList[int16](len(tile_dict))
	for tile, _ := range tile_dict {
		tile_list.Add(tile)
	}
	return tile_list
}

// Computes border and interior nodes of graph tile.
// If tile doesn't exist arrays will be empty.
func _GetBorderNodes(graph ITiledGraph, tile_id int16) (Array[int32], Array[int32]) {
	border := NewList[int32](100)
	interior := NewList[int32](100)

	explorer := graph.GetGraphExplorer()
	for i := 0; i < graph.NodeCount(); i++ {
		id := int32(i)
		tile := graph.GetNodeTile(id)
		if tile != tile_id {
			continue
		}
		is_border := false
		explorer.ForAdjacentEdges(int32(id), BACKWARD, ADJACENT_ALL, func(ref EdgeRef) {
			if is_border {
				return
			}
			if ref.IsCrossBorder() {
				border.Add(int32(id))
				is_border = true
			}
		})
		if !is_border {
			interior.Add(int32(id))
		}
	}
	return Array[int32](border), Array[int32](interior)
}

func _CalcFullSPT(graph ITiledGraph, start int32, flags Dict[int32, _Flag]) {
	heap := NewPriorityQueue[int32, int32](10)

	flags[start] = _Flag{pathlength: 0, visited: false, prevEdge: -1}
	heap.Enqueue(start, 0)

	explorer := graph.GetGraphExplorer()
	for {
		curr_id, ok := heap.Dequeue()
		if !ok {
			break
		}
		curr_flag := flags[curr_id]
		if curr_flag.visited {
			continue
		}
		curr_flag.visited = true
		explorer.ForAdjacentEdges(curr_id, FORWARD, ADJACENT_ALL, func(ref EdgeRef) {
			if !ref.IsEdge() || ref.IsCrossBorder() {
				return
			}
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			var other_flag _Flag
			if flags.ContainsKey(other_id) {
				other_flag = flags[other_id]
			} else {
				other_flag = _Flag{pathlength: 10000000, visited: false, prevEdge: -1}
			}
			if other_flag.visited {
				return
			}
			weight := explorer.GetEdgeWeight(ref)
			newlength := curr_flag.pathlength + weight
			if newlength < other_flag.pathlength {
				other_flag.pathlength = newlength
				other_flag.prevEdge = edge_id
				heap.Enqueue(other_id, newlength)
			}
			flags[other_id] = other_flag
		})
		flags[curr_id] = curr_flag
	}
}

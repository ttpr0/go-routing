package graph

import (
	"fmt"

	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// preprocess overlay
//*******************************************

// Creates overlay with skeleton cliques.
func PrepareSkeletonOverlay(graph *Graph, partition *Partition) *Overlay {
	skip_shortcuts := _NewShortcutStore(100, false)
	edge_types := NewArray[byte](graph.EdgeCount())

	_UpdateCrossBorder(graph, partition, edge_types)

	tiles := partition.GetTiles()

	tile_count := tiles.Length()
	c := 1
	for _, tile_id := range tiles {
		fmt.Printf("tile %v: %v / %v \n", tile_id, c, tile_count)
		fmt.Printf("tile %v: getting start nodes \n", tile_id)
		start_nodes, end_nodes := _GetInOutNodes(graph, tile_id, partition)
		fmt.Printf("tile %v: calculating skip edges \n", tile_id)
		_CalcSkipEdges(graph, start_nodes, end_nodes, edge_types)
		fmt.Printf("tile %v: finished \n", tile_id)
		c += 1
	}

	skip_topology := _CreateSkipTopology(graph, &skip_shortcuts, edge_types)

	return &Overlay{
		skip_shortcuts: skip_shortcuts,
		skip_topology:  *skip_topology,
		edge_types:     edge_types,
	}
}

// Creates overlay with full-shortcut cliques.
func PrepareOverlay(graph *Graph, partition *Partition) *Overlay {
	skip_shortcuts := _NewShortcutStore(100, false)
	edge_types := NewArray[byte](graph.EdgeCount())

	_UpdateCrossBorder(graph, partition, edge_types)

	tiles := partition.GetTiles()

	tile_count := tiles.Length()
	c := 1
	for _, tile_id := range tiles {
		fmt.Printf("tile %v: %v / %v \n", tile_id, c, tile_count)
		fmt.Printf("tile %v: getting start nodes \n", tile_id)
		start_nodes, end_nodes := _GetInOutNodes(graph, tile_id, partition)
		fmt.Printf("tile %v: calculating skip edges \n", tile_id)
		_CalcShortcutEdges(graph, start_nodes, end_nodes, edge_types, &skip_shortcuts)
		fmt.Printf("tile %v: finished \n", tile_id)
		c += 1
	}

	skip_topology := _CreateSkipTopology(graph, &skip_shortcuts, edge_types)

	return &Overlay{
		skip_shortcuts: skip_shortcuts,
		skip_topology:  *skip_topology,
		edge_types:     edge_types,
	}
}

//*******************************************
// preprocessing utility methods
//*******************************************

// return list of nodes that have at least one cross-border edge
//
// returns in_nodes, out_nodes
func _GetInOutNodes(graph *Graph, tile_id int16, partition *Partition) (List[int32], List[int32]) {
	in_list := NewList[int32](100)
	out_list := NewList[int32](100)

	explorer := graph.GetGraphExplorer()
	for i := 0; i < graph.NodeCount(); i++ {
		id := int32(i)
		tile := partition.GetNodeTile(id)
		if tile != tile_id {
			continue
		}
		is_added := false
		explorer.ForAdjacentEdges(int32(id), BACKWARD, ADJACENT_ALL, func(ref EdgeRef) {
			if is_added {
				return
			}
			other_tile := partition.GetNodeTile(ref.OtherID)
			if other_tile != tile {
				in_list.Add(int32(id))
				is_added = true
			}
		})

		is_added = false
		explorer.ForAdjacentEdges(int32(id), FORWARD, ADJACENT_ALL, func(ref EdgeRef) {
			if is_added {
				return
			}
			other_tile := partition.GetNodeTile(ref.OtherID)
			if other_tile != tile {
				out_list.Add(int32(id))
				is_added = true
			}
		})
	}
	return in_list, out_list
}

// sets edge type of cross border edges to 10
func _UpdateCrossBorder(graph *Graph, partition *Partition, edge_types Array[byte]) {
	for i := 0; i < graph.EdgeCount(); i++ {
		edge := graph.GetEdge(int32(i))
		if partition.GetNodeTile(edge.NodeA) != partition.GetNodeTile(edge.NodeB) {
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
func _CalcShortcutEdges(graph *Graph, start_nodes, end_nodes List[int32], edge_types Array[byte], shortcuts *_ShortcutStore) {
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
func _CreateSkipTopology(graph *Graph, shortcuts *_ShortcutStore, edge_types Array[byte]) *_AdjacencyArray {
	dyn_top := _NewAdjacencyList(graph.NodeCount())

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

	return _AdjacencyListToArray(&dyn_top)
}

//*******************************************
// preprocess cell-index
//*******************************************

// Creates GRASP cell-index for partitioned tiles.
func PrepareGRASPCellIndex(graph *Graph, partition *Partition) *CellIndex {
	tiles := partition.GetTiles()
	cell_index := _NewCellIndex()
	for index, tile := range tiles {
		fmt.Println("Process Tile:", index, "/", len(tiles))
		index_edges := NewList[Shortcut](4)
		b_nodes, i_nodes := _GetBorderNodes(graph, partition, tile)
		flags := NewDict[int32, _Flag](100)
		for _, b_node := range b_nodes {
			flags.Clear()
			_CalcFullSPT(graph, b_node, partition, flags)
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
	return &cell_index
}

// Computes border and interior nodes of graph tile.
// If tile doesn't exist arrays will be empty.
func _GetBorderNodes(graph IGraph, partition *Partition, tile_id int16) (Array[int32], Array[int32]) {
	border := NewList[int32](100)
	interior := NewList[int32](100)

	explorer := graph.GetGraphExplorer()
	for i := 0; i < graph.NodeCount(); i++ {
		id := int32(i)
		tile := partition.GetNodeTile(id)
		if tile != tile_id {
			continue
		}
		is_border := false
		explorer.ForAdjacentEdges(int32(id), BACKWARD, ADJACENT_ALL, func(ref EdgeRef) {
			if is_border {
				return
			}
			other_tile := partition.GetNodeTile(ref.OtherID)
			if tile != other_tile {
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

// Computes shortest-path distances to all interior nodes of cell.
func _CalcFullSPT(graph IGraph, start int32, partition *Partition, flags Dict[int32, _Flag]) {
	heap := NewPriorityQueue[int32, int32](10)

	flags[start] = _Flag{pathlength: 0, visited: false, prevEdge: -1}
	heap.Enqueue(start, 0)

	tile := partition.GetNodeTile(start)
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
			if !ref.IsEdge() {
				return
			}
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			other_tile := partition.GetNodeTile(other_id)
			if tile != other_tile {
				return
			}
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

package graph

import (
	"fmt"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// preprocess tiled-graph
//*******************************************

func PreprocessTiledGraph(graph *Graph, node_tiles Array[int16]) *TiledGraph {
	skip_store := TiledStore{
		node_tiles:   node_tiles,
		shortcuts:    NewList[Shortcut](100),
		edge_refs:    NewList[Tuple[int32, byte]](100),
		skip_weights: NewList[int32](100),
		edge_types:   NewArray[byte](graph.EdgeCount()),
	}

	UpdateCrossBorder(&skip_store, graph)

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
		start_nodes, end_nodes := GetInOutNodes(graph, &skip_store, tile_id)
		fmt.Printf("tile %v: calculating skip edges \n", tile_id)
		CalcSkipEdges(graph, start_nodes, end_nodes, &skip_store)
		fmt.Printf("tile %v: finished \n", tile_id)
		c += 1
	}

	skip_topology := CreateSkipTopology(graph, &skip_store)

	tiled_graph := &TiledGraph{
		store:         graph.store,
		topology:      graph.topology,
		skip_topology: *skip_topology,
		skip_store:    skip_store,
		weight:        graph.weight,
		index:         graph.index,
	}

	return tiled_graph
}

func PreprocessTiledGraph3(graph *Graph, node_tiles Array[int16]) *TiledGraph {
	skip_store := TiledStore{
		node_tiles:   node_tiles,
		shortcuts:    NewList[Shortcut](100),
		edge_refs:    NewList[Tuple[int32, byte]](100),
		skip_weights: NewList[int32](100),
		edge_types:   NewArray[byte](graph.EdgeCount()),
	}

	UpdateCrossBorder(&skip_store, graph)

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
		start_nodes, end_nodes := GetInOutNodes(graph, &skip_store, tile_id)
		fmt.Printf("tile %v: calculating skip edges \n", tile_id)
		CalcShortcutEdges(graph, start_nodes, end_nodes, &skip_store)
		fmt.Printf("tile %v: finished \n", tile_id)
		c += 1
	}

	skip_topology := CreateSkipTopology(graph, &skip_store)

	tiled_graph := &TiledGraph{
		store:         graph.store,
		topology:      graph.topology,
		skip_topology: *skip_topology,
		skip_store:    skip_store,
		weight:        graph.weight,
		index:         graph.index,
	}

	return tiled_graph
}

func PreprocessTiledGraph4(graph *TiledGraph) *TiledGraph {
	g := &Graph{
		store:    graph.store,
		topology: graph.topology,
		weight:   graph.weight,
		index:    graph.index,
	}

	tiles := graph.skip_store.node_tiles

	return PreprocessTiledGraph3(g, tiles)
}

//*******************************************
// preprocessing utility methods
//*******************************************

// return list of nodes that have at least one cross-border edge
//
// returns in_nodes, out_nodes
func GetInOutNodes(graph *Graph, skip_store *TiledStore, tile_id int16) (List[int32], List[int32]) {
	in_list := NewList[int32](100)
	out_list := NewList[int32](100)

	explorer := graph.GetDefaultExplorer()
	for i := 0; i < graph.NodeCount(); i++ {
		id := i
		tile := skip_store.GetNodeTile(int32(id))
		if tile != tile_id {
			continue
		}
		iter := explorer.GetAdjacentEdges(int32(id), BACKWARD, ADJACENT_ALL)
		for {
			ref, ok := iter.Next()
			if !ok {
				break
			}
			if skip_store.GetEdgeType(ref.EdgeID) == 10 {
				in_list.Add(int32(id))
				break
			}
		}

		iter = explorer.GetAdjacentEdges(int32(id), FORWARD, ADJACENT_ALL)
		for {
			ref, ok := iter.Next()
			if !ok {
				break
			}
			if skip_store.GetEdgeType(ref.EdgeID) == 10 {
				out_list.Add(int32(id))
				break
			}
		}
	}
	return in_list, out_list
}

// sets edge type of cross border edges to 10
func UpdateCrossBorder(skip_store *TiledStore, graph *Graph) {
	for i := 0; i < graph.EdgeCount(); i++ {
		edge := graph.GetEdge(int32(i))
		if skip_store.GetNodeTile(edge.NodeA) != skip_store.GetNodeTile(edge.NodeB) {
			skip_store.SetEdgeType(int32(i), 10)
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
func CalcSkipEdges(graph *Graph, start_nodes, end_nodes List[int32], skip_store *TiledStore) {
	explorer := graph.GetDefaultExplorer()
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
			iter := explorer.GetAdjacentEdges(curr_id, FORWARD, ADJACENT_ALL)
			for {
				ref, ok := iter.Next()
				if !ok {
					break
				}
				if !ref.IsEdge() {
					continue
				}
				edge_id := ref.EdgeID
				if skip_store.GetEdgeType(edge_id) == 10 {
					continue
				}
				other_id := explorer.GetOtherNode(ref, curr_id)
				var other_flag _Flag
				if flags.ContainsKey(other_id) {
					other_flag = flags[other_id]
				} else {
					other_flag = _Flag{pathlength: 10000000, visited: false, prevEdge: -1}
				}
				if other_flag.visited {
					continue
				}
				weight := explorer.GetEdgeWeight(ref)
				newlength := curr_flag.pathlength + weight
				if newlength < other_flag.pathlength {
					other_flag.pathlength = newlength
					other_flag.prevEdge = edge_id
					heap.Enqueue(other_id, newlength)
				}
				flags[other_id] = other_flag
			}
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
				skip_store.SetEdgeType(edge_id, 20)
				curr_id = explorer.GetOtherNode(EdgeRef{EdgeID: edge_id}, curr_id)
			}
		}
	}
}

// computes shortest paths from everyv start to end node and adds shortcuts
func CalcShortcutEdges(graph *Graph, start_nodes, end_nodes List[int32], skip_store *TiledStore) {
	explorer := graph.GetDefaultExplorer()
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
			iter := explorer.GetAdjacentEdges(curr_id, FORWARD, ADJACENT_ALL)
			for {
				ref, ok := iter.Next()
				if !ok {
					break
				}
				if !ref.IsEdge() {
					continue
				}
				edge_id := ref.EdgeID
				if skip_store.GetEdgeType(edge_id) == 10 {
					continue
				}
				other_id := ref.OtherID
				var other_flag _Flag
				if flags.ContainsKey(other_id) {
					other_flag = flags[other_id]
				} else {
					other_flag = _Flag{pathlength: 10000000, visited: false, prevEdge: -1}
				}
				if other_flag.visited {
					continue
				}
				weight := explorer.GetEdgeWeight(ref)
				newlength := curr_flag.pathlength + weight
				if newlength < other_flag.pathlength {
					other_flag.pathlength = newlength
					other_flag.prevEdge = edge_id
					heap.Enqueue(other_id, newlength)
				}
				flags[other_id] = other_flag
			}
			flags[curr_id] = curr_flag
		}

		for _, end := range end_nodes {
			if !flags.ContainsKey(end) {
				continue
			}
			path := make([]Tuple[int32, byte], 0)
			length := int32(flags[end].pathlength)
			// curr_id := end
			// var edge int32
			// for {
			// 	if curr_id == start {
			// 		break
			// 	}
			// 	edge = flags[curr_id].prevEdge
			// 	path = append(path, MakeTuple(edge, byte(0)))
			// 	curr_id = explorer.GetOtherNode(EdgeRef{EdgeID: edge}, curr_id)
			// }
			skip_store.AddShortcut(start, end, path, length)
		}
	}
}

//*******************************************
// create topology store
//*******************************************

// creates topology with cross-border edges (type 10), skip edges (type 20) and shortcuts (type 100)
func CreateSkipTopology(graph *Graph, skip_store *TiledStore) *TypedTopologyStore {
	dyn_top := NewDynamicTopology(graph.NodeCount())

	for i := 0; i < graph.EdgeCount(); i++ {
		edge_id := int32(i)
		edge_typ := skip_store.GetEdgeType(edge_id)
		if edge_typ != 10 && edge_typ != 20 {
			continue
		}
		edge := graph.GetEdge(edge_id)
		dyn_top.AddEdgeEntries(edge.NodeA, edge.NodeB, edge_id, edge_typ)
	}

	for i := 0; i < skip_store.ShortcutCount(); i++ {
		shc_id := int32(i)
		shc := skip_store.GetShortcut(shc_id)
		dyn_top.AddEdgeEntries(shc.NodeA, shc.NodeB, shc_id, 100)
	}

	return DynamicToTypedTopology(&dyn_top)
}

//*******************************************
// preprocess tiled-graph index
//*******************************************

func TransformToTiled2(graph *TiledGraph) *TiledGraph2 {
	tiles := GetTiles(graph)
	border_nodes := NewDict[int16, Array[int32]](len(tiles))
	interior_nodes := NewDict[int16, Array[int32]](len(tiles))
	border_range_map := NewDict[int16, Dict[int32, Array[float32]]](len(tiles))
	for index, tile := range tiles {
		fmt.Println("Process Tile:", index, "/", len(tiles))
		b_nodes, i_nodes := GetBorderNodes(graph, tile)
		border_nodes[tile] = b_nodes
		interior_nodes[tile] = i_nodes
		range_map := NewDict[int32, Array[float32]](len(b_nodes))
		flags := NewDict[int32, _Flag](100)
		for _, b_node := range b_nodes {
			flags.Clear()
			CalcFullSPT(graph, b_node, flags)
			ranges := NewArray[float32](len(i_nodes))
			for i, i_node := range i_nodes {
				if flags.ContainsKey(i_node) {
					flag := flags[i_node]
					ranges[i] = float32(flag.pathlength)
				} else {
					ranges[i] = 1000000
				}
			}
			range_map[b_node] = ranges
		}
		border_range_map[tile] = range_map
	}

	return &TiledGraph2{
		TiledGraph:       *graph,
		border_nodes:     border_nodes,
		interior_nodes:   interior_nodes,
		border_range_map: border_range_map,
	}
}

func GetTiles(graph *TiledGraph) List[int16] {
	tile_dict := NewDict[int16, bool](100)
	for _, tile_id := range graph.skip_store.node_tiles {
		if tile_dict.ContainsKey(tile_id) {
			continue
		}
		tile_dict[tile_id] = true
	}
	tiles := NewList[int16](len(tile_dict))
	for tile, _ := range tile_dict {
		tiles.Add(tile)
	}
	return tiles
}

// computes border and interior nodes of graph tile
func GetBorderNodes(graph *TiledGraph, tile_id int16) (Array[int32], Array[int32]) {
	border := NewList[int32](100)
	interior := NewList[int32](100)

	explorer := graph.GetDefaultExplorer()
	for id, tile := range graph.skip_store.node_tiles {
		if tile != tile_id {
			continue
		}
		iter := explorer.GetAdjacentEdges(int32(id), BACKWARD, ADJACENT_ALL)
		is_border := false
		for {
			ref, ok := iter.Next()
			if !ok {
				break
			}
			if ref.IsCrossBorder() {
				border.Add(int32(id))
				is_border = true
				break
			}
		}
		if !is_border {
			interior.Add(int32(id))
		}
	}
	return Array[int32](border), Array[int32](interior)
}

func CalcFullSPT(graph *TiledGraph, start int32, flags Dict[int32, _Flag]) {
	heap := NewPriorityQueue[int32, int32](10)

	flags[start] = _Flag{pathlength: 0, visited: false, prevEdge: -1}
	heap.Enqueue(start, 0)

	explorer := graph.GetDefaultExplorer()
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
		iter := explorer.GetAdjacentEdges(curr_id, FORWARD, ADJACENT_ALL)
		for {
			ref, ok := iter.Next()
			if !ok {
				break
			}
			if !ref.IsEdge() || ref.IsCrossBorder() {
				continue
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
				continue
			}
			weight := explorer.GetEdgeWeight(ref)
			newlength := curr_flag.pathlength + weight
			if newlength < other_flag.pathlength {
				other_flag.pathlength = newlength
				other_flag.prevEdge = edge_id
				heap.Enqueue(other_id, newlength)
			}
			flags[other_id] = other_flag
		}
		flags[curr_id] = curr_flag
	}
}

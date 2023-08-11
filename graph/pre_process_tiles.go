package graph

import (
	"fmt"
	"strconv"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// partition graph
//*******************************************

// computes node tiles geo-polygons
func CalcNodeTiles(geom IGeometry, features []geo.Feature) List[int16] {
	node_coords := geom.GetAllNodes()
	node_tiles := make([]int16, len(node_coords))

	c := 0
	for i, node := range node_coords {
		if c%1000 == 0 {
			fmt.Println("finished node ", c)
		}
		point := geo.NewPoint(node)
		node_tiles[i] = -1
		for _, feature := range features {
			polygon := feature.Geometry()
			if polygon.Contains(&point) {
				tile_id := feature.Properties()["TileID"]
				id, _ := strconv.Atoi(tile_id.(string))
				node_tiles[i] = int16(id)
				break
			}
		}
		c += 1
	}

	return node_tiles
}

//*******************************************
// preprocess tiled-graph
//*******************************************

func PreprocessTiledGraph(graph *Graph, node_tiles List[int16]) *TiledGraph {
	edge_types := NewArray[byte](graph.edges.EdgeCount())

	tiled_graph := &TiledGraph{
		topology:   graph.topology,
		nodes:      graph.nodes,
		node_tiles: NodeTileStore{node_tiles: Array[int16](node_tiles)},
		edges:      graph.edges,
		edge_types: edge_types,
		geom:       graph.geom,
		weight:     graph.weight,
		index:      graph.index,
	}

	UpdateCrossBorder(tiled_graph.edges.edges, tiled_graph.edge_types, node_tiles)

	tiles := NewDict[int16, bool](100)
	for _, tile_id := range tiled_graph.node_tiles.GetTiles() {
		if tiles.ContainsKey(tile_id) {
			continue
		}
		tiles[tile_id] = true
	}

	tile_count := tiles.Length()
	is_skip := make([]bool, graph.EdgeCount())
	c := 1
	for tile_id, _ := range tiles {
		fmt.Printf("tile %v: %v / %v \n", tile_id, c, tile_count)
		fmt.Printf("tile %v: getting start nodes \n", tile_id)
		start_nodes := GetStartNodes(tiled_graph, tile_id)
		fmt.Printf("tile %v: calculating skip edges \n", tile_id)
		CalcSkipEdges(tiled_graph, start_nodes, is_skip)
		fmt.Printf("tile %v: finished \n", tile_id)
		c += 1
	}
	UpdateSkipEdges(tiled_graph.edge_types, is_skip)

	return tiled_graph
}

// return list of nodes that have at least one cross-border edge
func GetStartNodes(graph *TiledGraph, tile_id int16) List[int32] {
	list := NewList[int32](100)

	explorer := graph.GetDefaultExplorer()
	for id, tile := range graph.node_tiles.GetTiles() {
		if tile != tile_id {
			continue
		}
		iter := explorer.GetAdjacentEdges(int32(id), BACKWARD)
		for {
			ref, ok := iter.Next()
			if !ok {
				break
			}
			if ref.IsCrossBorder() {
				list.Add(int32(id))
				break
			}
		}
	}
	return list
}

// marks every edge as that lies on a shortest path between border nodes
func CalcSkipEdges(graph *TiledGraph, start_nodes List[int32], is_skip []bool) {
	weight := graph.GetWeighting()

	explorer := graph.GetDefaultExplorer()
	for _, start := range start_nodes {
		heap := NewPriorityQueue[int32, int32](10)
		flags := NewDict[int32, _Flag](10)
		end_nodes := NewDict[int32, bool](100)

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
			iter := explorer.GetAdjacentEdges(curr_id, FORWARD)
			for {
				ref, ok := iter.Next()
				if !ok {
					break
				}
				if ref.IsSkip() {
					continue
				}
				edge_id := ref.EdgeID
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
				weight := weight.GetEdgeWeight(edge_id)
				newlength := curr_flag.pathlength + weight
				if newlength < other_flag.pathlength {
					other_flag.pathlength = newlength
					other_flag.prevEdge = edge_id
					if ref.IsCrossBorder() {
						end_nodes[other_id] = true
					} else {
						heap.Enqueue(other_id, newlength)
					}
				}
				flags[other_id] = other_flag
			}
			flags[curr_id] = curr_flag
		}

		for end, _ := range end_nodes {
			curr_id := end
			for {
				if curr_id == start {
					break
				}
				edge_id := flags[curr_id].prevEdge
				is_skip[edge_id] = true
				curr_id = explorer.GetOtherNode(EdgeRef{EdgeID: edge_id}, curr_id)
			}
		}
	}
}

type _Flag struct {
	pathlength int32
	prevEdge   int32
	visited    bool
}

// sets edge type of cross border edges to 10
func UpdateCrossBorder(edges Array[Edge], edge_types Array[byte], node_tiles List[int16]) {
	for i := 0; i < edges.Length(); i++ {
		edge := edges[i]
		if node_tiles[edge.NodeA] != node_tiles[edge.NodeB] {
			edge_types[i] = 10
		}
	}
}

// sets the edge type of skip edges to 20
func UpdateSkipEdges(edge_types Array[byte], is_skip []bool) {
	for i := 0; i < edge_types.Length(); i++ {
		if edge_types[i] != 10 && is_skip[i] {
			edge_types[i] = 20
		}
	}
}

//*******************************************
// preprocess tiled-graph 2
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
		topology:         graph.topology,
		nodes:            graph.nodes,
		node_tiles:       graph.node_tiles,
		edges:            graph.edges,
		geom:             graph.geom,
		weight:           graph.weight,
		index:            graph.index,
		border_nodes:     border_nodes,
		interior_nodes:   interior_nodes,
		border_range_map: border_range_map,
	}
}

func GetTiles(graph *TiledGraph) List[int16] {
	tile_dict := NewDict[int16, bool](100)
	for _, tile_id := range graph.node_tiles.GetTiles() {
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
	for id, tile := range graph.node_tiles.GetTiles() {
		if tile != tile_id {
			continue
		}
		iter := explorer.GetAdjacentEdges(int32(id), BACKWARD)
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
	weight := graph.GetWeighting()

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
		iter := explorer.GetAdjacentEdges(curr_id, FORWARD)
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
			weight := weight.GetEdgeWeight(edge_id)
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

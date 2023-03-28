package graph

import (
	"fmt"
	"strconv"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// preprocess tiled-graph
//*******************************************

func PreprocessTiledGraph(graph *Graph, features []geo.Feature) *TiledGraph {
	node_tiles := CalcNodeTiles(graph.geom, features)
	edge_refs := UpdateCrossBorder(graph.edges, graph.edge_refs, node_tiles)

	tiled_graph := &TiledGraph{
		nodes:           graph.nodes,
		node_attributes: graph.node_attributes,
		node_tiles:      node_tiles,
		edge_refs:       edge_refs,
		edges:           graph.edges,
		edge_attributes: graph.edge_attributes,
		geom:            graph.geom,
		weight:          graph.weight,
	}

	tiles := NewDict[int16, bool](100)
	for _, tile_id := range tiled_graph.node_tiles {
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
	edge_refs = UpdateSkipEdges(tiled_graph, is_skip)
	tiled_graph.edge_refs = edge_refs

	return tiled_graph
}

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

func UpdateCrossBorder(edges List[Edge], edge_refs List[EdgeRef], node_tiles List[int16]) List[EdgeRef] {
	for i := 0; i < edge_refs.Length(); i++ {
		edge_ref := edge_refs[i]
		if edge_ref.Type != 0 && edge_ref.Type != 1 {
			continue
		}
		edge := edges[edge_ref.EdgeID]
		if node_tiles[edge.NodeA] != node_tiles[edge.NodeB] {
			edge_ref.Type += 10
			edge_refs.Set(i, edge_ref)
		}
	}

	return edge_refs
}

func GetStartNodes(graph *TiledGraph, tile_id int16) List[int32] {
	list := NewList[int32](100)
	for id, tile := range graph.node_tiles {
		if tile != tile_id {
			continue
		}
		iter := graph.GetAdjacentEdges(int32(id))
		for {
			ref, ok := iter.Next()
			if !ok {
				break
			}
			if ref.IsCrossBorder() && ref.IsReversed() {
				list.Add(int32(id))
				break
			}
		}
	}
	return list
}

func CalcSkipEdges(graph *TiledGraph, start_nodes List[int32], is_skip []bool) {
	weight := graph.GetWeighting()

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
			iter := graph.GetAdjacentEdges(curr_id)
			for {
				ref, ok := iter.Next()
				if !ok {
					break
				}
				if ref.IsReversed() || ref.IsSkip() {
					continue
				}
				edge_id := ref.EdgeID
				other_id, _ := graph.GetOtherNode(edge_id, curr_id)
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
				curr_id, _ = graph.GetOtherNode(edge_id, curr_id)
			}
		}
	}
}

type _Flag struct {
	pathlength int32
	prevEdge   int32
	visited    bool
}

func UpdateSkipEdges(graph *TiledGraph, is_skip []bool) List[EdgeRef] {
	edgerefcount := graph.edge_refs.Length()
	for i := 0; i < edgerefcount; i++ {
		edgeref := graph.edge_refs.Get(i)
		if !edgeref.IsCrossBorder() && is_skip[edgeref.EdgeID] {
			edgeref.Type += 20
		}
		graph.edge_refs.Set(i, edgeref)
	}
	return graph.edge_refs
}

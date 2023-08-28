package graph

import (
	"fmt"
	"sort"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func BuildGraphIndex(g *Graph) {
	g.index = _BuildKDTreeIndex(g.geom.GetAllNodes())
}

func _BuildKDTreeIndex(node_geoms List[geo.Coord]) KDTree[int32] {
	tree := NewKDTree[int32](2)
	for i := 0; i < node_geoms.Length(); i++ {
		geom := node_geoms[i]
		tree.Insert(geom[:], int32(i))
	}
	return tree
}

func GraphToGeoJSON(graph *TiledGraph) (geo.FeatureCollection, geo.FeatureCollection) {
	geom := graph.GetGeometry()

	edges := NewList[geo.Feature](int(graph.EdgeCount()))
	for i := 0; i < graph.edges.EdgeCount(); i++ {
		edge := graph.GetEdge(int32(i))
		if i%1000 == 0 {
			fmt.Println("edge ", i)
		}
		var edge_type int16
		if graph.GetNodeTile(edge.NodeA) == graph.GetNodeTile(edge.NodeB) {
			edge_type = graph.GetNodeTile(edge.NodeA)
		} else {
			edge_type = -10
		}
		line := geo.NewLineString(geom.GetEdge(int32(i)))
		edges.Add(geo.NewFeature(&line, map[string]any{"index": i, "nodeA": edge.NodeA, "nodeB": edge.NodeB, "type": edge_type}))
	}

	nodes := NewList[geo.Feature](int(graph.NodeCount()))
	for i, node := range graph.topology.node_entries {
		if i%1000 == 0 {
			fmt.Println("node ", i)
		}
		e := NewList[int32](3)
		for j := 0; j < int(node.FWDEdgeCount); j++ {
			e.Add(graph.topology.fwd_edge_entries[int(node.FWDEdgeStart)+j].EdgeID)
		}
		node_tile := graph.node_tiles.GetNodeTile(int32(i))
		point := geo.NewPoint(geom.GetNode(int32(i)))
		nodes.Add(geo.NewFeature(&point, map[string]any{"index": i, "edgecount": node.FWDEdgeCount, "edges": e, "tile": node_tile}))
	}

	return geo.NewFeatureCollection(nodes), geo.NewFeatureCollection(edges)
}

func GraphToGeoJSON2(graph *Graph, node_tiles Array[int16]) (geo.FeatureCollection, geo.FeatureCollection) {
	geom := graph.GetGeometry()

	edges := NewList[geo.Feature](int(graph.EdgeCount()))
	for i := 0; i < graph.edges.EdgeCount(); i++ {
		edge := graph.GetEdge(int32(i))
		if i%1000 == 0 {
			fmt.Println("edge ", i)
		}
		var edge_type int16
		if node_tiles[edge.NodeA] == node_tiles[edge.NodeB] {
			edge_type = node_tiles[edge.NodeA]
		} else {
			edge_type = -10
		}
		line := geo.NewLineString(geom.GetEdge(int32(i)))
		edges.Add(geo.NewFeature(&line, map[string]any{"index": i, "nodeA": edge.NodeA, "nodeB": edge.NodeB, "type": edge_type}))
	}

	nodes := NewList[geo.Feature](int(graph.NodeCount()))
	for i, node := range graph.topology.node_entries {
		if i%1000 == 0 {
			fmt.Println("node ", i)
		}
		e := NewList[int32](3)
		for j := 0; j < int(node.FWDEdgeCount); j++ {
			e.Add(graph.topology.fwd_edge_entries[int(node.FWDEdgeStart)+j].EdgeID)
		}
		node_tile := node_tiles[i]
		point := geo.NewPoint(geom.GetNode(int32(i)))
		nodes.Add(geo.NewFeature(&point, map[string]any{"index": i, "edgecount": node.FWDEdgeCount, "edges": e, "tile": node_tile}))
	}

	return geo.NewFeatureCollection(nodes), geo.NewFeatureCollection(edges)
}

// checks graph topology
func CheckGraph(g IGraph) {
	explorer := g.GetDefaultExplorer()
	for i := 0; i < int(g.NodeCount()); i++ {
		adj_edges := explorer.GetAdjacentEdges(int32(i), FORWARD, ADJACENT_ALL)
		for {
			ref, ok := adj_edges.Next()
			if !ok {
				break
			}
			if ref.IsShortcut() {
				continue
			}
			edge := g.GetEdge(ref.EdgeID)
			if edge.NodeA != int32(i) {
				fmt.Println("error 81")
			}
			if edge.NodeB != ref.OtherID {
				fmt.Println("error 84")
			}
		}
		adj_edges = explorer.GetAdjacentEdges(int32(i), BACKWARD, ADJACENT_ALL)
		for {
			ref, ok := adj_edges.Next()
			if !ok {
				break
			}
			if ref.IsShortcut() {
				continue
			}
			edge := g.GetEdge(ref.EdgeID)
			if edge.NodeB != int32(i) {
				fmt.Println("error 95")
			}
			if edge.NodeA != ref.OtherID {
				fmt.Println("error 98")
			}
		}
	}
}

// checks topology of ch graph
func CheckCHGraph(g ICHGraph) {
	explorer := g.GetDefaultExplorer()
	for i := 0; i < int(g.NodeCount()); i++ {
		adj_edges := explorer.GetAdjacentEdges(int32(i), FORWARD, ADJACENT_ALL)
		for {
			ref, ok := adj_edges.Next()
			if !ok {
				break
			}
			if ref.IsShortcut() {
				edge := g.GetShortcut(ref.EdgeID)
				if edge.NodeA != int32(i) {
					fmt.Println("error 1")
				}
				if edge.NodeB != ref.OtherID {
					fmt.Println("error 2")
				}
			} else {
				edge := g.GetEdge(ref.EdgeID)
				if edge.NodeA != int32(i) {
					fmt.Println("error 3")
				}
				if edge.NodeB != ref.OtherID {
					fmt.Println("error 4")
				}
			}
		}
		adj_edges = explorer.GetAdjacentEdges(int32(i), BACKWARD, ADJACENT_ALL)
		for {
			ref, ok := adj_edges.Next()
			if !ok {
				break
			}
			if ref.IsShortcut() {
				edge := g.GetShortcut(ref.EdgeID)
				if edge.NodeB != int32(i) {
					fmt.Println("error 5")
				}
				if edge.NodeA != ref.OtherID {
					fmt.Println("error 6")
				}
			} else {
				edge := g.GetEdge(ref.EdgeID)
				if edge.NodeB != int32(i) {
					fmt.Println("error 7")
				}
				if edge.NodeA != ref.OtherID {
					fmt.Println("error 8")
				}
			}
		}
	}
}

func SortNodesByLevel(g *CHGraph) {
	indices := NewList[Tuple[int32, int16]](int(g.NodeCount()))
	for i := 0; i < int(g.NodeCount()); i++ {
		indices.Add(MakeTuple(int32(i), g.node_levels.GetNodeLevel(int32(i))))
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

func ReorderCHGraph(g *CHGraph, node_mapping Array[int32]) {
	g.nodes._ReorderNodes(node_mapping)
	g.edges._ReorderNodes(node_mapping)
	g.node_levels._ReorderNodes(node_mapping)
	g.shortcuts._ReorderNodes(node_mapping)
	g.topology._ReorderNodes(node_mapping)
	g.ch_topology._ReorderNodes(node_mapping)
	g.geom._ReorderNodes(node_mapping)
	g.weight._ReorderNodes(node_mapping)
	g.sh_weight._ReorderNodes(node_mapping)
	g.index = _BuildKDTreeIndex(g.geom.GetAllNodes())
}

package graph

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	. "github.com/ttpr0/go-routing/util"
)

// func GraphToGeoJSON(graph *TiledGraph) (geo.FeatureCollection, geo.FeatureCollection) {
// 	edges := NewList[geo.Feature](int(graph.EdgeCount()))
// 	for i := 0; i < graph.EdgeCount(); i++ {
// 		edge := graph.GetEdge(int32(i))
// 		if i%1000 == 0 {
// 			fmt.Println("edge ", i)
// 		}
// 		var edge_type int16
// 		if graph.GetNodeTile(edge.NodeA) == graph.GetNodeTile(edge.NodeB) {
// 			edge_type = graph.GetNodeTile(edge.NodeA)
// 		} else {
// 			edge_type = -10
// 		}
// 		line := geo.NewLineString(graph.GetEdgeGeom(int32(i)))
// 		edges.Add(geo.NewFeature(&line, map[string]any{"index": i, "nodeA": edge.NodeA, "nodeB": edge.NodeB, "type": edge_type}))
// 	}

// 	nodes := NewList[geo.Feature](int(graph.NodeCount()))
// 	for i, node := range graph.base.topology.node_entries {
// 		if i%1000 == 0 {
// 			fmt.Println("node ", i)
// 		}
// 		e := NewList[int32](3)
// 		for j := 0; j < int(node.FWDEdgeCount); j++ {
// 			e.Add(graph.base.topology.fwd_edge_entries[int(node.FWDEdgeStart)+j].EdgeID)
// 		}
// 		node_tile := graph.GetNodeTile(int32(i))
// 		point := geo.NewPoint(graph.GetNodeGeom(int32(i)))
// 		nodes.Add(geo.NewFeature(&point, map[string]any{"index": i, "edgecount": node.FWDEdgeCount, "edges": e, "tile": node_tile}))
// 	}

// 	return geo.NewFeatureCollection(nodes), geo.NewFeatureCollection(edges)
// }

// func GraphToGeoJSON2(graph *Graph, node_tiles Array[int16]) (geo.FeatureCollection, geo.FeatureCollection) {
// 	edges := NewList[geo.Feature](int(graph.EdgeCount()))
// 	for i := 0; i < graph.EdgeCount(); i++ {
// 		edge := graph.GetEdge(int32(i))
// 		if i%1000 == 0 {
// 			fmt.Println("edge ", i)
// 		}
// 		var edge_type int16
// 		if node_tiles[edge.NodeA] == node_tiles[edge.NodeB] {
// 			edge_type = node_tiles[edge.NodeA]
// 		} else {
// 			edge_type = -10
// 		}
// 		line := geo.NewLineString(graph.GetEdgeGeom(int32(i)))
// 		edges.Add(geo.NewFeature(&line, map[string]any{"index": i, "nodeA": edge.NodeA, "nodeB": edge.NodeB, "type": edge_type}))
// 	}

// 	nodes := NewList[geo.Feature](int(graph.NodeCount()))
// 	for i, node := range graph.base.topology.node_entries {
// 		if i%1000 == 0 {
// 			fmt.Println("node ", i)
// 		}
// 		e := NewList[int32](3)
// 		for j := 0; j < int(node.FWDEdgeCount); j++ {
// 			e.Add(graph.base.topology.fwd_edge_entries[int(node.FWDEdgeStart)+j].EdgeID)
// 		}
// 		node_tile := node_tiles[i]
// 		point := geo.NewPoint(graph.GetNodeGeom(int32(i)))
// 		nodes.Add(geo.NewFeature(&point, map[string]any{"index": i, "edgecount": node.FWDEdgeCount, "edges": e, "tile": node_tile}))
// 	}

// 	return geo.NewFeatureCollection(nodes), geo.NewFeatureCollection(edges)
// }

func GraphToMETIS(g IGraph) string {
	n := g.NodeCount()
	m := 0
	adj := NewArray[List[int32]](n)

	for i := 0; i < g.EdgeCount(); i++ {
		edge := g.GetEdge(int32(i))

		adj_a := adj[edge.NodeA]
		if !Contains(adj_a, edge.NodeB+1) {
			adj_a.Add(edge.NodeB + 1)
			m += 1
		}
		adj[edge.NodeA] = adj_a

		adj_b := adj[edge.NodeB]
		if !Contains(adj_b, edge.NodeA+1) {
			adj_b.Add(edge.NodeA + 1)
			m += 1
		}
		adj[edge.NodeB] = adj_b
	}
	m = m / 2

	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintln(n, m))
	for i := 0; i < adj.Length(); i++ {
		adj_nodes := adj[i]
		for _, node := range adj_nodes {
			builder.WriteString(fmt.Sprint(node, " "))
		}
		builder.WriteString("\n")
	}
	return builder.String()
}

func StoreNodeTiles(filename string, node_tiles Array[int16]) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("failed to create tile-file")
		return
	}
	defer file.Close()

	var builder strings.Builder
	for i := 0; i < node_tiles.Length(); i++ {
		builder.WriteString(fmt.Sprintln(node_tiles[i]))
	}
	file.Write([]byte(builder.String()))
}
func ReadNodeTiles(filename string) Array[int16] {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("failed to open csv file")
		return nil
	}
	defer file.Close()
	stat, _ := file.Stat()
	data := make([]byte, stat.Size())
	file.Read(data)
	s := string(data)
	tokens := strings.Split(s, "\n")

	tiles := NewArray[int16](len(tokens))
	for i := 0; i < tiles.Length(); i++ {
		val, _ := strconv.Atoi(tokens[i])
		tiles[i] = int16(val)
	}
	return tiles
}

func StoreNodeOrdering(filename string, contraction_order Array[int32]) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("failed to create csv file")
		return
	}
	defer file.Close()

	var builder strings.Builder
	builder.WriteString(fmt.Sprintln(contraction_order.Length()))
	for i := 0; i < contraction_order.Length()-1; i++ {
		builder.WriteString(fmt.Sprint(contraction_order[i]) + ",")
	}
	builder.WriteString(fmt.Sprint(contraction_order[contraction_order.Length()-1]))
	file.Write([]byte(builder.String()))
}
func ReadNodeOrdering(filename string) Array[int32] {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("failed to open csv file")
		return nil
	}
	defer file.Close()
	stat, _ := file.Stat()
	data := make([]byte, stat.Size())
	file.Read(data)
	s := string(data)
	tokens := strings.Split(s, "\n")

	num_nodes, _ := strconv.Atoi(tokens[0])
	ordering := NewArray[int32](num_nodes)

	for i := 1; i <= ordering.Length(); i++ {
		fields := strings.Fields(tokens[i])
		val1, _ := strconv.Atoi(fields[0])
		val2, _ := strconv.Atoi(fields[1])
		ordering[val2-1] = int32(val1 - 1)
	}
	return ordering
}

func _IsBorderNode3(graph IGraph, partition *Partition) Array[bool] {
	is_border := NewArray[bool](graph.NodeCount())

	explorer := graph.GetGraphExplorer()
	for i := 0; i < graph.NodeCount(); i++ {
		explorer.ForAdjacentEdges(int32(i), FORWARD, ADJACENT_ALL, func(ref EdgeRef) {
			if partition.GetNodeTile(int32(i)) != partition.GetNodeTile(ref.OtherID) {
				is_border[i] = true
			}
		})
		explorer.ForAdjacentEdges(int32(i), BACKWARD, ADJACENT_ALL, func(ref EdgeRef) {
			if partition.GetNodeTile(int32(i)) != partition.GetNodeTile(ref.OtherID) {
				is_border[i] = true
			}
		})
	}

	return is_border
}

// Reorders Array based on mapping (old-id -> new-id).
func Reorder[T any](arr Array[T], mapping Array[int32]) Array[T] {
	new_arr := NewArray[T](arr.Length())
	for i, id := range mapping {
		new_arr[id] = arr[i]
	}
	return new_arr
}

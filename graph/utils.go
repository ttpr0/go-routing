package graph

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func LoadOrCreate(graph_path string, osm_file string, partition_file string) ITiledGraph {
	// check if graph files already exist
	_, err1 := os.Stat(graph_path + "-nodes")
	_, err2 := os.Stat(graph_path + "-edges")
	_, err3 := os.Stat(graph_path + "-geom")
	_, err4 := os.Stat(graph_path + "-tiles")
	if errors.Is(err1, os.ErrNotExist) || errors.Is(err2, os.ErrNotExist) || errors.Is(err3, os.ErrNotExist) || errors.Is(err4, os.ErrNotExist) {
		// create graph
		g := ParseGraph(osm_file)

		file_str, _ := os.ReadFile(partition_file)
		collection := geo.FeatureCollection{}
		_ = json.Unmarshal(file_str, &collection)
		g.index = _BuildNodeIndex(g.geom.GetAllNodes())

		tg := PreprocessTiledGraph(g, collection.Features())

		StoreTiledGraph(tg, graph_path)

		return tg
	} else {
		return LoadTiledGraph(graph_path)
	}
}

func GraphToGeoJSON(graph *TiledGraph) (geo.FeatureCollection, geo.FeatureCollection) {
	geom := graph.GetGeometry()
	edge_types := make([]int16, int(graph.EdgeCount()))
	for i := 0; i < graph.fwd_edge_refs.Length(); i++ {
		edge_ref := graph.fwd_edge_refs[i]
		edge_types[edge_ref.EdgeID] = int16(edge_ref.Type)
	}

	edges := NewList[geo.Feature](int(graph.EdgeCount()))
	for i, edge := range graph.edges {
		if i%1000 == 0 {
			fmt.Println("edge ", i)
		}
		line := geo.NewLineString(geom.GetEdge(int32(i)))
		edges.Add(geo.NewFeature(&line, map[string]any{"index": i, "nodeA": edge.NodeA, "nodeB": edge.NodeB, "type": edge_types[i]}))
	}

	nodes := NewList[geo.Feature](int(graph.NodeCount()))
	for i, node := range graph.node_refs {
		if i%1000 == 0 {
			fmt.Println("node ", i)
		}
		e := NewList[int32](3)
		for j := 0; j < int(node.EdgeRefFWDCount); j++ {
			e.Add(graph.fwd_edge_refs[int(node.EdgeRefFWDStart)+j].EdgeID)
		}
		point := geo.NewPoint(geom.GetNode(int32(i)))
		nodes.Add(geo.NewFeature(&point, map[string]any{"index": i, "edgecount": node.EdgeRefFWDCount, "edges": e, "tile": graph.node_tiles[i]}))
	}

	return geo.NewFeatureCollection(nodes), geo.NewFeatureCollection(edges)
}

func CheckGraph(g IGraph) {
	for i := 0; i < int(g.NodeCount()); i++ {
		adj_edges := g.GetAdjacentEdges(int32(i), FORWARD)
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
			if edge.NodeB != ref.NodeID {
				fmt.Println("error 84")
			}
		}
		adj_edges = g.GetAdjacentEdges(int32(i), BACKWARD)
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
			if edge.NodeA != ref.NodeID {
				fmt.Println("error 98")
			}
		}
	}
}

func SortNodesByLevel(g *CHGraph) {
	indices := NewList[Tuple[int32, int16]](int(g.NodeCount()))
	for i := 0; i < int(g.NodeCount()); i++ {
		indices.Add(MakeTuple(int32(i), g.node_levels[i]))
	}
	sort.SliceStable(indices, func(i, j int) bool {
		return indices[i].B > indices[j].B
	})
	rev_indices := NewArray[int32](int(g.NodeCount()))
	for i := 0; i < int(g.NodeCount()); i++ {
		rev_indices[indices[i].A] = int32(i)
	}

	node_refs := NewList[NodeRef](g.node_refs.Length())
	nodes := NewList[Node](g.nodes.Length())
	node_levels := NewList[int16](g.node_levels.Length())
	fwd_edge_refs := NewList[EdgeRef](g.fwd_edge_refs.Length())
	bwd_edge_refs := NewList[EdgeRef](g.bwd_edge_refs.Length())
	node_geom := NewList[geo.Coord](g.nodes.Length())
	edges := NewList[Edge](g.edges.Length())
	shortcuts := NewList[Shortcut](g.shortcuts.Length())

	fwd_start := 0
	bwd_start := 0
	for i := 0; i < g.nodes.Length(); i++ {
		index := indices[i].A

		fwd_count := 0
		fwd_edges := g.GetAdjacentEdges(index, FORWARD)
		for {
			ref, ok := fwd_edges.Next()
			if !ok {
				break
			}
			ref.NodeID = rev_indices[ref.NodeID]
			fwd_edge_refs.Add(ref)
			fwd_count += 1
		}

		bwd_count := 0
		bwd_edges := g.GetAdjacentEdges(index, BACKWARD)
		for {
			ref, ok := bwd_edges.Next()
			if !ok {
				break
			}
			ref.NodeID = rev_indices[ref.NodeID]
			bwd_edge_refs.Add(ref)
			bwd_count += 1
		}

		node_levels.Add(g.node_levels[index])
		node_geom.Add(g.geom.GetNode(index))
		nodes.Add(g.nodes[index])
		node_refs.Add(NodeRef{
			EdgeRefFWDStart: int32(fwd_start),
			EdgeRefFWDCount: int16(fwd_count),
			EdgeRefBWDStart: int32(bwd_start),
			EdgeRefBWDCount: int16(bwd_count),
		})

		fwd_start += fwd_count
		bwd_start += bwd_count
	}
	for _, edge := range g.edges {
		edge.NodeA = rev_indices[edge.NodeA]
		edge.NodeB = rev_indices[edge.NodeB]
		edges.Add(edge)
	}
	for _, shc := range g.shortcuts {
		shc.NodeA = rev_indices[shc.NodeA]
		shc.NodeB = rev_indices[shc.NodeB]
		shortcuts.Add(shc)
	}

	g.node_refs = node_refs
	g.nodes = nodes
	g.node_levels = node_levels
	g.fwd_edge_refs = fwd_edge_refs
	g.bwd_edge_refs = bwd_edge_refs
	g.edges = edges
	g.shortcuts = shortcuts
	geom := g.geom.(*Geometry)
	geom.NodeGeometry = node_geom
	g.geom = geom
}

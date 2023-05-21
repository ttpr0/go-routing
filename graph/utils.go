package graph

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

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

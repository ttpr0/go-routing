package graph

import (
	"fmt"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func GraphToGeoJSON(graph *TiledGraph) (geo.FeatureCollection, geo.FeatureCollection) {
	geom := graph.GetGeometry()
	edge_types := make([]int16, int(graph.EdgeCount()))
	for i := 0; i < graph.edge_refs.Length(); i++ {
		edge_ref := graph.edge_refs[i]
		if edge_ref.IsReversed() {
			continue
		}
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
	for i, node := range graph.nodes {
		if i%1000 == 0 {
			fmt.Println("node ", i)
		}
		e := NewList[int32](3)
		for j := 0; j < int(node.EdgeRefCount); j++ {
			e.Add(graph.edge_refs[int(node.EdgeRefStart)+j].EdgeID)
		}
		point := geo.NewPoint(geom.GetNode(int32(i)))
		nodes.Add(geo.NewFeature(&point, map[string]any{"index": i, "edgecount": node.EdgeRefCount, "edges": e, "tile": graph.node_tiles[i]}))
	}

	return geo.NewFeatureCollection(nodes), geo.NewFeatureCollection(edges)
}

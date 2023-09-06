package graph

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// modifikation methods
//*******************************************

func RemoveNodes(graph *Graph, nodes List[int32]) *Graph {
	store := graph.store

	remove := NewArray[bool](store.NodeCount())
	for _, n := range nodes {
		remove[n] = true
	}

	new_nodes := NewList[Node](100)
	new_node_geoms := NewList[geo.Coord](100)
	mapping := NewArray[int32](store.NodeCount())
	id := int32(0)
	for i := 0; i < store.NodeCount(); i++ {
		if remove[i] {
			mapping[i] = -1
			continue
		}
		new_nodes.Add(store.GetNode(int32(i)))
		new_node_geoms.Add(store.GetNodeGeom(int32(i)))
		mapping[i] = id
		id += 1
	}
	new_edges := NewList[Edge](100)
	new_edge_geoms := NewList[geo.CoordArray](100)
	for i := 0; i < store.EdgeCount(); i++ {
		edge := store.GetEdge(int32(i))
		if remove[edge.NodeA] || remove[edge.NodeB] {
			continue
		}
		new_edges.Add(Edge{
			NodeA:    mapping[edge.NodeA],
			NodeB:    mapping[edge.NodeB],
			Type:     edge.Type,
			Length:   edge.Length,
			Maxspeed: edge.Maxspeed,
			Oneway:   edge.Oneway,
		})
		new_edge_geoms.Add(store.GetEdgeGeom(int32(i)))
	}

	new_store := GraphStore{
		nodes:      Array[Node](new_nodes),
		edges:      Array[Edge](new_edges),
		node_geoms: new_node_geoms,
		edge_geoms: new_edge_geoms,
	}
	return &Graph{
		store:    new_store,
		topology: _BuildTopology(new_store),
		weight:   _BuildWeighting(new_store),
		index:    _BuildKDTreeIndex(new_store),
	}
}

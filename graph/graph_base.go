package graph

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// graph base
//******************************************

type GraphBase struct {
	store    GraphStore
	topology AdjacencyArray
	index    KDTree[int32]
}

func (self *GraphBase) NodeCount() int {
	return self.store.NodeCount()
}
func (self *GraphBase) EdgeCount() int {
	return self.store.EdgeCount()
}
func (self *GraphBase) GetNode(node int32) Node {
	return self.store.GetNode(node)
}
func (self *GraphBase) GetEdge(edge int32) Edge {
	return self.store.GetEdge(edge)
}
func (self *GraphBase) GetNodeGeom(node int32) geo.Coord {
	return self.store.GetNodeGeom(node)
}
func (self *GraphBase) GetEdgeGeom(edge int32) geo.CoordArray {
	return self.store.GetEdgeGeom(edge)
}
func (self *GraphBase) GetAccessor() AdjArrayAccessor {
	accessor := self.topology.GetAccessor()
	return accessor
}
func (self *GraphBase) GetKDTree() KDTree[int32] {
	return self.index
}
func (self *GraphBase) _ReorderNodes(mapping Array[int32]) {
	self.store._ReorderNodes(mapping)
	self.topology._ReorderNodes(mapping)
	self.index = _BuildKDTreeIndex(self.store)
}
func (self *GraphBase) _RemoveNodes(nodes List[int32]) {
	store := self.store

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

	self.store = GraphStore{
		nodes:      Array[Node](new_nodes),
		edges:      Array[Edge](new_edges),
		node_geoms: new_node_geoms,
		edge_geoms: new_edge_geoms,
	}
	self.topology = _BuildTopology(self.store)
	self.index = _BuildKDTreeIndex(self.store)
}

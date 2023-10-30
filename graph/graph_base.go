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

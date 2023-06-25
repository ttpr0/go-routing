package graph

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type IGraph interface {
	GetGeometry() IGeometry
	GetWeighting() IWeighting
	GetOtherNode(edge, node int32) (int32, Direction)
	GetAdjacentEdges(node int32, direction Direction) IIterator[EdgeRef]
	ForEachEdge(node int32, f func(int32))
	NodeCount() int32
	EdgeCount() int32
	IsNode(node int32) bool
	GetNode(node int32) Node
	GetEdge(edge int32) Edge
	GetNodeIndex() KDTree[int32]
	GetClosestNode(point geo.Coord) (int32, bool)
}

type Graph struct {
	node_refs     List[NodeRef]
	nodes         List[Node]
	fwd_edge_refs List[EdgeRef]
	bwd_edge_refs List[EdgeRef]
	edges         List[Edge]
	geom          IGeometry
	weight        IWeighting
	index         KDTree[int32]
}

func (self *Graph) GetGeometry() IGeometry {
	return self.geom
}
func (self *Graph) GetWeighting() IWeighting {
	return self.weight
}
func (self *Graph) GetOtherNode(edge, node int32) (int32, Direction) {
	e := self.edges[edge]
	if node == e.NodeA {
		return e.NodeB, FORWARD
	}
	if node == e.NodeB {
		return e.NodeA, BACKWARD
	}
	return -1, 0
}
func (self *Graph) GetAdjacentEdges(node int32, direction Direction) IIterator[EdgeRef] {
	n := self.node_refs[node]
	if direction == FORWARD {
		return &EdgeRefIterator{
			state:     int(n.EdgeRefFWDStart),
			end:       int(n.EdgeRefFWDStart) + int(n.EdgeRefFWDCount),
			edge_refs: &self.fwd_edge_refs,
		}
	} else {
		return &EdgeRefIterator{
			state:     int(n.EdgeRefBWDStart),
			end:       int(n.EdgeRefBWDStart) + int(n.EdgeRefBWDCount),
			edge_refs: &self.bwd_edge_refs,
		}
	}
}
func (self *Graph) ForEachEdge(node int32, f func(int32)) {

}
func (self *Graph) NodeCount() int32 {
	return int32(len(self.nodes))
}
func (self *Graph) EdgeCount() int32 {
	return int32(len(self.edges))
}
func (self *Graph) IsNode(node int32) bool {
	if node < int32(len(self.nodes)) {
		return true
	} else {
		return false
	}
}
func (self *Graph) GetNode(node int32) Node {
	return self.nodes[node]
}
func (self *Graph) GetEdge(edge int32) Edge {
	return self.edges[edge]
}
func (self *Graph) GetNodeIndex() KDTree[int32] {
	return self.index
}
func (self *Graph) GetClosestNode(point geo.Coord) (int32, bool) {
	return self.index.GetClosest(point[:], 0.005)
}

type EdgeRefIterator struct {
	state     int
	end       int
	edge_refs *List[EdgeRef]
}

func (self *EdgeRefIterator) Next() (EdgeRef, bool) {
	if self.state == self.end {
		var t EdgeRef
		return t, false
	}
	self.state += 1
	return self.edge_refs.Get(self.state - 1), true
}

package graph

import (
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type IGraph interface {
	GetGeometry() IGeometry
	GetWeighting() IWeighting
	GetOtherNode(edge, node int32) (int32, Direction)
	GetAdjacentEdges(node int32) IIterator[EdgeRef]
	ForEachEdge(node int32, f func(int32))
	NodeCount() int32
	EdgeCount() int32
	IsNode(node int32) bool
	GetNode(node int32) NodeAttributes
	GetEdge(edge int32) EdgeAttributes
}

type Graph struct {
	nodes           List[Node]
	node_attributes List[NodeAttributes]
	edge_refs       List[EdgeRef]
	edges           List[Edge]
	edge_attributes List[EdgeAttributes]
	geom            IGeometry
	weight          IWeighting
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
func (self *Graph) GetAdjacentEdges(node int32) IIterator[EdgeRef] {
	n := self.nodes[node]
	return &EdgeRefIterator{
		state:     int(n.EdgeRefStart),
		end:       int(n.EdgeRefStart) + int(n.EdgeRefCount),
		edge_refs: &self.edge_refs,
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
func (self *Graph) GetNode(node int32) NodeAttributes {
	return self.node_attributes[node]
}
func (self *Graph) GetEdge(edge int32) EdgeAttributes {
	return self.edge_attributes[edge]
}

type EdgeRefIterator struct {
	state     int
	end       int
	edge_refs *List[EdgeRef]
}

func (self *EdgeRefIterator) Next() (EdgeRef, bool) {
	for {
	if self.state == self.end {
		var t EdgeRef
		return t, false
	}
	self.state += 1
		if self.edge_refs.Get(self.state-1).Type <= 1 {
	return self.edge_refs.Get(self.state - 1), true
		}
	}
}

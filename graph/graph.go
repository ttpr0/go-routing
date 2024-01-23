package graph

import (
	"github.com/ttpr0/go-routing/geo"
)

//*******************************************
// graph interfaces
//******************************************

type IGraph interface {
	GetGraphExplorer() IGraphExplorer
	GetIndex() IGraphIndex
	NodeCount() int
	EdgeCount() int
	IsNode(node int32) bool
	GetNode(node int32) Node
	GetEdge(edge int32) Edge
	GetNodeGeom(node int32) geo.Coord
	GetEdgeGeom(edge int32) geo.CoordArray
}

// not thread safe, use only one instance per thread
type IGraphExplorer interface {
	// Iterates through the adjacency of a node calling the callback for every edge.
	//
	// direction tells the traversel direction (FORWARD meand outgoing edges, BACKWARD ingoing edges)
	//
	// typ is basically a hint to tell which edges/sub-graph will be traversed
	ForAdjacentEdges(node int32, dir Direction, typ Adjacency, callback func(EdgeRef))
	GetEdgeWeight(edge EdgeRef) int32
	GetTurnCost(from EdgeRef, via int32, to EdgeRef) int32
	GetOtherNode(edge EdgeRef, node int32) int32
}

//*******************************************
// base-graph
//******************************************

type Graph struct {
	base   IGraphBase
	weight IWeighting
	index  IGraphIndex
}

func (self *Graph) GetGraphExplorer() IGraphExplorer {
	return &BaseGraphExplorer{
		graph:    self,
		accessor: self.base.GetAccessor(),
		weight:   self.weight,
	}
}
func (self *Graph) NodeCount() int {
	return self.base.NodeCount()
}
func (self *Graph) EdgeCount() int {
	return self.base.EdgeCount()
}
func (self *Graph) IsNode(node int32) bool {
	return int32(self.base.NodeCount()) < node
}
func (self *Graph) GetNode(node int32) Node {
	return self.base.GetNode(node)
}
func (self *Graph) GetEdge(edge int32) Edge {
	return self.base.GetEdge(edge)
}
func (self *Graph) GetNodeGeom(node int32) geo.Coord {
	return self.base.GetNodeGeom(node)
}
func (self *Graph) GetEdgeGeom(edge int32) geo.CoordArray {
	return self.base.GetEdgeGeom(edge)
}
func (self *Graph) GetIndex() IGraphIndex {
	return self.index
}

//*******************************************
// base-graph explorer
//******************************************

type BaseGraphExplorer struct {
	graph    *Graph
	accessor _AdjArrayAccessor
	weight   IWeighting
}

func (self *BaseGraphExplorer) ForAdjacentEdges(node int32, direction Direction, typ Adjacency, callback func(EdgeRef)) {
	if typ == ADJACENT_ALL || typ == ADJACENT_EDGES {
		self.accessor.SetBaseNode(node, direction)
		for self.accessor.Next() {
			edge_id := self.accessor.GetEdgeID()
			other_id := self.accessor.GetOtherID()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				_Type:   0,
			})
		}
	} else {
		panic("Adjacency-type not implemented for this graph.")
	}
}
func (self *BaseGraphExplorer) GetEdgeWeight(edge EdgeRef) int32 {
	return self.weight.GetEdgeWeight(edge.EdgeID)
}
func (self *BaseGraphExplorer) GetTurnCost(from EdgeRef, via int32, to EdgeRef) int32 {
	return self.weight.GetTurnCost(from.EdgeID, via, to.EdgeID)
}
func (self *BaseGraphExplorer) GetOtherNode(edge EdgeRef, node int32) int32 {
	e := self.graph.GetEdge(edge.EdgeID)
	if node == e.NodeA {
		return e.NodeB
	}
	if node == e.NodeB {
		return e.NodeA
	}
	return -1
}

package graph

import (
	"github.com/ttpr0/go-routing/comps"
	"github.com/ttpr0/go-routing/geo"
	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// graph interfaces
//******************************************

type IGraph interface {
	GetGraphExplorer() IGraphExplorer
	NodeCount() int
	EdgeCount() int
	IsNode(node int32) bool
	GetNode(node int32) structs.Node
	GetEdge(edge int32) structs.Edge
	GetNodeGeom(node int32) geo.Coord
	GetClosestNode(point geo.Coord) (int32, bool)
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
	base   comps.IGraphBase
	weight comps.IWeighting
	index  Optional[comps.IGraphIndex]
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
func (self *Graph) GetNode(node int32) structs.Node {
	return self.base.GetNode(node)
}
func (self *Graph) GetEdge(edge int32) structs.Edge {
	return self.base.GetEdge(edge)
}
func (self *Graph) GetNodeGeom(node int32) geo.Coord {
	return self.base.GetNode(node).Loc
}
func (self *Graph) GetClosestNode(point geo.Coord) (int32, bool) {
	if self.index.HasValue() {
		return self.index.Value.GetClosestNode(point)
	} else {
		self.index.Value = comps.NewGraphIndex(self.base)
		return self.index.Value.GetClosestNode(point)
	}
}

//*******************************************
// base-graph explorer
//******************************************

type BaseGraphExplorer struct {
	graph    *Graph
	accessor structs.IAdjAccessor
	weight   comps.IWeighting
}

func (self *BaseGraphExplorer) ForAdjacentEdges(node int32, direction Direction, typ Adjacency, callback func(EdgeRef)) {
	if typ == ADJACENT_ALL || typ == ADJACENT_EDGES {
		self.accessor.SetBaseNode(node, direction == FORWARD)
		for self.accessor.Next() {
			edge_id := self.accessor.GetEdgeID()
			other_id := self.accessor.GetOtherID()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				Type:    0,
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
	return 0
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

//*******************************************
// tc-graph
//******************************************

type TCGraph struct {
	base   comps.IGraphBase
	weight comps.ITCWeighting
	index  Optional[comps.IGraphIndex]
}

func (self *TCGraph) GetGraphExplorer() IGraphExplorer {
	return &TCGraphExplorer{
		graph:    self,
		accessor: self.base.GetAccessor(),
		weight:   self.weight,
	}
}
func (self *TCGraph) NodeCount() int {
	return self.base.NodeCount()
}
func (self *TCGraph) EdgeCount() int {
	return self.base.EdgeCount()
}
func (self *TCGraph) IsNode(node int32) bool {
	return int32(self.base.NodeCount()) < node
}
func (self *TCGraph) GetNode(node int32) structs.Node {
	return self.base.GetNode(node)
}
func (self *TCGraph) GetEdge(edge int32) structs.Edge {
	return self.base.GetEdge(edge)
}
func (self *TCGraph) GetNodeGeom(node int32) geo.Coord {
	return self.base.GetNode(node).Loc
}
func (self *TCGraph) GetClosestNode(point geo.Coord) (int32, bool) {
	if self.index.HasValue() {
		return self.index.Value.GetClosestNode(point)
	} else {
		self.index.Value = comps.NewGraphIndex(self.base)
		return self.index.Value.GetClosestNode(point)
	}
}

//*******************************************
// tc-graph explorer
//******************************************

type TCGraphExplorer struct {
	graph    IGraph
	accessor structs.IAdjAccessor
	weight   comps.ITCWeighting
}

func (self *TCGraphExplorer) ForAdjacentEdges(node int32, direction Direction, typ Adjacency, callback func(EdgeRef)) {
	if typ == ADJACENT_ALL || typ == ADJACENT_EDGES {
		self.accessor.SetBaseNode(node, direction == FORWARD)
		for self.accessor.Next() {
			edge_id := self.accessor.GetEdgeID()
			other_id := self.accessor.GetOtherID()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				Type:    0,
			})
		}
	} else {
		panic("Adjacency-type not implemented for this graph.")
	}
}
func (self *TCGraphExplorer) GetEdgeWeight(edge EdgeRef) int32 {
	return self.weight.GetEdgeWeight(edge.EdgeID)
}
func (self *TCGraphExplorer) GetTurnCost(from EdgeRef, via int32, to EdgeRef) int32 {
	return self.weight.GetTurnCost(from.EdgeID, via, to.EdgeID)
}
func (self *TCGraphExplorer) GetOtherNode(edge EdgeRef, node int32) int32 {
	e := self.graph.GetEdge(edge.EdgeID)
	if node == e.NodeA {
		return e.NodeB
	}
	if node == e.NodeB {
		return e.NodeA
	}
	return -1
}

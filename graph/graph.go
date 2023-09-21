package graph

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// graph interfaces
//******************************************

type IGraph interface {
	GetDefaultExplorer() IGraphExplorer
	GetGraphExplorer(weighting IWeighting) IGraphExplorer
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

type IGraphIndex interface {
	GetClosestNode(point geo.Coord) (int32, bool)
}

//*******************************************
// base-graph
//******************************************

type Graph struct {
	store    GraphStore
	topology TopologyStore
	weight   DefaultWeighting
	index    KDTree[int32]
}

func (self *Graph) GetDefaultExplorer() IGraphExplorer {
	return &BaseGraphExplorer{
		graph:    self,
		accessor: self.topology.GetAccessor(),
		weight:   &self.weight,
	}
}
func (self *Graph) GetGraphExplorer(weighting IWeighting) IGraphExplorer {
	return &BaseGraphExplorer{
		graph:    self,
		accessor: self.topology.GetAccessor(),
		weight:   weighting,
	}
}
func (self *Graph) NodeCount() int {
	return self.store.NodeCount()
}
func (self *Graph) EdgeCount() int {
	return self.store.EdgeCount()
}
func (self *Graph) IsNode(node int32) bool {
	return self.store.IsNode(node)
}
func (self *Graph) GetNode(node int32) Node {
	return self.store.GetNode(node)
}
func (self *Graph) GetEdge(edge int32) Edge {
	return self.store.GetEdge(edge)
}
func (self *Graph) GetNodeGeom(node int32) geo.Coord {
	return self.store.GetNodeGeom(node)
}
func (self *Graph) GetEdgeGeom(edge int32) geo.CoordArray {
	return self.store.GetEdgeGeom(edge)
}
func (self *Graph) GetIndex() IGraphIndex {
	return &BaseGraphIndex{
		index: self.index,
	}
}

//*******************************************
// base-graph explorer
//******************************************

type BaseGraphExplorer struct {
	graph    *Graph
	accessor TopologyAccessor
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

//*******************************************
// graph index
//******************************************

type BaseGraphIndex struct {
	index KDTree[int32]
}

func (self *BaseGraphIndex) GetClosestNode(point geo.Coord) (int32, bool) {
	return self.index.GetClosest(point[:], 0.005)
}

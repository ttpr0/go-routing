package graph

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type IGraph interface {
	GetGeometry() IGeometry
	GetWeighting() IWeighting
	GetDefaultExplorer() IGraphExplorer
	GetGraphExplorer(weighting IWeighting) IGraphExplorer
	NodeCount() int32
	EdgeCount() int32
	IsNode(node int32) bool
	GetNode(node int32) Node
	GetEdge(edge int32) Edge
	GetIndex() IGraphIndex
}

// not thread safe, use only one instance per thread
type IGraphExplorer interface {
	// multiple calls to this will overwrite underlying iterator object
	// use only one instance at a time
	GetAdjacentEdges(node int32, direction Direction) IIterator[EdgeRef]
	GetEdgeWeight(edge EdgeRef) int32
	GetTurnCost(from EdgeRef, via int32, to EdgeRef) int32
	GetOtherNode(edge EdgeRef, node int32) int32
}

type IGraphIndex interface {
	GetClosestNode(point geo.Coord) (int32, bool)
}

type Graph struct {
	nodes    NodeStore
	edges    EdgeStore
	topology TopologyStore
	geom     GeometryStore
	weight   DefaultWeighting
	index    KDTree[int32]
}

func (self *Graph) GetGeometry() IGeometry {
	return &self.geom
}
func (self *Graph) GetWeighting() IWeighting {
	return &self.weight
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
func (self *Graph) NodeCount() int32 {
	return int32(self.nodes.NodeCount())
}
func (self *Graph) EdgeCount() int32 {
	return int32(self.edges.EdgeCount())
}
func (self *Graph) IsNode(node int32) bool {
	return self.nodes.IsNode(node)
}
func (self *Graph) GetNode(node int32) Node {
	return self.nodes.GetNode(node)
}
func (self *Graph) GetEdge(edge int32) Edge {
	return self.edges.GetEdge(edge)
}
func (self *Graph) GetIndex() IGraphIndex {
	return &BaseGraphIndex{
		index: self.index,
	}
}

type BaseGraphExplorer struct {
	graph    *Graph
	accessor TopologyAccessor
	weight   IWeighting
}

func (self *BaseGraphExplorer) GetAdjacentEdges(node int32, direction Direction) IIterator[EdgeRef] {
	accessor := &self.accessor
	accessor.SetBaseNode(node, direction)
	return &EdgeRefIterator{
		accessor: accessor,
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

type EdgeRefIterator struct {
	accessor *TopologyAccessor
}

func (self *EdgeRefIterator) Next() (EdgeRef, bool) {
	ok := self.accessor.Next()
	if !ok {
		var t EdgeRef
		return t, false
	}
	edge_id := self.accessor.GetEdgeID()
	other_id := self.accessor.GetOtherID()
	return EdgeRef{
		EdgeID:  edge_id,
		OtherID: other_id,
		_Type:   0,
	}, true
}

type BaseGraphIndex struct {
	index KDTree[int32]
}

func (self *BaseGraphIndex) GetClosestNode(point geo.Coord) (int32, bool) {
	return self.index.GetClosest(point[:], 0.005)
}

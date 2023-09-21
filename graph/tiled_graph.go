package graph

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// tiled-graph interface
//******************************************

type ITiledGraph interface {
	// Base IGraph
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

	// Additional
	GetNodeTile(node int32) int16
	TileCount() int16
	GetShortcut(shc int32) Shortcut
	GetEdgesFromShortcut(edges *List[int32], shortcut_id int32)
}

//*******************************************
// tiled-graph
//******************************************

type TiledGraph struct {
	// Base Graph
	store    GraphStore
	topology AdjacencyArray
	weight   DefaultWeighting
	index    KDTree[int32]

	// Tiles Storage
	skip_store    TiledStore
	skip_topology AdjacencyArray
}

func (self *TiledGraph) GetDefaultExplorer() IGraphExplorer {
	return &TiledGraphExplorer{
		graph:         self,
		accessor:      self.topology.GetAccessor(),
		skip_accessor: self.skip_topology.GetAccessor(),
		weight:        &self.weight,
		skip_weight:   &DefaultWeighting{edge_weights: self.skip_store.skip_weights},
	}
}
func (self *TiledGraph) GetGraphExplorer(weighting IWeighting) IGraphExplorer {
	return &TiledGraphExplorer{
		graph:         self,
		accessor:      self.topology.GetAccessor(),
		skip_accessor: self.skip_topology.GetAccessor(),
		weight:        weighting,
		skip_weight:   &DefaultWeighting{edge_weights: self.skip_store.skip_weights},
	}
}
func (self *TiledGraph) GetNodeTile(node int32) int16 {
	return self.skip_store.GetNodeTile(node)
}
func (self *TiledGraph) NodeCount() int {
	return self.store.NodeCount()
}
func (self *TiledGraph) EdgeCount() int {
	return self.store.EdgeCount()
}
func (self *TiledGraph) TileCount() int16 {
	return self.skip_store.TileCount()
}
func (self *TiledGraph) IsNode(node int32) bool {
	return self.store.IsNode(node)
}
func (self *TiledGraph) GetNode(node int32) Node {
	return self.store.GetNode(node)
}
func (self *TiledGraph) GetEdge(edge int32) Edge {
	return self.store.GetEdge(edge)
}
func (self *TiledGraph) GetNodeGeom(node int32) geo.Coord {
	return self.store.GetNodeGeom(node)
}
func (self *TiledGraph) GetEdgeGeom(edge int32) geo.CoordArray {
	return self.store.GetEdgeGeom(edge)
}
func (self *TiledGraph) GetShortcut(shc int32) Shortcut {
	return self.skip_store.GetShortcut(shc)
}
func (self *TiledGraph) GetEdgesFromShortcut(edges *List[int32], shortcut_id int32) {
	for _, ref := range self.skip_store.GetEdgesFromShortcut(shortcut_id, false) {
		edges.Add(ref)
	}
}
func (self *TiledGraph) GetIndex() IGraphIndex {
	return &BaseGraphIndex{
		index: self.index,
	}
}

//*******************************************
// tiled-graph explorer
//******************************************

type TiledGraphExplorer struct {
	graph         *TiledGraph
	accessor      AdjArrayAccessor
	skip_accessor AdjArrayAccessor
	weight        IWeighting
	skip_weight   IWeighting
}

func (self *TiledGraphExplorer) ForAdjacentEdges(node int32, direction Direction, typ Adjacency, callback func(EdgeRef)) {
	if typ == ADJACENT_SKIP {
		self.skip_accessor.SetBaseNode(node, direction)
		for self.skip_accessor.Next() {
			edge_id := self.skip_accessor.GetEdgeID()
			other_id := self.skip_accessor.GetOtherID()
			typ := self.skip_accessor.GetType()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				_Type:   typ,
			})
		}
	} else if typ == ADJACENT_ALL || typ == ADJACENT_EDGES {
		self.accessor.SetBaseNode(node, direction)
		for self.accessor.Next() {
			edge_id := self.accessor.GetEdgeID()
			other_id := self.accessor.GetOtherID()
			typ := self.graph.skip_store.edge_types[edge_id]
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				_Type:   typ,
			})
		}
	} else {
		panic("Adjacency-type not implemented for this graph.")
	}
}
func (self *TiledGraphExplorer) GetEdgeWeight(edge EdgeRef) int32 {
	if edge.IsShortcut() {
		return self.skip_weight.GetEdgeWeight(edge.EdgeID)
	} else {
		return self.weight.GetEdgeWeight(edge.EdgeID)
	}
}
func (self *TiledGraphExplorer) GetTurnCost(from EdgeRef, via int32, to EdgeRef) int32 {
	return self.weight.GetTurnCost(from.EdgeID, via, to.EdgeID)
}
func (self *TiledGraphExplorer) GetOtherNode(edge EdgeRef, node int32) int32 {
	if edge.IsShortcut() {
		e := self.graph.GetShortcut(edge.EdgeID)
		if node == e.NodeA {
			return e.NodeB
		}
		if node == e.NodeB {
			return e.NodeA
		}
		return -1
	} else {
		e := self.graph.GetEdge(edge.EdgeID)
		if node == e.NodeA {
			return e.NodeB
		}
		if node == e.NodeB {
			return e.NodeA
		}
		return -1
	}
}

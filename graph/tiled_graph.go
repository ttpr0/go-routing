package graph

import (
	"errors"

	"github.com/ttpr0/go-routing/comps"
	"github.com/ttpr0/go-routing/geo"
	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// tiled-graph interface
//******************************************

type ITiledGraph interface {
	// Base IGraph
	GetGraphExplorer() IGraphExplorer
	NodeCount() int
	EdgeCount() int
	IsNode(node int32) bool
	GetNode(node int32) structs.Node
	GetEdge(edge int32) structs.Edge
	GetNodeGeom(node int32) geo.Coord

	// Additional
	GetNodeTile(node int32) int16
	TileCount() int16
	GetShortcut(shc int32) structs.Shortcut
	GetEdgesFromShortcut(shortcut_id int32, reversed bool, handler func(int32))
	HasCellIndex() bool
	GetIndexEdges(tile int16, dir Direction) (Array[structs.Shortcut], error)
}

//*******************************************
// tiled-graph
//******************************************

type TiledGraph struct {
	// Base Graph
	base   comps.IGraphBase
	weight comps.IWeighting

	// Tiles Storage
	partition  *comps.Partition
	overlay    *comps.Overlay
	cell_index Optional[*comps.CellIndex]
}

func (self *TiledGraph) GetGraphExplorer() IGraphExplorer {
	return &TiledGraphExplorer{
		graph:         self,
		accessor:      self.base.GetAccessor(),
		skip_accessor: self.overlay.GetAccessor(),
		weight:        self.weight,
	}
}
func (self *TiledGraph) GetNodeTile(node int32) int16 {
	return self.partition.GetNodeTile(node)
}
func (self *TiledGraph) NodeCount() int {
	return self.base.NodeCount()
}
func (self *TiledGraph) EdgeCount() int {
	return self.base.EdgeCount()
}
func (self *TiledGraph) TileCount() int16 {
	return self.partition.TileCount()
}
func (self *TiledGraph) IsNode(node int32) bool {
	return node < int32(self.base.NodeCount())
}
func (self *TiledGraph) GetNode(node int32) structs.Node {
	return self.base.GetNode(node)
}
func (self *TiledGraph) GetEdge(edge int32) structs.Edge {
	return self.base.GetEdge(edge)
}
func (self *TiledGraph) GetNodeGeom(node int32) geo.Coord {
	return self.GetNode(node).Loc
}
func (self *TiledGraph) GetShortcut(shc int32) structs.Shortcut {
	return self.overlay.GetShortcut(shc)
}
func (self *TiledGraph) GetEdgesFromShortcut(shc_id int32, reversed bool, handler func(int32)) {
	self.overlay.GetEdgesFromShortcut(shc_id, reversed, handler)
}
func (self *TiledGraph) GetIndexEdges(tile int16, dir Direction) (Array[structs.Shortcut], error) {
	if !self.cell_index.HasValue() {
		return nil, errors.New("graph doesnt have cell-index")
	}
	if dir == FORWARD {
		return self.cell_index.Value.GetFWDIndexEdges(tile), nil
	} else {
		return self.cell_index.Value.GetBWDIndexEdges(tile), nil
	}
}
func (self *TiledGraph) HasCellIndex() bool {
	return self.cell_index.HasValue()
}

//*******************************************
// tiled-graph explorer
//*******************************************

type TiledGraphExplorer struct {
	graph         *TiledGraph
	accessor      structs.IAdjAccessor
	skip_accessor structs.IAdjAccessor
	weight        comps.IWeighting
}

func (self *TiledGraphExplorer) ForAdjacentEdges(node int32, direction Direction, typ Adjacency, callback func(EdgeRef)) {
	if typ == ADJACENT_SKIP {
		self.skip_accessor.SetBaseNode(node, direction == FORWARD)
		for self.skip_accessor.Next() {
			edge_id := self.skip_accessor.GetEdgeID()
			other_id := self.skip_accessor.GetOtherID()
			typ := self.skip_accessor.GetType()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				Type:    typ,
			})
		}
	} else if typ == ADJACENT_ALL || typ == ADJACENT_EDGES {
		self.accessor.SetBaseNode(node, direction == FORWARD)
		for self.accessor.Next() {
			edge_id := self.accessor.GetEdgeID()
			other_id := self.accessor.GetOtherID()
			typ := self.graph.overlay.GetEdgeType(edge_id)
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				Type:    typ,
			})
		}
	} else {
		panic("Adjacency-type not implemented for this graph.")
	}
}
func (self *TiledGraphExplorer) GetEdgeWeight(edge EdgeRef) int32 {
	if edge.IsShortcut() {
		shc := self.graph.overlay.GetShortcut(edge.EdgeID)
		return shc.Weight
	} else {
		return self.weight.GetEdgeWeight(edge.EdgeID)
	}
}
func (self *TiledGraphExplorer) GetTurnCost(from EdgeRef, via int32, to EdgeRef) int32 {
	if from.IsShortcut() || to.IsShortcut() {
		return 0
	}
	return 0
}
func (self *TiledGraphExplorer) GetOtherNode(edge EdgeRef, node int32) int32 {
	if edge.IsShortcut() {
		e := self.graph.GetShortcut(edge.EdgeID)
		if node == e.From {
			return e.To
		}
		if node == e.To {
			return e.From
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

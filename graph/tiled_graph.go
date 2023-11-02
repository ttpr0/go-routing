package graph

import (
	"errors"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// tiled-graph interface
//******************************************

type ITiledGraph interface {
	// Base IGraph
	GetGraphExplorer() IGraphExplorer
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
	HasCellIndex() bool
	GetIndexEdges(tile int16, dir Direction) (Array[Shortcut], error)
}

//*******************************************
// tiled-graph
//******************************************

type TiledGraph struct {
	// Base Graph
	base   GraphBase
	weight IWeighting

	// Tiles Storage
	skip_shortcuts ShortcutStore
	skip_topology  AdjacencyArray
	node_tiles     Array[int16]
	edge_types     Array[byte]
	cell_index     Optional[_CellIndex] // Storage for indexing sp within cells
}

func (self *TiledGraph) GetGraphExplorer() IGraphExplorer {
	return &TiledGraphExplorer{
		graph:         self,
		accessor:      self.base.GetAccessor(),
		skip_accessor: self.skip_topology.GetAccessor(),
		weight:        self.weight,
	}
}
func (self *TiledGraph) GetNodeTile(node int32) int16 {
	return self.node_tiles[node]
}
func (self *TiledGraph) NodeCount() int {
	return self.base.NodeCount()
}
func (self *TiledGraph) EdgeCount() int {
	return self.base.EdgeCount()
}
func (self *TiledGraph) TileCount() int16 {
	max := int16(0)
	for i := 0; i < len(self.node_tiles); i++ {
		tile := self.node_tiles[i]
		if tile > max {
			max = tile
		}
	}
	return max - 1
}
func (self *TiledGraph) IsNode(node int32) bool {
	return self.base.NodeCount() < int(node)
}
func (self *TiledGraph) GetNode(node int32) Node {
	return self.base.GetNode(node)
}
func (self *TiledGraph) GetEdge(edge int32) Edge {
	return self.base.GetEdge(edge)
}
func (self *TiledGraph) GetNodeGeom(node int32) geo.Coord {
	return self.base.GetNodeGeom(node)
}
func (self *TiledGraph) GetEdgeGeom(edge int32) geo.CoordArray {
	return self.base.GetEdgeGeom(edge)
}
func (self *TiledGraph) GetShortcut(shc int32) Shortcut {
	return self.skip_shortcuts.GetShortcut(shc)
}
func (self *TiledGraph) GetEdgesFromShortcut(edges *List[int32], shc_id int32) {
	self.skip_shortcuts.GetEdgesFromShortcut(shc_id, false, func(edge int32) {
		edges.Add(edge)
	})
}
func (self *TiledGraph) GetIndexEdges(tile int16, dir Direction) (Array[Shortcut], error) {
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
func (self *TiledGraph) GetIndex() IGraphIndex {
	return &BaseGraphIndex{
		index: self.base.GetKDTree(),
	}
}

//*******************************************
// tiled-graph explorer
//*******************************************

type TiledGraphExplorer struct {
	graph         *TiledGraph
	accessor      AdjArrayAccessor
	skip_accessor AdjArrayAccessor
	weight        IWeighting
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
			typ := self.graph.edge_types[edge_id]
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
		shc := self.graph.skip_shortcuts.GetShortcut(edge.EdgeID)
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

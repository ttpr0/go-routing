package graph

import (
	"errors"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// tiled-graph
//******************************************

type TiledGraph2 struct {
	// Base Graph
	base   GraphBase
	weight IWeighting

	// ID-mapping
	id_mapping _IDMapping

	// Tiles Storage
	skip_shortcuts ShortcutStore
	skip_topology  AdjacencyArray
	node_tiles     Array[int16]
	edge_types     Array[byte]
	cell_index     Optional[_CellIndex] // Storage for indexing sp within cells
}

func (self *TiledGraph2) GetGraphExplorer() IGraphExplorer {
	return &TiledGraph2Explorer{
		graph:         self,
		id_mapping:    self.id_mapping,
		accessor:      self.base.GetAccessor(),
		skip_accessor: self.skip_topology.GetAccessor(),
		weight:        self.weight,
	}
}
func (self *TiledGraph2) GetNodeTile(node int32) int16 {
	return self.node_tiles[node]
}
func (self *TiledGraph2) NodeCount() int {
	return self.base.NodeCount()
}
func (self *TiledGraph2) EdgeCount() int {
	return self.base.EdgeCount()
}
func (self *TiledGraph2) TileCount() int16 {
	max := int16(0)
	for i := 0; i < len(self.node_tiles); i++ {
		tile := self.node_tiles[i]
		if tile > max {
			max = tile
		}
	}
	return max - 1
}
func (self *TiledGraph2) IsNode(node int32) bool {
	m_node := self.id_mapping.GetSource(node)
	return self.base.NodeCount() < int(m_node)
}
func (self *TiledGraph2) GetNode(node int32) Node {
	m_node := self.id_mapping.GetSource(node)
	return self.base.GetNode(m_node)
}
func (self *TiledGraph2) GetEdge(edge int32) Edge {
	e := self.base.GetEdge(edge)
	e.NodeA = self.id_mapping.GetTarget(e.NodeA)
	e.NodeB = self.id_mapping.GetTarget(e.NodeB)
	return e
}
func (self *TiledGraph2) GetNodeGeom(node int32) geo.Coord {
	m_node := self.id_mapping.GetSource(node)
	return self.base.GetNodeGeom(m_node)
}
func (self *TiledGraph2) GetEdgeGeom(edge int32) geo.CoordArray {
	return self.base.GetEdgeGeom(edge)
}
func (self *TiledGraph2) GetShortcut(shc int32) Shortcut {
	return self.skip_shortcuts.GetShortcut(shc)
}
func (self *TiledGraph2) GetEdgesFromShortcut(edges *List[int32], shc_id int32) {
	self.skip_shortcuts.GetEdgesFromShortcut(shc_id, false, func(edge int32) {
		edges.Add(edge)
	})
}
func (self *TiledGraph2) GetIndexEdges(tile int16, dir Direction) (Array[Shortcut], error) {
	if !self.cell_index.HasValue() {
		return nil, errors.New("graph doesnt have cell-index")
	}
	if dir == FORWARD {
		return self.cell_index.Value.GetFWDIndexEdges(tile), nil
	} else {
		return self.cell_index.Value.GetBWDIndexEdges(tile), nil
	}
}
func (self *TiledGraph2) HasCellIndex() bool {
	return self.cell_index.HasValue()
}
func (self *TiledGraph2) GetIndex() IGraphIndex {
	return &MappedGraphIndex{
		id_mapping: self.id_mapping,
		index:      self.base.GetKDTree(),
	}
}

//*******************************************
// tiled-graph explorer
//*******************************************

type TiledGraph2Explorer struct {
	graph         *TiledGraph2
	id_mapping    _IDMapping
	accessor      AdjArrayAccessor
	skip_accessor AdjArrayAccessor
	weight        IWeighting
}

func (self *TiledGraph2Explorer) ForAdjacentEdges(node int32, direction Direction, typ Adjacency, callback func(EdgeRef)) {
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
			m_other_id := self.id_mapping.GetTarget(other_id)
			typ := self.graph.edge_types[edge_id]
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: m_other_id,
				_Type:   typ,
			})
		}
	} else {
		panic("Adjacency-type not implemented for this graph.")
	}
}
func (self *TiledGraph2Explorer) GetEdgeWeight(edge EdgeRef) int32 {
	if edge.IsShortcut() {
		shc := self.graph.skip_shortcuts.GetShortcut(edge.EdgeID)
		return shc.Weight
	} else {
		return self.weight.GetEdgeWeight(edge.EdgeID)
	}
}
func (self *TiledGraph2Explorer) GetTurnCost(from EdgeRef, via int32, to EdgeRef) int32 {
	if from.IsShortcut() || to.IsShortcut() {
		return 0
	}
	return 0
}
func (self *TiledGraph2Explorer) GetOtherNode(edge EdgeRef, node int32) int32 {
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
		node_a := e.NodeA
		node_b := e.NodeB
		if node == node_a {
			return node_b
		}
		if node == node_b {
			return node_a
		}
		return -1
	}
}

package graph

import (
	"errors"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// ch-graph
//******************************************

type CHGraph2 struct {
	// Base Graph
	base   GraphBase
	weight IWeighting

	// ID-mapping
	id_mapping _IDMapping

	// Additional Storage
	ch_shortcuts ShortcutStore
	ch_topology  AdjacencyArray
	node_levels  Array[int16]

	// contraction order build with tiles
	_build_with_tiles bool
	node_tiles        Optional[Array[int16]]

	// index for PHAST
	_contains_dummies bool
	fwd_down_edges    Optional[Array[Shortcut]]
	bwd_down_edges    Optional[Array[Shortcut]]
}

func (self *CHGraph2) GetGraphExplorer() IGraphExplorer {
	return &CHGraph2Explorer{
		graph:       self,
		id_mapping:  self.id_mapping,
		accessor:    self.base.GetAccessor(),
		sh_accessor: self.ch_topology.GetAccessor(),
		weight:      self.weight,
	}
}

func (self *CHGraph2) GetNodeLevel(node int32) int16 {
	return self.node_levels[node]
}

func (self *CHGraph2) NodeCount() int {
	return self.base.NodeCount()
}

func (self *CHGraph2) EdgeCount() int {
	return self.base.EdgeCount()
}

func (self *CHGraph2) ShortcutCount() int {
	return self.ch_shortcuts.ShortcutCount()
}

func (self *CHGraph2) IsNode(node int32) bool {
	m_node := self.id_mapping.GetSource(node)
	return self.base.NodeCount() < int(m_node)
}

func (self *CHGraph2) GetNode(node int32) Node {
	m_node := self.id_mapping.GetSource(node)
	return self.base.GetNode(m_node)
}

func (self *CHGraph2) GetEdge(edge int32) Edge {
	e := self.base.GetEdge(edge)
	e.NodeA = self.id_mapping.GetTarget(e.NodeA)
	e.NodeB = self.id_mapping.GetTarget(e.NodeB)
	return e
}

func (self *CHGraph2) GetNodeGeom(node int32) geo.Coord {
	m_node := self.id_mapping.GetSource(node)
	return self.base.GetNodeGeom(m_node)
}
func (self *CHGraph2) GetEdgeGeom(edge int32) geo.CoordArray {
	return self.base.GetEdgeGeom(edge)
}

func (self *CHGraph2) GetShortcut(shortcut int32) Shortcut {
	return self.ch_shortcuts.GetShortcut(shortcut)
}

func (self *CHGraph2) GetEdgesFromShortcut(edges *List[int32], shc_id int32, reversed bool) {
	self.ch_shortcuts.GetEdgesFromShortcut(shc_id, false, func(edge int32) {
		edges.Add(edge)
	})
}
func (self *CHGraph2) GetDownEdges(dir Direction) (Array[Shortcut], error) {
	if dir == FORWARD {
		if self.fwd_down_edges.HasValue() {
			return self.fwd_down_edges.Value, nil
		} else {
			return nil, errors.New("forward downedges not build for this graph")
		}
	} else {
		if self.bwd_down_edges.HasValue() {
			return self.bwd_down_edges.Value, nil
		} else {
			return nil, errors.New("backward downedges not build for this graph")
		}
	}
}
func (self *CHGraph2) HasDownEdges(dir Direction) bool {
	if dir == FORWARD {
		return self.fwd_down_edges.HasValue()
	} else {
		return self.bwd_down_edges.HasValue()
	}
}
func (self *CHGraph2) GetNodeTile(node int32) int16 {
	if self.node_tiles.HasValue() {
		return self.node_tiles.Value[node]
	} else {
		return -1
	}
}
func (self *CHGraph2) TileCount() int {
	if self.node_tiles.HasValue() {
		max := int16(0)
		for i := 0; i < len(self.node_tiles.Value); i++ {
			tile := self.node_tiles.Value[i]
			if tile > max {
				max = tile
			}
		}
		return int(max + 1)
	} else {
		return -1
	}
}
func (self *CHGraph2) GetIndex() IGraphIndex {
	return &MappedGraphIndex{
		id_mapping: self.id_mapping,
		index:      self.base.GetKDTree(),
	}
}

//*******************************************
// ch-graph explorer
//******************************************

type CHGraph2Explorer struct {
	graph       *CHGraph2
	id_mapping  _IDMapping
	accessor    AdjArrayAccessor
	sh_accessor AdjArrayAccessor
	weight      IWeighting
}

func (self *CHGraph2Explorer) ForAdjacentEdges(node int32, direction Direction, typ Adjacency, callback func(EdgeRef)) {
	if typ == ADJACENT_ALL {
		m_node := self.id_mapping.GetSource(node)
		self.accessor.SetBaseNode(m_node, direction)
		self.sh_accessor.SetBaseNode(node, direction)
		for self.accessor.Next() {
			edge_id := self.accessor.GetEdgeID()
			other_id := self.accessor.GetOtherID()
			m_other_id := self.id_mapping.GetTarget(other_id)
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: m_other_id,
				_Type:   0,
			})
		}
		for self.sh_accessor.Next() {
			edge_id := self.sh_accessor.GetEdgeID()
			other_id := self.sh_accessor.GetOtherID()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				_Type:   100,
			})
		}
	} else if typ == ADJACENT_EDGES {
		m_node := self.id_mapping.GetSource(node)
		self.accessor.SetBaseNode(m_node, direction)
		for self.accessor.Next() {
			edge_id := self.accessor.GetEdgeID()
			other_id := self.accessor.GetOtherID()
			m_other_id := self.id_mapping.GetTarget(other_id)
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: m_other_id,
				_Type:   0,
			})
		}
	} else if typ == ADJACENT_SHORTCUTS {
		self.sh_accessor.SetBaseNode(node, direction)
		for self.sh_accessor.Next() {
			edge_id := self.sh_accessor.GetEdgeID()
			other_id := self.sh_accessor.GetOtherID()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				_Type:   100,
			})
		}
	} else if typ == ADJACENT_UPWARDS {
		m_node := self.id_mapping.GetSource(node)
		self.accessor.SetBaseNode(m_node, direction)
		self.sh_accessor.SetBaseNode(node, direction)
		this_level := self.graph.GetNodeLevel(node)
		for self.accessor.Next() {
			other_id := self.accessor.GetOtherID()
			m_other_id := self.id_mapping.GetTarget(other_id)
			if this_level >= self.graph.GetNodeLevel(m_other_id) {
				continue
			}
			edge_id := self.accessor.GetEdgeID()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: m_other_id,
				_Type:   0,
			})
		}
		for self.sh_accessor.Next() {
			other_id := self.sh_accessor.GetOtherID()
			if this_level >= self.graph.GetNodeLevel(other_id) {
				continue
			}
			edge_id := self.sh_accessor.GetEdgeID()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				_Type:   100,
			})
		}
	} else if typ == ADJACENT_DOWNWARDS {
		m_node := self.id_mapping.GetSource(node)
		self.accessor.SetBaseNode(m_node, direction)
		self.sh_accessor.SetBaseNode(node, direction)
		this_level := self.graph.GetNodeLevel(node)
		for self.accessor.Next() {
			other_id := self.accessor.GetOtherID()
			m_other_id := self.id_mapping.GetTarget(other_id)
			if this_level <= self.graph.GetNodeLevel(m_other_id) {
				continue
			}
			edge_id := self.accessor.GetEdgeID()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: m_other_id,
				_Type:   0,
			})
		}
		for self.sh_accessor.Next() {
			other_id := self.sh_accessor.GetOtherID()
			if this_level <= self.graph.GetNodeLevel(other_id) {
				continue
			}
			edge_id := self.sh_accessor.GetEdgeID()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				_Type:   100,
			})
		}
	} else {
		panic("Adjacency-type not implemented for this graph.")
	}
}
func (self *CHGraph2Explorer) GetEdgeWeight(edge EdgeRef) int32 {
	if edge.IsCHShortcut() {
		shc := self.graph.ch_shortcuts.GetShortcut(edge.EdgeID)
		return shc.Weight
	} else {
		return self.weight.GetEdgeWeight(edge.EdgeID)
	}
}
func (self *CHGraph2Explorer) GetTurnCost(from EdgeRef, via int32, to EdgeRef) int32 {
	if from.IsShortcut() || to.IsShortcut() {
		return 0
	}
	return 0
}
func (self *CHGraph2Explorer) GetOtherNode(edge EdgeRef, node int32) int32 {
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

//*******************************************
// graph index
//******************************************

type MappedGraphIndex struct {
	id_mapping _IDMapping
	index      KDTree[int32]
}

func (self *MappedGraphIndex) GetClosestNode(point geo.Coord) (int32, bool) {
	node, ok := self.index.GetClosest(point[:], 0.005)
	if !ok {
		return node, ok
	}
	mapped_node := self.id_mapping.GetTarget(node)
	return mapped_node, true
}

package graph

import (
	"errors"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// ch-graph interface
//******************************************

type ICHGraph interface {
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

	// CH Specific
	GetNodeLevel(node int32) int16
	ShortcutCount() int
	GetShortcut(shortcut int32) Shortcut
	GetEdgesFromShortcut(edges *List[int32], shortcut_id int32, reversed bool)
	HasDownEdges(dir Direction) bool
	GetDownEdges(dir Direction) (Array[Shortcut], error)
	GetNodeTile(node int32) int16
	TileCount() int
}

//*******************************************
// ch-graph
//******************************************

type CHGraph struct {
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

func (self *CHGraph) GetGraphExplorer() IGraphExplorer {
	return &CHGraphExplorer{
		graph:       self,
		accessor:    self.base.GetAccessor(),
		sh_accessor: self.ch_topology.GetAccessor(),
		weight:      self.weight,
	}
}

func (self *CHGraph) GetNodeLevel(node int32) int16 {
	return self.node_levels[node]
}

func (self *CHGraph) NodeCount() int {
	return self.base.NodeCount()
}

func (self *CHGraph) EdgeCount() int {
	return self.base.EdgeCount()
}

func (self *CHGraph) ShortcutCount() int {
	return self.ch_shortcuts.ShortcutCount()
}

func (self *CHGraph) IsNode(node int32) bool {
	return self.base.NodeCount() < int(node)
}

func (self *CHGraph) GetNode(node int32) Node {
	return self.base.GetNode(node)
}

func (self *CHGraph) GetEdge(edge int32) Edge {
	return self.base.GetEdge(edge)
}

func (self *CHGraph) GetNodeGeom(node int32) geo.Coord {
	return self.base.GetNodeGeom(node)
}
func (self *CHGraph) GetEdgeGeom(edge int32) geo.CoordArray {
	return self.base.GetEdgeGeom(edge)
}

func (self *CHGraph) GetShortcut(shortcut int32) Shortcut {
	return self.ch_shortcuts.GetShortcut(shortcut)
}

func (self *CHGraph) GetEdgesFromShortcut(edges *List[int32], shc_id int32, reversed bool) {
	self.ch_shortcuts.GetEdgesFromShortcut(shc_id, false, func(edge int32) {
		edges.Add(edge)
	})
}
func (self *CHGraph) GetDownEdges(dir Direction) (Array[Shortcut], error) {
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
func (self *CHGraph) HasDownEdges(dir Direction) bool {
	if dir == FORWARD {
		return self.fwd_down_edges.HasValue()
	} else {
		return self.bwd_down_edges.HasValue()
	}
}
func (self *CHGraph) GetNodeTile(node int32) int16 {
	if self.node_tiles.HasValue() {
		return self.node_tiles.Value[node]
	} else {
		return -1
	}
}
func (self *CHGraph) TileCount() int {
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
func (self *CHGraph) GetIndex() IGraphIndex {
	return &BaseGraphIndex{
		index: self.base.GetKDTree(),
	}
}

//*******************************************
// ch-graph explorer
//******************************************

type CHGraphExplorer struct {
	graph       *CHGraph
	accessor    AdjArrayAccessor
	sh_accessor AdjArrayAccessor
	weight      IWeighting
}

func (self *CHGraphExplorer) ForAdjacentEdges(node int32, direction Direction, typ Adjacency, callback func(EdgeRef)) {
	if typ == ADJACENT_ALL {
		self.accessor.SetBaseNode(node, direction)
		self.sh_accessor.SetBaseNode(node, direction)
		for self.accessor.Next() {
			edge_id := self.accessor.GetEdgeID()
			other_id := self.accessor.GetOtherID()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
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
		self.accessor.SetBaseNode(node, direction)
		self.sh_accessor.SetBaseNode(node, direction)
		this_level := self.graph.GetNodeLevel(node)
		for self.accessor.Next() {
			other_id := self.accessor.GetOtherID()
			if this_level >= self.graph.GetNodeLevel(other_id) {
				continue
			}
			edge_id := self.accessor.GetEdgeID()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
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
		self.accessor.SetBaseNode(node, direction)
		self.sh_accessor.SetBaseNode(node, direction)
		this_level := self.graph.GetNodeLevel(node)
		for self.accessor.Next() {
			other_id := self.accessor.GetOtherID()
			if this_level <= self.graph.GetNodeLevel(other_id) {
				continue
			}
			edge_id := self.accessor.GetEdgeID()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
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
func (self *CHGraphExplorer) GetEdgeWeight(edge EdgeRef) int32 {
	if edge.IsCHShortcut() {
		shc := self.graph.ch_shortcuts.GetShortcut(edge.EdgeID)
		return shc.Weight
	} else {
		return self.weight.GetEdgeWeight(edge.EdgeID)
	}
}
func (self *CHGraphExplorer) GetTurnCost(from EdgeRef, via int32, to EdgeRef) int32 {
	if from.IsShortcut() || to.IsShortcut() {
		return 0
	}
	return 0
}
func (self *CHGraphExplorer) GetOtherNode(edge EdgeRef, node int32) int32 {
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

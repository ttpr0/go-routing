package graph

import (
	"errors"

	"github.com/ttpr0/go-routing/comps"
	"github.com/ttpr0/go-routing/geo"
	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// ch-graph interface
//******************************************

type ICHGraph interface {
	// Base IGraph
	GetGraphExplorer() IGraphExplorer
	NodeCount() int
	EdgeCount() int
	IsNode(node int32) bool
	GetNode(node int32) structs.Node
	GetEdge(edge int32) structs.Edge
	GetNodeGeom(node int32) geo.Coord
	GetClosestNode(point geo.Coord) (int32, bool)

	// CH Specific
	GetNodeLevel(node int32) int16
	ShortcutCount() int
	GetShortcut(shortcut int32) structs.Shortcut
	GetEdgesFromShortcut(shortcut_id int32, reversed bool, handler func(int32))
	HasDownEdges(dir Direction) bool
	GetDownEdges(dir Direction) (Array[structs.Shortcut], error)
	GetNodeTile(node int32) int16
	TileCount() int
}

//*******************************************
// ch-graph
//******************************************

type CHGraph struct {
	// Base Graph
	base   comps.IGraphBase
	weight comps.IWeighting
	index  Optional[comps.IGraphIndex]

	// Additional Storage
	ch *comps.CH

	// contraction order build with tiles
	partition Optional[*comps.Partition]

	// index for PHAST
	ch_index Optional[*comps.CHIndex]
}

func (self *CHGraph) GetGraphExplorer() IGraphExplorer {
	return &CHGraphExplorer{
		graph:       self,
		accessor:    self.base.GetAccessor(),
		sh_accessor: self.ch.GetShortcutAccessor(),
		weight:      self.weight,
	}
}

func (self *CHGraph) GetNodeLevel(node int32) int16 {
	return self.ch.GetNodeLevel(node)
}

func (self *CHGraph) NodeCount() int {
	return self.base.NodeCount()
}

func (self *CHGraph) EdgeCount() int {
	return self.base.EdgeCount()
}

func (self *CHGraph) ShortcutCount() int {
	return self.ch.ShortcutCount()
}

func (self *CHGraph) IsNode(node int32) bool {
	return self.base.NodeCount() < int(node)
}

func (self *CHGraph) GetNode(node int32) structs.Node {
	return self.base.GetNode(node)
}

func (self *CHGraph) GetEdge(edge int32) structs.Edge {
	return self.base.GetEdge(edge)
}

func (self *CHGraph) GetNodeGeom(node int32) geo.Coord {
	return self.base.GetNode(node).Loc
}

func (self *CHGraph) GetClosestNode(point geo.Coord) (int32, bool) {
	if self.index.HasValue() {
		return self.index.Value.GetClosestNode(point)
	} else {
		self.index.Value = comps.NewGraphIndex(self.base)
		return self.index.Value.GetClosestNode(point)
	}
}

func (self *CHGraph) GetShortcut(shortcut int32) structs.Shortcut {
	return self.ch.GetShortcut(shortcut)
}

func (self *CHGraph) GetEdgesFromShortcut(shc_id int32, reversed bool, handler func(int32)) {
	self.ch.GetEdgesFromShortcut(shc_id, false, handler)
}
func (self *CHGraph) GetDownEdges(dir Direction) (Array[structs.Shortcut], error) {
	if !self.ch_index.HasValue() {
		return nil, errors.New("downedges not build for this graph")
	}
	if dir == FORWARD {
		down_edges := self.ch_index.Value.GetFWDDownEdges()
		if down_edges.Length() == 0 {
			return nil, errors.New("forward downedges not build for this graph")
		}
		return down_edges, nil
	} else {
		down_edges := self.ch_index.Value.GetBWDDownEdges()
		if down_edges.Length() == 0 {
			return nil, errors.New("backward downedges not build for this graph")
		}
		return down_edges, nil
	}
}
func (self *CHGraph) HasDownEdges(dir Direction) bool {
	if !self.ch_index.HasValue() {
		return false
	}
	if dir == FORWARD {
		edges := self.ch_index.Value.GetFWDDownEdges()
		return edges.Length() > 0
	} else {
		edges := self.ch_index.Value.GetBWDDownEdges()
		return edges.Length() > 0
	}
}
func (self *CHGraph) GetNodeTile(node int32) int16 {
	if self.partition.HasValue() {
		return self.partition.Value.GetNodeTile(node)
	} else {
		return -1
	}
}
func (self *CHGraph) TileCount() int {
	if self.partition.HasValue() {
		return int(self.partition.Value.TileCount())
	} else {
		return -1
	}
}

//*******************************************
// ch-graph explorer
//******************************************

type CHGraphExplorer struct {
	graph       *CHGraph
	accessor    structs.IAdjAccessor
	sh_accessor structs.IAdjAccessor
	weight      comps.IWeighting
}

func (self *CHGraphExplorer) ForAdjacentEdges(node int32, direction Direction, typ Adjacency, callback func(EdgeRef)) {
	if typ == ADJACENT_ALL {
		self.accessor.SetBaseNode(node, direction == FORWARD)
		self.sh_accessor.SetBaseNode(node, direction == FORWARD)
		for self.accessor.Next() {
			edge_id := self.accessor.GetEdgeID()
			other_id := self.accessor.GetOtherID()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				Type:    0,
			})
		}
		for self.sh_accessor.Next() {
			edge_id := self.sh_accessor.GetEdgeID()
			other_id := self.sh_accessor.GetOtherID()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				Type:    100,
			})
		}
	} else if typ == ADJACENT_EDGES {
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
	} else if typ == ADJACENT_SHORTCUTS {
		self.sh_accessor.SetBaseNode(node, direction == FORWARD)
		for self.sh_accessor.Next() {
			edge_id := self.sh_accessor.GetEdgeID()
			other_id := self.sh_accessor.GetOtherID()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				Type:    100,
			})
		}
	} else if typ == ADJACENT_UPWARDS {
		self.accessor.SetBaseNode(node, direction == FORWARD)
		self.sh_accessor.SetBaseNode(node, direction == FORWARD)
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
				Type:    0,
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
				Type:    100,
			})
		}
	} else if typ == ADJACENT_DOWNWARDS {
		self.accessor.SetBaseNode(node, direction == FORWARD)
		self.sh_accessor.SetBaseNode(node, direction == FORWARD)
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
				Type:    0,
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
				Type:    100,
			})
		}
	} else {
		panic("Adjacency-type not implemented for this graph.")
	}
}
func (self *CHGraphExplorer) GetEdgeWeight(edge EdgeRef) int32 {
	if edge.IsCHShortcut() {
		shc := self.graph.ch.GetShortcut(edge.EdgeID)
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

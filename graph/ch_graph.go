package graph

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// ch-graph interface
//******************************************

type ICHGraph interface {
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

	// CH Specific
	GetNodeLevel(node int32) int16
	ShortcutCount() int
	GetShortcut(shortcut int32) CHShortcut
	GetEdgesFromShortcut(edges *List[int32], shortcut_id int32, reversed bool)
}

//*******************************************
// ch-graph
//******************************************

type CHGraph struct {
	// Base Graph
	store    GraphStore
	topology TopologyStore
	weight   DefaultWeighting
	index    KDTree[int32]

	// Additional Storage
	ch_store    CHStore
	ch_topology TopologyStore
}

func (self *CHGraph) GetDefaultExplorer() IGraphExplorer {
	return &CHGraphExplorer{
		graph:       self,
		accessor:    self.topology.GetAccessor(),
		sh_accessor: self.ch_topology.GetAccessor(),
		weight:      &self.weight,
		sh_weight:   &DefaultWeighting{edge_weights: self.ch_store.sh_weight},
	}
}

func (self *CHGraph) GetGraphExplorer(weighting IWeighting) IGraphExplorer {
	return &CHGraphExplorer{
		graph:       self,
		accessor:    self.topology.GetAccessor(),
		sh_accessor: self.ch_topology.GetAccessor(),
		weight:      weighting,
		sh_weight:   &DefaultWeighting{edge_weights: self.ch_store.sh_weight},
	}
}

func (self *CHGraph) GetNodeLevel(node int32) int16 {
	return self.ch_store.GetNodeLevel(node)
}

func (self *CHGraph) NodeCount() int {
	return self.store.NodeCount()
}

func (self *CHGraph) EdgeCount() int {
	return self.store.EdgeCount()
}

func (self *CHGraph) ShortcutCount() int {
	return self.ch_store.ShortcutCount()
}

func (self *CHGraph) IsNode(node int32) bool {
	return self.store.IsNode(node)
}

func (self *CHGraph) GetNode(node int32) Node {
	return self.store.GetNode(node)
}

func (self *CHGraph) GetEdge(edge int32) Edge {
	return self.store.GetEdge(edge)
}

func (self *CHGraph) GetNodeGeom(node int32) geo.Coord {
	return self.store.GetNodeGeom(node)
}
func (self *CHGraph) GetEdgeGeom(edge int32) geo.CoordArray {
	return self.store.GetEdgeGeom(edge)
}

func (self *CHGraph) GetShortcut(shortcut int32) CHShortcut {
	return self.ch_store.GetShortcut(shortcut)
}

func (self *CHGraph) GetEdgesFromShortcut(edges *List[int32], shortcut_id int32, reversed bool) {
	self.ch_store._UnpackShortcutRecursive(edges, shortcut_id, reversed)
}
func (self *CHGraph) GetIndex() IGraphIndex {
	return &BaseGraphIndex{
		index: self.index,
	}
}

//*******************************************
// ch-graph explorer
//******************************************

type CHGraphExplorer struct {
	graph       *CHGraph
	accessor    TopologyAccessor
	sh_accessor TopologyAccessor
	weight      IWeighting
	sh_weight   IWeighting
}

func (self *CHGraphExplorer) GetAdjacentEdges(node int32, direction Direction, typ Adjacency) IIterator[EdgeRef] {
	if typ == ADJACENT_ALL {
		self.accessor.SetBaseNode(node, direction)
		self.sh_accessor.SetBaseNode(node, direction)
		return &CHEdgeRefIterator{
			accessor:    &self.accessor,
			ch_accessor: &self.sh_accessor,
			typ:         0,
		}
	} else if typ == ADJACENT_EDGES {
		self.accessor.SetBaseNode(node, direction)
		return &EdgeRefIterator{
			accessor: &self.accessor,
		}
	} else if typ == ADJACENT_SHORTCUTS {
		self.sh_accessor.SetBaseNode(node, direction)
		return &EdgeRefIterator{
			accessor: &self.sh_accessor,
		}
	} else if typ == ADJACENT_UPWARDS {
		self.accessor.SetBaseNode(node, direction)
		self.sh_accessor.SetBaseNode(node, direction)
		return &CHUpwardIterator{
			this_node:   node,
			node_levels: self.graph.ch_store.node_levels,
			accessor:    &self.accessor,
			ch_accessor: &self.sh_accessor,
			typ:         0,
		}
	} else if typ == ADJACENT_DOWNWARDS {
		self.accessor.SetBaseNode(node, direction)
		self.sh_accessor.SetBaseNode(node, direction)
		return &CHDownwardIterator{
			this_node:   node,
			node_levels: self.graph.ch_store.node_levels,
			accessor:    &self.accessor,
			ch_accessor: &self.sh_accessor,
			typ:         0,
		}
	} else {
		panic("Adjacency-type not implemented for this graph.")
	}
}
func (self *CHGraphExplorer) GetEdgeWeight(edge EdgeRef) int32 {
	if edge.IsCHShortcut() {
		return self.sh_weight.GetEdgeWeight(edge.EdgeID)
	} else {
		return self.weight.GetEdgeWeight(edge.EdgeID)
	}
}
func (self *CHGraphExplorer) GetTurnCost(from EdgeRef, via int32, to EdgeRef) int32 {
	return 0
}
func (self *CHGraphExplorer) GetOtherNode(edge EdgeRef, node int32) int32 {
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

//*******************************************
// edge-ref iterators
//******************************************

type CHEdgeRefIterator struct {
	accessor    *TopologyAccessor
	ch_accessor *TopologyAccessor
	typ         byte
}

func (self *CHEdgeRefIterator) Next() (EdgeRef, bool) {
	ok := self.accessor.Next()
	if !ok {
		if self.typ == 100 {
			var t EdgeRef
			return t, false
		}
		self.accessor = self.ch_accessor
		self.typ = 100
		ok := self.accessor.Next()
		if !ok {
			var t EdgeRef
			return t, false
		}
	}
	edge_id := self.accessor.GetEdgeID()
	other_id := self.accessor.GetOtherID()
	return EdgeRef{
		EdgeID:  edge_id,
		OtherID: other_id,
		_Type:   self.typ,
	}, true
}

type CHUpwardIterator struct {
	this_node   int32
	node_levels Array[int16]
	accessor    *TopologyAccessor
	ch_accessor *TopologyAccessor
	typ         byte
}

func (self *CHUpwardIterator) Next() (EdgeRef, bool) {
	if self.typ == 0 {
		for self.accessor.Next() {
			other_id := self.accessor.GetOtherID()
			if self.node_levels[self.this_node] >= self.node_levels[other_id] {
				continue
			}
			edge_id := self.accessor.GetEdgeID()
			return EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				_Type:   0,
			}, true
		}
		self.typ = 1
	}
	if self.typ == 1 {
		for self.ch_accessor.Next() {
			other_id := self.ch_accessor.GetOtherID()
			if self.node_levels[self.this_node] >= self.node_levels[other_id] {
				continue
			}
			edge_id := self.ch_accessor.GetEdgeID()
			return EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				_Type:   100,
			}, true
		}
		self.typ = 2
	}
	return EdgeRef{}, false
}

type CHDownwardIterator struct {
	this_node   int32
	node_levels Array[int16]
	accessor    *TopologyAccessor
	ch_accessor *TopologyAccessor
	typ         byte
}

func (self *CHDownwardIterator) Next() (EdgeRef, bool) {
	if self.typ == 0 {
		for self.accessor.Next() {
			other_id := self.accessor.GetOtherID()
			if self.node_levels[self.this_node] <= self.node_levels[other_id] {
				continue
			}
			edge_id := self.accessor.GetEdgeID()
			return EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				_Type:   0,
			}, true
		}
		self.typ = 1
	}
	if self.typ == 1 {
		for self.ch_accessor.Next() {
			other_id := self.ch_accessor.GetOtherID()
			if self.node_levels[self.this_node] <= self.node_levels[other_id] {
				continue
			}
			edge_id := self.ch_accessor.GetEdgeID()
			return EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				_Type:   100,
			}, true
		}
		self.typ = 2
	}
	return EdgeRef{}, false
}

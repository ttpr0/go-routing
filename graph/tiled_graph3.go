package graph

import (
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type ITiledGraph3 interface {
	GetGeometry() IGeometry
	GetWeighting() IWeighting
	GetDefaultExplorer() IGraphExplorer
	GetGraphExplorer(weighting IWeighting) IGraphExplorer
	GetNodeTile(node int32) int16
	NodeCount() int32
	EdgeCount() int32
	TileCount() int16
	IsNode(node int32) bool
	GetNode(node int32) Node
	GetEdge(edge int32) Edge
	GetEdgesFromShortcut(edges *List[int32], shortcut_id int32)
	GetIndex() IGraphIndex
}

type TiledGraph3 struct {
	nodes           NodeStore
	node_tiles      NodeTileStore
	topology        TopologyStore
	edges           EdgeStore
	border_topology TopologyStore
	skip_topology   TopologyStore
	skip_edges      ShortcutStore
	skip_weights    DefaultWeighting
	edge_types      Array[byte]
	geom            GeometryStore
	weight          DefaultWeighting
	index           KDTree[int32]
}

func (self *TiledGraph3) GetGeometry() IGeometry {
	return &self.geom
}
func (self *TiledGraph3) GetWeighting() IWeighting {
	return &self.weight
}
func (self *TiledGraph3) GetDefaultExplorer() IGraphExplorer {
	return &TiledGraph3Explorer{
		graph:           self,
		accessor:        self.topology.GetAccessor(),
		skip_accessor:   self.skip_topology.GetAccessor(),
		border_accessor: self.border_topology.GetAccessor(),
		weight:          &self.weight,
		skip_weight:     &self.skip_weights,
	}
}
func (self *TiledGraph3) GetGraphExplorer(weighting IWeighting) IGraphExplorer {
	return &TiledGraph3Explorer{
		graph:    self,
		accessor: self.topology.GetAccessor(),
		weight:   weighting,
	}
}
func (self *TiledGraph3) GetNodeTile(node int32) int16 {
	return self.node_tiles.GetNodeTile(node)
}
func (self *TiledGraph3) NodeCount() int32 {
	return int32(self.nodes.NodeCount())
}
func (self *TiledGraph3) EdgeCount() int32 {
	return int32(self.edges.EdgeCount())
}
func (self *TiledGraph3) TileCount() int16 {
	max := int16(0)
	for i := 0; i < int(self.NodeCount()); i++ {
		tile := self.node_tiles.GetNodeTile(int32(i))
		if tile > max {
			max = tile
		}
	}
	return max - 1
}
func (self *TiledGraph3) IsNode(node int32) bool {
	return self.nodes.IsNode(node)
}
func (self *TiledGraph3) GetNode(node int32) Node {
	return self.nodes.GetNode(node)
}
func (self *TiledGraph3) GetEdge(edge int32) Edge {
	return self.edges.GetEdge(edge)
}
func (self *TiledGraph3) GetEdgesFromShortcut(edges *List[int32], shortcut_id int32) {
	for _, ref := range self.skip_edges.GetEdgesFromShortcut(shortcut_id) {
		edges.Add(ref.A)
	}
}
func (self *TiledGraph3) GetIndex() IGraphIndex {
	return &BaseGraphIndex{
		index: self.index,
	}
}

type TiledGraph3Explorer struct {
	graph           *TiledGraph3
	accessor        TopologyAccessor
	skip_accessor   TopologyAccessor
	border_accessor TopologyAccessor
	weight          IWeighting
	skip_weight     IWeighting
}

func (self *TiledGraph3Explorer) GetAdjacentEdges(node int32, direction Direction, typ Adjacency) IIterator[EdgeRef] {
	if typ == ADJACENT_SKIP {
		self.skip_accessor.SetBaseNode(node, direction)
		self.border_accessor.SetBaseNode(node, direction)
		return &SkipEdgeRefIterator{
			skip_accessor: &self.skip_accessor,
			accessor:      &self.border_accessor,
			typ:           10,
		}
	} else {
		self.accessor.SetBaseNode(node, direction)
		return &TiledEdgeRefIterator{
			accessor:   &self.accessor,
			edge_types: self.graph.edge_types,
		}
	}
}
func (self *TiledGraph3Explorer) GetEdgeWeight(edge EdgeRef) int32 {
	if edge.IsShortcut() {
		return self.skip_weight.GetEdgeWeight(edge.EdgeID)
	} else {
		return self.weight.GetEdgeWeight(edge.EdgeID)
	}
}
func (self *TiledGraph3Explorer) GetTurnCost(from EdgeRef, via int32, to EdgeRef) int32 {
	return 0
}
func (self *TiledGraph3Explorer) GetOtherNode(edge EdgeRef, node int32) int32 {
	if edge.IsShortcut() {
		e := self.graph.skip_edges.GetShortcut(edge.EdgeID)
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

type SkipEdgeRefIterator struct {
	accessor      *TopologyAccessor
	skip_accessor *TopologyAccessor
	typ           byte
}

func (self *SkipEdgeRefIterator) Next() (EdgeRef, bool) {
	ok := self.accessor.Next()
	if !ok {
		if self.typ == 101 {
			var t EdgeRef
			return t, false
		}
		self.accessor = self.skip_accessor
		self.typ = 101
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

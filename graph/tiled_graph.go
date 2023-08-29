package graph

import (
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type ITiledGraph interface {
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
	GetShortcut(shc int32) Shortcut
	GetEdgesFromShortcut(edges *List[int32], shortcut_id int32)
	GetIndex() IGraphIndex
}

type TiledGraph struct {
	nodes         NodeStore
	node_tiles    NodeTileStore
	topology      TopologyStore
	edges         EdgeStore
	skip_topology TypedTopologyStore
	skip_edges    ShortcutStore
	skip_weights  DefaultWeighting
	edge_types    Array[byte]
	geom          GeometryStore
	weight        DefaultWeighting
	index         KDTree[int32]
}

func (self *TiledGraph) GetGeometry() IGeometry {
	return &self.geom
}
func (self *TiledGraph) GetWeighting() IWeighting {
	return &self.weight
}
func (self *TiledGraph) GetDefaultExplorer() IGraphExplorer {
	return &TiledGraphExplorer{
		graph:         self,
		accessor:      self.topology.GetAccessor(),
		skip_accessor: self.skip_topology.GetAccessor(),
		weight:        &self.weight,
		skip_weight:   &self.skip_weights,
	}
}
func (self *TiledGraph) GetGraphExplorer(weighting IWeighting) IGraphExplorer {
	return &TiledGraphExplorer{
		graph:         self,
		accessor:      self.topology.GetAccessor(),
		skip_accessor: self.skip_topology.GetAccessor(),
		weight:        weighting,
		skip_weight:   &self.skip_weights,
	}
}
func (self *TiledGraph) GetNodeTile(node int32) int16 {
	return self.node_tiles.GetNodeTile(node)
}
func (self *TiledGraph) ForEachEdge(node int32, f func(int32)) {

}
func (self *TiledGraph) NodeCount() int32 {
	return int32(self.nodes.NodeCount())
}
func (self *TiledGraph) EdgeCount() int32 {
	return int32(self.edges.EdgeCount())
}
func (self *TiledGraph) TileCount() int16 {
	max := int16(0)
	for i := 0; i < int(self.NodeCount()); i++ {
		tile := self.node_tiles.GetNodeTile(int32(i))
		if tile > max {
			max = tile
		}
	}
	return max - 1
}
func (self *TiledGraph) IsNode(node int32) bool {
	return self.nodes.IsNode(node)
}
func (self *TiledGraph) GetNode(node int32) Node {
	return self.nodes.GetNode(node)
}
func (self *TiledGraph) GetEdge(edge int32) Edge {
	return self.edges.GetEdge(edge)
}
func (self *TiledGraph) GetShortcut(shc int32) Shortcut {
	return self.skip_edges.GetShortcut(shc)
}
func (self *TiledGraph) GetEdgesFromShortcut(edges *List[int32], shortcut_id int32) {
	for _, ref := range self.skip_edges.GetEdgesFromShortcut(shortcut_id, false) {
		edges.Add(ref)
	}
}
func (self *TiledGraph) GetIndex() IGraphIndex {
	return &BaseGraphIndex{
		index: self.index,
	}
}

type TiledGraphExplorer struct {
	graph         *TiledGraph
	accessor      TopologyAccessor
	skip_accessor TypedTopologyAccessor
	weight        IWeighting
	skip_weight   IWeighting
}

func (self *TiledGraphExplorer) GetAdjacentEdges(node int32, direction Direction, typ Adjacency) IIterator[EdgeRef] {
	if typ == ADJACENT_SKIP {
		self.skip_accessor.SetBaseNode(node, direction)
		return &SkipEdgeRefIterator{
			accessor: &self.skip_accessor,
		}
	} else {
		self.accessor.SetBaseNode(node, direction)
		return &TiledEdgeRefIterator{
			accessor:   &self.accessor,
			edge_types: self.graph.edge_types,
		}
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

type TiledEdgeRefIterator struct {
	accessor   *TopologyAccessor
	edge_types Array[byte]
}

func (self *TiledEdgeRefIterator) Next() (EdgeRef, bool) {
	ok := self.accessor.Next()
	if !ok {
		return EdgeRef{}, false
	}
	edge_id := self.accessor.GetEdgeID()
	other_id := self.accessor.GetOtherID()
	typ := self.edge_types[edge_id]
	return EdgeRef{
		EdgeID:  edge_id,
		OtherID: other_id,
		_Type:   typ,
	}, true
}

type SkipEdgeRefIterator struct {
	accessor *TypedTopologyAccessor
}

func (self *SkipEdgeRefIterator) Next() (EdgeRef, bool) {
	ok := self.accessor.Next()
	if !ok {
		return EdgeRef{}, false
	}
	edge_id := self.accessor.GetEdgeID()
	other_id := self.accessor.GetOtherID()
	typ := self.accessor.GetType()
	return EdgeRef{
		EdgeID:  edge_id,
		OtherID: other_id,
		_Type:   typ,
	}, true
}

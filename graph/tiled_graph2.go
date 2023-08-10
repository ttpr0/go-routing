package graph

import (
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type ITiledGraph2 interface {
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
	GetIndex() IGraphIndex
	GetBorderNodes(tile int16) Array[int32]
	GetTileRanges(tile int16, border_node int32) IIterator[Tuple[int32, float32]]
}

type TiledGraph2 struct {
	nodes            NodeStore
	node_tiles       NodeTileStore
	topology         TopologyStore
	edges            EdgeStore
	geom             GeometryStore
	weight           DefaultWeighting
	index            KDTree[int32]
	border_nodes     Dict[int16, Array[int32]]
	interior_nodes   Dict[int16, Array[int32]]
	border_range_map Dict[int16, Dict[int32, Array[float32]]]
}

func (self *TiledGraph2) GetGeometry() IGeometry {
	return &self.geom
}
func (self *TiledGraph2) GetWeighting() IWeighting {
	return &self.weight
}
func (self *TiledGraph2) GetDefaultExplorer() IGraphExplorer {
	return &TiledGraph2Explorer{
		graph:  self,
		weight: &self.weight,
	}
}
func (self *TiledGraph2) GetGraphExplorer(weighting IWeighting) IGraphExplorer {
	return &TiledGraph2Explorer{
		graph:  self,
		weight: weighting,
	}
}
func (self *TiledGraph2) GetNodeTile(node int32) int16 {
	return self.node_tiles.GetNodeTile(node)
}
func (self *TiledGraph2) NodeCount() int32 {
	return int32(self.nodes.NodeCount())
}
func (self *TiledGraph2) EdgeCount() int32 {
	return int32(self.edges.EdgeCount())
}
func (self *TiledGraph2) TileCount() int16 {
	max := int16(0)
	for i := 0; i < int(self.NodeCount()); i++ {
		tile := self.node_tiles.GetNodeTile(int32(i))
		if tile > max {
			max = tile
		}
	}
	return max - 1
}
func (self *TiledGraph2) IsNode(node int32) bool {
	return self.nodes.IsNode(node)
}
func (self *TiledGraph2) GetNode(node int32) Node {
	return self.nodes.GetNode(node)
}
func (self *TiledGraph2) GetEdge(edge int32) Edge {
	return self.edges.GetEdge(edge)
}
func (self *TiledGraph2) GetIndex() IGraphIndex {
	return &BaseGraphIndex{
		index: self.index,
	}
}
func (self *TiledGraph2) GetBorderNodes(tile int16) Array[int32] {
	return self.border_nodes[tile]
}
func (self *TiledGraph2) GetTileRanges(tile int16, border_node int32) IIterator[Tuple[int32, float32]] {
	nodes := self.interior_nodes[tile]
	ranges := self.border_range_map[tile][border_node]
	return &BorderRangeIterator{0, nodes, ranges}
}

type BorderRangeIterator struct {
	state  int
	nodes  Array[int32]
	ranges Array[float32]
}

func (self *BorderRangeIterator) Next() (Tuple[int32, float32], bool) {
	if self.state == len(self.nodes) {
		var t Tuple[int32, float32]
		return t, false
	}
	node := self.nodes[self.state]
	dist := self.ranges[self.state]
	self.state += 1
	return MakeTuple(node, dist), true
}

type TiledGraph2Explorer struct {
	graph  *TiledGraph2
	weight IWeighting
}

func (self *TiledGraph2Explorer) GetAdjacentEdges(node int32, direction Direction) IIterator[EdgeRef] {
	start, count := self.graph.topology.GetNodeRef(node, direction)
	edge_refs := self.graph.topology.GetEdgeRefs(direction)
	return &EdgeRefIterator{
		state:     int(start),
		end:       int(start) + int(count),
		edge_refs: edge_refs,
	}
}
func (self *TiledGraph2Explorer) GetEdgeWeight(edge EdgeRef) int32 {
	return self.weight.GetEdgeWeight(edge.EdgeID)
}
func (self *TiledGraph2Explorer) GetTurnCost(from EdgeRef, via int32, to EdgeRef) int32 {
	return self.weight.GetTurnCost(from.EdgeID, via, to.EdgeID)
}
func (self *TiledGraph2Explorer) GetOtherNode(edge EdgeRef, node int32) int32 {
	e := self.graph.GetEdge(edge.EdgeID)
	if node == e.NodeA {
		return e.NodeB
	}
	if node == e.NodeB {
		return e.NodeA
	}
	return -1
}

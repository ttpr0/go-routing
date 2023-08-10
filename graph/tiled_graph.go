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
	GetIndex() IGraphIndex
}

type TiledGraph struct {
	nodes      NodeStore
	node_tiles NodeTileStore
	topology   TopologyStore
	edges      EdgeStore
	geom       GeometryStore
	weight     DefaultWeighting
	index      KDTree[int32]
}

func (self *TiledGraph) GetGeometry() IGeometry {
	return &self.geom
}
func (self *TiledGraph) GetWeighting() IWeighting {
	return &self.weight
}
func (self *TiledGraph) GetDefaultExplorer() IGraphExplorer {
	return &TiledGraphExplorer{
		graph:  self,
		weight: &self.weight,
	}
}
func (self *TiledGraph) GetGraphExplorer(weighting IWeighting) IGraphExplorer {
	return &TiledGraphExplorer{
		graph:  self,
		weight: weighting,
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
func (self *TiledGraph) GetIndex() IGraphIndex {
	return &BaseGraphIndex{
		index: self.index,
	}
}

type TiledGraphExplorer struct {
	graph  *TiledGraph
	weight IWeighting
}

func (self *TiledGraphExplorer) GetAdjacentEdges(node int32, direction Direction) IIterator[EdgeRef] {
	start, count := self.graph.topology.GetNodeRef(node, direction)
	edge_refs := self.graph.topology.GetEdgeRefs(direction)
	return &EdgeRefIterator{
		state:     int(start),
		end:       int(start) + int(count),
		edge_refs: edge_refs,
	}
}
func (self *TiledGraphExplorer) GetEdgeWeight(edge EdgeRef) int32 {
	return self.weight.GetEdgeWeight(edge.EdgeID)
}
func (self *TiledGraphExplorer) GetTurnCost(from EdgeRef, via int32, to EdgeRef) int32 {
	return self.weight.GetTurnCost(from.EdgeID, via, to.EdgeID)
}
func (self *TiledGraphExplorer) GetOtherNode(edge EdgeRef, node int32) int32 {
	e := self.graph.GetEdge(edge.EdgeID)
	if node == e.NodeA {
		return e.NodeB
	}
	if node == e.NodeB {
		return e.NodeA
	}
	return -1
}

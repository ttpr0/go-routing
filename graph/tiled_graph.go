package graph

import (
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type ITiledGraph interface {
	GetGeometry() IGeometry
	GetWeighting() IWeighting
	GetOtherNode(edge, node int32) (int32, Direction)
	GetAdjacentEdges(node int32) IIterator[EdgeRef]
	GetNodeTile(node int32) int16
	ForEachEdge(node int32, f func(int32))
	NodeCount() int32
	EdgeCount() int32
	TileCount() int16
	IsNode(node int32) bool
	GetNode(node int32) NodeAttributes
	GetEdge(edge int32) EdgeAttributes
}

type TiledGraph struct {
	nodes           List[Node]
	node_attributes List[NodeAttributes]
	node_tiles      List[int16]
	edge_refs       List[EdgeRef]
	edges           List[Edge]
	edge_attributes List[EdgeAttributes]
	geom            IGeometry
	weight          IWeighting
}

func (self *TiledGraph) GetGeometry() IGeometry {
	return self.geom
}
func (self *TiledGraph) GetWeighting() IWeighting {
	return self.weight
}
func (self *TiledGraph) GetOtherNode(edge, node int32) (int32, Direction) {
	e := self.edges[edge]
	if node == e.NodeA {
		return e.NodeB, FORWARD
	}
	if node == e.NodeB {
		return e.NodeA, BACKWARD
	}
	return -1, 0
}
func (self *TiledGraph) GetAdjacentEdges(node int32) IIterator[EdgeRef] {
	n := self.nodes[node]
	return &EdgeRefIterator{
		state:     int(n.EdgeRefStart),
		end:       int(n.EdgeRefStart) + int(n.EdgeRefCount),
		edge_refs: &self.edge_refs,
	}
}
func (self *TiledGraph) GetNodeTile(node int32) int16 {
	return self.node_tiles[node]
}
func (self *TiledGraph) ForEachEdge(node int32, f func(int32)) {

}
func (self *TiledGraph) NodeCount() int32 {
	return int32(len(self.nodes))
}
func (self *TiledGraph) EdgeCount() int32 {
	return int32(len(self.edges))
}
func (self *TiledGraph) TileCount() int16 {
	max := int16(0)
	for _, tile := range self.node_tiles {
		if tile > max {
			max = tile
		}
	}
	return max - 1
}
func (self *TiledGraph) IsNode(node int32) bool {
	if node < int32(len(self.nodes)) {
		return true
	} else {
		return false
	}
}
func (self *TiledGraph) GetNode(node int32) NodeAttributes {
	return self.node_attributes[node]
}
func (self *TiledGraph) GetEdge(edge int32) EdgeAttributes {
	return self.edge_attributes[edge]
}

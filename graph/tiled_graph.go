package graph

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type ITiledGraph interface {
	GetGeometry() IGeometry
	GetWeighting() IWeighting
	GetOtherNode(edge, node int32) (int32, Direction)
	GetAdjacentEdges(node int32, direction Direction) IIterator[EdgeRef]
	GetNodeTile(node int32) int16
	ForEachEdge(node int32, f func(int32))
	NodeCount() int32
	EdgeCount() int32
	TileCount() int16
	IsNode(node int32) bool
	GetNode(node int32) Node
	GetEdge(edge int32) Edge
	GetNodeIndex() KDTree[int32]
	GetClosestNode(point geo.Coord) (int32, bool)
}

type TiledGraph struct {
	node_refs     List[NodeRef]
	nodes         List[Node]
	node_tiles    List[int16]
	fwd_edge_refs List[EdgeRef]
	bwd_edge_refs List[EdgeRef]
	edges         List[Edge]
	geom          IGeometry
	weight        IWeighting
	index         KDTree[int32]
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
func (self *TiledGraph) GetAdjacentEdges(node int32, direction Direction) IIterator[EdgeRef] {
	n := self.node_refs[node]
	if direction == FORWARD {
		return &EdgeRefIterator{
			state:     int(n.EdgeRefFWDStart),
			end:       int(n.EdgeRefFWDStart) + int(n.EdgeRefFWDCount),
			edge_refs: &self.fwd_edge_refs,
		}
	} else {
		return &EdgeRefIterator{
			state:     int(n.EdgeRefBWDStart),
			end:       int(n.EdgeRefBWDStart) + int(n.EdgeRefBWDCount),
			edge_refs: &self.bwd_edge_refs,
		}
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
func (self *TiledGraph) GetNode(node int32) Node {
	return self.nodes[node]
}
func (self *TiledGraph) GetEdge(edge int32) Edge {
	return self.edges[edge]
}
func (self *TiledGraph) GetNodeIndex() KDTree[int32] {
	return self.index
}
func (self *TiledGraph) GetClosestNode(point geo.Coord) (int32, bool) {
	return self.index.GetClosest(point[:], 0.005)
}

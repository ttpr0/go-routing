package graph

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type ITiledGraph2 interface {
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
	GetBorderNodes(tile int16) Array[int32]
	GetTileRanges(tile int16, border_node int32) IIterator[Tuple[int32, float32]]
}

type TiledGraph2 struct {
	node_refs        List[NodeRef]
	nodes            List[Node]
	node_tiles       List[int16]
	fwd_edge_refs    List[EdgeRef]
	bwd_edge_refs    List[EdgeRef]
	edges            List[Edge]
	geom             IGeometry
	weight           IWeighting
	index            KDTree[int32]
	border_nodes     Dict[int16, Array[int32]]
	interior_nodes   Dict[int16, Array[int32]]
	border_range_map Dict[int16, Dict[int32, Array[float32]]]
}

func (self *TiledGraph2) GetGeometry() IGeometry {
	return self.geom
}
func (self *TiledGraph2) GetWeighting() IWeighting {
	return self.weight
}
func (self *TiledGraph2) GetOtherNode(edge, node int32) (int32, Direction) {
	e := self.edges[edge]
	if node == e.NodeA {
		return e.NodeB, FORWARD
	}
	if node == e.NodeB {
		return e.NodeA, BACKWARD
	}
	return -1, 0
}
func (self *TiledGraph2) GetAdjacentEdges(node int32, direction Direction) IIterator[EdgeRef] {
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
func (self *TiledGraph2) GetNodeTile(node int32) int16 {
	return self.node_tiles[node]
}
func (self *TiledGraph2) ForEachEdge(node int32, f func(int32)) {

}
func (self *TiledGraph2) NodeCount() int32 {
	return int32(len(self.nodes))
}
func (self *TiledGraph2) EdgeCount() int32 {
	return int32(len(self.edges))
}
func (self *TiledGraph2) TileCount() int16 {
	max := int16(0)
	for _, tile := range self.node_tiles {
		if tile > max {
			max = tile
		}
	}
	return max - 1
}
func (self *TiledGraph2) IsNode(node int32) bool {
	if node < int32(len(self.nodes)) {
		return true
	} else {
		return false
	}
}
func (self *TiledGraph2) GetNode(node int32) Node {
	return self.nodes[node]
}
func (self *TiledGraph2) GetEdge(edge int32) Edge {
	return self.edges[edge]
}
func (self *TiledGraph2) GetNodeIndex() KDTree[int32] {
	return self.index
}
func (self *TiledGraph2) GetClosestNode(point geo.Coord) (int32, bool) {
	return self.index.GetClosest(point[:], 0.005)
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
	self.state += 1
	return MakeTuple(self.nodes[self.state], self.ranges[self.state]), true
}

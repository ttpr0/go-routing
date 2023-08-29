package graph

import (
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// tiled-graph with border-interior index
//******************************************

type ITiledGraph2 interface {
	ITiledGraph

	GetBorderNodes(tile int16) Array[int32]
	GetTileRanges(tile int16, border_node int32) IIterator[Tuple[int32, float32]]
}

type TiledGraph2 struct {
	TiledGraph

	// Storage for indexing sp within cells
	border_nodes     Dict[int16, Array[int32]]
	interior_nodes   Dict[int16, Array[int32]]
	border_range_map Dict[int16, Dict[int32, Array[float32]]]
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

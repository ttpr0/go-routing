package comps

import (
	"github.com/ttpr0/go-routing/geo"
	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

// *******************************************
// graph index interface
// *******************************************

type IGraphIndex interface {
	GetClosestNode(point geo.Coord) (int32, bool)
}

//*******************************************
// graph index
//*******************************************

type BaseGraphIndex struct {
	index KDTree[int32]
}

func NewGraphIndex(base IGraphBase) IGraphIndex {
	index := _BuildKDTreeIndex(base)
	return &BaseGraphIndex{
		index: index,
	}
}

func (self *BaseGraphIndex) GetClosestNode(point geo.Coord) (int32, bool) {
	return self.index.GetClosest(point[:], 0.005)
}

type MappedGraphIndex struct {
	id_mapping structs.IDMapping
	index      IGraphIndex
}

func (self *MappedGraphIndex) GetClosestNode(point geo.Coord) (int32, bool) {
	node, ok := self.index.GetClosestNode(point)
	if !ok {
		return node, ok
	}
	mapped_node := self.id_mapping.GetTarget(node)
	return mapped_node, true
}

package graph

import (
	"github.com/ttpr0/go-routing/geo"
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

func NewBaseGraphIndex(base IGraphBase) *BaseGraphIndex {
	index := _BuildKDTreeIndex(base)
	return &BaseGraphIndex{
		index: index,
	}
}

func (self *BaseGraphIndex) GetClosestNode(point geo.Coord) (int32, bool) {
	return self.index.GetClosest(point[:], 0.005)
}

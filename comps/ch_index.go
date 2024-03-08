package comps

import (
	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// ch-data interface
//*******************************************

type ICHIndex interface {
}

//*******************************************
// ch-data
//*******************************************

func NewCHIndex(fwd_down_edges Array[structs.Shortcut], bwd_down_edges Array[structs.Shortcut]) *CHIndex {
	return &CHIndex{
		fwd_down_edges: fwd_down_edges,
		bwd_down_edges: bwd_down_edges,
	}
}

type CHIndex struct {
	fwd_down_edges Array[structs.Shortcut]
	bwd_down_edges Array[structs.Shortcut]
}

func (self *CHIndex) GetFWDDownEdges() Array[structs.Shortcut] {
	return self.fwd_down_edges
}
func (self *CHIndex) GetBWDDownEdges() Array[structs.Shortcut] {
	return self.bwd_down_edges
}

func (self *CHIndex) _ReorderNodes(mapping Array[int32]) {
	panic("not implemented")
}
func (self *CHIndex) _New() *CHIndex {
	return &CHIndex{}
}

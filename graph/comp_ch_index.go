package graph

import (
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

type CHIndex struct {
	fwd_down_edges Array[Shortcut]
	bwd_down_edges Array[Shortcut]
}

func (self *CHIndex) _ReorderNodes(mapping Array[int32]) {
	panic("not implemented")
}
func (self *CHIndex) _New() *CHIndex {
	return &CHIndex{}
}

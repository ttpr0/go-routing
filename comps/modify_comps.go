package comps

import (
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// modification methods
//*******************************************

type IReorderable interface {
	_ReorderNodes(mapping Array[int32])
}

// reorders nodes in graph using mapping
// mapping: old id -> new id
func ReorderNodes(comp IReorderable, mapping Array[int32]) {
	comp._ReorderNodes(mapping)
}

type IModifyable interface {
	_RemoveNodes(nodes List[int32])
	_RemoveEdges(edges List[int32])
}

// removes nodes from nodes-list by id keeping order in tact
//
// also removes all edges (or shortcuts) adjacent to removed nodes
func RemoveNodes(comp IModifyable, nodes List[int32]) {
	comp._RemoveNodes(nodes)
}

// removes edges from edges-list by id keeping order in tact
func RemoveEdges(comp IModifyable, edges List[int32]) {
	comp._RemoveEdges(edges)
}

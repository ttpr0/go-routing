package comps

import (
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// modification methods
//*******************************************

type IReorderable[T any] interface {
	_ReorderNodes(mapping Array[int32]) T
}

// reorders nodes in graph using mapping
// mapping: old id -> new id
func ReorderNodes[T IReorderable[T]](comp T, mapping Array[int32]) T {
	return comp._ReorderNodes(mapping)
}

type IModifyable[T any] interface {
	_RemoveNodes(nodes List[int32]) T
	_RemoveEdges(edges List[int32]) T
}

// removes nodes from nodes-list by id keeping order in tact
//
// also removes all edges (or shortcuts) adjacent to removed nodes
func RemoveNodes[T IModifyable[T]](comp T, nodes List[int32]) T {
	return comp._RemoveNodes(nodes)
}

// removes edges from edges-list by id keeping order in tact
func RemoveEdges[T IModifyable[T]](comp T, edges List[int32]) T {
	return comp._RemoveEdges(edges)
}

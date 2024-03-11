package onetomany

import (
	. "github.com/ttpr0/go-routing/util"
)

type IOneToMany interface {
	CreateSolver() ISolver
}

type ISolver interface {
	// Computes distances from start nodes to all other nodes.
	//
	// Multiple start nodes are specified as (node, initial distance) tuples to account for start locations not identical to graph node locations.
	CalcDistanceFromStart(starts Array[Tuple[int32, int32]]) error

	// Returns the computed distance.
	GetDistance(node int32) int32
}

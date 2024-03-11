package nearest

import (
	. "github.com/ttpr0/go-routing/util"
)

type INearest interface {
	CreateSolver() ISolver
}

type ISolver interface {
	// Computes the nearest neighbour (source node) for all other nodes.
	//
	// Source nodes are specified using an array of (node, initial distance) tuples to account for start locations not identical to graph node locations.
	CalcNearestNeighbours(sources List[Array[Tuple[int32, int32]]]) error

	// Returns the id (in the specified source list) of the nearest neighbour.
	GetNeighbour(node int32) int32

	// Returns the distance to the nearest neighbour.
	GetDistance(node int32) int32
}

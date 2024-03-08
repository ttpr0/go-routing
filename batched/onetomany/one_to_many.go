package onetomany

import (
	. "github.com/ttpr0/go-routing/util"
)

type IOneToMany interface {
	CreateSolver() ISolver
}

type ISolver interface {
	CalcDistanceFromStart(start int32) error
	CalcDistanceFromStarts(starts Array[Tuple[int32, int32]]) error
	GetDistance(node int32) int32
}

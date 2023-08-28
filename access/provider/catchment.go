package provider

import (
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type ICatchment interface {
	GetNeighbours(destination int) IIterator[int]
}

type Catchment struct {
	sources []List[int]
}

func NewCatchment(sources []List[int]) Catchment {
	return Catchment{
		sources: sources,
	}
}

func (self *Catchment) GetNeighbours(destination int) IIterator[int] {
	agg := self.sources[destination]
	if agg == nil {
		return EmptyIterator[int]{}
	}
	return NewListIterator(agg)
}

type EmptyIterator[T any] struct {
}

func (self EmptyIterator[T]) Next() (T, bool) {
	var t T
	return t, false
}

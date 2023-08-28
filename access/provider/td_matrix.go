package provider

import (
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type ITDMatrix interface {
	GetRange(source, destination int) float32
}

type TDMatrix struct {
	durations Matrix[float32]
}

func NewTDMatrix(durations Matrix[float32]) *TDMatrix {
	return &TDMatrix{
		durations: durations,
	}
}

func (self *TDMatrix) GetRange(source, destination int) float32 {
	return self.durations.Get(source, destination)
}

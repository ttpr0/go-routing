package view

import (
	"github.com/ttpr0/go-routing/geo"
	. "github.com/ttpr0/go-routing/util"
)

type IPointView interface {
	GetCoordinate(index int) geo.Coord

	GetWeight(index int) int32

	PointCount() int
}

type PointView struct {
	points  Array[geo.Coord]
	weights Array[int32]
}

func NewPointView(points Array[geo.Coord], weights Array[int32]) *PointView {
	return &PointView{
		points:  points,
		weights: weights,
	}
}

func (self *PointView) GetCoordinate(index int) geo.Coord {
	return self.points[index]
}
func (self *PointView) GetWeight(index int) int32 {
	if self.weights == nil {
		return 1
	}
	return self.weights[index]
}
func (self *PointView) PointCount() int {
	return len(self.points)
}

package view

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type IPointView interface {
	GetCoordinate(index int) geo.Coord

	GetWeight(index int) int32

	PointCount() int

	GetPointsInEnvelop(envelope geo.Envelope) List[int]
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
func (self *PointView) GetPointsInEnvelop(envelope geo.Envelope) List[int] {
	panic("not implemented")
}

type IndexPointView struct {
	index   KDTree[int]
	points  Array[geo.Coord]
	weights Array[int32]
}

func NewIndexPointView(points Array[geo.Coord], weights Array[int32]) *IndexPointView {
	return &IndexPointView{
		points:  points,
		weights: weights,
	}
}

func (self *IndexPointView) GetCoordinate(index int) geo.Coord {
	return self.points[index]
}
func (self *IndexPointView) GetWeight(index int) int32 {
	if self.weights == nil {
		return 1
	}
	return self.weights[index]
}
func (self *IndexPointView) PointCount() int {
	return len(self.points)
}
func (self *IndexPointView) GetPointsInEnvelop(envelope geo.Envelope) List[int] {
	panic("not implemented")
}

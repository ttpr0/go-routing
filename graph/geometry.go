package graph

import "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"

type IGeometry interface {
	GetNode(node int32) Coord
	GetEdge(edge int32) CoordArray
	GetAllNodes() []Coord
	GetAllEdges() []CoordArray
}

type Coord struct {
	Lon float32
	Lat float32
}

type CoordArray []Coord

func (self *CoordArray) GetIterator() util.IIterator[Coord] {
	return &CoordArrayIterator{self, 0}
}

type CoordArrayIterator struct {
	coords *CoordArray
	curr   int
}

func (self *CoordArrayIterator) Next() (Coord, bool) {
	if len(*self.coords) <= self.curr {
		return Coord{}, false
	} else {
		self.curr += 1
		return (*self.coords)[self.curr-1], true
	}
}

type Geometry struct {
	NodeGeometry []Coord
	EdgeGeometry []CoordArray
}

func (self *Geometry) GetNode(node int32) Coord {
	return self.NodeGeometry[node]
}
func (self *Geometry) GetEdge(edge int32) CoordArray {
	return self.EdgeGeometry[edge]
}
func (self *Geometry) GetAllNodes() []Coord {
	return self.NodeGeometry
}
func (self *Geometry) GetAllEdges() []CoordArray {
	return self.EdgeGeometry
}

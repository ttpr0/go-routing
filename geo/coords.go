package geo

import (
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type Coord [2]float32

type CoordArray []Coord

func (self *CoordArray) GetIterator() IIterator[Coord] {
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

type Envelope [4]float32

func (self *Envelope) ContainsCoord(coord Coord) bool {
	return coord[0] > self[0] && coord[1] > self[1] && coord[0] < self[2] && coord[1] < self[3]
}

func (self *Envelope) ContainsEnvelope(envelope Envelope) bool {
	return envelope[0] > self[0] && envelope[1] > self[1] && envelope[2] < self[2] && envelope[3] < self[3]
}

package geo

import (
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type Coord struct {
	Lon float32
	Lat float32
}

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

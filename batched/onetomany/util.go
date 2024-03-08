package onetomany

import (
	. "github.com/ttpr0/go-routing/util"
)

type DistFlag struct {
	Dist int32
}

func (self *DistFlag) GetDist() int32 {
	return self.Dist
}

type PQItem struct {
	item int32
	dist int32
}

type TransitItem struct {
	time      int32
	departure int32
	stop      int32
}

type TransitFlag struct {
	trips List[Tuple[int32, int32]]
}

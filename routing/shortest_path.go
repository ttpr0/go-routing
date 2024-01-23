package routing

import (
	"github.com/ttpr0/go-routing/geo"
	. "github.com/ttpr0/go-routing/util"
)

type IShortestPath interface {
	CalcShortestPath() bool
	Steps(int, *List[geo.CoordArray]) bool
	GetShortestPath() Path
}

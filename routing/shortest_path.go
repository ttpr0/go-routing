package routing

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type IShortestPath interface {
	CalcShortestPath() bool
	Steps(int, *util.List[geo.CoordArray]) bool
	GetShortestPath() Path
}

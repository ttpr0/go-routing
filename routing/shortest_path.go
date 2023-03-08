package routing

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type IShortestPath interface {
	CalcShortestPath() bool
	Steps(int, *util.List[graph.CoordArray]) bool
	GetShortestPath() Path
}

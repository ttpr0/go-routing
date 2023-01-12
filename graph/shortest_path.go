package graph

import "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"

type IShortestPath interface {
	CalcShortestPath() bool
	Steps(int, *util.List[CoordArray]) bool
	GetShortestPath() Path
}

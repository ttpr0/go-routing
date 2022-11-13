package graph

type IShortestPath interface {
	CalcShortestPath() bool
	GetShortestPath() Path
}

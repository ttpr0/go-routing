package routing

type IShortestPath interface {
	CalcShortestPath() bool
	Steps(int, func(int32)) bool
	GetShortestPath() Path
}

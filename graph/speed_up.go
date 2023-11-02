package graph

import (
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// speed-up data
//*******************************************

type SpeedUpType byte

const (
	CH    SpeedUpType = 0
	TILED SpeedUpType = 1
)

type ISpeedUpData interface {
	Type() SpeedUpType
}

type ISpeedUpHandler interface {
	Load(dir string, name string) ISpeedUpData
	Store(dir string, name string, data ISpeedUpData)
	Remove(dir string, name string)
	_ReorderNodes(dir string, name string, mapping Array[int32], typ ReorderType)  // Reorder nodes of base-graph
	_ReorderNodesInplace(data ISpeedUpData, mapping Array[int32], typ ReorderType) // Reorder nodes of base-graph
}

var SPEEDUP_HANDLERS = Dict[SpeedUpType, ISpeedUpHandler]{
	CH:    _CHDataHandler{},
	TILED: _TiledDataHandler{},
}

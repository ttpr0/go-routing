package graph

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// enums
//*******************************************

type Direction byte

const (
	BACKWARD Direction = 0
	FORWARD  Direction = 1
)

type RoadType int8

const (
	MOTORWAY       RoadType = 1
	MOTORWAY_LINK  RoadType = 2
	TRUNK          RoadType = 3
	TRUNK_LINK     RoadType = 4
	PRIMARY        RoadType = 5
	PRIMARY_LINK   RoadType = 6
	SECONDARY      RoadType = 7
	SECONDARY_LINK RoadType = 8
	TERTIARY       RoadType = 9
	TERTIARY_LINK  RoadType = 10
	RESIDENTIAL    RoadType = 11
	LIVING_STREET  RoadType = 12
	UNCLASSIFIED   RoadType = 13
	ROAD           RoadType = 14
	TRACK          RoadType = 15
)

//*******************************************
// parser structs
//*******************************************

type TempNode struct {
	Point geo.Coord
	Count int32
}
type OSMNode struct {
	Point geo.Coord
	Type  int32
	Edges List[int32]
}
type OSMEdge struct {
	NodeA     int
	NodeB     int
	Oneway    bool
	Type      RoadType
	Templimit int32
	Length    float32
	Weight    float32
	Nodes     List[geo.Coord]
}

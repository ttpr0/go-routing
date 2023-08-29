package graph

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

type Adjacency byte

const (
	ADJACENT_EDGES     Adjacency = 0
	ADJACENT_SHORTCUTS Adjacency = 1
	ADJACENT_ALL       Adjacency = 2
	ADJACENT_SKIP      Adjacency = 3
	ADJACENT_UPWARDS   Adjacency = 4
	ADJACENT_DOWNWARDS Adjacency = 5
)

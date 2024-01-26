package graph

//*******************************************
// enums
//*******************************************

type Direction byte

const (
	BACKWARD Direction = 0
	FORWARD  Direction = 1
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

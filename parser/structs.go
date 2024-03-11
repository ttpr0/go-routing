package parser

import (
	"github.com/ttpr0/go-routing/attr"
	"github.com/ttpr0/go-routing/geo"
	. "github.com/ttpr0/go-routing/util"
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
	Attr  attr.NodeAttribs
	Edges List[int32]
}
type OSMEdge struct {
	NodeA int
	NodeB int
	Attr  attr.EdgeAttribs
	Nodes List[geo.Coord]
}

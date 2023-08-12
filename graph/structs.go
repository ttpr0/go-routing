package graph

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// graph structs
//*******************************************

type Edge struct {
	NodeA    int32
	NodeB    int32
	Type     RoadType
	Length   float32
	Maxspeed byte
	Oneway   bool
}

type Node struct {
	Type int8
}

type Shortcut struct {
	NodeA         int32
	NodeB         int32
	_EdgeRefStart int32
	_EdgeRefCount int16
}

type CHShortcut struct {
	NodeA  int32
	NodeB  int32
	_Edges [2]Tuple[int32, byte]
}

type NodeRef struct {
	EdgeRefFWDStart int32
	EdgeRefFWDCount int16
	EdgeRefBWDStart int32
	EdgeRefBWDCount int16
}

type EdgeRef struct {
	EdgeID  int32
	_Type   byte
	OtherID int32
}

func (self EdgeRef) IsEdge() bool {
	return self._Type < 100
}
func (self EdgeRef) IsCrossBorder() bool {
	return self._Type == 10
}
func (self EdgeRef) IsSkip() bool {
	return self._Type == 20
}
func (self EdgeRef) IsShortcut() bool {
	return self._Type >= 100
}
func (self EdgeRef) IsCHShortcut() bool {
	return self._Type == 100
}

func CreateEdgeRef(edge int32) EdgeRef {
	return EdgeRef{
		EdgeID:  edge,
		_Type:   0,
		OtherID: -1,
	}
}
func CreateCHShortcutRef(edge int32) EdgeRef {
	return EdgeRef{
		EdgeID:  edge,
		_Type:   100,
		OtherID: -1,
	}
}

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

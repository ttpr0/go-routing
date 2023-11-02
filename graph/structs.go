package graph

import (
	"unsafe"

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

//*******************************************
// edgeref struct
//*******************************************

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
// shortcut struct
//*******************************************

func NewShortcut(from, to, weight int32) Shortcut {
	return Shortcut{
		From:   from,
		To:     to,
		Weight: weight,
	}
}

type Shortcut struct {
	From     int32
	To       int32
	Weight   int32
	_payload [4]byte
}

// Payload size is 4 bytes.
//
// Be carefull, this method is unsafe.
func Shortcut_set_payload[T int32 | int16 | int8 | uint32 | uint16 | uint8 | bool](edge *Shortcut, value T, pos int) {
	*(*T)(unsafe.Pointer(&edge._payload[pos])) = value
}

// Payload size is 4 bytes.
//
// Be carefull, this method is unsafe.
func Shortcut_get_payload[T int32 | int16 | int8 | uint32 | uint16 | uint8 | bool](edge *Shortcut, pos int) T {
	return *(*T)(unsafe.Pointer(&edge._payload[pos]))
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

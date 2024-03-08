package structs

import (
	"unsafe"

	"github.com/ttpr0/go-routing/geo"
)

//*******************************************
// graph structs
//*******************************************

type Edge struct {
	NodeA int32
	NodeB int32
}

type Node struct {
	Loc geo.Coord
}

type Connection struct {
	StopA   int32
	StopB   int32
	RouteID int32
}

//*******************************************
// shortcut struct
//*******************************************

type Shortcut struct {
	From    int32
	To      int32
	Weight  int32
	Payload [4]byte
}

func NewShortcut(from, to, weight int32) Shortcut {
	return Shortcut{
		From:   from,
		To:     to,
		Weight: weight,
	}
}

// Payload size is 4 bytes.
//
// Be carefull, this method is unsafe.
func Shortcut_set_payload[T int32 | int16 | int8 | uint32 | uint16 | uint8 | bool](edge *Shortcut, value T, pos int) {
	*(*T)(unsafe.Pointer(&edge.Payload[pos])) = value
}

// Payload size is 4 bytes.
//
// Be carefull, this method is unsafe.
func Shortcut_get_payload[T int32 | int16 | int8 | uint32 | uint16 | uint8 | bool](edge *Shortcut, pos int) T {
	return *(*T)(unsafe.Pointer(&edge.Payload[pos]))
}

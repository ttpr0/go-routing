package comps

import (
	"errors"
	"os"

	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// tiled-graph cell-index
//*******************************************

func NewCellIndex() CellIndex {
	return CellIndex{
		fwd_index_edges: NewDict[int16, Array[structs.Shortcut]](100),
		bwd_index_edges: NewDict[int16, Array[structs.Shortcut]](100),
	}
}

type CellIndex struct {
	fwd_index_edges Dict[int16, Array[structs.Shortcut]]
	bwd_index_edges Dict[int16, Array[structs.Shortcut]]
}

func (self *CellIndex) GetFWDIndexEdges(tile int16) Array[structs.Shortcut] {
	return self.fwd_index_edges[tile]
}
func (self *CellIndex) GetBWDIndexEdges(tile int16) Array[structs.Shortcut] {
	return self.bwd_index_edges[tile]
}
func (self *CellIndex) SetFWDIndexEdges(tile int16, edges Array[structs.Shortcut]) {
	self.fwd_index_edges[tile] = edges
}
func (self *CellIndex) SetBWDIndexEdges(tile int16, edges Array[structs.Shortcut]) {
	self.bwd_index_edges[tile] = edges
}

func (self *CellIndex) _ReorderNodes(mapping Array[int32]) *CellIndex {
	new_fwd := NewDict[int16, Array[structs.Shortcut]](100)
	for tile, edges := range self.fwd_index_edges {
		new_edges := NewArray[structs.Shortcut](edges.Length())
		for i, edge := range edges {
			new_edges[i] = structs.Shortcut{
				From:    mapping[edge.From],
				To:      mapping[edge.To],
				Weight:  edge.Weight,
				Payload: edge.Payload,
			}
		}
		new_fwd[tile] = new_edges
	}
	new_bwd := NewDict[int16, Array[structs.Shortcut]](100)
	for tile, edges := range self.bwd_index_edges {
		new_edges := NewArray[structs.Shortcut](edges.Length())
		for i, edge := range edges {
			new_edges[i] = structs.Shortcut{
				From:    mapping[edge.From],
				To:      mapping[edge.To],
				Weight:  edge.Weight,
				Payload: edge.Payload,
			}
		}
		new_bwd[tile] = new_edges
	}

	return &CellIndex{
		fwd_index_edges: new_fwd,
		bwd_index_edges: new_bwd,
	}
}
func (self *CellIndex) _Store(path string) {
	filename := path + "-tileranges"
	writer := NewBufferWriter()

	fwd_tilecount := self.fwd_index_edges.Length()
	Write[int32](writer, int32(fwd_tilecount))
	bwd_tilecount := self.bwd_index_edges.Length()
	Write[int32](writer, int32(bwd_tilecount))

	for tile, edges := range self.fwd_index_edges {
		Write[int16](writer, tile)
		Write[int32](writer, int32(edges.Length()))
		for _, edge := range edges {
			Write[int32](writer, edge.From)
			Write[int32](writer, edge.To)
			Write[int32](writer, edge.Weight)
			// Write[[4]byte](writer, edge._payload)
		}
	}
	for tile, edges := range self.bwd_index_edges {
		Write[int16](writer, tile)
		Write[int32](writer, int32(edges.Length()))
		for _, edge := range edges {
			Write[int32](writer, edge.From)
			Write[int32](writer, edge.To)
			Write[int32](writer, edge.Weight)
			// Write[[4]byte](writer, edge._payload)
		}
	}

	rangesfile, _ := os.Create(filename)
	defer rangesfile.Close()
	rangesfile.Write(writer.Bytes())
}
func (self *CellIndex) _New() *CellIndex {
	return &CellIndex{}
}
func (self *CellIndex) _Load(path string) {
	file := path + "-tileranges"
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	tiledata, _ := os.ReadFile(file)
	reader := NewBufferReader(tiledata)

	fwd_tilecount := Read[int32](reader)
	bwd_tilecount := Read[int32](reader)

	fwd_index_edges := NewDict[int16, Array[structs.Shortcut]](int(fwd_tilecount))
	for i := 0; i < int(fwd_tilecount); i++ {
		tile := Read[int16](reader)
		count := Read[int32](reader)
		edges := NewArray[structs.Shortcut](int(count))
		for j := 0; j < int(count); j++ {
			edge := structs.Shortcut{}
			edge.From = Read[int32](reader)
			edge.To = Read[int32](reader)
			edge.Weight = Read[int32](reader)
			// edge.Payload = Read[[4]byte](reader)
			edges[j] = edge
		}
		fwd_index_edges[tile] = edges
	}
	bwd_index_edges := NewDict[int16, Array[structs.Shortcut]](int(bwd_tilecount))
	for i := 0; i < int(bwd_tilecount); i++ {
		tile := Read[int16](reader)
		count := Read[int32](reader)
		edges := NewArray[structs.Shortcut](int(count))
		for j := 0; j < int(count); j++ {
			edge := structs.Shortcut{}
			edge.From = Read[int32](reader)
			edge.To = Read[int32](reader)
			edge.Weight = Read[int32](reader)
			// edge._payload = Read[[4]byte](reader)
			edges[j] = edge
		}
		bwd_index_edges[tile] = edges
	}

	*self = CellIndex{
		fwd_index_edges: fwd_index_edges,
		bwd_index_edges: bwd_index_edges,
	}
}
func (self *CellIndex) _Remove(path string) {
	os.Remove(path + "-tileranges")
}

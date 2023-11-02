package graph

import (
	"errors"
	"os"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// tiled-graph cell-index
//*******************************************

func _NewCellIndex() _CellIndex {
	return _CellIndex{
		fwd_index_edges: NewDict[int16, Array[Shortcut]](100),
		bwd_index_edges: NewDict[int16, Array[Shortcut]](100),
	}
}

type _CellIndex struct {
	fwd_index_edges Dict[int16, Array[Shortcut]]
	bwd_index_edges Dict[int16, Array[Shortcut]]
}

func (self *_CellIndex) GetFWDIndexEdges(tile int16) Array[Shortcut] {
	return self.fwd_index_edges[tile]
}
func (self *_CellIndex) GetBWDIndexEdges(tile int16) Array[Shortcut] {
	return self.bwd_index_edges[tile]
}
func (self *_CellIndex) SetFWDIndexEdges(tile int16, edges Array[Shortcut]) {
	self.fwd_index_edges[tile] = edges
}
func (self *_CellIndex) SetBWDIndexEdges(tile int16, edges Array[Shortcut]) {
	self.bwd_index_edges[tile] = edges
}

func (self *_CellIndex) _ReorderNodes(mapping Array[int32]) {
	panic("not implemented")
}

//*******************************************
// tiled-graph cell-index io
//*******************************************

func _StoreCellIndex(index _CellIndex, filename string) {
	writer := NewBufferWriter()

	fwd_tilecount := index.fwd_index_edges.Length()
	Write[int32](writer, int32(fwd_tilecount))
	bwd_tilecount := index.bwd_index_edges.Length()
	Write[int32](writer, int32(bwd_tilecount))

	for tile, edges := range index.fwd_index_edges {
		Write[int16](writer, tile)
		Write[int32](writer, int32(edges.Length()))
		for _, edge := range edges {
			Write[int32](writer, edge.From)
			Write[int32](writer, edge.To)
			Write[int32](writer, edge.Weight)
			// Write[[4]byte](writer, edge._payload)
		}
	}
	for tile, edges := range index.bwd_index_edges {
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

func _LoadCellIndex(file string) _CellIndex {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	tiledata, _ := os.ReadFile(file)
	reader := NewBufferReader(tiledata)

	fwd_tilecount := Read[int32](reader)
	bwd_tilecount := Read[int32](reader)

	fwd_index_edges := NewDict[int16, Array[Shortcut]](int(fwd_tilecount))
	for i := 0; i < int(fwd_tilecount); i++ {
		tile := Read[int16](reader)
		count := Read[int32](reader)
		edges := NewArray[Shortcut](int(count))
		for j := 0; i < int(count); j++ {
			edge := Shortcut{}
			edge.From = Read[int32](reader)
			edge.To = Read[int32](reader)
			edge.Weight = Read[int32](reader)
			edge._payload = Read[[4]byte](reader)
			edges[j] = edge
		}
		fwd_index_edges[tile] = edges
	}
	bwd_index_edges := NewDict[int16, Array[Shortcut]](int(bwd_tilecount))
	for i := 0; i < int(bwd_tilecount); i++ {
		tile := Read[int16](reader)
		count := Read[int32](reader)
		edges := NewArray[Shortcut](int(count))
		for j := 0; i < int(count); j++ {
			edge := Shortcut{}
			edge.From = Read[int32](reader)
			edge.To = Read[int32](reader)
			edge.Weight = Read[int32](reader)
			// edge._payload = Read[[4]byte](reader)
			edges[j] = edge
		}
		bwd_index_edges[tile] = edges
	}

	return _CellIndex{
		fwd_index_edges: fwd_index_edges,
		bwd_index_edges: bwd_index_edges,
	}
}

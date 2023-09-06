package graph

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// tiled-graph with border-interior index
//******************************************

type TiledGraph3 struct {
	TiledGraph

	// Storage for indexing sp within cells
	tile_ranges Dict[int16, Tuple[int32, int32]]
	index_edges Array[TiledSHEdge]
}

func (self *TiledGraph3) GetDownEdges(tile int16, dir Direction) Array[TiledSHEdge] {
	tr := self.tile_ranges[tile]
	return self.index_edges[tr.A : tr.A+tr.B]
}

type TiledSHEdge struct {
	From   int32
	To     int32
	Weight int32
}

func _StoreTileRanges2(tile_ranges Dict[int16, Tuple[int32, int32]], index_edges Array[TiledSHEdge], filename string) {
	tilebuffer := bytes.Buffer{}

	tilecount := tile_ranges.Length()
	binary.Write(&tilebuffer, binary.LittleEndian, int32(tilecount))
	edgecount := index_edges.Length()
	binary.Write(&tilebuffer, binary.LittleEndian, int32(edgecount))

	for tile, ran := range tile_ranges {
		binary.Write(&tilebuffer, binary.LittleEndian, tile)
		binary.Write(&tilebuffer, binary.LittleEndian, ran.A)
		binary.Write(&tilebuffer, binary.LittleEndian, ran.B)
	}

	for _, edge := range index_edges {
		binary.Write(&tilebuffer, binary.LittleEndian, edge.From)
		binary.Write(&tilebuffer, binary.LittleEndian, edge.To)
		binary.Write(&tilebuffer, binary.LittleEndian, edge.Weight)
	}

	rangesfile, _ := os.Create(filename)
	defer rangesfile.Close()
	rangesfile.Write(tilebuffer.Bytes())
}

func _LoadTileRanges2(file string) (Dict[int16, Tuple[int32, int32]], Array[TiledSHEdge]) {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	tiledata, _ := os.ReadFile(file)
	tilereader := bytes.NewReader(tiledata)
	var tilecount int32
	binary.Read(tilereader, binary.LittleEndian, &tilecount)
	var edgecount int32
	binary.Read(tilereader, binary.LittleEndian, &edgecount)

	tile_ranges := NewDict[int16, Tuple[int32, int32]](int(tilecount))
	index_edges := NewList[TiledSHEdge](int(edgecount))

	for i := 0; i < int(tilecount); i++ {
		var tile int16
		binary.Read(tilereader, binary.LittleEndian, &tile)
		var s int32
		binary.Read(tilereader, binary.LittleEndian, &s)
		var c int32
		binary.Read(tilereader, binary.LittleEndian, &c)
		tile_ranges[tile] = MakeTuple(s, c)
	}
	for i := 0; i < int(edgecount); i++ {
		var from int32
		binary.Read(tilereader, binary.LittleEndian, &from)
		var to int32
		binary.Read(tilereader, binary.LittleEndian, &to)
		var weight int32
		binary.Read(tilereader, binary.LittleEndian, &weight)
		index_edges.Add(TiledSHEdge{
			From:   from,
			To:     to,
			Weight: weight,
		})
	}

	return tile_ranges, Array[TiledSHEdge](index_edges)
}

func TransformToTiled3(graph *TiledGraph) *TiledGraph3 {
	tiles := GetTiles(graph)
	index_edges := NewList[TiledSHEdge](100)
	tile_ranges := NewDict[int16, Tuple[int32, int32]](tiles.Length())
	for index, tile := range tiles {
		fmt.Println("Process Tile:", index, "/", len(tiles))
		start := index_edges.Length()
		count := 0
		b_nodes, i_nodes := GetBorderNodes(graph, tile)
		flags := NewDict[int32, _Flag](100)
		for _, b_node := range b_nodes {
			flags.Clear()
			CalcFullSPT(graph, b_node, flags)
			for _, i_node := range i_nodes {
				if flags.ContainsKey(i_node) {
					flag := flags[i_node]
					index_edges.Add(TiledSHEdge{
						From:   b_node,
						To:     i_node,
						Weight: flag.pathlength,
					})
					count += 1
				}
			}
		}
		tile_ranges[tile] = MakeTuple(int32(start), int32(count))
	}

	return &TiledGraph3{
		TiledGraph:  *graph,
		tile_ranges: tile_ranges,
		index_edges: Array[TiledSHEdge](index_edges),
	}
}

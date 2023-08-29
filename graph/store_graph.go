package graph

import (
	"bytes"
	"encoding/binary"
	"os"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// graph io
//*******************************************

func StoreGraph(graph *Graph, filename string) {
	graph.topology._Store(filename + "-graph")
	_StoreGraphStorage(graph.store, filename)
	_StoreDefaultWeighting(&graph.weight, filename+"-fastest_weighting")

	// tc_weight := _CreateTCWeighting(graph)
	// _StoreTCWeighting(tc_weight, filename+"-tc_weighting")
}

func StoreTiledGraph(graph *TiledGraph, filename string) {
	graph.topology._Store(filename + "-graph")
	_StoreGraphStorage(graph.store, filename)
	_StoreTypedTopology(&graph.skip_topology, filename+"-skip_topology")
	_StoreDefaultWeighting(&graph.weight, filename+"-fastest_weighting")
	_StoreTiledStorage(graph.skip_store, filename)
}

func StoreTiledGraph2(graph *TiledGraph2, filename string) {
	graph.topology._Store(filename + "-graph")
	_StoreGraphStorage(graph.store, filename)
	_StoreDefaultWeighting(&graph.weight, filename+"-fastest_weighting")
	_StoreTiledStorage(graph.skip_store, filename)
	_StoreTileRanges(graph.border_nodes, graph.interior_nodes, graph.border_range_map, filename+"-tileranges")
}

func StoreCHGraph(graph *CHGraph, filename string) {
	graph.topology._Store(filename + "-graph")
	_StoreGraphStorage(graph.store, filename)
	graph.ch_topology._Store(filename + "-ch_graph")
	_StoreDefaultWeighting(&graph.weight, filename+"-fastest_weighting")
	_StoreCHStorage(graph.ch_store, filename)
}

//*******************************************
// store graph information
//*******************************************

func _StoreTileRanges(border_nodes Dict[int16, Array[int32]], interior_nodes Dict[int16, Array[int32]], border_range_map Dict[int16, Dict[int32, Array[float32]]], filename string) {
	tilebuffer := bytes.Buffer{}

	tilecount := len(border_nodes)
	binary.Write(&tilebuffer, binary.LittleEndian, int32(tilecount))

	for tile, b_nodes := range border_nodes {
		binary.Write(&tilebuffer, binary.LittleEndian, tile)
		binary.Write(&tilebuffer, binary.LittleEndian, int32(len(b_nodes)))
		for _, node := range b_nodes {
			binary.Write(&tilebuffer, binary.LittleEndian, node)
		}
		i_nodes := interior_nodes[tile]
		binary.Write(&tilebuffer, binary.LittleEndian, int32(len(i_nodes)))
		for _, node := range i_nodes {
			binary.Write(&tilebuffer, binary.LittleEndian, node)
		}
		range_map := border_range_map[tile]
		for _, node := range b_nodes {
			ranges := range_map[node]
			for _, dist := range ranges {
				binary.Write(&tilebuffer, binary.LittleEndian, dist)
			}
		}
	}

	rangesfile, _ := os.Create(filename)
	defer rangesfile.Close()
	rangesfile.Write(tilebuffer.Bytes())
}

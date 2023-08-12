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
	graph.nodes._Store(filename + "-nodes")
	graph.edges._Store(filename + "-edges")
	graph.geom._Store(filename + "-geom")
	_StoreDefaultWeighting(&graph.weight, filename+"-fastest_weighting")

	tc_weight := _CreateTCWeighting(graph)
	_StoreTCWeighting(tc_weight, filename+"-tc_weighting")

}

func StoreTiledGraph(graph *TiledGraph, filename string) {
	graph.topology._Store(filename + "-graph")
	graph.nodes._Store(filename + "-nodes")
	graph.edges._Store(filename + "-edges")
	graph.geom._Store(filename + "-geom")
	graph.skip_topology._Store(filename + "-skip_topology")
	_StoreDefaultWeighting(&graph.weight, filename+"-fastest_weighting")
	_StoreNodeTileStore(&graph.node_tiles, filename+"-tiles")
	_StoreEdgeTypes(graph.edge_types, filename+"-tiles_types")
}

func StoreTiledGraph2(graph *TiledGraph2, filename string) {
	graph.topology._Store(filename + "-graph")
	graph.nodes._Store(filename + "-nodes")
	graph.edges._Store(filename + "-edges")
	graph.geom._Store(filename + "-geom")
	_StoreDefaultWeighting(&graph.weight, filename+"-fastest_weighting")
	_StoreNodeTileStore(&graph.node_tiles, filename+"-tiles")
	_StoreEdgeTypes(graph.edge_types, filename+"-tiles_types")
	_StoreTileRanges(graph.border_nodes, graph.interior_nodes, graph.border_range_map, filename+"-tileranges")
}

func StoreTiledGraph3(graph *TiledGraph3, filename string) {
	graph.topology._Store(filename + "-graph")
	graph.nodes._Store(filename + "-nodes")
	graph.edges._Store(filename + "-edges")
	graph.geom._Store(filename + "-geom")
	graph.skip_topology._Store(filename + "-skip_topology")
	_StoreDefaultWeighting(&graph.weight, filename+"-fastest_weighting")
	_StoreNodeTileStore(&graph.node_tiles, filename+"-tiles")
	_StoreEdgeTypes(graph.edge_types, filename+"-tiles_types")
	graph.border_topology._Store(filename + "-skip3_border_topology")
	graph.skip_topology._Store(filename + "-skip3_topology")
	_StoreShortcutStore(&graph.skip_edges, &graph.skip_weights, filename+"-skip3_shortcuts")
}

func StoreCHGraph(graph *CHGraph, filename string) {
	graph.topology._Store(filename + "-graph")
	graph.nodes._Store(filename + "-nodes")
	graph.edges._Store(filename + "-edges")
	graph.geom._Store(filename + "-geom")
	graph.ch_topology._Store(filename + "-ch_graph")
	_StoreDefaultWeighting(&graph.weight, filename+"-fastest_weighting")
	_StoreCHLevelStore(&graph.node_levels, filename+"-level")
	_StoreCHShortcutStore(&graph.shortcuts, &graph.sh_weight, filename+"-shortcut")
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

func _StoreEdgeTypes(edge_types Array[byte], filename string) {
	typesbuffer := bytes.Buffer{}

	for i := 0; i < edge_types.Length(); i++ {
		binary.Write(&typesbuffer, binary.LittleEndian, edge_types[i])
	}

	typesfile, _ := os.Create(filename)
	defer typesfile.Close()
	typesfile.Write(typesbuffer.Bytes())
}

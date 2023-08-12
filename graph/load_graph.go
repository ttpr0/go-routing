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
// graph io
//*******************************************

func LoadGraph(file string) IGraph {
	nodes := _LoadNodeStore(file + "-nodes")
	nodecount := nodes.NodeCount()
	edges := _LoadEdgeStore(file + "-edges")
	edgecount := edges.EdgeCount()
	topology := _LoadTopologyStore(file+"-graph", nodecount)
	geoms := _LoadGeometryStore(file+"-geom", nodecount, edgecount)
	weights := _LoadDefaultWeighting(file+"-fastest_weighting", edgecount)
	index := BuildNodeIndex(geoms.GetAllNodes())

	return &Graph{
		nodes:    *nodes,
		topology: *topology,
		edges:    *edges,
		geom:     *geoms,
		weight:   *weights,
		index:    index,
	}
}

func LoadCHGraph(file string) ICHGraph {
	nodes := _LoadNodeStore(file + "-nodes")
	nodecount := nodes.NodeCount()
	edges := _LoadEdgeStore(file + "-edges")
	edgecount := edges.EdgeCount()
	topology := _LoadTopologyStore(file+"-graph", nodecount)
	geoms := _LoadGeometryStore(file+"-geom", nodecount, edgecount)
	weights := _LoadDefaultWeighting(file+"-fastest_weighting", edgecount)
	ch_topology := _LoadTopologyStore(file+"-ch_graph", nodecount)
	levels := _LoadCHLevelStore(file+"-level", nodecount)
	shortcuts, sh_weights := _LoadCHShortcutStore(file + "-shortcut")
	chg := &CHGraph{
		nodes:       *nodes,
		node_levels: *levels,
		edges:       *edges,
		shortcuts:   *shortcuts,
		topology:    *topology,
		ch_topology: *ch_topology,
		geom:        *geoms,
		weight:      *weights,
		sh_weight:   *sh_weights,
	}
	SortNodesByLevel(chg)
	chg.index = BuildNodeIndex(chg.geom.GetAllNodes())
	return chg
}

func LoadTiledGraph(file string) ITiledGraph {
	nodes := _LoadNodeStore(file + "-nodes")
	nodecount := nodes.NodeCount()
	edges := _LoadEdgeStore(file + "-edges")
	edgecount := edges.EdgeCount()
	topology := _LoadTopologyStore(file+"-graph", nodecount)
	skip_topology := _LoadTopologyStore(file+"-skip_topology", nodecount)
	geoms := _LoadGeometryStore(file+"-geom", nodecount, edgecount)
	weights := _LoadDefaultWeighting(file+"-fastest_weighting", edgecount)
	edge_types := _LoadEdgeTypes(file+"-tiles_types", edgecount)
	node_tiles := _LoadNodeTileStore(file+"-tiles", nodecount)
	fmt.Println("start buidling index")
	index := BuildNodeIndex(geoms.GetAllNodes())
	fmt.Println("finished building index")

	return &TiledGraph{
		nodes:         *nodes,
		node_tiles:    *node_tiles,
		topology:      *topology,
		edges:         *edges,
		skip_topology: *skip_topology,
		edge_types:    edge_types,
		geom:          *geoms,
		weight:        *weights,
		index:         index,
	}
}

func LoadTiledGraph2(file string) ITiledGraph2 {
	nodes := _LoadNodeStore(file + "-nodes")
	nodecount := nodes.NodeCount()
	edges := _LoadEdgeStore(file + "-edges")
	edgecount := edges.EdgeCount()
	topology := _LoadTopologyStore(file+"-graph", nodecount)
	geoms := _LoadGeometryStore(file+"-geom", nodecount, edgecount)
	weights := _LoadDefaultWeighting(file+"-fastest_weighting", edgecount)
	edge_types := _LoadEdgeTypes(file+"-tiles_types", edgecount)
	node_tiles := _LoadNodeTileStore(file+"-tiles", nodecount)
	fmt.Println("start buidling index")
	index := BuildNodeIndex(geoms.GetAllNodes())
	fmt.Println("finished building index")
	border_nodes, interior_nodes, border_range_map := _LoadTileRanges(file + "-tileranges")

	return &TiledGraph2{
		nodes:            *nodes,
		node_tiles:       *node_tiles,
		topology:         *topology,
		edges:            *edges,
		edge_types:       edge_types,
		geom:             *geoms,
		weight:           *weights,
		index:            index,
		border_nodes:     border_nodes,
		interior_nodes:   interior_nodes,
		border_range_map: border_range_map,
	}
}

//*******************************************
// load graph information
//*******************************************

func _LoadTileRanges(file string) (Dict[int16, Array[int32]], Dict[int16, Array[int32]], Dict[int16, Dict[int32, Array[float32]]]) {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	tiledata, _ := os.ReadFile(file)
	tilereader := bytes.NewReader(tiledata)
	var tilecount int32
	binary.Read(tilereader, binary.LittleEndian, &tilecount)

	border_nodes := NewDict[int16, Array[int32]](100)
	interior_nodes := NewDict[int16, Array[int32]](100)
	border_range_map := NewDict[int16, Dict[int32, Array[float32]]](100)

	for i := 0; i < int(tilecount); i++ {
		var tile int16
		binary.Read(tilereader, binary.LittleEndian, &tile)
		var b_node_count int32
		binary.Read(tilereader, binary.LittleEndian, &b_node_count)
		b_nodes := NewArray[int32](int(b_node_count))
		for j := 0; j < int(b_node_count); j++ {
			var node int32
			binary.Read(tilereader, binary.LittleEndian, &node)
			b_nodes[j] = node
		}
		border_nodes[tile] = b_nodes
		var i_node_count int32
		binary.Read(tilereader, binary.LittleEndian, &i_node_count)
		i_nodes := NewArray[int32](int(i_node_count))
		for j := 0; j < int(i_node_count); j++ {
			var node int32
			binary.Read(tilereader, binary.LittleEndian, &node)
			i_nodes[j] = node
		}
		interior_nodes[tile] = i_nodes
		range_map := NewDict[int32, Array[float32]](int(b_node_count))
		for _, b_node := range b_nodes {
			ranges := NewArray[float32](int(i_node_count))
			for j, _ := range i_nodes {
				var dist float32
				binary.Read(tilereader, binary.LittleEndian, &dist)
				ranges[j] = dist
			}
			range_map[b_node] = ranges
		}
		border_range_map[tile] = range_map
	}

	return border_nodes, interior_nodes, border_range_map
}

func _LoadEdgeTypes(file string, edgecount int) Array[byte] {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	tiledata, _ := os.ReadFile(file)
	tilereader := bytes.NewReader(tiledata)
	edge_types := NewArray[byte](edgecount)
	for i := 0; i < edgecount; i++ {
		var t byte
		binary.Read(tilereader, binary.LittleEndian, &t)
		edge_types[i] = t
	}

	return edge_types
}

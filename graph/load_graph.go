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
	store := _LoadGraphStorage(file)
	nodecount := store.NodeCount()
	edgecount := store.EdgeCount()
	topology := _LoadTopologyStore(file+"-graph", nodecount)
	weights := _LoadDefaultWeighting(file+"-fastest_weighting", edgecount)
	index := _BuildKDTreeIndex(store)

	return &Graph{
		store:    store,
		topology: *topology,
		weight:   *weights,
		index:    index,
	}
}

func LoadCHGraph(file string) ICHGraph {
	store := _LoadGraphStorage(file)
	nodecount := store.NodeCount()
	edgecount := store.EdgeCount()
	topology := _LoadTopologyStore(file+"-graph", nodecount)
	weights := _LoadDefaultWeighting(file+"-fastest_weighting", edgecount)
	ch_topology := _LoadTopologyStore(file+"-ch_graph", nodecount)
	ch_store := _LoadCHStorage(file, nodecount)
	chg := &CHGraph{
		store:       store,
		topology:    *topology,
		ch_topology: *ch_topology,
		ch_store:    ch_store,
		weight:      *weights,
	}
	SortNodesByLevel(chg)
	chg.index = _BuildKDTreeIndex(chg.store)
	return chg
}

func LoadTiledGraph(file string) ITiledGraph {
	store := _LoadGraphStorage(file)
	nodecount := store.NodeCount()
	edgecount := store.EdgeCount()
	topology := _LoadTopologyStore(file+"-graph", nodecount)
	skip_topology := _LoadTypedTopology(file+"-skip_topology", nodecount)
	weights := _LoadDefaultWeighting(file+"-fastest_weighting", edgecount)
	skip_store := _LoadTiledStorage(file, nodecount, edgecount)
	fmt.Println("start buidling index")
	index := _BuildKDTreeIndex(store)
	fmt.Println("finished building index")

	return &TiledGraph{
		store:         store,
		topology:      *topology,
		skip_topology: *skip_topology,
		skip_store:    skip_store,
		weight:        *weights,
		index:         index,
	}
}

func LoadTiledGraph2(file string) ITiledGraph2 {
	tg := LoadTiledGraph(file).(*TiledGraph)
	border_nodes, interior_nodes, border_range_map := _LoadTileRanges(file + "-tileranges")

	return &TiledGraph2{
		TiledGraph:       *tg,
		border_nodes:     border_nodes,
		interior_nodes:   interior_nodes,
		border_range_map: border_range_map,
	}
}

func LoadTiledGraph3(file string) *TiledGraph3 {
	tg := LoadTiledGraph(file).(*TiledGraph)
	tile_ranges, index_edges := _LoadTileRanges2(file + "-tileranges")

	return &TiledGraph3{
		TiledGraph:  *tg,
		tile_ranges: tile_ranges,
		index_edges: index_edges,
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

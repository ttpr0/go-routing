package graph

import (
	"fmt"
)

//*******************************************
// graph io
//*******************************************

func LoadGraph(file string) *Graph {
	store := _LoadGraphStorage(file)
	nodecount := store.NodeCount()
	edgecount := store.EdgeCount()
	topology := _LoadAdjacency(file+"-graph", false, nodecount)
	weights := _LoadDefaultWeighting(file+"-fastest_weighting", edgecount)
	index := _BuildKDTreeIndex(store)

	return &Graph{
		store:    store,
		topology: *topology,
		weight:   *weights,
		index:    index,
	}
}

func LoadGraph2(file string) *Graph {
	store := _LoadGraphStorageMin(file)
	nodecount := store.NodeCount()
	edgecount := store.EdgeCount()
	topology := _LoadAdjacency(file+"-graph", false, nodecount)
	weights := _LoadDefaultWeighting(file+"-fastest_weighting", edgecount)
	index := _BuildKDTreeIndex(store)

	return &Graph{
		store:    store,
		topology: *topology,
		weight:   *weights,
		index:    index,
	}
}

func LoadCHGraph(file string) *CHGraph {
	store := _LoadGraphStorage(file)
	nodecount := store.NodeCount()
	edgecount := store.EdgeCount()
	topology := _LoadAdjacency(file+"-graph", false, nodecount)
	weights := _LoadDefaultWeighting(file+"-fastest_weighting", edgecount)
	ch_topology := _LoadAdjacency(file+"-ch_graph", false, nodecount)
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

func LoadCHGraph2(file string) *CHGraph {
	store := _LoadGraphStorage(file)
	nodecount := store.NodeCount()
	edgecount := store.EdgeCount()
	topology := _LoadAdjacency(file+"-graph", false, nodecount)
	weights := _LoadDefaultWeighting(file+"-fastest_weighting", edgecount)
	ch_topology := _LoadAdjacency(file+"-ch_graph", false, nodecount)
	ch_store := _LoadCHStorage(file, nodecount)
	chg := &CHGraph{
		store:       store,
		topology:    *topology,
		ch_topology: *ch_topology,
		ch_store:    ch_store,
		weight:      *weights,
	}
	chg.index = _BuildKDTreeIndex(chg.store)
	return chg
}

func LoadTiledGraph(file string) *TiledGraph {
	store := _LoadGraphStorage(file)
	nodecount := store.NodeCount()
	edgecount := store.EdgeCount()
	topology := _LoadAdjacency(file+"-graph", false, nodecount)
	skip_topology := _LoadAdjacency(file+"-skip_topology", true, nodecount)
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

func LoadTiledGraph3(file string) *TiledGraph3 {
	tg := LoadTiledGraph(file)
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

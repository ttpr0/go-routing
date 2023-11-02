package graph

// import (
// 	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
// )

//*******************************************
// graph io
//*******************************************

// func LoadGraph(file string) *Graph {
// 	store := _LoadGraphStorage(file)
// 	nodecount := store.NodeCount()
// 	edgecount := store.EdgeCount()
// 	topology := _LoadAdjacency(file+"-graph", false, nodecount)
// 	weights := _LoadDefaultWeighting(file+"-fastest_weighting", edgecount)
// 	index := _BuildKDTreeIndex(store)

// 	return &Graph{
// 		base: GraphBase{
// 			store:    store,
// 			topology: *topology,
// 			index:    index,
// 		},
// 		weight: weights,
// 	}
// }

// func LoadGraph2(file string) *Graph {
// 	store := _LoadGraphStorageMin(file)
// 	nodecount := store.NodeCount()
// 	edgecount := store.EdgeCount()
// 	topology := _LoadAdjacency(file+"-graph", false, nodecount)
// 	weights := _LoadDefaultWeighting(file+"-fastest_weighting", edgecount)
// 	index := _BuildKDTreeIndex(store)

// 	return &Graph{
// 		base: GraphBase{
// 			store:    store,
// 			topology: *topology,
// 			index:    index,
// 		},
// 		weight: weights,
// 	}
// }

// func LoadCHGraph(file string) *CHGraph {
// 	store := _LoadGraphStorage(file)
// 	nodecount := store.NodeCount()
// 	edgecount := store.EdgeCount()
// 	topology := _LoadAdjacency(file+"-graph", false, nodecount)
// 	weights := _LoadDefaultWeighting(file+"-fastest_weighting", edgecount)
// 	ch_topology := _LoadAdjacency(file+"-ch_graph", false, nodecount)
// 	ch_shortcuts := _LoadShortcuts(file + "-shortcuts")
// 	node_levels := ReadArrayFromFile[int16](file + "-level")
// 	chg := &CHGraph{
// 		base: GraphBase{
// 			store:    store,
// 			topology: *topology,
// 		},
// 		weight: weights,

// 		ch_shortcuts: ch_shortcuts,
// 		ch_topology:  *ch_topology,
// 		node_levels:  node_levels,
// 	}
// 	SortNodesByLevel(chg)
// 	chg.base.index = _BuildKDTreeIndex(chg.base.store)
// 	return chg
// }

// func LoadCHGraph2(file string) *CHGraph {
// 	store := _LoadGraphStorage(file)
// 	nodecount := store.NodeCount()
// 	edgecount := store.EdgeCount()
// 	topology := _LoadAdjacency(file+"-graph", false, nodecount)
// 	weights := _LoadDefaultWeighting(file+"-fastest_weighting", edgecount)
// 	ch_topology := _LoadAdjacency(file+"-ch_graph", false, nodecount)
// 	ch_shortcuts := _LoadShortcuts(file + "-shortcuts")
// 	node_levels := ReadArrayFromFile[int16](file + "-level")
// 	chg := &CHGraph{
// 		base: GraphBase{
// 			store:    store,
// 			topology: *topology,
// 		},
// 		weight: weights,

// 		ch_shortcuts: ch_shortcuts,
// 		ch_topology:  *ch_topology,
// 		node_levels:  node_levels,
// 	}
// 	chg.base.index = _BuildKDTreeIndex(chg.base.store)
// 	return chg
// }

// func LoadTiledGraph(file string) *TiledGraph {
// 	store := _LoadGraphStorage(file)
// 	nodecount := store.NodeCount()
// 	edgecount := store.EdgeCount()
// 	topology := _LoadAdjacency(file+"-graph", false, nodecount)
// 	weights := _LoadDefaultWeighting(file+"-fastest_weighting", edgecount)
// 	skip_shortcuts := _LoadShortcuts(file + "skip_shortcuts")
// 	skip_topology := _LoadAdjacency(file+"-skip_topology", true, nodecount)
// 	node_tiles := ReadArrayFromFile[int16](file + "-tiles")
// 	edge_types := ReadArrayFromFile[byte](file + "-tiles_types")

// 	var cell_index Optional[_CellIndex]
// 	_, err := os.Stat(file + "tileranges")
// 	if errors.Is(err, os.ErrNotExist) {
// 		cell_index = None[_CellIndex]()
// 	} else {
// 		cell_index = Some(_LoadCellIndex(file + "tileranges"))
// 	}

// 	fmt.Println("start buidling index")
// 	index := _BuildKDTreeIndex(store)
// 	fmt.Println("finished building index")

// 	return &TiledGraph{
// 		base: GraphBase{
// 			store:    store,
// 			topology: *topology,
// 			index:    index,
// 		},
// 		weight: weights,

// 		skip_shortcuts: skip_shortcuts,
// 		skip_topology:  *skip_topology,
// 		node_tiles:     node_tiles,
// 		edge_types:     edge_types,
// 		cell_index:     cell_index,
// 	}
// }

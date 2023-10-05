package graph

//*******************************************
// graph io
//*******************************************

func StoreGraph(graph *Graph, filename string) {
	_StoreAdjacency(&graph.topology, false, filename+"-graph")
	_StoreGraphStorage(graph.store, filename)
	_StoreDefaultWeighting(&graph.weight, filename+"-fastest_weighting")

	// tc_weight := _CreateTCWeighting(graph)
	// _StoreTCWeighting(tc_weight, filename+"-tc_weighting")
}

func StoreTiledGraph(graph *TiledGraph, filename string) {
	_StoreAdjacency(&graph.topology, false, filename+"-graph")
	_StoreGraphStorage(graph.store, filename)
	_StoreAdjacency(&graph.skip_topology, true, filename+"-skip_topology")
	_StoreDefaultWeighting(&graph.weight, filename+"-fastest_weighting")
	_StoreTiledStorage(graph.skip_store, filename)
}

func StoreTiledGraph3(graph *TiledGraph3, filename string) {
	StoreTiledGraph(&graph.TiledGraph, filename)
	_StoreTileRanges2(graph.tile_ranges, graph.index_edges, filename+"-tileranges")
}

func StoreCHGraph(graph *CHGraph, filename string) {
	_StoreAdjacency(&graph.topology, false, filename+"-graph")
	_StoreGraphStorage(graph.store, filename)
	_StoreAdjacency(&graph.ch_topology, false, filename+"-ch_graph")
	_StoreDefaultWeighting(&graph.weight, filename+"-fastest_weighting")
	_StoreCHStorage(graph.ch_store, filename)
}

func StoreCHGraph4(graph *CHGraph4, filename string) {
	StoreCHGraph(&graph.CHGraph, filename)
	_StoreTiledNodeTiles(graph.node_tiles, filename+"-tiles")
}

//*******************************************
// store graph information
//*******************************************

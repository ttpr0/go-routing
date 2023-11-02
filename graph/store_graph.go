package graph

//*******************************************
// graph io
//*******************************************

func StoreGraph(graph *Graph, filename string) {
	_StoreAdjacency(&graph.base.topology, false, filename+"-graph")
	_StoreGraphStorage(graph.base.store, filename)
	panic("TODO")
	// _StoreDefaultWeighting(&graph.weight, filename+"-fastest_weighting")

	// tc_weight := _CreateTCWeighting(graph)
	// _StoreTCWeighting(tc_weight, filename+"-tc_weighting")
}

func StoreTiledGraph(graph *TiledGraph, filename string) {
	_StoreAdjacency(&graph.base.topology, false, filename+"-graph")
	_StoreGraphStorage(graph.base.store, filename)
	panic("TODO")
	// _StoreDefaultWeighting(&graph.weight, filename+"-fastest_weighting")
	_StoreShortcuts(graph.skip_shortcuts, filename+"-skip_shortcuts")
	_StoreAdjacency(&graph.skip_topology, true, filename+"-skip_topology")
	WriteArrayToFile[int16](graph.node_tiles, filename+"-tiles")
	WriteArrayToFile[byte](graph.edge_types, filename+"-tiles_types")
	if graph.HasCellIndex() {
		_StoreCellIndex(graph.cell_index.Value, filename+"-tileranges")
	}
}

func StoreCHGraph(graph *CHGraph, filename string) {
	_StoreAdjacency(&graph.base.topology, false, filename+"-graph")
	_StoreGraphStorage(graph.base.store, filename)
	panic("TODO")
	// _StoreDefaultWeighting(&graph.weight, filename+"-fastest_weighting")
	_StoreShortcuts(graph.ch_shortcuts, filename+"-shortcut")
	_StoreAdjacency(&graph.ch_topology, false, filename+"-ch_graph")
	WriteArrayToFile[int16](graph.node_levels, filename+"-level")
	if graph._build_with_tiles {
		WriteArrayToFile[int16](graph.node_tiles.Value, filename+"-ch_tiles")
	}
}

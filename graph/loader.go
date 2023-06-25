package graph

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// graph io
//*******************************************

func StoreGraph(graph *Graph, filename string) {
	_StoreNodes(graph.nodes, graph.node_refs, graph.fwd_edge_refs, graph.bwd_edge_refs, filename+"-nodes")
	_StoreEdges(graph.edges, graph.weight, filename+"-edges")
	_StoreGeom(graph.geom, filename+"-geom")
}

func StoreTiledGraph(graph *TiledGraph, filename string) {
	_StoreNodes(graph.nodes, graph.node_refs, graph.fwd_edge_refs, graph.bwd_edge_refs, filename+"-nodes")
	_StoreEdges(graph.edges, graph.weight, filename+"-edges")
	_StoreGeom(graph.geom, filename+"-geom")
	_StoreNodeTiles(graph.node_tiles, filename+"-tiles")
}

func StoreTiledGraph2(graph *TiledGraph2, filename string) {
	_StoreNodes(graph.nodes, graph.node_refs, graph.fwd_edge_refs, graph.bwd_edge_refs, filename+"-nodes")
	_StoreEdges(graph.edges, graph.weight, filename+"-edges")
	_StoreGeom(graph.geom, filename+"-geom")
	_StoreNodeTiles(graph.node_tiles, filename+"-tiles")
	_StoreTileRanges(graph.border_nodes, graph.interior_nodes, graph.border_range_map, filename+"-tileranges")
}

func StoreCHGraph(graph *CHGraph, filename string) {
	_StoreNodes(graph.nodes, graph.node_refs, graph.fwd_edge_refs, graph.bwd_edge_refs, filename+"-nodes")
	_StoreEdges(graph.edges, graph.weight, filename+"-edges")
	_StoreGeom(graph.geom, filename+"-geom")

	shcbuffer := bytes.Buffer{}
	lvlbuffer := bytes.Buffer{}
	nodecount := graph.nodes.Length()
	shortcutcount := graph.shortcuts.Length()
	binary.Write(&shcbuffer, binary.LittleEndian, int32(shortcutcount))

	for i := 0; i < shortcutcount; i++ {
		shortcut := graph.shortcuts.Get(i)
		weight := graph.sh_weight.GetEdgeWeight(int32(i))
		binary.Write(&shcbuffer, binary.LittleEndian, int32(shortcut.NodeA))
		binary.Write(&shcbuffer, binary.LittleEndian, int32(shortcut.NodeB))
		binary.Write(&shcbuffer, binary.LittleEndian, uint32(weight))
		for _, edge := range shortcut.Edges {
			binary.Write(&shcbuffer, binary.LittleEndian, edge.A)
			binary.Write(&shcbuffer, binary.LittleEndian, edge.B == 2 || edge.B == 3)
		}
	}
	for i := 0; i < nodecount; i++ {
		binary.Write(&lvlbuffer, binary.LittleEndian, graph.GetNodeLevel(int32(i)))
	}

	shcfile, _ := os.Create(filename + "-shortcut")
	defer shcfile.Close()
	shcfile.Write(shcbuffer.Bytes())
	lvlfile, _ := os.Create(filename + "-level")
	defer lvlfile.Close()
	lvlfile.Write(lvlbuffer.Bytes())
}

func LoadGraph(file string) IGraph {
	nodes, node_refs, fwd_edge_refs, bwd_edge_refs := _LoadNodes(file + "-nodes")
	nodecount := nodes.Length()
	edges, edge_weights := _LoadEdges(file + "-edges")
	edgecount := edges.Length()
	node_geoms, edge_geoms := _LoadGeom(file+"-geom", nodecount, edgecount)
	index := _BuildNodeIndex(node_geoms)

	return &Graph{
		node_refs:     node_refs,
		nodes:         nodes,
		fwd_edge_refs: fwd_edge_refs,
		bwd_edge_refs: bwd_edge_refs,
		edges:         edges,
		geom:          &Geometry{node_geoms, edge_geoms},
		weight:        &Weighting{edge_weights},
		index:         index,
	}
}

func LoadCHGraph(file string) ICHGraph {
	nodes, node_refs, fwd_edge_refs, bwd_edge_refs := _LoadNodes(file + "-nodes")
	nodecount := nodes.Length()
	edges, edge_weights := _LoadEdges(file + "-edges")
	edgecount := edges.Length()
	node_geoms, edge_geoms := _LoadGeom(file+"-geom", nodecount, edgecount)
	levels := _LoadCHLevels(file+"-level", nodecount)
	shortcuts, shortcut_weights := _LoadCHShortcuts(file + "-shortcut")
	chg := &CHGraph{
		node_refs:     node_refs,
		nodes:         nodes,
		node_levels:   levels,
		fwd_edge_refs: fwd_edge_refs,
		bwd_edge_refs: bwd_edge_refs,
		edges:         edges,
		shortcuts:     shortcuts,
		geom:          &Geometry{node_geoms, edge_geoms},
		weight:        &Weighting{edge_weights},
		sh_weight:     &Weighting{shortcut_weights},
	}
	SortNodesByLevel(chg)
	chg.index = _BuildNodeIndex(chg.geom.GetAllNodes())
	return chg
}

func LoadTiledGraph(file string) ITiledGraph {
	nodes, node_refs, fwd_edge_refs, bwd_edge_refs := _LoadNodes(file + "-nodes")
	nodecount := nodes.Length()
	edges, edge_weights := _LoadEdges(file + "-edges")
	edgecount := edges.Length()
	node_geoms, edge_geoms := _LoadGeom(file+"-geom", nodecount, edgecount)
	node_tiles := _LoadNodeTiles(file+"-tiles", nodecount)
	fmt.Println("start buidling index")
	index := _BuildNodeIndex(node_geoms)
	fmt.Println("finished building index")

	return &TiledGraph{
		node_refs:     node_refs,
		nodes:         nodes,
		node_tiles:    node_tiles,
		fwd_edge_refs: fwd_edge_refs,
		bwd_edge_refs: bwd_edge_refs,
		edges:         edges,
		geom:          &Geometry{node_geoms, edge_geoms},
		weight:        &Weighting{edge_weights},
		index:         index,
	}
}

func LoadTiledGraph2(file string) ITiledGraph2 {
	nodes, node_refs, fwd_edge_refs, bwd_edge_refs := _LoadNodes(file + "-nodes")
	nodecount := nodes.Length()
	edges, edge_weights := _LoadEdges(file + "-edges")
	edgecount := edges.Length()
	node_geoms, edge_geoms := _LoadGeom(file+"-geom", nodecount, edgecount)
	node_tiles := _LoadNodeTiles(file+"-tiles", nodecount)
	fmt.Println("start buidling index")
	index := _BuildNodeIndex(node_geoms)
	fmt.Println("finished building index")
	border_nodes, interior_nodes, border_range_map := _LoadTileRanges(file + "tileranges")

	return &TiledGraph2{
		node_refs:        node_refs,
		nodes:            nodes,
		node_tiles:       node_tiles,
		fwd_edge_refs:    fwd_edge_refs,
		bwd_edge_refs:    bwd_edge_refs,
		edges:            edges,
		geom:             &Geometry{node_geoms, edge_geoms},
		weight:           &Weighting{edge_weights},
		index:            index,
		border_nodes:     border_nodes,
		interior_nodes:   interior_nodes,
		border_range_map: border_range_map,
	}
}

//*******************************************
// store graph information
//*******************************************

func _StoreNodes(nodes List[Node], node_refs List[NodeRef], fwd_edge_refs List[EdgeRef], bwd_edge_refs List[EdgeRef], filename string) {
	nodesbuffer := bytes.Buffer{}

	nodecount := nodes.Length()
	fwd_edgerefcount := fwd_edge_refs.Length()
	bwd_edgerefcount := bwd_edge_refs.Length()
	binary.Write(&nodesbuffer, binary.LittleEndian, int32(nodecount))
	binary.Write(&nodesbuffer, binary.LittleEndian, int32(fwd_edgerefcount))
	binary.Write(&nodesbuffer, binary.LittleEndian, int32(bwd_edgerefcount))

	for i := 0; i < nodecount; i++ {
		node := nodes.Get(i)
		node_ref := node_refs.Get(i)
		binary.Write(&nodesbuffer, binary.LittleEndian, node.Type)
		binary.Write(&nodesbuffer, binary.LittleEndian, node_ref.EdgeRefFWDStart)
		binary.Write(&nodesbuffer, binary.LittleEndian, node_ref.EdgeRefFWDCount)
		binary.Write(&nodesbuffer, binary.LittleEndian, node_ref.EdgeRefBWDStart)
		binary.Write(&nodesbuffer, binary.LittleEndian, node_ref.EdgeRefBWDCount)
	}
	for i := 0; i < fwd_edgerefcount; i++ {
		edgeref := fwd_edge_refs.Get(i)
		binary.Write(&nodesbuffer, binary.LittleEndian, edgeref.EdgeID)
		binary.Write(&nodesbuffer, binary.LittleEndian, edgeref._Type)
		binary.Write(&nodesbuffer, binary.LittleEndian, edgeref.OtherID)
		binary.Write(&nodesbuffer, binary.LittleEndian, edgeref.Weight)
	}
	for i := 0; i < bwd_edgerefcount; i++ {
		edgeref := bwd_edge_refs.Get(i)
		binary.Write(&nodesbuffer, binary.LittleEndian, edgeref.EdgeID)
		binary.Write(&nodesbuffer, binary.LittleEndian, edgeref._Type)
		binary.Write(&nodesbuffer, binary.LittleEndian, edgeref.OtherID)
		binary.Write(&nodesbuffer, binary.LittleEndian, edgeref.Weight)
	}

	nodesfile, _ := os.Create(filename)
	defer nodesfile.Close()
	nodesfile.Write(nodesbuffer.Bytes())
}

func _StoreEdges(edges List[Edge], weight IWeighting, filename string) {
	edgesbuffer := bytes.Buffer{}

	edgecount := edges.Length()
	binary.Write(&edgesbuffer, binary.LittleEndian, int32(edgecount))

	for i := 0; i < edgecount; i++ {
		edge := edges.Get(i)
		edge_weight := weight.GetEdgeWeight(int32(i))
		binary.Write(&edgesbuffer, binary.LittleEndian, int32(edge.NodeA))
		binary.Write(&edgesbuffer, binary.LittleEndian, int32(edge.NodeB))
		binary.Write(&edgesbuffer, binary.LittleEndian, uint8(edge_weight))
		binary.Write(&edgesbuffer, binary.LittleEndian, byte(edge.Type))
		binary.Write(&edgesbuffer, binary.LittleEndian, edge.Length)
		binary.Write(&edgesbuffer, binary.LittleEndian, uint8(edge.Maxspeed))
	}

	edgesfile, _ := os.Create(filename)
	defer edgesfile.Close()
	edgesfile.Write(edgesbuffer.Bytes())
}

func _StoreGeom(geom IGeometry, filename string) {
	geombuffer := bytes.Buffer{}

	nodecount := len(geom.GetAllNodes())
	edgecount := len(geom.GetAllEdges())

	for i := 0; i < nodecount; i++ {
		point := geom.GetNode(int32(i))
		binary.Write(&geombuffer, binary.LittleEndian, point[0])
		binary.Write(&geombuffer, binary.LittleEndian, point[1])
	}
	c := 0
	for i := 0; i < edgecount; i++ {
		nc := len(geom.GetEdge(int32(i)))
		binary.Write(&geombuffer, binary.LittleEndian, int32(c))
		binary.Write(&geombuffer, binary.LittleEndian, uint8(nc))
		c += nc * 8
	}
	for i := 0; i < edgecount; i++ {
		coords := geom.GetEdge(int32(i))
		for _, coord := range coords {
			binary.Write(&geombuffer, binary.LittleEndian, coord[0])
			binary.Write(&geombuffer, binary.LittleEndian, coord[1])
		}
	}

	geomfile, _ := os.Create(filename)
	defer geomfile.Close()
	geomfile.Write(geombuffer.Bytes())
}

func _StoreNodeTiles(node_tiles List[int16], filename string) {
	tilesbuffer := bytes.Buffer{}

	for i := 0; i < node_tiles.Length(); i++ {
		binary.Write(&tilesbuffer, binary.LittleEndian, node_tiles[i])
	}

	tilesfile, _ := os.Create(filename)
	defer tilesfile.Close()
	tilesfile.Write(tilesbuffer.Bytes())
}

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

//*******************************************
// load graph information
//*******************************************

func _LoadNodes(file string) (List[Node], List[NodeRef], List[EdgeRef], List[EdgeRef]) {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	nodedata, _ := os.ReadFile(file)
	nodereader := bytes.NewReader(nodedata)
	var nodecount int32
	binary.Read(nodereader, binary.LittleEndian, &nodecount)
	var fwd_edgerefcount int32
	binary.Read(nodereader, binary.LittleEndian, &fwd_edgerefcount)
	var bwd_edgerefcount int32
	binary.Read(nodereader, binary.LittleEndian, &bwd_edgerefcount)
	nodes := NewList[Node](int(nodecount))
	node_refs := NewList[NodeRef](int(nodecount))
	fwd_edge_refs := NewList[EdgeRef](int(fwd_edgerefcount))
	bwd_edge_refs := NewList[EdgeRef](int(bwd_edgerefcount))
	for i := 0; i < int(nodecount); i++ {
		var t int8
		binary.Read(nodereader, binary.LittleEndian, &t)
		var s1 int32
		binary.Read(nodereader, binary.LittleEndian, &s1)
		var c1 int16
		binary.Read(nodereader, binary.LittleEndian, &c1)
		var s2 int32
		binary.Read(nodereader, binary.LittleEndian, &s2)
		var c2 int16
		binary.Read(nodereader, binary.LittleEndian, &c2)
		nodes.Add(Node{
			Type: t,
		})
		node_refs.Add(NodeRef{
			EdgeRefFWDStart: s1,
			EdgeRefFWDCount: c1,
			EdgeRefBWDStart: s2,
			EdgeRefBWDCount: c2,
		})
	}
	for i := 0; i < int(fwd_edgerefcount); i++ {
		var id int32
		binary.Read(nodereader, binary.LittleEndian, &id)
		var t byte
		binary.Read(nodereader, binary.LittleEndian, &t)
		var nid int32
		binary.Read(nodereader, binary.LittleEndian, &nid)
		var w int32
		binary.Read(nodereader, binary.LittleEndian, &w)
		fwd_edge_refs.Add(EdgeRef{
			EdgeID:  id,
			_Type:   t,
			OtherID: nid,
			Weight:  w,
		})
	}
	for i := 0; i < int(bwd_edgerefcount); i++ {
		var id int32
		binary.Read(nodereader, binary.LittleEndian, &id)
		var t byte
		binary.Read(nodereader, binary.LittleEndian, &t)
		var nid int32
		binary.Read(nodereader, binary.LittleEndian, &nid)
		var w int32
		binary.Read(nodereader, binary.LittleEndian, &w)
		bwd_edge_refs.Add(EdgeRef{
			EdgeID:  id,
			_Type:   t,
			OtherID: nid,
			Weight:  w,
		})
	}

	return nodes, node_refs, fwd_edge_refs, bwd_edge_refs
}

func _LoadEdges(file string) (List[Edge], List[int32]) {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	edgedata, _ := os.ReadFile(file)
	edgereader := bytes.NewReader(edgedata)
	var edgecount int32
	binary.Read(edgereader, binary.LittleEndian, &edgecount)
	edges := NewList[Edge](int(edgecount))
	edge_weights := NewList[int32](int(edgecount))
	for i := 0; i < int(edgecount); i++ {
		var a int32
		binary.Read(edgereader, binary.LittleEndian, &a)
		var b int32
		binary.Read(edgereader, binary.LittleEndian, &b)
		var w uint8
		binary.Read(edgereader, binary.LittleEndian, &w)
		var t byte
		binary.Read(edgereader, binary.LittleEndian, &t)
		var l float32
		binary.Read(edgereader, binary.LittleEndian, &l)
		var m uint8
		binary.Read(edgereader, binary.LittleEndian, &m)
		edges.Add(Edge{
			NodeA:    a,
			NodeB:    b,
			Type:     RoadType(t),
			Length:   l,
			Maxspeed: m,
		})
		edge_weights.Add(int32(w))
	}

	return edges, edge_weights
}

func _LoadGeom(file string, nodecount, edgecount int) ([]geo.Coord, []geo.CoordArray) {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	geomdata, _ := os.ReadFile(file)
	startindex := nodecount*8 + edgecount*5
	geomreader := bytes.NewReader(geomdata)
	linereader := bytes.NewReader(geomdata[startindex:])
	node_geoms := make([]geo.Coord, nodecount)
	for i := 0; i < int(nodecount); i++ {
		var lon float32
		binary.Read(geomreader, binary.LittleEndian, &lon)
		var lat float32
		binary.Read(geomreader, binary.LittleEndian, &lat)
		node_geoms[i] = geo.Coord{lon, lat}
	}
	edge_geoms := make([]geo.CoordArray, edgecount)
	for i := 0; i < int(edgecount); i++ {
		var s int32
		binary.Read(geomreader, binary.LittleEndian, &s)
		var c byte
		binary.Read(geomreader, binary.LittleEndian, &c)
		points := make([]geo.Coord, c)
		for j := 0; j < int(c); j++ {
			var lon float32
			binary.Read(linereader, binary.LittleEndian, &lon)
			var lat float32
			binary.Read(linereader, binary.LittleEndian, &lat)
			points[j][0] = lon
			points[j][1] = lat
		}
		edge_geoms[i] = points
	}

	return node_geoms, edge_geoms
}

func _LoadNodeTiles(file string, nodecount int) List[int16] {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	tiledata, _ := os.ReadFile(file)
	tilereader := bytes.NewReader(tiledata)
	node_tiles := NewList[int16](int(nodecount))
	for i := 0; i < int(nodecount); i++ {
		var t int16
		binary.Read(tilereader, binary.LittleEndian, &t)
		node_tiles.Add(t)
	}

	return node_tiles
}

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

func _LoadCHLevels(file string, nodecount int) List[int16] {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	leveldata, _ := os.ReadFile(file)
	levelreader := bytes.NewReader(leveldata)
	levels := NewList[int16](int(nodecount))
	for i := 0; i < int(nodecount); i++ {
		var l int16
		binary.Read(levelreader, binary.LittleEndian, &l)
		levels.Add(l)
	}

	return levels
}

func _LoadCHShortcuts(file string) (List[CHShortcut], List[int32]) {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	shortcutdata, _ := os.ReadFile(file)
	shortcutreader := bytes.NewReader(shortcutdata)
	var shortcutcount int32
	binary.Read(shortcutreader, binary.LittleEndian, &shortcutcount)
	shortcuts := NewList[CHShortcut](int(shortcutcount))
	shortcut_weights := NewList[int32](int(shortcutcount))
	for i := 0; i < int(shortcutcount); i++ {
		var node_a int32
		binary.Read(shortcutreader, binary.LittleEndian, &node_a)
		var node_b int32
		binary.Read(shortcutreader, binary.LittleEndian, &node_b)
		var weight uint32
		binary.Read(shortcutreader, binary.LittleEndian, &weight)
		shortcut := CHShortcut{
			NodeA: node_a,
			NodeB: node_b,
			Edges: [2]Tuple[int32, byte]{},
		}
		for j := 0; j < 2; j++ {
			var id int32
			binary.Read(shortcutreader, binary.LittleEndian, &id)
			var is bool
			binary.Read(shortcutreader, binary.LittleEndian, &is)
			if is {
				shortcut.Edges[j] = MakeTuple(id, byte(2))
			} else {
				shortcut.Edges[j] = MakeTuple(id, byte(0))
			}
		}
		shortcuts.Add(shortcut)
		shortcut_weights.Add(int32(weight))
	}

	return shortcuts, shortcut_weights
}

//*******************************************
// build graph information
//*******************************************

func _BuildNodeIndex(node_geoms List[geo.Coord]) KDTree[int32] {
	tree := NewKDTree[int32](2)
	for i := 0; i < node_geoms.Length(); i++ {
		geom := node_geoms[i]
		tree.Insert(geom[:], int32(i))
	}
	return tree
}

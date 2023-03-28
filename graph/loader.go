package graph

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// graph io
//*******************************************

func StoreGraph(graph *Graph, filename string) {
	_StoreNodes(graph.nodes, graph.node_attributes, graph.edge_refs, filename+"-nodes")
	_StoreEdges(graph.edges, graph.edge_attributes, graph.weight, filename+"-edges")
	_StoreGeom(graph.geom, filename+"-geom")
}

func StoreTiledGraph(graph *TiledGraph, filename string) {
	_StoreNodes(graph.nodes, graph.node_attributes, graph.edge_refs, filename+"-nodes")
	_StoreEdges(graph.edges, graph.edge_attributes, graph.weight, filename+"-edges")
	_StoreGeom(graph.geom, filename+"-geom")
	_StoreNodeTiles(graph.node_tiles, filename+"-tiles")
}

func LoadGraph(file string) IGraph {
	nodes, node_attribs, edge_refs := _LoadNodes(file + "-nodes")
	nodecount := nodes.Length()
	edges, edge_attribs, edge_weights := _LoadEdges(file + "-edges")
	edgecount := edges.Length()
	node_geoms, edge_geoms := _LoadGeom(file+"-geom", nodecount, edgecount)

	return &Graph{
		nodes:           nodes,
		node_attributes: node_attribs,
		edge_refs:       edge_refs,
		edges:           edges,
		edge_attributes: edge_attribs,
		geom:            &Geometry{node_geoms, edge_geoms},
		weight:          &Weighting{edge_weights},
	}
}

func LoadCHGraph(file string) ICHGraph {
	nodes, node_attribs, edge_refs := _LoadNodes(file + "-nodes")
	nodecount := nodes.Length()
	edges, edge_attribs, edge_weights := _LoadEdges(file + "-edges")
	edgecount := edges.Length()
	node_geoms, edge_geoms := _LoadGeom(file+"-geom", nodecount, edgecount)
	levels := _LoadCHLevels(file+"-level", nodecount)
	shortcuts, shortcut_weights := _LoadCHShortcuts(file + "-shortcut")

	return &CHGraph{
		nodes:           nodes,
		node_attributes: node_attribs,
		node_levels:     levels,
		edge_refs:       edge_refs,
		edges:           edges,
		edge_attributes: edge_attribs,
		shortcuts:       shortcuts,
		geom:            &Geometry{node_geoms, edge_geoms},
		weight:          &Weighting{edge_weights},
		sh_weight:       &Weighting{shortcut_weights},
	}
}

func LoadTiledGraph(file string) ITiledGraph {
	nodes, node_attribs, edge_refs := _LoadNodes(file + "-nodes")
	nodecount := nodes.Length()
	edges, edge_attribs, edge_weights := _LoadEdges(file + "-edges")
	edgecount := edges.Length()
	node_geoms, edge_geoms := _LoadGeom(file+"-geom", nodecount, edgecount)
	node_tiles := _LoadNodeTiles(file+"-tiles", nodecount)

	return &TiledGraph{
		nodes:           nodes,
		node_attributes: node_attribs,
		node_tiles:      node_tiles,
		edge_refs:       edge_refs,
		edges:           edges,
		edge_attributes: edge_attribs,
		geom:            &Geometry{node_geoms, edge_geoms},
		weight:          &Weighting{edge_weights},
	}
}

//*******************************************
// store graph information
//*******************************************

func _StoreNodes(nodes List[Node], node_attributes List[NodeAttributes], edge_refs List[EdgeRef], filename string) {
	nodesbuffer := bytes.Buffer{}

	nodecount := nodes.Length()
	edgerefcount := edge_refs.Length()
	binary.Write(&nodesbuffer, binary.LittleEndian, int32(nodecount))
	binary.Write(&nodesbuffer, binary.LittleEndian, int32(edgerefcount))

	for i := 0; i < nodecount; i++ {
		node := nodes.Get(i)
		node_attrib := node_attributes.Get(i)
		binary.Write(&nodesbuffer, binary.LittleEndian, node_attrib.Type)
		binary.Write(&nodesbuffer, binary.LittleEndian, node.EdgeRefStart)
		binary.Write(&nodesbuffer, binary.LittleEndian, node.EdgeRefCount)
	}
	for i := 0; i < edgerefcount; i++ {
		edgeref := edge_refs.Get(i)
		binary.Write(&nodesbuffer, binary.LittleEndian, int32(edgeref.EdgeID))
		binary.Write(&nodesbuffer, binary.LittleEndian, edgeref.Type)
	}

	nodesfile, _ := os.Create(filename)
	defer nodesfile.Close()
	nodesfile.Write(nodesbuffer.Bytes())
}

func _StoreEdges(edges List[Edge], edge_attributes List[EdgeAttributes], weight IWeighting, filename string) {
	edgesbuffer := bytes.Buffer{}

	edgecount := edges.Length()
	binary.Write(&edgesbuffer, binary.LittleEndian, int32(edgecount))

	for i := 0; i < edgecount; i++ {
		edge := edges.Get(i)
		edge_attrib := edge_attributes.Get(i)
		edge_weight := weight.GetEdgeWeight(int32(i))
		binary.Write(&edgesbuffer, binary.LittleEndian, int32(edge.NodeA))
		binary.Write(&edgesbuffer, binary.LittleEndian, int32(edge.NodeB))
		binary.Write(&edgesbuffer, binary.LittleEndian, uint8(edge_weight))
		binary.Write(&edgesbuffer, binary.LittleEndian, byte(edge_attrib.Type))
		binary.Write(&edgesbuffer, binary.LittleEndian, edge_attrib.Length)
		binary.Write(&edgesbuffer, binary.LittleEndian, uint8(edge_attrib.Maxspeed))
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

//*******************************************
// load graph information
//*******************************************

func _LoadNodes(file string) (List[Node], List[NodeAttributes], List[EdgeRef]) {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	nodedata, _ := os.ReadFile(file)
	nodereader := bytes.NewReader(nodedata)
	var nodecount int32
	binary.Read(nodereader, binary.LittleEndian, &nodecount)
	var edgerefcount int32
	binary.Read(nodereader, binary.LittleEndian, &edgerefcount)
	nodes := NewList[Node](int(nodecount))
	node_attribs := NewList[NodeAttributes](int(nodecount))
	edge_refs := NewList[EdgeRef](int(edgerefcount))
	for i := 0; i < int(nodecount); i++ {
		var t int8
		binary.Read(nodereader, binary.LittleEndian, &t)
		var s int32
		binary.Read(nodereader, binary.LittleEndian, &s)
		var c int16
		binary.Read(nodereader, binary.LittleEndian, &c)
		nodes.Add(Node{
			EdgeRefStart: s,
			EdgeRefCount: c,
		})
		node_attribs.Add(NodeAttributes{Type: t})
	}
	for i := 0; i < int(edgerefcount); i++ {
		var id int32
		binary.Read(nodereader, binary.LittleEndian, &id)
		var r byte
		binary.Read(nodereader, binary.LittleEndian, &r)
		edge_refs.Add(EdgeRef{
			EdgeID: id,
			Type:   r,
		})
	}

	return nodes, node_attribs, edge_refs
}

func _LoadEdges(file string) (List[Edge], List[EdgeAttributes], List[int32]) {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	edgedata, _ := os.ReadFile(file)
	edgereader := bytes.NewReader(edgedata)
	var edgecount int32
	binary.Read(edgereader, binary.LittleEndian, &edgecount)
	edges := NewList[Edge](int(edgecount))
	edge_attribs := NewList[EdgeAttributes](int(edgecount))
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
			NodeA: a,
			NodeB: b,
		})
		edge_attribs.Add(EdgeAttributes{
			Type:     RoadType(t),
			Length:   l,
			Maxspeed: m,
		})
		edge_weights.Add(int32(w))
	}

	return edges, edge_attribs, edge_weights
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

func _LoadCHShortcuts(file string) (List[Shortcut], List[int32]) {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	shortcutdata, _ := os.ReadFile(file)
	shortcutreader := bytes.NewReader(shortcutdata)
	var shortcutcount int32
	binary.Read(shortcutreader, binary.LittleEndian, &shortcutcount)
	shortcuts := NewList[Shortcut](int(shortcutcount))
	shortcut_weights := NewList[int32](int(shortcutcount))
	for i := 0; i < int(shortcutcount); i++ {
		var node_a int32
		binary.Read(shortcutreader, binary.LittleEndian, &node_a)
		var node_b int32
		binary.Read(shortcutreader, binary.LittleEndian, &node_b)
		var weight uint32
		binary.Read(shortcutreader, binary.LittleEndian, &weight)
		shortcut := Shortcut{
			NodeA: node_a,
			NodeB: node_b,
			Edges: [2]EdgeRef{},
		}
		for j := 0; j < 2; j++ {
			var id int32
			binary.Read(shortcutreader, binary.LittleEndian, &id)
			var is bool
			binary.Read(shortcutreader, binary.LittleEndian, &is)
			if is {
				shortcut.Edges[j] = EdgeRef{
					EdgeID: id,
					Type:   2,
				}
			} else {
				shortcut.Edges[j] = EdgeRef{
					EdgeID: id,
					Type:   0,
				}
			}
		}
		shortcuts.Add(shortcut)
		shortcut_weights.Add(int32(weight))
	}

	return shortcuts, shortcut_weights
}

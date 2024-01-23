package graph

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	"github.com/ttpr0/go-routing/geo"
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// graph base interface
//*******************************************

type IGraphBase interface {
	NodeCount() int
	EdgeCount() int
	GetNode(node int32) Node
	IsNode(node int32) bool
	GetEdge(edge int32) Edge
	IsEdge(edge int32) bool
	GetNodeGeom(node int32) geo.Coord
	GetEdgeGeom(edge int32) geo.CoordArray
	GetAccessor() _AdjArrayAccessor
	GetNodeDegree(node int32, dir Direction) int16
}

//*******************************************
// graph base
//*******************************************

var _ IGraphBase = &GraphBase{}

type GraphBase struct {
	nodes      Array[Node]
	edges      Array[Edge]
	node_geoms []geo.Coord
	edge_geoms []geo.CoordArray
	topology   _AdjacencyArray
}

func (self *GraphBase) NodeCount() int {
	return len(self.nodes)
}
func (self *GraphBase) EdgeCount() int {
	return len(self.edges)
}
func (self *GraphBase) IsNode(node int32) bool {
	if node < int32(len(self.nodes)) {
		return true
	} else {
		return false
	}
}
func (self *GraphBase) GetNode(node int32) Node {
	return self.nodes[node]
}
func (self *GraphBase) IsEdge(edge int32) bool {
	if edge < int32(len(self.edges)) {
		return true
	} else {
		return false
	}
}
func (self *GraphBase) GetEdge(edge int32) Edge {
	return self.edges[edge]
}
func (self *GraphBase) GetNodeGeom(node int32) geo.Coord {
	return self.node_geoms[node]
}
func (self *GraphBase) GetEdgeGeom(edge int32) geo.CoordArray {
	geom := self.edge_geoms[edge]
	if geom == nil {
		e := self.GetEdge(edge)
		geom = make([]geo.Coord, 2)
		geom[0] = self.GetNodeGeom(e.NodeA)
		geom[1] = self.GetNodeGeom(e.NodeB)
	}
	return geom
}
func (self *GraphBase) GetAccessor() _AdjArrayAccessor {
	accessor := self.topology.GetAccessor()
	return accessor
}
func (self *GraphBase) GetNodeDegree(node int32, dir Direction) int16 {
	return self.topology.GetDegree(node, dir)
}

//*******************************************
// modification methods
//*******************************************

func (self *GraphBase) _ReorderNodes(mapping Array[int32]) {
	// nodes
	new_nodes := NewArray[Node](self.NodeCount())
	for i, id := range mapping {
		new_nodes[id] = self.nodes[i]
	}
	self.nodes = new_nodes

	// edges
	for i := 0; i < self.EdgeCount(); i++ {
		edge := self.edges[i]
		edge.NodeA = mapping[edge.NodeA]
		edge.NodeB = mapping[edge.NodeB]
		self.edges[i] = edge
	}

	// geom
	new_node_geoms := NewArray[geo.Coord](len(self.nodes))
	for i, id := range mapping {
		new_node_geoms[id] = self.node_geoms[i]
	}
	self.node_geoms = new_node_geoms

	// others
	self.topology._ReorderNodes(mapping)
}
func (self *GraphBase) _RemoveNodes(nodes List[int32]) {
	remove := NewArray[bool](self.NodeCount())
	for _, n := range nodes {
		remove[n] = true
	}

	new_nodes := NewList[Node](100)
	new_node_geoms := NewList[geo.Coord](100)
	mapping := NewArray[int32](self.NodeCount())
	id := int32(0)
	for i := 0; i < self.NodeCount(); i++ {
		if remove[i] {
			mapping[i] = -1
			continue
		}
		new_nodes.Add(self.GetNode(int32(i)))
		new_node_geoms.Add(self.GetNodeGeom(int32(i)))
		mapping[i] = id
		id += 1
	}
	new_edges := NewList[Edge](100)
	new_edge_geoms := NewList[geo.CoordArray](100)
	for i := 0; i < self.EdgeCount(); i++ {
		edge := self.GetEdge(int32(i))
		if remove[edge.NodeA] || remove[edge.NodeB] {
			continue
		}
		new_edges.Add(Edge{
			NodeA:    mapping[edge.NodeA],
			NodeB:    mapping[edge.NodeB],
			Type:     edge.Type,
			Length:   edge.Length,
			Maxspeed: edge.Maxspeed,
			Oneway:   edge.Oneway,
		})
		new_edge_geoms.Add(self.GetEdgeGeom(int32(i)))
	}

	self.nodes = Array[Node](new_nodes)
	self.edges = Array[Edge](new_edges)
	self.node_geoms = new_node_geoms
	self.edge_geoms = new_edge_geoms
	self.topology = _BuildTopology(self.nodes, self.edges)
}

//*******************************************
// load and store methods
//*******************************************

func (self *GraphBase) _Store(path string) {
	_StoreGraphNodes(self.nodes, path+"-nodes")
	_StoreGraphEdges(self.edges, path+"-edges")
	_StoreGraphGeom(self.node_geoms, self.edge_geoms, path+"-geom")
	_StoreAdjacency(&self.topology, false, path+"-graph")
}

func (self *GraphBase) _New() *GraphBase {
	return &GraphBase{}
}
func (self *GraphBase) _Load2(path string) {
	nodes := _LoadGraphNodes(path + "-nodes")
	nodecount := len(nodes)
	edges := _LoadGraphEdges(path + "-edges")
	edgecount := len(edges)
	node_geoms, edge_geoms := _LoadGraphGeom(path+"-geom", nodecount, edgecount)
	topology := _LoadAdjacency(path+"-graph", false)

	*self = GraphBase{
		nodes:      nodes,
		edges:      edges,
		node_geoms: node_geoms,
		edge_geoms: edge_geoms,
		topology:   *topology,
	}
}

func (self *GraphBase) _Load(path string) {
	nodes := _LoadGraphNodes(path + "-nodes")
	nodecount := len(nodes)
	edges := _LoadGraphEdges(path + "-edges")
	edgecount := len(edges)
	node_geoms, edge_geoms := _LoadGraphGeomMin(path+"-geom", nodecount, edgecount)
	topology := _LoadAdjacency(path+"-graph", false)

	*self = GraphBase{
		nodes:      nodes,
		edges:      edges,
		node_geoms: node_geoms,
		edge_geoms: edge_geoms,
		topology:   *topology,
	}
}

//*******************************************
// load and store components
//*******************************************

func _StoreGraphNodes(nodes Array[Node], filename string) {
	nodesbuffer := bytes.Buffer{}

	nodecount := nodes.Length()
	binary.Write(&nodesbuffer, binary.LittleEndian, int32(nodecount))

	for i := 0; i < nodecount; i++ {
		node := nodes.Get(i)
		binary.Write(&nodesbuffer, binary.LittleEndian, node.Type)
	}

	nodesfile, _ := os.Create(filename)
	defer nodesfile.Close()
	nodesfile.Write(nodesbuffer.Bytes())
}

func _LoadGraphNodes(file string) Array[Node] {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	nodedata, _ := os.ReadFile(file)
	nodereader := bytes.NewReader(nodedata)
	var nodecount int32
	binary.Read(nodereader, binary.LittleEndian, &nodecount)
	nodes := NewList[Node](int(nodecount))
	for i := 0; i < int(nodecount); i++ {
		var t int8
		binary.Read(nodereader, binary.LittleEndian, &t)
		nodes.Add(Node{
			Type: t,
		})
	}

	return Array[Node](nodes)
}

func _StoreGraphEdges(edges Array[Edge], filename string) {
	edgesbuffer := bytes.Buffer{}

	edgecount := edges.Length()
	binary.Write(&edgesbuffer, binary.LittleEndian, int32(edgecount))

	for i := 0; i < edgecount; i++ {
		edge := edges.Get(i)
		binary.Write(&edgesbuffer, binary.LittleEndian, int32(edge.NodeA))
		binary.Write(&edgesbuffer, binary.LittleEndian, int32(edge.NodeB))
		binary.Write(&edgesbuffer, binary.LittleEndian, byte(edge.Type))
		binary.Write(&edgesbuffer, binary.LittleEndian, edge.Length)
		binary.Write(&edgesbuffer, binary.LittleEndian, uint8(edge.Maxspeed))
	}

	edgesfile, _ := os.Create(filename)
	defer edgesfile.Close()
	edgesfile.Write(edgesbuffer.Bytes())
}

func _LoadGraphEdges(file string) Array[Edge] {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	edgedata, _ := os.ReadFile(file)
	edgereader := bytes.NewReader(edgedata)
	var edgecount int32
	binary.Read(edgereader, binary.LittleEndian, &edgecount)
	edges := NewList[Edge](int(edgecount))
	for i := 0; i < int(edgecount); i++ {
		var a int32
		binary.Read(edgereader, binary.LittleEndian, &a)
		var b int32
		binary.Read(edgereader, binary.LittleEndian, &b)
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
	}

	return Array[Edge](edges)
}

func _StoreGraphGeom(nodes []geo.Coord, edges []geo.CoordArray, filename string) {
	geombuffer := bytes.Buffer{}

	nodecount := len(nodes)
	edgecount := len(edges)

	for i := 0; i < nodecount; i++ {
		point := nodes[i]
		binary.Write(&geombuffer, binary.LittleEndian, point[0])
		binary.Write(&geombuffer, binary.LittleEndian, point[1])
	}
	c := 0
	for i := 0; i < edgecount; i++ {
		nc := len(edges[i])
		if nc > 255 {
			nc = 255
		}
		binary.Write(&geombuffer, binary.LittleEndian, int32(c))
		binary.Write(&geombuffer, binary.LittleEndian, uint8(nc))
		c += nc * 8
	}
	for i := 0; i < edgecount; i++ {
		coords := edges[i]
		nc := len(coords)
		if nc > 255 {
			nc = 255
		}
		for j := 0; j < nc; j++ {
			coord := coords[j]
			binary.Write(&geombuffer, binary.LittleEndian, coord[0])
			binary.Write(&geombuffer, binary.LittleEndian, coord[1])
		}
	}

	geomfile, _ := os.Create(filename)
	defer geomfile.Close()
	geomfile.Write(geombuffer.Bytes())
}

func _LoadGraphGeom(file string, nodecount, edgecount int) ([]geo.Coord, []geo.CoordArray) {
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

func _LoadGraphGeomMin(file string, nodecount, edgecount int) ([]geo.Coord, []geo.CoordArray) {
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
		for j := 0; j < int(c); j++ {
			var lon float32
			binary.Read(linereader, binary.LittleEndian, &lon)
			var lat float32
			binary.Read(linereader, binary.LittleEndian, &lat)
		}
		edge_geoms[i] = nil
	}

	return node_geoms, edge_geoms
}

package graph

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"strings"
)

type IGraph interface {
	GetGeometry() IGeometry
	GetWeighting() IWeighting
	GetOtherNode(edge, node int32) (int32, Direction)
	GetAdjacentEdges(node int32) []int32
	ForEachEdge(node int32, f func(int32))
	NodeCount() int32
	EdgeCount() int32
	IsNode(node int32) bool
	GetNode(node int32) NodeAttributes
	GetEdge(edge int32) EdgeAttributes
}

type Edge struct {
	NodeA int32
	NodeB int32
}

type EdgeAttributes struct {
	Type     RoadType
	Length   float32
	Maxspeed byte
	Oneway   bool
}

type Node struct {
	Edges []int32
}
type NodeAttributes struct {
	Type int8
}

type Direction byte

const (
	BACKWARD Direction = 0
	FORWARD  Direction = 1
)

type RoadType int8

const (
	MOTORWAY       RoadType = 1
	MOTORWAY_LINK  RoadType = 2
	TRUNK          RoadType = 3
	TRUNK_LINK     RoadType = 4
	PRIMARY        RoadType = 5
	PRIMARY_LINK   RoadType = 6
	SECONDARY      RoadType = 7
	SECONDARY_LINK RoadType = 8
	TERTIARY       RoadType = 9
	TERTIARY_LINK  RoadType = 10
	RESIDENTIAL    RoadType = 11
	LIVING_STREET  RoadType = 12
	UNCLASSIFIED   RoadType = 13
	ROAD           RoadType = 14
	TRACK          RoadType = 15
)

type Graph struct {
	edges           []Edge
	edge_attributes []EdgeAttributes
	nodes           []Node
	node_attributes []NodeAttributes
	geom            IGeometry
	weight          IWeighting
}

func (self *Graph) GetGeometry() IGeometry {
	return self.geom
}
func (self *Graph) GetWeighting() IWeighting {
	return self.weight
}
func (self *Graph) GetOtherNode(edge, node int32) (int32, Direction) {
	e := self.edges[edge]
	if node == e.NodeA {
		return e.NodeB, FORWARD
	}
	if node == e.NodeB {
		return e.NodeA, BACKWARD
	}
	return 0, 0
}
func (self *Graph) GetAdjacentEdges(node int32) []int32 {
	return self.nodes[node].Edges
}
func (self *Graph) ForEachEdge(node int32, f func(int32)) {

}
func (self *Graph) NodeCount() int32 {
	return int32(len(self.nodes))
}
func (self *Graph) EdgeCount() int32 {
	return int32(len(self.edges))
}
func (self *Graph) IsNode(node int32) bool {
	if node < int32(len(self.nodes)) {
		return true
	} else {
		return false
	}
}
func (self *Graph) GetNode(node int32) NodeAttributes {
	return self.node_attributes[node]
}
func (self *Graph) GetEdge(edge int32) EdgeAttributes {
	return self.edge_attributes[edge]
}

func LoadGraph(file string) IGraph {
	file_info, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) || strings.Split(file_info.Name(), ".")[1] != "graph" {
		panic("file not found")
	}

	graphdata, _ := os.ReadFile(file)
	graphreader := bytes.NewReader(graphdata)
	var nodecount int32
	binary.Read(graphreader, binary.LittleEndian, &nodecount)
	var edgecount int32
	binary.Read(graphreader, binary.LittleEndian, &edgecount)
	startindex := 8 + nodecount*5 + edgecount*8
	edgerefreader := bytes.NewReader(graphdata[startindex:])
	nodearr := make([]Node, nodecount)
	for i := 0; i < int(nodecount); i++ {
		var s int32
		binary.Read(graphreader, binary.LittleEndian, &s)
		var c int8
		binary.Read(graphreader, binary.LittleEndian, &c)
		edges := make([]int32, c)
		for j := 0; j < int(c); j++ {
			var e int32
			binary.Read(edgerefreader, binary.LittleEndian, &e)
			edges[j] = e
		}
		nodearr[i] = Node{Edges: edges}
	}
	edgearr := make([]Edge, edgecount)
	for i := 0; i < int(edgecount); i++ {
		var start int32
		binary.Read(graphreader, binary.LittleEndian, &start)
		var end int32
		binary.Read(graphreader, binary.LittleEndian, &end)
		edgearr[i] = Edge{NodeA: start, NodeB: end}
	}

	attribdata, _ := os.ReadFile(strings.Replace(file, ".graph", "-attrib", 1))
	attribreader := bytes.NewReader(attribdata)
	nodeattribarr := make([]NodeAttributes, nodecount)
	for i := 0; i < int(nodecount); i++ {
		var t int8
		binary.Read(attribreader, binary.LittleEndian, &t)
		nodeattribarr[i] = NodeAttributes{Type: t}
	}
	edgeattribarr := make([]EdgeAttributes, edgecount)
	for i := 0; i < int(edgecount); i++ {
		var t int8
		binary.Read(attribreader, binary.LittleEndian, &t)
		var length float32
		binary.Read(attribreader, binary.LittleEndian, &length)
		var maxspeed byte
		binary.Read(attribreader, binary.LittleEndian, &maxspeed)
		var oneway bool
		binary.Read(attribreader, binary.LittleEndian, &oneway)
		edgeattribarr[i] = EdgeAttributes{Type: RoadType(t), Length: length, Maxspeed: maxspeed, Oneway: oneway}
	}

	weightdata, _ := os.ReadFile(strings.Replace(file, ".graph", "-weight", 1))
	weightreader := bytes.NewReader(weightdata)
	edgeweights := make([]int32, edgecount)
	for i := 0; i < int(edgecount); i++ {
		var w byte
		binary.Read(weightreader, binary.LittleEndian, &w)
		edgeweights[i] = int32(w)
	}

	geomdata, _ := os.ReadFile(strings.Replace(file, ".graph", "-geom", 1))
	startindex = nodecount*8 + edgecount*5
	geomreader := bytes.NewReader(geomdata)
	linereader := bytes.NewReader(geomdata[startindex:])
	pointarr := make([]Coord, nodecount)
	for i := 0; i < int(nodecount); i++ {
		var lon float32
		binary.Read(geomreader, binary.LittleEndian, &lon)
		var lat float32
		binary.Read(geomreader, binary.LittleEndian, &lat)
		pointarr[i] = Coord{lon, lat}
	}
	linearr := make([]CoordArray, edgecount)
	for i := 0; i < int(edgecount); i++ {
		var s int32
		binary.Read(geomreader, binary.LittleEndian, &s)
		var c byte
		binary.Read(geomreader, binary.LittleEndian, &c)
		points := make([]Coord, c)
		for j := 0; j < int(c); j++ {
			var lon float32
			binary.Read(linereader, binary.LittleEndian, &lon)
			var lat float32
			binary.Read(linereader, binary.LittleEndian, &lat)
			points[j].Lon = lon
			points[j].Lat = lat
		}
		linearr[i] = points
	}

	return &Graph{
		edges:           edgearr,
		edge_attributes: edgeattribarr,
		nodes:           nodearr,
		node_attributes: nodeattribarr,
		geom:            &Geometry{pointarr, linearr},
		weight:          &Weighting{edgeweights},
	}
}

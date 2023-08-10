package graph

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type IGeometry interface {
	GetNode(node int32) geo.Coord
	GetEdge(edge int32) geo.CoordArray
	GetAllNodes() []geo.Coord
	GetAllEdges() []geo.CoordArray
}

type GeometryStore struct {
	nodes []geo.Coord
	edges []geo.CoordArray
}

func (self *GeometryStore) GetNode(node int32) geo.Coord {
	return self.nodes[node]
}
func (self *GeometryStore) GetEdge(edge int32) geo.CoordArray {
	return self.edges[edge]
}
func (self *GeometryStore) GetAllNodes() []geo.Coord {
	return self.nodes
}
func (self *GeometryStore) GetAllEdges() []geo.CoordArray {
	return self.edges
}

// reorders node information in geometry store,
// node mapping: old id -> new id
func (self *GeometryStore) _ReorderNodes(mapping Array[int32]) {
	new_nodes := NewArray[geo.Coord](len(self.nodes))
	for i, id := range mapping {
		new_nodes[id] = self.nodes[i]
	}
	self.nodes = new_nodes
}

func (self *GeometryStore) _Store(filename string) {
	geombuffer := bytes.Buffer{}

	nodecount := len(self.nodes)
	edgecount := len(self.edges)

	for i := 0; i < nodecount; i++ {
		point := self.GetNode(int32(i))
		binary.Write(&geombuffer, binary.LittleEndian, point[0])
		binary.Write(&geombuffer, binary.LittleEndian, point[1])
	}
	c := 0
	for i := 0; i < edgecount; i++ {
		nc := len(self.GetEdge(int32(i)))
		binary.Write(&geombuffer, binary.LittleEndian, int32(c))
		binary.Write(&geombuffer, binary.LittleEndian, uint8(nc))
		c += nc * 8
	}
	for i := 0; i < edgecount; i++ {
		coords := self.GetEdge(int32(i))
		for _, coord := range coords {
			binary.Write(&geombuffer, binary.LittleEndian, coord[0])
			binary.Write(&geombuffer, binary.LittleEndian, coord[1])
		}
	}

	geomfile, _ := os.Create(filename)
	defer geomfile.Close()
	geomfile.Write(geombuffer.Bytes())
}

func _LoadGeometryStore(file string, nodecount, edgecount int) *GeometryStore {
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

	return &GeometryStore{
		nodes: node_geoms,
		edges: edge_geoms,
	}
}

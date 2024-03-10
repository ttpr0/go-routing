package attr

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	"github.com/ttpr0/go-routing/geo"
	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

type IAttributes interface {
	GetNodeAttribs(node int32) NodeAttribs
	GetEdgeAttribs(edge int32) EdgeAttribs
	GetNodeGeom(node int32) geo.Coord
	GetEdgeGeom(edge int32) geo.CoordArray
}

type GraphAttributes struct {
	node_attribs Array[NodeAttribs]
	edge_attribs Array[EdgeAttribs]
	node_geoms   []geo.Coord
	edge_geoms   []geo.CoordArray
}

func New(nodes Array[NodeAttribs], edges Array[EdgeAttribs], node_geoms Array[geo.Coord], edge_geoms Array[geo.CoordArray]) *GraphAttributes {
	return &GraphAttributes{
		node_attribs: nodes,
		edge_attribs: edges,
		node_geoms:   node_geoms,
		edge_geoms:   edge_geoms,
	}
}

func (self *GraphAttributes) GetNodeAttribs(node int32) NodeAttribs {
	return self.node_attribs[node]
}
func (self *GraphAttributes) GetEdgeAttribs(edge int32) EdgeAttribs {
	return self.edge_attribs[edge]
}
func (self *GraphAttributes) GetNodeGeom(node int32) geo.Coord {
	return self.node_geoms[node]
}
func (self *GraphAttributes) GetEdgeGeom(edge int32) geo.CoordArray {
	geom := self.edge_geoms[edge]
	return geom
}

func NewMappedAttributes(attributes IAttributes, node_mapping Optional[structs.IDMapping], edge_mapping Optional[structs.IDMapping]) *MappedAttributes {
	return &MappedAttributes{
		attributes:   attributes,
		node_mapping: node_mapping,
		edge_mapping: edge_mapping,
	}
}

type MappedAttributes struct {
	attributes   IAttributes
	node_mapping Optional[structs.IDMapping]
	edge_mapping Optional[structs.IDMapping]
}

func (self *MappedAttributes) GetNodeAttribs(node int32) NodeAttribs {
	var m_node int32
	if self.node_mapping.HasValue() {
		m_node = self.node_mapping.Value.GetTarget(node)
	} else {
		m_node = node
	}
	return self.attributes.GetNodeAttribs(m_node)
}
func (self *MappedAttributes) GetEdgeAttribs(edge int32) EdgeAttribs {
	var m_edge int32
	if self.edge_mapping.HasValue() {
		m_edge = self.edge_mapping.Value.GetTarget(edge)
	} else {
		m_edge = edge
	}
	return self.attributes.GetEdgeAttribs(m_edge)
}
func (self *MappedAttributes) GetNodeGeom(node int32) geo.Coord {
	var m_node int32
	if self.node_mapping.HasValue() {
		m_node = self.node_mapping.Value.GetTarget(node)
	} else {
		m_node = node
	}
	return self.attributes.GetNodeGeom(m_node)
}
func (self *MappedAttributes) GetEdgeGeom(edge int32) geo.CoordArray {
	var m_edge int32
	if self.edge_mapping.HasValue() {
		m_edge = self.edge_mapping.Value.GetTarget(edge)
	} else {
		m_edge = edge
	}
	geom := self.attributes.GetEdgeGeom(m_edge)
	return geom
}

//*******************************************
// modification methods
//*******************************************

func (self *GraphAttributes) ReorderNodes(mapping Array[int32]) {
	// nodes
	new_nodes := NewArray[NodeAttribs](len(self.node_attribs))
	for i, id := range mapping {
		new_nodes[id] = self.node_attribs[i]
	}
	self.node_attribs = new_nodes

	// geom
	new_node_geoms := NewArray[geo.Coord](len(self.node_geoms))
	for i, id := range mapping {
		new_node_geoms[id] = self.node_geoms[i]
	}
	self.node_geoms = new_node_geoms
}
func (self *GraphAttributes) RemoveNodes(nodes List[int32]) {
	remove := NewArray[bool](len(self.node_attribs))
	for _, n := range nodes {
		remove[n] = true
	}

	new_nodes := NewList[NodeAttribs](100)
	new_node_geoms := NewList[geo.Coord](100)
	for i := 0; i < len(self.node_attribs); i++ {
		if remove[i] {
			continue
		}
		new_nodes.Add(self.node_attribs[i])
		new_node_geoms.Add(self.node_geoms[i])
	}

	self.node_attribs = Array[NodeAttribs](new_nodes)
	self.node_geoms = new_node_geoms
}
func (self *GraphAttributes) RemoveEdges(edges List[int32]) {
	remove := NewArray[bool](len(self.edge_attribs))
	for _, n := range edges {
		remove[n] = true
	}

	new_edges := NewList[EdgeAttribs](100)
	new_edge_geoms := NewList[geo.CoordArray](100)
	for i := 0; i < len(self.edge_attribs); i++ {
		if remove[i] {
			continue
		}
		new_edges.Add(self.edge_attribs[i])
		new_edge_geoms.Add(self.GetEdgeGeom(int32(i)))
	}

	self.edge_attribs = Array[EdgeAttribs](new_edges)
	self.edge_geoms = new_edge_geoms
}

//*******************************************
// load and store methods
//*******************************************

func Store(attr *GraphAttributes, path string) {
	attrib_writer := NewBufferWriter()
	nodecount := attr.node_attribs.Length()
	edgecount := attr.edge_attribs.Length()
	Write[int32](attrib_writer, int32(nodecount))
	Write[int32](attrib_writer, int32(edgecount))
	for i := 0; i < nodecount; i++ {
		node := attr.node_attribs[i]
		Write[int8](attrib_writer, node.Type)
	}
	for i := 0; i < edgecount; i++ {
		edge := attr.edge_attribs[i]
		Write(attrib_writer, byte(edge.Type))
		Write(attrib_writer, edge.Length)
		Write(attrib_writer, uint8(edge.Maxspeed))
	}
	attrfile, _ := os.Create(path + "-attrib")
	defer attrfile.Close()
	attrfile.Write(attrib_writer.Bytes())

	_StoreGraphGeom(attr.node_geoms, attr.edge_geoms, path+"-geom")
}

func Load(path string) *GraphAttributes {
	_, err := os.Stat(path + "-attrib")
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + path + "-attrib")
	}

	attrdata, _ := os.ReadFile(path + "-attrib")
	attr_reader := NewBufferReader(attrdata)

	nodecount := int(Read[int32](attr_reader))
	edgecount := int(Read[int32](attr_reader))

	nodes := NewArray[NodeAttribs](nodecount)
	edges := NewArray[EdgeAttribs](edgecount)
	for i := 0; i < nodecount; i++ {
		typ := Read[int8](attr_reader)
		nodes[i] = NodeAttribs{Type: typ}
	}
	for i := 0; i < edgecount; i++ {
		typ := Read[byte](attr_reader)
		length := Read[float32](attr_reader)
		maxspeed := Read[uint8](attr_reader)
		edges[i] = EdgeAttribs{
			Type:     RoadType(typ),
			Length:   length,
			Maxspeed: maxspeed,
		}
	}

	node_geoms, edge_geoms := _LoadGraphGeom(path+"-geom", nodecount, edgecount)

	return &GraphAttributes{
		node_attribs: nodes,
		edge_attribs: edges,
		node_geoms:   node_geoms,
		edge_geoms:   edge_geoms,
	}
}

func LoadMin(path string) *GraphAttributes {
	_, err := os.Stat(path + "-attrib")
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + path + "-attrib")
	}

	attrdata, _ := os.ReadFile(path + "-attrib")
	attr_reader := NewBufferReader(attrdata)

	nodecount := int(Read[int32](attr_reader))
	edgecount := int(Read[int32](attr_reader))

	nodes := NewArray[NodeAttribs](nodecount)
	edges := NewArray[EdgeAttribs](edgecount)
	for i := 0; i < nodecount; i++ {
		typ := Read[int8](attr_reader)
		nodes[i] = NodeAttribs{Type: typ}
	}
	for i := 0; i < edgecount; i++ {
		typ := Read[byte](attr_reader)
		length := Read[float32](attr_reader)
		maxspeed := Read[uint8](attr_reader)
		edges[i] = EdgeAttribs{
			Type:     RoadType(typ),
			Length:   length,
			Maxspeed: maxspeed,
		}
	}

	node_geoms, edge_geoms := _LoadGraphGeomMin(path+"-geom", nodecount, edgecount)

	return &GraphAttributes{
		node_attribs: nodes,
		edge_attribs: edges,
		node_geoms:   node_geoms,
		edge_geoms:   edge_geoms,
	}
}

//*******************************************
// load and store components
//*******************************************

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

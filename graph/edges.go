package graph

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type EdgeStore struct {
	edges Array[Edge]
}

func (self *EdgeStore) GetEdge(index int32) Edge {
	return self.edges[index]
}
func (self *EdgeStore) SetEdge(index int32, edge Edge) {
	self.edges[index] = edge
}
func (self *EdgeStore) IsEdge(index int32) bool {
	if index < int32(len(self.edges)) {
		return true
	} else {
		return false
	}
}
func (self *EdgeStore) EdgeCount() int {
	return self.edges.Length()
}
func (self *EdgeStore) GetAllEdges() Array[Edge] {
	return self.edges
}

// reorders node information in edgestore,
// mapping: old id -> new id
func (self *EdgeStore) _ReorderNodes(mapping Array[int32]) {
	for i := 0; i < self.EdgeCount(); i++ {
		edge := self.edges[i]
		edge.NodeA = mapping[edge.NodeA]
		edge.NodeB = mapping[edge.NodeB]
		self.edges[i] = edge
	}
}

func (self *EdgeStore) _Store(filename string) {
	edgesbuffer := bytes.Buffer{}

	edgecount := self.edges.Length()
	binary.Write(&edgesbuffer, binary.LittleEndian, int32(edgecount))

	for i := 0; i < edgecount; i++ {
		edge := self.edges.Get(i)
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

func _LoadEdgeStore(file string) *EdgeStore {
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

	return &EdgeStore{
		edges: Array[Edge](edges),
	}
}

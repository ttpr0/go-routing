package graph

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type NodeStore struct {
	nodes Array[Node]
}

func (self *NodeStore) GetNode(index int32) Node {
	return self.nodes[index]
}
func (self *NodeStore) SetNode(index int32, node Node) {
	self.nodes[index] = node
}
func (self *NodeStore) IsNode(index int32) bool {
	if index < int32(len(self.nodes)) {
		return true
	} else {
		return false
	}
}
func (self *NodeStore) NodeCount() int {
	return self.nodes.Length()
}

// reorders nodes in nodestore,
// mapping: old id -> new id
func (self *NodeStore) _ReorderNodes(mapping Array[int32]) {
	new_nodes := NewArray[Node](self.NodeCount())
	for i, id := range mapping {
		new_nodes[id] = self.nodes[i]
	}
	self.nodes = new_nodes
}

func (self *NodeStore) _Store(filename string) {
	nodesbuffer := bytes.Buffer{}

	nodecount := self.nodes.Length()
	binary.Write(&nodesbuffer, binary.LittleEndian, int32(nodecount))

	for i := 0; i < nodecount; i++ {
		node := self.nodes.Get(i)
		binary.Write(&nodesbuffer, binary.LittleEndian, node.Type)
	}

	nodesfile, _ := os.Create(filename)
	defer nodesfile.Close()
	nodesfile.Write(nodesbuffer.Bytes())
}

func _LoadNodeStore(file string) *NodeStore {
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

	return &NodeStore{
		nodes: Array[Node](nodes),
	}
}

package comps

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// graph base interface
//*******************************************

type IGraphBase interface {
	NodeCount() int
	EdgeCount() int
	GetNode(node int32) structs.Node
	IsNode(node int32) bool
	GetEdge(edge int32) structs.Edge
	IsEdge(edge int32) bool
	GetAccessor() structs.IAdjAccessor
	GetNodeDegree(node int32, forward bool) int16
}

//*******************************************
// graph base
//*******************************************

var _ IGraphBase = &GraphBase{}

type GraphBase struct {
	nodes    Array[structs.Node]
	edges    Array[structs.Edge]
	topology structs.AdjacencyArray
}

func NewGraphBase(nodes Array[structs.Node], edges Array[structs.Edge]) *GraphBase {
	topology := _BuildTopology(nodes, edges)
	return &GraphBase{
		nodes:    nodes,
		edges:    edges,
		topology: topology,
	}
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
func (self *GraphBase) GetNode(node int32) structs.Node {
	return self.nodes[node]
}
func (self *GraphBase) IsEdge(edge int32) bool {
	if edge < int32(len(self.edges)) {
		return true
	} else {
		return false
	}
}
func (self *GraphBase) GetEdge(edge int32) structs.Edge {
	return self.edges[edge]
}
func (self *GraphBase) GetAccessor() structs.IAdjAccessor {
	accessor := self.topology.GetAccessor()
	return &accessor
}
func (self *GraphBase) GetNodeDegree(node int32, forward bool) int16 {
	return self.topology.GetDegree(node, forward)
}

//*******************************************
// modification methods
//*******************************************

func (self *GraphBase) _ReorderNodes(mapping Array[int32]) *GraphBase {
	// nodes
	new_nodes := NewArray[structs.Node](self.NodeCount())
	for i, id := range mapping {
		new_nodes[id] = self.nodes[i]
	}

	// edges
	new_edges := NewArray[structs.Edge](self.EdgeCount())
	for i := 0; i < self.EdgeCount(); i++ {
		edge := self.edges[i]
		edge.NodeA = mapping[edge.NodeA]
		edge.NodeB = mapping[edge.NodeB]
		new_edges[i] = edge
	}

	// others
	new_topology := self.topology.ReorderNodes(mapping)

	return &GraphBase{
		nodes:    new_nodes,
		edges:    new_edges,
		topology: new_topology,
	}
}
func (self *GraphBase) _RemoveNodes(nodes List[int32]) *GraphBase {
	remove := NewArray[bool](self.NodeCount())
	for _, n := range nodes {
		remove[n] = true
	}

	new_nodes := NewList[structs.Node](100)
	mapping := NewArray[int32](self.NodeCount())
	id := int32(0)
	for i := 0; i < self.NodeCount(); i++ {
		if remove[i] {
			mapping[i] = -1
			continue
		}
		new_nodes.Add(self.GetNode(int32(i)))
		mapping[i] = id
		id += 1
	}
	new_edges := NewList[structs.Edge](100)
	for i := 0; i < self.EdgeCount(); i++ {
		edge := self.GetEdge(int32(i))
		if remove[edge.NodeA] || remove[edge.NodeB] {
			continue
		}
		new_edges.Add(structs.Edge{
			NodeA: mapping[edge.NodeA],
			NodeB: mapping[edge.NodeB],
		})
	}

	return &GraphBase{
		nodes:    Array[structs.Node](new_nodes),
		edges:    Array[structs.Edge](new_edges),
		topology: _BuildTopology(Array[structs.Node](new_nodes), Array[structs.Edge](new_edges)),
	}
}
func (self *GraphBase) _RemoveEdges(edges List[int32]) *GraphBase {
	panic("not implemented")
}

//*******************************************
// load and store methods
//*******************************************

func (self *GraphBase) _Store(path string) {
	_StoreGraphNodes(self.nodes, path+"-nodes")
	_StoreGraphEdges(self.edges, path+"-edges")
	structs.StoreAdjacency(&self.topology, false, path+"-graph")
}

func (self *GraphBase) _New() *GraphBase {
	return &GraphBase{}
}
func (self *GraphBase) _Load2(path string) {
	nodes := _LoadGraphNodes(path + "-nodes")
	edges := _LoadGraphEdges(path + "-edges")
	topology := structs.LoadAdjacency(path+"-graph", false)

	*self = GraphBase{
		nodes:    nodes,
		edges:    edges,
		topology: *topology,
	}
}

func (self *GraphBase) _Load(path string) {
	nodes := _LoadGraphNodes(path + "-nodes")
	edges := _LoadGraphEdges(path + "-edges")
	topology := structs.LoadAdjacency(path+"-graph", false)

	*self = GraphBase{
		nodes:    nodes,
		edges:    edges,
		topology: *topology,
	}
}

//*******************************************
// load and store components
//*******************************************

func _StoreGraphNodes(nodes Array[structs.Node], filename string) {
	nodesbuffer := bytes.Buffer{}

	nodecount := nodes.Length()
	binary.Write(&nodesbuffer, binary.LittleEndian, int32(nodecount))

	for i := 0; i < nodecount; i++ {
		node := nodes.Get(i)
		binary.Write(&nodesbuffer, binary.LittleEndian, node.Loc)
	}

	nodesfile, _ := os.Create(filename)
	defer nodesfile.Close()
	nodesfile.Write(nodesbuffer.Bytes())
}

func _LoadGraphNodes(file string) Array[structs.Node] {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	nodedata, _ := os.ReadFile(file)
	nodereader := bytes.NewReader(nodedata)
	var nodecount int32
	binary.Read(nodereader, binary.LittleEndian, &nodecount)
	nodes := NewList[structs.Node](int(nodecount))
	for i := 0; i < int(nodecount); i++ {
		var c [2]float32
		binary.Read(nodereader, binary.LittleEndian, &c)
		nodes.Add(structs.Node{
			Loc: c,
		})
	}

	return Array[structs.Node](nodes)
}

func _StoreGraphEdges(edges Array[structs.Edge], filename string) {
	edgesbuffer := bytes.Buffer{}

	edgecount := edges.Length()
	binary.Write(&edgesbuffer, binary.LittleEndian, int32(edgecount))

	for i := 0; i < edgecount; i++ {
		edge := edges.Get(i)
		binary.Write(&edgesbuffer, binary.LittleEndian, int32(edge.NodeA))
		binary.Write(&edgesbuffer, binary.LittleEndian, int32(edge.NodeB))
	}

	edgesfile, _ := os.Create(filename)
	defer edgesfile.Close()
	edgesfile.Write(edgesbuffer.Bytes())
}

func _LoadGraphEdges(file string) Array[structs.Edge] {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	edgedata, _ := os.ReadFile(file)
	edgereader := bytes.NewReader(edgedata)
	var edgecount int32
	binary.Read(edgereader, binary.LittleEndian, &edgecount)
	edges := NewList[structs.Edge](int(edgecount))
	for i := 0; i < int(edgecount); i++ {
		var a int32
		binary.Read(edgereader, binary.LittleEndian, &a)
		var b int32
		binary.Read(edgereader, binary.LittleEndian, &b)
		edges.Add(structs.Edge{
			NodeA: a,
			NodeB: b,
		})
	}

	return Array[structs.Edge](edges)
}

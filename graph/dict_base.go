package graph

import (
	"github.com/ttpr0/go-routing/geo"
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// dictionary graph
//******************************************

// Graph implementation using dictionaries.
// Mainly for testing purposes.
type DictBase struct {
	nodes        Dict[int32, Node]
	edges        Dict[int32, Edge]
	fwd_edgerefs Dict[int32, List[EdgeRef]]
	bwd_edgerefs Dict[int32, List[EdgeRef]]
	weights      Dict[int32, int32]

	max_node_id int32
	max_edge_id int32
}

func NewDictBase() *DictBase {
	return &DictBase{
		nodes:        NewDict[int32, Node](10),
		edges:        NewDict[int32, Edge](10),
		fwd_edgerefs: NewDict[int32, List[EdgeRef]](10),
		bwd_edgerefs: NewDict[int32, List[EdgeRef]](10),
		weights:      NewDict[int32, int32](10),

		max_node_id: 0,
		max_edge_id: 0,
	}
}

func (self *DictBase) NodeCount() int {
	return int(self.max_node_id)
}
func (self *DictBase) EdgeCount() int {
	return int(self.max_edge_id)
}
func (self *DictBase) IsNode(node int32) bool {
	return self.nodes.ContainsKey(node)
}
func (self *DictBase) GetNode(node int32) Node {
	return self.nodes[node]
}
func (self *DictBase) IsEdge(edge int32) bool {
	return self.edges.ContainsKey(edge)
}
func (self *DictBase) GetEdge(edge int32) Edge {
	return self.edges[edge]
}
func (self *DictBase) GetAccessor() IAdjacencyAccessor {
	return &_DictBaseAccessor{
		base: self,
	}
}
func (self *DictBase) GetNodeDegree(node int32, dir Direction) int16 {
	if dir == FORWARD {
		edge_refs := self.fwd_edgerefs[node]
		return int16(edge_refs.Length())
	} else {
		edge_refs := self.bwd_edgerefs[node]
		return int16(edge_refs.Length())
	}
}
func (self *DictBase) GetWeighting() IWeighting {
	return NewDynamicWeighting(func(edge int32) int32 {
		return self.weights[edge]
	})
}

func (self *DictBase) AddNode(id int32, point geo.Coord) {
	if self.nodes.ContainsKey(id) {
		panic("node already exists")
	}
	if id >= self.max_node_id {
		self.max_node_id = id + 1
	}
	self.nodes[id] = Node{Loc: point}
	self.fwd_edgerefs[id] = NewList[EdgeRef](2)
	self.bwd_edgerefs[id] = NewList[EdgeRef](2)
}
func (self *DictBase) AddEdge(node_a, node_b int32, weight int32) {
	if !self.nodes.ContainsKey(node_a) {
		self.AddNode(node_a, geo.Coord{})
	}
	if !self.nodes.ContainsKey(node_b) {
		self.AddNode(node_b, geo.Coord{})
	}
	id := self.max_edge_id
	self.max_edge_id = id + 1
	self.edges[id] = Edge{
		NodeA: node_a,
		NodeB: node_b,
	}
	self.weights[id] = weight
	fwd_edge_refs := self.fwd_edgerefs[node_a]
	fwd_edge_refs.Add(EdgeRef{EdgeID: id, OtherID: node_b, _Type: 0})
	self.fwd_edgerefs[node_a] = fwd_edge_refs
	bwd_edge_refs := self.bwd_edgerefs[node_b]
	bwd_edge_refs.Add(EdgeRef{EdgeID: id, OtherID: node_a, _Type: 0})
	self.bwd_edgerefs[node_b] = bwd_edge_refs
}
func (self *DictBase) RemoveNode(id int32) {
	if !self.nodes.ContainsKey(id) {
		panic("node doesn't exists")
	}
	self.nodes.Delete(id)
	for _, ref := range self.fwd_edgerefs[id] {
		self.RemoveEdge(ref.EdgeID)
	}
	self.fwd_edgerefs.Delete(id)
	for _, ref := range self.bwd_edgerefs[id] {
		self.RemoveEdge(ref.EdgeID)
	}
	self.bwd_edgerefs.Delete(id)
}
func (self *DictBase) RemoveEdge(id int32) {
	if !self.edges.ContainsKey(id) {
		panic("edge doesn't exists")
	}
	edge := self.edges[id]
	// remove fwd edgeref
	fwd_edgerefs := self.fwd_edgerefs[edge.NodeA]
	var index int
	for i, ref := range fwd_edgerefs {
		if ref.EdgeID == id {
			index = i
			break
		}
	}
	fwd_edgerefs.Remove(index)
	self.fwd_edgerefs[edge.NodeA] = fwd_edgerefs
	// remove bwd edgeref
	bwd_edgerefs := self.bwd_edgerefs[edge.NodeB]
	for i, ref := range bwd_edgerefs {
		if ref.EdgeID == id {
			index = i
			break
		}
	}
	bwd_edgerefs.Remove(index)
	self.bwd_edgerefs[edge.NodeB] = bwd_edgerefs
	// remove edge
	self.edges.Delete(id)
}

//*******************************************
// dict-base adjacency accessor
//******************************************

type _DictBaseAccessor struct {
	base          *DictBase
	state         int32
	end           int32
	edge_refs     List[EdgeRef]
	curr_edge_id  int32
	curr_other_id int32
	curr_type     byte
}

func (self *_DictBaseAccessor) SetBaseNode(node int32, dir Direction) {
	var edge_refs List[EdgeRef]
	if dir == FORWARD {
		edge_refs = self.base.fwd_edgerefs[node]

	} else {
		edge_refs = self.base.bwd_edgerefs[node]
	}
	self.state = 0
	self.end = int32(edge_refs.Length())
	self.edge_refs = edge_refs
}
func (self *_DictBaseAccessor) Next() bool {
	if self.state == self.end {
		return false
	}
	ref := self.edge_refs[self.state]
	self.curr_edge_id = ref.EdgeID
	self.curr_other_id = ref.OtherID
	self.curr_type = ref._Type
	self.state += 1
	return true
}
func (self *_DictBaseAccessor) GetEdgeID() int32 {
	return self.curr_edge_id
}
func (self *_DictBaseAccessor) GetOtherID() int32 {
	return self.curr_other_id
}
func (self *_DictBaseAccessor) GetType() byte {
	return self.curr_type
}

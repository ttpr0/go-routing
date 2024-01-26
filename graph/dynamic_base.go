package graph

import (
	"github.com/ttpr0/go-routing/geo"
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// dynamic graph
//*******************************************

type DynamicBase struct {
	// graph store
	nodes   List[Node]
	edges   List[Edge]
	is_node List[bool]
	is_edge List[bool]

	// graph adjacency
	topology _AdjacencyList

	// weighting
	edge_weights List[int32]
}

func NewDynamicBase(init_cap int) *DynamicBase {
	return &DynamicBase{
		nodes:        NewList[Node](init_cap),
		edges:        NewList[Edge](init_cap),
		is_node:      NewList[bool](init_cap),
		is_edge:      NewList[bool](init_cap),
		topology:     _NewAdjacencyList(init_cap),
		edge_weights: NewList[int32](init_cap),
	}
}

// func NewDynGraphFromGraph(graph *Graph) *DynamicBase {
// 	is_node := NewList[bool](graph.NodeCount())
// 	for i := 0; i < graph.NodeCount(); i++ {
// 		is_node.Add(true)
// 	}
// 	is_edge := NewList[bool](graph.EdgeCount())
// 	for i := 0; i < graph.EdgeCount(); i++ {
// 		is_edge.Add(true)
// 	}
// 	topology := NewAdjacencyList(graph.NodeCount())
// 	accessor := graph.base.GetAccessor()
// 	for i := 0; i < graph.NodeCount(); i++ {
// 		accessor.SetBaseNode(int32(i), FORWARD)
// 		for accessor.Next() {
// 			topology.AddFWDEntry(int32(i), accessor.GetOtherID(), accessor.GetEdgeID(), 0)
// 		}
// 		accessor.SetBaseNode(int32(i), BACKWARD)
// 		for accessor.Next() {
// 			topology.AddBWDEntry(accessor.GetOtherID(), int32(i), accessor.GetEdgeID(), 0)
// 		}
// 	}

// 	panic("TODO")
// 	return &DynamicBase{
// 		nodes:      List[Node](graph.base.nodes),
// 		edges:      List[Edge](graph.base.edges),
// 		node_geoms: graph.base.node_geoms,
// 		edge_geoms: graph.base.edge_geoms,
// 		is_node:    is_node,
// 		is_edge:    is_edge,
// 		topology:   topology,
// 		// edge_weights: graph.weight.edge_weights,
// 	}
// }

func (self *DynamicBase) NodeCount() int {
	return self.nodes.Length()
}
func (self *DynamicBase) EdgeCount() int {
	return self.edges.Length()
}
func (self *DynamicBase) IsNode(node int32) bool {
	return self.is_node[node]
}
func (self *DynamicBase) IsEdge(edge int32) bool {
	return self.is_edge[edge]
}
func (self *DynamicBase) GetNode(node int32) Node {
	return self.nodes[node]
}
func (self *DynamicBase) GetEdge(edge int32) Edge {
	return self.edges[edge]
}
func (self *DynamicBase) GetAccessor() IAdjacencyAccessor {
	accessor := self.topology.GetAccessor()
	return &accessor
}
func (self *DynamicBase) GetNodeDegree(node int32, dir Direction) int16 {
	return self.topology.GetDegree(node, dir)
}
func (self *DynamicBase) GetWeighting() IWeighting {
	return NewDynamicWeighting(func(edge int32) int32 {
		return self.edge_weights[edge]
	})
}

func (self *DynamicBase) AddNode(point geo.Coord) int32 {
	id := int32(self.NodeCount())
	self.nodes.Add(Node{Loc: point})
	self.is_node.Add(true)
	if self.topology.node_entries.Length() < self.NodeCount() {
		self.topology.AddNodeEntry()
	}
	return id
}
func (self *DynamicBase) AddEdge(node_a, node_b int32, weight int32) int32 {
	id := int32(self.EdgeCount())
	edge := Edge{NodeA: node_a, NodeB: node_b}
	self.edges.Add(edge)
	self.is_edge.Add(true)
	self.topology.AddFWDEntry(edge.NodeA, edge.NodeB, id, 0)
	self.topology.AddBWDEntry(edge.NodeA, edge.NodeB, id, 0)
	self.edge_weights.Add(weight)
	return id
}
func (self *DynamicBase) RemoveNode(id int32) {
	if self.IsNode(id) {
		self.is_node[id] = false
	}
}
func (self *DynamicBase) RemoveEdge(id int32) {
	if self.IsEdge(id) {
		self.is_edge[id] = false
	}
}

//*******************************************
// convert from/to graph
//*******************************************

// func (self *DynamicBase) ConvertToGraph() *Graph {
// 	new_nodes := NewList[Node](100)
// 	new_node_geoms := NewList[geo.Coord](100)
// 	mapping := NewArray[int32](self.NodeCount())
// 	id := int32(0)
// 	for i := 0; i < self.NodeCount(); i++ {
// 		if !self.is_node[i] {
// 			mapping[i] = -1
// 			continue
// 		}
// 		new_nodes.Add(self.GetNode(int32(i)))
// 		new_node_geoms.Add(self.GetNodeGeom(int32(i)))
// 		mapping[i] = id
// 		id += 1
// 	}
// 	new_edges := NewList[Edge](100)
// 	new_edge_geoms := NewList[geo.CoordArray](100)
// 	for i := 0; i < self.EdgeCount(); i++ {
// 		if !self.is_edge[i] {
// 			continue
// 		}
// 		edge := self.GetEdge(int32(i))
// 		if !self.is_node[edge.NodeA] || !self.is_node[edge.NodeB] {
// 			continue
// 		}
// 		new_edges.Add(Edge{
// 			NodeA:    mapping[edge.NodeA],
// 			NodeB:    mapping[edge.NodeB],
// 			Type:     edge.Type,
// 			Length:   edge.Length,
// 			Maxspeed: edge.Maxspeed,
// 			Oneway:   edge.Oneway,
// 		})
// 		new_edge_geoms.Add(self.GetEdgeGeom(int32(i)))
// 	}

// 	base := GraphBase{
// 		nodes:      Array[Node](new_nodes),
// 		edges:      Array[Edge](new_edges),
// 		node_geoms: new_node_geoms,
// 		edge_geoms: new_edge_geoms,
// 	}
// 	base.topology = _BuildTopology(base.nodes, base.edges)
// 	weight := BuildDefaultWeighting(base)
// 	index := NewBaseGraphIndex(&base)
// 	return &Graph{
// 		base:   base,
// 		weight: weight,
// 		index:  index,
// 	}
// }

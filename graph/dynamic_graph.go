package graph

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// dynamic graph
//*******************************************

type DynGraph struct {
	// graph store
	nodes      List[Node]
	edges      List[Edge]
	node_geoms List[geo.Coord]
	edge_geoms List[geo.CoordArray]
	is_node    List[bool]
	is_edge    List[bool]

	// graph adjacency
	topology AdjacencyList

	// weighting
	edge_weights List[int32]
}

func NewDynGraph(init_cap int) *DynGraph {
	return &DynGraph{
		nodes:        NewList[Node](init_cap),
		edges:        NewList[Edge](init_cap),
		node_geoms:   NewList[geo.Coord](init_cap),
		edge_geoms:   NewList[geo.CoordArray](init_cap),
		is_node:      NewList[bool](init_cap),
		is_edge:      NewList[bool](init_cap),
		topology:     NewAdjacencyList(init_cap),
		edge_weights: NewList[int32](init_cap),
	}
}

func NewDynGraphFromGraph(graph *Graph) *DynGraph {
	is_node := NewList[bool](graph.NodeCount())
	for i := 0; i < graph.NodeCount(); i++ {
		is_node.Add(true)
	}
	is_edge := NewList[bool](graph.EdgeCount())
	for i := 0; i < graph.EdgeCount(); i++ {
		is_edge.Add(true)
	}
	topology := NewAdjacencyList(graph.NodeCount())
	accessor := graph.base.topology.GetAccessor()
	for i := 0; i < graph.NodeCount(); i++ {
		accessor.SetBaseNode(int32(i), FORWARD)
		for accessor.Next() {
			topology.AddFWDEntry(int32(i), accessor.GetOtherID(), accessor.GetEdgeID(), 0)
		}
		accessor.SetBaseNode(int32(i), BACKWARD)
		for accessor.Next() {
			topology.AddBWDEntry(accessor.GetOtherID(), int32(i), accessor.GetEdgeID(), 0)
		}
	}

	panic("TODO")
	return &DynGraph{
		nodes:      List[Node](graph.base.store.nodes),
		edges:      List[Edge](graph.base.store.edges),
		node_geoms: graph.base.store.node_geoms,
		edge_geoms: graph.base.store.edge_geoms,
		is_node:    is_node,
		is_edge:    is_edge,
		topology:   topology,
		// edge_weights: graph.weight.edge_weights,
	}
}

func (self *DynGraph) GetDefaultExplorer() IGraphExplorer {
	return &_DynGraphExplorer{
		graph:    self,
		accessor: self.topology.GetAccessor(),
	}
}
func (self *DynGraph) NodeCount() int {
	return self.nodes.Length()
}
func (self *DynGraph) EdgeCount() int {
	return self.edges.Length()
}
func (self *DynGraph) IsNode(node int32) bool {
	return self.is_node[node]
}
func (self *DynGraph) IsEdge(edge int32) bool {
	return self.is_edge[edge]
}
func (self *DynGraph) GetNode(node int32) Node {
	return self.nodes[node]
}
func (self *DynGraph) GetEdge(edge int32) Edge {
	return self.edges[edge]
}
func (self *DynGraph) GetNodeGeom(node int32) geo.Coord {
	return self.node_geoms[node]
}
func (self *DynGraph) GetEdgeGeom(edge int32) geo.CoordArray {
	return self.edge_geoms[edge]
}
func (self *DynGraph) GetIndex() IGraphIndex {
	return &DynGraphIndex{
		nodes: self.node_geoms,
	}
}

func (self *DynGraph) AddNode(node Node, point geo.Coord) int32 {
	id := int32(self.NodeCount())
	self.nodes.Add(node)
	self.is_node.Add(true)
	self.node_geoms.Add(point)
	if self.topology.node_entries.Length() < self.NodeCount() {
		self.topology.AddNodeEntry()
	}
	return id
}
func (self *DynGraph) AddEdge(edge Edge, points geo.CoordArray) int32 {
	id := int32(self.EdgeCount())
	self.edges.Add(edge)
	self.is_edge.Add(true)
	self.edge_geoms.Add(points)
	self.topology.AddFWDEntry(edge.NodeA, edge.NodeB, id, 0)
	self.topology.AddBWDEntry(edge.NodeA, edge.NodeB, id, 0)
	self.edge_weights.Add(int32(edge.Length / float32(edge.Maxspeed)))
	return id
}
func (self *DynGraph) RemoveNode(id int32) {
	if self.IsNode(id) {
		self.is_node[id] = false
	}
}
func (self *DynGraph) RemoveEdge(id int32) {
	if self.IsEdge(id) {
		self.is_edge[id] = false
	}
}

//*******************************************
// dyn graph explorer
//*******************************************

type _DynGraphExplorer struct {
	graph    *DynGraph
	accessor AdjListAccessor
}

func (self *_DynGraphExplorer) ForAdjacentEdges(node int32, direction Direction, typ Adjacency, callback func(EdgeRef)) {
	if !self.graph.IsNode(node) {
		panic("invalid node")
	}
	if typ == ADJACENT_ALL || typ == ADJACENT_EDGES {
		accessor := &self.accessor
		accessor.SetBaseNode(node, direction)
		for self.accessor.Next() {
			edge_id := self.accessor.GetEdgeID()
			if !self.graph.IsEdge(edge_id) {
				continue
			}
			other_id := self.accessor.GetOtherID()
			if !self.graph.IsNode(other_id) {
				continue
			}
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				_Type:   0,
			})
		}
	} else {
		panic("Adjacency-type not implemented for this graph.")
	}
}
func (self *_DynGraphExplorer) GetEdgeWeight(edge EdgeRef) int32 {
	return self.graph.edge_weights[edge.EdgeID]
}
func (self *_DynGraphExplorer) GetTurnCost(from EdgeRef, via int32, to EdgeRef) int32 {
	return 0
}
func (self *_DynGraphExplorer) GetOtherNode(edge EdgeRef, node int32) int32 {
	e := self.graph.GetEdge(edge.EdgeID)
	if node == e.NodeA {
		return e.NodeB
	}
	if node == e.NodeB {
		return e.NodeA
	}
	return -1
}

//*******************************************
// others
//*******************************************

type DynGraphIndex struct {
	nodes List[geo.Coord]
}

func (self *DynGraphIndex) GetClosestNode(point geo.Coord) (int32, bool) {
	distance := -1.0
	closest_id := int32(0)
	geom := self.nodes
	for id, coord := range geom {
		newdistance := geo.EuclideanDistance(point, coord)
		if distance == -1 {
			distance = newdistance
			closest_id = int32(id)
		} else if newdistance < distance {
			distance = newdistance
			closest_id = int32(id)
		}
	}
	return closest_id, true
}

//*******************************************
// convert from/to graph
//*******************************************

func (self *DynGraph) ConvertToGraph() *Graph {
	new_nodes := NewList[Node](100)
	new_node_geoms := NewList[geo.Coord](100)
	mapping := NewArray[int32](self.NodeCount())
	id := int32(0)
	for i := 0; i < self.NodeCount(); i++ {
		if !self.is_node[i] {
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
		if !self.is_edge[i] {
			continue
		}
		edge := self.GetEdge(int32(i))
		if !self.is_node[edge.NodeA] || !self.is_node[edge.NodeB] {
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

	new_store := GraphStore{
		nodes:      Array[Node](new_nodes),
		edges:      Array[Edge](new_edges),
		node_geoms: new_node_geoms,
		edge_geoms: new_edge_geoms,
	}
	base := GraphBase{
		store:    new_store,
		topology: _BuildTopology(new_store),
		index:    _BuildKDTreeIndex(new_store),
	}
	weight := BuildDefaultWeighting(base)
	return &Graph{
		base:   base,
		weight: weight,
	}
}

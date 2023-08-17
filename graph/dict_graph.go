package graph

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

// Graph implementation using dictionaries.
// Mainly for testing purposes.
type DictGraph struct {
	nodes        Dict[int32, Node]
	edges        Dict[int32, Edge]
	fwd_edgerefs Dict[int32, List[EdgeRef]]
	bwd_edgerefs Dict[int32, List[EdgeRef]]
	geom         DictGeometry
	weight       DictWeighting

	max_node_id int32
	max_edge_id int32
}

func NewDictGraph() *DictGraph {
	return &DictGraph{
		nodes:        NewDict[int32, Node](10),
		edges:        NewDict[int32, Edge](10),
		fwd_edgerefs: NewDict[int32, List[EdgeRef]](10),
		bwd_edgerefs: NewDict[int32, List[EdgeRef]](10),
		geom:         DictGeometry{nodes: NewDict[int32, geo.Coord](10), edges: NewDict[int32, geo.CoordArray](10)},
		weight:       DictWeighting{weights: NewDict[int32, int32](10)},

		max_node_id: 0,
		max_edge_id: 0,
	}
}

func (self *DictGraph) GetGeometry() IGeometry {
	return &self.geom
}
func (self *DictGraph) GetWeighting() IWeighting {
	return &self.weight
}
func (self *DictGraph) GetDefaultExplorer() IGraphExplorer {
	return &DictGraphExplorer{
		graph:  self,
		weight: &self.weight,
	}
}
func (self *DictGraph) GetGraphExplorer(weighting IWeighting) IGraphExplorer {
	return &DictGraphExplorer{
		graph:  self,
		weight: weighting,
	}
}
func (self *DictGraph) NodeCount() int32 {
	return self.max_node_id
}
func (self *DictGraph) EdgeCount() int32 {
	return self.max_edge_id
}
func (self *DictGraph) IsNode(node int32) bool {
	return self.nodes.ContainsKey(node)
}
func (self *DictGraph) GetNode(node int32) Node {
	return self.nodes[node]
}
func (self *DictGraph) GetEdge(edge int32) Edge {
	return self.edges[edge]
}
func (self *DictGraph) GetIndex() IGraphIndex {
	return &DictGraphIndex{
		nodes: self.geom.nodes,
	}
}

func (self *DictGraph) AddNode(id int32, node Node, point geo.Coord) {
	if self.nodes.ContainsKey(id) {
		panic("node already exists")
	}
	if id >= self.max_node_id {
		self.max_node_id = id + 1
	}
	self.nodes[id] = node
	self.geom.nodes[id] = point
	self.fwd_edgerefs[id] = NewList[EdgeRef](2)
	self.bwd_edgerefs[id] = NewList[EdgeRef](2)
}
func (self *DictGraph) AddEdge(edge Edge, points geo.CoordArray) {
	if !self.nodes.ContainsKey(edge.NodeA) {
		self.AddNode(edge.NodeA, Node{}, points[0])
	}
	if !self.nodes.ContainsKey(edge.NodeB) {
		self.AddNode(edge.NodeB, Node{}, points[len(points)-1])
	}
	id := self.max_edge_id
	self.max_edge_id = id + 1
	self.edges[id] = edge
	self.geom.edges[id] = points
	self.weight.weights[id] = int32(edge.Length / float32(edge.Maxspeed))
	fwd_edge_refs := self.fwd_edgerefs[edge.NodeA]
	fwd_edge_refs.Add(EdgeRef{EdgeID: id, OtherID: edge.NodeB, _Type: 0})
	self.fwd_edgerefs[edge.NodeA] = fwd_edge_refs
	bwd_edge_refs := self.bwd_edgerefs[edge.NodeB]
	bwd_edge_refs.Add(EdgeRef{EdgeID: id, OtherID: edge.NodeA, _Type: 0})
	self.bwd_edgerefs[edge.NodeB] = bwd_edge_refs
}
func (self *DictGraph) AddDummyEdge(node_a, node_b int32, weight int32) {
	if !self.nodes.ContainsKey(node_a) {
		self.AddNode(node_a, Node{}, geo.Coord{})
	}
	if !self.nodes.ContainsKey(node_b) {
		self.AddNode(node_b, Node{}, geo.Coord{})
	}
	id := self.max_edge_id
	self.max_edge_id = id + 1
	self.edges[id] = Edge{
		NodeA: node_a,
		NodeB: node_b,
	}
	self.geom.edges[id] = geo.CoordArray{}
	self.weight.weights[id] = weight
	fwd_edge_refs := self.fwd_edgerefs[node_a]
	fwd_edge_refs.Add(EdgeRef{EdgeID: id, OtherID: node_b, _Type: 0})
	self.fwd_edgerefs[node_a] = fwd_edge_refs
	bwd_edge_refs := self.bwd_edgerefs[node_b]
	bwd_edge_refs.Add(EdgeRef{EdgeID: id, OtherID: node_a, _Type: 0})
	self.bwd_edgerefs[node_b] = bwd_edge_refs
}

type DictGraphExplorer struct {
	graph  *DictGraph
	weight IWeighting
}

func (self *DictGraphExplorer) GetAdjacentEdges(node int32, direction Direction, typ Adjacency) IIterator[EdgeRef] {
	if direction == FORWARD {
		edge_refs := self.graph.fwd_edgerefs[node]
		return NewListIterator(edge_refs)
	} else {
		edge_refs := self.graph.bwd_edgerefs[node]
		return NewListIterator(edge_refs)
	}
}
func (self *DictGraphExplorer) GetEdgeWeight(edge EdgeRef) int32 {
	return self.weight.GetEdgeWeight(edge.EdgeID)
}
func (self *DictGraphExplorer) GetTurnCost(from EdgeRef, via int32, to EdgeRef) int32 {
	return self.weight.GetTurnCost(from.EdgeID, via, to.EdgeID)
}
func (self *DictGraphExplorer) GetOtherNode(edge EdgeRef, node int32) int32 {
	e := self.graph.GetEdge(edge.EdgeID)
	if node == e.NodeA {
		return e.NodeB
	}
	if node == e.NodeB {
		return e.NodeA
	}
	return -1
}

type DictGraphIndex struct {
	nodes Dict[int32, geo.Coord]
}

func (self *DictGraphIndex) GetClosestNode(point geo.Coord) (int32, bool) {
	distance := -1.0
	closest_id := int32(0)
	geom := self.nodes
	for id, coord := range geom {
		newdistance := geo.EuclideanDistance(point, coord)
		if distance == -1 {
			distance = newdistance
			closest_id = id
		} else if newdistance < distance {
			distance = newdistance
			closest_id = id
		}
	}
	return closest_id, true
}

type DictWeighting struct {
	weights Dict[int32, int32]
}

func (self *DictWeighting) GetEdgeWeight(edge int32) int32 {
	return self.weights[edge]
}
func (self *DictWeighting) GetTurnCost(from, via, to int32) int32 {
	return 0
}

type DictGeometry struct {
	nodes Dict[int32, geo.Coord]
	edges Dict[int32, geo.CoordArray]
}

func (self *DictGeometry) GetNode(node int32) geo.Coord {
	return self.nodes[node]
}

func (self *DictGeometry) GetEdge(edge int32) geo.CoordArray {
	return self.edges[edge]
}

func (self *DictGeometry) GetAllNodes() []geo.Coord {
	nodes := make([]geo.Coord, 0, 10)
	for _, coord := range self.nodes {
		nodes = append(nodes, coord)
	}
	return nodes
}

func (self *DictGeometry) GetAllEdges() []geo.CoordArray {
	edges := make([]geo.CoordArray, 0, 10)
	for _, coords := range self.edges {
		edges = append(edges, coords)
	}
	return edges
}

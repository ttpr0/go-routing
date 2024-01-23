package graph

import (
	"errors"
	"os"

	"github.com/ttpr0/go-routing/geo"
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// transit-graph
//******************************************

type TransitGraph struct {
	base   IGraphBase
	weight IWeighting
	index  IGraphIndex

	id_mapping       _IDMapping
	transit_stops    Array[_TransitStop]
	transit_edges    Array[_TransitEdge]
	transit_topology _AdjacencyArray
	transit_weights  Array[_TransitEdgeWeight]
}

func (self *TransitGraph) GetGraphExplorer() *TransitGraphExplorer {
	return &TransitGraphExplorer{
		graph:    self,
		accessor: self.base.GetAccessor(),
		weight:   self.weight,

		transit_accessor: self.transit_topology.GetAccessor(),
	}
}
func (self *TransitGraph) NodeCount() int {
	return self.base.NodeCount()
}
func (self *TransitGraph) EdgeCount() int {
	return self.base.EdgeCount()
}
func (self *TransitGraph) IsNode(node int32) bool {
	return int32(self.base.NodeCount()) < node
}
func (self *TransitGraph) GetNode(node int32) Node {
	return self.base.GetNode(node)
}
func (self *TransitGraph) GetEdge(edge int32) Edge {
	return self.base.GetEdge(edge)
}
func (self *TransitGraph) GetNodeGeom(node int32) geo.Coord {
	return self.base.GetNodeGeom(node)
}
func (self *TransitGraph) GetEdgeGeom(edge int32) geo.CoordArray {
	return self.base.GetEdgeGeom(edge)
}
func (self *TransitGraph) GetIndex() IGraphIndex {
	return self.index
}
func (self *TransitGraph) GetBaseGraph() *Graph {
	return &Graph{
		base:   self.base,
		weight: self.weight,
	}
}

//*******************************************
// transit-graph explorer
//*******************************************

type TransitGraphExplorer struct {
	graph    *TransitGraph
	accessor _AdjArrayAccessor
	weight   IWeighting

	transit_accessor _AdjArrayAccessor
}

func (self *TransitGraphExplorer) ForAdjacentEdges(node int32, typ Adjacency, arival int32, day WeekDay, callback func(EdgeRef, int)) {
	if typ == ADJACENT_EDGES {
		self.accessor.SetBaseNode(node, FORWARD)
		for self.accessor.Next() {
			edge_id := self.accessor.GetEdgeID()
			other_id := self.accessor.GetOtherID()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				_Type:   0,
			}, -1)
		}
	} else if typ == ADJACENT_ALL {
		self.accessor.SetBaseNode(node, FORWARD)
		for self.accessor.Next() {
			edge_id := self.accessor.GetEdgeID()
			other_id := self.accessor.GetOtherID()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				_Type:   0,
			}, -1)
		}

		m_node := self.graph.id_mapping.GetTarget(node)
		if m_node != -1 {
			self.transit_accessor.SetBaseNode(m_node, FORWARD)
			for self.transit_accessor.Next() {
				edge_id := self.transit_accessor.GetEdgeID()
				other_id := self.transit_accessor.GetOtherID()
				m_other_id := self.graph.id_mapping.GetSource(other_id)

				index := -1
				curr_arival := int32(1000000000)
				weights := self.graph.transit_weights[edge_id].weights
				for i := 0; i < len(weights); i++ {
					w := weights[i]
					if w.A != day {
						continue
					}
					if w.B < arival {
						continue
					}
					if w.B > curr_arival {
						break
					}
					if w.C < curr_arival {
						curr_arival = w.C
						index = i
					}
				}
				if index == -1 {
					continue
				}

				callback(EdgeRef{
					EdgeID:  edge_id,
					OtherID: m_other_id,
					_Type:   100,
				}, index)
			}
		}
	} else {
		panic("Adjacency-type not implemented for this graph.")
	}
}
func (self *TransitGraphExplorer) GetEdgeWeight(edge EdgeRef, index int) int32 {
	if edge.IsEdge() {
		return self.weight.GetEdgeWeight(edge.EdgeID)
	} else {
		edge_weights := self.graph.transit_weights[edge.EdgeID].weights
		trip := edge_weights[index]
		w := trip.C - trip.B
		if w < 0 {
			panic("edge weight less than zero")
		}
		return w
	}
}
func (self *TransitGraphExplorer) GetTurnCost(from EdgeRef, arival int32, via int32, to EdgeRef, to_index int) int32 {
	if !from.IsEdge() && to.IsEdge() {
		return 0
	}
	if from.IsEdge() && to.IsEdge() {
		return self.weight.GetTurnCost(from.EdgeID, via, to.EdgeID)
	} else {
		to_weights := self.graph.transit_weights[to.EdgeID].weights
		w := to_weights[to_index].B - arival
		if w < 0 {
			panic("turn cost less than zero")
		}
		return w
	}
}
func (self *TransitGraphExplorer) GetOtherNode(edge EdgeRef, node int32) int32 {
	if edge.IsEdge() {
		e := self.graph.GetEdge(edge.EdgeID)
		if node == e.NodeA {
			return e.NodeB
		}
		if node == e.NodeB {
			return e.NodeA
		}
		return -1
	} else {
		e := self.graph.transit_edges[edge.EdgeID]
		node_a := self.graph.id_mapping.GetSource(e.node_a)
		node_b := self.graph.id_mapping.GetSource(e.node_b)
		if node == node_a {
			return node_b
		}
		if node == node_b {
			return node_a
		}
		return -1
	}
}

//*******************************************
// transit data
//*******************************************

type _TransitData struct {
	transit_stops    Array[_TransitStop]
	transit_edges    Array[_TransitEdge]
	transit_topology _AdjacencyArray
	transit_weights  Array[_TransitEdgeWeight]
}

type _TransitStop struct {
	coord geo.Coord
}

type _TransitEdge struct {
	node_a int32
	node_b int32
}

type _TransitEdgeWeight struct {
	weights List[Triple[WeekDay, int32, int32]]
}

type WeekDay byte

const (
	MONDAY    WeekDay = 1
	THUESDAY  WeekDay = 2
	WEDNESDAY WeekDay = 3
	THURSDAY  WeekDay = 4
	FRIDAY    WeekDay = 5
	SATURDAY  WeekDay = 6
	SUNDAY    WeekDay = 7
)

//*******************************************
// build transit graph
//*******************************************

func BuildTransitGraph(base *GraphBase, weight IWeighting, data *_TransitData) *TransitGraph {
	// map transit nodes to graph nodes
	mapping := NewArray[[2]int32](base.NodeCount())
	for i := 0; i < base.NodeCount(); i++ {
		mapping[i] = [2]int32{-1, -1}
	}
	index := NewBaseGraphIndex(base)
	for i := 0; i < data.transit_stops.Length(); i++ {
		stop := data.transit_stops[i]
		closest, ok := index.index.GetClosest(stop.coord[:], 0.02)
		if !ok {
			continue
		}
		mapping[closest][0] = int32(i)
		mapping[i][1] = closest
	}

	return &TransitGraph{
		base:   base,
		weight: weight,

		id_mapping: _IDMapping{mapping: mapping},

		transit_stops:    data.transit_stops,
		transit_edges:    data.transit_edges,
		transit_weights:  data.transit_weights,
		transit_topology: data.transit_topology,
	}
}

//*******************************************
// load transit data
//*******************************************

func LoadTransitData(file string) *_TransitData {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	data, _ := os.ReadFile(file)
	reader := NewBufferReader(data)

	stop_count := Read[int32](reader)

	transit_stops := NewArray[_TransitStop](int(stop_count))
	transit_edges := NewList[_TransitEdge](10)
	transit_weights := NewList[_TransitEdgeWeight](10)
	for i := 0; i < int(stop_count); i++ {
		lon := Read[float64](reader)
		lat := Read[float64](reader)
		transit_stops[i] = _TransitStop{
			coord: geo.Coord{float32(lon), float32(lat)},
		}
		neigh_count := Read[int32](reader)
		for j := 0; j < int(neigh_count); j++ {
			neigh_id := Read[int32](reader)
			trip_count := Read[int32](reader)
			trips := NewList[Triple[WeekDay, int32, int32]](10)
			for k := 0; k < int(trip_count); k++ {
				day := Read[int32](reader)
				arive := Read[int32](reader)
				depart := Read[int32](reader)
				trips.Add(MakeTriple(WeekDay(day), arive, depart))
			}
			transit_edges.Add(_TransitEdge{
				node_a: int32(i),
				node_b: neigh_id,
			})
			transit_weights.Add(_TransitEdgeWeight{
				weights: trips,
			})
		}
	}

	dyn := _NewAdjacencyList(transit_stops.Length())
	for id, edge := range transit_edges {
		dyn.AddFWDEntry(edge.node_a, edge.node_b, int32(id), 130)
	}
	transit_topology := _AdjacencyListToArray(&dyn)

	return &_TransitData{
		transit_stops:    transit_stops,
		transit_edges:    Array[_TransitEdge](transit_edges),
		transit_weights:  Array[_TransitEdgeWeight](transit_weights),
		transit_topology: *transit_topology,
	}
}

package graph

import (
	"github.com/ttpr0/go-routing/comps"
	"github.com/ttpr0/go-routing/geo"
	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// transit-graph
//******************************************

type TransitGraph struct {
	base   comps.IGraphBase
	index  Optional[comps.IGraphIndex]
	weight comps.ITCWeighting

	transit        *comps.Transit
	transit_weight *comps.TransitWeighting
}

func (self *TransitGraph) GetGraphExplorer() IGraphExplorer {
	return &TCGraphExplorer{
		graph:    self,
		accessor: self.base.GetAccessor(),
		weight:   self.weight,
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
func (self *TransitGraph) GetNode(node int32) structs.Node {
	return self.base.GetNode(node)
}
func (self *TransitGraph) GetEdge(edge int32) structs.Edge {
	return self.base.GetEdge(edge)
}
func (self *TransitGraph) GetNodeGeom(node int32) geo.Coord {
	return self.base.GetNode(node).Loc
}
func (self *TransitGraph) GetClosestNode(point geo.Coord) (int32, bool) {
	if self.index.HasValue() {
		return self.index.Value.GetClosestNode(point)
	} else {
		self.index.Value = comps.NewGraphIndex(self.base)
		return self.index.Value.GetClosestNode(point)
	}
}

func (self *TransitGraph) StopCount() int {
	return self.transit.StopCount()
}
func (self *TransitGraph) GetStop(stop int32) structs.Node {
	return self.transit.GetStop(stop)
}
func (self *TransitGraph) IsStop(node int32) bool {
	return self.transit.MapNodeToStop(node) != -1
}
func (self *TransitGraph) MapStopToNode(stop int32) int32 {
	return self.transit.MapStopToNode(stop)
}
func (self *TransitGraph) MapNodeToStop(node int32) int32 {
	return self.transit.MapNodeToStop(node)
}
func (self *TransitGraph) ConnectionCount() int {
	return self.transit.ConnectionCount()
}
func (self *TransitGraph) GetConnection(connection int32) structs.Connection {
	return self.transit.GetConnection(connection)
}
func (self *TransitGraph) GetTransitExplorer() *TransitGraphExplorer {
	return &TransitGraphExplorer{
		graph:            self,
		transit_accessor: self.transit.GetAccessor(),
		transit_weight:   self.transit_weight,
	}
}

//*******************************************
// transit-graph explorer
//*******************************************

type TransitGraphExplorer struct {
	graph            *TransitGraph
	transit_accessor structs.IAdjAccessor
	transit_weight   comps.ITransitWeighting
}

func (self *TransitGraphExplorer) ForAdjacentEdges(stop int32, dir Direction, typ Adjacency, callback func(EdgeRef)) {
	if typ == ADJACENT_ALL {
		self.transit_accessor.SetBaseNode(stop, dir == FORWARD)
		for self.transit_accessor.Next() {
			edge_id := self.transit_accessor.GetEdgeID()
			other_id := self.transit_accessor.GetOtherID()
			type_ := self.transit_accessor.GetType()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				Type:    type_,
			})
		}
	} else {
		panic("Adjacency-type not implemented for this graph.")
	}
}
func (self *TransitGraphExplorer) GetShortcutWeight(conn EdgeRef) int32 {
	if !conn.IsShortcut() {
		return 0
	} else {
		shc := self.graph.transit.GetShortcut(conn.EdgeID)
		return shc.Weight
	}
}
func (self *TransitGraphExplorer) GetConnectionWeight(conn EdgeRef, from int32) Optional[comps.ConnectionWeight] {
	if int(conn.EdgeID) >= self.graph.ConnectionCount() || conn.IsShortcut() {
		return None[comps.ConnectionWeight]()
	}
	return self.transit_weight.GetNextWeight(conn.EdgeID, from)
}
func (self *TransitGraphExplorer) GetConnectionWeights(conn EdgeRef, from int32, to int32) []comps.ConnectionWeight {
	if int(conn.EdgeID) >= self.graph.ConnectionCount() || conn.IsShortcut() {
		return nil
	}
	return self.transit_weight.GetWeightsInRange(conn.EdgeID, from, to)
}
func (self *TransitGraphExplorer) GetOtherStop(conn EdgeRef, node int32) int32 {
	if conn.IsShortcut() {
		e := self.graph.transit.GetShortcut(conn.EdgeID)
		if node == e.From {
			return e.To
		}
		if node == e.To {
			return e.From
		}
		return -1
	} else {
		e := self.graph.transit.GetConnection(conn.EdgeID)
		if node == e.StopA {
			return e.StopB
		}
		if node == e.StopB {
			return e.StopA
		}
		return -1
	}
}

package graph

import (
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// tiled-graph with up topology for rphast
//******************************************

type TiledGraph4 struct {
	TiledGraph3

	// Backward-Up Topology
	up_topology TypedTopologyStore
}

func (self *TiledGraph4) GetDefaultExplorer() IGraphExplorer {
	return &TiledGraph4Explorer{
		graph:         &self.TiledGraph3.TiledGraph,
		accessor:      self.topology.GetAccessor(),
		skip_accessor: self.skip_topology.GetAccessor(),
		up_accessor:   self.up_topology.GetAccessor(),
		weight:        &self.weight,
		skip_weight:   &DefaultWeighting{edge_weights: self.skip_store.skip_weights},
	}
}

type TiledGraph4Explorer struct {
	graph         *TiledGraph
	accessor      TopologyAccessor
	skip_accessor TypedTopologyAccessor
	up_accessor   TypedTopologyAccessor
	weight        IWeighting
	skip_weight   IWeighting
}

func (self *TiledGraph4Explorer) GetAdjacentEdges(node int32, direction Direction, typ Adjacency) IIterator[EdgeRef] {
	if typ == ADJACENT_SKIP {
		self.skip_accessor.SetBaseNode(node, direction)
		return &TypedEdgeRefIterator{
			accessor: &self.skip_accessor,
		}
	} else if typ == ADJACENT_ALL || typ == ADJACENT_EDGES {
		self.accessor.SetBaseNode(node, direction)
		return &TiledEdgeRefIterator{
			accessor:   &self.accessor,
			edge_types: self.graph.skip_store.edge_types,
		}
	} else if typ == ADJACENT_UPWARDS {
		self.up_accessor.SetBaseNode(node, direction)
		return &TypedEdgeRefIterator{
			accessor: &self.up_accessor,
		}
	} else {
		panic("Adjacency-type not implemented for this graph.")
	}
}
func (self *TiledGraph4Explorer) GetEdgeWeight(edge EdgeRef) int32 {
	if edge.IsShortcut() {
		return self.skip_weight.GetEdgeWeight(edge.EdgeID)
	} else {
		return self.weight.GetEdgeWeight(edge.EdgeID)
	}
}
func (self *TiledGraph4Explorer) GetTurnCost(from EdgeRef, via int32, to EdgeRef) int32 {
	return self.weight.GetTurnCost(from.EdgeID, via, to.EdgeID)
}
func (self *TiledGraph4Explorer) GetOtherNode(edge EdgeRef, node int32) int32 {
	if edge.IsShortcut() {
		e := self.graph.GetShortcut(edge.EdgeID)
		if node == e.NodeA {
			return e.NodeB
		}
		if node == e.NodeB {
			return e.NodeA
		}
		return -1
	} else {
		e := self.graph.GetEdge(edge.EdgeID)
		if node == e.NodeA {
			return e.NodeB
		}
		if node == e.NodeB {
			return e.NodeA
		}
		return -1
	}
}

func TransformToTiled5(graph *TiledGraph3) *TiledGraph4 {
	dyn := NewDynamicTopology(graph.NodeCount())

	for _, edge := range graph.index_edges {
		dyn.AddBWDEntry(edge.From, edge.To, 0, 0)
	}

	return &TiledGraph4{
		TiledGraph3: *graph,

		up_topology: *DynamicToTypedTopology(&dyn),
	}
}

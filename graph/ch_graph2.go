package graph

func CreateCHGraph2(g *CHGraph) *CHGraph2 {
	up_top := NewDynamicTopology(g.NodeCount())
	down_top := NewDynamicTopology(g.NodeCount())

	accessor := g.topology.GetAccessor()
	ch_accessor := g.ch_topology.GetAccessor()
	for i := 0; i < g.NodeCount(); i++ {
		accessor.SetBaseNode(int32(i), FORWARD)
		for {
			ok := accessor.Next()
			if !ok {
				break
			}
			edge_id := accessor.GetEdgeID()
			edge := g.GetEdge(edge_id)
			if g.GetNodeLevel(edge.NodeA) > g.GetNodeLevel(edge.NodeB) {
				down_top.AddFWDEntry(edge.NodeA, edge.NodeB, edge_id, 0)
				up_top.AddBWDEntry(edge.NodeA, edge.NodeB, edge_id, 0)
			} else if g.GetNodeLevel(edge.NodeA) < g.GetNodeLevel(edge.NodeB) {
				up_top.AddFWDEntry(edge.NodeA, edge.NodeB, edge_id, 0)
				down_top.AddBWDEntry(edge.NodeA, edge.NodeB, edge_id, 0)
			}
		}
		ch_accessor.SetBaseNode(int32(i), FORWARD)
		for {
			ok := ch_accessor.Next()
			if !ok {
				break
			}
			shc_id := ch_accessor.GetEdgeID()
			shc := g.GetShortcut(shc_id)
			if g.GetNodeLevel(shc.NodeA) > g.GetNodeLevel(shc.NodeB) {
				down_top.AddFWDEntry(shc.NodeA, shc.NodeB, shc_id, 100)
				up_top.AddBWDEntry(shc.NodeA, shc.NodeB, shc_id, 100)
			} else if g.GetNodeLevel(shc.NodeA) < g.GetNodeLevel(shc.NodeB) {
				up_top.AddFWDEntry(shc.NodeA, shc.NodeB, shc_id, 100)
				down_top.AddBWDEntry(shc.NodeA, shc.NodeB, shc_id, 100)
			}
		}
	}

	return &CHGraph2{
		CHGraph: *g,

		up_topology:   *DynamicToTypedTopology(&up_top),
		down_topology: *DynamicToTypedTopology(&down_top),
	}
}

type CHGraph2 struct {
	CHGraph

	// additional topologies
	up_topology   TypedTopologyStore
	down_topology TypedTopologyStore
}

func (self *CHGraph2) GetDefaultExplorer() IGraphExplorer {
	return &CHGraph2Explorer{
		graph:         self,
		accessor:      self.topology.GetAccessor(),
		sh_accessor:   self.ch_topology.GetAccessor(),
		up_accessor:   self.up_topology.GetAccessor(),
		down_accessor: self.down_topology.GetAccessor(),
		weight:        &self.weight,
		sh_weight:     &DefaultWeighting{edge_weights: self.ch_store.sh_weight},
	}
}

type CHGraph2Explorer struct {
	graph         *CHGraph2
	accessor      TopologyAccessor
	sh_accessor   TopologyAccessor
	up_accessor   TypedTopologyAccessor
	down_accessor TypedTopologyAccessor
	weight        IWeighting
	sh_weight     IWeighting
}

func (self *CHGraph2Explorer) ForAdjacentEdges(node int32, direction Direction, typ Adjacency, callback func(EdgeRef)) {
	if typ == ADJACENT_ALL {
		self.accessor.SetBaseNode(node, direction)
		self.sh_accessor.SetBaseNode(node, direction)
		for self.accessor.Next() {
			edge_id := self.accessor.GetEdgeID()
			other_id := self.accessor.GetOtherID()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				_Type:   0,
			})
		}
		for self.sh_accessor.Next() {
			edge_id := self.sh_accessor.GetEdgeID()
			other_id := self.sh_accessor.GetOtherID()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				_Type:   100,
			})
		}
	} else if typ == ADJACENT_EDGES {
		self.accessor.SetBaseNode(node, direction)
		for self.accessor.Next() {
			edge_id := self.accessor.GetEdgeID()
			other_id := self.accessor.GetOtherID()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				_Type:   0,
			})
		}
	} else if typ == ADJACENT_SHORTCUTS {
		self.sh_accessor.SetBaseNode(node, direction)
		for self.sh_accessor.Next() {
			edge_id := self.sh_accessor.GetEdgeID()
			other_id := self.sh_accessor.GetOtherID()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				_Type:   100,
			})
		}
	} else if typ == ADJACENT_UPWARDS {
		self.up_accessor.SetBaseNode(node, direction)
		for self.up_accessor.Next() {
			edge_id := self.up_accessor.GetEdgeID()
			other_id := self.up_accessor.GetOtherID()
			typ := self.up_accessor.GetType()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				_Type:   typ,
			})
		}
	} else if typ == ADJACENT_DOWNWARDS {
		self.down_accessor.SetBaseNode(node, direction)
		for self.down_accessor.Next() {
			edge_id := self.down_accessor.GetEdgeID()
			other_id := self.down_accessor.GetOtherID()
			typ := self.down_accessor.GetType()
			callback(EdgeRef{
				EdgeID:  edge_id,
				OtherID: other_id,
				_Type:   typ,
			})
		}
	} else {
		panic("Adjacency-type not implemented for this graph.")
	}
}
func (self *CHGraph2Explorer) GetEdgeWeight(edge EdgeRef) int32 {
	if edge.IsCHShortcut() {
		return self.sh_weight.GetEdgeWeight(edge.EdgeID)
	} else {
		return self.weight.GetEdgeWeight(edge.EdgeID)
	}
}
func (self *CHGraph2Explorer) GetTurnCost(from EdgeRef, via int32, to EdgeRef) int32 {
	return 0
}
func (self *CHGraph2Explorer) GetOtherNode(edge EdgeRef, node int32) int32 {
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

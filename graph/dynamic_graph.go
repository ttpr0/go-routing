package graph

import (
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// dynamic graph structs
//*******************************************

type DynamicNodeRef struct {
	FWDEdgeRefs List[EdgeRef]
	BWDEdgeRefs List[EdgeRef]
}

//*******************************************
// dynamic graph
//*******************************************

type DynamicGraph struct {
	node_refs   List[DynamicNodeRef]
	nodes       List[Node]
	node_levels List[int16]
	edges       List[Edge]
	shortcuts   List[Shortcut]
	geom        IGeometry
	weight      List[int32]
	sh_weight   List[int32]
	index       KDTree[int32]
}

func (self *DynamicGraph) GetGeometry() IGeometry {
	return self.geom
}
func (self *DynamicGraph) GetOtherNode(edge, node int32, is_shortcut bool) (int32, Direction) {
	if is_shortcut {
		e := self.shortcuts[edge]
		if node == e.NodeA {
			return e.NodeB, FORWARD
		}
		if node == e.NodeB {
			return e.NodeA, BACKWARD
		}
		return -1, 0
	} else {
		e := self.edges[edge]
		if node == e.NodeA {
			return e.NodeB, FORWARD
		}
		if node == e.NodeB {
			return e.NodeA, BACKWARD
		}
		return -1, 0
	}
}
func (self *DynamicGraph) GetAdjacentEdges(node int32, direction Direction) IIterator[EdgeRef] {
	n := self.node_refs[node]
	if direction == FORWARD {
		return &EdgeRefIterator{
			state:     0,
			end:       len(n.FWDEdgeRefs),
			edge_refs: &n.FWDEdgeRefs,
		}
	} else {
		return &EdgeRefIterator{
			state:     0,
			end:       len(n.BWDEdgeRefs),
			edge_refs: &n.BWDEdgeRefs,
		}
	}
}
func (self *DynamicGraph) NodeCount() int {
	return len(self.nodes)
}
func (self *DynamicGraph) EdgeCount() int {
	return len(self.edges)
}
func (self *DynamicGraph) IsNode(node int32) bool {
	if node < int32(len(self.nodes)) {
		return true
	} else {
		return false
	}
}
func (self *DynamicGraph) GetNode(node int32) Node {
	return self.nodes[node]
}
func (self *DynamicGraph) GetEdge(edge int32) Edge {
	return self.edges[edge]
}
func (self *DynamicGraph) GetShortcut(id int32) Shortcut {
	return self.shortcuts[id]
}
func (self *DynamicGraph) GetNodeIndex() KDTree[int32] {
	return self.index
}

func (self *DynamicGraph) GetWeight(id int32, is_shortcut bool) int32 {
	if is_shortcut {
		return self.sh_weight[id]
	} else {
		return self.weight[id]
	}
}
func (self *DynamicGraph) SetWeight(id int32, is_shortcut bool, weight int32) {
	if is_shortcut {
		self.sh_weight[id] = weight
	} else {
		self.weight[id] = weight
	}
}

func (self *DynamicGraph) GetNeigbours(id int32, min_level int16) ([]int32, []int32) {
	in_neigbours := NewList[int32](4)
	out_neigbours := NewList[int32](4)
	node := self.node_refs[id]
	for _, ref := range node.FWDEdgeRefs {
		other_id := ref.NodeID
		if other_id == id || Contains(out_neigbours, other_id) {
			continue
		}
		if self.node_levels[other_id] < min_level {
			continue
		}
		out_neigbours.Add(other_id)
	}
	for _, ref := range node.BWDEdgeRefs {
		other_id := ref.NodeID
		if other_id == id || Contains(in_neigbours, other_id) {
			continue
		}
		if self.node_levels[other_id] < min_level {
			continue
		}
		in_neigbours.Add(other_id)
	}
	return in_neigbours, out_neigbours
}
func (self *DynamicGraph) GetNodeLevel(id int32) int16 {
	return self.node_levels[id]
}
func (self *DynamicGraph) SetNodeLevel(id int32, level int16) {
	self.node_levels[id] = level
}
func (self *DynamicGraph) AddShortcut(node_a, node_b int32, edges [2]Tuple[int32, byte]) {
	if node_a == node_b {
		return
	}

	weight := int32(0)
	weight += self.GetWeight(edges[0].A, edges[1].B == 2 || edges[1].B == 3)
	weight += self.GetWeight(edges[1].A, edges[1].B == 2 || edges[1].B == 3)
	shortcut := Shortcut{
		NodeA: node_a,
		NodeB: node_b,
		Edges: edges,
	}
	shc_id := self.shortcuts.Length()
	self.sh_weight.Add(weight)
	self.shortcuts.Add(shortcut)

	node := self.node_refs[node_a]
	node.FWDEdgeRefs.Add(EdgeRef{
		EdgeID: int32(shc_id),
		Type:   2,
		NodeID: node_b,
		Weight: weight,
	})
	self.node_refs[node_a] = node
	node = self.node_refs[node_b]
	node.BWDEdgeRefs.Add(EdgeRef{
		EdgeID: int32(shc_id),
		Type:   3,
		NodeID: node_b,
		Weight: weight,
	})
	self.node_refs[node_b] = node
}
func (self *DynamicGraph) GetWeightBetween(from, to int32) int32 {
	for _, ref := range self.node_refs[from].FWDEdgeRefs {
		if ref.NodeID == to {
			return ref.Weight
		}
	}
	return -1
}

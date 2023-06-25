package graph

import (
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type ICHGraph interface {
	GetGeometry() IGeometry
	GetWeighting() IWeighting
	GetShortcutWeighting() IWeighting
	GetOtherNode(edge, node int32) (int32, Direction)
	GetOtherShortcutNode(shortcut, node int32) (int32, Direction)
	GetNodeLevel(node int32) int16
	GetAdjacentEdges(node int32, direction Direction) IIterator[EdgeRef]
	ForEachEdge(node int32, f func(int32))
	NodeCount() int32
	EdgeCount() int32
	ShortcutCount() int32
	IsNode(node int32) bool
	GetNode(node int32) Node
	GetEdge(edge int32) Edge
	GetShortcut(shortcut int32) CHShortcut
	GetEdgesFromShortcut(edges *List[int32], shortcut_id int32, reversed bool)
	GetNodeIndex() KDTree[int32]
	GetClosestNode(point geo.Coord) (int32, bool)
}

type CHGraph struct {
	node_refs     List[NodeRef]
	nodes         List[Node]
	node_levels   List[int16]
	fwd_edge_refs List[EdgeRef]
	bwd_edge_refs List[EdgeRef]
	edges         List[Edge]
	shortcuts     List[CHShortcut]
	geom          IGeometry
	weight        IWeighting
	sh_weight     IWeighting
	index         KDTree[int32]
}

func (self *CHGraph) GetGeometry() IGeometry {
	return self.geom
}

func (self *CHGraph) GetWeighting() IWeighting {
	return self.weight
}

func (self *CHGraph) GetShortcutWeighting() IWeighting {
	return self.sh_weight
}

func (self *CHGraph) GetOtherNode(edge int32, node int32) (int32, Direction) {
	e := self.edges[edge]
	if node == e.NodeA {
		return e.NodeB, FORWARD
	}
	if node == e.NodeB {
		return e.NodeA, BACKWARD
	}
	return 0, 0
}

func (self *CHGraph) GetOtherShortcutNode(shortcut int32, node int32) (int32, Direction) {
	e := self.shortcuts[shortcut]
	if node == e.NodeA {
		return e.NodeB, FORWARD
	}
	if node == e.NodeB {
		return e.NodeA, BACKWARD
	}
	return 0, 0
}

func (self *CHGraph) GetNodeLevel(node int32) int16 {
	return self.node_levels[node]
}

func (self *CHGraph) GetAdjacentEdges(node int32, direction Direction) IIterator[EdgeRef] {
	n := self.node_refs[node]
	if direction == FORWARD {
		return &EdgeRefIterator{
			state:     int(n.EdgeRefFWDStart),
			end:       int(n.EdgeRefFWDStart) + int(n.EdgeRefFWDCount),
			edge_refs: &self.fwd_edge_refs,
		}
	} else {
		return &EdgeRefIterator{
			state:     int(n.EdgeRefBWDStart),
			end:       int(n.EdgeRefBWDStart) + int(n.EdgeRefBWDCount),
			edge_refs: &self.bwd_edge_refs,
		}
	}
}

func (self *CHGraph) ForEachEdge(node int32, f func(int32)) {
	panic("not implemented") // TODO: Implement
}

func (self *CHGraph) NodeCount() int32 {
	return int32(len(self.nodes))
}

func (self *CHGraph) EdgeCount() int32 {
	return int32(len(self.edges))
}

func (self *CHGraph) ShortcutCount() int32 {
	return int32(len(self.shortcuts))
}

func (self *CHGraph) IsNode(node int32) bool {
	if node < int32(len(self.nodes)) {
		return true
	} else {
		return false
	}
}

func (self *CHGraph) GetNode(node int32) Node {
	return self.nodes[node]
}

func (self *CHGraph) GetEdge(edge int32) Edge {
	return self.edges[edge]
}

func (self *CHGraph) GetShortcut(shortcut int32) CHShortcut {
	return self.shortcuts[shortcut]
}

func (self *CHGraph) GetEdgesFromShortcut(edges *List[int32], shortcut_id int32, reversed bool) {
	shortcut := self.GetShortcut(shortcut_id)
	if reversed {
		e := shortcut.Edges[1]
		if e.B == 2 || e.B == 3 {
			self.GetEdgesFromShortcut(edges, e.A, reversed)
		} else {
			edges.Add(e.A)
		}
		e = shortcut.Edges[0]
		if e.B == 2 || e.B == 3 {
			self.GetEdgesFromShortcut(edges, e.A, reversed)
		} else {
			edges.Add(e.A)
		}
	} else {
		e := shortcut.Edges[0]
		if e.B == 2 || e.B == 3 {
			self.GetEdgesFromShortcut(edges, e.A, reversed)
		} else {
			edges.Add(e.A)
		}
		e = shortcut.Edges[1]
		if e.B == 2 || e.B == 3 {
			self.GetEdgesFromShortcut(edges, e.A, reversed)
		} else {
			edges.Add(e.A)
		}
	}
}
func (self *CHGraph) GetNodeIndex() KDTree[int32] {
	return self.index
}
func (self *CHGraph) GetClosestNode(point geo.Coord) (int32, bool) {
	return self.index.GetClosest(point[:], 0.005)
}

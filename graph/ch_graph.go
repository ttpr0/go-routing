package graph

import (
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type ICHGraph interface {
	GetGeometry() IGeometry
	GetWeighting() IWeighting
	GetShortcutWeighting() IWeighting
	GetDefaultExplorer() IGraphExplorer
	GetGraphExplorer(weighting IWeighting) IGraphExplorer
	GetNodeLevel(node int32) int16
	NodeCount() int32
	EdgeCount() int32
	ShortcutCount() int32
	IsNode(node int32) bool
	GetNode(node int32) Node
	GetEdge(edge int32) Edge
	GetShortcut(shortcut int32) CHShortcut
	GetEdgesFromShortcut(edges *List[int32], shortcut_id int32, reversed bool)
	GetIndex() IGraphIndex
}

type CHGraph struct {
	nodes       NodeStore
	edges       EdgeStore
	shortcuts   CHShortcutStore
	topology    TopologyStore
	node_levels CHLevelStore
	geom        GeometryStore
	weight      DefaultWeighting
	sh_weight   DefaultWeighting
	index       KDTree[int32]
}

func (self *CHGraph) GetGeometry() IGeometry {
	return &self.geom
}

func (self *CHGraph) GetWeighting() IWeighting {
	return &self.weight
}

func (self *CHGraph) GetShortcutWeighting() IWeighting {
	return &self.sh_weight
}

func (self *CHGraph) GetDefaultExplorer() IGraphExplorer {
	return &CHGraphExplorer{
		graph:     self,
		weight:    &self.weight,
		sh_weight: &self.sh_weight,
	}
}

func (self *CHGraph) GetGraphExplorer(weighting IWeighting) IGraphExplorer {
	return &CHGraphExplorer{
		graph:     self,
		weight:    weighting,
		sh_weight: &self.sh_weight,
	}
}

func (self *CHGraph) GetNodeLevel(node int32) int16 {
	return self.node_levels.GetNodeLevel(node)
}

func (self *CHGraph) NodeCount() int32 {
	return int32(self.nodes.NodeCount())
}

func (self *CHGraph) EdgeCount() int32 {
	return int32(self.edges.EdgeCount())
}

func (self *CHGraph) ShortcutCount() int32 {
	return int32(self.shortcuts.ShortcutCount())
}

func (self *CHGraph) IsNode(node int32) bool {
	return self.nodes.IsNode(node)
}

func (self *CHGraph) GetNode(node int32) Node {
	return self.nodes.GetNode(node)
}

func (self *CHGraph) GetEdge(edge int32) Edge {
	return self.edges.GetEdge(edge)
}

func (self *CHGraph) GetShortcut(shortcut int32) CHShortcut {
	return self.shortcuts.GetShortcut(shortcut)
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
func (self *CHGraph) GetIndex() IGraphIndex {
	return &BaseGraphIndex{
		index: self.index,
	}
}

type CHGraphExplorer struct {
	graph     *CHGraph
	weight    IWeighting
	sh_weight IWeighting
}

func (self *CHGraphExplorer) GetAdjacentEdges(node int32, direction Direction) IIterator[EdgeRef] {
	start, count := self.graph.topology.GetNodeRef(node, direction)
	edge_refs := self.graph.topology.GetEdgeRefs(direction)
	return &EdgeRefIterator{
		state:     int(start),
		end:       int(start) + int(count),
		edge_refs: edge_refs,
	}
}
func (self *CHGraphExplorer) GetEdgeWeight(edge EdgeRef) int32 {
	if edge.IsCHShortcut() {
		return self.sh_weight.GetEdgeWeight(edge.EdgeID)
	} else {
		return self.weight.GetEdgeWeight(edge.EdgeID)
	}
}
func (self *CHGraphExplorer) GetTurnCost(from EdgeRef, via int32, to EdgeRef) int32 {
	return 0
}
func (self *CHGraphExplorer) GetOtherNode(edge EdgeRef, node int32) int32 {
	if edge.IsCHShortcut() {
		e := self.graph.shortcuts.GetShortcut(edge.EdgeID)
		if node == e.NodeA {
			return e.NodeB
		}
		if node == e.NodeB {
			return e.NodeA
		}
		return 0
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

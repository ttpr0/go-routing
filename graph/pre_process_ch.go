package graph

import (
	"fmt"
	"sort"

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
// transform to/from dynamic graph
//*******************************************

func TransformToDynamicGraph(g *Graph) *DynamicGraph {
	node_refs := NewList[DynamicNodeRef](g.node_refs.Length())
	node_levels := NewList[int16](g.node_refs.Length())
	for i := 0; i < g.nodes.Length(); i++ {
		fwd_refs := NewList[EdgeRef](4)
		bwd_refs := NewList[EdgeRef](4)
		fwd_edges := g.GetAdjacentEdges(int32(i), FORWARD)
		for {
			ref, ok := fwd_edges.Next()
			if !ok {
				break
			}
			fwd_refs.Add(ref)
		}
		bwd_edges := g.GetAdjacentEdges(int32(i), BACKWARD)
		for {
			ref, ok := bwd_edges.Next()
			if !ok {
				break
			}
			bwd_refs.Add(ref)
		}

		node_refs.Add(DynamicNodeRef{
			FWDEdgeRefs: fwd_refs,
			BWDEdgeRefs: bwd_refs,
		})
		node_levels.Add(0)
	}

	dg := DynamicGraph{
		node_refs:   node_refs,
		nodes:       g.nodes,
		node_levels: node_levels,
		edges:       g.edges,
		shortcuts:   NewList[Shortcut](100),
		geom:        g.geom,
		weight:      g.weight.(*Weighting).EdgeWeight,
		sh_weight:   NewList[int32](100),
		index:       g.index,
	}

	return &dg
}

func TransformFromDynamicGraph(dg *DynamicGraph) *CHGraph {
	node_refs := NewList[NodeRef](dg.node_refs.Length())
	fwd_edge_refs := NewList[EdgeRef](dg.EdgeCount())
	bwd_edge_refs := NewList[EdgeRef](dg.EdgeCount())

	fwd_start := 0
	bwd_start := 0
	for i := 0; i < dg.nodes.Length(); i++ {
		fwd_count := 0
		bwd_count := 0

		fwd_refs := dg.node_refs[i].FWDEdgeRefs
		bwd_refs := dg.node_refs[i].BWDEdgeRefs

		for _, ref := range fwd_refs {
			fwd_edge_refs.Add(ref)
			fwd_count += 1
		}
		for _, ref := range bwd_refs {
			bwd_edge_refs.Add(ref)
			bwd_count += 1
		}

		node_refs.Add(NodeRef{
			EdgeRefFWDStart: int32(fwd_start),
			EdgeRefFWDCount: int16(fwd_count),
			EdgeRefBWDStart: int32(bwd_start),
			EdgeRefBWDCount: int16(bwd_count),
		})
		fwd_start += fwd_count
		bwd_start += bwd_count
	}

	g := CHGraph{
		node_refs:     node_refs,
		nodes:         dg.nodes,
		node_levels:   dg.node_levels,
		fwd_edge_refs: fwd_edge_refs,
		bwd_edge_refs: bwd_edge_refs,
		edges:         dg.edges,
		shortcuts:     dg.shortcuts,
		geom:          dg.geom,
		weight:        &Weighting{dg.weight},
		sh_weight:     &Weighting{dg.sh_weight},
		index:         dg.index,
	}

	return &g
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
	weight += self.GetWeight(edges[0].A, edges[0].B == 2 || edges[0].B == 3)
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
		NodeID: node_a,
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

//*******************************************
// preprocess ch
//*******************************************

func CalcContraction(graph *DynamicGraph) {
	fmt.Println("started contracting graph")
	// initialize graph
	//graph.resetContraction();
	for i := 0; i < graph.NodeCount(); i++ {
		graph.SetNodeLevel(int32(i), 0)
	}

	level := int16(0)
	nodes := NewList[int32](graph.NodeCount())
	for {
		// get all nodes on level
		for i := 0; i < graph.NodeCount(); i++ {
			if graph.GetNodeLevel(int32(i)) >= level {
				nodes.Add(int32(i))
			}
		}
		if nodes.Length() == 0 {
			break
		}

		// sort nodes by number of adjacent edges
		fmt.Println("start ordering nodes")
		sort.Slice(nodes, func(i, j int) bool {
			a := nodes[i]
			b := nodes[j]
			ec_a := len(graph.node_refs[a].FWDEdgeRefs) + len(graph.node_refs[a].BWDEdgeRefs)
			ec_b := len(graph.node_refs[b].FWDEdgeRefs) + len(graph.node_refs[b].BWDEdgeRefs)
			return ec_a < ec_b
		})
		fmt.Println("finished ordering nodes")

		// contract nodes
		sc1 := graph.shortcuts.Length()
		nc1 := 0
		for i := 0; i < graph.NodeCount(); i++ {
			if graph.GetNodeLevel(int32(i)) == level {
				nc1 += 1
			}
		}
		count := 0
		for i := 0; i < nodes.Length(); i++ {
			node_id := nodes[i]
			if graph.GetNodeLevel(node_id) > level {
				continue
			}
			count += 1
			if count%1000 == 0 {
				fmt.Println("node :", count)
			}
			if count == 35393 {
				fmt.Println("test")
			}
			in_neigbours, out_neigbours := graph.GetNeigbours(node_id, level)
			for i := 0; i < len(in_neigbours); i++ {
				for j := 0; j < len(out_neigbours); j++ {
					from := in_neigbours[i]
					to := out_neigbours[j]
					if from == to {
						continue
					}
					add_shortcut, edges := CalcShortcut(graph, from, to, node_id, level)
					if !add_shortcut {
						continue
					}
					graph.AddShortcut(from, to, edges)
				}
			}
			for i := 0; i < len(in_neigbours); i++ {
				graph.SetNodeLevel(in_neigbours[i], int16(level+1))
			}
			for i := 0; i < len(out_neigbours); i++ {
				graph.SetNodeLevel(out_neigbours[i], int16(level+1))
			}
		}
		sc2 := graph.shortcuts.Length()
		nc2 := 0
		for i := 0; i < graph.NodeCount(); i++ {
			if graph.GetNodeLevel(int32(i)) == int16(level+1) {
				nc2 += 1
			}
		}
		fmt.Println("contracted level", level+1, ":", sc2-sc1, "shortcuts added,", nc1-nc2, "/", nc1, "nodes contracted")

		// advance level
		level += 1
		nodes.Clear()
	}
	fmt.Println("finished contracting graph")
}

func CalcShortcut(graph *DynamicGraph, start, end, contract int32, level int16) (bool, [2]Tuple[int32, byte]) {
	w1 := graph.GetWeightBetween(start, contract)
	if w1 == -1 {
		return false, [2]Tuple[int32, byte]{}
	}
	w2 := graph.GetWeightBetween(contract, end)
	if w2 == -1 {
		return false, [2]Tuple[int32, byte]{}
	}
	max_weight := w1 + w2

	heap := NewPriorityQueue[int32, int32](10)
	flags := NewDict[int32, FlagSH](10)

	flags[start] = FlagSH{pathlength: 0, visited: false, prevEdge: -1, isShortcut: false}
	heap.Enqueue(start, 0)

	var curr_id int32
	for {
		curr_id, _ := heap.Dequeue()
		curr_flag := flags[curr_id]
		if curr_id == end {
			break
		}
		if curr_flag.visited {
			continue
		}
		curr_flag.visited = true
		flags[curr_id] = curr_flag
		// curr_node := graph.GetNode(curr_id)
		edges := graph.GetAdjacentEdges(curr_id, FORWARD)
		for {
			ref, ok := edges.Next()
			if !ok {
				break
			}
			edge_id := ref.EdgeID
			other_id := ref.NodeID
			if graph.GetNodeLevel(other_id) < level {
				continue
			}
			var other_flag FlagSH
			if flags.ContainsKey(other_id) {
				other_flag = flags[other_id]
			} else {
				other_flag = FlagSH{pathlength: 10000000, visited: false, prevEdge: -1, isShortcut: false}
			}
			weight := graph.GetWeight(edge_id, ref.IsShortcut())
			newlength := curr_flag.pathlength + weight
			if newlength > max_weight {
				continue
			}
			if newlength < other_flag.pathlength {
				other_flag.pathlength = newlength
				other_flag.prevEdge = edge_id
				other_flag.isShortcut = ref.IsShortcut()
				heap.Enqueue(other_id, newlength)
			}
			flags[other_id] = other_flag
		}
	}

	curr_id = end
	curr_flag := flags[curr_id]
	prev_id, _ := graph.GetOtherNode(curr_flag.prevEdge, curr_id, curr_flag.isShortcut)
	if prev_id != contract {
		return false, [2]Tuple[int32, byte]{}
	}
	prev_flag := flags[prev_id]
	prev_prev_id, _ := graph.GetOtherNode(prev_flag.prevEdge, prev_id, prev_flag.isShortcut)
	if prev_prev_id != start {
		return false, [2]Tuple[int32, byte]{}
	}
	var pt byte
	if prev_flag.isShortcut {
		pt = 2
	} else {
		pt = 0
	}
	var ct byte
	if curr_flag.isShortcut {
		ct = 2
	} else {
		ct = 0
	}
	return true, [2]Tuple[int32, byte]{
		{prev_flag.prevEdge, pt},
		{curr_flag.prevEdge, ct},
	}
}

type FlagSH struct {
	pathlength int32
	prevEdge   int32
	isShortcut bool
	visited    bool
}

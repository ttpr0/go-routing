package graph

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// dynamic graph
//*******************************************

type DynamicGraph struct {
	// added attributes to build ch
	ch_topology Array[DynamicNodeRef]
	node_levels Array[int16]
	shortcuts   List[CHShortcut]
	sh_weight   List[int32]

	// underlying base graph
	nodes    NodeStore
	edges    EdgeStore
	topology TopologyStore
	geom     GeometryStore
	weight   DefaultWeighting
	index    KDTree[int32]
}

type DynamicNodeRef struct {
	FWDEdgeRefs List[EdgeRef]
	BWDEdgeRefs List[EdgeRef]
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
		e := self.edges.GetEdge(edge)
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
	if direction == FORWARD {
		return &DynamicEdgeRefIterator{
			state:        0,
			edge_refs:    self.topology.GetAdjacentEdgeRefs(node, FORWARD),
			shotcut_refs: self.ch_topology[node].FWDEdgeRefs,
			typ:          0,
		}
	} else {
		return &DynamicEdgeRefIterator{
			state:        0,
			edge_refs:    self.topology.GetAdjacentEdgeRefs(node, BACKWARD),
			shotcut_refs: self.ch_topology[node].BWDEdgeRefs,
			typ:          0,
		}
	}
}
func (self *DynamicGraph) NodeCount() int {
	return self.nodes.NodeCount()
}
func (self *DynamicGraph) EdgeCount() int {
	return self.edges.EdgeCount()
}
func (self *DynamicGraph) GetNode(node int32) Node {
	return self.nodes.GetNode(node)
}
func (self *DynamicGraph) GetEdge(edge int32) Edge {
	return self.edges.GetEdge(edge)
}
func (self *DynamicGraph) GetShortcut(id int32) CHShortcut {
	return self.shortcuts[id]
}
func (self *DynamicGraph) GetWeight(id int32, is_shortcut bool) int32 {
	if is_shortcut {
		return self.sh_weight[id]
	} else {
		return self.weight.GetEdgeWeight(id)
	}
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
	shortcut := CHShortcut{
		NodeA:  node_a,
		NodeB:  node_b,
		_Edges: edges,
	}
	shc_id := self.shortcuts.Length()
	self.sh_weight.Add(weight)
	self.shortcuts.Add(shortcut)

	node := self.ch_topology[node_a]
	node.FWDEdgeRefs.Add(EdgeRef{
		EdgeID:  int32(shc_id),
		_Type:   100,
		OtherID: node_b,
	})
	self.ch_topology[node_a] = node
	node = self.ch_topology[node_b]
	node.BWDEdgeRefs.Add(EdgeRef{
		EdgeID:  int32(shc_id),
		_Type:   100,
		OtherID: node_a,
	})
	self.ch_topology[node_b] = node
}
func (self *DynamicGraph) GetWeightBetween(from, to int32) int32 {
	for _, ref := range self.topology.GetAdjacentEdgeRefs(from, FORWARD) {
		if ref.OtherID == to {
			return self.weight.GetEdgeWeight(ref.EdgeID)
		}
	}
	for _, ref := range self.ch_topology[from].FWDEdgeRefs {
		if ref.OtherID == to {
			return self.sh_weight[int(ref.EdgeID)]
		}
	}
	return -1
}

type DynamicEdgeRefIterator struct {
	state        int
	edge_refs    List[_EdgeEntry]
	shotcut_refs List[EdgeRef]
	typ          byte
}

func (self *DynamicEdgeRefIterator) Next() (EdgeRef, bool) {
	if self.typ == 0 {
		if self.state < len(self.edge_refs) {
			ref := self.edge_refs[self.state]
			self.state += 1
			return EdgeRef{
				EdgeID:  ref.EdgeID,
				OtherID: ref.OtherID,
				_Type:   self.typ,
			}, true
		} else {
			self.state = 0
			self.typ = 100
		}
	}
	if self.typ == 100 {
		if self.state < len(self.shotcut_refs) {
			ref := self.shotcut_refs[self.state]
			self.state += 1
			return EdgeRef{
				EdgeID:  ref.EdgeID,
				OtherID: ref.OtherID,
				_Type:   self.typ,
			}, true
		} else {
			return EdgeRef{}, false
		}
	}
	return EdgeRef{}, false
}

//*******************************************
// transform to/from dynamic graph
//*******************************************

func TransformToDynamicGraph(g *Graph) *DynamicGraph {
	ch_topology := NewArray[DynamicNodeRef](g.nodes.NodeCount())
	node_levels := NewArray[int16](g.nodes.NodeCount())

	for i := 0; i < g.nodes.NodeCount(); i++ {
		ch_topology[i] = DynamicNodeRef{
			FWDEdgeRefs: NewList[EdgeRef](4),
			BWDEdgeRefs: NewList[EdgeRef](4),
		}
		node_levels[i] = 0
	}

	dg := DynamicGraph{
		ch_topology: ch_topology,
		node_levels: node_levels,
		shortcuts:   NewList[CHShortcut](100),
		sh_weight:   NewList[int32](100),
		nodes:       g.nodes,
		edges:       g.edges,
		topology:    g.topology,
		geom:        g.geom,
		weight:      g.weight,
		index:       g.index,
	}

	return &dg
}

func TransformFromDynamicGraph(dg *DynamicGraph) *CHGraph {
	node_refs := NewList[_NodeEntry](dg.nodes.NodeCount())
	fwd_edge_refs := NewList[_EdgeEntry](dg.shortcuts.Length())
	bwd_edge_refs := NewList[_EdgeEntry](dg.sh_weight.Length())

	fwd_start := 0
	bwd_start := 0
	for i := 0; i < dg.nodes.NodeCount(); i++ {
		fwd_count := 0
		bwd_count := 0

		fwd_refs := dg.ch_topology[i].FWDEdgeRefs
		bwd_refs := dg.ch_topology[i].BWDEdgeRefs

		for _, ref := range fwd_refs {
			fwd_edge_refs.Add(_EdgeEntry{EdgeID: ref.EdgeID, OtherID: ref.OtherID, Type: ref._Type})
			fwd_count += 1
		}
		for _, ref := range bwd_refs {
			bwd_edge_refs.Add(_EdgeEntry{EdgeID: ref.EdgeID, OtherID: ref.OtherID, Type: ref._Type})
			bwd_count += 1
		}

		node_refs.Add(_NodeEntry{
			FWDEdgeStart: int32(fwd_start),
			FWDEdgeCount: int16(fwd_count),
			BWDEdgeStart: int32(bwd_start),
			BWDEdgeCount: int16(bwd_count),
		})
		fwd_start += fwd_count
		bwd_start += bwd_count
	}

	g := CHGraph{
		nodes:       dg.nodes,
		edges:       dg.edges,
		topology:    dg.topology,
		ch_topology: TopologyStore{node_entries: Array[_NodeEntry](node_refs), fwd_edge_entries: Array[_EdgeEntry](fwd_edge_refs), bwd_edge_entries: Array[_EdgeEntry](bwd_edge_refs)},
		shortcuts:   CHShortcutStore{Array[CHShortcut](dg.shortcuts)},
		node_levels: CHLevelStore{dg.node_levels},
		geom:        dg.geom,
		weight:      dg.weight,
		sh_weight:   DefaultWeighting{dg.sh_weight},
		index:       dg.index,
	}

	return &g
}

//*******************************************
// ch utility
//*******************************************

// * searches for neighbours using edges and shortcuts for a node
//
// * is-contracted is used to limit search to nodes that have not been contracted yet (bool array containing every node in graph)
//
// * returns in-neighbours and out-neughbours
func FindNeighbours(graph *DynamicGraph, id int32, is_contracted Array[bool]) ([]int32, []int32) {
	// compute out-going neighbours
	out_neigbours := NewList[int32](4)
	for _, ref := range graph.ch_topology[id].FWDEdgeRefs {
		other_id := ref.OtherID
		if other_id == id || Contains(out_neigbours, other_id) {
			continue
		}
		if is_contracted[other_id] {
			continue
		}
		out_neigbours.Add(other_id)
	}
	for _, ref := range graph.topology.GetAdjacentEdgeRefs(id, FORWARD) {
		other_id := ref.OtherID
		if other_id == id || Contains(out_neigbours, other_id) {
			continue
		}
		if is_contracted[other_id] {
			continue
		}
		out_neigbours.Add(other_id)
	}

	// compute in-going neighbours
	in_neigbours := NewList[int32](4)
	for _, ref := range graph.ch_topology[id].BWDEdgeRefs {
		other_id := ref.OtherID
		if other_id == id || Contains(in_neigbours, other_id) {
			continue
		}
		if is_contracted[other_id] {
			continue
		}
		in_neigbours.Add(other_id)
	}
	for _, ref := range graph.topology.GetAdjacentEdgeRefs(id, BACKWARD) {
		other_id := ref.OtherID
		if other_id == id || Contains(in_neigbours, other_id) {
			continue
		}
		if is_contracted[other_id] {
			continue
		}
		in_neigbours.Add(other_id)
	}
	return in_neigbours, out_neigbours
}

// computes if a shortcut has to be added for the node contract between start and end
// is_contracted contains true for every node that is already contracted (will not be used while finding shortest path)
// returns true if a shortcut is needed and the two coresponding edges
func CalcShortcut(start, end, contract int32, graph *DynamicGraph, heap PriorityQueue[int32, int32], flags Dict[int32, FlagSH], is_contracted Array[bool]) (bool, [2]Tuple[int32, byte]) {
	w1 := graph.GetWeightBetween(start, contract)
	if w1 == -1 {
		return false, [2]Tuple[int32, byte]{}
	}
	w2 := graph.GetWeightBetween(contract, end)
	if w2 == -1 {
		return false, [2]Tuple[int32, byte]{}
	}
	max_weight := w1 + w2

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
			other_id := ref.OtherID
			if is_contracted[other_id] {
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

	is_contracted := NewArray[bool](graph.NodeCount())
	heap := NewPriorityQueue[int32, int32](10)
	flags := NewDict[int32, FlagSH](10)
	level := int16(0)
	nodes := NewList[int32](graph.NodeCount())
	for {
		// get all non contracted
		for i := 0; i < graph.NodeCount(); i++ {
			if !is_contracted[i] {
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
			_, c1 := graph.topology.GetNodeRef(a, FORWARD)
			_, c2 := graph.topology.GetNodeRef(a, BACKWARD)
			c3 := len(graph.ch_topology[a].FWDEdgeRefs) + len(graph.ch_topology[a].BWDEdgeRefs)
			count_a := c1 + c2 + int16(c3)
			b := nodes[j]
			_, c1 = graph.topology.GetNodeRef(b, FORWARD)
			_, c2 = graph.topology.GetNodeRef(b, BACKWARD)
			c3 = len(graph.ch_topology[b].FWDEdgeRefs) + len(graph.ch_topology[b].BWDEdgeRefs)
			count_b := c1 + c2 + int16(c3)
			return count_a < count_b
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
			in_neigbours, out_neigbours := FindNeighbours(graph, node_id, is_contracted)
			for i := 0; i < len(in_neigbours); i++ {
				for j := 0; j < len(out_neigbours); j++ {
					from := in_neigbours[i]
					to := out_neigbours[j]
					if from == to {
						continue
					}
					heap.Clear()
					flags.Clear()
					add_shortcut, edges := CalcShortcut(from, to, node_id, graph, heap, flags, is_contracted)
					if !add_shortcut {
						continue
					}
					graph.AddShortcut(from, to, edges)
				}
			}
			is_contracted[node_id] = true
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

//*******************************************
// preprocess ch 2
//*******************************************

func CalcContraction2(graph *DynamicGraph, contraction_order Array[int32]) {
	fmt.Println("started contracting graph")
	// initialize graph
	for i := 0; i < graph.NodeCount(); i++ {
		graph.SetNodeLevel(int32(i), 0)
	}
	is_contracted := NewArray[bool](graph.NodeCount())
	heap := NewPriorityQueue[int32, int32](10)
	flags := NewDict[int32, FlagSH](10)

	count := 0
	dt_1 := int64(0)
	dt_2 := int64(0)
	for _, node_id := range contraction_order {
		count += 1
		if count%1000 == 0 {
			fmt.Println("node :", count, "/", graph.NodeCount(), "contracted in", dt_1, "ns /", dt_2, "ns")
			dt_1 = 0
			dt_2 = 0
		}

		t1 := time.Now()

		// contract nodes
		level := graph.GetNodeLevel(node_id)
		in_neigbours, out_neigbours := FindNeighbours(graph, node_id, is_contracted)
		t2 := time.Now()
		for i := 0; i < len(in_neigbours); i++ {
			for j := 0; j < len(out_neigbours); j++ {
				from := in_neigbours[i]
				to := out_neigbours[j]
				if from == to {
					continue
				}
				heap.Clear()
				flags.Clear()
				add_shortcut, edges := CalcShortcut(from, to, node_id, graph, heap, flags, is_contracted)
				if !add_shortcut {
					continue
				}
				graph.AddShortcut(from, to, edges)
			}
		}
		dt_2 += time.Since(t2).Nanoseconds()
		is_contracted[node_id] = true
		for i := 0; i < len(in_neigbours); i++ {
			nb := in_neigbours[i]
			graph.SetNodeLevel(nb, Max(level+1, graph.GetNodeLevel(nb)))
		}
		for i := 0; i < len(out_neigbours); i++ {
			nb := out_neigbours[i]
			graph.SetNodeLevel(nb, Max(level+1, graph.GetNodeLevel(nb)))
		}

		dt_1 += time.Since(t1).Nanoseconds()
	}
	fmt.Println("finished contracting graph")
}

func SimpleNodeOrdering(graph *DynamicGraph) Array[int32] {
	nodes := NewArray[int32](graph.NodeCount())
	for i := 0; i < graph.NodeCount(); i++ {
		nodes[i] = int32(i)
	}

	// sort nodes by number of adjacent edges
	fmt.Println("start ordering nodes")
	sort.Slice(nodes, func(i, j int) bool {
		a := nodes[i]
		_, c1 := graph.topology.GetNodeRef(a, FORWARD)
		_, c2 := graph.topology.GetNodeRef(a, BACKWARD)
		c3 := len(graph.ch_topology[a].FWDEdgeRefs) + len(graph.ch_topology[a].BWDEdgeRefs)
		count_a := c1 + c2 + int16(c3)
		b := nodes[j]
		_, c1 = graph.topology.GetNodeRef(b, FORWARD)
		_, c2 = graph.topology.GetNodeRef(b, BACKWARD)
		c3 = len(graph.ch_topology[b].FWDEdgeRefs) + len(graph.ch_topology[b].BWDEdgeRefs)
		count_b := c1 + c2 + int16(c3)
		return count_a < count_b
	})
	fmt.Println("finished ordering nodes")

	return nodes
}

// computes n random shortest paths and sorts nodes by number of paths they are on
func ShortestPathNodeOrdering(graph IGraph, n int) Array[int32] {
	fmt.Println("start computing random shortest paths")
	sp_counts := NewArray[int32](int(graph.NodeCount()))
	heap := NewPriorityQueue[int32, float64](100)
	flags := NewArray[flag_d](int(graph.NodeCount()))
	c := 0
	for i := 0; i < n; i++ {
		c += 1
		if c%100 == 0 {
			fmt.Println(c, "/", n)
		}
		start := rand.Int31n(graph.NodeCount())
		end := rand.Int31n(graph.NodeCount())
		MarkNodesOnPath(start, end, sp_counts, graph, heap, flags)
	}
	fmt.Println("finished shortest paths")

	nodes := NewArray[int32](int(graph.NodeCount()))
	for i := 0; i < int(graph.NodeCount()); i++ {
		nodes[i] = int32(i)
	}
	// sort nodes by number of shortest path they are on
	fmt.Println("start ordering nodes")
	sort.Slice(nodes, func(i, j int) bool {
		a := nodes[i]
		count_a := sp_counts[a]
		b := nodes[j]
		count_b := sp_counts[b]
		return count_a < count_b
	})
	fmt.Println("finished ordering nodes")

	return nodes
}

func StoreNodeOrdering(filename string, contraction_order Array[int32]) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("failed to create csv file")
		return
	}
	defer file.Close()

	var builder strings.Builder
	for i := 0; i < contraction_order.Length()-1; i++ {
		builder.WriteString(fmt.Sprint(contraction_order[i]) + ",")
	}
	builder.WriteString(fmt.Sprint(contraction_order[contraction_order.Length()-1]))
	file.Write([]byte(builder.String()))
}
func ReadNodeOrdering(filename string) Array[int32] {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("failed to open csv file")
		return nil
	}
	defer file.Close()
	stat, _ := file.Stat()
	data := make([]byte, stat.Size())
	file.Read(data)
	s := string(data)
	tokens := strings.Split(s, ",")

	ordering := NewArray[int32](len(tokens))
	for i := 0; i < ordering.Length(); i++ {
		val, _ := strconv.Atoi(tokens[i])
		ordering[i] = int32(val)
	}
	return ordering
}

type flag_d struct {
	path_length float64
	prev_edge   int32
	visited     bool
}

func MarkNodesOnPath(start, end int32, sp_counts Array[int32], graph IGraph, heap PriorityQueue[int32, float64], flags Array[flag_d]) {
	for i := 0; i < len(flags); i++ {
		flags[i] = flag_d{
			path_length: 1000000000,
			prev_edge:   -1,
			visited:     false,
		}
	}
	flags[start].path_length = 0
	heap.Clear()
	heap.Enqueue(start, 0)

	explorer := graph.GetDefaultExplorer()
	for {
		curr_id, ok := heap.Dequeue()
		if !ok {
			return
		}
		if curr_id == end {
			break
		}
		curr_flag := flags[curr_id]
		if curr_flag.visited {
			continue
		}
		curr_flag.visited = true
		edges := explorer.GetAdjacentEdges(curr_id, FORWARD, ADJACENT_ALL)
		for {
			ref, ok := edges.Next()
			if !ok {
				break
			}
			if !ref.IsEdge() {
				continue
			}
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			other_flag := flags[other_id]
			if other_flag.visited {
				continue
			}
			new_length := curr_flag.path_length + float64(explorer.GetEdgeWeight(ref))
			if other_flag.path_length > new_length {
				other_flag.prev_edge = edge_id
				other_flag.path_length = new_length
				heap.Enqueue(other_id, new_length)
			}
			flags[other_id] = other_flag
		}
		flags[curr_id] = curr_flag
	}

	curr_id := end
	var edge int32
	for {
		sp_counts[curr_id] += 1
		if curr_id == start {
			break
		}
		edge = flags[curr_id].prev_edge
		curr_id = explorer.GetOtherNode(CreateEdgeRef(edge), curr_id)
	}
}

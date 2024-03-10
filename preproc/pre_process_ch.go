package preproc

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/ttpr0/go-routing/comps"
	"github.com/ttpr0/go-routing/graph"
	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// preprocessing graph
//*******************************************

type CHPreprocGraph struct {
	// added attributes to build ch
	ch_topology structs.AdjacencyList
	node_levels Array[int16]
	shortcuts   structs.ShortcutStore

	// underlying base graph
	base   comps.IGraphBase
	weight comps.IWeighting
}

func (self *CHPreprocGraph) GetExplorer() *CHPreprocGraphExplorer {
	return &CHPreprocGraphExplorer{
		graph:       self,
		accessor:    self.base.GetAccessor(),
		sh_accessor: self.ch_topology.GetAccessor(),
	}
}
func (self *CHPreprocGraph) NodeCount() int {
	return self.base.NodeCount()
}
func (self *CHPreprocGraph) EdgeCount() int {
	return self.base.EdgeCount()
}
func (self *CHPreprocGraph) GetNode(node int32) structs.Node {
	return self.base.GetNode(node)
}
func (self *CHPreprocGraph) GetEdge(edge int32) structs.Edge {
	return self.base.GetEdge(edge)
}
func (self *CHPreprocGraph) GetShortcut(id int32) structs.Shortcut {
	return self.shortcuts.GetShortcut(id)
}
func (self *CHPreprocGraph) GetWeight(id int32, is_shortcut bool) int32 {
	if is_shortcut {
		shc := self.shortcuts.GetShortcut(id)
		return shc.Weight
	} else {
		return self.weight.GetEdgeWeight(id)
	}
}
func (self *CHPreprocGraph) GetNodeLevel(id int32) int16 {
	return self.node_levels[id]
}
func (self *CHPreprocGraph) SetNodeLevel(id int32, level int16) {
	self.node_levels[id] = level
}
func (self *CHPreprocGraph) AddShortcut(node_a, node_b int32, edges [2]Tuple[int32, byte]) {
	if node_a == node_b {
		return
	}

	weight := int32(0)
	weight += self.GetWeight(edges[0].A, edges[0].B == 2 || edges[0].B == 3)
	weight += self.GetWeight(edges[1].A, edges[1].B == 2 || edges[1].B == 3)
	shc := structs.Shortcut{
		From:   node_a,
		To:     node_b,
		Weight: weight,
	}
	shc_id, _ := self.shortcuts.AddCHShortcut(shc, edges)

	self.ch_topology.AddEdgeEntries(node_a, node_b, int32(shc_id), 100)
}

type CHPreprocGraphExplorer struct {
	graph       *CHPreprocGraph
	accessor    structs.IAdjAccessor
	sh_accessor structs.AdjListAccessor
}

func (self *CHPreprocGraphExplorer) ForAdjacentEdges(node int32, direction graph.Direction, typ graph.Adjacency, callback func(graph.EdgeRef)) {
	self.accessor.SetBaseNode(node, direction == graph.FORWARD)
	self.sh_accessor.SetBaseNode(node, direction == graph.FORWARD)
	for self.accessor.Next() {
		edge_id := self.accessor.GetEdgeID()
		other_id := self.accessor.GetOtherID()
		callback(graph.EdgeRef{
			EdgeID:  edge_id,
			OtherID: other_id,
			Type:    0,
		})
	}
	for self.sh_accessor.Next() {
		edge_id := self.sh_accessor.GetEdgeID()
		other_id := self.sh_accessor.GetOtherID()
		callback(graph.EdgeRef{
			EdgeID:  edge_id,
			OtherID: other_id,
			Type:    100,
		})
	}
}
func (self *CHPreprocGraphExplorer) GetEdgeWeight(edge graph.EdgeRef) int32 {
	return self.graph.GetWeight(edge.EdgeID, edge.IsCHShortcut())
}
func (self *CHPreprocGraphExplorer) GetTurnCost(from graph.EdgeRef, via int32, to graph.EdgeRef) int32 {
	return 0
}
func (self *CHPreprocGraphExplorer) GetOtherNode(edge_id, node int32, is_shortcut bool) int32 {
	if is_shortcut {
		e := self.graph.GetShortcut(edge_id)
		if node == e.From {
			return e.To
		}
		if node == e.To {
			return e.From
		}
		return -1
	} else {
		e := self.graph.GetEdge(edge_id)
		if node == e.NodeA {
			return e.NodeB
		}
		if node == e.NodeB {
			return e.NodeA
		}
		return -1
	}
}
func (self *CHPreprocGraphExplorer) GetWeightBetween(from, to int32) int32 {
	self.accessor.SetBaseNode(from, true)
	for self.accessor.Next() {
		edge_id := self.accessor.GetEdgeID()
		other_id := self.accessor.GetOtherID()
		if other_id == to {
			return self.graph.GetWeight(edge_id, false)
		}
	}
	self.sh_accessor.SetBaseNode(from, true)
	for self.sh_accessor.Next() {
		edge_id := self.sh_accessor.GetEdgeID()
		other_id := self.sh_accessor.GetOtherID()
		if other_id == to {
			return self.graph.GetWeight(edge_id, true)
		}
	}
	return -1
}
func (self *CHPreprocGraphExplorer) GetEdgeBetween(from, to int32) (graph.EdgeRef, bool) {
	self.accessor.SetBaseNode(from, true)
	for self.accessor.Next() {
		edge_id := self.accessor.GetEdgeID()
		other_id := self.accessor.GetOtherID()
		if other_id == to {
			return graph.EdgeRef{
				EdgeID:  edge_id,
				Type:    0,
				OtherID: to,
			}, true
		}
	}
	self.sh_accessor.SetBaseNode(from, true)
	for self.sh_accessor.Next() {
		edge_id := self.sh_accessor.GetEdgeID()
		other_id := self.sh_accessor.GetOtherID()
		if other_id == to {
			return graph.EdgeRef{
				EdgeID:  edge_id,
				Type:    100,
				OtherID: to,
			}, true
		}
	}
	return graph.EdgeRef{}, false
}

//*******************************************
// transform to/from dynamic graph
//*******************************************

func TransformToCHPreprocGraph(base comps.IGraphBase, weight comps.IWeighting) *CHPreprocGraph {
	ch_topology := structs.NewAdjacencyList(base.NodeCount())
	node_levels := NewArray[int16](base.NodeCount())

	for i := 0; i < base.NodeCount(); i++ {
		node_levels[i] = 0
	}

	dg := CHPreprocGraph{
		base:   base,
		weight: weight,

		shortcuts:   structs.NewShortcutStore(100, true),
		ch_topology: ch_topology,
		node_levels: node_levels,
	}

	return &dg
}

func TransformToCHData(dg *CHPreprocGraph) *comps.CH {
	return comps.NewCH(dg.shortcuts, *structs.AdjacencyListToArray(&dg.ch_topology), dg.node_levels)
}

//*******************************************
// ch utility
//*******************************************

// * searches for neighbours using edges and shortcuts for a node
//
// * is-contracted is used to limit search to nodes that have not been contracted yet (bool array containing every node in graph)
//
// * returns in-neighbours and out-neughbours
func _FindNeighbours(explorer *CHPreprocGraphExplorer, id int32, is_contracted Array[bool]) ([]int32, []int32) {
	// compute out-going neighbours
	out_neigbours := NewList[int32](4)
	explorer.ForAdjacentEdges(id, graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
		other_id := ref.OtherID
		if other_id == id || Contains(out_neigbours, other_id) {
			return
		}
		if is_contracted[other_id] {
			return
		}
		out_neigbours.Add(other_id)
	})

	// compute in-going neighbours
	in_neigbours := NewList[int32](4)
	explorer.ForAdjacentEdges(id, graph.BACKWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
		other_id := ref.OtherID
		if other_id == id || Contains(in_neigbours, other_id) {
			return
		}
		if is_contracted[other_id] {
			return
		}
		in_neigbours.Add(other_id)
	})

	return in_neigbours, out_neigbours
}

// Performs a local dijkstra search from start until all targets are found or hop_limit reached.
// Flags will be set in flags-array.
// is_contracted contains true for every node that is already contracted (will not be used while finding shortest path).
func _RunLocalSearch(start int32, targets List[int32], explorer *CHPreprocGraphExplorer, heap PriorityQueue[int32, int32], flags Flags[_FlagSH], is_contracted Array[bool], hop_limit int32) {
	*flags.Get(start) = _FlagSH{
		curr_length: 0,
	}
	target_count := targets.Length()
	for _, target := range targets {
		*flags.Get(target) = _FlagSH{
			curr_length: 1000000000,
			_is_target:  true,
		}
	}
	start_flag := flags.Get(start)
	start_flag.curr_length = 0
	heap.Enqueue(start, 0)

	found_count := 0
	for {
		curr_id, ok := heap.Dequeue()
		if !ok {
			break
		}
		curr_flag := flags.Get(curr_id)
		if curr_flag.visited {
			continue
		}
		curr_flag.visited = true
		if curr_flag._is_target {
			found_count += 1
		}
		if found_count >= target_count {
			break
		}
		if curr_flag.curr_hops >= hop_limit {
			continue
		}
		explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			if is_contracted[other_id] {
				return
			}
			other_flag := flags.Get(other_id)
			weight := explorer.GetEdgeWeight(ref)
			newlength := curr_flag.curr_length + weight
			if newlength < other_flag.curr_length {
				other_flag.curr_length = newlength
				other_flag.curr_hops = curr_flag.curr_hops + 1
				other_flag.prev_edge = edge_id
				other_flag.prev_node = curr_id
				other_flag.is_shortcut = ref.IsShortcut()
				heap.Enqueue(other_id, newlength)
			}
		})
	}
}

type _FlagSH struct {
	curr_length int32
	curr_hops   int32
	prev_edge   int32
	prev_node   int32
	is_shortcut bool
	visited     bool
	_is_target  bool
}

// Returns the neccessary shortcut between from and to.
// If no shortcut is needed false will be returned.
func _GetShortcut(from, to, via int32, explorer *CHPreprocGraphExplorer, flags Flags[_FlagSH]) ([2]Tuple[int32, byte], bool) {
	edges := [2]Tuple[int32, byte]{}

	to_flag := flags.Get(to)
	// is target hasnt been found by search always add shortcut
	if !to_flag.visited {
		t_edge, _ := explorer.GetEdgeBetween(via, to)
		if t_edge.IsCHShortcut() {
			edges[0] = MakeTuple(t_edge.EdgeID, byte(2))
		} else {
			edges[0] = MakeTuple(t_edge.EdgeID, byte(0))
		}
		f_edge, _ := explorer.GetEdgeBetween(from, via)
		if f_edge.IsCHShortcut() {
			edges[1] = MakeTuple(f_edge.EdgeID, byte(2))
		} else {
			edges[1] = MakeTuple(f_edge.EdgeID, byte(0))
		}
		return edges, true
	} else {
		// check if shortest path goes through node
		if to_flag.prev_node != via {
			return edges, false
		}
		node_flag := flags.Get(via)
		if node_flag.prev_node != from {
			return edges, false
		}

		// capture edges that form shortcut
		if to_flag.is_shortcut {
			edges[0] = MakeTuple(to_flag.prev_edge, byte(2))
		} else {
			edges[0] = MakeTuple(to_flag.prev_edge, byte(0))
		}
		if node_flag.is_shortcut {
			edges[1] = MakeTuple(node_flag.prev_edge, byte(2))
		} else {
			edges[1] = MakeTuple(node_flag.prev_edge, byte(0))
		}
		return edges, true
	}
}

//*******************************************
// preprocess ch
//*******************************************

func CalcContraction(base comps.IGraphBase, weight comps.IWeighting) *comps.CH {
	graph := TransformToCHPreprocGraph(base, weight)

	fmt.Println("started contracting graph")
	// initialize graph
	//graph.resetContraction();
	for i := 0; i < graph.NodeCount(); i++ {
		graph.SetNodeLevel(int32(i), 0)
	}

	is_contracted := NewArray[bool](graph.NodeCount())
	heap := NewPriorityQueue[int32, int32](10)
	flags := NewFlags[_FlagSH](int32(graph.NodeCount()), _FlagSH{curr_length: 100000000})
	level := int16(0)
	nodes := NewList[int32](graph.NodeCount())
	explorer := graph.GetExplorer()
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
			c1 := graph.base.GetNodeDegree(a, true) + graph.base.GetNodeDegree(a, false)
			c2 := graph.ch_topology.GetDegree(a, true) + graph.ch_topology.GetDegree(a, false)
			count_a := c1 + c2
			b := nodes[j]
			c1 = graph.base.GetNodeDegree(b, true) + graph.base.GetNodeDegree(b, false)
			c2 = graph.ch_topology.GetDegree(b, true) + graph.ch_topology.GetDegree(b, false)
			count_b := c1 + c2
			return count_a < count_b
		})
		fmt.Println("finished ordering nodes")

		// contract nodes
		sc1 := graph.shortcuts.ShortcutCount()
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
			in_neigbours, out_neigbours := _FindNeighbours(explorer, node_id, is_contracted)
			for i := 0; i < len(in_neigbours); i++ {
				from := in_neigbours[i]
				heap.Clear()
				flags.Reset()
				_RunLocalSearch(from, out_neigbours, explorer, heap, flags, is_contracted, 1000000)
				for j := 0; j < len(out_neigbours); j++ {
					to := out_neigbours[j]
					if from == to {
						continue
					}
					edges, shortcut_needed := _GetShortcut(from, to, node_id, explorer, flags)
					if !shortcut_needed {
						continue
					}
					// add shortcut to graph
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
		sc2 := graph.shortcuts.ShortcutCount()
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

	ch_data := TransformToCHData(graph)
	return ch_data
}

//*******************************************
// preprocess ch 2
//*******************************************

func CalcContraction2(base comps.IGraphBase, weight comps.IWeighting, contraction_order Array[int32]) *comps.CH {
	graph := TransformToCHPreprocGraph(base, weight)

	fmt.Println("started contracting graph")
	// initialize graph
	for i := 0; i < graph.NodeCount(); i++ {
		graph.SetNodeLevel(int32(i), 0)
	}
	is_contracted := NewArray[bool](graph.NodeCount())
	heap := NewPriorityQueue[int32, int32](10)
	flags := NewFlags[_FlagSH](int32(graph.NodeCount()), _FlagSH{curr_length: 100000000})
	explorer := graph.GetExplorer()

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
		in_neigbours, out_neigbours := _FindNeighbours(explorer, node_id, is_contracted)
		t2 := time.Now()
		for i := 0; i < len(in_neigbours); i++ {
			from := in_neigbours[i]
			heap.Clear()
			flags.Reset()
			_RunLocalSearch(from, out_neigbours, explorer, heap, flags, is_contracted, 1000000)
			for j := 0; j < len(out_neigbours); j++ {
				to := out_neigbours[j]
				if from == to {
					continue
				}
				edges, shortcut_needed := _GetShortcut(from, to, node_id, explorer, flags)
				if !shortcut_needed {
					continue
				}
				// add shortcut to graph
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

	ch_data := TransformToCHData(graph)
	return ch_data
}

func SimpleNodeOrdering(graph *CHPreprocGraph) Array[int32] {
	nodes := NewArray[int32](graph.NodeCount())
	for i := 0; i < graph.NodeCount(); i++ {
		nodes[i] = int32(i)
	}

	// sort nodes by number of adjacent edges
	fmt.Println("start ordering nodes")
	sort.Slice(nodes, func(i, j int) bool {
		a := nodes[i]
		c1 := graph.base.GetNodeDegree(a, true) + graph.base.GetNodeDegree(a, false)
		c2 := graph.ch_topology.GetDegree(a, true) + graph.ch_topology.GetDegree(a, false)
		count_a := c1 + c2
		b := nodes[j]
		c1 = graph.base.GetNodeDegree(b, true) + graph.base.GetNodeDegree(b, false)
		c2 = graph.ch_topology.GetDegree(b, true) + graph.ch_topology.GetDegree(b, false)
		count_b := c1 + c2
		return count_a < count_b
	})
	fmt.Println("finished ordering nodes")

	return nodes
}

// computes n random shortest paths and sorts nodes by number of paths they are on
func ShortestPathNodeOrdering(g graph.IGraph, n int) Array[int32] {
	fmt.Println("start computing random shortest paths")
	sp_counts := NewArray[int32](int(g.NodeCount()))
	heap := NewPriorityQueue[int32, float64](100)
	flags := NewArray[flag_d](int(g.NodeCount()))
	c := 0
	for i := 0; i < n; i++ {
		c += 1
		if c%100 == 0 {
			fmt.Println(c, "/", n)
		}
		start := rand.Int31n(int32(g.NodeCount()))
		end := rand.Int31n(int32(g.NodeCount()))
		MarkNodesOnPath(start, end, sp_counts, g, heap, flags)
	}
	fmt.Println("finished shortest paths")

	nodes := NewArray[int32](int(g.NodeCount()))
	for i := 0; i < int(g.NodeCount()); i++ {
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

type flag_d struct {
	path_length float64
	prev_edge   int32
	visited     bool
}

func MarkNodesOnPath(start, end int32, sp_counts Array[int32], g graph.IGraph, heap PriorityQueue[int32, float64], flags Array[flag_d]) {
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

	explorer := g.GetGraphExplorer()
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
		explorer.ForAdjacentEdges(curr_id, graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			if !ref.IsEdge() {
				return
			}
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			other_flag := flags[other_id]
			if other_flag.visited {
				return
			}
			new_length := curr_flag.path_length + float64(explorer.GetEdgeWeight(ref))
			if other_flag.path_length > new_length {
				other_flag.prev_edge = edge_id
				other_flag.path_length = new_length
				heap.Enqueue(other_id, new_length)
			}
			flags[other_id] = other_flag
		})
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
		curr_id = explorer.GetOtherNode(graph.CreateEdgeRef(edge), curr_id)
	}
}

//*******************************************
// preprocess ch 3
//*******************************************

// Computes contraction using 2*ED + CN + EC + 5*L with hop-limits.
func CalcContraction3(base comps.IGraphBase, weight comps.IWeighting) *comps.CH {
	graph := TransformToCHPreprocGraph(base, weight)

	fmt.Println("started contracting graph...")

	// initialize
	is_contracted := NewArray[bool](graph.NodeCount())
	node_levels := NewArray[int16](graph.NodeCount())
	contracted_neighbours := NewArray[int](graph.NodeCount())
	shortcut_edgecount := NewList[int8](10)

	// initialize routing components
	node_count := graph.NodeCount()
	edge_count := graph.EdgeCount()
	heap := NewPriorityQueue[int32, int32](10)
	flags := NewFlags[_FlagSH](int32(graph.NodeCount()), _FlagSH{curr_length: 100000000})
	explorer := graph.GetExplorer()
	hop_limit := int32(5)

	// compute node priorities
	fmt.Println("computing priorities...")
	node_priorities := NewArray[int](graph.NodeCount())
	for i := 0; i < graph.NodeCount(); i++ {
		node_priorities[i] = _ComputeNodePriority(int32(i), explorer, heap, flags, is_contracted, node_levels, contracted_neighbours, shortcut_edgecount, hop_limit)
	}

	// put nodes into priority queue
	contraction_order := NewPriorityQueue[Tuple[int32, int], int](graph.NodeCount())
	for i := 0; i < graph.NodeCount(); i++ {
		prio := node_priorities[i]
		contraction_order.Enqueue(MakeTuple(int32(i), prio), prio)
	}

	fmt.Println("start contracting nodes...")
	count := 0
	for {
		temp, ok := contraction_order.Dequeue()
		if !ok {
			break
		}
		node_id := temp.A
		node_prio := temp.B
		if is_contracted[node_id] || node_prio != node_priorities[node_id] {
			continue
		}
		node_count -= 1
		count += 1
		if count%1000 == 0 {
			fmt.Println("	node :", count, "/", graph.NodeCount())
		}

		// contract node
		level := node_levels[node_id]
		in_neigbours, out_neigbours := _FindNeighbours(explorer, node_id, is_contracted)
		ed := len(in_neigbours) + len(out_neigbours)
		for i := 0; i < len(in_neigbours); i++ {
			from := in_neigbours[i]
			heap.Clear()
			flags.Reset()
			_RunLocalSearch(from, out_neigbours, explorer, heap, flags, is_contracted, hop_limit)
			for j := 0; j < len(out_neigbours); j++ {
				to := out_neigbours[j]
				if from == to {
					continue
				}
				edges, shortcut_needed := _GetShortcut(from, to, node_id, explorer, flags)
				if !shortcut_needed {
					continue
				}
				// add shortcut to graph
				graph.AddShortcut(from, to, edges)
				ed -= 1

				// compute number of edges representing the shortcut (limited to 3)
				ec := int8(0)
				if edges[0].B == 0 {
					ec += 1
				} else {
					ec += shortcut_edgecount[edges[0].A]
				}
				if edges[1].B == 0 {
					ec += 1
				} else {
					ec += shortcut_edgecount[edges[1].A]
				}
				if ec > 3 {
					ec = 3
				}
				shortcut_edgecount.Add(ec)
			}
		}
		edge_count -= ed
		if edge_count*2/node_count > 5 {
			hop_limit = 10
		}
		if edge_count*2/node_count > 10 {
			hop_limit = 10000000
		}
		// set node to contracted
		is_contracted[node_id] = true

		// update neighbours
		for i := 0; i < len(in_neigbours); i++ {
			nb := in_neigbours[i]
			node_levels[nb] = Max(level+1, node_levels[nb])
			contracted_neighbours[nb] += 1
			prio := _ComputeNodePriority(nb, explorer, heap, flags, is_contracted, node_levels, contracted_neighbours, shortcut_edgecount, hop_limit)
			node_priorities[nb] = prio
			contraction_order.Enqueue(MakeTuple(nb, prio), prio)
		}
		for i := 0; i < len(out_neigbours); i++ {
			nb := out_neigbours[i]
			node_levels[nb] = Max(level+1, node_levels[nb])
			contracted_neighbours[nb] += 1
			prio := _ComputeNodePriority(nb, explorer, heap, flags, is_contracted, node_levels, contracted_neighbours, shortcut_edgecount, hop_limit)
			node_priorities[nb] = prio
			contraction_order.Enqueue(MakeTuple(nb, prio), prio)
		}
	}
	for i := 0; i < graph.NodeCount(); i++ {
		graph.SetNodeLevel(int32(i), node_levels[i])
	}
	fmt.Println("finished contracting graph")

	ch_data := TransformToCHData(graph)
	return ch_data
}

func _ComputeNodePriority(node int32, explorer *CHPreprocGraphExplorer, heap PriorityQueue[int32, int32], flags Flags[_FlagSH], is_contracted Array[bool], node_levels Array[int16], contracted_neighbours Array[int], shortcut_edgecount List[int8], hop_limit int32) int {
	in_neigbours, out_neigbours := _FindNeighbours(explorer, node, is_contracted)
	edge_diff := -(len(in_neigbours) + len(out_neigbours))
	edge_count := int8(0)
	for i := 0; i < len(in_neigbours); i++ {
		from := in_neigbours[i]
		flags.Reset()
		heap.Clear()
		_RunLocalSearch(from, out_neigbours, explorer, heap, flags, is_contracted, hop_limit)
		for j := 0; j < len(out_neigbours); j++ {
			to := out_neigbours[j]
			if from == to {
				continue
			}
			edges, shortcut_needed := _GetShortcut(from, to, node, explorer, flags)
			if !shortcut_needed {
				continue
			}
			edge_diff += 1
			// compute number of edges representing the shortcut (limited to 3)
			ec := int8(0)
			if edges[0].B == 0 {
				ec += 1
			} else {
				ec += shortcut_edgecount[edges[0].A]
			}
			if edges[1].B == 0 {
				ec += 1
			} else {
				ec += shortcut_edgecount[edges[1].A]
			}
			if ec > 3 {
				ec = 3
			}
			edge_count += ec
		}
	}

	return 2*edge_diff + contracted_neighbours[node] + int(edge_count) + 5*int(node_levels[node])
}

// Computes contraction using 2*ED + CN.
func CalcContraction4(base comps.IGraphBase, weight comps.IWeighting) *comps.CH {
	graph := TransformToCHPreprocGraph(base, weight)

	fmt.Println("started contracting graph...")

	// initialize
	is_contracted := NewArray[bool](graph.NodeCount())
	node_levels := NewArray[int16](graph.NodeCount())
	contracted_neighbours := NewArray[int](graph.NodeCount())

	// initialize routing components
	heap := NewPriorityQueue[int32, int32](10)
	flags := NewFlags[_FlagSH](int32(graph.NodeCount()), _FlagSH{curr_length: 100000000})
	explorer := graph.GetExplorer()

	// compute node priorities
	fmt.Println("computing priorities...")
	node_priorities := NewArray[int](graph.NodeCount())
	for i := 0; i < graph.NodeCount(); i++ {
		node_priorities[i] = _ComputeNodePriority2(int32(i), explorer, heap, flags, is_contracted, node_levels, contracted_neighbours)
	}

	// put nodes into priority queue
	contraction_order := NewPriorityQueue[Tuple[int32, int], int](graph.NodeCount())
	for i := 0; i < graph.NodeCount(); i++ {
		prio := node_priorities[i]
		contraction_order.Enqueue(MakeTuple(int32(i), prio), prio)
	}

	fmt.Println("start contracting nodes...")
	count := 0
	for {
		temp, ok := contraction_order.Dequeue()
		if !ok {
			break
		}
		node_id := temp.A
		node_prio := temp.B
		if is_contracted[node_id] || node_prio != node_priorities[node_id] {
			continue
		}

		count += 1
		if count%1000 == 0 {
			fmt.Println("	node :", count, "/", graph.NodeCount())
		}

		// contract node
		level := node_levels[node_id]
		in_neigbours, out_neigbours := _FindNeighbours(explorer, node_id, is_contracted)
		for i := 0; i < len(in_neigbours); i++ {
			from := in_neigbours[i]
			heap.Clear()
			flags.Reset()
			_RunLocalSearch(from, out_neigbours, explorer, heap, flags, is_contracted, 1000000)
			for j := 0; j < len(out_neigbours); j++ {
				to := out_neigbours[j]
				if from == to {
					continue
				}
				edges, shortcut_needed := _GetShortcut(from, to, node_id, explorer, flags)
				if !shortcut_needed {
					continue
				}
				// add shortcut to graph
				graph.AddShortcut(from, to, edges)
			}
		}
		// set node to contracted
		is_contracted[node_id] = true

		// update neighbours
		for i := 0; i < len(in_neigbours); i++ {
			nb := in_neigbours[i]
			node_levels[nb] = Max(level+1, node_levels[nb])
			contracted_neighbours[nb] += 1
			prio := _ComputeNodePriority2(nb, explorer, heap, flags, is_contracted, node_levels, contracted_neighbours)
			node_priorities[nb] = prio
			contraction_order.Enqueue(MakeTuple(nb, prio), prio)
		}
		for i := 0; i < len(out_neigbours); i++ {
			nb := out_neigbours[i]
			node_levels[nb] = Max(level+1, node_levels[nb])
			contracted_neighbours[nb] += 1
			prio := _ComputeNodePriority2(nb, explorer, heap, flags, is_contracted, node_levels, contracted_neighbours)
			node_priorities[nb] = prio
			contraction_order.Enqueue(MakeTuple(nb, prio), prio)
		}
	}
	for i := 0; i < graph.NodeCount(); i++ {
		graph.SetNodeLevel(int32(i), node_levels[i])
	}
	fmt.Println("finished contracting graph")

	ch_data := TransformToCHData(graph)
	return ch_data
}

func _ComputeNodePriority2(node int32, explorer *CHPreprocGraphExplorer, heap PriorityQueue[int32, int32], flags Flags[_FlagSH], is_contracted Array[bool], node_levels Array[int16], contracted_neighbours Array[int]) int {
	in_neigbours, out_neigbours := _FindNeighbours(explorer, node, is_contracted)
	edge_diff := -(len(in_neigbours) + len(out_neigbours))
	for i := 0; i < len(in_neigbours); i++ {
		from := in_neigbours[i]
		flags.Reset()
		heap.Clear()
		_RunLocalSearch(from, out_neigbours, explorer, heap, flags, is_contracted, 1000000)
		for j := 0; j < len(out_neigbours); j++ {
			to := out_neigbours[j]
			if from == to {
				continue
			}
			_, shortcut_needed := _GetShortcut(from, to, node, explorer, flags)
			if !shortcut_needed {
				continue
			}
			edge_diff += 1
		}
	}

	// return 2*edge_diff + contracted_neighbours[node] + 5*int(node_levels[node])
	return 2*edge_diff + contracted_neighbours[node]
}

//*******************************************
// preprocess ch using partitions
//*******************************************

// Computes contraction using 2*ED + CN + EC + 5*L.
// Ignores border nodes until all interior nodes are contracted.
func CalcContraction5(base comps.IGraphBase, weight comps.IWeighting, partition *comps.Partition) *comps.CH {
	graph := TransformToCHPreprocGraph(base, weight)

	fmt.Println("started contracting graph...")

	// initialize
	is_contracted := NewArray[bool](graph.NodeCount())
	node_levels := NewArray[int16](graph.NodeCount())
	contracted_neighbours := NewArray[int](graph.NodeCount())
	shortcut_edgecount := NewList[int8](10)

	// initialize routing components
	heap := NewPriorityQueue[int32, int32](10)
	flags := NewFlags[_FlagSH](int32(graph.NodeCount()), _FlagSH{curr_length: 100000000})
	explorer := graph.GetExplorer()

	// compute node priorities
	fmt.Println("computing priorities...")
	is_border := _IsBorderNode(graph, partition)
	border_count := 0
	node_priorities := NewArray[int](graph.NodeCount())
	contraction_order := NewPriorityQueue[Tuple[int32, int], int](graph.NodeCount())
	for i := 0; i < graph.NodeCount(); i++ {
		if is_border[i] {
			node_priorities[i] = 10000000000
			border_count += 1
		}
		prio := _ComputeNodePriority(int32(i), explorer, heap, flags, is_contracted, node_levels, contracted_neighbours, shortcut_edgecount, 100000)
		node_priorities[i] = prio
		contraction_order.Enqueue(MakeTuple(int32(i), prio), prio)
	}

	fmt.Println("start contracting nodes...")
	contract_count := 0
	is_border_contraction := false
	for {
		temp, ok := contraction_order.Dequeue()
		if !ok {
			break
		}
		node_id := temp.A
		node_prio := temp.B
		if is_contracted[node_id] || node_prio != node_priorities[node_id] {
			continue
		}

		contract_count += 1
		if contract_count%1000 == 0 {
			fmt.Println("	node :", contract_count, "/", graph.NodeCount())
		}

		if contract_count == graph.NodeCount()-border_count {
			is_border_contraction = true
			for i := 0; i < graph.NodeCount(); i++ {
				if !is_border[i] {
					continue
				}
				prio := _ComputeNodePriority(int32(i), explorer, heap, flags, is_contracted, node_levels, contracted_neighbours, shortcut_edgecount, 100000)
				node_priorities[i] = prio
				contraction_order.Enqueue(MakeTuple(int32(i), prio), prio)
			}
		}

		// contract node
		level := node_levels[node_id]
		in_neigbours, out_neigbours := _FindNeighbours(explorer, node_id, is_contracted)
		for i := 0; i < len(in_neigbours); i++ {
			from := in_neigbours[i]
			heap.Clear()
			flags.Reset()
			_RunLocalSearch(from, out_neigbours, explorer, heap, flags, is_contracted, 1000000)
			for j := 0; j < len(out_neigbours); j++ {
				to := out_neigbours[j]
				if from == to {
					continue
				}
				edges, shortcut_needed := _GetShortcut(from, to, node_id, explorer, flags)
				if !shortcut_needed {
					continue
				}
				// add shortcut to graph
				graph.AddShortcut(from, to, edges)

				// compute number of edges representing the shortcut (limited to 3)
				ec := int8(0)
				if edges[0].B == 0 {
					ec += 1
				} else {
					ec += shortcut_edgecount[edges[0].A]
				}
				if edges[1].B == 0 {
					ec += 1
				} else {
					ec += shortcut_edgecount[edges[1].A]
				}
				if ec > 3 {
					ec = 3
				}
				shortcut_edgecount.Add(ec)
			}
		}
		// set node to contracted
		is_contracted[node_id] = true

		// update neighbours
		for i := 0; i < len(in_neigbours); i++ {
			nb := in_neigbours[i]
			node_levels[nb] = Max(level+1, node_levels[nb])
			contracted_neighbours[nb] += 1
			if is_border[nb] && !is_border_contraction {
				continue
			}
			prio := _ComputeNodePriority(nb, explorer, heap, flags, is_contracted, node_levels, contracted_neighbours, shortcut_edgecount, 100000)
			node_priorities[nb] = prio
			contraction_order.Enqueue(MakeTuple(nb, prio), prio)
		}
		for i := 0; i < len(out_neigbours); i++ {
			nb := out_neigbours[i]
			node_levels[nb] = Max(level+1, node_levels[nb])
			contracted_neighbours[nb] += 1
			if is_border[nb] && !is_border_contraction {
				continue
			}
			prio := _ComputeNodePriority(nb, explorer, heap, flags, is_contracted, node_levels, contracted_neighbours, shortcut_edgecount, 100000)
			node_priorities[nb] = prio
			contraction_order.Enqueue(MakeTuple(nb, prio), prio)
		}
	}
	for i := 0; i < graph.NodeCount(); i++ {
		graph.SetNodeLevel(int32(i), node_levels[i])
	}
	fmt.Println("finished contracting graph")

	ch_data := TransformToCHData(graph)
	return ch_data
}

func _IsBorderNode(g *CHPreprocGraph, partition *comps.Partition) Array[bool] {
	is_border := NewArray[bool](g.NodeCount())

	explorer := g.GetExplorer()
	for i := 0; i < g.NodeCount(); i++ {
		explorer.ForAdjacentEdges(int32(i), graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			if partition.GetNodeTile(int32(i)) != partition.GetNodeTile(ref.OtherID) {
				is_border[i] = true
			}
		})
		explorer.ForAdjacentEdges(int32(i), graph.BACKWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			if partition.GetNodeTile(int32(i)) != partition.GetNodeTile(ref.OtherID) {
				is_border[i] = true
			}
		})
	}

	return is_border
}

// Computes contraction using 2*ED + CN + EC + 5*L.
// Ignores border nodes until all interior nodes are contracted.
func CalcPartialContraction5(base comps.IGraphBase, weight comps.IWeighting, partition *comps.Partition) *comps.CH {
	graph := TransformToCHPreprocGraph(base, weight)

	fmt.Println("started contracting graph...")

	// initialize
	is_contracted := NewArray[bool](graph.NodeCount())
	node_levels := NewArray[int16](graph.NodeCount())
	contracted_neighbours := NewArray[int](graph.NodeCount())
	shortcut_edgecount := NewList[int8](10)

	// initialize routing components
	heap := NewPriorityQueue[int32, int32](10)
	flags := NewFlags[_FlagSH](int32(graph.NodeCount()), _FlagSH{curr_length: 100000000})
	explorer := graph.GetExplorer()

	// compute node priorities
	fmt.Println("computing priorities...")
	is_border := _IsBorderNode(graph, partition)
	node_priorities := NewArray[int](graph.NodeCount())
	contraction_order := NewPriorityQueue[Tuple[int32, int], int](graph.NodeCount())
	for i := 0; i < graph.NodeCount(); i++ {
		if is_border[i] {
			node_priorities[i] = 10000000000
			continue
		}
		prio := _ComputeNodePriority(int32(i), explorer, heap, flags, is_contracted, node_levels, contracted_neighbours, shortcut_edgecount, 100000)
		node_priorities[i] = prio
		contraction_order.Enqueue(MakeTuple(int32(i), prio), prio)
	}

	fmt.Println("start contracting nodes...")
	contract_count := 0
	for {
		temp, ok := contraction_order.Dequeue()
		if !ok {
			break
		}
		node_id := temp.A
		node_prio := temp.B
		if node_prio == 10000000000 {
			break
		}
		if is_contracted[node_id] || node_prio != node_priorities[node_id] {
			continue
		}

		contract_count += 1
		if contract_count%1000 == 0 {
			fmt.Println("	node :", contract_count, "/", graph.NodeCount())
		}

		// contract node
		level := node_levels[node_id]
		in_neigbours, out_neigbours := _FindNeighbours(explorer, node_id, is_contracted)
		for i := 0; i < len(in_neigbours); i++ {
			from := in_neigbours[i]
			heap.Clear()
			flags.Reset()
			_RunLocalSearch(from, out_neigbours, explorer, heap, flags, is_contracted, 1000000)
			for j := 0; j < len(out_neigbours); j++ {
				to := out_neigbours[j]
				if from == to {
					continue
				}
				edges, shortcut_needed := _GetShortcut(from, to, node_id, explorer, flags)
				if !shortcut_needed {
					continue
				}
				// add shortcut to graph
				graph.AddShortcut(from, to, edges)

				// compute number of edges representing the shortcut (limited to 3)
				ec := int8(0)
				if edges[0].B == 0 {
					ec += 1
				} else {
					ec += shortcut_edgecount[edges[0].A]
				}
				if edges[1].B == 0 {
					ec += 1
				} else {
					ec += shortcut_edgecount[edges[1].A]
				}
				if ec > 3 {
					ec = 3
				}
				shortcut_edgecount.Add(ec)
			}
		}
		// set node to contracted
		is_contracted[node_id] = true

		// update neighbours
		for i := 0; i < len(in_neigbours); i++ {
			nb := in_neigbours[i]
			node_levels[nb] = Max(level+1, node_levels[nb])
			contracted_neighbours[nb] += 1
			if is_border[nb] {
				continue
			}
			prio := _ComputeNodePriority(nb, explorer, heap, flags, is_contracted, node_levels, contracted_neighbours, shortcut_edgecount, 100000)
			node_priorities[nb] = prio
			contraction_order.Enqueue(MakeTuple(nb, prio), prio)
		}
		for i := 0; i < len(out_neigbours); i++ {
			nb := out_neigbours[i]
			node_levels[nb] = Max(level+1, node_levels[nb])
			contracted_neighbours[nb] += 1
			if is_border[nb] {
				continue
			}
			prio := _ComputeNodePriority(nb, explorer, heap, flags, is_contracted, node_levels, contracted_neighbours, shortcut_edgecount, 100000)
			node_priorities[nb] = prio
			contraction_order.Enqueue(MakeTuple(nb, prio), prio)
		}
	}
	for i := 0; i < graph.NodeCount(); i++ {
		graph.SetNodeLevel(int32(i), node_levels[i])
	}
	fmt.Println("finished contracting graph")

	ch_data := TransformToCHData(graph)
	return ch_data
}

// Computes contraction using 2*ED + CN + EC + 5*L without hop-limits.
func CalcContraction6(base comps.IGraphBase, weight comps.IWeighting) *comps.CH {
	graph := TransformToCHPreprocGraph(base, weight)
	fmt.Println("started contracting graph...")

	// initialize
	is_contracted := NewArray[bool](graph.NodeCount())
	node_levels := NewArray[int16](graph.NodeCount())
	contracted_neighbours := NewArray[int](graph.NodeCount())
	shortcut_edgecount := NewList[int8](10)

	// initialize routing components
	heap := NewPriorityQueue[int32, int32](10)
	flags := NewFlags[_FlagSH](int32(graph.NodeCount()), _FlagSH{curr_length: 100000000})
	explorer := graph.GetExplorer()
	hop_limit := int32(10000000)

	// compute node priorities
	fmt.Println("computing priorities...")
	node_priorities := NewArray[int](graph.NodeCount())
	for i := 0; i < graph.NodeCount(); i++ {
		node_priorities[i] = _ComputeNodePriority(int32(i), explorer, heap, flags, is_contracted, node_levels, contracted_neighbours, shortcut_edgecount, hop_limit)
	}

	// put nodes into priority queue
	contraction_order := NewPriorityQueue[Tuple[int32, int], int](graph.NodeCount())
	for i := 0; i < graph.NodeCount(); i++ {
		prio := node_priorities[i]
		contraction_order.Enqueue(MakeTuple(int32(i), prio), prio)
	}

	fmt.Println("start contracting nodes...")
	count := 0
	for {
		temp, ok := contraction_order.Dequeue()
		if !ok {
			break
		}
		node_id := temp.A
		node_prio := temp.B
		if is_contracted[node_id] || node_prio != node_priorities[node_id] {
			continue
		}
		count += 1
		if count%1000 == 0 {
			fmt.Println("	node :", count, "/", graph.NodeCount())
		}

		// contract node
		level := node_levels[node_id]
		in_neigbours, out_neigbours := _FindNeighbours(explorer, node_id, is_contracted)
		ed := len(in_neigbours) + len(out_neigbours)
		for i := 0; i < len(in_neigbours); i++ {
			from := in_neigbours[i]
			heap.Clear()
			flags.Reset()
			_RunLocalSearch(from, out_neigbours, explorer, heap, flags, is_contracted, hop_limit)
			for j := 0; j < len(out_neigbours); j++ {
				to := out_neigbours[j]
				if from == to {
					continue
				}
				edges, shortcut_needed := _GetShortcut(from, to, node_id, explorer, flags)
				if !shortcut_needed {
					continue
				}
				// add shortcut to graph
				graph.AddShortcut(from, to, edges)
				ed -= 1

				// compute number of edges representing the shortcut (limited to 3)
				ec := int8(0)
				if edges[0].B == 0 {
					ec += 1
				} else {
					ec += shortcut_edgecount[edges[0].A]
				}
				if edges[1].B == 0 {
					ec += 1
				} else {
					ec += shortcut_edgecount[edges[1].A]
				}
				if ec > 3 {
					ec = 3
				}
				shortcut_edgecount.Add(ec)
			}
		}
		// set node to contracted
		is_contracted[node_id] = true

		// update neighbours
		for i := 0; i < len(in_neigbours); i++ {
			nb := in_neigbours[i]
			node_levels[nb] = Max(level+1, node_levels[nb])
			contracted_neighbours[nb] += 1
			prio := _ComputeNodePriority(nb, explorer, heap, flags, is_contracted, node_levels, contracted_neighbours, shortcut_edgecount, hop_limit)
			node_priorities[nb] = prio
			contraction_order.Enqueue(MakeTuple(nb, prio), prio)
		}
		for i := 0; i < len(out_neigbours); i++ {
			nb := out_neigbours[i]
			node_levels[nb] = Max(level+1, node_levels[nb])
			contracted_neighbours[nb] += 1
			prio := _ComputeNodePriority(nb, explorer, heap, flags, is_contracted, node_levels, contracted_neighbours, shortcut_edgecount, hop_limit)
			node_priorities[nb] = prio
			contraction_order.Enqueue(MakeTuple(nb, prio), prio)
		}
	}
	for i := 0; i < graph.NodeCount(); i++ {
		graph.SetNodeLevel(int32(i), node_levels[i])
	}
	fmt.Println("finished contracting graph")

	ch_data := TransformToCHData(graph)
	return ch_data
}

//*******************************************
// preprocess ch-index
//*******************************************

// Computes CH down-edges used in PHAST.
func PreparePHASTIndex(base comps.IGraphBase, weight comps.IWeighting, ch *comps.CH) *comps.CHIndex {
	temp_graph := graph.BuildCHGraph(base, weight, None[comps.IGraphIndex](), ch, None[*comps.CHIndex]())
	explorer := temp_graph.GetGraphExplorer()

	fwd_down_edges := NewList[structs.Shortcut](temp_graph.NodeCount())
	bwd_down_edges := NewList[structs.Shortcut](temp_graph.NodeCount())

	for i := 0; i < temp_graph.NodeCount(); i++ {
		this_id := int32(i)
		count := 0
		explorer.ForAdjacentEdges(this_id, graph.FORWARD, graph.ADJACENT_DOWNWARDS, func(ref graph.EdgeRef) {
			other_id := ref.OtherID
			fwd_down_edges.Add(structs.Shortcut{
				From:   this_id,
				To:     other_id,
				Weight: explorer.GetEdgeWeight(ref),
			})
			count += 1
		})
		for j := fwd_down_edges.Length() - count; j < fwd_down_edges.Length(); j++ {
			ch_edge := fwd_down_edges[j]
			structs.Shortcut_set_payload(&ch_edge, int32(count), 0)
			fwd_down_edges[j] = ch_edge
		}

		count = 0
		explorer.ForAdjacentEdges(this_id, graph.BACKWARD, graph.ADJACENT_DOWNWARDS, func(ref graph.EdgeRef) {
			other_id := ref.OtherID
			bwd_down_edges.Add(structs.Shortcut{
				From:   this_id,
				To:     other_id,
				Weight: explorer.GetEdgeWeight(ref),
			})
			count += 1
		})
		for j := bwd_down_edges.Length() - count; j < bwd_down_edges.Length(); j++ {
			ch_edge := bwd_down_edges[j]
			structs.Shortcut_set_payload(&ch_edge, int32(count), 0)
			bwd_down_edges[j] = ch_edge
		}
	}

	// sort down edges by node level
	sort.SliceStable(fwd_down_edges, func(i, j int) bool {
		e_i := fwd_down_edges[i]
		level_i := ch.GetNodeLevel(e_i.From)
		e_j := fwd_down_edges[j]
		level_j := ch.GetNodeLevel(e_j.From)
		return level_i > level_j
	})
	sort.SliceStable(bwd_down_edges, func(i, j int) bool {
		e_i := bwd_down_edges[i]
		level_i := ch.GetNodeLevel(e_i.From)
		e_j := bwd_down_edges[j]
		level_j := ch.GetNodeLevel(e_j.From)
		return level_i > level_j
	})

	return comps.NewCHIndex(Array[structs.Shortcut](fwd_down_edges), Array[structs.Shortcut](bwd_down_edges))
}

// Computes CH down-edges used in PHAST+GS.
// Has to be reordered with tile-level-ordering before calling this function.
func PrepareGSPHASTIndex(base comps.IGraphBase, weight comps.IWeighting, ch *comps.CH, partition *comps.Partition) *comps.CHIndex {
	g := graph.BuildGraph(base, weight, None[comps.IGraphIndex]())

	is_border := _IsBorderNode3(g, partition)

	temp_graph := graph.BuildCHGraph(base, weight, None[comps.IGraphIndex](), ch, None[*comps.CHIndex]())
	explorer := temp_graph.GetGraphExplorer()
	border_count := 0

	// initialize down edges lists
	fwd_down_edges := NewList[structs.Shortcut](temp_graph.NodeCount())
	bwd_down_edges := NewList[structs.Shortcut](temp_graph.NodeCount())

	// add overlay down-edges
	fwd_down_edges.Add(_CreateDummyShortcut(-1))
	bwd_down_edges.Add(_CreateDummyShortcut(-1))
	fwd_other_edges := NewDict[int16, List[structs.Shortcut]](100)
	bwd_other_edges := NewDict[int16, List[structs.Shortcut]](100)
	for i := 0; i < temp_graph.NodeCount(); i++ {
		this_id := int32(i)
		this_tile := partition.GetNodeTile(this_id)
		if !is_border[this_id] {
			border_count = i + 1
			break
		}
		explorer.ForAdjacentEdges(this_id, graph.FORWARD, graph.ADJACENT_DOWNWARDS, func(ref graph.EdgeRef) {
			other_id := ref.OtherID
			other_tile := partition.GetNodeTile(other_id)
			edge := structs.Shortcut{
				From:   this_id,
				To:     other_id,
				Weight: explorer.GetEdgeWeight(ref),
			}
			structs.Shortcut_set_payload(&edge, other_tile, 0)
			if !is_border[other_id] {
				var edges List[structs.Shortcut]
				if fwd_other_edges.ContainsKey(this_tile) {
					edges = fwd_other_edges[this_tile]
				} else {
					edges = NewList[structs.Shortcut](10)
				}
				edges.Add(edge)
				fwd_other_edges[this_tile] = edges
			} else {
				fwd_down_edges.Add(edge)
			}
		})
		explorer.ForAdjacentEdges(this_id, graph.BACKWARD, graph.ADJACENT_DOWNWARDS, func(ref graph.EdgeRef) {
			other_id := ref.OtherID
			other_tile := partition.GetNodeTile(other_id)
			edge := structs.Shortcut{
				From:   this_id,
				To:     other_id,
				Weight: explorer.GetEdgeWeight(ref),
			}
			structs.Shortcut_set_payload(&edge, other_tile, 0)
			if !is_border[other_id] {
				var edges List[structs.Shortcut]
				if bwd_other_edges.ContainsKey(this_tile) {
					edges = bwd_other_edges[this_tile]
				} else {
					edges = NewList[structs.Shortcut](10)
				}
				edges.Add(edge)
				bwd_other_edges[this_tile] = edges
			} else {
				bwd_down_edges.Add(edge)
			}
		})
	}
	// add other down edges
	curr_tile := int16(-1)
	for i := border_count; i < temp_graph.NodeCount(); i++ {
		this_id := int32(i)
		this_tile := partition.GetNodeTile(this_id)
		if this_tile != curr_tile {
			fwd_down_edges.Add(_CreateDummyShortcut(this_tile))
			bwd_down_edges.Add(_CreateDummyShortcut(this_tile))
			curr_tile = this_tile
			if fwd_other_edges.ContainsKey(this_tile) {
				edges := fwd_other_edges[this_tile]
				for _, edge := range edges {
					fwd_down_edges.Add(edge)
				}
			}
			if bwd_other_edges.ContainsKey(this_tile) {
				edges := bwd_other_edges[this_tile]
				for _, edge := range edges {
					bwd_down_edges.Add(edge)
				}
			}
		}
		explorer.ForAdjacentEdges(this_id, graph.FORWARD, graph.ADJACENT_DOWNWARDS, func(ref graph.EdgeRef) {
			other_id := ref.OtherID
			fwd_down_edges.Add(structs.Shortcut{
				From:   this_id,
				To:     other_id,
				Weight: explorer.GetEdgeWeight(ref),
			})
		})
		explorer.ForAdjacentEdges(this_id, graph.BACKWARD, graph.ADJACENT_DOWNWARDS, func(ref graph.EdgeRef) {
			other_id := ref.OtherID
			bwd_down_edges.Add(structs.Shortcut{
				From:   this_id,
				To:     other_id,
				Weight: explorer.GetEdgeWeight(ref),
			})
		})
	}

	// set count in dummy edges
	fwd_id := 0
	fwd_count := 0
	for i := 0; i < fwd_down_edges.Length(); i++ {
		edge := fwd_down_edges[i]
		is_dummy := structs.Shortcut_get_payload[bool](&edge, 2)
		if is_dummy {
			// set count in previous dummy
			fwd_down_edges[fwd_id].To = int32(fwd_count)
			// reset count
			fwd_id = i
			fwd_count = 0
			continue
		}
		fwd_count += 1
	}
	fwd_down_edges[fwd_id].To = int32(fwd_count)
	bwd_id := 0
	bwd_count := 0
	for i := 0; i < bwd_down_edges.Length(); i++ {
		edge := bwd_down_edges[i]
		is_dummy := structs.Shortcut_get_payload[bool](&edge, 2)
		if is_dummy {
			// set count in previous dummy
			bwd_down_edges[bwd_id].To = int32(bwd_count)
			// reset count
			bwd_id = i
			bwd_count = 0
			continue
		}
		bwd_count += 1
	}
	bwd_down_edges[bwd_id].To = int32(bwd_count)

	return comps.NewCHIndex(Array[structs.Shortcut](fwd_down_edges), Array[structs.Shortcut](bwd_down_edges))
}

func _CreateDummyShortcut(to_tile int16) structs.Shortcut {
	dummy := structs.Shortcut{}
	structs.Shortcut_set_payload(&dummy, to_tile, 0)
	structs.Shortcut_set_payload(&dummy, true, 2)
	return dummy
}

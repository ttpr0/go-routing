package graph

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// preprocessing graph
//*******************************************

type CHPreprocGraph struct {
	// added attributes to build ch
	ch_topology AdjacencyList
	node_levels Array[int16]
	shortcuts   List[CHShortcut]
	sh_weight   List[int32]

	// underlying base graph
	store    GraphStore
	topology AdjacencyArray
	weight   DefaultWeighting
	index    KDTree[int32]
}

type DynamicNodeRef struct {
	FWDEdgeRefs List[EdgeRef]
	BWDEdgeRefs List[EdgeRef]
}

func (self *CHPreprocGraph) GetExplorer() *CHPreprocGraphExplorer {
	return &CHPreprocGraphExplorer{
		graph:       self,
		accessor:    self.topology.GetAccessor(),
		sh_accessor: self.ch_topology.GetAccessor(),
	}
}
func (self *CHPreprocGraph) NodeCount() int {
	return self.store.NodeCount()
}
func (self *CHPreprocGraph) EdgeCount() int {
	return self.store.EdgeCount()
}
func (self *CHPreprocGraph) GetNode(node int32) Node {
	return self.store.GetNode(node)
}
func (self *CHPreprocGraph) GetEdge(edge int32) Edge {
	return self.store.GetEdge(edge)
}
func (self *CHPreprocGraph) GetShortcut(id int32) CHShortcut {
	return self.shortcuts[id]
}
func (self *CHPreprocGraph) GetWeight(id int32, is_shortcut bool) int32 {
	if is_shortcut {
		return self.sh_weight[id]
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
	shortcut := CHShortcut{
		NodeA:  node_a,
		NodeB:  node_b,
		_Edges: edges,
	}
	shc_id := self.shortcuts.Length()
	self.sh_weight.Add(weight)
	self.shortcuts.Add(shortcut)

	self.ch_topology.AddEdgeEntries(node_a, node_b, int32(shc_id), 100)
}

type CHPreprocGraphExplorer struct {
	graph       *CHPreprocGraph
	accessor    AdjArrayAccessor
	sh_accessor AdjListAccessor
}

func (self *CHPreprocGraphExplorer) ForAdjacentEdges(node int32, direction Direction, typ Adjacency, callback func(EdgeRef)) {
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
}
func (self *CHPreprocGraphExplorer) GetEdgeWeight(edge EdgeRef) int32 {
	return self.graph.GetWeight(edge.EdgeID, edge.IsCHShortcut())
}
func (self *CHPreprocGraphExplorer) GetTurnCost(from EdgeRef, via int32, to EdgeRef) int32 {
	return 0
}
func (self *CHPreprocGraphExplorer) GetOtherNode(edge_id, node int32, is_shortcut bool) int32 {
	if is_shortcut {
		e := self.graph.GetShortcut(edge_id)
		if node == e.NodeA {
			return e.NodeB
		}
		if node == e.NodeB {
			return e.NodeA
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
	self.accessor.SetBaseNode(from, FORWARD)
	for self.accessor.Next() {
		edge_id := self.accessor.GetEdgeID()
		other_id := self.accessor.GetOtherID()
		if other_id == to {
			return self.graph.GetWeight(edge_id, false)
		}
	}
	self.sh_accessor.SetBaseNode(from, FORWARD)
	for self.sh_accessor.Next() {
		edge_id := self.sh_accessor.GetEdgeID()
		other_id := self.sh_accessor.GetOtherID()
		if other_id == to {
			return self.graph.GetWeight(edge_id, true)
		}
	}
	return -1
}
func (self *CHPreprocGraphExplorer) GetEdgeBetween(from, to int32) (EdgeRef, bool) {
	self.accessor.SetBaseNode(from, FORWARD)
	for self.accessor.Next() {
		edge_id := self.accessor.GetEdgeID()
		other_id := self.accessor.GetOtherID()
		if other_id == to {
			return EdgeRef{
				EdgeID:  edge_id,
				_Type:   0,
				OtherID: to,
			}, true
		}
	}
	self.sh_accessor.SetBaseNode(from, FORWARD)
	for self.sh_accessor.Next() {
		edge_id := self.sh_accessor.GetEdgeID()
		other_id := self.sh_accessor.GetOtherID()
		if other_id == to {
			return EdgeRef{
				EdgeID:  edge_id,
				_Type:   100,
				OtherID: to,
			}, true
		}
	}
	return EdgeRef{}, false
}

//*******************************************
// transform to/from dynamic graph
//*******************************************

func TransformToCHPreprocGraph(g *Graph) *CHPreprocGraph {
	ch_topology := NewAdjacencyList(g.NodeCount())
	node_levels := NewArray[int16](g.NodeCount())

	for i := 0; i < g.NodeCount(); i++ {
		node_levels[i] = 0
	}

	dg := CHPreprocGraph{
		ch_topology: ch_topology,
		node_levels: node_levels,
		shortcuts:   NewList[CHShortcut](100),
		sh_weight:   NewList[int32](100),
		store:       g.store,
		topology:    g.topology,
		weight:      g.weight,
		index:       g.index,
	}

	return &dg
}

func TransformFromCHPreprocGraph(dg *CHPreprocGraph) *CHGraph {
	g := CHGraph{
		store:       dg.store,
		topology:    dg.topology,
		ch_topology: *AdjacencyListToArray(&dg.ch_topology),
		ch_store: CHStore{
			shortcuts:   Array[CHShortcut](dg.shortcuts),
			node_levels: dg.node_levels,
			sh_weight:   Array[int32](dg.sh_weight),
		},
		weight: dg.weight,
		index:  dg.index,
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
func FindNeighbours(explorer *CHPreprocGraphExplorer, id int32, is_contracted Array[bool]) ([]int32, []int32) {
	// compute out-going neighbours
	out_neigbours := NewList[int32](4)
	explorer.ForAdjacentEdges(id, FORWARD, ADJACENT_ALL, func(ref EdgeRef) {
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
	explorer.ForAdjacentEdges(id, BACKWARD, ADJACENT_ALL, func(ref EdgeRef) {
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
func _RunLocalSearch(start int32, targets List[int32], explorer *CHPreprocGraphExplorer, heap PriorityQueue[int32, int32], flags Array[_FlagSH], flag_count int32, is_contracted Array[bool], hop_limit int32) {
	flags[start] = _FlagSH{
		_counter:    flag_count,
		curr_length: 0,
	}
	target_count := targets.Length()
	for _, target := range targets {
		flags[target] = _FlagSH{
			_counter:    flag_count,
			curr_length: 1000000000,
			_is_target:  true,
		}
	}
	start_flag := flags[start]
	start_flag.curr_length = 0
	flags[start] = start_flag
	heap.Enqueue(start, 0)

	found_count := 0
	for {
		curr_id, ok := heap.Dequeue()
		if !ok {
			break
		}
		curr_flag := flags[curr_id]
		if curr_flag.visited {
			continue
		}
		curr_flag.visited = true
		flags[curr_id] = curr_flag
		if curr_flag._is_target {
			found_count += 1
		}
		if found_count >= target_count {
			break
		}
		if curr_flag.curr_hops >= hop_limit {
			continue
		}
		explorer.ForAdjacentEdges(curr_id, FORWARD, ADJACENT_ALL, func(ref EdgeRef) {
			edge_id := ref.EdgeID
			other_id := ref.OtherID
			if is_contracted[other_id] {
				return
			}
			other_flag := flags[other_id]
			if other_flag._counter != flag_count {
				other_flag = _FlagSH{
					_counter:    flag_count,
					curr_length: 1000000000,
				}
			}
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
			flags[other_id] = other_flag
		})
	}
}

type _FlagSH struct {
	_counter    int32
	curr_length int32
	curr_hops   int32
	prev_edge   int32
	prev_node   int32
	is_shortcut bool
	visited     bool
	_is_target  bool
}

//*******************************************
// preprocess ch
//*******************************************

func CalcContraction(graph *CHPreprocGraph) {
	fmt.Println("started contracting graph")
	// initialize graph
	//graph.resetContraction();
	for i := 0; i < graph.NodeCount(); i++ {
		graph.SetNodeLevel(int32(i), 0)
	}

	is_contracted := NewArray[bool](graph.NodeCount())
	heap := NewPriorityQueue[int32, int32](10)
	flags := NewArray[_FlagSH](graph.NodeCount())
	flag_count := int32(1)
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
			c1 := graph.topology.GetDegree(a, FORWARD) + graph.topology.GetDegree(a, BACKWARD)
			c2 := graph.ch_topology.GetDegree(a, FORWARD) + graph.ch_topology.GetDegree(a, BACKWARD)
			count_a := c1 + c2
			b := nodes[j]
			c1 = graph.topology.GetDegree(b, FORWARD) + graph.topology.GetDegree(b, BACKWARD)
			c2 = graph.ch_topology.GetDegree(b, FORWARD) + graph.ch_topology.GetDegree(b, BACKWARD)
			count_b := c1 + c2
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
			in_neigbours, out_neigbours := FindNeighbours(explorer, node_id, is_contracted)
			for i := 0; i < len(in_neigbours); i++ {
				from := in_neigbours[i]
				heap.Clear()
				_RunLocalSearch(from, out_neigbours, explorer, heap, flags, flag_count, is_contracted, 1000000)
				for j := 0; j < len(out_neigbours); j++ {
					to := out_neigbours[j]
					if from == to {
						continue
					}
					edges := [2]Tuple[int32, byte]{}

					to_flag := flags[to]
					// is target hasnt been found by search always add shortcut
					if !to_flag.visited || to_flag._counter != flag_count {
						t_edge, _ := explorer.GetEdgeBetween(node_id, to)
						if t_edge.IsCHShortcut() {
							edges[0] = MakeTuple(t_edge.EdgeID, byte(2))
						} else {
							edges[0] = MakeTuple(t_edge.EdgeID, byte(0))
						}
						f_edge, _ := explorer.GetEdgeBetween(from, node_id)
						if f_edge.IsCHShortcut() {
							edges[1] = MakeTuple(f_edge.EdgeID, byte(2))
						} else {
							edges[1] = MakeTuple(f_edge.EdgeID, byte(0))
						}
					} else {
						// check if shortest path goes through node
						if to_flag.prev_node != node_id {
							continue
						}
						node_flag := flags[node_id]
						if node_flag.prev_node != from {
							continue
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
					}

					// add shortcut to graph
					graph.AddShortcut(from, to, edges)
				}
				flag_count += 1
				if flag_count > 1000 {
					flag_count = 3
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

func CalcContraction2(graph *CHPreprocGraph, contraction_order Array[int32]) {
	fmt.Println("started contracting graph")
	// initialize graph
	for i := 0; i < graph.NodeCount(); i++ {
		graph.SetNodeLevel(int32(i), 0)
	}
	is_contracted := NewArray[bool](graph.NodeCount())
	heap := NewPriorityQueue[int32, int32](10)
	flags := NewArray[_FlagSH](graph.NodeCount())
	flag_count := int32(1)
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
		in_neigbours, out_neigbours := FindNeighbours(explorer, node_id, is_contracted)
		t2 := time.Now()
		for i := 0; i < len(in_neigbours); i++ {
			from := in_neigbours[i]
			heap.Clear()
			_RunLocalSearch(from, out_neigbours, explorer, heap, flags, flag_count, is_contracted, 1000000)
			for j := 0; j < len(out_neigbours); j++ {
				to := out_neigbours[j]
				if from == to {
					continue
				}
				edges := [2]Tuple[int32, byte]{}

				to_flag := flags[to]
				// is target hasnt been found by search always add shortcut
				if !to_flag.visited || to_flag._counter != flag_count {
					t_edge, _ := explorer.GetEdgeBetween(node_id, to)
					if t_edge.IsCHShortcut() {
						edges[0] = MakeTuple(t_edge.EdgeID, byte(2))
					} else {
						edges[0] = MakeTuple(t_edge.EdgeID, byte(0))
					}
					f_edge, _ := explorer.GetEdgeBetween(from, node_id)
					if f_edge.IsCHShortcut() {
						edges[1] = MakeTuple(f_edge.EdgeID, byte(2))
					} else {
						edges[1] = MakeTuple(f_edge.EdgeID, byte(0))
					}
				} else {
					// check if shortest path goes through node
					if to_flag.prev_node != node_id {
						continue
					}
					node_flag := flags[node_id]
					if node_flag.prev_node != from {
						continue
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
				}

				// add shortcut to graph
				graph.AddShortcut(from, to, edges)
			}
			flag_count += 1
			if flag_count > 1000 {
				flag_count = 3
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

func SimpleNodeOrdering(graph *CHPreprocGraph) Array[int32] {
	nodes := NewArray[int32](graph.NodeCount())
	for i := 0; i < graph.NodeCount(); i++ {
		nodes[i] = int32(i)
	}

	// sort nodes by number of adjacent edges
	fmt.Println("start ordering nodes")
	sort.Slice(nodes, func(i, j int) bool {
		a := nodes[i]
		c1 := graph.topology.GetDegree(a, FORWARD) + graph.topology.GetDegree(a, BACKWARD)
		c2 := graph.ch_topology.GetDegree(a, FORWARD) + graph.ch_topology.GetDegree(a, BACKWARD)
		count_a := c1 + c2
		b := nodes[j]
		c1 = graph.topology.GetDegree(b, FORWARD) + graph.topology.GetDegree(b, BACKWARD)
		c2 = graph.ch_topology.GetDegree(b, FORWARD) + graph.ch_topology.GetDegree(b, BACKWARD)
		count_b := c1 + c2
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
		start := rand.Int31n(int32(graph.NodeCount()))
		end := rand.Int31n(int32(graph.NodeCount()))
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
		explorer.ForAdjacentEdges(curr_id, FORWARD, ADJACENT_ALL, func(ref EdgeRef) {
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
		curr_id = explorer.GetOtherNode(CreateEdgeRef(edge), curr_id)
	}
}

//*******************************************
// preprocess ch 3
//*******************************************

func CalcContraction3(graph *CHPreprocGraph) {
	fmt.Println("started contracting graph...")

	// initialize
	is_contracted := NewArray[bool](graph.NodeCount())
	node_levels := NewArray[int16](graph.NodeCount())
	contracted_neighbours := NewArray[int](graph.NodeCount())
	shortcut_edgecount := NewList[int8](10)

	// initialize routing components
	heap := NewPriorityQueue[int32, int32](10)
	flags := NewArray[_FlagSH](graph.NodeCount())
	explorer := graph.GetExplorer()

	// compute node priorities
	fmt.Println("computing priorities...")
	node_priorities := NewArray[int](graph.NodeCount())
	for i := 0; i < graph.NodeCount(); i++ {
		node_priorities[i] = _ComputeNodePriority(int32(i), explorer, heap, flags, [2]int32{1, 2}, is_contracted, node_levels, contracted_neighbours, shortcut_edgecount)
	}

	// put nodes into priority queue
	contraction_order := NewPriorityQueue[Tuple[int32, int], int](graph.NodeCount())
	for i := 0; i < graph.NodeCount(); i++ {
		prio := node_priorities[i]
		contraction_order.Enqueue(MakeTuple(int32(i), prio), prio)
	}

	fmt.Println("start contracting nodes...")
	flag_count := int32(3)
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
		in_neigbours, out_neigbours := FindNeighbours(explorer, node_id, is_contracted)
		for i := 0; i < len(in_neigbours); i++ {
			from := in_neigbours[i]
			heap.Clear()
			_RunLocalSearch(from, out_neigbours, explorer, heap, flags, flag_count, is_contracted, 1000000)
			for j := 0; j < len(out_neigbours); j++ {
				to := out_neigbours[j]
				if from == to {
					continue
				}
				edges := [2]Tuple[int32, byte]{}

				to_flag := flags[to]
				// is target hasnt been found by search always add shortcut
				if !to_flag.visited || to_flag._counter != flag_count {
					t_edge, _ := explorer.GetEdgeBetween(node_id, to)
					if t_edge.IsCHShortcut() {
						edges[0] = MakeTuple(t_edge.EdgeID, byte(2))
					} else {
						edges[0] = MakeTuple(t_edge.EdgeID, byte(0))
					}
					f_edge, _ := explorer.GetEdgeBetween(from, node_id)
					if f_edge.IsCHShortcut() {
						edges[1] = MakeTuple(f_edge.EdgeID, byte(2))
					} else {
						edges[1] = MakeTuple(f_edge.EdgeID, byte(0))
					}
				} else {
					// check if shortest path goes through node
					if to_flag.prev_node != node_id {
						continue
					}
					node_flag := flags[node_id]
					if node_flag.prev_node != from {
						continue
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
			flag_count += 1
			if flag_count > 1000 {
				flag_count = 3
			}
		}
		// set node to contracted
		is_contracted[node_id] = true

		// update neighbours
		for i := 0; i < len(in_neigbours); i++ {
			nb := in_neigbours[i]
			node_levels[nb] = Max(level+1, node_levels[nb])
			contracted_neighbours[nb] += 1
			prio := _ComputeNodePriority(nb, explorer, heap, flags, [2]int32{1, 2}, is_contracted, node_levels, contracted_neighbours, shortcut_edgecount)
			node_priorities[nb] = prio
			contraction_order.Enqueue(MakeTuple(nb, prio), prio)
		}
		for i := 0; i < len(out_neigbours); i++ {
			nb := out_neigbours[i]
			node_levels[nb] = Max(level+1, node_levels[nb])
			contracted_neighbours[nb] += 1
			prio := _ComputeNodePriority(nb, explorer, heap, flags, [2]int32{1, 2}, is_contracted, node_levels, contracted_neighbours, shortcut_edgecount)
			node_priorities[nb] = prio
			contraction_order.Enqueue(MakeTuple(nb, prio), prio)
		}
	}
	for i := 0; i < graph.NodeCount(); i++ {
		graph.SetNodeLevel(int32(i), node_levels[i])
	}
	fmt.Println("finished contracting graph")
}

func _ComputeNodePriority(node int32, explorer *CHPreprocGraphExplorer, heap PriorityQueue[int32, int32], flags Array[_FlagSH], flag_counts [2]int32, is_contracted Array[bool], node_levels Array[int16], contracted_neighbours Array[int], shortcut_edgecount List[int8]) int {
	in_neigbours, out_neigbours := FindNeighbours(explorer, node, is_contracted)
	edge_diff := -(len(in_neigbours) + len(out_neigbours))
	edge_count := int8(0)
	for i := 0; i < len(in_neigbours); i++ {
		from := in_neigbours[i]
		flag_count := flag_counts[i%2]
		heap.Clear()
		_RunLocalSearch(from, out_neigbours, explorer, heap, flags, flag_count, is_contracted, 1000000)
		for j := 0; j < len(out_neigbours); j++ {
			to := out_neigbours[j]
			if from == to {
				continue
			}
			edges := [2]Tuple[int32, byte]{}
			to_flag := flags[to]
			if to_flag._counter != flag_count {
				t_edge, _ := explorer.GetEdgeBetween(node, to)
				if t_edge.IsCHShortcut() {
					edges[0] = MakeTuple(t_edge.EdgeID, byte(2))
				} else {
					edges[0] = MakeTuple(t_edge.EdgeID, byte(0))
				}
				f_edge, _ := explorer.GetEdgeBetween(from, node)
				if f_edge.IsCHShortcut() {
					edges[1] = MakeTuple(f_edge.EdgeID, byte(2))
				} else {
					edges[1] = MakeTuple(f_edge.EdgeID, byte(0))
				}
			} else {
				// check if shortest path goes through node
				if to_flag.prev_node != node {
					continue
				}
				node_flag := flags[node]
				if node_flag.prev_node != from {
					continue
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

func CalcContraction4(graph *CHPreprocGraph) {
	fmt.Println("started contracting graph...")

	// initialize
	is_contracted := NewArray[bool](graph.NodeCount())
	node_levels := NewArray[int16](graph.NodeCount())
	contracted_neighbours := NewArray[int](graph.NodeCount())

	// initialize routing components
	heap := NewPriorityQueue[int32, int32](10)
	flags := NewArray[_FlagSH](graph.NodeCount())
	explorer := graph.GetExplorer()

	// compute node priorities
	fmt.Println("computing priorities...")
	node_priorities := NewArray[int](graph.NodeCount())
	for i := 0; i < graph.NodeCount(); i++ {
		node_priorities[i] = _ComputeNodePriority2(int32(i), explorer, heap, flags, [2]int32{1, 2}, is_contracted, node_levels, contracted_neighbours)
	}

	// put nodes into priority queue
	contraction_order := NewPriorityQueue[Tuple[int32, int], int](graph.NodeCount())
	for i := 0; i < graph.NodeCount(); i++ {
		prio := node_priorities[i]
		contraction_order.Enqueue(MakeTuple(int32(i), prio), prio)
	}

	fmt.Println("start contracting nodes...")
	flag_count := int32(3)
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
		in_neigbours, out_neigbours := FindNeighbours(explorer, node_id, is_contracted)
		for i := 0; i < len(in_neigbours); i++ {
			from := in_neigbours[i]
			heap.Clear()
			_RunLocalSearch(from, out_neigbours, explorer, heap, flags, flag_count, is_contracted, 1000000)
			for j := 0; j < len(out_neigbours); j++ {
				to := out_neigbours[j]
				if from == to {
					continue
				}
				edges := [2]Tuple[int32, byte]{}

				to_flag := flags[to]
				// is target hasnt been found by search always add shortcut
				if !to_flag.visited || to_flag._counter != flag_count {
					t_edge, _ := explorer.GetEdgeBetween(node_id, to)
					if t_edge.IsCHShortcut() {
						edges[0] = MakeTuple(t_edge.EdgeID, byte(2))
					} else {
						edges[0] = MakeTuple(t_edge.EdgeID, byte(0))
					}
					f_edge, _ := explorer.GetEdgeBetween(from, node_id)
					if f_edge.IsCHShortcut() {
						edges[1] = MakeTuple(f_edge.EdgeID, byte(2))
					} else {
						edges[1] = MakeTuple(f_edge.EdgeID, byte(0))
					}
				} else {
					// check if shortest path goes through node
					if to_flag.prev_node != node_id {
						continue
					}
					node_flag := flags[node_id]
					if node_flag.prev_node != from {
						continue
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
				}

				// add shortcut to graph
				graph.AddShortcut(from, to, edges)
			}
			flag_count += 1
			if flag_count > 1000 {
				flag_count = 3
			}
		}
		// set node to contracted
		is_contracted[node_id] = true

		// update neighbours
		for i := 0; i < len(in_neigbours); i++ {
			nb := in_neigbours[i]
			node_levels[nb] = Max(level+1, node_levels[nb])
			contracted_neighbours[nb] += 1
			prio := _ComputeNodePriority2(nb, explorer, heap, flags, [2]int32{1, 2}, is_contracted, node_levels, contracted_neighbours)
			node_priorities[nb] = prio
			contraction_order.Enqueue(MakeTuple(nb, prio), prio)
		}
		for i := 0; i < len(out_neigbours); i++ {
			nb := out_neigbours[i]
			node_levels[nb] = Max(level+1, node_levels[nb])
			contracted_neighbours[nb] += 1
			prio := _ComputeNodePriority2(nb, explorer, heap, flags, [2]int32{1, 2}, is_contracted, node_levels, contracted_neighbours)
			node_priorities[nb] = prio
			contraction_order.Enqueue(MakeTuple(nb, prio), prio)
		}
	}
	for i := 0; i < graph.NodeCount(); i++ {
		graph.SetNodeLevel(int32(i), node_levels[i])
	}
	fmt.Println("finished contracting graph")
}

func _ComputeNodePriority2(node int32, explorer *CHPreprocGraphExplorer, heap PriorityQueue[int32, int32], flags Array[_FlagSH], flag_counts [2]int32, is_contracted Array[bool], node_levels Array[int16], contracted_neighbours Array[int]) int {
	in_neigbours, out_neigbours := FindNeighbours(explorer, node, is_contracted)
	edge_diff := -(len(in_neigbours) + len(out_neigbours))
	for i := 0; i < len(in_neigbours); i++ {
		from := in_neigbours[i]
		flag_count := flag_counts[i%2]
		heap.Clear()
		_RunLocalSearch(from, out_neigbours, explorer, heap, flags, flag_count, is_contracted, 1000000)
		for j := 0; j < len(out_neigbours); j++ {
			to := out_neigbours[j]
			if from == to {
				continue
			}
			edges := [2]Tuple[int32, byte]{}
			to_flag := flags[to]
			if to_flag._counter != flag_count {
				t_edge, _ := explorer.GetEdgeBetween(node, to)
				if t_edge.IsCHShortcut() {
					edges[0] = MakeTuple(t_edge.EdgeID, byte(2))
				} else {
					edges[0] = MakeTuple(t_edge.EdgeID, byte(0))
				}
				f_edge, _ := explorer.GetEdgeBetween(from, node)
				if f_edge.IsCHShortcut() {
					edges[1] = MakeTuple(f_edge.EdgeID, byte(2))
				} else {
					edges[1] = MakeTuple(f_edge.EdgeID, byte(0))
				}
			} else {
				// check if shortest path goes through node
				if to_flag.prev_node != node {
					continue
				}
				node_flag := flags[node]
				if node_flag.prev_node != from {
					continue
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
func CalcContraction5(graph *CHPreprocGraph, node_tiles Array[int16]) {
	fmt.Println("started contracting graph...")

	// initialize
	is_contracted := NewArray[bool](graph.NodeCount())
	node_levels := NewArray[int16](graph.NodeCount())
	contracted_neighbours := NewArray[int](graph.NodeCount())
	shortcut_edgecount := NewList[int8](10)

	// initialize routing components
	heap := NewPriorityQueue[int32, int32](10)
	flags := NewArray[_FlagSH](graph.NodeCount())
	explorer := graph.GetExplorer()

	// compute node priorities
	fmt.Println("computing priorities...")
	is_border := _IsBorderNode(graph, node_tiles)
	border_count := 0
	node_priorities := NewArray[int](graph.NodeCount())
	contraction_order := NewPriorityQueue[Tuple[int32, int], int](graph.NodeCount())
	for i := 0; i < graph.NodeCount(); i++ {
		if is_border[i] {
			node_priorities[i] = 10000000000
			border_count += 1
		}
		prio := _ComputeNodePriority(int32(i), explorer, heap, flags, [2]int32{1, 2}, is_contracted, node_levels, contracted_neighbours, shortcut_edgecount)
		node_priorities[i] = prio
		contraction_order.Enqueue(MakeTuple(int32(i), prio), prio)
	}

	fmt.Println("start contracting nodes...")
	flag_count := int32(3)
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
				prio := _ComputeNodePriority(int32(i), explorer, heap, flags, [2]int32{1, 2}, is_contracted, node_levels, contracted_neighbours, shortcut_edgecount)
				node_priorities[i] = prio
				contraction_order.Enqueue(MakeTuple(int32(i), prio), prio)
			}
		}

		// contract node
		level := node_levels[node_id]
		in_neigbours, out_neigbours := FindNeighbours(explorer, node_id, is_contracted)
		for i := 0; i < len(in_neigbours); i++ {
			from := in_neigbours[i]
			heap.Clear()
			_RunLocalSearch(from, out_neigbours, explorer, heap, flags, flag_count, is_contracted, 1000000)
			for j := 0; j < len(out_neigbours); j++ {
				to := out_neigbours[j]
				if from == to {
					continue
				}
				edges := [2]Tuple[int32, byte]{}

				to_flag := flags[to]
				// is target hasnt been found by search always add shortcut
				if !to_flag.visited || to_flag._counter != flag_count {
					t_edge, _ := explorer.GetEdgeBetween(node_id, to)
					if t_edge.IsCHShortcut() {
						edges[0] = MakeTuple(t_edge.EdgeID, byte(2))
					} else {
						edges[0] = MakeTuple(t_edge.EdgeID, byte(0))
					}
					f_edge, _ := explorer.GetEdgeBetween(from, node_id)
					if f_edge.IsCHShortcut() {
						edges[1] = MakeTuple(f_edge.EdgeID, byte(2))
					} else {
						edges[1] = MakeTuple(f_edge.EdgeID, byte(0))
					}
				} else {
					// check if shortest path goes through node
					if to_flag.prev_node != node_id {
						continue
					}
					node_flag := flags[node_id]
					if node_flag.prev_node != from {
						continue
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
			flag_count += 1
			if flag_count > 1000 {
				flag_count = 3
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
			prio := _ComputeNodePriority(nb, explorer, heap, flags, [2]int32{1, 2}, is_contracted, node_levels, contracted_neighbours, shortcut_edgecount)
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
			prio := _ComputeNodePriority(nb, explorer, heap, flags, [2]int32{1, 2}, is_contracted, node_levels, contracted_neighbours, shortcut_edgecount)
			node_priorities[nb] = prio
			contraction_order.Enqueue(MakeTuple(nb, prio), prio)
		}
	}
	for i := 0; i < graph.NodeCount(); i++ {
		graph.SetNodeLevel(int32(i), node_levels[i])
	}
	fmt.Println("finished contracting graph")
}

func _IsBorderNode(graph *CHPreprocGraph, node_tiles Array[int16]) Array[bool] {
	is_border := NewArray[bool](graph.NodeCount())

	explorer := graph.GetExplorer()
	for i := 0; i < graph.NodeCount(); i++ {
		explorer.ForAdjacentEdges(int32(i), FORWARD, ADJACENT_ALL, func(ref EdgeRef) {
			if node_tiles[i] != node_tiles[ref.OtherID] {
				is_border[i] = true
			}
		})
		explorer.ForAdjacentEdges(int32(i), BACKWARD, ADJACENT_ALL, func(ref EdgeRef) {
			if node_tiles[i] != node_tiles[ref.OtherID] {
				is_border[i] = true
			}
		})
	}

	return is_border
}

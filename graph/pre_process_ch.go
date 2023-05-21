package graph

import (
	"fmt"
	"sort"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

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
					if i == j {
						continue
					}
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

type Flag struct {
	pathlength int32
	prevEdge   int32
	visited    bool
}

type FlagSH struct {
	pathlength int32
	prevEdge   int32
	isShortcut bool
	visited    bool
}

type FlagCH struct {
	pathlength1 int32
	prevEdge1   int32
	isShortcut1 bool
	visited1    bool

	pathlength2 int32
	prevEdge2   int32
	isShortcut2 bool
	visited2    bool
}

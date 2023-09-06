package main

import (
	"fmt"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/algorithm"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func prepare() {
	g := graph.LoadGraph("./graphs/niedersachsen")

	// compute closely connected components
	groups := NewArray[int](g.NodeCount())
	group := 1
	for i := 0; i < g.NodeCount(); i++ {
		if groups[i] != 0 {
			continue
		}
		fmt.Println("iteration:", group)
		start := int32(i)
		visited := algorithm.CalcBidirectBFS(g, start)
		for i := 0; i < g.NodeCount(); i++ {
			if visited[i] {
				if groups[i] != 0 {
					fmt.Println("failure 1")
				}
				groups[i] = group
			}
		}
		group += 1
	}

	// get largest group
	counts := NewArray[int](group + 1)
	for i := 0; i < g.NodeCount(); i++ {
		counts[groups[i]] += 1
	}
	max_group := 0
	for i := 0; i < counts.Length(); i++ {
		if counts[i] > counts[max_group] {
			max_group = i
		}
	}

	// get nodes to be removed
	remove := NewList[int32](100)
	for i := 0; i < g.NodeCount(); i++ {
		if groups[i] != max_group {
			remove.Add(int32(i))
		}
	}
	fmt.Println("remove", remove.Length(), "nodes")

	// remove nodes from graph
	new_g := graph.RemoveNodes(g.(*graph.Graph), remove)

	graph.StoreGraph(new_g, "./graphs/niedersachsen_sub")
}

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
	groups := algorithm.ConnectedComponents(g)

	// get largest group
	max_group := GetMostCommon(groups)

	// get nodes to be removed
	remove := NewList[int32](100)
	for i := 0; i < g.NodeCount(); i++ {
		if groups[i] != max_group {
			remove.Add(int32(i))
		}
	}
	fmt.Println("remove", remove.Length(), "nodes")

	// remove nodes from graph
	new_g := graph.RemoveNodes(g, remove)

	graph.StoreGraph(new_g, "./graphs/niedersachsen_sub")
}

func GetMostCommon[T comparable](arr Array[T]) T {
	var max_val T
	max_count := 0
	counts := NewDict[T, int](10)
	for i := 0; i < arr.Length(); i++ {
		val := arr[i]
		count := counts[val]
		count += 1
		if count > max_count {
			max_count = count
			max_val = val
		}
		counts[val] = count
	}
	return max_val
}

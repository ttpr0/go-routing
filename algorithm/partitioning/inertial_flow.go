package partitioning

import (
	"fmt"
	"sort"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func SortNodes(g graph.IGraph, coord_func func(geo.Coord) float32) Array[int32] {
	nodes := NewArray[int32](int(g.NodeCount()))
	for i := 0; i < int(g.NodeCount()); i++ {
		nodes[i] = int32(i)
	}

	geom := g.GetGeometry()

	sort.Slice(nodes, func(i, j int) bool {
		coord_i := geom.GetNode(int32(i))
		val_i := coord_func(coord_i)
		coord_j := geom.GetNode(int32(j))
		val_j := coord_func(coord_j)
		return val_i > val_j
	})

	return nodes
}

func InertialFlow(g graph.IGraph) Array[int16] {
	order := SortNodes(g, func(coord geo.Coord) float32 {
		return coord[0]
	})
	k := order.Length() / 4

	alg := NewEdmondsKarp(g, order[:k], 1, order[3*k:], 2, order[k:3*k], 0)
	flow := alg.ComputeMaxFlow()
	fmt.Println("computed flow", flow)
	alg.ComputeMinCut()

	return alg.GetNodeTiles()
}

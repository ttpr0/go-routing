package partitioning

import (
	"fmt"
	"math"
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
		node_a := nodes[i]
		coord_a := geom.GetNode(node_a)
		val_a := coord_func(coord_a)
		node_b := nodes[j]
		coord_b := geom.GetNode(node_b)
		val_b := coord_func(coord_b)
		return val_a < val_b
	})

	return nodes
}

func CreateOrders(g graph.IGraph) List[Array[int32]] {
	sin_45 := float32(math.Sin(math.Pi / 4))
	cos_45 := float32(math.Cos(math.Pi / 4))
	factors := []Tuple[float32, float32]{
		// vertical
		{1, 0},
		{-1, 0},
		// horizontal
		{0, 1},
		{0, -1},
		// diagonal ll to ur
		{-sin_45, cos_45},
		{sin_45, -cos_45},
		// diagonal ul to lr
		{sin_45, cos_45},
		{-sin_45, -cos_45},
	}
	orders := NewList[Array[int32]](8)
	for _, factor := range factors {
		orders.Add(SortNodes(g, func(coord geo.Coord) float32 {
			return factor.A*coord[0] + factor.B*coord[1]
		}))
	}
	return orders
}

func InertialFlow(g graph.IGraph) Array[int16] {
	// compute node orderings by location embedding
	orders := CreateOrders(g)

	// create source and sink thresholds
	k := 0.25
	so_c := int(float64(g.NodeCount()) * k)
	si_c := int(float64(g.NodeCount()) * (1 - k))

	var max_alg *EdmondsKarp
	max_flow := -1
	fmt.Println("start computing flows")
	for _, order := range orders {
		alg := NewEdmondsKarp(g, order[:so_c], 1, order[si_c:], 2, order[so_c:si_c], 0)
		flow := alg.ComputeMaxFlow()
		fmt.Println("computed flow:", flow)
		if flow > max_flow {
			max_flow = flow
			max_alg = alg
		}
	}
	if max_alg == nil {
		panic("no min cut found")
	}
	fmt.Println("start computing min cut")
	max_alg.ComputeMinCut()

	fmt.Println("inertial flow finished")
	return max_alg.GetNodeTiles()
}

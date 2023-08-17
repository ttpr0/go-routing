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

	// create node_tiles array
	node_tiles := NewArray[int16](int(g.NodeCount()))

	// init processing queue containing to be splitted tiles
	proc_queue := NewQueue[int16]()
	proc_queue.Push(0)

	// keep track of tile ids
	max_tile := int16(0)

	// iterate until no more tiles to be processed
	for proc_queue.Size() > 0 {
		curr_tile, _ := proc_queue.Pop()
		source_tile := max_tile + 1
		sink_tile := max_tile + 2

		// dont process if node count is small enough
		node_count := 0
		for i := 0; i < node_tiles.Length(); i++ {
			if node_tiles[i] == curr_tile {
				node_count += 1
			}
		}
		if node_count < 10000 {
			continue
		}

		// compute max-flow for every direction
		var max_alg *EdmondsKarp
		max_flow := -1
		fmt.Println("start computing flows")
		for _, order := range orders {
			// select nodes from current tile
			nodes := NewList[int32](node_count)
			for _, node := range order {
				if node_tiles[node] == curr_tile {
					nodes.Add(node)
				}
			}
			// create source and sink thresholds
			k := 0.25
			so_c := int(float64(nodes.Length()) * k)
			si_c := int(float64(nodes.Length()) * (1 - k))

			// compute max-flow for current direction
			alg := NewEdmondsKarp(g, nodes[:so_c], source_tile, nodes[si_c:], sink_tile, nodes[so_c:si_c], curr_tile)
			flow := alg.ComputeMaxFlow()
			fmt.Println("computed flow:", flow)

			// select minimum max-flow
			if flow < max_flow || max_flow == -1 {
				max_flow = flow
				max_alg = alg
			}
		}
		// compute min-cut on minimum max-flow
		fmt.Println("start computing min cut")
		max_alg.ComputeMinCut()

		// set computed tiles
		tiles := max_alg.GetNodeTiles()
		for i := 0; i < node_tiles.Length(); i++ {
			if node_tiles[i] != curr_tile {
				continue
			}
			node_tiles[i] = tiles[i]
		}
		max_tile += 2

		// add new tiles to processing queue
		proc_queue.Push(source_tile)
		proc_queue.Push(sink_tile)
	}

	fmt.Println("inertial flow finished")
	return node_tiles
}

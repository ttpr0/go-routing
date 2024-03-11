package main

import (
	"github.com/ttpr0/go-routing/algorithm"
	"github.com/ttpr0/go-routing/algorithm/partitioning"
	"github.com/ttpr0/go-routing/comps"
	"github.com/ttpr0/go-routing/graph"
	"github.com/ttpr0/go-routing/preproc"
	. "github.com/ttpr0/go-routing/util"
)

func CreatePartition(base *comps.GraphBase, nodes_per_cell int) *comps.Partition {
	eq_weight := comps.NewEqualWeighting()
	g := graph.BuildGraph(base, eq_weight)
	tiles := partitioning.InertialFlow(g)
	return comps.NewPartition(tiles)
}

func CreateGRASP(base *comps.GraphBase, weight comps.IWeighting, partition *comps.Partition, skeleton bool) (*comps.GraphBase, *comps.Partition, *comps.Overlay, *comps.CellIndex, Array[int32]) {
	g := graph.BuildGraph(base, weight)
	var overlay *comps.Overlay
	if skeleton {
		overlay = preproc.PrepareSkeletonOverlay(g, partition)
	} else {
		overlay = preproc.PrepareOverlay(g, partition)
	}

	mapping := preproc.ComputeTileOrdering(g, partition)
	new_base := comps.ReorderNodes(base, mapping)
	new_overlay := comps.ReorderNodes(overlay, mapping)
	new_partition := comps.ReorderNodes(partition, mapping)

	g = graph.BuildGraph(new_base, weight)
	cell_index := preproc.PrepareGRASPCellIndex(g, partition)

	return new_base, new_partition, new_overlay, cell_index, mapping
}

func CreateIsoPHAST(base *comps.GraphBase, weight comps.IWeighting, partition *comps.Partition) (*comps.GraphBase, *comps.Partition, *comps.Overlay, *comps.CellIndex, Array[int32]) {
	overlay, cell_index := preproc.PrepareIsoPHAST(base, weight, partition)

	g := graph.BuildGraph(base, weight)
	mapping := preproc.ComputeTileOrdering(g, partition)
	new_base := comps.ReorderNodes(base, mapping)
	new_overlay := comps.ReorderNodes(overlay, mapping)
	new_partition := comps.ReorderNodes(partition, mapping)
	new_cell_index := comps.ReorderNodes(cell_index, mapping)

	return new_base, new_partition, new_overlay, new_cell_index, mapping
}

func CreateCH(base *comps.GraphBase, weight comps.IWeighting) (*comps.GraphBase, *comps.CH, Array[int32]) {
	ch := preproc.CalcContraction6(base, weight)

	g := graph.BuildGraph(base, weight)
	ordering := preproc.ComputeLevelOrdering(g, ch)
	new_base := comps.ReorderNodes(base, ordering)
	new_ch := comps.ReorderNodes(ch, ordering)

	return new_base, new_ch, ordering
}

func CreateTiledCH(base *comps.GraphBase, weight comps.IWeighting, partition *comps.Partition) (*comps.GraphBase, *comps.Partition, *comps.CH, Array[int32]) {
	ch := preproc.CalcContraction5(base, weight, partition)

	g := graph.BuildGraph(base, weight)
	ordering := preproc.ComputeTileLevelOrdering(g, partition, ch)
	new_base := comps.ReorderNodes(base, ordering)
	new_ch := comps.ReorderNodes(ch, ordering)
	new_partition := comps.ReorderNodes(partition, ordering)

	return new_base, new_partition, new_ch, ordering
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

func RemoveConnectedComponents(base *comps.GraphBase) (List[int32], List[int32]) {
	eq_weight := comps.NewEqualWeighting()
	g := graph.BuildGraph(base, eq_weight)
	groups := algorithm.ConnectedComponents(g)
	max_group := GetMostCommon(groups)
	// get nodes and edges to be removed
	e := g.GetGraphExplorer()
	nodes_remove := NewArray[bool](g.NodeCount())
	edges_remove := NewArray[bool](g.EdgeCount())
	for i := 0; i < g.NodeCount(); i++ {
		if groups[i] == max_group {
			continue
		}
		// mark nodes not part of largest group
		nodes_remove[i] = true
		// mark all in- and out-going edges of those nodes
		e.ForAdjacentEdges(int32(i), graph.FORWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			edges_remove[ref.EdgeID] = true
		})
		e.ForAdjacentEdges(int32(i), graph.BACKWARD, graph.ADJACENT_ALL, func(ref graph.EdgeRef) {
			edges_remove[ref.EdgeID] = true
		})
	}

	// get remove lists
	remove_nodes := NewList[int32](100)
	remove_edges := NewList[int32](100)
	for i := 0; i < g.NodeCount(); i++ {
		if nodes_remove[i] {
			remove_nodes.Add(int32(i))
		}
	}
	for i := 0; i < g.EdgeCount(); i++ {
		if edges_remove[i] {
			remove_edges.Add(int32(i))
		}
	}

	return remove_nodes, remove_edges
}

package graph

import (
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func CreateCHGraph3(g *CHGraph) *CHGraph3 {
	fwd_down_edges := NewList[CHEdge](g.NodeCount())
	bwd_down_edges := NewList[CHEdge](g.NodeCount())

	explorer := g.GetDefaultExplorer()
	for i := 0; i < g.NodeCount(); i++ {
		this_id := int32(i)
		count := 0
		explorer.ForAdjacentEdges(this_id, FORWARD, ADJACENT_DOWNWARDS, func(ref EdgeRef) {
			other_id := ref.OtherID
			fwd_down_edges.Add(CHEdge{
				From:   this_id,
				To:     other_id,
				Weight: explorer.GetEdgeWeight(ref),
			})
			count += 1
		})
		for j := fwd_down_edges.Length() - count; j < fwd_down_edges.Length(); j++ {
			ch_edge := fwd_down_edges[j]
			ch_edge.Count = int32(count)
			fwd_down_edges[j] = ch_edge
		}

		count = 0
		explorer.ForAdjacentEdges(this_id, BACKWARD, ADJACENT_DOWNWARDS, func(ref EdgeRef) {
			other_id := ref.OtherID
			bwd_down_edges.Add(CHEdge{
				From:   this_id,
				To:     other_id,
				Weight: explorer.GetEdgeWeight(ref),
			})
			count += 1
		})
		for j := bwd_down_edges.Length() - count; j < bwd_down_edges.Length(); j++ {
			ch_edge := bwd_down_edges[j]
			ch_edge.Count = int32(count)
			bwd_down_edges[j] = ch_edge
		}
	}

	return &CHGraph3{
		CHGraph: *g,

		fwd_down_edges: Array[CHEdge](fwd_down_edges),
		bwd_down_edges: Array[CHEdge](bwd_down_edges),
	}
}

type CHGraph3 struct {
	CHGraph

	// stores all fowwards-down edges
	fwd_down_edges Array[CHEdge]
	// stores all backwards-down edges
	bwd_down_edges Array[CHEdge]
}

func (self *CHGraph3) GetDownEdges(dir Direction) Array[CHEdge] {
	if dir == FORWARD {
		return self.fwd_down_edges
	} else {
		return self.bwd_down_edges
	}
}

type CHEdge struct {
	From   int32
	To     int32
	Weight int32
	Count  int32
}

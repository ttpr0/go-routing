package preproc

import (
	"github.com/ttpr0/go-routing/comps"
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
)

func _IsBorderNode3(g graph.IGraph, partition *comps.Partition) Array[bool] {
	is_border := NewArray[bool](g.NodeCount())

	explorer := g.GetGraphExplorer()
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

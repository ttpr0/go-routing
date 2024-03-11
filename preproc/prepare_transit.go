package preproc

import (
	"github.com/ttpr0/go-routing/batched/onetomany"
	"github.com/ttpr0/go-routing/comps"
	"github.com/ttpr0/go-routing/graph"
	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// prepare transit-data
//*******************************************

func PrepareTransit(g graph.IGraph, stops Array[structs.Node], connections Array[structs.Connection], max_transfer_range int32) *comps.Transit {
	mapping := NewArray[[2]int32](g.NodeCount())
	for i := 0; i < g.NodeCount(); i++ {
		mapping[i] = [2]int32{-1, -1}
	}
	tree := NewKDTree[int32](2)
	for i := 0; i < g.NodeCount(); i++ {
		coord := g.GetNodeGeom(int32(i))
		tree.Insert(coord[:], int32(i))
	}
	for i := 0; i < stops.Length(); i++ {
		stop := stops[i]
		closest, ok := tree.GetClosest(stop.Loc[:], 0.05)
		if !ok {
			continue
		}
		mapping[closest][0] = int32(i)
		mapping[i][1] = closest
	}
	id_mapping := structs.NewIDMapping(mapping)
	shortcuts := structs.NewShortcutStore(100, false)
	otm := onetomany.NewRangeDijkstraTC(g, max_transfer_range)
	solver := otm.CreateSolver()
	for i := 0; i < stops.Length(); i++ {
		s_node := id_mapping.GetSource(int32(i))
		if s_node == -1 {
			continue
		}
		solver.CalcDistanceFromStart(s_node)
		for j := 0; j < stops.Length(); j++ {
			if i == j {
				continue
			}
			t_node := id_mapping.GetSource(int32(j))
			if t_node == -1 {
				continue
			}
			dist := solver.GetDistance(t_node)
			if dist > max_transfer_range {
				continue
			}
			// TODO: edges from shortcuts
			shortcuts.AddShortcut(structs.Shortcut{From: int32(i), To: int32(j), Weight: dist}, []int32{})
		}
	}

	return comps.NewTransit(id_mapping, stops, connections, shortcuts)
}

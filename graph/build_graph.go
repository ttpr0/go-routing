package graph

import (
	"github.com/ttpr0/go-routing/comps"
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// build graphs
//*******************************************

func BuildGraph(base comps.IGraphBase, weight comps.IWeighting, index Optional[comps.IGraphIndex]) *Graph {
	return &Graph{
		base:   base,
		weight: weight,
		index:  index,
	}
}

func BuildTCGraph(base comps.IGraphBase, weight comps.ITCWeighting, index Optional[comps.IGraphIndex]) *TCGraph {
	return &TCGraph{
		base:   base,
		weight: weight,
		index:  index,
	}
}

func BuildCHGraph(base comps.IGraphBase, weight comps.IWeighting, index Optional[comps.IGraphIndex], ch_data *comps.CH, ch_index Optional[*comps.CHIndex]) *CHGraph {
	return &CHGraph{
		base:   base,
		weight: weight,
		index:  index,

		ch:        ch_data,
		partition: None[*comps.Partition](),
		ch_index:  ch_index,
	}
}

func BuildPartitionedCHGraph(base comps.IGraphBase, weight comps.IWeighting, index Optional[comps.IGraphIndex], ch_data *comps.CH, partition Optional[*comps.Partition], ch_index Optional[*comps.CHIndex]) *CHGraph {
	return &CHGraph{
		base:   base,
		weight: weight,
		index:  index,

		ch:        ch_data,
		partition: partition,
		ch_index:  ch_index,
	}
}

func BuildTiledGraph(base comps.IGraphBase, weight comps.IWeighting, index Optional[comps.IGraphIndex], partition *comps.Partition, overlay *comps.Overlay, cell_index Optional[*comps.CellIndex]) *TiledGraph {
	return &TiledGraph{
		base:   base,
		weight: weight,
		index:  index,

		partition:  partition,
		overlay:    overlay,
		cell_index: cell_index,
	}
}

func BuildTransitGraph(base comps.IGraphBase, weight comps.ITCWeighting, index Optional[comps.IGraphIndex], transit *comps.Transit, transit_weight *comps.TransitWeighting) *TransitGraph {
	return &TransitGraph{
		base:   base,
		index:  index,
		weight: weight,

		transit:        transit,
		transit_weight: transit_weight,
	}
}

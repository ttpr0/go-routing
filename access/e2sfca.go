package access

import (
	"github.com/ttpr0/go-routing/access/decay"
	"github.com/ttpr0/go-routing/access/view"
	"github.com/ttpr0/go-routing/algorithm"
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
)

/**
 * Computes the enhanced two-step-floating-catchment-area accessibility
 * introduced by Luo and Qi (2009).
 * Formula:
 * $A_i = \sum_j{\frac{S_j}{\sum_i{D_i * w_{ij}}} * w_{ij}}$
 * $S_j$ denotes the weight of the reachable supply $j$, $D_i$ the demand of the
 * demand point $i$ and $w_{ij} ~ d_{ij}$ the travel-friction (distance decay)
 * between them.
 *
 * @param g        Graph-Instance.
 * @param demand   Demand locations and weights ($D_i$).
 * @param supply   Supply locations and weights ($S_j$).
 * @param decay    Distance decay.
 * @return enhanced two-step-floating-catchment-area value for every demand
 *         point.
 */
func CalcEnhanced2SFCA(g graph.IGraph, dem view.IPointView, sup view.IPointView, dec decay.IDistanceDecay) []float32 {
	populationWeights := NewArray[float32](dem.PointCount())
	facilityWeights := NewArray[float32](sup.PointCount())

	max_range := dec.GetMaxDistance()
	invertedMapping := NewDict[int, List[Tuple[int, float32]]](10)

	index := g.GetIndex()
	demand_nodes := NewArray[int32](dem.PointCount())
	for i := 0; i < dem.PointCount(); i++ {
		loc := dem.GetCoordinate(i)
		id, ok := index.GetClosestNode(loc)
		if ok {
			demand_nodes[i] = id
		} else {
			demand_nodes[i] = -1
		}
	}
	supply_nodes := NewArray[int32](sup.PointCount())
	for i := 0; i < sup.PointCount(); i++ {
		loc := sup.GetCoordinate(i)
		id, ok := index.GetClosestNode(loc)
		if ok {
			supply_nodes[i] = id
		} else {
			supply_nodes[i] = -1
		}
	}
	matrix := algorithm.DijkstraTDMatrix(g, supply_nodes, demand_nodes, max_range)

	for f := 0; f < sup.PointCount(); f++ {
		weight := float32(0)
		for p := 0; p < dem.PointCount(); p++ {
			dist := matrix.Get(f, p)
			if dist < 0 {
				continue
			}
			rangeFactor := dec.GetDistanceWeight(dist)
			populationCount := dem.GetWeight(p)
			weight += float32(populationCount) * rangeFactor

			var refs List[Tuple[int, float32]]
			if !invertedMapping.ContainsKey(p) {
				refs = NewList[Tuple[int, float32]](4)
			} else {
				refs = invertedMapping.Get(p)
			}
			refs.Add(MakeTuple(f, dist))
			invertedMapping.Set(p, refs)
		}
		if weight == 0 {
			facilityWeights[f] = 0
		} else {
			facilityWeights[f] = float32(sup.GetWeight(f)) / weight
		}
	}

	for index, refs := range invertedMapping {
		if refs == nil {
			continue
		} else {
			weight := float32(0)
			for _, fref := range refs {
				rangeFactor := dec.GetDistanceWeight(fref.B)
				weight += facilityWeights[fref.A] * rangeFactor
			}
			populationWeights[index] = weight
		}
	}

	return populationWeights
}

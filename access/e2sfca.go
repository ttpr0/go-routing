package access

import (
	"sync"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/access/decay"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/access/provider"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/access/view"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/routing"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func CalcEnhanced2SFCA(g graph.IGraph, supply_locs, demand_locs Array[geo.Coord], supply_weights, demand_weights Array[int32], max_range float32) []float32 {
	index := g.GetIndex()
	population_nodes := NewArray[int32](len(demand_locs))
	for i, loc := range demand_locs {
		id, ok := index.GetClosestNode(loc)
		if ok {
			population_nodes[i] = id
		} else {
			population_nodes[i] = -1
		}
	}
	facility_chan := make(chan Tuple[geo.Coord, float32], len(supply_locs))
	for i, facility := range supply_locs {
		facility_chan <- MakeTuple(facility, float32(supply_weights[i]))
	}

	access := NewArray[float32](len(demand_locs))
	wg := sync.WaitGroup{}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			spt := routing.NewSPT2(g)
			for {
				if len(facility_chan) == 0 {
					break
				}
				temp := <-facility_chan
				facility := temp.A
				weight := temp.B
				id, ok := index.GetClosestNode(facility)
				if !ok {
					continue
				}
				spt.Init(id, float64(max_range))
				spt.CalcSPT()
				flags := spt.GetSPT()

				facility_weight := float32(0.0)
				for i, node := range population_nodes {
					if node == -1 {
						continue
					}
					flag := flags[node]
					if !flag.Visited {
						continue
					}
					distance_decay := float32(1 - flag.PathLength/float64(max_range))
					facility_weight += float32(demand_weights[i]) * distance_decay
				}
				for i, node := range population_nodes {
					if node == -1 {
						continue
					}
					flag := flags[node]
					if !flag.Visited {
						continue
					}
					distance_decay := float32(1 - flag.PathLength/float64(max_range))
					access[i] += (weight / facility_weight) * distance_decay
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	max_val := float32(0.0)
	for _, val := range access {
		if val > max_val {
			max_val = val
		}
	}
	for i, val := range access {
		if val == 0 {
			access[i] = -9999
		} else {
			access[i] = val * 100 / max_val
		}
	}

	return access
}

/**
 * Computes the enhanced two-step-floating-catchment-area accessibility
 * introduced by Luo and Qi (2009).
 * Formula:
 * $A_i = \sum_j{\frac{S_j}{\sum_i{D_i * w_{ij}}} * w_{ij}}$
 * $S_j$ denotes the weight of the reachable supply $j$, $D_i$ the demand of the
 * demand point $i$ and $w_{ij} ~ d_{ij}$ the travel-friction (distance decay)
 * between them.
 *
 * @param demand   Demand locations and weights ($D_i$).
 * @param supply   Supply locations and weights ($S_j$).
 * @param decay    Distance decay.
 * @param provider Routing API provider.
 * @param options  Computation mode ("isochrones", "matrix") and Ranges of
 *                 isochrones used in computation of distances $d_{ij}$.
 * @return enhanced two-step-floating-catchment-area value for every demand
 *         point.
 */
func CalcEnhanced2SFCA2(dem view.IPointView, sup view.IPointView, dec decay.IDistanceDecay, prov provider.IRoutingProvider, options provider.RoutingOptions) []float32 {
	populationWeights := NewArray[float32](dem.PointCount())
	facilityWeights := NewArray[float32](sup.PointCount())

	invertedMapping := NewDict[int, List[Tuple[int, float32]]](10)

	matrix := prov.RequestTDMatrix(dem, sup, options)
	if matrix == nil {
		return populationWeights
	}
	for f := 0; f < sup.PointCount(); f++ {
		weight := float32(0)
		for p := 0; p < dem.PointCount(); p++ {
			dist := matrix.GetRange(f, p)
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

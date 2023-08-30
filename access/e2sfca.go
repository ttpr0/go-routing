package access

import (
	"sync"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/access/decay"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/access/view"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/algorithm"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/routing"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func CalcEnhanced2SFCA(g graph.IGraph, dem view.IPointView, sup view.IPointView, dec decay.IDistanceDecay) []float32 {
	index := g.GetIndex()
	population_nodes := NewArray[int32](dem.PointCount())
	for i := 0; i < dem.PointCount(); i++ {
		loc := dem.GetCoordinate(i)
		id, ok := index.GetClosestNode(loc)
		if ok {
			population_nodes[i] = id
		} else {
			population_nodes[i] = -1
		}
	}
	facility_chan := make(chan Tuple[geo.Coord, float32], sup.PointCount())
	for i := 0; i < sup.PointCount(); i++ {
		loc := sup.GetCoordinate(i)
		w := sup.GetWeight(i)
		facility_chan <- MakeTuple(loc, float32(w))
	}

	max_range := dec.GetMaxDistance()

	access := NewArray[float32](dem.PointCount())
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
					distance_decay := dec.GetDistanceWeight(float32(flag.PathLength))
					facility_weight += float32(dem.GetWeight(i)) * distance_decay
				}
				for i, node := range population_nodes {
					if node == -1 {
						continue
					}
					flag := flags[node]
					if !flag.Visited {
						continue
					}
					distance_decay := dec.GetDistanceWeight(float32(flag.PathLength))
					access[i] += (weight / facility_weight) * distance_decay
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

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
 * @param g        Graph-Instance.
 * @param demand   Demand locations and weights ($D_i$).
 * @param supply   Supply locations and weights ($S_j$).
 * @param decay    Distance decay.
 * @return enhanced two-step-floating-catchment-area value for every demand
 *         point.
 */
func CalcEnhanced2SFCA2(g graph.IGraph, dem view.IPointView, sup view.IPointView, dec decay.IDistanceDecay) []float32 {
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

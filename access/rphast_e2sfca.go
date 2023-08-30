package access

import (
	"sync"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/access/decay"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/access/view"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/algorithm"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func CalcRPHASTEnhanced2SFCA(g graph.ICHGraph, dem view.IPointView, sup view.IPointView, dec decay.IDistanceDecay) []float32 {
	node_queue := NewQueue[int32]()
	index := g.GetIndex()
	population_nodes := NewArray[int32](dem.PointCount())
	for i := 0; i < dem.PointCount(); i++ {
		loc := dem.GetCoordinate(i)
		id, ok := index.GetClosestNode(loc)
		if ok {
			population_nodes[i] = id
			node_queue.Push(id)
		} else {
			population_nodes[i] = -1
		}
	}
	facility_chan := make(chan Tuple[geo.Coord, float32], sup.PointCount())
	for i := 0; i < sup.PointCount(); i++ {
		loc := sup.GetCoordinate(i)
		weight := sup.GetWeight(i)
		facility_chan <- MakeTuple(loc, float32(weight))
	}

	explorer := g.GetDefaultExplorer()
	graph_subset := NewArray[bool](int(g.NodeCount()))
	for {
		if node_queue.Size() == 0 {
			break
		}
		node, _ := node_queue.Pop()
		if graph_subset[node] {
			continue
		}
		graph_subset[node] = true
		node_level := g.GetNodeLevel(node)
		edges := explorer.GetAdjacentEdges(node, graph.BACKWARD, graph.ADJACENT_ALL)
		for {
			ref, ok := edges.Next()
			if !ok {
				break
			}
			if graph_subset[ref.OtherID] {
				continue
			}
			if node_level >= g.GetNodeLevel(ref.OtherID) {
				continue
			}
			node_queue.Push(ref.OtherID)
		}
	}

	max_range := dec.GetMaxDistance()

	access := NewArray[float32](dem.PointCount())
	wg := sync.WaitGroup{}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			spt := algorithm.NewRPHAST(g, graph_subset)
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
					if flag.PathLength > float64(max_range) {
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
					if flag.PathLength > float64(max_range) {
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

package access

import (
	"sync"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/algorithm"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func CalcRPHASTEnhanced2SFCA(g graph.ICHGraph, supply_locs, demand_locs Array[geo.Coord], supply_weights, demand_weights Array[int32], max_range float32) []float32 {
	node_queue := NewQueue[int32]()
	index := g.GetIndex()
	population_nodes := NewArray[int32](len(demand_locs))
	for i, loc := range demand_locs {
		id, ok := index.GetClosestNode(loc)
		if ok {
			population_nodes[i] = id
			node_queue.Push(id)
		} else {
			population_nodes[i] = -1
		}
	}
	facility_chan := make(chan Tuple[geo.Coord, float32], len(supply_locs))
	for i, facility := range supply_locs {
		facility_chan <- MakeTuple(facility, float32(supply_weights[i]))
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

	access := NewArray[float32](len(demand_locs))
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
					distance_decay := float32(1 - flag.PathLength/float64(max_range))
					facility_weight += float32(demand_weights[i]) * distance_decay
				}
				for i, node := range population_nodes {
					if node == -1 {
						continue
					}
					flag := flags[node]
					if flag.PathLength > float64(max_range) {
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

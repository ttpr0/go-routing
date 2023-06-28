package access

import (
	"sync"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/routing"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func CalcTiledEnhanced2SFCA(g graph.ITiledGraph2, supply_locs, demand_locs [][2]float32, supply_weights, demand_weights []float32, max_range float32) []float32 {
	population_nodes := NewArray[int32](len(demand_locs))
	node_flag := NewArray[bool](int(g.NodeCount()))
	node_tiles := NewDict[int16, bool](10)
	for i, loc := range demand_locs {
		id, ok := g.GetClosestNode(loc)
		if ok {
			population_nodes[i] = id
			node_flag[id] = true
			tile := g.GetNodeTile(id)
			node_tiles[tile] = true
		} else {
			population_nodes[i] = -1
		}
	}
	facility_chan := make(chan Tuple[[2]float32, float32], len(supply_locs))
	for i, facility := range supply_locs {
		facility_chan <- MakeTuple(facility, supply_weights[i])
	}

	access := NewArray[float32](len(demand_locs))
	wg := sync.WaitGroup{}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			spt := routing.NewSPT3(g)
			for {
				if len(facility_chan) == 0 {
					break
				}
				temp := <-facility_chan
				facility := temp.A
				weight := temp.B
				id, ok := g.GetClosestNode(facility)
				if !ok {
					continue
				}
				spt.Init(id, float64(max_range))
				spt.CalcSPT()
				flags := spt.GetSPT()
				active_tiles := spt.GetActiveTiles()

				for tile, _ := range active_tiles {
					if !node_tiles.ContainsKey(tile) {
						continue
					}
					for _, border_node := range g.GetBorderNodes(tile) {
						b_flag := flags[border_node]
						if !b_flag.Visited {
							continue
						}
						b_range := float32(b_flag.PathLength)
						iter := g.GetTileRanges(tile, border_node)
						for {
							item, ok := iter.Next()
							if !ok {
								break
							}
							node := item.A
							if !node_flag[node] {
								continue
							}
							dist := item.B
							if dist == 1000000 {
								continue
							}
							flag := flags[node]
							if flag.PathLength > float64(b_range+dist) {
								flag.PathLength = float64(b_range + dist)
							}
							flag.Visited = true
							flags[node] = flag
						}
					}
				}
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
					facility_weight += demand_weights[i] * distance_decay
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

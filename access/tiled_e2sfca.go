package access

import (
	"sync"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/access/decay"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/access/view"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/routing"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func CalcTiled2SFCA(g graph.ITiledGraph2, dem view.IPointView, sup view.IPointView, dec decay.IDistanceDecay) []float32 {
	index := g.GetIndex()
	population_nodes := NewArray[int32](dem.PointCount())
	node_flag := NewArray[bool](int(g.NodeCount()))
	node_tiles := NewDict[int16, bool](10)
	for i := 0; i < dem.PointCount(); i++ {
		loc := dem.GetCoordinate(i)
		id, ok := index.GetClosestNode(loc)
		if ok {
			population_nodes[i] = id
			node_flag[id] = true
			tile := g.GetNodeTile(id)
			node_tiles[tile] = true
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

	max_range := dec.GetMaxDistance()

	access := NewArray[float32](dem.PointCount())
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
				id, ok := index.GetClosestNode(facility)
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

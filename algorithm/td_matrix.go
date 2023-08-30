package algorithm

import (
	"sync"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/routing"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func DijkstraTDMatrix(g graph.IGraph, sources Array[int32], destinations Array[int32], max_range float32) Matrix[float32] {
	facility_chan := make(chan Tuple[int, int32], sources.Length())
	for i := 0; i < sources.Length(); i++ {
		facility_chan <- MakeTuple(i, sources[i])
	}

	matrix := NewMatrix[float32](sources.Length(), destinations.Length())
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
				f := temp.A
				f_node := temp.B
				if f_node == -1 {
					for i := 0; i < destinations.Length(); i++ {
						matrix.Set(f, i, -1)
					}
					continue
				}
				spt.Init(f_node, float64(max_range))
				spt.CalcSPT()
				flags := spt.GetSPT()

				for p, p_node := range destinations {
					if p_node == -1 {
						matrix.Set(f, p, -1)
						continue
					}
					flag := flags[p_node]
					if !flag.Visited {
						matrix.Set(f, p, -1)
						continue
					}
					matrix.Set(f, p, float32(flag.PathLength))
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	return matrix
}

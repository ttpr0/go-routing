package provider

import (
	"sync"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/access/view"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/routing"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

var GRAPH graph.IGraph = graph.LoadGraph("./graphs/niedersachsen")

type IRoutingProvider interface {
	SetProfile(profile string)

	SetRangeType(range_type string)

	SetParameter(name string, value any)

	RequestTDMatrix(dem view.IPointView, sup view.IPointView, options RoutingOptions) ITDMatrix

	RequestNearest(dem view.IPointView, sup view.IPointView, options RoutingOptions) INNTable

	RequestKNearest(dem view.IPointView, sup view.IPointView, k int, options RoutingOptions) IKNNTable

	RequestCatchment(dem view.IPointView, sup view.IPointView, dist float32, options RoutingOptions) ICatchment
}

type RoutingOptions struct {
	Mode     string
	MaxRange float32
}

func (self *RoutingOptions) GetMode() string {
	if self.Mode == "" {
		return "dijkstra"
	}
	return self.Mode
}

func (self *RoutingOptions) GetMaxRange() float32 {
	return self.MaxRange
}

type RoutingProvider struct {
}

func (self *RoutingProvider) SetProfile(profile string) {
}
func (self *RoutingProvider) SetRangeType(range_type string) {
}
func (self *RoutingProvider) SetParameter(name string, value any) {
}

func (self *RoutingProvider) RequestTDMatrix(dem view.IPointView, sup view.IPointView, options RoutingOptions) ITDMatrix {
	index := GRAPH.GetIndex()
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
	facility_chan := make(chan Tuple[int, geo.Coord], sup.PointCount())
	for i := 0; i < sup.PointCount(); i++ {
		facility_chan <- MakeTuple(i, sup.GetCoordinate(i))
	}
	max_range := options.GetMaxRange()

	matrix := NewMatrix[float32](sup.PointCount(), dem.PointCount())
	wg := sync.WaitGroup{}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			spt := routing.NewSPT2(GRAPH)
			for {
				if len(facility_chan) == 0 {
					break
				}
				temp := <-facility_chan
				f := temp.A
				facility := temp.B
				id, ok := index.GetClosestNode(facility)
				if !ok {
					continue
				}
				spt.Init(id, float64(max_range))
				spt.CalcSPT()
				flags := spt.GetSPT()

				for p, node := range population_nodes {
					if node == -1 {
						matrix.Set(f, p, -1)
						continue
					}
					flag := flags[node]
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

	return NewTDMatrix(matrix)
}

func (self *RoutingProvider) RequestNearest(dem view.IPointView, sup view.IPointView, options RoutingOptions) INNTable {
	panic("not implemented") // TODO: Implement
}

func (self *RoutingProvider) RequestKNearest(dem view.IPointView, sup view.IPointView, k int, options RoutingOptions) IKNNTable {
	panic("not implemented") // TODO: Implement
}

func (self *RoutingProvider) RequestCatchment(dem view.IPointView, sup view.IPointView, dist float32, options RoutingOptions) ICatchment {
	panic("not implemented") // TODO: Implement
}

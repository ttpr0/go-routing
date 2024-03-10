package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/ttpr0/go-routing/batched/onetomany"
	"github.com/ttpr0/go-routing/geo"
	. "github.com/ttpr0/go-routing/util"
)

type MatrixRequest struct {
	Sources      Array[geo.Coord] `json:"sources"`
	Destinations Array[geo.Coord] `json:"destinations"`
	Profile      string           `json:"profile"`
	Metric       string           `json:"metric"`
	MaxRange     float32          `json:"max_range"`
	TimeWindow   [2]int32         `json:"time_window"`
}

type MatrixResponse struct {
	Distances Matrix[float32] `json:"distances"`
}

func HandleMatrixRequest(w http.ResponseWriter, r *http.Request) {
	req := ReadRequestBody[MatrixRequest](r)
	var max_range float32
	if req.MaxRange > 0 {
		max_range = req.MaxRange
	} else {
		max_range = 100000000
	}
	profile := MANAGER.GetMatchingProfile(DRIVING, CAR, FASTEST)
	if !profile.HasValue() {
		panic("Profile not found")
	}
	g_ := profile.Value.GetGraph()
	if !g_.HasValue() {
		panic("Graph not found")
	}
	g := g_.Value

	fmt.Println("Run Matrix Request")

	source_count := req.Sources.Length()
	source_chan := make(chan Tuple[int, int32], source_count)
	for i := 0; i < source_count; i++ {
		loc := req.Sources[i]
		id, ok := g.GetClosestNode(loc)
		if ok {
			source_chan <- MakeTuple(i, id)
		} else {
			source_chan <- MakeTuple(i, int32(-1))
		}
	}
	close(source_chan)
	target_count := req.Destinations.Length()
	target_nodes := NewArray[int32](target_count)
	for i := 0; i < req.Destinations.Length(); i++ {
		loc := req.Destinations[i]
		id, ok := g.GetClosestNode(loc)
		if ok {
			target_nodes[i] = id
		} else {
			target_nodes[i] = -1
		}
	}

	matrix := NewMatrix[float32](source_count, target_count)
	otm := onetomany.NewRangeDijkstra(g, int32(max_range))
	wg := sync.WaitGroup{}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			solver := otm.CreateSolver()
			for {
				// read supply entry from chan
				temp, ok := <-source_chan
				if !ok {
					break
				}
				s := temp.A
				s_node := temp.B
				// if no node set all distances to -1
				if s_node == -1 {
					for i := 0; i < target_nodes.Length(); i++ {
						matrix.Set(s, i, -1)
					}
					continue
				}

				solver.CalcDistanceFromStart(s_node)

				// set distances in matrix
				for t, t_node := range target_nodes {
					if t_node == -1 {
						matrix.Set(s, t, -1)
						continue
					}
					dist := solver.GetDistance(t_node)
					if dist > int32(max_range) {
						matrix.Set(s, t, -1)
						continue
					}
					matrix.Set(s, t, float32(dist))
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	resp := MatrixResponse{Distances: matrix}
	fmt.Println("reponse build")
	WriteResponse(w, resp, http.StatusOK)
}

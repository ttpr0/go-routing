package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/routing"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type FCARequest struct {
	PopulationLocations [][2]float32 `json:"population_locations"`
	PopulationWeights   []float32    `json:"population_weights"`
	FacilityLocations   [][2]float32 `json:"facility_locations"`
	MaxRange            float64      `json:"max_range"`
}

type FCAResponse struct {
	Access []float32 `json:"access"`
}

func HandleFCARequest(w http.ResponseWriter, r *http.Request) {
	data, _ := io.ReadAll(r.Body)
	req := FCARequest{}
	err := json.Unmarshal(data, &req)
	if err != nil {
		fmt.Println(err.Error())
	}

	node_index := GRAPH.GetNodeIndex()
	population_nodes := NewArray[int32](len(req.PopulationLocations))
	for i, loc := range req.PopulationLocations {
		id, ok := node_index.GetClosest(loc[:], 0.005)
		if ok {
			population_nodes[i] = id
		} else {
			population_nodes[i] = -1
		}
	}
	facility_chan := make(chan [2]float32, len(req.FacilityLocations))
	for _, facility := range req.FacilityLocations {
		facility_chan <- facility
	}

	access := NewArray[float32](len(req.PopulationLocations))
	wg := sync.WaitGroup{}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			spt := routing.NewSPT2(GRAPH)
			for {
				if len(facility_chan) == 0 {
					break
				}
				facility := <-facility_chan
				id, ok := node_index.GetClosest(facility[:], 0.005)
				if !ok {
					continue
				}
				spt.Init(id, req.MaxRange)
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
					distance_decay := float32(1 - flag.PathLength/req.MaxRange)
					facility_weight += req.PopulationWeights[i] * distance_decay
				}
				for i, node := range population_nodes {
					if node == -1 {
						continue
					}
					flag := flags[node]
					if !flag.Visited {
						continue
					}
					distance_decay := float32(1 - flag.PathLength/req.MaxRange)
					access[i] += (1 / facility_weight) * distance_decay
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
		access[i] = val * 100 / max_val
	}

	resp := FCAResponse{Access: access}
	data, _ = json.Marshal(resp)
	w.Write(data)
}

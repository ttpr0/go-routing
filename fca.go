package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/access"
)

type FCARequest struct {
	PopulationLocations [][2]float32 `json:"population_locations"`
	PopulationDemand    []float32    `json:"population_weights"`
	FacilityLocations   [][2]float32 `json:"facility_locations"`
	FacilityCapacities  []float32    `json:"facility_capacities"`
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

	res := access.CalcEnhanced2SFCA(GRAPH, req.FacilityLocations, req.PopulationLocations, req.FacilityCapacities, req.PopulationDemand, float32(req.MaxRange))

	resp := FCAResponse{Access: res}
	data, _ = json.Marshal(resp)
	w.Write(data)
}

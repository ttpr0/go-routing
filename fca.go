package main

import (
	"fmt"
	"net/http"

	"github.com/ttpr0/go-routing/access"
)

type FCARequest struct {
	Demand        DemandRequestParams  `json:"demand"`
	DistanceDecay DecayRequestParams   `json:"distance_decay"`
	Routing       RoutingRequestParams `json:"routing"`
	Supply        SupplyRequestParams  `json:"supply"`
	Mode          string               `json:"mode"`
}

type FCAResponse struct {
	Access []float32 `json:"access"`
}

func HandleFCARequest(w http.ResponseWriter, r *http.Request) {
	req := ReadRequestBody[FCARequest](r)
	fmt.Println("Run FCA Request")

	demand_view := GetDemandView(req.Demand)
	if demand_view == nil {
		WriteResponse(w, NewErrorResponse("2sfca/enhanced", "failed to get demand-view, parameters are invalid"), http.StatusBadRequest)
		return
	}
	supply_view := GetSupplyView(req.Supply)
	if supply_view == nil {
		WriteResponse(w, NewErrorResponse("2sfca/enhanced", "failed to get supply-view, parameters are invalid"), http.StatusBadRequest)
		return
	}
	distance_decay := GetDistanceDecay(req.DistanceDecay)
	if distance_decay == nil {
		WriteResponse(w, NewErrorResponse("2sfca/enhanced", "failed to get distance-decay, parameters are invalid"), http.StatusBadRequest)
		return
	}

	var weights []float32
	if req.Mode == "tiled" {
		fmt.Println("run tiled fca")
		weights = access.CalcEnhanced2SFCA(GRAPH, demand_view, supply_view, distance_decay)
	} else if req.Mode == "ch" {
		fmt.Println("run ch fca")
		// weights = access.CalcRPHAST2SFCA(GRAPH, demand_view, supply_view, distance_decay)
	} else if req.Mode == "default" {
		fmt.Println("run default fca")
		weights = access.CalcEnhanced2SFCA(GRAPH, demand_view, supply_view, distance_decay)
	} else {
		weights = access.CalcEnhanced2SFCA(GRAPH, demand_view, supply_view, distance_decay)
	}

	max_weight := float32(0)
	for i := 0; i < len(weights); i++ {
		w := weights[i]
		if w > max_weight {
			max_weight = w
		}
	}
	factor := 100 / max_weight
	for i := 0; i < len(weights); i++ {
		w := weights[i]
		if w != 0 {
			w = w * factor
		} else {
			w = -9999
		}
		weights[i] = w
	}

	resp := FCAResponse{weights}
	fmt.Println("reponse build")
	WriteResponse(w, resp, http.StatusOK)
}

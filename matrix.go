package main

import (
	"fmt"
	"net/http"

	"github.com/ttpr0/go-routing/algorithm"
	"github.com/ttpr0/go-routing/geo"
	. "github.com/ttpr0/go-routing/util"
)

type MatrixRequest struct {
	Sources      Array[geo.Coord] `json:"sources"`
	Destinations Array[geo.Coord] `json:"destinations"`
	MaxRange     float32          `json:"max_range"`
}

type MatrixResponse struct {
	Durations Matrix[float32] `json:"durations"`
}

func HandleMatrixRequest(w http.ResponseWriter, r *http.Request) {
	req := ReadRequestBody[MatrixRequest](r)

	fmt.Println("Run Matrix Request")

	source_nodes := NewArray[int32](req.Sources.Length())
	for i := 0; i < req.Sources.Length(); i++ {
		loc := req.Sources[i]
		id, ok := GRAPH.GetClosestNode(loc)
		if ok {
			source_nodes[i] = id
		} else {
			source_nodes[i] = -1
		}
	}
	destination_nodes := NewArray[int32](req.Destinations.Length())
	for i := 0; i < req.Destinations.Length(); i++ {
		loc := req.Destinations[i]
		id, ok := GRAPH.GetClosestNode(loc)
		if ok {
			destination_nodes[i] = id
		} else {
			destination_nodes[i] = -1
		}
	}

	var max_range float32
	if req.MaxRange > 0 {
		max_range = req.MaxRange
	} else {
		max_range = 100000000
	}

	matrix := algorithm.DijkstraTDMatrix(GRAPH, source_nodes, destination_nodes, max_range)

	resp := MatrixResponse{Durations: matrix}
	fmt.Println("reponse build")
	WriteResponse(w, resp, http.StatusOK)
}

package main

import (
	"github.com/ttpr0/go-routing/isochrone"
)

//**********************************************************
// isochrone request
//**********************************************************

type IsochroneRequest struct {
	Locations [][]float32 `json:"locations"`
	Range     []int32     `json:"range"`
}

//**********************************************************
// isochrone handler
//**********************************************************

func HandleIsochroneRequest(req IsochroneRequest) Result {
	// get profile
	profile_, res := GetRequestProfile(MANAGER, "driving-car", "time")
	if !profile_.HasValue() {
		return res
	}
	profile := profile_.Value
	g_ := profile.GetGraph()
	if !g_.HasValue() {
		return BadRequest("Graph not found")
	}
	g := g_.Value
	att := profile.GetAttributes()
	loc := [2]float32{req.Locations[0][0], req.Locations[0][1]}
	resp := isochrone.ComputeIsochrone(g, att, loc, req.Range)
	return OK(&resp)
}

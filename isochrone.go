package main

import (
	"github.com/ttpr0/go-routing/isochrone"
	"github.com/ttpr0/go-routing/routing"
)

//**********************************************************
// isochrone request
//**********************************************************

type IsochroneRequest struct {
	Locations [][]float32 `json:"locations"`
	Range     []int32     `json:"range"`
	Profile   string      `json:"profile"`
	Metric    string      `json:"metric"`
}

//**********************************************************
// isochrone handler
//**********************************************************

func HandleIsochroneRequest(req IsochroneRequest) Result {
	// get profile
	loc := [2]float32{req.Locations[0][0], req.Locations[0][1]}
	if req.Profile == "" {
		req.Profile = "driving-car"
	}
	if req.Metric == "" {
		req.Metric = "time"
	}
	profile_, res := GetRequestProfile(MANAGER, req.Profile, req.Metric)
	if !profile_.HasValue() {
		return res
	}
	profile := profile_.Value
	att := profile.GetAttributes()
	var spt routing.IShortestPathTree
	g_ := profile.GetTransitGraph("monday")
	if g_.HasValue() {
		g := g_.Value
		spt = routing.NewShortestPathTree4(g, 36000, 43200)
	} else {
		g_ := profile.GetGraph()
		if !g_.HasValue() {
			return BadRequest("Graph not found")
		}
		g := g_.Value
		spt = routing.NewShortestPathTree5(g)
	}
	resp := isochrone.ComputeIsochrone(spt, att, loc, req.Range)
	return OK(&resp)
}

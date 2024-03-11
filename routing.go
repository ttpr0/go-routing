package main

import (
	"fmt"
	"math/rand"

	"github.com/ttpr0/go-routing/geo"
	"github.com/ttpr0/go-routing/routing"
	. "github.com/ttpr0/go-routing/util"
	"golang.org/x/exp/slog"
)

type RoutingRequest struct {
	Start     []float32 `json:"start"`
	End       []float32 `json:"end"`
	Key       int32     `json:"key"`
	Draw      bool      `json:"drawRouting"`
	Alg       string    `json:"algorithm"`
	Stepcount int       `json:"stepount"`
}

type DrawContextRequest struct {
	Start     []float32 `json:"start"`
	End       []float32 `json:"end"`
	Algorithm string    `json:"algorithm"`
}

type DrawRoutingRequest struct {
	Key       int `json:"key"`
	Stepcount int `json:"stepcount"`
}

type DrawContextResponse struct {
	Key int `json:"key"`
}

type RoutingResponse struct {
	Type     string           `json:"type"`
	Finished bool             `json:"finished"`
	Features []GeoJSONFeature `json:"features"`
	Key      int              `json:"key"`
}

func NewRoutingResponse(lines []geo.CoordArray, finished bool, key int) RoutingResponse {
	resp := RoutingResponse{}
	resp.Type = "FeatureCollection"
	resp.Finished = finished
	resp.Key = key
	resp.Features = make([]GeoJSONFeature, 0, 10)
	for _, line := range lines {
		obj := NewGeoJSONFeature()
		obj.Geom["type"] = "LineString"
		iter := line.GetIterator()
		arr := make([][2]float32, 0, 2)
		for {
			coord, ok := iter.Next()
			if !ok {
				break
			}
			arr = append(arr, coord)
		}
		obj.Geom["coordinates"] = arr
		obj.Props["value"] = 0
		resp.Features = append(resp.Features, obj)
	}
	return resp
}

func HandleRoutingRequest(req RoutingRequest) Result {
	if req.Draw {
		return BadRequest("Draw not implemented")
	}
	start := geo.Coord{req.Start[0], req.Start[1]}
	end := geo.Coord{req.End[0], req.End[1]}
	profile_ := ProfileFromAlg(req.Alg)
	if !profile_.HasValue() {
		return BadRequest("Profile not found")
	}
	profile := profile_.Value
	attr := profile.GetAttributes()
	var alg routing.IShortestPath
	switch req.Alg {
	case "Dijkstra":
		g := profile.GetGraph()
		alg = routing.NewDijkstra(g.Value, GetClosestNode(start, g.Value), GetClosestNode(end, g.Value))
	case "A*":
		g := profile.GetGraph()
		alg = routing.NewAStar(g.Value, GetClosestNode(start, g.Value), GetClosestNode(end, g.Value))
	case "Bidirect-Dijkstra":
		g := profile.GetGraph()
		alg = routing.NewBidirectDijkstra(g.Value, GetClosestNode(start, g.Value), GetClosestNode(end, g.Value))
	case "Bidirect-A*":
		g := profile.GetGraph()
		alg = routing.NewBidirectAStar(g.Value, GetClosestNode(start, g.Value), GetClosestNode(end, g.Value))
	// case "Distributed-Dijkstra":
	// 	alg = routing.NewDistributedDijkstra(MANAGER, GetClosestNode(start, GRAPH), GetClosestNode(end, GRAPH))
	case "BODijkstra":
		g := profile.GetTiledGraph()
		alg = routing.NewBODijkstra(g.Value, GetClosestNode(start, g.Value), GetClosestNode(end, g.Value))
	case "CH":
		g := profile.GetCHGraph()
		alg = routing.NewCH(g.Value, GetClosestNode(start, g.Value), GetClosestNode(end, g.Value))
	default:
		g := profile.GetGraph()
		alg = routing.NewDijkstra(g.Value, GetClosestNode(start, g.Value), GetClosestNode(end, g.Value))
	}
	slog.Debug(fmt.Sprintf("Using algorithm: %v", req.Alg))
	slog.Debug(fmt.Sprintf("Start Caluclating shortest path between %v and %v", start, end))
	ok := alg.CalcShortestPath()
	if !ok {
		slog.Debug("routing failed")
		return BadRequest("routing failed")
	}
	slog.Debug("shortest path found")
	path := alg.GetShortestPath()
	slog.Debug("start building response")
	resp := NewRoutingResponse(path.GetGeometry(attr), true, int(req.Key))
	slog.Debug("reponse build")
	return OK(resp)
}

var algs_dict Dict[int, Tuple[IRoutingProfile, routing.IShortestPath]] = NewDict[int, Tuple[IRoutingProfile, routing.IShortestPath]](10)

func HandleCreateContextRequest(req DrawContextRequest) Result {
	// process request
	start := geo.Coord{req.Start[0], req.Start[1]}
	end := geo.Coord{req.End[0], req.End[1]}
	profile_ := ProfileFromAlg(req.Algorithm)
	if !profile_.HasValue() {
		return BadRequest("Profile not found")
	}
	profile := profile_.Value
	var alg routing.IShortestPath
	switch req.Algorithm {
	case "Dijkstra":
		g := profile.GetGraph()
		alg = routing.NewDijkstra(g.Value, GetClosestNode(start, g.Value), GetClosestNode(end, g.Value))
	case "A*":
		g := profile.GetGraph()
		alg = routing.NewAStar(g.Value, GetClosestNode(start, g.Value), GetClosestNode(end, g.Value))
	case "Bidirect-Dijkstra":
		g := profile.GetGraph()
		alg = routing.NewBidirectDijkstra(g.Value, GetClosestNode(start, g.Value), GetClosestNode(end, g.Value))
	case "Bidirect-A*":
		g := profile.GetGraph()
		alg = routing.NewBidirectAStar(g.Value, GetClosestNode(start, g.Value), GetClosestNode(end, g.Value))
	// case "Distributed-Dijkstra":
	// 	alg = routing.NewDistributedDijkstra(MANAGER, GetClosestNode(start, GRAPH), GetClosestNode(end, GRAPH))
	case "BODijkstra":
		g := profile.GetTiledGraph()
		alg = routing.NewBODijkstra(g.Value, GetClosestNode(start, g.Value), GetClosestNode(end, g.Value))
	case "CH":
		g := profile.GetCHGraph()
		alg = routing.NewCH(g.Value, GetClosestNode(start, g.Value), GetClosestNode(end, g.Value))
	default:
		g := profile.GetGraph()
		alg = routing.NewDijkstra(g.Value, GetClosestNode(start, g.Value), GetClosestNode(end, g.Value))
	}
	key := -1
	for {
		k := rand.Intn(1000)
		if !algs_dict.ContainsKey(k) {
			algs_dict[k] = MakeTuple(profile, alg)
			key = k
			break
		}
	}
	resp := DrawContextResponse{key}
	return OK(resp)
}

func HandleRoutingStepRequest(req DrawRoutingRequest) Result {
	// process request
	var profile IRoutingProfile
	var alg routing.IShortestPath
	if req.Key != -1 && algs_dict.ContainsKey(req.Key) {
		item := algs_dict[req.Key]
		profile = item.A
		alg = item.B
	} else {
		return BadRequest("key not found")
	}
	attr := profile.GetAttributes()

	edges := NewList[geo.CoordArray](10)
	finished := !alg.Steps(req.Stepcount, func(edge int32) {
		edges.Add(attr.GetEdgeGeom(edge))
	})
	var resp RoutingResponse
	if finished {
		path := alg.GetShortestPath()
		resp = NewRoutingResponse(path.GetGeometry(attr), true, req.Key)
		algs_dict.Delete(req.Key)
	} else {
		resp = NewRoutingResponse(edges, finished, req.Key)
	}

	return OK(resp)
}

func ProfileFromAlg(alg string) Optional[IRoutingProfile] {
	var profile Optional[IRoutingProfile]
	switch alg {
	case "Dijkstra":
		profile = MANAGER.GetProfile("driving-car-ch")
	case "A*":
		profile = MANAGER.GetProfile("driving-car-ch")
	case "Bidirect-Dijkstra":
		profile = MANAGER.GetProfile("driving-car-ch")
	case "Bidirect-A*":
		profile = MANAGER.GetProfile("driving-car-ch")
	// case "Distributed-Dijkstra":
	// 	alg = routing.NewDistributedDijkstra(MANAGER, GetClosestNode(start, GRAPH), GetClosestNode(end, GRAPH))
	case "BODijkstra":
		profile = MANAGER.GetProfile("driving-car-overlay")
	case "CH":
		profile = MANAGER.GetProfile("driving-car-ch")
	default:
		profile = MANAGER.GetProfile("driving-car-ch")
	}
	return profile
}

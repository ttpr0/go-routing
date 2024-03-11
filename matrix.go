package main

import (
	"strings"
	"sync"

	"github.com/ttpr0/go-routing/batched/onetomany"
	"github.com/ttpr0/go-routing/geo"
	. "github.com/ttpr0/go-routing/util"
	"golang.org/x/exp/slog"
)

type MatrixRequest struct {
	Sources      Array[geo.Coord] `json:"sources"`
	Destinations Array[geo.Coord] `json:"destinations"`
	Profile      string           `json:"profile"`
	Metric       string           `json:"metric"`
	MaxRange     int32            `json:"max_range"`
	TimeWindow   [2]int32         `json:"time_window"`
	ScheduleDay  string           `json:"schedule_day"`
}

type MatrixResponse struct {
	Distances Matrix[float32] `json:"distances"`
}

func HandleMatrixRequest(req MatrixRequest) Result {
	var max_range int32
	if req.MaxRange > 0 {
		max_range = req.MaxRange
	} else {
		max_range = 100000000
	}
	// get profile
	var profile IRoutingProfile
	{
		tokens := strings.Split(req.Profile, "-")
		if len(tokens) != 2 {
			return BadRequest("Invalid profile")
		}
		typ, err := ProfileTypeFromString(tokens[0])
		if err != nil {
			return BadRequest("Invalid profile type")
		}
		vehicle, err := VehicleTypeFromString(tokens[1])
		if err != nil {
			return BadRequest("Invalid vehicle type")
		}
		var metric MetricType
		switch req.Metric {
		case "time":
			metric = FASTEST
		case "distance":
			metric = SHORTEST
		default:
			return BadRequest("Invalid metric type")
		}
		profile := MANAGER.GetMatchingProfile(typ, vehicle, metric)
		if !profile.HasValue() {
			return BadRequest("Profile not found")
		}
	}
	// get graph
	var source_nodes Array[int32]
	var target_nodes Array[int32]
	var otm onetomany.IOneToMany
	{
		transit_g := profile.GetTransitGraph(req.ScheduleDay)
		if transit_g.HasValue() {
			source_nodes = MapCoordsToNodes(transit_g.Value, req.Sources)
			target_nodes = MapCoordsToNodes(transit_g.Value, req.Destinations)
			otm = onetomany.NewTransitDijkstra(transit_g.Value, max_range, req.TimeWindow[0], req.TimeWindow[1])
		} else {
			ch_g := profile.GetCHGraph()
			if ch_g.HasValue() {
				source_nodes = MapCoordsToNodes(ch_g.Value, req.Sources)
				target_nodes = MapCoordsToNodes(ch_g.Value, req.Destinations)
				otm = onetomany.NewRangeRPHAST(ch_g.Value, target_nodes, max_range)
			} else {
				overlay_g := profile.GetTiledGraph()
				if overlay_g.HasValue() {
					source_nodes = MapCoordsToNodes(overlay_g.Value, req.Sources)
					target_nodes = MapCoordsToNodes(overlay_g.Value, req.Destinations)
					otm = onetomany.NewGRASP(overlay_g.Value, target_nodes, max_range)
				} else {
					s_g := profile.GetGraph()
					if !s_g.HasValue() {
						return BadRequest("Graph not found")
					}
					source_nodes = MapCoordsToNodes(s_g.Value, req.Sources)
					target_nodes = MapCoordsToNodes(s_g.Value, req.Destinations)
					otm = onetomany.NewRangeDijkstra(s_g.Value, max_range)
				}
			}
		}
	}

	slog.Info("Run Matrix Request")

	source_chan := make(chan Tuple[int, int32], source_nodes.Length())
	for i := 0; i < source_nodes.Length(); i++ {
		source_chan <- MakeTuple(i, source_nodes[i])
	}
	close(source_chan)

	matrix := NewMatrix[float32](source_nodes.Length(), target_nodes.Length())
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
	slog.Info("Matrix reponse build")
	return OK(resp)
}

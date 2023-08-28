package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"

	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/access/decay"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/access/provider"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/access/view"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/graph"
)

func GetClosestNode2(point geo.Coord, graph graph.IGraph) int32 {
	distance := -1.0
	id := 0
	newdistance := 0.0
	geom := graph.GetGeometry()
	for i := 0; i < len(geom.GetAllNodes()); i++ {
		p := geom.GetNode(int32(i))
		newdistance = math.Sqrt(math.Pow(float64(point[1])-float64(p[1]), 2) + math.Pow(float64(point[0])-float64(p[0]), 2))
		if distance == -1 {
			distance = newdistance
			id = i
		} else if newdistance < distance {
			distance = newdistance
			id = i
		}
	}
	return int32(id)
}

func GetClosestNode(point geo.Coord, graph graph.IGraph) int32 {
	index := graph.GetIndex()
	id, _ := index.GetClosestNode(point)
	return id
}

type GeoJSONFeature struct {
	Type  string         `json:"type"`
	Geom  map[string]any `json:"geometry"`
	Props map[string]any `json:"properties"`
}

func NewGeoJSONFeature() GeoJSONFeature {
	line := GeoJSONFeature{}
	line.Type = "Feature"
	line.Geom = make(map[string]any)
	line.Props = make(map[string]any)
	return line
}

func ReadRequestBody[T any](r *http.Request) T {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err.Error())
	}
	var req T
	err = json.Unmarshal(data, &req)
	if err != nil {
		fmt.Println(err.Error())
	}
	return req
}

func WriteResponse[T any](w http.ResponseWriter, resp T, status int) {
	data, _ := json.Marshal(resp)
	w.WriteHeader(status)
	w.Write(data)
}

func GetRoutingProvider(param RoutingRequestParams) provider.IRoutingProvider {
	var provider provider.RoutingProvider

	if param.Profile != "" {
		provider.SetProfile(param.Profile)
	}
	if param.RangeType != "" {
		provider.SetRangeType(param.RangeType)
	}
	if param.LocationType != "" {
		provider.SetParameter("location_type", param.LocationType)
	}

	return &provider
}

func GetDemandView(param DemandRequestParams) view.IPointView {
	var demand_view view.IPointView
	if param.Locations != nil && param.Weights != nil {
		demand_view = view.NewPointView(param.Locations, param.Weights)
	} else if param.Locations != nil {
		demand_view = view.NewPointView(param.Locations, nil)
	}
	return demand_view
}

func GetSupplyView(param SupplyRequestParams) view.IPointView {
	var supply_view view.IPointView
	if param.Locations != nil && param.Weights != nil {
		supply_view = view.NewPointView(param.Locations, param.Weights)
	} else if param.Locations != nil {
		supply_view = view.NewPointView(param.Locations, nil)
	}
	return supply_view
}

func GetDistanceDecay(param DecayRequestParams) decay.IDistanceDecay {
	switch param.Type {
	case "hybrid":
		if param.Ranges == nil || param.RangeFactors == nil {
			return nil
		}
		if len(param.Ranges) == 0 || len(param.RangeFactors) != len(param.Ranges) {
			return nil
		}
		return decay.NewHybridDecay(param.Ranges, param.RangeFactors)
	case "linear":
		if param.MaxRange <= 0 {
			return nil
		}
		return decay.NewLinearDecay(param.MaxRange)
	default:
		return nil
	}
}

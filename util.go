package main

import (
	"os"
	"strings"

	"github.com/ttpr0/go-routing/attr"
	"github.com/ttpr0/go-routing/comps"
	"github.com/ttpr0/go-routing/geo"
	. "github.com/ttpr0/go-routing/util"
)

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

func BuildFastestWeighting(base comps.IGraphBase, attributes *attr.GraphAttributes) *comps.DefaultWeighting {
	weights := comps.NewDefaultWeighting(base)
	for i := 0; i < base.EdgeCount(); i++ {
		attr := attributes.GetEdgeAttribs(int32(i))
		w := attr.Length * 3.6 / float32(attr.Maxspeed)
		if w < 1 {
			w = 1
		}
		weights.SetEdgeWeight(int32(i), int32(w))
	}

	return weights
}

func BuildShortestWeighting(base comps.IGraphBase, attributes *attr.GraphAttributes) *comps.DefaultWeighting {
	weights := comps.NewDefaultWeighting(base)
	for i := 0; i < base.EdgeCount(); i++ {
		attr := attributes.GetEdgeAttribs(int32(i))
		w := attr.Length
		if w < 1 {
			w = 1
		}
		weights.SetEdgeWeight(int32(i), int32(w))
	}

	return weights
}

func BuildPedestrianWeighting(base comps.IGraphBase, attributes *attr.GraphAttributes) *comps.DefaultWeighting {
	weights := comps.NewDefaultWeighting(base)
	for i := 0; i < base.EdgeCount(); i++ {
		attr := attributes.GetEdgeAttribs(int32(i))
		w := attr.Length * 3.6 / 3
		if w < 1 {
			w = 1
		}
		weights.SetEdgeWeight(int32(i), int32(w))
	}

	return weights
}

func BuildTCWeighting(base comps.IGraphBase, attributes *attr.GraphAttributes) *comps.TCWeighting {
	weight := comps.NewTCWeighting(base)

	for i := 0; i < int(base.EdgeCount()); i++ {
		attr := attributes.GetEdgeAttribs(int32(i))
		weight.SetEdgeWeight(int32(i), int32(attr.Length/float32(attr.Maxspeed)))
	}

	return weight
}

func IsDirectoryEmpty(path string) bool {
	files, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	return len(files) == 0
}

func GetRequestProfile(manager *RoutingManager, profile, metric string) (Optional[IRoutingProfile], Result) {
	var prof IRoutingProfile
	{
		tokens := strings.Split(profile, "-")
		if len(tokens) != 2 {
			return None[IRoutingProfile](), BadRequest("Invalid profile")
		}
		typ, err := ProfileTypeFromString(tokens[0])
		if err != nil {
			return None[IRoutingProfile](), BadRequest("Invalid profile type")
		}
		vehicle, err := VehicleTypeFromString(tokens[1])
		if err != nil {
			return None[IRoutingProfile](), BadRequest("Invalid vehicle type")
		}
		var metr MetricType
		switch metric {
		case "time":
			metr = FASTEST
		case "distance":
			metr = SHORTEST
		default:
			return None[IRoutingProfile](), BadRequest("Invalid metric type")
		}
		prof_ := MANAGER.GetMatchingProfile(typ, vehicle, metr)
		if !prof_.HasValue() {
			return None[IRoutingProfile](), BadRequest("Profile not found")
		}
		prof = prof_.Value
	}
	return Some(prof), OK("")
}

func MapCoordsToNodes(att attr.IAttributes, coords []geo.Coord) Array[int32] {
	nodes := make([]int32, len(coords))
	for i, coord := range coords {
		id, ok := att.GetClosestNode(coord)
		if ok {
			nodes[i] = id
		} else {
			nodes[i] = -1
		}
	}
	return nodes
}

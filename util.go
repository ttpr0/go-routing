package main

import (
	"os"
	"strings"

	"github.com/ttpr0/go-routing/attr"
	"github.com/ttpr0/go-routing/geo"
	"github.com/ttpr0/go-routing/parser"
	. "github.com/ttpr0/go-routing/util"
)

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

func GetDecoder(typ ProfileType) parser.IOSMDecoder {
	var decoder parser.IOSMDecoder
	switch typ {
	case DRIVING:
		decoder = &parser.DrivingDecoder{}
	case CYCLING:
		decoder = &parser.CyclingDecoder{}
	case WALKING:
		decoder = &parser.WalkingDecoder{}
	}
	return decoder
}

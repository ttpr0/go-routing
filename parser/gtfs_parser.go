package parser

import (
	"encoding/json"
	"os"
	"os/exec"

	"github.com/ttpr0/go-routing/comps"
	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// gtfs parser
//*******************************************

type GTFSOutput struct {
	Stops     List[GTFSNode]                      `json:"stops"`
	Conns     List[GTFSConn]                      `json:"conns"`
	Schedules Dict[string, Array[List[[2]int32]]] `json:"schedules"`
}

type GTFSNode struct {
	Lon float32 `json:"lon"`
	Lat float32 `json:"lat"`
}

type GTFSConn struct {
	StopA   int32 `json:"stop_a"`
	StopB   int32 `json:"stop_b"`
	RouteID int32 `json:"route_id"`
}

func ParseGtfs(gtfs_path string, filter_polygon string) (Array[structs.Node], Array[structs.Connection], Dict[string, Array[[]comps.ConnectionWeight]]) {
	cmd := exec.Command("python", "./parser/gtfs_parser.py", gtfs_path, "-f "+filter_polygon, "-o ./temp.json")
	if err := cmd.Run(); err != nil {
		panic(err)
	}

	file_str, err := os.ReadFile("./temp.json")
	if err != nil {
		panic("missing test data file: geom.json")
	}
	output := GTFSOutput{}
	err = json.Unmarshal(file_str, &output)

	stops := NewArray[structs.Node](output.Stops.Length())
	for i := 0; i < output.Stops.Length(); i++ {
		stops[i] = structs.Node{
			Loc: [2]float32{output.Stops[i].Lon, output.Stops[i].Lat},
		}
	}
	conns := NewArray[structs.Connection](output.Conns.Length())
	for i := 0; i < output.Conns.Length(); i++ {
		conns[i] = structs.Connection{
			StopA:   output.Conns[i].StopA,
			StopB:   output.Conns[i].StopB,
			RouteID: output.Conns[i].RouteID,
		}
	}
	schedules := NewDict[string, Array[[]comps.ConnectionWeight]](output.Schedules.Length())
	for k, v := range output.Schedules {
		schedule := NewArray[[]comps.ConnectionWeight](conns.Length())
		for i := 0; i < conns.Length(); i++ {
			s := v[i]
			weights := make([]comps.ConnectionWeight, s.Length())
			for j := 0; j < s.Length(); j++ {
				weights[j] = comps.ConnectionWeight{Departure: s[j][0], Arrival: s[j][1]}
			}
			schedule[i] = weights
		}
		schedules[k] = schedule
	}

	return stops, conns, schedules
}

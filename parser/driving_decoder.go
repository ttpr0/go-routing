package parser

import (
	"github.com/ttpr0/go-routing/attr"
	. "github.com/ttpr0/go-routing/util"
)

type DrivingDecoder struct {
}

var driving_types = Dict[string, bool]{"motorway": true, "motorway_link": true, "trunk": true, "trunk_link": true,
	"primary": true, "primary_link": true, "secondary": true, "secondary_link": true, "tertiary": true, "tertiary_link": true,
	"residential": true, "living_street": true, "service": true, "track": true, "unclassified": true, "road": true}

func (self *DrivingDecoder) IsValidHighway(tags Dict[string, string]) bool {
	if !tags.ContainsKey("highway") {
		return false
	}
	if !driving_types.ContainsKey(tags.Get("highway")) {
		return false
	}
	return true
}
func (self *DrivingDecoder) DecodeNode(tags Dict[string, string]) attr.NodeAttribs {
	return attr.NodeAttribs{Type: 0}
}
func (self *DrivingDecoder) DecodeEdge(tags Dict[string, string]) attr.EdgeAttribs {
	templimit := tags.Get("maxspeed")
	str_type := tags.Get("highway")
	oneway := tags.Get("oneway")
	track_type := tags.Get("tracktype")
	surface := tags.Get("surface")
	e := attr.EdgeAttribs{}
	e.Type = _GetType(str_type)
	// e.Templimit = GetTemplimit(templimit, e.Type)
	e.Maxspeed = byte(_GetORSTravelSpeed(e.Type, templimit, track_type, surface))
	e.Oneway = _IsOneway(oneway, e.Type)
	return e
}

package attr

import (
	"encoding/json"
	"errors"
)

//*******************************************
// enums
//*******************************************

type RoadType int8

const (
	MOTORWAY       RoadType = 1
	MOTORWAY_LINK  RoadType = 2
	TRUNK          RoadType = 3
	TRUNK_LINK     RoadType = 4
	PRIMARY        RoadType = 5
	PRIMARY_LINK   RoadType = 6
	SECONDARY      RoadType = 7
	SECONDARY_LINK RoadType = 8
	TERTIARY       RoadType = 9
	TERTIARY_LINK  RoadType = 10
	RESIDENTIAL    RoadType = 11
	LIVING_STREET  RoadType = 12
	UNCLASSIFIED   RoadType = 13
	ROAD           RoadType = 14
	TRACK          RoadType = 15
)

func (self RoadType) String() string {
	switch self {
	case MOTORWAY:
		return "motorway"
	case MOTORWAY_LINK:
		return "motorway_link"
	case TRUNK:
		return "trunk"
	case TRUNK_LINK:
		return "trunk_link"
	case PRIMARY:
		return "primary"
	case PRIMARY_LINK:
		return "primary_link"
	case SECONDARY:
		return "secondary"
	case SECONDARY_LINK:
		return "secondary_link"
	case TERTIARY:
		return "tertiary"
	case TERTIARY_LINK:
		return "tertiary_link"
	case RESIDENTIAL:
		return "residential"
	case LIVING_STREET:
		return "living_street"
	case UNCLASSIFIED:
		return "unclassified"
	case ROAD:
		return "road"
	case TRACK:
		return "track"
	}
	return ""
}

func RoadTypeFromString(typ string) RoadType {
	switch typ {
	case "motorway":
		return MOTORWAY
	case "motorway_link":
		return MOTORWAY_LINK
	case "trunk":
		return TRUNK
	case "trunk_link":
		return TRUNK_LINK
	case "primary":
		return PRIMARY
	case "primary_link":
		return PRIMARY_LINK
	case "secondary":
		return SECONDARY
	case "secondary_link":
		return SECONDARY_LINK
	case "tertiary":
		return TERTIARY
	case "tertiary_link":
		return TERTIARY_LINK
	case "residential":
		return RESIDENTIAL
	case "living_street":
		return LIVING_STREET
	case "unclassified":
		return UNCLASSIFIED
	case "road":
		return ROAD
	case "track":
		return TRACK
	}
	return 0
}

func (self RoadType) MarshalJSON() ([]byte, error) {
	return json.Marshal(self.String())
}
func (self *RoadType) UnmarshalJSON(data []byte) error {
	var typ string
	if err := json.Unmarshal(data, &typ); err != nil {
		return err
	}
	prof_typ := RoadTypeFromString(typ)
	if prof_typ == 0 {
		return errors.New("invalid road type")
	}
	*self = prof_typ
	return nil
}

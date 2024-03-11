package parser

import (
	"strconv"

	"github.com/ttpr0/go-routing/attr"
)

//*******************************************
// utility methods
//*******************************************

func _IsOneway(oneway string, str_type attr.RoadType) bool {
	if str_type == attr.MOTORWAY || str_type == attr.TRUNK || str_type == attr.MOTORWAY_LINK || str_type == attr.TRUNK_LINK {
		return true
	} else if oneway == "yes" {
		return true
	}
	return false
}

func _GetType(typ string) attr.RoadType {
	switch typ {
	case "motorway":
		return attr.MOTORWAY
	case "motorway_link":
		return attr.MOTORWAY_LINK
	case "trunk":
		return attr.TRUNK
	case "trunk_link":
		return attr.TRUNK_LINK
	case "primary":
		return attr.PRIMARY
	case "primary_link":
		return attr.PRIMARY_LINK
	case "secondary":
		return attr.SECONDARY
	case "secondary_link":
		return attr.SECONDARY_LINK
	case "tertiary":
		return attr.TERTIARY
	case "tertiary_link":
		return attr.TERTIARY_LINK
	case "residential":
		return attr.RESIDENTIAL
	case "living_street":
		return attr.LIVING_STREET
	case "unclassified":
		return attr.UNCLASSIFIED
	case "road":
		return attr.ROAD
	case "track":
		return attr.TRACK
	}
	return 0
}

func _GetTemplimit(templimit string, streettype attr.RoadType) int32 {
	var w int32
	if templimit == "" {
		if streettype == attr.MOTORWAY || streettype == attr.TRUNK {
			w = 130
		} else if streettype == attr.MOTORWAY_LINK || streettype == attr.TRUNK_LINK {
			w = 50
		} else if streettype == attr.PRIMARY || streettype == attr.SECONDARY {
			w = 90
		} else if streettype == attr.TERTIARY {
			w = 70
		} else if streettype == attr.PRIMARY_LINK || streettype == attr.SECONDARY_LINK || streettype == attr.TERTIARY_LINK {
			w = 30
		} else if streettype == attr.RESIDENTIAL {
			w = 40
		} else if streettype == attr.LIVING_STREET {
			w = 10
		} else {
			w = 25
		}
	} else if templimit == "walk" {
		w = 10
	} else if templimit == "none" {
		w = 130
	} else {
		t, err := strconv.Atoi(templimit)
		if err != nil {
			w = 20
		} else {
			w = int32(t)
		}
	}
	return w
}

func _GetORSTravelSpeed(streettype attr.RoadType, maxspeed string, tracktype string, surface string) int32 {
	var speed int32

	// check if maxspeed is set
	if maxspeed != "" {
		if maxspeed == "walk" {
			speed = 10
		} else if maxspeed == "none" {
			speed = 110
		} else {
			t, err := strconv.Atoi(maxspeed)
			if err != nil {
				speed = 20
			} else {
				speed = int32(t)
			}
		}
		speed = int32(0.9 * float32(speed))
	}

	// set defaults
	if maxspeed == "" {
		switch streettype {
		case attr.MOTORWAY:
			speed = 100
		case attr.TRUNK:
			speed = 85
		case attr.MOTORWAY_LINK, attr.TRUNK_LINK:
			speed = 60
		case attr.PRIMARY:
			speed = 65
		case attr.SECONDARY:
			speed = 60
		case attr.TERTIARY:
			speed = 50
		case attr.PRIMARY_LINK, attr.SECONDARY_LINK:
			speed = 50
		case attr.TERTIARY_LINK:
			speed = 40
		case attr.UNCLASSIFIED:
			speed = 30
		case attr.RESIDENTIAL:
			speed = 30
		case attr.LIVING_STREET:
			speed = 10
		case attr.ROAD:
			speed = 20
		case attr.TRACK:
			if tracktype == "" {
				speed = 15
			} else {
				switch tracktype {
				case "grade1":
					speed = 40
				case "grade2":
					speed = 30
				case "grade3":
					speed = 20
				case "grade4":
					speed = 15
				case "grade5":
					speed = 10
				default:
					speed = 15
				}
			}
		default:
			speed = 20
		}
	}

	// check if surface is set
	if surface != "" {
		switch surface {
		case "cement", "compacted":
			if speed > 80 {
				speed = 80
			}
		case "fine_gravel":
			if speed > 60 {
				speed = 60
			}
		case "paving_stones", "metal", "bricks":
			if speed > 40 {
				speed = 40
			}
		case "grass", "wood", "sett", "grass_paver", "gravel", "unpaved", "ground", "dirt", "pebblestone", "tartan":
			if speed > 30 {
				speed = 30
			}
		case "cobblestone", "clay":
			if speed > 20 {
				speed = 20
			}
		case "earth", "stone", "rocky", "sand":
			if speed > 15 {
				speed = 15
			}
		case "mud":
			if speed > 10 {
				speed = 10
			}
		}
	}

	if speed == 0 {
		speed = 10
	}
	return speed
}

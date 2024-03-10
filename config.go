package main

import (
	"encoding/json"
	"os"

	. "github.com/ttpr0/go-routing/util"
	"gopkg.in/yaml.v3"
)

//**********************************************************
// config
//**********************************************************

func ReadConfig(file string) Config {
	data, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	var config Config
	yaml.Unmarshal(data, &config)
	return config
}

type Config struct {
	Build struct {
		Source   SourceOptions                 `yaml:"source"`
		Profiles Dict[string, *ProfileOptions] `yaml:"profiles"`
	} `yaml:"build"`
	BuildGraphs bool `yaml:"build-graphs"`
	Services    struct {
	} `yaml:"services"`
}

type SourceOptions struct {
	OSM  string `yaml:"osm"`
	GTFS string `yaml:"gtfs"`
}

//**********************************************************
// profile options
//**********************************************************

type ProfileOptions struct {
	Value IProfileOptions
}

func (self *ProfileOptions) UnmarshalYAML(value *yaml.Node) error {
	m := map[string]interface{}{}
	if err := value.Decode(&m); err != nil {
		return err
	}
	typ := ProfileTypeFromString(m["type"].(string))
	switch typ {
	case DRIVING:
		val := DrivingOptions{}
		value.Decode(&val)
		self.Value = val
	case WALKING:
		val := WalkingOptions{}
		value.Decode(&val)
		self.Value = val
	case TRANSIT:
		val := TransitOptions{}
		value.Decode(&val)
		self.Value = val
	default:
		self.Value = nil
	}
	return nil
}

type IProfileOptions interface {
	Type() ProfileType
}

type DrivingOptions struct {
	Vehicle     VehicleType `yaml:"vehicle"`
	Metric      MetricType  `yaml:"metric"`
	Preperation struct {
		Contraction     bool   `yaml:"contraction"`
		Overlay         bool   `yaml:"overlay"`
		MaxNodesPerCell int    `yaml:"max-nodes-per-cell"`
		OverlayMethod   string `yaml:"overlay-method"`
	} `yaml:"preperation"`
}

func (self DrivingOptions) Type() ProfileType {
	return DRIVING
}

type WalkingOptions struct {
	Vehicle VehicleType `yaml:"vehicle"`
	Metric  MetricType  `yaml:"metric"`
}

func (self WalkingOptions) Type() ProfileType {
	return WALKING
}

type TransitOptions struct {
	Vehicle VehicleType `yaml:"vehicle"`
	// Metric      MetricType  `yaml:"metric"`
	Preperation struct {
		FilterPolygon    string `yaml:"filter-polygon"`
		MaxTransferRange int32  `yaml:"max-transfer-range"`
	} `yaml:"preperation"`
}

func (self TransitOptions) Type() ProfileType {
	return TRANSIT
}

//**********************************************************
// enums
//**********************************************************

type ProfileType byte

const (
	DRIVING ProfileType = 0
	WALKING ProfileType = 1
	TRANSIT ProfileType = 2
)

func (self ProfileType) String() string {
	switch self {
	case DRIVING:
		return "driving"
	case WALKING:
		return "walking"
	case TRANSIT:
		return "transit"
	default:
		panic("unknown profile type")
	}
}
func (self ProfileType) MarshalJSON() ([]byte, error) {
	return json.Marshal(self.String())
}
func (self *ProfileType) UnmarshalJSON(data []byte) error {
	var typ string
	if err := json.Unmarshal(data, &typ); err != nil {
		return err
	}
	*self = ProfileTypeFromString(typ)
	return nil
}

func ProfileTypeFromString(s string) ProfileType {
	switch s {
	case "driving":
		return DRIVING
	case "walking":
		return WALKING
	case "transit":
		return TRANSIT
	default:
		panic("unknown profile type")
	}
}

type MetricType byte

const (
	FASTEST  MetricType = 0
	SHORTEST MetricType = 1
)

func (self MetricType) String() string {
	switch self {
	case FASTEST:
		return "fastest"
	case SHORTEST:
		return "shortest"
	default:
		panic("unknown metric type")
	}
}
func (self MetricType) MarshalJSON() ([]byte, error) {
	return json.Marshal(self.String())
}
func (self *MetricType) UnmarshalJSON(data []byte) error {
	var typ string
	if err := json.Unmarshal(data, &typ); err != nil {
		return err
	}
	*self = MetricTypeFromString(typ)
	return nil
}
func (self MetricType) MarshalYAML() (any, error) {
	return self.String(), nil
}
func (self *MetricType) UnmarshalYAML(value *yaml.Node) error {
	*self = MetricTypeFromString(value.Value)
	return nil
}

func MetricTypeFromString(s string) MetricType {
	switch s {
	case "fastest":
		return FASTEST
	case "shortest":
		return SHORTEST
	default:
		panic("unknown metric type")
	}
}

type VehicleType byte

const (
	CAR  VehicleType = 0
	FOOT VehicleType = 1
)

func (self VehicleType) String() string {
	switch self {
	case CAR:
		return "car"
	case FOOT:
		return "foot"
	default:
		panic("unknown vehicle type")
	}
}
func (self VehicleType) MarshalJSON() ([]byte, error) {
	return json.Marshal(self.String())
}
func (self *VehicleType) UnmarshalJSON(data []byte) error {
	var typ string
	if err := json.Unmarshal(data, &typ); err != nil {
		return err
	}
	*self = VehicleTypeFromString(typ)
	return nil
}
func (self VehicleType) MarshalYAML() (any, error) {
	return self.String(), nil
}
func (self *VehicleType) UnmarshalYAML(value *yaml.Node) error {
	*self = VehicleTypeFromString(value.Value)
	return nil
}

func VehicleTypeFromString(s string) VehicleType {
	switch s {
	case "car":
		return CAR
	case "foot":
		return FOOT
	default:
		panic("unknown vehicle type")
	}
}

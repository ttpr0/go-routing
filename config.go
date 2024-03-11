package main

import (
	"encoding/json"
	"errors"
	"os"

	. "github.com/ttpr0/go-routing/util"
	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

//**********************************************************
// config
//**********************************************************

func ReadConfig(file string) Config {
	slog.Info("Reading config file")
	data, err := os.ReadFile(file)
	if err != nil {
		slog.Error("failed to read config file: " + err.Error())
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
	typ, err := ProfileTypeFromString(m["type"].(string))
	if err != nil {
		return err
	}
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
	Preparation struct {
		Contraction     bool   `yaml:"contraction"`
		Overlay         bool   `yaml:"overlay"`
		MaxNodesPerCell int    `yaml:"max-nodes-per-cell"`
		OverlayMethod   string `yaml:"overlay-method"`
	} `yaml:"preparation"`
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

type CyclingOptions struct {
	Vehicle VehicleType `yaml:"vehicle"`
	Metric  MetricType  `yaml:"metric"`
}

func (self CyclingOptions) Type() ProfileType {
	return CYCLING
}

type TransitOptions struct {
	Vehicle VehicleType `yaml:"vehicle"`
	// Metric      MetricType  `yaml:"metric"`
	Preparation struct {
		FilterPolygon    string `yaml:"filter-polygon"`
		MaxTransferRange int32  `yaml:"max-transfer-range"`
	} `yaml:"preparation"`
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
	CYCLING ProfileType = 2
	TRANSIT ProfileType = 3
)

func (self ProfileType) String() string {
	switch self {
	case DRIVING:
		return "driving"
	case WALKING:
		return "walking"
	case CYCLING:
		return "cycling"
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
	prof_typ, err := ProfileTypeFromString(typ)
	*self = prof_typ
	return err
}

func ProfileTypeFromString(s string) (ProfileType, error) {
	switch s {
	case "driving":
		return DRIVING, nil
	case "walking":
		return WALKING, nil
	case "cycling":
		return CYCLING, nil
	case "transit":
		return TRANSIT, nil
	default:
		return DRIVING, errors.New("unknown profile type")
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
	err := json.Unmarshal(data, &typ)
	if err != nil {
		return err
	}
	*self, err = MetricTypeFromString(typ)
	return err
}
func (self MetricType) MarshalYAML() (any, error) {
	return self.String(), nil
}
func (self *MetricType) UnmarshalYAML(value *yaml.Node) error {
	typ, err := MetricTypeFromString(value.Value)
	if err != nil {
		return err
	}
	*self = typ
	return nil
}

func MetricTypeFromString(s string) (MetricType, error) {
	switch s {
	case "fastest":
		return FASTEST, nil
	case "shortest":
		return SHORTEST, nil
	default:
		return FASTEST, errors.New("unknown metric type")
	}
}

type VehicleType byte

const (
	CAR  VehicleType = 0
	FOOT VehicleType = 1
	BIKE VehicleType = 2
)

func (self VehicleType) String() string {
	switch self {
	case CAR:
		return "car"
	case FOOT:
		return "foot"
	case BIKE:
		return "bike"
	default:
		panic("unknown vehicle type")
	}
}
func (self VehicleType) MarshalJSON() ([]byte, error) {
	return json.Marshal(self.String())
}
func (self *VehicleType) UnmarshalJSON(data []byte) error {
	var typ string
	err := json.Unmarshal(data, &typ)
	if err != nil {
		return err
	}
	*self, err = VehicleTypeFromString(typ)
	return err
}
func (self VehicleType) MarshalYAML() (any, error) {
	return self.String(), nil
}
func (self *VehicleType) UnmarshalYAML(value *yaml.Node) error {
	typ, err := VehicleTypeFromString(value.Value)
	if err != nil {
		return err
	}
	*self = typ
	return nil
}

func VehicleTypeFromString(s string) (VehicleType, error) {
	switch s {
	case "car":
		return CAR, nil
	case "foot":
		return FOOT, nil
	case "bike":
		return BIKE, nil
	default:
		return CAR, errors.New("unknown vehicle type")
	}
}

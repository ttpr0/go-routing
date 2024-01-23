package main

import (
	"github.com/ttpr0/go-routing/geo"
	. "github.com/ttpr0/go-routing/util"
)

type RoutingRequestParams struct {
	// *************************************
	// standard routing params
	// *************************************
	Profile string `json:"profile"`

	RangeType string `json:"range_type"`

	// *************************************
	// additional routing params
	// *************************************
	AvoidBorders string `json:"avoid_borders"`

	AvoidFeatures []string `json:"avoid_features"`

	AvoidPolygons geo.Feature `json:"avoid_polygons"`

	// *************************************
	// direction params
	// *************************************
	LocationType string `json:"location_type"`
}

type DemandRequestParams struct {
	// *************************************
	// create new view from locations and weights
	// *************************************
	Locations Array[geo.Coord] `json:"demand_locations"`

	Weights Array[int32] `json:"demand_weights"`
}

type SupplyRequestParams struct {
	// *************************************
	// create new view from locations and weights
	// *************************************
	Locations Array[geo.Coord] `json:"supply_locations"`

	Weights Array[int32] `json:"supply_weights"`
}

type DecayRequestParams struct {
	Type         string    `json:"decay_type"`
	MaxRange     float32   `json:"max_range"`
	Ranges       []float32 `json:"ranges"`
	RangeFactors []float32 `json:"range_factors"`
}

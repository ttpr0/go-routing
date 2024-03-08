package main

import (
	"github.com/ttpr0/go-routing/geo"
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

package main

import (
	"github.com/ttpr0/go-routing/access/decay"
	"github.com/ttpr0/go-routing/access/view"
	"github.com/ttpr0/go-routing/attr"
	"github.com/ttpr0/go-routing/geo"
	"github.com/ttpr0/go-routing/graph"
)

func LoadOrCreate(graph_path string, osm_file string, partition_file string) graph.ITiledGraph {
	// // check if graph files already exist
	// _, err1 := os.Stat(graph_path + "-nodes")
	// _, err2 := os.Stat(graph_path + "-edges")
	// _, err3 := os.Stat(graph_path + "-geom")
	// _, err4 := os.Stat(graph_path + "-tiles")
	// if errors.Is(err1, os.ErrNotExist) || errors.Is(err2, os.ErrNotExist) || errors.Is(err3, os.ErrNotExist) || errors.Is(err4, os.ErrNotExist) {
	// 	// create graph
	// 	g := graph.ParseGraph(osm_file)

	// 	file_str, _ := os.ReadFile(partition_file)
	// 	collection := geo.FeatureCollection{}
	// 	_ = json.Unmarshal(file_str, &collection)

	// 	graph.BuildGraphIndex(g)

	// 	tiles := partitioning.GeometricPartitioning(g, collection.Features())
	// 	tg := graph.PreprocessTiledGraph(g, tiles)

	// 	graph.StoreTiledGraph(tg, graph_path)

	// 	return tg
	// } else {
	// 	return graph.LoadTiledGraph(graph_path)
	// }
	return nil
}

func GetClosestNode(point geo.Coord, graph graph.IGraph) int32 {
	index := graph.GetIndex()
	id, _ := index.GetClosestNode(point)
	return id
}

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

func GetDemandView(param DemandRequestParams) view.IPointView {
	var demand_view view.IPointView
	if param.Locations != nil && param.Weights != nil {
		demand_view = view.NewPointView(param.Locations, param.Weights)
	} else if param.Locations != nil {
		demand_view = view.NewPointView(param.Locations, nil)
	}
	return demand_view
}

func GetSupplyView(param SupplyRequestParams) view.IPointView {
	var supply_view view.IPointView
	if param.Locations != nil && param.Weights != nil {
		supply_view = view.NewPointView(param.Locations, param.Weights)
	} else if param.Locations != nil {
		supply_view = view.NewPointView(param.Locations, nil)
	}
	return supply_view
}

func GetDistanceDecay(param DecayRequestParams) decay.IDistanceDecay {
	switch param.Type {
	case "hybrid":
		if param.Ranges == nil || param.RangeFactors == nil {
			return nil
		}
		if len(param.Ranges) == 0 || len(param.RangeFactors) != len(param.Ranges) {
			return nil
		}
		return decay.NewHybridDecay(param.Ranges, param.RangeFactors)
	case "linear":
		if param.MaxRange <= 0 {
			return nil
		}
		return decay.NewLinearDecay(param.MaxRange)
	default:
		return nil
	}
}

func BuildDefaultWeighting(base graph.IGraphBase, attributes *attr.GraphAttributes) *graph.DefaultWeighting {
	weights := graph.NewDefaultWeighting(base)
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

func BuildPedestrianWeighting(base graph.IGraphBase, attributes *attr.GraphAttributes) *graph.DefaultWeighting {
	weights := graph.NewDefaultWeighting(base)
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

func BuildTCWeighting(base graph.IGraphBase, attributes *attr.GraphAttributes) *graph.TCWeighting {
	weight := graph.NewTCWeighting(base)

	for i := 0; i < int(base.EdgeCount()); i++ {
		attr := attributes.GetEdgeAttribs(int32(i))
		weight.SetEdgeWeight(int32(i), int32(attr.Length/float32(attr.Maxspeed)))
	}

	return weight
}

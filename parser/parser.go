package parser

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
	"github.com/ttpr0/go-routing/attr"
	"github.com/ttpr0/go-routing/comps"
	"github.com/ttpr0/go-routing/geo"
	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
	"golang.org/x/exp/slog"
)

func ParseGraph(pbf_file string, decoder IOSMDecoder) (*comps.GraphBase, *attr.GraphAttributes) {
	nodes := NewList[OSMNode](10000)
	edges := NewList[OSMEdge](10000)
	index_mapping := NewDict[int64, int](10000)
	_ParseOsm(pbf_file, decoder, &nodes, &edges, &index_mapping)
	print("edges: ", edges.Length(), ", nodes: ", nodes.Length())
	base, attr := _CreateGraphBase(&nodes, &edges)
	return base, attr
}

func _ParseOsm(filename string, decoder IOSMDecoder, nodes *List[OSMNode], edges *List[OSMEdge], index_mapping *Dict[int64, int]) {
	osm_nodes := NewDict[int64, TempNode](1000)

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := osmpbf.New(context.Background(), file, runtime.GOMAXPROCS(-1))
	_InitWayHandler(scanner, decoder, &osm_nodes)
	scanner.Close()
	file.Seek(0, 0)
	scanner = osmpbf.New(context.Background(), file, runtime.GOMAXPROCS(-1))
	_NodeHandler(scanner, decoder, &osm_nodes, nodes, index_mapping)
	scanner.Close()
	file.Seek(0, 0)
	scanner = osmpbf.New(context.Background(), file, runtime.GOMAXPROCS(-1))
	_WayHandler(scanner, decoder, edges, &osm_nodes, index_mapping)
	scanner.Close()
	for i := 0; i < edges.Length(); i++ {
		e := edges.Get(i)
		node_a := nodes.Get(e.NodeA)
		node_a.Edges.Add(int32(i))
		nodes.Set(e.NodeA, node_a)
		node_b := nodes.Get(e.NodeB)
		node_b.Edges.Add(int32(i))
		nodes.Set(e.NodeB, node_b)
	}
}

func _CreateGraphBase(osmnodes *List[OSMNode], osmedges *List[OSMEdge]) (*comps.GraphBase, *attr.GraphAttributes) {
	nodes := NewList[structs.Node](osmnodes.Length())
	edges := NewList[structs.Edge](osmedges.Length() * 2)
	node_attrs := NewList[attr.NodeAttribs](osmnodes.Length())
	edge_attrs := NewList[attr.EdgeAttribs](osmedges.Length() * 2)
	node_geoms := NewList[geo.Coord](osmnodes.Length())
	edge_geoms := NewList[geo.CoordArray](osmedges.Length() * 2)

	edge_index_mapping := NewDict[int, int](osmedges.Length())
	for i, osmedge := range *osmedges {
		edge := structs.Edge{
			NodeA: int32(osmedge.NodeA),
			NodeB: int32(osmedge.NodeB),
		}
		edge_attr := osmedge.Attr
		edges.Add(edge)
		edge_attrs.Add(edge_attr)
		edge_geoms.Add(geo.CoordArray(osmedge.Nodes))
		edge_index_mapping[i] = edges.Length() - 1
		if !osmedge.Attr.Oneway {
			edge = structs.Edge{
				NodeA: int32(osmedge.NodeB),
				NodeB: int32(osmedge.NodeA),
			}
			edge_attr = osmedge.Attr
			edges.Add(edge)
			edge_attrs.Add(edge_attr)
			edge_geoms.Add(geo.CoordArray(osmedge.Nodes))
		}
	}

	for _, osmnode := range *osmnodes {
		node := structs.Node{
			Loc: osmnode.Point,
		}
		node_attr := osmnode.Attr
		nodes.Add(node)
		node_attrs.Add(node_attr)
		node_geoms.Add(osmnode.Point)
	}

	base := comps.NewGraphBase(Array[structs.Node](nodes), Array[structs.Edge](edges))
	attr := attr.New(Array[attr.NodeAttribs](node_attrs), Array[attr.EdgeAttribs](edge_attrs), Array[geo.Coord](node_geoms), Array[geo.CoordArray](edge_geoms))
	return base, attr
}

//*******************************************
// osm handler methods
//*******************************************

func _InitWayHandler(scanner *osmpbf.Scanner, decoder IOSMDecoder, osm_nodes *Dict[int64, TempNode]) {
	scanner.SkipNodes = true
	scanner.SkipRelations = true
	for scanner.Scan() {
		switch object := scanner.Object().(type) {
		case *osm.Way:
			tags := Dict[string, string](object.TagMap())
			if !decoder.IsValidHighway(tags) {
				continue
			}
			nodes := object.Nodes.NodeIDs()
			l := len(nodes)
			for i := 0; i < l; i++ {
				ndref := nodes[i].FeatureID().Ref()
				if !osm_nodes.ContainsKey(ndref) {
					(*osm_nodes)[ndref] = TempNode{geo.Coord{0, 0}, 1}
				} else {
					node := (*osm_nodes)[ndref]
					node.Count += 1
					(*osm_nodes)[ndref] = node
				}
			}
			node_a := (*osm_nodes)[nodes[0].FeatureID().Ref()]
			node_b := (*osm_nodes)[nodes[l-1].FeatureID().Ref()]
			node_a.Count += 1
			node_b.Count += 1
			(*osm_nodes)[nodes[0].FeatureID().Ref()] = node_a
			(*osm_nodes)[nodes[l-1].FeatureID().Ref()] = node_b
		default:
			continue
		}
	}
}

func _NodeHandler(scanner *osmpbf.Scanner, decoder IOSMDecoder, osm_nodes *Dict[int64, TempNode], nodes *List[OSMNode], index_mapping *Dict[int64, int]) {
	i := 0
	c := 0

	scanner.SkipWays = true
	scanner.SkipRelations = true
	for scanner.Scan() {
		switch object := scanner.Object().(type) {
		case *osm.Node:
			id := object.FeatureID().Ref()
			if !osm_nodes.ContainsKey(id) {
				continue
			}
			tags := object.TagMap()
			c += 1
			if c%1000 == 0 {
				slog.Debug(fmt.Sprintf("%v", c))
			}
			on := osm_nodes.Get(id)
			if on.Count > 1 {
				node_attr := decoder.DecodeNode(tags)
				node := OSMNode{geo.Coord{float32(object.Lon), float32(object.Lat)}, node_attr, NewList[int32](3)}
				nodes.Add(node)
				index_mapping.Set(id, i)
				i += 1
			}
			on.Point[0] = float32(object.Lon)
			on.Point[1] = float32(object.Lat)
			osm_nodes.Set(id, on)
		default:
			continue
		}
	}
}

func _WayHandler(scanner *osmpbf.Scanner, decoder IOSMDecoder, edges *List[OSMEdge], osm_nodes *Dict[int64, TempNode], index_mapping *Dict[int64, int]) {
	c := 0
	scanner.SkipNodes = true
	scanner.SkipRelations = true
	for scanner.Scan() {
		switch object := scanner.Object().(type) {
		case *osm.Way:
			tags := Dict[string, string](object.TagMap())
			if !decoder.IsValidHighway(tags) {
				continue
			}
			c += 1
			if c%1000 == 0 {
				slog.Debug(fmt.Sprintf("%v", c))
			}

			nodes := object.Nodes.NodeIDs()
			l := len(nodes)
			start := nodes[0].FeatureID().Ref()
			// end := nodes[l-1].FeatureID().Ref()
			curr := int64(0)
			e := OSMEdge{}
			for i := 0; i < l; i++ {
				curr = nodes[i].FeatureID().Ref()
				on := osm_nodes.Get(curr)
				e.Nodes.Add(on.Point)
				if on.Count > 1 && curr != start {
					edge_att := decoder.DecodeEdge(tags)
					e.NodeA = index_mapping.Get(start)
					e.NodeB = index_mapping.Get(curr)
					e.Attr = edge_att
					edges.Add(e)
					start = curr
					e = OSMEdge{}
					e.Nodes.Add(on.Point)
				}
			}
		default:
			continue
		}
	}
}

//*******************************************
// osm decoder
//*******************************************

type IOSMDecoder interface {
	IsValidHighway(tags Dict[string, string]) bool
	DecodeNode(tags Dict[string, string]) attr.NodeAttribs
	DecodeEdge(tags Dict[string, string]) attr.EdgeAttribs
}

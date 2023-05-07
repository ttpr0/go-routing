package graph

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func ParseGraph(pbf_file string) *Graph {
	nodes := NewList[OSMNode](10000)
	edges := NewList[OSMEdge](10000)
	index_mapping := NewDict[int64, int](10000)
	ParseOsm(pbf_file, &nodes, &edges, &index_mapping)
	print("edges: ", edges.Length(), ", nodes: ", nodes.Length())
	CalcEdgeWeights(&edges)
	graph := CreateGraph(&nodes, &edges)
	return graph
}

func ParseOsm(filename string, nodes *List[OSMNode], edges *List[OSMEdge], index_mapping *Dict[int64, int]) {
	osm_nodes := NewDict[int64, TempNode](1000)

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := osmpbf.New(context.Background(), file, runtime.GOMAXPROCS(-1))
	InitWayHandler(scanner, &osm_nodes)
	scanner.Close()
	file.Seek(0, 0)
	scanner = osmpbf.New(context.Background(), file, runtime.GOMAXPROCS(-1))
	NodeHandler(scanner, &osm_nodes, nodes, index_mapping)
	scanner.Close()
	file.Seek(0, 0)
	scanner = osmpbf.New(context.Background(), file, runtime.GOMAXPROCS(-1))
	WayHandler(scanner, edges, &osm_nodes, index_mapping)
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

func CreateGraph(osmnodes *List[OSMNode], osmedges *List[OSMEdge]) *Graph {
	nodes := NewList[Node](osmedges.Length())
	node_attributes := NewList[NodeAttributes](osmnodes.Length())
	edges := NewList[Edge](osmedges.Length() * 2)
	edge_attributes := NewList[EdgeAttributes](osmedges.Length() * 2)
	edgerefs := NewList[EdgeRef](osmedges.Length() * 2)
	edge_weights := NewList[int32](osmedges.Length() * 2)
	node_geoms := NewList[geo.Coord](osmnodes.Length())
	edge_geoms := NewList[geo.CoordArray](osmedges.Length() * 2)

	edge_index_mapping := NewDict[int, int](osmedges.Length())
	for i, osmedge := range *osmedges {
		edge := Edge{
			NodeA: int32(osmedge.NodeA),
			NodeB: int32(osmedge.NodeB),
		}
		edge_attrib := EdgeAttributes{
			Type:     osmedge.Type,
			Maxspeed: byte(osmedge.Templimit),
			Length:   osmedge.Length,
		}
		edge_weight := int32(osmedge.Weight)
		edges.Add(edge)
		edge_attributes.Add(edge_attrib)
		edge_weights.Add(edge_weight)
		edge_geoms.Add(geo.CoordArray(osmedge.Nodes))
		edge_index_mapping[i] = edges.Length() - 1
		if !osmedge.Oneway {
			edge = Edge{
				NodeA: int32(osmedge.NodeB),
				NodeB: int32(osmedge.NodeA),
			}
			edge_attrib := EdgeAttributes{
				Type:     osmedge.Type,
				Maxspeed: byte(osmedge.Templimit),
				Length:   osmedge.Length,
			}
			edge_weight := int32(osmedge.Weight)
			edges.Add(edge)
			edge_attributes.Add(edge_attrib)
			edge_weights.Add(edge_weight)
			edge_geoms.Add(geo.CoordArray(osmedge.Nodes))
		}
	}

	for id, osmnode := range *osmnodes {
		node := Node{}
		node_attrib := NodeAttributes{}
		node.EdgeRefStart = int32(edgerefs.Length())
		for _, edgeid := range osmnode.Edges {
			index := edge_index_mapping[int(edgeid)]
			edge := edges[index]
			if edge.NodeA == int32(id) {
				edgeref := EdgeRef{
					EdgeID: int32(index),
					Type:   0,
				}
				edgerefs.Add(edgeref)
			} else if edge.NodeB == int32(id) {
				edgeref := EdgeRef{
					EdgeID: int32(index),
					Type:   1,
				}
				edgerefs.Add(edgeref)
			}
			if index == edges.Length()-1 {
				continue
			}
			edge = edges[index+1]
			if edge.NodeA == int32(id) {
				edgeref := EdgeRef{
					EdgeID: int32(index + 1),
					Type:   0,
				}
				edgerefs.Add(edgeref)
			} else if edge.NodeB == int32(id) {
				edgeref := EdgeRef{
					EdgeID: int32(index + 1),
					Type:   1,
				}
				edgerefs.Add(edgeref)
			}
		}
		node.EdgeRefCount = int16(edgerefs.Length() - int(node.EdgeRefStart))
		nodes.Add(node)
		node_attributes.Add(node_attrib)
		node_geoms.Add(osmnode.Point)
	}

	return &Graph{
		nodes:           nodes,
		node_attributes: node_attributes,
		edge_refs:       edgerefs,
		edges:           edges,
		edge_attributes: edge_attributes,
		geom:            &Geometry{node_geoms, edge_geoms},
		weight:          &Weighting{edge_weights},
	}
}

//*******************************************
// osm handler methods
//*******************************************

func InitWayHandler(scanner *osmpbf.Scanner, osm_nodes *Dict[int64, TempNode]) {
	// i := 0
	types := Dict[string, bool]{"motorway": true, "motorway_link": true, "trunk": true, "trunk_link": true,
		"primary": true, "primary_link": true, "secondary": true, "secondary_link": true, "tertiary": true, "tertiary_link": true,
		"residential": true, "living_street": true, "service": true, "track": true, "unclassified": true, "road": true}

	scanner.SkipNodes = true
	scanner.SkipRelations = true
	for scanner.Scan() {
		switch object := scanner.Object().(type) {
		case *osm.Way:
			tags := Dict[string, string](object.TagMap())
			if !tags.ContainsKey("highway") {
				continue
			}
			if !types.ContainsKey(tags.Get("highway")) {
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

func NodeHandler(scanner *osmpbf.Scanner, osm_nodes *Dict[int64, TempNode], nodes *List[OSMNode], index_mapping *Dict[int64, int]) {
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
			c += 1
			if c%1000 == 0 {
				fmt.Println(c)
			}
			on := osm_nodes.Get(id)
			if on.Count > 1 {
				node := OSMNode{geo.Coord{float32(object.Lon), float32(object.Lat)}, 0, NewList[int32](3)}
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

func WayHandler(scanner *osmpbf.Scanner, edges *List[OSMEdge], osm_nodes *Dict[int64, TempNode], index_mapping *Dict[int64, int]) {
	// i := 0
	types := Dict[string, bool]{"motorway": true, "motorway_link": true, "trunk": true, "trunk_link": true,
		"primary": true, "primary_link": true, "secondary": true, "secondary_link": true, "tertiary": true, "tertiary_link": true,
		"residential": true, "living_street": true, "service": true, "track": true, "unclassified": true, "road": true}
	c := 0

	scanner.SkipNodes = true
	scanner.SkipRelations = true
	for scanner.Scan() {
		switch object := scanner.Object().(type) {
		case *osm.Way:
			tags := Dict[string, string](object.TagMap())
			if !tags.ContainsKey("highway") {
				continue
			}
			if !types.ContainsKey(tags.Get("highway")) {
				continue
			}
			c += 1
			if c%1000 == 0 {
				fmt.Println(c)
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
					templimit := tags.Get("maxspeed")
					str_type := tags.Get("highway")
					oneway := tags.Get("oneway")
					e.Type = GetType(str_type)
					e.Templimit = GetTemplimit(templimit, e.Type)
					e.Oneway = IsOneway(oneway, e.Type)
					e.NodeA = index_mapping.Get(start)
					e.NodeB = index_mapping.Get(curr)
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
// utility methods
//*******************************************

func IsOneway(oneway string, str_type RoadType) bool {
	if str_type == MOTORWAY || str_type == TRUNK || str_type == MOTORWAY_LINK || str_type == TRUNK_LINK {
		return true
	} else if oneway == "yes" {
		return true
	}
	return false
}

func GetType(typ string) RoadType {
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

func GetTemplimit(templimit string, streettype RoadType) int32 {
	var w int32
	if templimit == "" {
		if streettype == MOTORWAY || streettype == TRUNK {
			w = 130
		} else if streettype == MOTORWAY_LINK || streettype == TRUNK_LINK {
			w = 50
		} else if streettype == PRIMARY || streettype == SECONDARY {
			w = 90
		} else if streettype == TERTIARY {
			w = 70
		} else if streettype == PRIMARY_LINK || streettype == SECONDARY_LINK || streettype == TERTIARY_LINK {
			w = 30
		} else if streettype == RESIDENTIAL {
			w = 40
		} else if streettype == LIVING_STREET {
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

func CalcEdgeWeights(edges *List[OSMEdge]) {
	for i := 0; i < edges.Length(); i++ {
		e := edges.Get(i)
		e.Length = float32(geo.HaversineLength(geo.CoordArray(e.Nodes)))
		e.Weight = e.Length * 3.6 / float32(e.Templimit)
		if e.Weight > 255 {
			e.Weight = 255
		}
		if e.Weight < 1 {
			e.Weight = 1
		}
		edges.Set(i, e)
	}
}

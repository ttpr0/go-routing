package graph

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
	"github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

func ParseGraph(pbf_file, out_file string) {
	nodes := NewList[OSMNode](10000)
	edges := NewList[OSMEdge](10000)
	index_mapping := NewDict[int64, int](10000)
	ParseOsm(pbf_file, &nodes, &edges, &index_mapping)
	print("edges: ", edges.Length(), ", nodes: ", nodes.Length())
	CalcEdgeWeights(&edges)
	graph := CreateGraph(&nodes, &edges)
	StoreGraph(graph, out_file)
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

func StoreGraph(graph *Graph, filename string) {
	nodesbuffer := bytes.Buffer{}
	edgesbuffer := bytes.Buffer{}
	geombuffer := bytes.Buffer{}
	nodecount := graph.nodes.Length()
	edgerefcount := graph.edge_refs.Length()
	binary.Write(&nodesbuffer, binary.LittleEndian, int32(nodecount))
	binary.Write(&nodesbuffer, binary.LittleEndian, int32(edgerefcount))
	edgecount := graph.edges.Length()
	binary.Write(&edgesbuffer, binary.LittleEndian, int32(edgecount))
	for i := 0; i < nodecount; i++ {
		node := graph.nodes.Get(i)
		node_attrib := graph.node_attributes.Get(i)
		point := graph.geom.GetNode(int32(i))
		binary.Write(&nodesbuffer, binary.LittleEndian, node_attrib.Type)
		binary.Write(&nodesbuffer, binary.LittleEndian, node.EdgeRefStart)
		binary.Write(&nodesbuffer, binary.LittleEndian, node.EdgeRefCount)
		binary.Write(&geombuffer, binary.LittleEndian, point.Lon)
		binary.Write(&geombuffer, binary.LittleEndian, point.Lat)
	}
	for i := 0; i < edgerefcount; i++ {
		edgeref := graph.edge_refs.Get(i)
		binary.Write(&nodesbuffer, binary.LittleEndian, int32(edgeref.EdgeID))
		binary.Write(&nodesbuffer, binary.LittleEndian, edgeref.Type)
	}
	c := 0
	for i := 0; i < edgecount; i++ {
		edge := graph.edges.Get(i)
		edge_attrib := graph.edge_attributes.Get(i)
		edge_weight := graph.weight.GetEdgeWeight(int32(i))
		binary.Write(&edgesbuffer, binary.LittleEndian, int32(edge.NodeA))
		binary.Write(&edgesbuffer, binary.LittleEndian, int32(edge.NodeB))
		binary.Write(&edgesbuffer, binary.LittleEndian, uint8(edge_weight))
		binary.Write(&edgesbuffer, binary.LittleEndian, byte(edge_attrib.Type))
		binary.Write(&edgesbuffer, binary.LittleEndian, edge_attrib.Length)
		binary.Write(&edgesbuffer, binary.LittleEndian, uint8(edge_attrib.Maxspeed))
		nc := len(graph.geom.GetEdge(int32(i)))
		binary.Write(&geombuffer, binary.LittleEndian, int32(c))
		binary.Write(&geombuffer, binary.LittleEndian, uint8(nc))
		c += nc * 8
	}
	for i := 0; i < edgecount; i++ {
		coords := graph.geom.GetEdge(int32(i))
		for _, coord := range coords {
			binary.Write(&geombuffer, binary.LittleEndian, coord.Lon)
			binary.Write(&geombuffer, binary.LittleEndian, coord.Lat)
		}
	}

	nodesfile, _ := os.Create(filename + "-nodes")
	defer nodesfile.Close()
	nodesfile.Write(nodesbuffer.Bytes())
	edgesfile, _ := os.Create(filename + "-edges")
	defer edgesfile.Close()
	edgesfile.Write(edgesbuffer.Bytes())
	geomfile, _ := os.Create(filename + "-geom")
	defer geomfile.Close()
	geomfile.Write(geombuffer.Bytes())
}

func LoadGraph(file string) IGraph {
	_, err := os.Stat(file + "-nodes")
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file + "-nodes")
	}
	_, err = os.Stat(file + "-edges")
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file + "-edges")
	}
	_, err = os.Stat(file + "-geom")
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file + "-geom")
	}

	nodedata, _ := os.ReadFile(file + "-nodes")
	nodereader := bytes.NewReader(nodedata)
	var nodecount int32
	binary.Read(nodereader, binary.LittleEndian, &nodecount)
	var edgerefcount int32
	binary.Read(nodereader, binary.LittleEndian, &edgerefcount)
	nodes := NewList[Node](int(nodecount))
	node_attribs := NewList[NodeAttributes](int(nodecount))
	edge_refs := NewList[EdgeRef](int(edgerefcount))
	for i := 0; i < int(nodecount); i++ {
		var t int8
		binary.Read(nodereader, binary.LittleEndian, &t)
		var s int32
		binary.Read(nodereader, binary.LittleEndian, &s)
		var c int16
		binary.Read(nodereader, binary.LittleEndian, &c)
		nodes.Add(Node{
			EdgeRefStart: s,
			EdgeRefCount: c,
		})
		node_attribs.Add(NodeAttributes{Type: t})
	}
	for i := 0; i < int(edgerefcount); i++ {
		var id int32
		binary.Read(nodereader, binary.LittleEndian, &id)
		var r byte
		binary.Read(nodereader, binary.LittleEndian, &r)
		edge_refs.Add(EdgeRef{
			EdgeID: id,
			Type:   r,
		})
	}

	edgedata, _ := os.ReadFile(file + "-edges")
	edgereader := bytes.NewReader(edgedata)
	var edgecount int32
	binary.Read(edgereader, binary.LittleEndian, &edgecount)
	edges := NewList[Edge](int(edgecount))
	edge_attribs := NewList[EdgeAttributes](int(edgecount))
	edge_weights := NewList[int32](int(edgecount))
	for i := 0; i < int(edgecount); i++ {
		var a int32
		binary.Read(edgereader, binary.LittleEndian, &a)
		var b int32
		binary.Read(edgereader, binary.LittleEndian, &b)
		var w uint8
		binary.Read(edgereader, binary.LittleEndian, &w)
		var t byte
		binary.Read(edgereader, binary.LittleEndian, &t)
		var l float32
		binary.Read(edgereader, binary.LittleEndian, &l)
		var m uint8
		binary.Read(edgereader, binary.LittleEndian, &m)
		edges.Add(Edge{
			NodeA: a,
			NodeB: b,
		})
		edge_attribs.Add(EdgeAttributes{
			Type:     RoadType(t),
			Length:   l,
			Maxspeed: m,
		})
		edge_weights.Add(int32(w))
	}

	geomdata, _ := os.ReadFile(file + "-geom")
	startindex := nodecount*8 + edgecount*5
	geomreader := bytes.NewReader(geomdata)
	linereader := bytes.NewReader(geomdata[startindex:])
	node_geoms := make([]geo.Coord, nodecount)
	for i := 0; i < int(nodecount); i++ {
		var lon float32
		binary.Read(geomreader, binary.LittleEndian, &lon)
		var lat float32
		binary.Read(geomreader, binary.LittleEndian, &lat)
		node_geoms[i] = geo.Coord{lon, lat}
	}
	edge_geoms := make([]geo.CoordArray, edgecount)
	for i := 0; i < int(edgecount); i++ {
		var s int32
		binary.Read(geomreader, binary.LittleEndian, &s)
		var c byte
		binary.Read(geomreader, binary.LittleEndian, &c)
		points := make([]geo.Coord, c)
		for j := 0; j < int(c); j++ {
			var lon float32
			binary.Read(linereader, binary.LittleEndian, &lon)
			var lat float32
			binary.Read(linereader, binary.LittleEndian, &lat)
			points[j].Lon = lon
			points[j].Lat = lat
		}
		edge_geoms[i] = points
	}

	return &Graph{
		nodes:           nodes,
		node_attributes: node_attribs,
		edge_refs:       edge_refs,
		edges:           edges,
		edge_attributes: edge_attribs,
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
			on.Point.Lon = float32(object.Lon)
			on.Point.Lat = float32(object.Lat)
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

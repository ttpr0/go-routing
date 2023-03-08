package graph

import (
	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type ICHGraph interface {
	GetGeometry() IGeometry
	GetWeighting() IWeighting
	GetShortcutWeighting() IWeighting
	GetOtherNode(edge, node int32) (int32, Direction)
	GetOtherShortcutNode(shortcut, node int32) (int32, Direction)
	GetNodeLevel(node int32) int16
	GetAdjacentEdges(node int32) IIterator[EdgeRef]
	GetAdjacentShortcuts(node int32) IIterator[EdgeRef]
	ForEachEdge(node int32, f func(int32))
	NodeCount() int32
	EdgeCount() int32
	ShortcutCount() int32
	IsNode(node int32) bool
	GetNode(node int32) NodeAttributes
	GetEdge(edge int32) EdgeAttributes
	GetShortcut(shortcut int32) Shortcut
	GetEdgesFromShortcut(edges *List[int32], shortcut_id int32, reversed bool)
}

type CHGraph struct {
	nodes           List[Node]
	node_attributes List[NodeAttributes]
	node_levels     List[int16]
	edge_refs       List[EdgeRef]
	edges           List[Edge]
	edge_attributes List[EdgeAttributes]
	shortcuts       List[Shortcut]
	geom            IGeometry
	weight          IWeighting
	sh_weight       IWeighting
}

func (self *CHGraph) GetGeometry() IGeometry {
	return self.geom
}

func (self *CHGraph) GetWeighting() IWeighting {
	return self.weight
}

func (self *CHGraph) GetShortcutWeighting() IWeighting {
	return self.sh_weight
}

func (self *CHGraph) GetOtherNode(edge int32, node int32) (int32, Direction) {
	e := self.edges[edge]
	if node == e.NodeA {
		return e.NodeB, FORWARD
	}
	if node == e.NodeB {
		return e.NodeA, BACKWARD
	}
	return 0, 0
}

func (self *CHGraph) GetOtherShortcutNode(shortcut int32, node int32) (int32, Direction) {
	e := self.shortcuts[shortcut]
	if node == e.NodeA {
		return e.NodeB, FORWARD
	}
	if node == e.NodeB {
		return e.NodeA, BACKWARD
	}
	return 0, 0
}

func (self *CHGraph) GetNodeLevel(node int32) int16 {
	return self.node_levels[node]
}

func (self *CHGraph) GetAdjacentEdges(node int32) IIterator[EdgeRef] {
	n := self.nodes[node]
	return &EdgeRefIterator{
		state:     int(n.EdgeRefStart),
		end:       int(n.EdgeRefStart) + int(n.EdgeRefCount),
		edge_refs: &self.edge_refs,
	}
}

func (self *CHGraph) GetAdjacentShortcuts(node int32) IIterator[EdgeRef] {
	n := self.nodes[node]
	return &CHEdgeRefIterator{
		state:     int(n.EdgeRefStart),
		end:       int(n.EdgeRefStart) + int(n.EdgeRefCount),
		edge_refs: &self.edge_refs,
	}
}

func (self *CHGraph) ForEachEdge(node int32, f func(int32)) {
	panic("not implemented") // TODO: Implement
}

func (self *CHGraph) NodeCount() int32 {
	return int32(len(self.nodes))
}

func (self *CHGraph) EdgeCount() int32 {
	return int32(len(self.edges))
}

func (self *CHGraph) ShortcutCount() int32 {
	return int32(len(self.shortcuts))
}

func (self *CHGraph) IsNode(node int32) bool {
	if node < int32(len(self.nodes)) {
		return true
	} else {
		return false
	}
}

func (self *CHGraph) GetNode(node int32) NodeAttributes {
	return self.node_attributes[node]
}

func (self *CHGraph) GetEdge(edge int32) EdgeAttributes {
	return self.edge_attributes[edge]
}

func (self *CHGraph) GetShortcut(shortcut int32) Shortcut {
	return self.shortcuts[shortcut]
}

func (self *CHGraph) GetEdgesFromShortcut(edges *List[int32], shortcut_id int32, reversed bool) {
	shortcut := self.GetShortcut(shortcut_id)
	if reversed {
		e := shortcut.Edges[1]
		if e.IsShortcut() {
			self.GetEdgesFromShortcut(edges, e.EdgeID, reversed)
		} else {
			edges.Add(e.EdgeID)
		}
		e = shortcut.Edges[0]
		if e.IsShortcut() {
			self.GetEdgesFromShortcut(edges, e.EdgeID, reversed)
		} else {
			edges.Add(e.EdgeID)
		}
	} else {
		e := shortcut.Edges[0]
		if e.IsShortcut() {
			self.GetEdgesFromShortcut(edges, e.EdgeID, reversed)
		} else {
			edges.Add(e.EdgeID)
		}
		e = shortcut.Edges[1]
		if e.IsShortcut() {
			self.GetEdgesFromShortcut(edges, e.EdgeID, reversed)
		} else {
			edges.Add(e.EdgeID)
		}
	}
}

// func LoadCHGraph(file string) ICHGraph {
// 	file_info, err := os.Stat(file)
// 	if errors.Is(err, os.ErrNotExist) || strings.Split(file_info.Name(), ".")[1] != "graph" {
// 		panic("file not found")
// 	}

// 	graphdata, _ := os.ReadFile(file)
// 	graphreader := bytes.NewReader(graphdata)
// 	var nodecount int32
// 	binary.Read(graphreader, binary.LittleEndian, &nodecount)
// 	var edgecount int32
// 	binary.Read(graphreader, binary.LittleEndian, &edgecount)
// 	startindex := 8 + nodecount*5 + edgecount*8
// 	edgerefreader := bytes.NewReader(graphdata[startindex:])
// 	nodearr := make([]CHNode, nodecount)
// 	for i := 0; i < int(nodecount); i++ {
// 		var s int32
// 		binary.Read(graphreader, binary.LittleEndian, &s)
// 		var c int8
// 		binary.Read(graphreader, binary.LittleEndian, &c)
// 		edges := make([]int32, c)
// 		for j := 0; j < int(c); j++ {
// 			var e int32
// 			binary.Read(edgerefreader, binary.LittleEndian, &e)
// 			edges[j] = e
// 		}
// 		nodearr[i] = CHNode{Edges: edges}
// 	}
// 	edgearr := make([]Edge, edgecount)
// 	for i := 0; i < int(edgecount); i++ {
// 		var start int32
// 		binary.Read(graphreader, binary.LittleEndian, &start)
// 		var end int32
// 		binary.Read(graphreader, binary.LittleEndian, &end)
// 		edgearr[i] = Edge{NodeA: start, NodeB: end}
// 	}

// 	attribdata, _ := os.ReadFile(strings.Replace(file, ".graph", "-attrib", 1))
// 	attribreader := bytes.NewReader(attribdata)
// 	nodeattribarr := make([]NodeAttributes, nodecount)
// 	for i := 0; i < int(nodecount); i++ {
// 		var t int8
// 		binary.Read(attribreader, binary.LittleEndian, &t)
// 		nodeattribarr[i] = NodeAttributes{Type: t}
// 	}
// 	edgeattribarr := make([]EdgeAttributes, edgecount)
// 	for i := 0; i < int(edgecount); i++ {
// 		var t int8
// 		binary.Read(attribreader, binary.LittleEndian, &t)
// 		var length float32
// 		binary.Read(attribreader, binary.LittleEndian, &length)
// 		var maxspeed byte
// 		binary.Read(attribreader, binary.LittleEndian, &maxspeed)
// 		var oneway bool
// 		binary.Read(attribreader, binary.LittleEndian, &oneway)
// 		edgeattribarr[i] = EdgeAttributes{Type: RoadType(t), Length: length, Maxspeed: maxspeed, Oneway: oneway}
// 	}

// 	weightdata, _ := os.ReadFile(strings.Replace(file, ".graph", "-weight", 1))
// 	weightreader := bytes.NewReader(weightdata)
// 	edgeweights := make([]int32, edgecount)
// 	for i := 0; i < int(edgecount); i++ {
// 		var w byte
// 		binary.Read(weightreader, binary.LittleEndian, &w)
// 		edgeweights[i] = int32(w)
// 	}

// 	geomdata, _ := os.ReadFile(strings.Replace(file, ".graph", "-geom", 1))
// 	startindex = nodecount*8 + edgecount*5
// 	geomreader := bytes.NewReader(geomdata)
// 	linereader := bytes.NewReader(geomdata[startindex:])
// 	pointarr := make([]geo.Coord, nodecount)
// 	for i := 0; i < int(nodecount); i++ {
// 		var lon float32
// 		binary.Read(geomreader, binary.LittleEndian, &lon)
// 		var lat float32
// 		binary.Read(geomreader, binary.LittleEndian, &lat)
// 		pointarr[i] = geo.Coord{lon, lat}
// 	}
// 	linearr := make([]geo.CoordArray, edgecount)
// 	for i := 0; i < int(edgecount); i++ {
// 		var s int32
// 		binary.Read(geomreader, binary.LittleEndian, &s)
// 		var c byte
// 		binary.Read(geomreader, binary.LittleEndian, &c)
// 		points := make([]geo.Coord, c)
// 		for j := 0; j < int(c); j++ {
// 			var lon float32
// 			binary.Read(linereader, binary.LittleEndian, &lon)
// 			var lat float32
// 			binary.Read(linereader, binary.LittleEndian, &lat)
// 			points[j].Lon = lon
// 			points[j].Lat = lat
// 		}
// 		linearr[i] = points
// 	}

// 	leveldata, _ := os.ReadFile(strings.Replace(file, ".graph", "-level", 1))
// 	levelreader := bytes.NewReader(leveldata)
// 	for i := 0; i < int(nodecount); i++ {
// 		var l int16
// 		binary.Read(levelreader, binary.LittleEndian, &l)
// 		nodearr[i].Level = l
// 		nodearr[i].Shortcuts = make([]int32, 0, 3)
// 	}

// 	shdata, _ := os.ReadFile(strings.Replace(file, ".graph", "-shortcut", 1))
// 	shreader := bytes.NewReader(shdata)
// 	var shcount int32
// 	binary.Read(shreader, binary.LittleEndian, &shcount)
// 	startindex = shcount * 12
// 	shweights := make([]int32, shcount)
// 	sharr := make([]Shortcut, shcount)
// 	for i := 0; i < int(shcount); i++ {
// 		var start int32
// 		binary.Read(shreader, binary.LittleEndian, &start)
// 		var end int32
// 		binary.Read(shreader, binary.LittleEndian, &end)
// 		var weight int32
// 		binary.Read(shreader, binary.LittleEndian, &weight)
// 		sharr[i] = Shortcut{NodeA: start, NodeB: end, Oneway: false}
// 		shweights[i] = weight
// 		nodearr[start].Shortcuts = append(nodearr[start].Shortcuts, int32(i))
// 		nodearr[end].Shortcuts = append(nodearr[end].Shortcuts, int32(i))
// 	}
// 	for i := 0; i < int(shcount); i++ {
// 		edges := make([]EdgeRef, 2)
// 		for j := 0; j < 2; j++ {
// 			var id int32
// 			binary.Read(shreader, binary.LittleEndian, &id)
// 			var is bool
// 			binary.Read(shreader, binary.LittleEndian, &is)
// 			edges[j] = EdgeRef{id, is}
// 		}
// 		sharr[i].egdes = edges
// 	}

// 	return &CHGraph{
// 		edges:           edgearr,
// 		edge_attributes: edgeattribarr,
// 		nodes:           nodearr,
// 		node_attributes: nodeattribarr,
// 		shortcuts:       sharr,
// 		geom:            &Geometry{pointarr, linearr},
// 		weight:          &Weighting{edgeweights},
// 		sh_weight:       &Weighting{shweights},
// 	}
// }

type CHEdgeRefIterator struct {
	state     int
	end       int
	edge_refs *List[EdgeRef]
}

func (self *CHEdgeRefIterator) Next() (EdgeRef, bool) {
	if self.state == self.end {
		var t EdgeRef
		return t, false
	}
	self.state += 1
	return self.edge_refs.Get(self.state - 1), true
}

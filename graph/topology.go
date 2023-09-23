package graph

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// adjacency interfaces
//*******************************************

type IAdjacency interface {
	GetAccessor() IAdjacencyAccessor
}

type IAdjacencyAccessor interface {
	SetBaseNode(node int32, dir Direction)
	Next() bool
	GetEdgeID() int32
	GetOtherID() int32
	GetType() byte
}

//*******************************************
// adjacency structs
//*******************************************

type _NodeEntry struct {
	FWDEdgeStart int32
	FWDEdgeCount int16
	BWDEdgeStart int32
	BWDEdgeCount int16
}

type _DynamicNodeEntry struct {
	FWDEdges List[_EdgeEntry]
	BWDEdges List[_EdgeEntry]
}

type _EdgeEntry struct {
	EdgeID  int32
	OtherID int32
	Type    byte
}

//*******************************************
// adjacency array implementation
//*******************************************

type AdjacencyArray struct {
	node_entries     Array[_NodeEntry]
	fwd_edge_entries Array[_EdgeEntry]
	bwd_edge_entries Array[_EdgeEntry]
}

// return the node degree for given direction
func (self *AdjacencyArray) GetDegree(index int32, dir Direction) int16 {
	ref := self.node_entries[index]
	if dir == FORWARD {
		return ref.FWDEdgeCount
	} else {
		return ref.BWDEdgeCount
	}
}

func (self *AdjacencyArray) GetAccessor() AdjArrayAccessor {
	return AdjArrayAccessor{
		topology: self,
	}
}

type AdjArrayAccessor struct {
	topology      *AdjacencyArray
	state         int32
	end           int32
	edge_refs     Array[_EdgeEntry]
	curr_edge_id  int32
	curr_other_id int32
	curr_type     byte
}

func (self *AdjArrayAccessor) SetBaseNode(node int32, dir Direction) {
	ref := self.topology.node_entries[node]
	if dir == FORWARD {
		start, count := ref.FWDEdgeStart, int32(ref.FWDEdgeCount)
		self.state = start
		self.end = start + count
		self.edge_refs = self.topology.fwd_edge_entries
	} else {
		start, count := ref.BWDEdgeStart, int32(ref.BWDEdgeCount)
		self.state = start
		self.end = start + count
		self.edge_refs = self.topology.bwd_edge_entries
	}
}
func (self *AdjArrayAccessor) Next() bool {
	if self.state == self.end {
		return false
	}
	ref := self.edge_refs[self.state]
	self.curr_edge_id = ref.EdgeID
	self.curr_other_id = ref.OtherID
	self.curr_type = ref.Type
	self.state += 1
	return true
}
func (self *AdjArrayAccessor) GetEdgeID() int32 {
	return self.curr_edge_id
}
func (self *AdjArrayAccessor) GetOtherID() int32 {
	return self.curr_other_id
}
func (self *AdjArrayAccessor) GetType() byte {
	return self.curr_type
}

//*******************************************
// adjacency list implementation
//*******************************************

type AdjacencyList struct {
	node_entries Array[_DynamicNodeEntry]
}

func NewAdjacencyList(node_count int) AdjacencyList {
	topology := NewArray[_DynamicNodeEntry](node_count)

	for i := 0; i < node_count; i++ {
		topology[i] = _DynamicNodeEntry{
			FWDEdges: NewList[_EdgeEntry](4),
			BWDEdges: NewList[_EdgeEntry](4),
		}
	}

	return AdjacencyList{
		node_entries: topology,
	}
}

// return the node degree for given direction
func (self *AdjacencyList) GetDegree(index int32, dir Direction) int16 {
	ref := self.node_entries[index]
	if dir == FORWARD {
		return int16(ref.FWDEdges.Length())
	} else {
		return int16(ref.BWDEdges.Length())
	}
}
func (self *AdjacencyList) AddNodeEntry() {
	nodes := List[_DynamicNodeEntry](self.node_entries)
	nodes.Add(_DynamicNodeEntry{
		FWDEdges: NewList[_EdgeEntry](4),
		BWDEdges: NewList[_EdgeEntry](4),
	})
	self.node_entries = Array[_DynamicNodeEntry](nodes)
}
func (self *AdjacencyList) AddEdgeEntries(node_a, node_b, edge_id int32, edge_typ byte) {
	fwd_edges := self.node_entries[node_a].FWDEdges
	fwd_edges.Add(_EdgeEntry{
		EdgeID:  edge_id,
		OtherID: node_b,
		Type:    edge_typ,
	})
	self.node_entries[node_a].FWDEdges = fwd_edges
	bwd_edges := self.node_entries[node_b].BWDEdges
	bwd_edges.Add(_EdgeEntry{
		EdgeID:  edge_id,
		OtherID: node_a,
		Type:    edge_typ,
	})
	self.node_entries[node_b].BWDEdges = bwd_edges
}

// adds forward entry to adjacency
//
// refers to edge between node_a and node_b, entry will be added at node_a
func (self *AdjacencyList) AddFWDEntry(node_a, node_b, edge_id int32, edge_typ byte) {
	fwd_edges := self.node_entries[node_a].FWDEdges
	fwd_edges.Add(_EdgeEntry{
		EdgeID:  edge_id,
		OtherID: node_b,
		Type:    edge_typ,
	})
	self.node_entries[node_a].FWDEdges = fwd_edges
}

// adds backward entry to adjacency
//
// refers to edge between node_a and node_b, entry will be added at node_b
func (self *AdjacencyList) AddBWDEntry(node_a, node_b, edge_id int32, edge_typ byte) {
	bwd_edges := self.node_entries[node_b].BWDEdges
	bwd_edges.Add(_EdgeEntry{
		EdgeID:  edge_id,
		OtherID: node_a,
		Type:    edge_typ,
	})
	self.node_entries[node_b].BWDEdges = bwd_edges
}
func (self *AdjacencyList) GetAccessor() AdjListAccessor {
	return AdjListAccessor{
		topology: self,
	}
}

type AdjListAccessor struct {
	topology      *AdjacencyList
	state         int32
	end           int32
	edge_refs     List[_EdgeEntry]
	curr_edge_id  int32
	curr_other_id int32
	curr_type     byte
}

func (self *AdjListAccessor) SetBaseNode(node int32, dir Direction) {
	ref := self.topology.node_entries[node]
	if dir == FORWARD {
		self.state = 0
		self.end = int32(len(ref.FWDEdges))
		self.edge_refs = ref.FWDEdges
	} else {
		self.state = 0
		self.end = int32(len(ref.BWDEdges))
		self.edge_refs = ref.BWDEdges
	}
}
func (self *AdjListAccessor) Next() bool {
	if self.state == self.end {
		return false
	}
	ref := self.edge_refs[self.state]
	self.curr_edge_id = ref.EdgeID
	self.curr_other_id = ref.OtherID
	self.curr_type = ref.Type
	self.state += 1
	return true
}
func (self *AdjListAccessor) GetEdgeID() int32 {
	return self.curr_edge_id
}
func (self *AdjListAccessor) GetOtherID() int32 {
	return self.curr_other_id
}
func (self *AdjListAccessor) Type() byte {
	return self.curr_type
}

//*******************************************
// utility methods on topology stores
//*******************************************

func _StoreUntypedAdjacency(store *AdjacencyArray, filename string) {
	topologybuffer := bytes.Buffer{}

	fwd_edgerefcount := store.fwd_edge_entries.Length()
	bwd_edgerefcount := store.bwd_edge_entries.Length()
	binary.Write(&topologybuffer, binary.LittleEndian, int32(fwd_edgerefcount))
	binary.Write(&topologybuffer, binary.LittleEndian, int32(bwd_edgerefcount))

	for i := 0; i < store.node_entries.Length(); i++ {
		node_ref := store.node_entries.Get(i)
		binary.Write(&topologybuffer, binary.LittleEndian, node_ref.FWDEdgeStart)
		binary.Write(&topologybuffer, binary.LittleEndian, node_ref.FWDEdgeCount)
		binary.Write(&topologybuffer, binary.LittleEndian, node_ref.BWDEdgeStart)
		binary.Write(&topologybuffer, binary.LittleEndian, node_ref.BWDEdgeCount)
	}
	for i := 0; i < fwd_edgerefcount; i++ {
		edgeref := store.fwd_edge_entries.Get(i)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref.EdgeID)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref.OtherID)
	}
	for i := 0; i < bwd_edgerefcount; i++ {
		edgeref := store.bwd_edge_entries.Get(i)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref.EdgeID)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref.OtherID)
	}

	topologyfile, _ := os.Create(filename)
	defer topologyfile.Close()
	topologyfile.Write(topologybuffer.Bytes())
}

func _LoadUntypedAdjacency(file string, nodecount int) *AdjacencyArray {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	topologydata, _ := os.ReadFile(file)
	topologyreader := bytes.NewReader(topologydata)
	var fwd_edgerefcount int32
	binary.Read(topologyreader, binary.LittleEndian, &fwd_edgerefcount)
	var bwd_edgerefcount int32
	binary.Read(topologyreader, binary.LittleEndian, &bwd_edgerefcount)
	node_refs := NewList[_NodeEntry](int(nodecount))
	fwd_edge_refs := NewList[_EdgeEntry](int(fwd_edgerefcount))
	bwd_edge_refs := NewList[_EdgeEntry](int(bwd_edgerefcount))
	for i := 0; i < int(nodecount); i++ {
		var s1 int32
		binary.Read(topologyreader, binary.LittleEndian, &s1)
		var c1 int16
		binary.Read(topologyreader, binary.LittleEndian, &c1)
		var s2 int32
		binary.Read(topologyreader, binary.LittleEndian, &s2)
		var c2 int16
		binary.Read(topologyreader, binary.LittleEndian, &c2)
		node_refs.Add(_NodeEntry{
			FWDEdgeStart: s1,
			FWDEdgeCount: c1,
			BWDEdgeStart: s2,
			BWDEdgeCount: c2,
		})
	}
	for i := 0; i < int(fwd_edgerefcount); i++ {
		var id int32
		binary.Read(topologyreader, binary.LittleEndian, &id)
		var nid int32
		binary.Read(topologyreader, binary.LittleEndian, &nid)
		fwd_edge_refs.Add(_EdgeEntry{
			EdgeID:  id,
			OtherID: nid,
			Type:    0,
		})
	}
	for i := 0; i < int(bwd_edgerefcount); i++ {
		var id int32
		binary.Read(topologyreader, binary.LittleEndian, &id)
		var nid int32
		binary.Read(topologyreader, binary.LittleEndian, &nid)
		bwd_edge_refs.Add(_EdgeEntry{
			EdgeID:  id,
			OtherID: nid,
			Type:    0,
		})
	}

	return &AdjacencyArray{
		node_entries:     Array[_NodeEntry](node_refs),
		fwd_edge_entries: Array[_EdgeEntry](fwd_edge_refs),
		bwd_edge_entries: Array[_EdgeEntry](bwd_edge_refs),
	}
}

// reorders nodes in topologystore,
// mapping: old id -> new id
func (self *AdjacencyArray) _ReorderNodes(mapping Array[int32]) {
	node_refs := NewArray[_NodeEntry](self.node_entries.Length())
	for i, id := range mapping {
		node_refs[id] = self.node_entries[i]
	}

	fwd_edge_refs := NewList[_EdgeEntry](self.fwd_edge_entries.Length())
	bwd_edge_refs := NewList[_EdgeEntry](self.bwd_edge_entries.Length())

	fwd_start := 0
	bwd_start := 0
	for i := 0; i < node_refs.Length(); i++ {
		node_ref := node_refs[i]
		fwd_count := 0
		fwd_edges := self.fwd_edge_entries[node_ref.FWDEdgeStart : node_ref.FWDEdgeStart+int32(node_ref.FWDEdgeCount)]
		for _, ref := range fwd_edges {
			ref.OtherID = mapping[ref.OtherID]
			fwd_edge_refs.Add(ref)
			fwd_count += 1
		}

		bwd_count := 0
		bwd_edges := self.bwd_edge_entries[node_ref.BWDEdgeStart : node_ref.BWDEdgeStart+int32(node_ref.BWDEdgeCount)]
		for _, ref := range bwd_edges {
			ref.OtherID = mapping[ref.OtherID]
			bwd_edge_refs.Add(ref)
			bwd_count += 1
		}

		node_refs[i] = _NodeEntry{
			FWDEdgeStart: int32(fwd_start),
			FWDEdgeCount: int16(fwd_count),
			BWDEdgeStart: int32(bwd_start),
			BWDEdgeCount: int16(bwd_count),
		}

		fwd_start += fwd_count
		bwd_start += bwd_count
	}
	self.node_entries = Array[_NodeEntry](node_refs)
	self.fwd_edge_entries = Array[_EdgeEntry](fwd_edge_refs)
	self.bwd_edge_entries = Array[_EdgeEntry](bwd_edge_refs)
}

func _StoreTypedAdjacency(store *AdjacencyArray, filename string) {
	topologybuffer := bytes.Buffer{}

	fwd_edgerefcount := store.fwd_edge_entries.Length()
	bwd_edgerefcount := store.bwd_edge_entries.Length()
	binary.Write(&topologybuffer, binary.LittleEndian, int32(fwd_edgerefcount))
	binary.Write(&topologybuffer, binary.LittleEndian, int32(bwd_edgerefcount))

	for i := 0; i < store.node_entries.Length(); i++ {
		node_ref := store.node_entries.Get(i)
		binary.Write(&topologybuffer, binary.LittleEndian, node_ref.FWDEdgeStart)
		binary.Write(&topologybuffer, binary.LittleEndian, node_ref.FWDEdgeCount)
		binary.Write(&topologybuffer, binary.LittleEndian, node_ref.BWDEdgeStart)
		binary.Write(&topologybuffer, binary.LittleEndian, node_ref.BWDEdgeCount)
	}
	for i := 0; i < fwd_edgerefcount; i++ {
		edgeref := store.fwd_edge_entries.Get(i)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref.EdgeID)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref.Type)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref.OtherID)
	}
	for i := 0; i < bwd_edgerefcount; i++ {
		edgeref := store.bwd_edge_entries.Get(i)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref.EdgeID)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref.Type)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref.OtherID)
	}

	topologyfile, _ := os.Create(filename)
	defer topologyfile.Close()
	topologyfile.Write(topologybuffer.Bytes())
}

func _LoadTypedAdjacency(file string, nodecount int) *AdjacencyArray {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	topologydata, _ := os.ReadFile(file)
	topologyreader := bytes.NewReader(topologydata)
	var fwd_edgerefcount int32
	binary.Read(topologyreader, binary.LittleEndian, &fwd_edgerefcount)
	var bwd_edgerefcount int32
	binary.Read(topologyreader, binary.LittleEndian, &bwd_edgerefcount)
	node_refs := NewList[_NodeEntry](int(nodecount))
	fwd_edge_refs := NewList[_EdgeEntry](int(fwd_edgerefcount))
	bwd_edge_refs := NewList[_EdgeEntry](int(bwd_edgerefcount))
	for i := 0; i < int(nodecount); i++ {
		var s1 int32
		binary.Read(topologyreader, binary.LittleEndian, &s1)
		var c1 int16
		binary.Read(topologyreader, binary.LittleEndian, &c1)
		var s2 int32
		binary.Read(topologyreader, binary.LittleEndian, &s2)
		var c2 int16
		binary.Read(topologyreader, binary.LittleEndian, &c2)
		node_refs.Add(_NodeEntry{
			FWDEdgeStart: s1,
			FWDEdgeCount: c1,
			BWDEdgeStart: s2,
			BWDEdgeCount: c2,
		})
	}
	for i := 0; i < int(fwd_edgerefcount); i++ {
		var id int32
		binary.Read(topologyreader, binary.LittleEndian, &id)
		var t byte
		binary.Read(topologyreader, binary.LittleEndian, &t)
		var nid int32
		binary.Read(topologyreader, binary.LittleEndian, &nid)
		fwd_edge_refs.Add(_EdgeEntry{
			EdgeID:  id,
			Type:    t,
			OtherID: nid,
		})
	}
	for i := 0; i < int(bwd_edgerefcount); i++ {
		var id int32
		binary.Read(topologyreader, binary.LittleEndian, &id)
		var t byte
		binary.Read(topologyreader, binary.LittleEndian, &t)
		var nid int32
		binary.Read(topologyreader, binary.LittleEndian, &nid)
		bwd_edge_refs.Add(_EdgeEntry{
			EdgeID:  id,
			Type:    t,
			OtherID: nid,
		})
	}

	return &AdjacencyArray{
		node_entries:     Array[_NodeEntry](node_refs),
		fwd_edge_entries: Array[_EdgeEntry](fwd_edge_refs),
		bwd_edge_entries: Array[_EdgeEntry](bwd_edge_refs),
	}
}

func AdjacencyListToArray(dyn *AdjacencyList) *AdjacencyArray {
	node_refs := NewList[_NodeEntry](dyn.node_entries.Length())
	fwd_edge_refs := NewList[_EdgeEntry](dyn.node_entries.Length())
	bwd_edge_refs := NewList[_EdgeEntry](dyn.node_entries.Length())

	fwd_start := 0
	bwd_start := 0
	for i := 0; i < dyn.node_entries.Length(); i++ {
		fwd_count := 0
		bwd_count := 0

		fwd_refs := dyn.node_entries[i].FWDEdges
		bwd_refs := dyn.node_entries[i].BWDEdges

		for _, ref := range fwd_refs {
			fwd_edge_refs.Add(_EdgeEntry{EdgeID: ref.EdgeID, OtherID: ref.OtherID, Type: ref.Type})
			fwd_count += 1
		}
		for _, ref := range bwd_refs {
			bwd_edge_refs.Add(_EdgeEntry{EdgeID: ref.EdgeID, OtherID: ref.OtherID, Type: ref.Type})
			bwd_count += 1
		}

		node_refs.Add(_NodeEntry{
			FWDEdgeStart: int32(fwd_start),
			FWDEdgeCount: int16(fwd_count),
			BWDEdgeStart: int32(bwd_start),
			BWDEdgeCount: int16(bwd_count),
		})
		fwd_start += fwd_count
		bwd_start += bwd_count
	}

	return &AdjacencyArray{
		node_entries:     Array[_NodeEntry](node_refs),
		fwd_edge_entries: Array[_EdgeEntry](fwd_edge_refs),
		bwd_edge_entries: Array[_EdgeEntry](bwd_edge_refs),
	}
}

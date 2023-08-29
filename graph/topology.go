package graph

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// topology structs
//*******************************************

type _NodeEntry struct {
	FWDEdgeStart int32
	FWDEdgeCount int16
	BWDEdgeStart int32
	BWDEdgeCount int16
}

type _DynamicNodeEntry struct {
	FWDEdges List[_TypedEdgeEntry]
	BWDEdges List[_TypedEdgeEntry]
}

type _EdgeEntry struct {
	EdgeID  int32
	OtherID int32
}

type _TypedEdgeEntry struct {
	EdgeID  int32
	OtherID int32
	Type    byte
}

//*******************************************
// topology store
//*******************************************

type TopologyStore struct {
	node_entries     Array[_NodeEntry]
	fwd_edge_entries Array[_EdgeEntry]
	bwd_edge_entries Array[_EdgeEntry]
}

// return the node degree for given direction
func (self *TopologyStore) GetDegree(node int32, dir Direction) int16 {
	ref := self.node_entries[node]
	if dir == FORWARD {
		return ref.FWDEdgeCount
	} else {
		return ref.BWDEdgeCount
	}
}
func (self *TopologyStore) GetAccessor() TopologyAccessor {
	return TopologyAccessor{
		topology: self,
	}
}

type TypedTopologyStore struct {
	node_entries     Array[_NodeEntry]
	fwd_edge_entries Array[_TypedEdgeEntry]
	bwd_edge_entries Array[_TypedEdgeEntry]
}

// return the node degree for given direction
func (self *TypedTopologyStore) GetDegree(index int32, dir Direction) int16 {
	ref := self.node_entries[index]
	if dir == FORWARD {
		return ref.FWDEdgeCount
	} else {
		return ref.BWDEdgeCount
	}
}

func (self *TypedTopologyStore) GetAccessor() TypedTopologyAccessor {
	return TypedTopologyAccessor{
		topology: self,
	}
}

type DynamicTopologyStore struct {
	node_entries Array[_DynamicNodeEntry]
}

func NewDynamicTopology(node_count int) DynamicTopologyStore {
	topology := NewArray[_DynamicNodeEntry](node_count)

	for i := 0; i < node_count; i++ {
		topology[i] = _DynamicNodeEntry{
			FWDEdges: NewList[_TypedEdgeEntry](4),
			BWDEdges: NewList[_TypedEdgeEntry](4),
		}
	}

	return DynamicTopologyStore{
		node_entries: topology,
	}
}

// return the node degree for given direction
func (self *DynamicTopologyStore) GetDegree(index int32, dir Direction) int16 {
	ref := self.node_entries[index]
	if dir == FORWARD {
		return int16(ref.FWDEdges.Length())
	} else {
		return int16(ref.BWDEdges.Length())
	}
}

func (self *DynamicTopologyStore) AddEdgeEntries(node_a, node_b, edge_id int32, edge_typ byte) {
	fwd_edges := self.node_entries[node_a].FWDEdges
	fwd_edges.Add(_TypedEdgeEntry{
		EdgeID:  edge_id,
		OtherID: node_b,
		Type:    edge_typ,
	})
	self.node_entries[node_a].FWDEdges = fwd_edges
	bwd_edges := self.node_entries[node_b].BWDEdges
	bwd_edges.Add(_TypedEdgeEntry{
		EdgeID:  edge_id,
		OtherID: node_a,
		Type:    edge_typ,
	})
	self.node_entries[node_b].BWDEdges = bwd_edges
}
func (self *DynamicTopologyStore) GetAccessor() DynamicTopologyAccessor {
	return DynamicTopologyAccessor{
		topology: self,
	}
}

//*******************************************
// utility methods on topology stores
//*******************************************

// reorders nodes in topologystore,
// mapping: old id -> new id
func (self *TopologyStore) _ReorderNodes(mapping Array[int32]) {
	node_refs := NewArray[_NodeEntry](self.node_entries.Length())
	for i, id := range mapping {
		node_refs[id] = self.node_entries[i]
	}
	self.node_entries = Array[_NodeEntry](node_refs)

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
	self.fwd_edge_entries = Array[_EdgeEntry](fwd_edge_refs)
	self.bwd_edge_entries = Array[_EdgeEntry](bwd_edge_refs)
}

func (self *TopologyStore) _Store(filename string) {
	topologybuffer := bytes.Buffer{}

	fwd_edgerefcount := self.fwd_edge_entries.Length()
	bwd_edgerefcount := self.bwd_edge_entries.Length()
	binary.Write(&topologybuffer, binary.LittleEndian, int32(fwd_edgerefcount))
	binary.Write(&topologybuffer, binary.LittleEndian, int32(bwd_edgerefcount))

	for i := 0; i < self.node_entries.Length(); i++ {
		node_ref := self.node_entries.Get(i)
		binary.Write(&topologybuffer, binary.LittleEndian, node_ref.FWDEdgeStart)
		binary.Write(&topologybuffer, binary.LittleEndian, node_ref.FWDEdgeCount)
		binary.Write(&topologybuffer, binary.LittleEndian, node_ref.BWDEdgeStart)
		binary.Write(&topologybuffer, binary.LittleEndian, node_ref.BWDEdgeCount)
	}
	for i := 0; i < fwd_edgerefcount; i++ {
		edgeref := self.fwd_edge_entries.Get(i)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref.EdgeID)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref.OtherID)
	}
	for i := 0; i < bwd_edgerefcount; i++ {
		edgeref := self.bwd_edge_entries.Get(i)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref.EdgeID)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref.OtherID)
	}

	topologyfile, _ := os.Create(filename)
	defer topologyfile.Close()
	topologyfile.Write(topologybuffer.Bytes())
}

func _LoadTopologyStore(file string, nodecount int) *TopologyStore {
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
		})
	}

	return &TopologyStore{
		node_entries:     Array[_NodeEntry](node_refs),
		fwd_edge_entries: Array[_EdgeEntry](fwd_edge_refs),
		bwd_edge_entries: Array[_EdgeEntry](bwd_edge_refs),
	}
}

// reorders nodes in topologystore,
// mapping: old id -> new id
func _ReorderTypedTopology(store *TypedTopologyStore, mapping Array[int32]) {
	node_refs := NewArray[_NodeEntry](store.node_entries.Length())
	for i, id := range mapping {
		node_refs[id] = store.node_entries[i]
	}
	store.node_entries = Array[_NodeEntry](node_refs)

	fwd_edge_refs := NewList[_TypedEdgeEntry](store.fwd_edge_entries.Length())
	bwd_edge_refs := NewList[_TypedEdgeEntry](store.bwd_edge_entries.Length())

	fwd_start := 0
	bwd_start := 0
	for i := 0; i < node_refs.Length(); i++ {
		node_ref := node_refs[i]
		fwd_count := 0
		fwd_edges := store.fwd_edge_entries[node_ref.FWDEdgeStart : node_ref.FWDEdgeStart+int32(node_ref.FWDEdgeCount)]
		for _, ref := range fwd_edges {
			ref.OtherID = mapping[ref.OtherID]
			fwd_edge_refs.Add(ref)
			fwd_count += 1
		}

		bwd_count := 0
		bwd_edges := store.bwd_edge_entries[node_ref.BWDEdgeStart : node_ref.BWDEdgeStart+int32(node_ref.BWDEdgeCount)]
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
	store.fwd_edge_entries = Array[_TypedEdgeEntry](fwd_edge_refs)
	store.bwd_edge_entries = Array[_TypedEdgeEntry](bwd_edge_refs)
}

func _StoreTypedTopology(store *TypedTopologyStore, filename string) {
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

func _LoadTypedTopology(file string, nodecount int) *TypedTopologyStore {
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
	fwd_edge_refs := NewList[_TypedEdgeEntry](int(fwd_edgerefcount))
	bwd_edge_refs := NewList[_TypedEdgeEntry](int(bwd_edgerefcount))
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
		fwd_edge_refs.Add(_TypedEdgeEntry{
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
		bwd_edge_refs.Add(_TypedEdgeEntry{
			EdgeID:  id,
			Type:    t,
			OtherID: nid,
		})
	}

	return &TypedTopologyStore{
		node_entries:     Array[_NodeEntry](node_refs),
		fwd_edge_entries: Array[_TypedEdgeEntry](fwd_edge_refs),
		bwd_edge_entries: Array[_TypedEdgeEntry](bwd_edge_refs),
	}
}

func DynamicToTopology(dyn *DynamicTopologyStore) *TopologyStore {
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
			fwd_edge_refs.Add(_EdgeEntry{EdgeID: ref.EdgeID, OtherID: ref.OtherID})
			fwd_count += 1
		}
		for _, ref := range bwd_refs {
			bwd_edge_refs.Add(_EdgeEntry{EdgeID: ref.EdgeID, OtherID: ref.OtherID})
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

	return &TopologyStore{
		node_entries:     Array[_NodeEntry](node_refs),
		fwd_edge_entries: Array[_EdgeEntry](fwd_edge_refs),
		bwd_edge_entries: Array[_EdgeEntry](bwd_edge_refs),
	}
}

func DynamicToTypedTopology(dyn *DynamicTopologyStore) *TypedTopologyStore {
	node_refs := NewList[_NodeEntry](dyn.node_entries.Length())
	fwd_edge_refs := NewList[_TypedEdgeEntry](dyn.node_entries.Length())
	bwd_edge_refs := NewList[_TypedEdgeEntry](dyn.node_entries.Length())

	fwd_start := 0
	bwd_start := 0
	for i := 0; i < dyn.node_entries.Length(); i++ {
		fwd_count := 0
		bwd_count := 0

		fwd_refs := dyn.node_entries[i].FWDEdges
		bwd_refs := dyn.node_entries[i].BWDEdges

		for _, ref := range fwd_refs {
			fwd_edge_refs.Add(_TypedEdgeEntry{EdgeID: ref.EdgeID, OtherID: ref.OtherID, Type: ref.Type})
			fwd_count += 1
		}
		for _, ref := range bwd_refs {
			bwd_edge_refs.Add(_TypedEdgeEntry{EdgeID: ref.EdgeID, OtherID: ref.OtherID, Type: ref.Type})
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

	return &TypedTopologyStore{
		node_entries:     Array[_NodeEntry](node_refs),
		fwd_edge_entries: Array[_TypedEdgeEntry](fwd_edge_refs),
		bwd_edge_entries: Array[_TypedEdgeEntry](bwd_edge_refs),
	}
}

//*******************************************
// topology accessor
//*******************************************

type TopologyAccessor struct {
	topology      *TopologyStore
	state         int32
	end           int32
	edge_refs     Array[_EdgeEntry]
	curr_edge_id  int32
	curr_other_id int32
}

func (self *TopologyAccessor) SetBaseNode(node int32, dir Direction) {
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
func (self *TopologyAccessor) Next() bool {
	if self.state == self.end {
		return false
	}
	ref := self.edge_refs[self.state]
	self.curr_edge_id = ref.EdgeID
	self.curr_other_id = ref.OtherID
	self.state += 1
	return true
}
func (self *TopologyAccessor) GetEdgeID() int32 {
	return self.curr_edge_id
}
func (self *TopologyAccessor) GetOtherID() int32 {
	return self.curr_other_id
}
func (self *TopologyAccessor) HasNext() bool {
	if self.state == self.end {
		return false
	}
	return true
}

type TypedTopologyAccessor struct {
	topology      *TypedTopologyStore
	state         int32
	end           int32
	edge_refs     Array[_TypedEdgeEntry]
	curr_edge_id  int32
	curr_other_id int32
	curr_type     byte
}

func (self *TypedTopologyAccessor) SetBaseNode(node int32, dir Direction) {
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
func (self *TypedTopologyAccessor) Next() bool {
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
func (self *TypedTopologyAccessor) GetEdgeID() int32 {
	return self.curr_edge_id
}
func (self *TypedTopologyAccessor) GetOtherID() int32 {
	return self.curr_other_id
}
func (self *TypedTopologyAccessor) GetType() byte {
	return self.curr_type
}

type DynamicTopologyAccessor struct {
	topology      *DynamicTopologyStore
	state         int32
	end           int32
	edge_refs     List[_TypedEdgeEntry]
	curr_edge_id  int32
	curr_other_id int32
	curr_type     byte
}

func (self *DynamicTopologyAccessor) SetBaseNode(node int32, dir Direction) {
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
func (self *DynamicTopologyAccessor) Next() bool {
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
func (self *DynamicTopologyAccessor) GetEdgeID() int32 {
	return self.curr_edge_id
}
func (self *DynamicTopologyAccessor) GetOtherID() int32 {
	return self.curr_other_id
}
func (self *DynamicTopologyAccessor) Type() byte {
	return self.curr_type
}

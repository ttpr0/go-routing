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

type _EdgeEntry struct {
	EdgeID  int32
	Type    byte
	OtherID int32
}

//*******************************************
// topology store
//*******************************************

type TopologyStore struct {
	node_entries     Array[_NodeEntry]
	fwd_edge_entries Array[_EdgeEntry]
	bwd_edge_entries Array[_EdgeEntry]
}

func (self *TopologyStore) GetNodeRef(index int32, dir Direction) (int32, int16) {
	ref := self.node_entries[index]
	if dir == FORWARD {
		return ref.FWDEdgeStart, ref.FWDEdgeCount
	} else {
		return ref.BWDEdgeStart, ref.BWDEdgeCount
	}
}
func (self *TopologyStore) GetEdgeRefs(dir Direction) Array[_EdgeEntry] {
	if dir == FORWARD {
		return self.fwd_edge_entries
	} else {
		return self.bwd_edge_entries
	}
}

// used in ch preprocessing
func (self *TopologyStore) GetAdjacentEdgeRefs(node int32, dir Direction) List[_EdgeEntry] {
	if dir == FORWARD {
		ref := self.node_entries[node]
		return List[_EdgeEntry](self.fwd_edge_entries[ref.FWDEdgeStart : ref.FWDEdgeStart+int32(ref.FWDEdgeCount)])
	} else {
		ref := self.node_entries[node]
		return List[_EdgeEntry](self.bwd_edge_entries[ref.BWDEdgeStart : ref.BWDEdgeStart+int32(ref.BWDEdgeCount)])
	}
}

func (self *TopologyStore) GetAccessor() TopologyAccessor {
	return TopologyAccessor{
		topology: self,
	}
}

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
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref.Type)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref.OtherID)
	}
	for i := 0; i < bwd_edgerefcount; i++ {
		edgeref := self.bwd_edge_entries.Get(i)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref.EdgeID)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref.Type)
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

	return &TopologyStore{
		node_entries:     Array[_NodeEntry](node_refs),
		fwd_edge_entries: Array[_EdgeEntry](fwd_edge_refs),
		bwd_edge_entries: Array[_EdgeEntry](bwd_edge_refs),
	}
}

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

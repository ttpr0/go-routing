package graph

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type TopologyStore struct {
	node_refs     List[NodeRef]
	fwd_edge_refs List[EdgeRef]
	bwd_edge_refs List[EdgeRef]
}

func (self *TopologyStore) GetNodeRef(index int32, dir Direction) (int32, int16) {
	ref := self.node_refs[index]
	if dir == FORWARD {
		return ref.EdgeRefFWDStart, ref.EdgeRefFWDCount
	} else {
		return ref.EdgeRefBWDStart, ref.EdgeRefBWDCount
	}
}
func (self *TopologyStore) GetEdgeRefs(dir Direction) Array[EdgeRef] {
	if dir == FORWARD {
		return Array[EdgeRef](self.fwd_edge_refs)
	} else {
		return Array[EdgeRef](self.bwd_edge_refs)
	}
}

// used in ch preprocessing
func (self *TopologyStore) GetAdjacentEdgeRefs(node int32, dir Direction) List[EdgeRef] {
	if dir == FORWARD {
		ref := self.node_refs[node]
		return self.fwd_edge_refs[ref.EdgeRefFWDStart : ref.EdgeRefFWDStart+int32(ref.EdgeRefFWDCount)]
	} else {
		ref := self.node_refs[node]
		return self.bwd_edge_refs[ref.EdgeRefBWDStart : ref.EdgeRefBWDStart+int32(ref.EdgeRefBWDCount)]
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
	node_refs := NewArray[NodeRef](self.node_refs.Length())
	for i, id := range mapping {
		node_refs[id] = self.node_refs[i]
	}
	self.node_refs = List[NodeRef](node_refs)

	fwd_edge_refs := NewList[EdgeRef](self.fwd_edge_refs.Length())
	bwd_edge_refs := NewList[EdgeRef](self.bwd_edge_refs.Length())

	fwd_start := 0
	bwd_start := 0
	for i := 0; i < node_refs.Length(); i++ {
		node_ref := node_refs[i]
		fwd_count := 0
		fwd_edges := self.fwd_edge_refs[node_ref.EdgeRefFWDStart : node_ref.EdgeRefFWDStart+int32(node_ref.EdgeRefFWDCount)]
		for _, ref := range fwd_edges {
			ref.OtherID = mapping[ref.OtherID]
			fwd_edge_refs.Add(ref)
			fwd_count += 1
		}

		bwd_count := 0
		bwd_edges := self.bwd_edge_refs[node_ref.EdgeRefBWDStart : node_ref.EdgeRefBWDStart+int32(node_ref.EdgeRefBWDCount)]
		for _, ref := range bwd_edges {
			ref.OtherID = mapping[ref.OtherID]
			bwd_edge_refs.Add(ref)
			bwd_count += 1
		}

		node_refs[i] = NodeRef{
			EdgeRefFWDStart: int32(fwd_start),
			EdgeRefFWDCount: int16(fwd_count),
			EdgeRefBWDStart: int32(bwd_start),
			EdgeRefBWDCount: int16(bwd_count),
		}

		fwd_start += fwd_count
		bwd_start += bwd_count
	}
	self.fwd_edge_refs = fwd_edge_refs
	self.bwd_edge_refs = bwd_edge_refs
}

func (self *TopologyStore) _Store(filename string) {
	topologybuffer := bytes.Buffer{}

	fwd_edgerefcount := self.fwd_edge_refs.Length()
	bwd_edgerefcount := self.bwd_edge_refs.Length()
	binary.Write(&topologybuffer, binary.LittleEndian, int32(fwd_edgerefcount))
	binary.Write(&topologybuffer, binary.LittleEndian, int32(bwd_edgerefcount))

	for i := 0; i < self.node_refs.Length(); i++ {
		node_ref := self.node_refs.Get(i)
		binary.Write(&topologybuffer, binary.LittleEndian, node_ref.EdgeRefFWDStart)
		binary.Write(&topologybuffer, binary.LittleEndian, node_ref.EdgeRefFWDCount)
		binary.Write(&topologybuffer, binary.LittleEndian, node_ref.EdgeRefBWDStart)
		binary.Write(&topologybuffer, binary.LittleEndian, node_ref.EdgeRefBWDCount)
	}
	for i := 0; i < fwd_edgerefcount; i++ {
		edgeref := self.fwd_edge_refs.Get(i)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref.EdgeID)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref._Type)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref.OtherID)
	}
	for i := 0; i < bwd_edgerefcount; i++ {
		edgeref := self.bwd_edge_refs.Get(i)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref.EdgeID)
		binary.Write(&topologybuffer, binary.LittleEndian, edgeref._Type)
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
	node_refs := NewList[NodeRef](int(nodecount))
	fwd_edge_refs := NewList[EdgeRef](int(fwd_edgerefcount))
	bwd_edge_refs := NewList[EdgeRef](int(bwd_edgerefcount))
	for i := 0; i < int(nodecount); i++ {
		var s1 int32
		binary.Read(topologyreader, binary.LittleEndian, &s1)
		var c1 int16
		binary.Read(topologyreader, binary.LittleEndian, &c1)
		var s2 int32
		binary.Read(topologyreader, binary.LittleEndian, &s2)
		var c2 int16
		binary.Read(topologyreader, binary.LittleEndian, &c2)
		node_refs.Add(NodeRef{
			EdgeRefFWDStart: s1,
			EdgeRefFWDCount: c1,
			EdgeRefBWDStart: s2,
			EdgeRefBWDCount: c2,
		})
	}
	for i := 0; i < int(fwd_edgerefcount); i++ {
		var id int32
		binary.Read(topologyreader, binary.LittleEndian, &id)
		var t byte
		binary.Read(topologyreader, binary.LittleEndian, &t)
		var nid int32
		binary.Read(topologyreader, binary.LittleEndian, &nid)
		fwd_edge_refs.Add(EdgeRef{
			EdgeID:  id,
			_Type:   t,
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
		bwd_edge_refs.Add(EdgeRef{
			EdgeID:  id,
			_Type:   t,
			OtherID: nid,
		})
	}

	return &TopologyStore{
		node_refs:     node_refs,
		fwd_edge_refs: fwd_edge_refs,
		bwd_edge_refs: bwd_edge_refs,
	}
}

type TopologyAccessor struct {
	topology      *TopologyStore
	state         int32
	end           int32
	edge_refs     Array[EdgeRef]
	curr_edge_id  int32
	curr_other_id int32
}

func (self *TopologyAccessor) SetBaseNode(node int32, dir Direction) {
	ref := self.topology.node_refs[node]
	if dir == FORWARD {
		start, count := ref.EdgeRefFWDStart, int32(ref.EdgeRefFWDCount)
		self.state = start
		self.end = start + count
		self.edge_refs = Array[EdgeRef](self.topology.fwd_edge_refs)
	} else {
		start, count := ref.EdgeRefBWDStart, int32(ref.EdgeRefBWDCount)
		self.state = start
		self.end = start + count
		self.edge_refs = Array[EdgeRef](self.topology.bwd_edge_refs)
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

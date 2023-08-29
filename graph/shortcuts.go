package graph

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type ShortcutStore struct {
	shortcuts List[Shortcut]
	edge_refs List[Tuple[int32, byte]]
}

func (self *ShortcutStore) GetShortcut(index int32) Shortcut {
	return self.shortcuts[index]
}
func (self *ShortcutStore) SetShortcut(index int32, shortcut Shortcut) {
	self.shortcuts[index] = shortcut
}
func (self *ShortcutStore) AddShortcut(node_a, node_b int32, edges []Tuple[int32, byte]) int32 {
	index := self.shortcuts.Length()
	edge_ref_start := self.edge_refs.Length()
	edge_ref_count := 0
	for _, ref := range edges {
		edge_ref_count += 1
		self.edge_refs.Add(ref)
	}
	self.shortcuts.Add(Shortcut{
		NodeA:         node_a,
		NodeB:         node_b,
		_EdgeRefStart: int32(edge_ref_start),
		_EdgeRefCount: int16(edge_ref_count),
	})
	return int32(index)
}
func (self *ShortcutStore) IsShortcut(index int32) bool {
	if index < int32(len(self.shortcuts)) {
		return true
	} else {
		return false
	}
}
func (self *ShortcutStore) ShortcutCount() int {
	return self.shortcuts.Length()
}
func (self *ShortcutStore) GetEdgesFromShortcut(shc_id int32, reversed bool) List[int32] {
	edges := NewList[int32](2)
	shortcut := self.shortcuts[shc_id]
	for _, ref := range self.edge_refs[shortcut._EdgeRefStart : shortcut._EdgeRefStart+int32(shortcut._EdgeRefCount)] {
		edges.Add(ref.A)
	}
	return edges
}

// reorders node information in edgestore,
// mapping: old id -> new id
func (self *ShortcutStore) _ReorderNodes(mapping Array[int32]) {
	for i := 0; i < self.ShortcutCount(); i++ {
		shc := self.shortcuts[i]
		shc.NodeA = mapping[shc.NodeA]
		shc.NodeB = mapping[shc.NodeB]
		self.shortcuts[i] = shc
	}
}

func _StoreShortcutStore(sh_store *ShortcutStore, sh_weight *DefaultWeighting, filename string) {
	shcbuffer := bytes.Buffer{}
	shortcutcount := sh_store.shortcuts.Length()
	binary.Write(&shcbuffer, binary.LittleEndian, int32(shortcutcount))
	edgerefcount := sh_store.edge_refs.Length()
	binary.Write(&shcbuffer, binary.LittleEndian, int32(edgerefcount))

	for i := 0; i < shortcutcount; i++ {
		shortcut := sh_store.shortcuts.Get(i)
		weight := sh_weight.GetEdgeWeight(int32(i))
		binary.Write(&shcbuffer, binary.LittleEndian, int32(shortcut.NodeA))
		binary.Write(&shcbuffer, binary.LittleEndian, int32(shortcut.NodeB))
		binary.Write(&shcbuffer, binary.LittleEndian, shortcut._EdgeRefStart)
		binary.Write(&shcbuffer, binary.LittleEndian, shortcut._EdgeRefCount)
		binary.Write(&shcbuffer, binary.LittleEndian, uint32(weight))
	}
	for i := 0; i < edgerefcount; i++ {
		edgeref := sh_store.edge_refs[i]
		binary.Write(&shcbuffer, binary.LittleEndian, edgeref.A)
		binary.Write(&shcbuffer, binary.LittleEndian, edgeref.B)
	}

	shcfile, _ := os.Create(filename)
	defer shcfile.Close()
	shcfile.Write(shcbuffer.Bytes())
}

func _LoadShortcutStore(file string) (*ShortcutStore, *DefaultWeighting) {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	shortcutdata, _ := os.ReadFile(file)
	shortcutreader := bytes.NewReader(shortcutdata)
	var shortcutcount int32
	binary.Read(shortcutreader, binary.LittleEndian, &shortcutcount)
	var edgerefcount int32
	binary.Read(shortcutreader, binary.LittleEndian, &edgerefcount)
	shortcuts := NewList[Shortcut](int(shortcutcount))
	edgerefs := NewList[Tuple[int32, byte]](int(edgerefcount))
	shortcut_weights := NewList[int32](int(shortcutcount))
	for i := 0; i < int(shortcutcount); i++ {
		var node_a int32
		binary.Read(shortcutreader, binary.LittleEndian, &node_a)
		var node_b int32
		binary.Read(shortcutreader, binary.LittleEndian, &node_b)
		var s int32
		binary.Read(shortcutreader, binary.LittleEndian, &s)
		var c int16
		binary.Read(shortcutreader, binary.LittleEndian, &c)
		var weight uint32
		binary.Read(shortcutreader, binary.LittleEndian, &weight)
		shortcut := Shortcut{
			NodeA:         node_a,
			NodeB:         node_b,
			_EdgeRefStart: s,
			_EdgeRefCount: c,
		}
		shortcuts.Add(shortcut)
		shortcut_weights.Add(int32(weight))
	}
	for i := 0; i < int(edgerefcount); i++ {
		var e int32
		binary.Read(shortcutreader, binary.LittleEndian, &e)
		var t byte
		binary.Read(shortcutreader, binary.LittleEndian, &t)
		edgerefs.Add(MakeTuple(e, t))
	}

	shortcut_store := &ShortcutStore{
		shortcuts: shortcuts,
		edge_refs: edgerefs,
	}
	weight := &DefaultWeighting{
		edge_weights: shortcut_weights,
	}
	return shortcut_store, weight
}

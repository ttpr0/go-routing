package graph

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type CHShortcutStore struct {
	shortcuts Array[CHShortcut]
}

func (self *CHShortcutStore) GetShortcut(index int32) CHShortcut {
	return self.shortcuts[index]
}
func (self *CHShortcutStore) SetShortcut(index int32, shortcut CHShortcut) {
	self.shortcuts[index] = shortcut
}
func (self *CHShortcutStore) IsShortcut(index int32) bool {
	if index < int32(len(self.shortcuts)) {
		return true
	} else {
		return false
	}
}
func (self *CHShortcutStore) ShortcutCount() int {
	return self.shortcuts.Length()
}
func (self *CHShortcutStore) GetEdgesFromShortcut(shc_id int32, reversed bool) List[int32] {
	edges := NewList[int32](2)
	self._UnpackShortcutRecursive(&edges, shc_id, reversed)
	return edges
}
func (self *CHShortcutStore) _UnpackShortcutRecursive(edges *List[int32], shc_id int32, reversed bool) {
	shortcut := self.GetShortcut(shc_id)
	if reversed {
		e := shortcut._Edges[1]
		if e.B == 2 || e.B == 3 {
			self._UnpackShortcutRecursive(edges, e.A, reversed)
		} else {
			edges.Add(e.A)
		}
		e = shortcut._Edges[0]
		if e.B == 2 || e.B == 3 {
			self._UnpackShortcutRecursive(edges, e.A, reversed)
		} else {
			edges.Add(e.A)
		}
	} else {
		e := shortcut._Edges[0]
		if e.B == 2 || e.B == 3 {
			self._UnpackShortcutRecursive(edges, e.A, reversed)
		} else {
			edges.Add(e.A)
		}
		e = shortcut._Edges[1]
		if e.B == 2 || e.B == 3 {
			self._UnpackShortcutRecursive(edges, e.A, reversed)
		} else {
			edges.Add(e.A)
		}
	}
}

// reorders node information in shortcutstore,
// mapping: old id -> new id
func (self *CHShortcutStore) _ReorderNodes(mapping Array[int32]) {
	for i := 0; i < self.ShortcutCount(); i++ {
		edge := self.shortcuts[i]
		edge.NodeA = mapping[edge.NodeA]
		edge.NodeB = mapping[edge.NodeB]
		self.shortcuts[i] = edge
	}
}

func _StoreCHShortcutStore(sh_store *CHShortcutStore, sh_weight *DefaultWeighting, filename string) {
	shcbuffer := bytes.Buffer{}
	shortcutcount := sh_store.shortcuts.Length()
	binary.Write(&shcbuffer, binary.LittleEndian, int32(shortcutcount))

	for i := 0; i < shortcutcount; i++ {
		shortcut := sh_store.shortcuts.Get(i)
		weight := sh_weight.GetEdgeWeight(int32(i))
		binary.Write(&shcbuffer, binary.LittleEndian, int32(shortcut.NodeA))
		binary.Write(&shcbuffer, binary.LittleEndian, int32(shortcut.NodeB))
		binary.Write(&shcbuffer, binary.LittleEndian, uint32(weight))
		for _, edge := range shortcut._Edges {
			binary.Write(&shcbuffer, binary.LittleEndian, edge.A)
			binary.Write(&shcbuffer, binary.LittleEndian, edge.B == 2 || edge.B == 3)
		}
	}

	shcfile, _ := os.Create(filename)
	defer shcfile.Close()
	shcfile.Write(shcbuffer.Bytes())
}

func _LoadCHShortcutStore(file string) (*CHShortcutStore, *DefaultWeighting) {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	shortcutdata, _ := os.ReadFile(file)
	shortcutreader := bytes.NewReader(shortcutdata)
	var shortcutcount int32
	binary.Read(shortcutreader, binary.LittleEndian, &shortcutcount)
	shortcuts := NewList[CHShortcut](int(shortcutcount))
	shortcut_weights := NewList[int32](int(shortcutcount))
	for i := 0; i < int(shortcutcount); i++ {
		var node_a int32
		binary.Read(shortcutreader, binary.LittleEndian, &node_a)
		var node_b int32
		binary.Read(shortcutreader, binary.LittleEndian, &node_b)
		var weight uint32
		binary.Read(shortcutreader, binary.LittleEndian, &weight)
		shortcut := CHShortcut{
			NodeA:  node_a,
			NodeB:  node_b,
			_Edges: [2]Tuple[int32, byte]{},
		}
		for j := 0; j < 2; j++ {
			var id int32
			binary.Read(shortcutreader, binary.LittleEndian, &id)
			var is bool
			binary.Read(shortcutreader, binary.LittleEndian, &is)
			if is {
				shortcut._Edges[j] = MakeTuple(id, byte(2))
			} else {
				shortcut._Edges[j] = MakeTuple(id, byte(0))
			}
		}
		shortcuts.Add(shortcut)
		shortcut_weights.Add(int32(weight))
	}

	shortcut_store := &CHShortcutStore{
		shortcuts: Array[CHShortcut](shortcuts),
	}
	weight := &DefaultWeighting{
		edge_weights: shortcut_weights,
	}
	return shortcut_store, weight
}

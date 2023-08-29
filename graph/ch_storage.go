package graph

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type CHStore struct {
	shortcuts   Array[CHShortcut]
	node_levels Array[int16]
	sh_weight   Array[int32]
}

func (self *CHStore) GetNodeLevel(node int32) int16 {
	return self.node_levels[node]
}

func (self *CHStore) ShortcutCount() int {
	return len(self.shortcuts)
}

func (self *CHStore) GetShortcut(shc int32) CHShortcut {
	return self.shortcuts[shc]
}

func (self *CHStore) GetEdgesFromShortcut(shc_id int32, reversed bool) List[int32] {
	edges := NewList[int32](2)
	self._UnpackShortcutRecursive(&edges, shc_id, reversed)
	return edges
}
func (self *CHStore) _UnpackShortcutRecursive(edges *List[int32], shc_id int32, reversed bool) {
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

func _StoreCHStorage(store CHStore, file string) {
	_StoreCHShortcuts(store.shortcuts, store.sh_weight, file+"-shortcut")
	_StoreCHLevels(store.node_levels, file+"-level")
}

func _LoadCHStorage(file string, nodecount int) CHStore {
	shortcuts, weights := _LoadCHShortcuts(file + "-shortcut")
	node_levels := _LoadCHLevels(file+"-level", nodecount)

	return CHStore{
		shortcuts:   shortcuts,
		node_levels: node_levels,
		sh_weight:   weights,
	}
}

//*******************************************
// reorder nodes
//*******************************************

// reorders node information in edgestore,
// mapping: old id -> new id
func (self *CHStore) _ReorderNodes(mapping Array[int32]) {
	// shortcuts
	for i := 0; i < self.ShortcutCount(); i++ {
		edge := self.shortcuts[i]
		edge.NodeA = mapping[edge.NodeA]
		edge.NodeB = mapping[edge.NodeB]
		self.shortcuts[i] = edge
	}

	// levels
	new_levels := NewArray[int16](self.node_levels.Length())
	for i, id := range mapping {
		new_levels[id] = self.node_levels[i]
	}
	self.node_levels = new_levels
}

//*******************************************
// load and store components
//*******************************************

func _StoreCHShortcuts(shortcuts Array[CHShortcut], sh_weight Array[int32], filename string) {
	shcbuffer := bytes.Buffer{}
	shortcutcount := shortcuts.Length()
	binary.Write(&shcbuffer, binary.LittleEndian, int32(shortcutcount))

	for i := 0; i < shortcutcount; i++ {
		shortcut := shortcuts.Get(i)
		weight := sh_weight[i]
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

func _LoadCHShortcuts(file string) (Array[CHShortcut], Array[int32]) {
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

	return Array[CHShortcut](shortcuts), Array[int32](shortcut_weights)
}

func _StoreCHLevels(ch_levels Array[int16], filename string) {
	lvlbuffer := bytes.Buffer{}
	nodecount := ch_levels.Length()
	for i := 0; i < nodecount; i++ {
		binary.Write(&lvlbuffer, binary.LittleEndian, ch_levels[i])
	}

	lvlfile, _ := os.Create(filename)
	defer lvlfile.Close()
	lvlfile.Write(lvlbuffer.Bytes())
}

func _LoadCHLevels(file string, nodecount int) Array[int16] {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	leveldata, _ := os.ReadFile(file)
	levelreader := bytes.NewReader(leveldata)
	levels := NewList[int16](int(nodecount))
	for i := 0; i < int(nodecount); i++ {
		var l int16
		binary.Read(levelreader, binary.LittleEndian, &l)
		levels.Add(l)
	}

	return Array[int16](levels)
}

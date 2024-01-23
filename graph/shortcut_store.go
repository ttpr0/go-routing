package graph

import (
	"errors"
	"os"

	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// shortcut storage
//*******************************************

func _NewShortcutStore(cap int, is_static bool) _ShortcutStore {
	if is_static {
		return _ShortcutStore{
			shortcuts:        NewList[Shortcut](cap),
			_has_static_size: true,
			_is_node_based:   true,
			edges_stat:       Some(NewList[[2]Tuple[int32, byte]](cap)),
		}
	} else {
		return _ShortcutStore{
			shortcuts:        NewList[Shortcut](cap),
			_has_static_size: false,
			_is_node_based:   true,
			edgerefs_dyn:     Some(NewList[Tuple[int32, int16]](cap)),
			edges_dyn:        Some(NewList[int32](cap)),
		}
	}
}

type _ShortcutStore struct {
	// list of shortcuts
	shortcuts List[Shortcut]

	_has_static_size bool // true if every edges has two child edges
	_is_node_based   bool // true if shortcuts are between nodes

	// Dynamic storage of underlying edges.
	//
	// Every shortcut can consist of multiple edges.
	edgerefs_dyn Optional[List[Tuple[int32, int16]]]
	edges_dyn    Optional[List[int32]]

	// Static storage of underlying edges.
	//
	// Shortcuts consist of a tree of edges and shortcuts.
	// Every Shortcut has two childs that can be either a shortcut (2) or an edge (0).
	edges_stat Optional[List[[2]Tuple[int32, byte]]]
}

// Number of shortcuts currently stored.
func (self *_ShortcutStore) ShortcutCount() int {
	return self.shortcuts.Length()
}

// Returns shortcut with id.
func (self *_ShortcutStore) GetShortcut(shc_id int32) Shortcut {
	return self.shortcuts[shc_id]
}

// Adds a new shortcut with static edges-count.
func (self *_ShortcutStore) AddCHShortcut(shc Shortcut, edges [2]Tuple[int32, byte]) (int32, error) {
	if !self._has_static_size || !self.edges_stat.HasValue() {
		return 0, errors.New("ShortcutStore has dynamic edgerefs")
	}
	index := self.shortcuts.Length()
	self.shortcuts.Add(shc)
	self.edges_stat.Value.Add(edges)
	return int32(index), nil
}

// Adds a new shortcut with dynamic edges-count.
func (self *_ShortcutStore) AddShortcut(shc Shortcut, edges []int32) (int32, error) {
	if self._has_static_size || !self.edgerefs_dyn.HasValue() {
		return 0, errors.New("ShortcutStore has static edgerefs")
	}
	index := self.shortcuts.Length()
	edge_ref_start := self.edgerefs_dyn.Value.Length()
	edge_ref_count := 0
	for _, ref := range edges {
		edge_ref_count += 1
		self.edges_dyn.Value.Add(ref)
	}
	self.shortcuts.Add(shc)
	self.edgerefs_dyn.Value.Add(MakeTuple(int32(edge_ref_start), int16(edge_ref_count)))
	return int32(index), nil
}

// Retrives underlying edges from shortcut.
func (self *_ShortcutStore) GetEdgesFromShortcut(shc_id int32, reversed bool, handler func(int32)) {
	if !self._has_static_size {
		edgeref := self.edgerefs_dyn.Value[shc_id]
		edges := self.edges_dyn.Value[edgeref.A : edgeref.A+int32(edgeref.B)]
		if reversed {
			for i := len(edges) - 1; i >= 0; i-- {
				handler(edges[i])
			}
		} else {
			for i := 0; i < len(edges); i++ {
				handler(edges[i])
			}
		}
	} else {
		self._UnpackCHShortcutRecursive(shc_id, reversed, handler)
	}
}
func (self *_ShortcutStore) _UnpackCHShortcutRecursive(shc_id int32, reversed bool, handler func(int32)) {
	edges := self.edges_stat.Value[shc_id]
	if reversed {
		e := edges[1]
		if e.B == 2 {
			self._UnpackCHShortcutRecursive(e.A, reversed, handler)
		} else {
			handler(e.A)
		}
		e = edges[0]
		if e.B == 2 {
			self._UnpackCHShortcutRecursive(e.A, reversed, handler)
		} else {
			handler(e.A)
		}
	} else {
		e := edges[0]
		if e.B == 2 {
			self._UnpackCHShortcutRecursive(e.A, reversed, handler)
		} else {
			handler(e.A)
		}
		e = edges[1]
		if e.B == 2 {
			self._UnpackCHShortcutRecursive(e.A, reversed, handler)
		} else {
			handler(e.A)
		}
	}
}

//*******************************************
// reorder shortcut storage nodes
//*******************************************

// reorders node information,
// mapping: old id -> new id
func (self *_ShortcutStore) _ReorderNodes(mapping Array[int32]) {
	if self._is_node_based {
		// shortcuts
		for i := 0; i < self.ShortcutCount(); i++ {
			shc := self.shortcuts[i]
			shc.From = mapping[shc.From]
			shc.To = mapping[shc.To]
			self.shortcuts[i] = shc
		}
	}
}

// reorders edge information,
// mapping: old id -> new id
func (self *_ShortcutStore) _ReorderEdges(mapping Array[int32]) {
	if !self._is_node_based {
		// shortcuts
		for i := 0; i < self.ShortcutCount(); i++ {
			shc := self.shortcuts[i]
			shc.From = mapping[shc.From]
			shc.To = mapping[shc.To]
			self.shortcuts[i] = shc
		}
	}
	if self._has_static_size {
		for i := 0; i < self.edges_stat.Value.Length(); i++ {
			edge := self.edges_stat.Value[i]
			if edge[0].B == 0 {
				edge[0].A = mapping[edge[0].A]
			}
			if edge[1].B == 0 {
				edge[1].A = mapping[edge[1].A]
			}
			self.edges_stat.Value[i] = edge
		}
	} else {
		for i := 0; i < self.edges_dyn.Value.Length(); i++ {
			edge := self.edges_dyn.Value[i]
			edge = mapping[edge]
			self.edges_dyn.Value[i] = edge
		}
	}
}

//*******************************************
// shortcut storage IO
//*******************************************

func _StoreShortcuts(store _ShortcutStore, filename string) {
	writer := NewBufferWriter()

	// write header
	shortcutcount := store.shortcuts.Length()
	Write[int32](writer, int32(shortcutcount))
	Write[bool](writer, store._has_static_size)
	if store._has_static_size {
		Write[int32](writer, 0)
	} else {
		edgescount := store.edges_dyn.Value.Length()
		Write[int32](writer, int32(edgescount))
	}

	// write shortcuts
	for i := 0; i < shortcutcount; i++ {
		shc := store.shortcuts[i]
		Write[int32](writer, shc.From)
		Write[int32](writer, shc.To)
		Write[int32](writer, shc.Weight)
		Write[[4]byte](writer, shc._payload)
	}

	// write shortcut edges
	if store._has_static_size {
		for i := 0; i < shortcutcount; i++ {
			edges := store.edges_stat.Value[i]
			for _, edge := range edges {
				Write[int32](writer, edge.A)
				Write[bool](writer, edge.B == 2)
			}
		}
	} else {
		for i := 0; i < shortcutcount; i++ {
			edgeref := store.edgerefs_dyn.Value[i]
			Write[int32](writer, edgeref.A)
			Write[int16](writer, edgeref.B)
		}
		for i := 0; i < store.edges_dyn.Value.Length(); i++ {
			edge := store.edges_dyn.Value[i]
			Write[int32](writer, edge)
		}
	}

	shcfile, _ := os.Create(filename)
	defer shcfile.Close()
	shcfile.Write(writer.Bytes())
}

func _LoadShortcuts(file string) _ShortcutStore {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	shortcutdata, _ := os.ReadFile(file)
	reader := NewBufferReader(shortcutdata)

	shortcutcount := Read[int32](reader)
	_has_static_size := Read[bool](reader)
	edges_count := Read[int32](reader)

	shortcuts := NewList[Shortcut](int(shortcutcount))
	for i := 0; i < int(shortcutcount); i++ {
		from := Read[int32](reader)
		to := Read[int32](reader)
		weight := Read[int32](reader)
		payload := Read[[4]byte](reader)
		shortcut := Shortcut{
			From:     from,
			To:       to,
			Weight:   weight,
			_payload: payload,
		}
		shortcuts.Add(shortcut)
	}
	if _has_static_size {
		edges_stat := NewList[[2]Tuple[int32, byte]](int(shortcutcount))
		for i := 0; i < int(shortcutcount); i++ {
			edges := [2]Tuple[int32, byte]{}
			for j := 0; j < 2; j++ {
				e_id := Read[int32](reader)
				is_s := Read[bool](reader)
				e_ty := byte(0)
				if is_s {
					e_ty = 2
				}
				edges[j] = MakeTuple(e_id, e_ty)
			}
			edges_stat.Add(edges)
		}
		return _ShortcutStore{
			shortcuts:        shortcuts,
			_has_static_size: true,

			edges_stat: Some(edges_stat),
		}
	} else {
		edgerefs_dyn := NewList[Tuple[int32, int16]](int(shortcutcount))
		for i := 0; i < int(shortcutcount); i++ {
			e_s := Read[int32](reader)
			e_c := Read[int16](reader)
			edgerefs_dyn.Add(MakeTuple(e_s, e_c))
		}
		edges_dyn := NewList[int32](int(edges_count))
		for i := 0; i < int(edges_count); i++ {
			e_id := Read[int32](reader)
			edges_dyn.Add(e_id)
		}

		return _ShortcutStore{
			shortcuts:        shortcuts,
			_has_static_size: false,

			edgerefs_dyn: Some(edgerefs_dyn),
			edges_dyn:    Some(edges_dyn),
		}
	}
}

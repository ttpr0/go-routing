package graph

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type TiledStore struct {
	node_tiles   Array[int16]
	shortcuts    List[Shortcut]
	edge_refs    List[Tuple[int32, byte]]
	skip_weights List[int32]
	edge_types   Array[byte]
}

func (self *TiledStore) GetNodeTile(node int32) int16 {
	return self.node_tiles[node]
}
func (self *TiledStore) SetNodeTile(node int32, tile int16) {
	self.node_tiles[node] = tile
}
func (self *TiledStore) GetEdgeType(edge int32) byte {
	return self.edge_types[edge]
}
func (self *TiledStore) SetEdgeType(edge int32, typ byte) {
	self.edge_types[edge] = typ
}
func (self *TiledStore) TileCount() int16 {
	max := int16(0)
	for i := 0; i < len(self.node_tiles); i++ {
		tile := self.node_tiles[i]
		if tile > max {
			max = tile
		}
	}
	return max - 1
}
func (self *TiledStore) ShortcutCount() int {
	return self.shortcuts.Length()
}
func (self *TiledStore) GetShortcut(shc int32) Shortcut {
	return self.shortcuts[shc]
}
func (self *TiledStore) AddShortcut(node_a, node_b int32, edges []Tuple[int32, byte], weight int32) int32 {
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
	self.skip_weights.Add(weight)
	return int32(index)
}
func (self *TiledStore) GetEdgesFromShortcut(shc_id int32, reversed bool) List[int32] {
	edges := NewList[int32](2)
	shortcut := self.shortcuts[shc_id]
	for _, ref := range self.edge_refs[shortcut._EdgeRefStart : shortcut._EdgeRefStart+int32(shortcut._EdgeRefCount)] {
		edges.Add(ref.A)
	}
	return edges
}

func _StoreTiledStorage(store TiledStore, file string) {
	_StoreTiledNodeTiles(store.node_tiles, file+"-tiles")
	_StoreTiledEdgeTypes(store.edge_types, file+"-tiles_types")
	_StoreTiledShortcuts(store.shortcuts, store.edge_refs, store.skip_weights, file+"-skip_shortcuts")
}

func _LoadTiledStorage(file string, nodecount, edgecount int) TiledStore {
	node_tiles := _LoadTiledNodeTiles(file+"-tiles", nodecount)
	edge_types := _LoadTiledEdgeTypes(file+"-tiles_types", edgecount)
	shortcuts, edge_refs, sh_weights := _LoadTiledShortcuts(file + "-skip_shortcuts")

	return TiledStore{
		node_tiles:   node_tiles,
		edge_types:   edge_types,
		shortcuts:    shortcuts,
		edge_refs:    edge_refs,
		skip_weights: sh_weights,
	}
}

//*******************************************
// reorder nodes
//*******************************************

// reorders node information in edgestore,
// mapping: old id -> new id
func (self *TiledStore) _ReorderNodes(mapping Array[int32]) {
	// node tiles
	new_tiles := NewArray[int16](self.node_tiles.Length())
	for i, id := range mapping {
		new_tiles[id] = self.node_tiles[i]
	}
	self.node_tiles = new_tiles

	// shortcuts
	for i := 0; i < self.ShortcutCount(); i++ {
		shc := self.shortcuts[i]
		shc.NodeA = mapping[shc.NodeA]
		shc.NodeB = mapping[shc.NodeB]
		self.shortcuts[i] = shc
	}
}

//*******************************************
// load and store components
//*******************************************

func _StoreTiledNodeTiles(node_tiles Array[int16], filename string) {
	tilesbuffer := bytes.Buffer{}

	for i := 0; i < node_tiles.Length(); i++ {
		binary.Write(&tilesbuffer, binary.LittleEndian, node_tiles[i])
	}

	tilesfile, _ := os.Create(filename)
	defer tilesfile.Close()
	tilesfile.Write(tilesbuffer.Bytes())
}

func _LoadTiledNodeTiles(file string, nodecount int) Array[int16] {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	tiledata, _ := os.ReadFile(file)
	tilereader := bytes.NewReader(tiledata)
	node_tiles := NewList[int16](int(nodecount))
	for i := 0; i < int(nodecount); i++ {
		var t int16
		binary.Read(tilereader, binary.LittleEndian, &t)
		node_tiles.Add(t)
	}

	return Array[int16](node_tiles)
}

func _StoreTiledEdgeTypes(edge_types Array[byte], filename string) {
	typesbuffer := bytes.Buffer{}

	for i := 0; i < edge_types.Length(); i++ {
		binary.Write(&typesbuffer, binary.LittleEndian, edge_types[i])
	}

	typesfile, _ := os.Create(filename)
	defer typesfile.Close()
	typesfile.Write(typesbuffer.Bytes())
}

func _LoadTiledEdgeTypes(file string, edgecount int) Array[byte] {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	tiledata, _ := os.ReadFile(file)
	tilereader := bytes.NewReader(tiledata)
	edge_types := NewArray[byte](edgecount)
	for i := 0; i < edgecount; i++ {
		var t byte
		binary.Read(tilereader, binary.LittleEndian, &t)
		edge_types[i] = t
	}

	return edge_types
}

func _StoreTiledShortcuts(shortcuts List[Shortcut], edge_refs List[Tuple[int32, byte]], sh_weight List[int32], filename string) {
	shcbuffer := bytes.Buffer{}
	shortcutcount := shortcuts.Length()
	binary.Write(&shcbuffer, binary.LittleEndian, int32(shortcutcount))
	edgerefcount := edge_refs.Length()
	binary.Write(&shcbuffer, binary.LittleEndian, int32(edgerefcount))

	for i := 0; i < shortcutcount; i++ {
		shortcut := shortcuts.Get(i)
		weight := sh_weight[i]
		binary.Write(&shcbuffer, binary.LittleEndian, int32(shortcut.NodeA))
		binary.Write(&shcbuffer, binary.LittleEndian, int32(shortcut.NodeB))
		binary.Write(&shcbuffer, binary.LittleEndian, shortcut._EdgeRefStart)
		binary.Write(&shcbuffer, binary.LittleEndian, shortcut._EdgeRefCount)
		binary.Write(&shcbuffer, binary.LittleEndian, uint32(weight))
	}
	for i := 0; i < edgerefcount; i++ {
		edgeref := edge_refs[i]
		binary.Write(&shcbuffer, binary.LittleEndian, edgeref.A)
		binary.Write(&shcbuffer, binary.LittleEndian, edgeref.B)
	}

	shcfile, _ := os.Create(filename)
	defer shcfile.Close()
	shcfile.Write(shcbuffer.Bytes())
}

func _LoadTiledShortcuts(file string) (List[Shortcut], List[Tuple[int32, byte]], List[int32]) {
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

	return shortcuts, edgerefs, shortcut_weights
}

package graph

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

type NodeTileStore struct {
	node_tiles Array[int16]
}

func (self *NodeTileStore) GetNodeTile(index int32) int16 {
	return self.node_tiles[index]
}
func (self *NodeTileStore) SetNodeTile(index int32, tile int16) {
	self.node_tiles[index] = tile
}
func (self *NodeTileStore) GetTiles() Array[int16] {
	return self.node_tiles
}

// reorders nodes in tilestore,
// node mapping: old id -> new id
func (self *NodeTileStore) _ReorderNodes(mapping Array[int32]) {
	new_tiles := NewArray[int16](self.node_tiles.Length())
	for i, id := range mapping {
		new_tiles[id] = self.node_tiles[i]
	}
	self.node_tiles = new_tiles
}

func _StoreNodeTileStore(tile_store *NodeTileStore, filename string) {
	tilesbuffer := bytes.Buffer{}

	for i := 0; i < tile_store.node_tiles.Length(); i++ {
		binary.Write(&tilesbuffer, binary.LittleEndian, tile_store.node_tiles[i])
	}

	tilesfile, _ := os.Create(filename)
	defer tilesfile.Close()
	tilesfile.Write(tilesbuffer.Bytes())
}

func _LoadNodeTileStore(file string, nodecount int) *NodeTileStore {
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

	return &NodeTileStore{
		node_tiles: Array[int16](node_tiles),
	}
}

package comps

import (
	"os"

	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// partition interface
//*******************************************

type IPartition interface {
	GetNodeTile(node int32) int16
}

//*******************************************
// partition-data
//*******************************************

func NewPartition(tiles Array[int16]) *Partition {
	return &Partition{
		node_tiles: tiles,
	}
}

type Partition struct {
	node_tiles Array[int16]
}

func (self *Partition) GetNodeTile(node int32) int16 {
	return self.node_tiles[node]
}
func (self *Partition) SetNodeTile(node int32, tile int16) {
	self.node_tiles[node] = tile
}
func (self *Partition) TileCount() int16 {
	max := int16(0)
	for i := 0; i < len(self.node_tiles); i++ {
		tile := self.node_tiles[i]
		if tile > max {
			max = tile
		}
	}
	return max + 1
}
func (self *Partition) GetTiles() List[int16] {
	tile_dict := NewDict[int16, bool](100)
	for i := 0; i < self.node_tiles.Length(); i++ {
		tile_id := self.node_tiles[i]
		if tile_dict.ContainsKey(tile_id) {
			continue
		}
		tile_dict[tile_id] = true
	}
	tile_list := NewList[int16](len(tile_dict))
	for tile, _ := range tile_dict {
		tile_list.Add(tile)
	}
	return tile_list
}

func (self *Partition) _ReorderNodes(mapping Array[int32]) *Partition {
	new_tiles := Reorder[int16](self.node_tiles, mapping)

	return &Partition{
		node_tiles: new_tiles,
	}
}
func (self *Partition) _New() *Partition {
	return &Partition{}
}
func (self *Partition) _Load(path string) {
	node_tiles := ReadArrayFromFile[int16](path + "-tiles")

	*self = Partition{
		node_tiles: node_tiles,
	}
}
func (self *Partition) _Store(path string) {
	WriteArrayToFile[int16](self.node_tiles, path+"-tiles")
}
func (self *Partition) _Remove(path string) {
	os.Remove(path + "-tiles")
}

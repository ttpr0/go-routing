package graph

import (
	"os"

	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// overlay-data
//*******************************************

type Overlay struct {
	skip_shortcuts _ShortcutStore
	skip_topology  _AdjacencyArray
	edge_types     Array[byte]
}

func (self *Overlay) _ReorderNodes(mapping Array[int32]) {
	self.skip_shortcuts._ReorderNodes(mapping)
	self.skip_topology._ReorderNodes(mapping)
}
func (self *Overlay) _New() *Overlay {
	return &Overlay{}
}
func (self *Overlay) _Load(path string) {
	skip_shortcuts := _LoadShortcuts(path + "skip_shortcuts")
	skip_topology := _LoadAdjacency(path+"-skip_topology", true)
	edge_types := ReadArrayFromFile[byte](path + "-tiles_types")

	*self = Overlay{
		skip_shortcuts: skip_shortcuts,
		skip_topology:  *skip_topology,
		edge_types:     edge_types,
	}
}
func (self *Overlay) _Store(path string) {
	_StoreShortcuts(self.skip_shortcuts, path+"-skip_shortcuts")
	_StoreAdjacency(&self.skip_topology, true, path+"-skip_topology")
	WriteArrayToFile[byte](self.edge_types, path+"-tiles_types")
}
func (self *Overlay) _Remove(path string) {
	os.Remove(path + "-skip_shortcuts")
	os.Remove(path + "-skip_topology")
	os.Remove(path + "-tiles_types")
}

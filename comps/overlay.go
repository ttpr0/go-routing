package comps

import (
	"os"

	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// overlay-data
//*******************************************

func NewOverlay(shortcuts structs.ShortcutStore, topology structs.AdjacencyArray, edge_types Array[byte]) *Overlay {
	return &Overlay{
		skip_shortcuts: shortcuts,
		skip_topology:  topology,
		edge_types:     edge_types,
	}
}

type Overlay struct {
	skip_shortcuts structs.ShortcutStore
	skip_topology  structs.AdjacencyArray
	edge_types     Array[byte]
}

func (self *Overlay) ShortcutCount() int {
	return self.skip_shortcuts.ShortcutCount()
}
func (self *Overlay) GetShortcut(shc_id int32) structs.Shortcut {
	return self.skip_shortcuts.GetShortcut(shc_id)
}
func (self *Overlay) GetEdgesFromShortcut(edge int32, reverse bool, callback func(int32)) {
	self.skip_shortcuts.GetEdgesFromShortcut(edge, reverse, callback)
}
func (self *Overlay) GetAccessor() structs.IAdjAccessor {
	acc := self.skip_topology.GetAccessor()
	return &acc
}
func (self *Overlay) GetEdgeType(edge int32) byte {
	return self.edge_types[edge]
}

func (self *Overlay) _ReorderNodes(mapping Array[int32]) *Overlay {
	new_shortcuts := self.skip_shortcuts.ReorderNodes(mapping)
	new_topology := self.skip_topology.ReorderNodes(mapping)

	return &Overlay{
		skip_shortcuts: new_shortcuts,
		skip_topology:  new_topology,
		edge_types:     self.edge_types.Copy(),
	}
}
func (self *Overlay) _New() *Overlay {
	return &Overlay{}
}
func (self *Overlay) _Load(path string) {
	skip_shortcuts := structs.LoadShortcuts(path + "-skip_shortcuts")
	skip_topology := structs.LoadAdjacency(path+"-skip_topology", true)
	edge_types := ReadArrayFromFile[byte](path + "-tiles_types")

	*self = Overlay{
		skip_shortcuts: skip_shortcuts,
		skip_topology:  *skip_topology,
		edge_types:     edge_types,
	}
}
func (self *Overlay) _Store(path string) {
	structs.StoreShortcuts(self.skip_shortcuts, path+"-skip_shortcuts")
	structs.StoreAdjacency(&self.skip_topology, true, path+"-skip_topology")
	WriteArrayToFile[byte](self.edge_types, path+"-tiles_types")
}
func (self *Overlay) _Remove(path string) {
	os.Remove(path + "-skip_shortcuts")
	os.Remove(path + "-skip_topology")
	os.Remove(path + "-tiles_types")
}

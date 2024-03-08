package comps

import (
	"os"

	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// ch-data interface
//*******************************************

type ICHData interface {
	GetNodeLevel(node int32) int16
}

//*******************************************
// ch-data
//*******************************************

func NewCH(shortcuts structs.ShortcutStore, topology structs.AdjacencyArray, node_levels Array[int16]) *CH {
	return &CH{
		shortcuts:   shortcuts,
		topology:    topology,
		node_levels: node_levels,
	}
}

type CH struct {
	shortcuts   structs.ShortcutStore
	topology    structs.AdjacencyArray
	node_levels Array[int16]
}

func (self *CH) GetNodeLevel(node int32) int16 {
	return self.node_levels[node]
}
func (self *CH) ShortcutCount() int {
	return self.shortcuts.ShortcutCount()
}
func (self *CH) GetShortcut(shc_id int32) structs.Shortcut {
	return self.shortcuts.GetShortcut(shc_id)
}
func (self *CH) GetEdgesFromShortcut(edge int32, reverse bool, callback func(int32)) {
	self.shortcuts.GetEdgesFromShortcut(edge, reverse, callback)
}
func (self *CH) GetShortcutAccessor() structs.IAdjAccessor {
	acc := self.topology.GetAccessor()
	return &acc
}

func (self *CH) _ReorderNodes(mapping Array[int32]) {
	self.shortcuts.ReorderNodes(mapping)
	self.topology.ReorderNodes(mapping)
	self.node_levels = Reorder[int16](self.node_levels, mapping)
}
func (self *CH) _New() *CH {
	return &CH{}
}
func (self *CH) _Load(path string) {
	ch_topology := structs.LoadAdjacency(path+"-ch_graph", false)
	ch_shortcuts := structs.LoadShortcuts(path + "-shortcut")
	node_levels := ReadArrayFromFile[int16](path + "-level")

	*self = CH{
		shortcuts:   ch_shortcuts,
		topology:    *ch_topology,
		node_levels: node_levels,
	}
}
func (self *CH) _Store(path string) {
	structs.StoreShortcuts(self.shortcuts, path+"-shortcut")
	structs.StoreAdjacency(&self.topology, false, path+"-ch_graph")
	WriteArrayToFile[int16](self.node_levels, path+"-level")
}
func (self *CH) _Remove(path string) {
	os.Remove(path + "-shortcut")
	os.Remove(path + "-ch_graph")
	os.Remove(path + "-level")
}

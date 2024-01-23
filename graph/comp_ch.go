package graph

import (
	"os"

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

type CH struct {
	shortcuts   _ShortcutStore
	topology    _AdjacencyArray
	node_levels Array[int16]
}

func (self *CH) _ReorderNodes(mapping Array[int32]) {
	self.shortcuts._ReorderNodes(mapping)
	self.topology._ReorderNodes(mapping)
	self.node_levels = Reorder[int16](self.node_levels, mapping)
}
func (self *CH) _New() *CH {
	return &CH{}
}
func (self *CH) _Load(path string) {
	ch_topology := _LoadAdjacency(path+"-ch_graph", false)
	ch_shortcuts := _LoadShortcuts(path + "-shortcut")
	node_levels := ReadArrayFromFile[int16](path + "-level")

	*self = CH{
		shortcuts:   ch_shortcuts,
		topology:    *ch_topology,
		node_levels: node_levels,
	}
}
func (self *CH) _Store(path string) {
	_StoreShortcuts(self.shortcuts, path+"-shortcut")
	_StoreAdjacency(&self.topology, false, path+"-ch_graph")
	WriteArrayToFile[int16](self.node_levels, path+"-level")
}
func (self *CH) _Remove(path string) {
	os.Remove(path + "-shortcut")
	os.Remove(path + "-ch_graph")
	os.Remove(path + "-level")
}

package comps

import (
	"os"

	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

//*******************************************
// overlay-data
//*******************************************

func NewTransit(id_mapping structs.IDMapping, stops Array[structs.Node], connections Array[structs.Connection], shortcuts structs.ShortcutStore) *Transit {
	topology := structs.NewAdjacencyList(stops.Length())
	for i := 0; i < connections.Length(); i++ {
		conn := connections[i]
		topology.AddEdgeEntries(conn.StopA, conn.StopB, int32(i), 30)
	}
	for i := 0; i < shortcuts.ShortcutCount(); i++ {
		shc := shortcuts.GetShortcut(int32(i))
		topology.AddEdgeEntries(shc.From, shc.To, int32(i), 100)
	}
	return &Transit{
		id_mapping:  id_mapping,
		stops:       stops,
		connections: connections,
		shortcuts:   shortcuts,
		topology:    *structs.AdjacencyListToArray(&topology),
	}
}

type Transit struct {
	id_mapping  structs.IDMapping
	stops       Array[structs.Node]
	connections Array[structs.Connection]
	shortcuts   structs.ShortcutStore
	topology    structs.AdjacencyArray
}

func (self *Transit) MapNodeToStop(node int32) int32 {
	return self.id_mapping.GetTarget(node)
}
func (self *Transit) MapStopToNode(stop int32) int32 {
	return self.id_mapping.GetSource(stop)
}
func (self *Transit) StopCount() int {
	return self.stops.Length()
}
func (self *Transit) GetStop(stop int32) structs.Node {
	return self.stops[stop]
}
func (self *Transit) ConnectionCount() int {
	return self.connections.Length()
}
func (self *Transit) GetConnection(connection int32) structs.Connection {
	return self.connections[connection]
}
func (self *Transit) ShortcutCount() int {
	return self.shortcuts.ShortcutCount()
}
func (self *Transit) GetShortcut(shortcut int32) structs.Shortcut {
	return self.shortcuts.GetShortcut(shortcut)
}
func (self *Transit) GetEdgesFromShortcut(shortcut int32, reversed bool, callback func(int32)) {
	self.shortcuts.GetEdgesFromShortcut(shortcut, reversed, callback)
}
func (self *Transit) GetAccessor() structs.IAdjAccessor {
	acc := self.topology.GetAccessor()
	return &acc
}

func (self *Transit) _New() *Transit {
	return &Transit{}
}
func (self *Transit) _Load(path string) {
	id_mapping := structs.LoadIDMapping(path + "-id_mapping")
	stops := ReadArrayFromFile[structs.Node](path + "-stops")
	connections := ReadArrayFromFile[structs.Connection](path + "-connections")
	shortcuts := structs.LoadShortcuts(path + "-shortcut")
	topology := structs.LoadAdjacency(path+"-transit_graph", true)

	*self = Transit{
		id_mapping:  id_mapping,
		stops:       stops,
		connections: connections,
		shortcuts:   shortcuts,
		topology:    *topology,
	}
}
func (self *Transit) _Store(path string) {
	structs.StoreIDMapping(self.id_mapping, path+"-id_mapping")
	structs.StoreShortcuts(self.shortcuts, path+"-shortcut")
	structs.StoreAdjacency(&self.topology, false, path+"-transit_graph")
	WriteArrayToFile(self.stops, path+"-stops")
	WriteArrayToFile(self.connections, path+"-connections")
}
func (self *Transit) _Remove(path string) {
	os.Remove(path + "-id_mapping")
	os.Remove(path + "-shortcut")
	os.Remove(path + "-transit_graph")
	os.Remove(path + "-stops")
	os.Remove(path + "-connections")
}

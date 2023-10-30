package graph

import (
	"os"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// ch-data
//*******************************************

type _CHData struct {
	// ID-mapping
	id_mapping _IDMapping

	// Metadata
	_base_weighting   string
	_build_with_tiles bool // true if contraction used tiles
	_contains_dummies bool // true if down_edges contain dummies

	// Additional Storage
	shortcuts   ShortcutStore
	topology    AdjacencyArray
	node_levels Array[int16]

	// contraction order build with tiles
	node_tiles Optional[Array[int16]]

	// index for PHAST
	fwd_down_edges Optional[Array[Shortcut]]
	bwd_down_edges Optional[Array[Shortcut]]
}

func (self *_CHData) Type() SpeedUpType {
	return CH
}

func (self *_CHData) _ReorderNodes(mapping Array[int32]) {
	self.id_mapping.ReorderTargets(mapping)
	self.shortcuts._ReorderNodes(mapping)
	self.topology._ReorderNodes(mapping)
	self.node_levels = Reorder[int16](self.node_levels, mapping)

	if self._build_with_tiles {
		self.node_tiles.Value = Reorder[int16](self.node_tiles.Value, mapping)
	}

	if self.fwd_down_edges.HasValue() || self.bwd_down_edges.HasValue() {
		panic("not implemented")
	}
}

//*******************************************
// ch-data handler
//*******************************************

type _CHDataHandler struct{}

func (self _CHDataHandler) Load(dir string, name string, nodecount int) ISpeedUpData {
	id_mapping := _LoadIDMapping(dir + name + "-mapping")
	ch_topology := _LoadAdjacency(dir+name+"-ch_graph", false, nodecount)
	ch_shortcuts := _LoadShortcuts(dir + name + "-shortcut")
	node_levels := ReadArrayFromFile[int16](dir + name + "-level")
	meta := ReadJSONFromFile[_CHDataMeta](dir + name + "-meta")

	return &_CHData{
		id_mapping: id_mapping,

		_base_weighting:   meta.BaseWeighting,
		_build_with_tiles: meta.BuildWithTiles,
		_contains_dummies: meta.ContainsDummies,

		shortcuts:   ch_shortcuts,
		topology:    *ch_topology,
		node_levels: node_levels,
	}
}
func (self _CHDataHandler) Store(dir string, name string, data ISpeedUpData) {
	ch_data := data.(*_CHData)

	_StoreIDMapping(ch_data.id_mapping, dir+name+"-mapping")
	_StoreShortcuts(ch_data.shortcuts, dir+name+"-shortcut")
	_StoreAdjacency(&ch_data.topology, false, dir+name+"-ch_graph")
	WriteArrayToFile[int16](ch_data.node_levels, dir+name+"-level")
	if ch_data._build_with_tiles {
		WriteArrayToFile[int16](ch_data.node_tiles.Value, dir+name+"-ch_tiles")
	}
	meta := _CHDataMeta{
		BaseWeighting:   ch_data._base_weighting,
		BuildWithTiles:  ch_data._build_with_tiles,
		ContainsDummies: ch_data._contains_dummies,
	}
	WriteJSONToFile(meta, dir+name+"-meta")
}
func (self _CHDataHandler) Remove(dir string, name string) {
	os.Remove(dir + name + "-mapping")
	os.Remove(dir + name + "-shortcut")
	os.Remove(dir + name + "-ch_graph")
	os.Remove(dir + name + "-level")
	os.Remove(dir + name + "-ch_tiles")
	os.Remove(dir + name + "-meta")
}
func (self _CHDataHandler) _ReorderNodes(dir string, name string, mapping Array[int32]) {
	id_mapping := _LoadIDMapping(dir + name + "-mapping")
	id_mapping.ReorderSources(mapping)
	_StoreIDMapping(id_mapping, dir+name+"-mapping")
}
func (self _CHDataHandler) _ReorderNodesInplace(data ISpeedUpData, mapping Array[int32]) {
	ch_data := data.(*_CHData)
	ch_data.id_mapping.ReorderSources(mapping)
}

type _CHDataMeta struct {
	BaseWeighting   string `json:base_weighting`
	BuildWithTiles  bool   `json:built_with_tiles`
	ContainsDummies bool   `json:contains_dummies`
}

package graph

import (
	"errors"
	"os"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// ch-data
//*******************************************

type _TiledData struct {
	// ID-mapping
	id_mapping _IDMapping

	// Metadata
	_base_weighting string

	// Tiles Storage
	skip_shortcuts ShortcutStore
	skip_topology  AdjacencyArray
	node_tiles     Array[int16]
	edge_types     Array[byte]
	cell_index     Optional[_CellIndex] // Storage for indexing sp within cells
}

func (self *_TiledData) Type() SpeedUpType {
	return TILED
}

func (self *_TiledData) _ReorderNodes(mapping Array[int32]) {
	self.id_mapping.ReorderTargets(mapping)
	self.skip_shortcuts._ReorderNodes(mapping)
	self.skip_topology._ReorderNodes(mapping)
	Reorder[int16](self.node_tiles, mapping)
	if self.cell_index.HasValue() {
		self.cell_index.Value._ReorderNodes(mapping)
	}
}

//*******************************************
// ch-data handler
//*******************************************

type _TiledDataHandler struct{}

func (self _TiledDataHandler) Load(dir string, name string, nodecount int) ISpeedUpData {
	id_mapping := _LoadIDMapping(dir + name + "-mapping")
	skip_shortcuts := _LoadShortcuts(dir + name + "skip_shortcuts")
	skip_topology := _LoadAdjacency(dir+name+"-skip_topology", true, nodecount)
	node_tiles := ReadArrayFromFile[int16](dir + name + "-tiles")
	edge_types := ReadArrayFromFile[byte](dir + name + "-tiles_types")

	var cell_index Optional[_CellIndex]
	_, err := os.Stat(dir + name + "tileranges")
	if errors.Is(err, os.ErrNotExist) {
		cell_index = None[_CellIndex]()
	} else {
		cell_index = Some(_LoadCellIndex(dir + name + "tileranges"))
	}

	meta := ReadJSONFromFile[_TiledDataMeta](dir + name + "-meta")

	return &_TiledData{
		id_mapping: id_mapping,

		_base_weighting: meta.BaseWeighting,

		skip_shortcuts: skip_shortcuts,
		skip_topology:  *skip_topology,
		node_tiles:     node_tiles,
		edge_types:     edge_types,
		cell_index:     cell_index,
	}
}
func (self _TiledDataHandler) Store(dir string, name string, data ISpeedUpData) {
	tiled_data := data.(*_TiledData)

	_StoreIDMapping(tiled_data.id_mapping, dir+name+"-mapping")
	_StoreShortcuts(tiled_data.skip_shortcuts, dir+name+"-skip_shortcuts")
	_StoreAdjacency(&tiled_data.skip_topology, true, dir+name+"-skip_topology")
	WriteArrayToFile[int16](tiled_data.node_tiles, dir+name+"-tiles")
	WriteArrayToFile[byte](tiled_data.edge_types, dir+name+"-tiles_types")
	if tiled_data.cell_index.HasValue() {
		_StoreCellIndex(tiled_data.cell_index.Value, dir+name+"-tileranges")
	}
	meta := _TiledDataMeta{
		BaseWeighting: tiled_data._base_weighting,
	}
	WriteJSONToFile(meta, dir+name+"-meta")
}
func (self _TiledDataHandler) Remove(dir string, name string) {
	os.Remove(dir + name + "-mapping")
	os.Remove(dir + name + "-skip_shortcuts")
	os.Remove(dir + name + "-skip_topology")
	os.Remove(dir + name + "-tiles")
	os.Remove(dir + name + "-tiles_types")
	os.Remove(dir + name + "-tileranges")
	os.Remove(dir + name + "-meta")
}
func (self _TiledDataHandler) _ReorderNodes(dir string, name string, mapping Array[int32]) {
	id_mapping := _LoadIDMapping(dir + name + "-mapping")
	id_mapping.ReorderSources(mapping)
	_StoreIDMapping(id_mapping, dir+name+"-mapping")
}
func (self _TiledDataHandler) _ReorderNodesInplace(data ISpeedUpData, mapping Array[int32]) {
	tiled_data := data.(*_TiledData)
	tiled_data.id_mapping.ReorderSources(mapping)
}

type _TiledDataMeta struct {
	BaseWeighting string `json:base_weighting`
}

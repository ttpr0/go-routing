package graph

import (
	"runtime"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// speed-up data
//*******************************************

type SpeedUpType byte

const (
	CH    SpeedUpType = 0
	TILED SpeedUpType = 1
)

type ISpeedUpData interface {
	Type() SpeedUpType
}

type ISpeedUpHandler interface {
	Load(dir string, name string, nodecount int) ISpeedUpData
	Store(dir string, name string, data ISpeedUpData)
	Remove(dir string, name string)
	_ReorderNodes(dir string, name string, mapping Array[int32])  // Reorder nodes of base-graph
	_ReorderNodesInplace(data ISpeedUpData, mapping Array[int32]) // Reorder nodes of base-graph
}

var SPEEDUP_HANDLERS = Dict[SpeedUpType, ISpeedUpHandler]{
	CH:    _CHDataHandler{},
	TILED: _TiledDataHandler{},
}

//*******************************************
// graph-dir
//*******************************************

func NewGraphDir(base GraphBase, dir string) GraphDir {
	meta := _GraphDirMeta{
		NodeCount:  int32(base.NodeCount()),
		EdgeCount:  int32(base.EdgeCount()),
		Weightings: NewList[_WeightingMeta](2),
		SpeedUps:   NewList[_SpeedUpMeta](2),
	}
	_StoreGraphDirMeta(meta, dir+"-meta")
	return GraphDir{
		base_dir: dir,
		metadata: meta,

		base: base,

		weightings: NewDict[string, IWeighting](2),
		speed_ups:  NewDict[string, ISpeedUpData](2),
	}
}

func OpenGraphDir(dir string) GraphDir {
	metadata := _LoadGraphDirMeta(dir + "-meta")

	store := _LoadGraphStorage(dir)
	nodecount := store.NodeCount()
	topology := _LoadAdjacency(dir+"-graph", false, nodecount)
	index := _BuildKDTreeIndex(store)

	return GraphDir{
		base_dir: dir,
		metadata: metadata,

		base: GraphBase{
			store:    store,
			topology: *topology,
			index:    index,
		},

		weightings: NewDict[string, IWeighting](2),
		speed_ups:  NewDict[string, ISpeedUpData](2),
	}
}

type GraphDir struct {
	base_dir string
	metadata _GraphDirMeta

	base GraphBase

	weightings Dict[string, IWeighting]

	speed_ups Dict[string, ISpeedUpData]
}

func (self *GraphDir) GetGraphBase() GraphBase {
	return self.base
}

func (self *GraphDir) StoreGraphBase() {
	_StoreAdjacency(&self.base.topology, false, self.base_dir+"-graph")
	_StoreGraphStorage(self.base.store, self.base_dir)
}

//*******************************************
// modify graph-dir data
//*******************************************

// reorders node information of base-graph,
// mapping: old id -> new id
//
// If "sync" is true also unloaded speed-ups and weightings will get updated.
func (self *GraphDir) ReorderBaseNodes(mapping Array[int32], sync bool) {
	self.base._ReorderNodes(mapping)
	if sync {
		self.StoreGraphBase()
	}
	for _, w_m := range self.metadata.Weightings {
		if self.speed_ups.ContainsKey(w_m.Name) {
			w := self.weightings[w_m.Name]
			w_h := WEIGHTING_HANDLERS[w.Type()]
			w_h._ReorderNodesInplace(w, mapping)
			if sync {
				w_h.Store(self.base_dir, w_m.Name, w)
			}
		} else if sync {
			su_h := WEIGHTING_HANDLERS[w_m.Type]
			su_h._ReorderNodes(self.base_dir, w_m.Name, mapping)
		}
	}
	for _, su_m := range self.metadata.SpeedUps {
		if self.speed_ups.ContainsKey(su_m.Name) {
			su := self.speed_ups[su_m.Name]
			su_h := SPEEDUP_HANDLERS[su.Type()]
			su_h._ReorderNodesInplace(su, mapping)
			if sync {
				su_h.Store(self.base_dir, su_m.Name, su)
			}
		} else if sync {
			su_h := SPEEDUP_HANDLERS[su_m.Type]
			su_h._ReorderNodes(self.base_dir, su_m.Name, mapping)
		}
	}
}

//*******************************************
// get graphs
//*******************************************

func GetBaseGraph(dir GraphDir, weight string) *Graph {
	return &Graph{
		base:         dir.base,
		weight:       dir.weightings[weight],
		_weight_name: weight,
	}
}

func GetCHGraph(dir GraphDir, name string) *CHGraph {
	su := dir.speed_ups[name]
	data := su.(*_CHData)

	return &CHGraph{
		base:   dir.base,
		weight: dir.weightings[data._base_weighting],

		id_mapping: data.id_mapping,

		_build_with_tiles: data._build_with_tiles,

		ch_shortcuts: data.shortcuts,
		ch_topology:  data.topology,
		node_levels:  data.node_levels,
	}
}

func GetCHGraph2(dir GraphDir, name string) *CHGraph2 {
	su := dir.speed_ups[name]
	data := su.(*_CHData)

	return &CHGraph2{
		base:   dir.base,
		weight: dir.weightings[data._base_weighting],

		id_mapping: data.id_mapping,

		_build_with_tiles: data._build_with_tiles,

		ch_shortcuts: data.shortcuts,
		ch_topology:  data.topology,
		node_levels:  data.node_levels,
	}
}

func GetTiledGraph(dir GraphDir, name string) *TiledGraph2 {
	su := dir.speed_ups[name]
	data := su.(*_TiledData)

	return &TiledGraph2{
		base:   dir.base,
		weight: dir.weightings[data._base_weighting],

		id_mapping: data.id_mapping,

		skip_shortcuts: data.skip_shortcuts,
		skip_topology:  data.skip_topology,
		node_tiles:     data.node_tiles,
		edge_types:     data.edge_types,
		cell_index:     data.cell_index,
	}
}

//*******************************************
// manage weightings
//*******************************************

func (self *GraphDir) AddWeighting(name string, weight IWeighting) {
	if self.weightings.ContainsKey(name) {
		panic("weighting already exists, remove first")
	}
	self.weightings[name] = weight
	self.metadata.Weightings.Add(_WeightingMeta{
		Name: name,
		Type: weight.Type(),
	})
	_StoreGraphDirMeta(self.metadata, self.base_dir+"-meta")
}
func (self *GraphDir) GetWeighting(name string) IWeighting {
	if !self.weightings.ContainsKey(name) {
		panic("weighting " + name + " doosnt exist")
	}
	return self.weightings[name]
}
func (self *GraphDir) RemoveWeighting(name string, typ WeightType) {
	w_handler := WEIGHTING_HANDLERS[typ]
	w_handler.Remove(self.base_dir, name)
	if self.weightings.ContainsKey(name) {
		self.weightings.Delete(name)
		runtime.GC()
	}
	i := FindFirst(self.metadata.Weightings, func(w_m _WeightingMeta) bool {
		if w_m.Name == name {
			return true
		} else {
			return false
		}
	})
	if i == -1 {
		panic("this should not have happened")
	}
	self.metadata.Weightings.Remove(i)
	_StoreGraphDirMeta(self.metadata, self.base_dir+"-meta")
}
func (self *GraphDir) LoadWeighting(name string, typ WeightType) {
	if self.weightings.ContainsKey(name) {
		panic("weighting " + name + " is already loaded")
	}
	w_handler := WEIGHTING_HANDLERS[typ]
	weight := w_handler.Load(self.base_dir, name, self.base.NodeCount(), self.base.EdgeCount())
	self.weightings[name] = weight
}
func (self *GraphDir) UnloadWeighting(name string) {
	if !self.weightings.ContainsKey(name) {
		panic("weighting " + name + " doosnt exist")
	}
	self.weightings.Delete(name)
	runtime.GC()
}
func (self *GraphDir) StoreWeighting(name string) {
	if !self.weightings.ContainsKey(name) {
		panic("weighting " + name + " doosnt exist")
	}
	weight := self.weightings[name]
	w_type := weight.Type()
	w_handler := WEIGHTING_HANDLERS[w_type]
	w_handler.Store(self.base_dir, name, weight)
}

//*******************************************
// manage speed-ups
//*******************************************

func (self *GraphDir) AddSpeedUp(name string, data ISpeedUpData) {
	if self.speed_ups.ContainsKey(name) {
		panic("speed up already exists, remove first")
	}
	self.speed_ups[name] = data
	self.metadata.SpeedUps.Add(_SpeedUpMeta{
		Name: name,
		Type: data.Type(),
	})
	_StoreGraphDirMeta(self.metadata, self.base_dir+"-meta")
}
func (self *GraphDir) RemoveSpeedUp(name string, typ SpeedUpType) {
	su_handler := SPEEDUP_HANDLERS[typ]
	su_handler.Remove(self.base_dir, name)
	if self.speed_ups.ContainsKey(name) {
		self.speed_ups.Delete(name)
		runtime.GC()
	}
	i := FindFirst(self.metadata.SpeedUps, func(s_m _SpeedUpMeta) bool {
		if s_m.Name == name {
			return true
		} else {
			return false
		}
	})
	if i == -1 {
		panic("this should not have happened")
	}
	self.metadata.SpeedUps.Remove(i)
	_StoreGraphDirMeta(self.metadata, self.base_dir+"-meta")
}
func (self *GraphDir) UpdateSpeedUp(name string, data ISpeedUpData) {
	if !self.speed_ups.ContainsKey(name) {
		panic("speed up " + name + " isnt loaded.")
	}
	self.speed_ups[name] = data
}
func (self *GraphDir) LoadSpeedUp(name string, typ SpeedUpType) {
	if self.speed_ups.ContainsKey(name) {
		panic("speed up " + name + " is already loaded")
	}
	su_handler := SPEEDUP_HANDLERS[typ]
	su_data := su_handler.Load(self.base_dir, name, self.base.NodeCount())
	self.speed_ups[name] = su_data
}
func (self *GraphDir) UnloadSpeedUp(name string) {
	if !self.speed_ups.ContainsKey(name) {
		panic("speed up " + name + " doosnt exist")
	}
	self.speed_ups.Delete(name)
	runtime.GC()
}
func (self *GraphDir) StoreSpeedUp(name string) {
	if !self.speed_ups.ContainsKey(name) {
		panic("speed up " + name + " doosnt exist")
	}
	speed_up := self.speed_ups[name]
	su_type := speed_up.Type()
	su_handler := SPEEDUP_HANDLERS[su_type]
	su_handler.Store(self.base_dir, name, speed_up)
}

//*******************************************
// graph-dir metadata
//*******************************************

type _GraphDirMeta struct {
	NodeCount  int32
	EdgeCount  int32
	Weightings List[_WeightingMeta]
	SpeedUps   List[_SpeedUpMeta]
}

type _WeightingMeta struct {
	Name string     `json:"name"`
	Type WeightType `json:"type"`
}

type _SpeedUpMeta struct {
	Name string      `json:"name"`
	Type SpeedUpType `json:"type"`
}

func _LoadGraphDirMeta(file string) _GraphDirMeta {
	meta := ReadJSONFromFile[_GraphDirMeta](file)
	return meta
}
func _StoreGraphDirMeta(meta _GraphDirMeta, file string) {
	WriteJSONToFile(meta, file)
}

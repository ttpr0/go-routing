package graph

import (
	"runtime"

	. "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"
)

//*******************************************
// graph-dir
//*******************************************

func NewGraphDir(dir string) GraphDir {
	meta := _GraphDirMeta{
		Bases:      NewList[_BaseMeta](1),
		Weightings: NewList[_WeightingMeta](2),
		SpeedUps:   NewList[_SpeedUpMeta](2),
	}
	_StoreGraphDirMeta(meta, dir+"-meta")
	return GraphDir{
		base_dir: dir,
		metadata: meta,

		bases:      NewDict[string, *GraphBase](1),
		weightings: NewDict[string, IWeighting](2),
		speed_ups:  NewDict[string, ISpeedUpData](2),
	}
}

func OpenGraphDir(dir string) GraphDir {
	metadata := _LoadGraphDirMeta(dir + "-meta")

	return GraphDir{
		base_dir: dir,
		metadata: metadata,

		bases:      NewDict[string, *GraphBase](1),
		weightings: NewDict[string, IWeighting](2),
		speed_ups:  NewDict[string, ISpeedUpData](2),
	}
}

type GraphDir struct {
	base_dir string
	metadata _GraphDirMeta

	bases Dict[string, *GraphBase]

	weightings Dict[string, IWeighting]

	speed_ups Dict[string, ISpeedUpData]
}

func (self *GraphDir) GetBaseDir() string {
	return self.base_dir
}
func (self *GraphDir) GraphBaseCount() int {
	return self.metadata.Bases.Length()
}
func (self *GraphDir) GraphBaseMetadata(index int) _BaseMeta {
	return self.metadata.Bases[index]
}
func (self *GraphDir) WeightingCount() int {
	return self.metadata.Weightings.Length()
}
func (self *GraphDir) WeightingMetadata(index int) _WeightingMeta {
	return self.metadata.Weightings[index]
}

func (self *GraphDir) SpeedUpCount() int {
	return self.metadata.SpeedUps.Length()
}
func (self *GraphDir) SpeedUpMetadata(index int) _SpeedUpMeta {
	return self.metadata.SpeedUps[index]
}

//*******************************************
// manage graph bases
//*******************************************

// TODO: join add+store and load+get together

func (self *GraphDir) AddGraphBase(name string, base GraphBase) {
	if self.bases.ContainsKey(name) {
		panic("base already exists, remove first")
	}
	self.bases[name] = &base
	self.metadata.AddBaseMeta(_BaseMeta{
		Name: name,
	})
	_StoreGraphDirMeta(self.metadata, self.base_dir+"-meta")
}
func (self *GraphDir) GetGraphBase(name string) *GraphBase {
	return self.bases[name]
}
func (self *GraphDir) LoadGraphBase(name string) {
	if self.bases.ContainsKey(name) {
		panic("base " + name + " is already loaded")
	}
	store := _LoadGraphStorage(self.base_dir + name)
	nodecount := store.NodeCount()
	topology := _LoadAdjacency(self.base_dir+name+"-graph", false, nodecount)
	index := _BuildKDTreeIndex(store)
	self.bases[name] = &GraphBase{
		store:    store,
		topology: *topology,
		index:    index,
	}
}
func (self *GraphDir) UnloadGraphBase(name string) {
	if !self.bases.ContainsKey(name) {
		panic("base " + name + " doosnt exist")
	}
	self.bases.Delete(name)
	runtime.GC()
}
func (self *GraphDir) RemoveGraphBase(name string) {
	panic("TODO: Remove files of graph-base")
	if self.bases.ContainsKey(name) {
		self.bases.Delete(name)
		runtime.GC()
	}
	self.metadata.RemoveBaseMeta(name)
	_StoreGraphDirMeta(self.metadata, self.base_dir+"-meta")
}
func (self *GraphDir) StoreGraphBase(name string) {
	if !self.bases.ContainsKey(name) {
		panic("base " + name + " doosnt exist")
	}
	base := self.bases[name]
	_StoreAdjacency(&base.topology, false, self.base_dir+name+"-graph")
	_StoreGraphStorage(base.store, self.base_dir+name)
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
func (self *GraphDir) RemoveWeighting(name string) {
	meta := self.metadata.GetWeightMeta(name)
	w_handler := WEIGHTING_HANDLERS[meta.Type]
	w_handler.Remove(self.base_dir, name)
	if self.weightings.ContainsKey(name) {
		self.weightings.Delete(name)
		runtime.GC()
	}
	self.metadata.RemoveWeightMeta(name)
	_StoreGraphDirMeta(self.metadata, self.base_dir+"-meta")
}
func (self *GraphDir) LoadWeighting(name string) {
	if self.weightings.ContainsKey(name) {
		panic("weighting " + name + " is already loaded")
	}
	meta := self.metadata.GetWeightMeta(name)
	w_handler := WEIGHTING_HANDLERS[meta.Type]
	weight := w_handler.Load(self.base_dir, name)
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

func (self *GraphDir) GetSpeedUp(name string) ISpeedUpData {
	if !self.speed_ups.ContainsKey(name) {
		panic("speed-up " + name + " doosnt exist")
	}
	return self.speed_ups[name]
}
func (self *GraphDir) AddSpeedUp(name string, data ISpeedUpData) {
	if self.speed_ups.ContainsKey(name) {
		panic("speed up already exists, remove first")
	}
	self.speed_ups[name] = data
	self.metadata.AddSpeedUpMeta(_SpeedUpMeta{
		Name: name,
		Type: data.Type(),
	})
	_StoreGraphDirMeta(self.metadata, self.base_dir+"-meta")
}
func (self *GraphDir) RemoveSpeedUp(name string) {
	meta := self.metadata.GetSpeedUpMeta(name)
	su_handler := SPEEDUP_HANDLERS[meta.Type]
	su_handler.Remove(self.base_dir, name)
	if self.speed_ups.ContainsKey(name) {
		self.speed_ups.Delete(name)
		runtime.GC()
	}
	self.metadata.RemoveSpeedUpMeta(name)
	_StoreGraphDirMeta(self.metadata, self.base_dir+"-meta")
}
func (self *GraphDir) UpdateSpeedUp(name string, data ISpeedUpData) {
	if !self.speed_ups.ContainsKey(name) {
		panic("speed up " + name + " isnt loaded.")
	}
	self.speed_ups[name] = data
}
func (self *GraphDir) LoadSpeedUp(name string) {
	if self.speed_ups.ContainsKey(name) {
		panic("speed up " + name + " is already loaded")
	}
	meta := self.metadata.GetSpeedUpMeta(name)
	su_handler := SPEEDUP_HANDLERS[meta.Type]
	su_data := su_handler.Load(self.base_dir, name)
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
	Bases      List[_BaseMeta]      `json:"bases"`
	Weightings List[_WeightingMeta] `json:"weightings"`
	SpeedUps   List[_SpeedUpMeta]   `json:"speedups"`
}

func (self *_GraphDirMeta) _FindBaseMeta(name string) int {
	return FindFirst(self.Bases, func(m _BaseMeta) bool {
		if m.Name == name {
			return true
		} else {
			return false
		}
	})
}
func (self *_GraphDirMeta) AddBaseMeta(meta _BaseMeta) {
	i := self._FindBaseMeta(meta.Name)
	if i == -1 {
		self.Bases.Add(meta)
	} else {
		panic("meta already exists")
	}
}
func (self *_GraphDirMeta) RemoveBaseMeta(name string) {
	i := self._FindBaseMeta(name)
	if i != -1 {
		self.Bases.Remove(i)
	}
}
func (self *_GraphDirMeta) GetBaseMeta(name string) _BaseMeta {
	i := self._FindBaseMeta(name)
	if i != -1 {
		return self.Bases[i]
	} else {
		panic("meta doesnt exist")
	}
}

func (self *_GraphDirMeta) _FindWeightMeta(name string) int {
	return FindFirst(self.Weightings, func(m _WeightingMeta) bool {
		if m.Name == name {
			return true
		} else {
			return false
		}
	})
}
func (self *_GraphDirMeta) AddWeightMeta(meta _WeightingMeta) {
	i := self._FindWeightMeta(meta.Name)
	if i == -1 {
		self.Weightings.Add(meta)
	} else {
		panic("meta already exists")
	}
}
func (self *_GraphDirMeta) RemoveWeightMeta(name string) {
	i := self._FindWeightMeta(name)
	if i != -1 {
		self.Weightings.Remove(i)
	}
}
func (self *_GraphDirMeta) GetWeightMeta(name string) _WeightingMeta {
	i := self._FindWeightMeta(name)
	if i != -1 {
		return self.Weightings[i]
	} else {
		panic("meta doesnt exist")
	}
}

func (self *_GraphDirMeta) _FindSpeedUpMeta(name string) int {
	return FindFirst(self.SpeedUps, func(m _SpeedUpMeta) bool {
		if m.Name == name {
			return true
		} else {
			return false
		}
	})
}
func (self *_GraphDirMeta) AddSpeedUpMeta(meta _SpeedUpMeta) {
	i := self._FindSpeedUpMeta(meta.Name)
	if i == -1 {
		self.SpeedUps.Add(meta)
	} else {
		panic("meta already exists")
	}
}
func (self *_GraphDirMeta) RemoveSpeedUpMeta(name string) {
	i := self._FindSpeedUpMeta(name)
	if i != -1 {
		self.SpeedUps.Remove(i)
	}
}
func (self *_GraphDirMeta) GetSpeedUpMeta(name string) _SpeedUpMeta {
	i := self._FindSpeedUpMeta(name)
	if i != -1 {
		return self.SpeedUps[i]
	} else {
		panic("meta doesnt exist")
	}
}

type _BaseMeta struct {
	Name string `json:"name"`
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

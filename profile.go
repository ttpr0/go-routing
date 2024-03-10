package main

import (
	"encoding/json"
	"fmt"

	"github.com/ttpr0/go-routing/attr"
	"github.com/ttpr0/go-routing/comps"
	"github.com/ttpr0/go-routing/graph"
	"github.com/ttpr0/go-routing/parser"
	"github.com/ttpr0/go-routing/preproc"
	"github.com/ttpr0/go-routing/structs"
	. "github.com/ttpr0/go-routing/util"
)

//**********************************************************
// profile
//**********************************************************

type IRoutingProfile interface {
	Profile() ProfileType
	Vehicle() VehicleType
	Metric() MetricType
	SetManager(manager *RoutingManager)

	GetGraph() Optional[graph.IGraph]
	GetCHGraph() Optional[graph.ICHGraph]
	GetTiledGraph() Optional[graph.ITiledGraph]
	GetTransitGraph(schedule string) Optional[*graph.TransitGraph]

	GetAttributes() attr.IAttributes

	_GetMetadata() ProfileMeta
}

type PrepDict = Dict[ProfileType, Tuple[*comps.GraphBase, *attr.GraphAttributes]]

type ProfileHandler struct {
	Build func(string, SourceOptions, IProfileOptions, PrepDict) IRoutingProfile
	Load  func(string, ProfileMeta) IRoutingProfile
}

var PROFILE_HANDLERS = Dict[ProfileType, ProfileHandler]{
	DRIVING: {
		Build: BuildDrivingProfile,
		Load:  LoadDrivingProfile,
	},
}

type ProfileMeta struct {
	Type ProfileType     `json:"type"`
	Meta json.RawMessage `json:"meta"`
}

//**********************************************************
// driving profile
//**********************************************************

type DrivingProfile struct {
	manager *RoutingManager
	metric  MetricType
	vehicle VehicleType

	base              comps.IGraphBase
	index             Optional[comps.IGraphIndex]
	attr_node_mapping Optional[structs.IDMapping]
	attr_edge_mapping Optional[structs.IDMapping]
	weight            Optional[comps.IWeighting]
	tc_weight         Optional[comps.ITCWeighting]
	ch_speed_up       Optional[DrivingCHSpeedUp]
	overlay_speed_up  Optional[DrivingOverlaySpeedUp]
}

type DrivingCHSpeedUp struct {
	ch       *comps.CH
	ch_index Optional[*comps.CHIndex]
}
type DrivingOverlaySpeedUp struct {
	partition  *comps.Partition
	overlay    *comps.Overlay
	cell_index *comps.CellIndex
}

func (self *DrivingProfile) Profile() ProfileType {
	return DRIVING
}
func (self *DrivingProfile) Vehicle() VehicleType {
	return self.vehicle
}
func (self *DrivingProfile) Metric() MetricType {
	return self.metric
}
func (self *DrivingProfile) SetManager(manager *RoutingManager) {
	self.manager = manager
}
func (self *DrivingProfile) GetGraph() Optional[graph.IGraph] {
	base := self.base
	if !self.index.HasValue() {
		self.index = Some(comps.NewGraphIndex(base))
	}
	index := self.index
	if self.weight.HasValue() {
		g := graph.BuildGraph(base, self.weight.Value, index)
		return Some(graph.IGraph(g))
	}
	if self.tc_weight.HasValue() {
		g := graph.BuildTCGraph(base, self.tc_weight.Value, index)
		return Some(graph.IGraph(g))
	}
	return None[graph.IGraph]()
}
func (self *DrivingProfile) GetCHGraph() Optional[graph.ICHGraph] {
	base := self.base
	if !self.index.HasValue() {
		self.index = Some(comps.NewGraphIndex(base))
	}
	index := self.index
	if !self.weight.HasValue() {
		return None[graph.ICHGraph]()
	}
	weight := self.weight.Value
	if !self.ch_speed_up.HasValue() {
		return None[graph.ICHGraph]()
	}
	speed_up := self.ch_speed_up.Value
	ch := speed_up.ch
	if !speed_up.ch_index.HasValue() {
		speed_up.ch_index = Some(preproc.PreparePHASTIndex(base, weight, ch))
		self.ch_speed_up = Some(speed_up)
	}
	ch_index := speed_up.ch_index
	g := graph.BuildCHGraph(base, weight, index, ch, ch_index)
	return Some(graph.ICHGraph(g))
}
func (self *DrivingProfile) GetTiledGraph() Optional[graph.ITiledGraph] {
	base := self.base
	if !self.index.HasValue() {
		self.index = Some(comps.NewGraphIndex(base))
	}
	index := self.index
	if !self.weight.HasValue() {
		return None[graph.ITiledGraph]()
	}
	weight := self.weight.Value
	if !self.overlay_speed_up.HasValue() {
		return None[graph.ITiledGraph]()
	}
	speed_up := self.overlay_speed_up.Value
	partition := speed_up.partition
	overlay := speed_up.overlay
	cell_index := speed_up.cell_index
	g := graph.BuildTiledGraph(base, weight, index, partition, overlay, Some(cell_index))
	return Some(graph.ITiledGraph(g))
}
func (self *DrivingProfile) GetTransitGraph(schedule string) Optional[*graph.TransitGraph] {
	return None[*graph.TransitGraph]()
}
func (self *DrivingProfile) GetAttributes() attr.IAttributes {
	att := self.manager._GetAttributes(DRIVING)
	return attr.NewMappedAttributes(att, self.attr_node_mapping, self.attr_edge_mapping)
}
func (self *DrivingProfile) _GetMetadata() ProfileMeta {
	meta := DrivingMeta{
		Metric:  self.metric,
		Vehicle: self.vehicle,

		TurnCosts: self.tc_weight.HasValue(),

		CH:      self.ch_speed_up.HasValue(),
		Overlay: self.overlay_speed_up.HasValue(),
	}
	meta_str, _ := json.Marshal(meta)
	return ProfileMeta{
		Type: DRIVING,
		Meta: meta_str,
	}
}

type DrivingMeta struct {
	Metric  MetricType  `json:"metric"`
	Vehicle VehicleType `json:"vehicle"`

	TurnCosts bool `json:"turn-costs"`

	CH      bool `json:"ch"`
	Overlay bool `json:"overlay"`
}

func LoadDrivingProfile(path string, p_meta ProfileMeta) IRoutingProfile {
	if p_meta.Type != DRIVING {
		panic("not a driving profile")
	}
	meta := DrivingMeta{}
	json.Unmarshal(p_meta.Meta, &meta)

	prefix := path

	base := comps.Load[*comps.GraphBase](prefix + "-base")
	attr_node_mapping := structs.LoadIDMapping(prefix + "-attr_node_mapping")
	var weight Optional[comps.IWeighting]
	var tc_weight Optional[comps.ITCWeighting]
	if meta.TurnCosts {
		weight = None[comps.IWeighting]()
		tc_weight = Some(comps.ITCWeighting(comps.Load[*comps.TCWeighting](prefix + "-weight")))
	} else {
		weight = Some(comps.IWeighting(comps.Load[*comps.DefaultWeighting](prefix + "-weight")))
		tc_weight = None[comps.ITCWeighting]()
	}
	var ch_speed_up Optional[DrivingCHSpeedUp]
	var overlay_speed_up Optional[DrivingOverlaySpeedUp]
	if meta.CH {
		ch_speed_up = Some(DrivingCHSpeedUp{
			ch: comps.Load[*comps.CH](prefix + "-ch"),
		})
	} else if meta.Overlay {
		overlay_speed_up = Some(DrivingOverlaySpeedUp{
			partition:  comps.Load[*comps.Partition](prefix + "-partition"),
			overlay:    comps.Load[*comps.Overlay](prefix + "-overlay"),
			cell_index: comps.Load[*comps.CellIndex](prefix + "-cell_index"),
		})
	}

	return &DrivingProfile{
		metric:  meta.Metric,
		vehicle: meta.Vehicle,

		base:              base,
		attr_node_mapping: Some(attr_node_mapping),
		weight:            weight,
		tc_weight:         tc_weight,

		ch_speed_up:      ch_speed_up,
		overlay_speed_up: overlay_speed_up,
	}
}

func BuildDrivingProfile(out_path string, source_ SourceOptions, options_ IProfileOptions, prep_cache PrepDict) IRoutingProfile {
	options := options_.(DrivingOptions)
	osm := source_.OSM

	var base *comps.GraphBase
	var attributes *attr.GraphAttributes
	if prep_cache.ContainsKey(DRIVING) {
		item := prep_cache.Get(DRIVING)
		base = item.A
		attributes = item.B
	} else {
		// parse graph from osm
		base, attributes = parser.ParseGraph(osm)
		// remove closely connected components
		remove := RemoveConnectedComponents(base)
		fmt.Println("remove", remove.Length(), "nodes")
		base = comps.RemoveNodes(base, remove)
		attributes.RemoveNodes(remove)
		prep_cache.Set(DRIVING, MakeTuple(base, attributes))
	}

	// build profile
	profile := &DrivingProfile{
		metric:  options.Metric,
		vehicle: options.Vehicle,
	}

	// build metric
	var weight *comps.DefaultWeighting
	switch profile.metric {
	case FASTEST:
		weight = BuildFastestWeighting(base, attributes)
	case SHORTEST:
		weight = BuildShortestWeighting(base, attributes)
	default:
		panic("unknown metric-type")
	}

	// store prefix
	prefix := out_path

	// node mapping of attributes of nodes are reordered
	attr_node_mapping := structs.NewIdendityMapping(base.NodeCount())

	if options.Preperation.Contraction {
		new_base, ch, ordering := CreateCH(base, weight)
		base = new_base
		attr_node_mapping.ReorderTargets(ordering)
		profile.attr_node_mapping = Some(attr_node_mapping)
		profile.base = base
		profile.weight = Some(comps.IWeighting(weight))
		ch_speed_up := DrivingCHSpeedUp{
			ch: ch,
		}
		profile.ch_speed_up = Some(ch_speed_up)
		structs.StoreIDMapping(attr_node_mapping, prefix+"-attr_node_mapping")
		comps.Store(base, prefix+"-base")
		comps.Store(weight, prefix+"-weight")
		comps.Store(ch, prefix+"-ch")
	} else if options.Preperation.Overlay {
		partition := CreatePartition(base, options.Preperation.MaxNodesPerCell)
		var overlay *comps.Overlay
		var cell_index *comps.CellIndex
		var ordering Array[int32]
		switch options.Preperation.OverlayMethod {
		case "skeleton":
			base, partition, overlay, cell_index, ordering = CreateGRASP(base, weight, partition, true)
		case "isophast":
			base, partition, overlay, cell_index, ordering = CreateIsoPHAST(base, weight, partition)
		default:
			base, partition, overlay, cell_index, ordering = CreateGRASP(base, weight, partition, false)
		}
		attr_node_mapping.ReorderTargets(ordering)
		profile.attr_node_mapping = Some(attr_node_mapping)
		profile.base = base
		profile.weight = Some(comps.IWeighting(weight))
		overlay_speed_up := DrivingOverlaySpeedUp{
			partition:  partition,
			overlay:    overlay,
			cell_index: cell_index,
		}
		profile.overlay_speed_up = Some(overlay_speed_up)
		structs.StoreIDMapping(attr_node_mapping, prefix+"-attr_node_mapping")
		comps.Store(base, prefix+"-base")
		comps.Store(weight, prefix+"-weight")
		comps.Store(overlay, prefix+"-overlay")
		comps.Store(cell_index, prefix+"-cell_index")
		comps.Store(partition, prefix+"-partition")
	} else {
		profile.attr_node_mapping = Some(attr_node_mapping)
		profile.base = base
		profile.weight = Some(comps.IWeighting(weight))
		structs.StoreIDMapping(attr_node_mapping, prefix+"-attr_node_mapping")
		comps.Store(base, prefix+"-base")
		comps.Store(weight, prefix+"-weight")
	}

	return profile
}

//**********************************************************
// walking profile
//**********************************************************

type WalkingProfile struct {
	manager *RoutingManager
	metric  MetricType
	vehicle VehicleType

	base      comps.IGraphBase
	index     Optional[comps.IGraphIndex]
	weight    Optional[comps.IWeighting]
	tc_weight Optional[comps.ITCWeighting]
}

func (self *WalkingProfile) Profile() ProfileType {
	return TRANSIT
}
func (self *WalkingProfile) Vehicle() VehicleType {
	return self.vehicle
}
func (self *WalkingProfile) Metric() MetricType {
	return self.metric
}
func (self *WalkingProfile) SetManager(manager *RoutingManager) {
	self.manager = manager
}
func (self *WalkingProfile) GetGraph() Optional[graph.IGraph] {
	base := self.base
	if !self.index.HasValue() {
		self.index = Some(comps.NewGraphIndex(base))
	}
	index := self.index
	if self.weight.HasValue() {
		g := graph.BuildGraph(base, self.weight.Value, index)
		return Some(graph.IGraph(g))
	}
	if self.tc_weight.HasValue() {
		g := graph.BuildTCGraph(base, self.tc_weight.Value, index)
		return Some(graph.IGraph(g))
	}
	return None[graph.IGraph]()
}
func (self *WalkingProfile) GetCHGraph() Optional[graph.ICHGraph] {
	return None[graph.ICHGraph]()
}
func (self *WalkingProfile) GetTiledGraph() Optional[graph.ITiledGraph] {
	return None[graph.ITiledGraph]()
}
func (self *WalkingProfile) GetTransitGraph(schedule string) Optional[*graph.TransitGraph] {
	return None[*graph.TransitGraph]()
}
func (self *WalkingProfile) GetAttributes() attr.IAttributes {
	att := self.manager._GetAttributes(WALKING)
	return attr.NewMappedAttributes(att, None[structs.IDMapping](), None[structs.IDMapping]())
}
func (self *WalkingProfile) _GetMetadata() ProfileMeta {
	meta := WalkingMeta{
		Metric:  self.metric,
		Vehicle: self.vehicle,

		TurnCosts: self.tc_weight.HasValue(),
	}
	meta_str, _ := json.Marshal(meta)
	return ProfileMeta{
		Type: WALKING,
		Meta: meta_str,
	}
}

type WalkingMeta struct {
	Metric  MetricType  `json:"metric"`
	Vehicle VehicleType `json:"vehicle"`

	TurnCosts bool `json:"turn-costs"`
}

func LoadWalkingProfile(path string, p_meta ProfileMeta) IRoutingProfile {
	if p_meta.Type != WALKING {
		panic("not a walking profile")
	}
	meta := WalkingMeta{}
	json.Unmarshal(p_meta.Meta, &meta)

	prefix := path

	base := comps.Load[*comps.GraphBase](prefix + "-base")
	var weight Optional[comps.IWeighting]
	var tc_weight Optional[comps.ITCWeighting]
	if meta.TurnCosts {
		weight = None[comps.IWeighting]()
		tc_weight = Some(comps.ITCWeighting(comps.Load[*comps.TCWeighting](prefix + "-weight")))
	} else {
		weight = Some(comps.IWeighting(comps.Load[*comps.DefaultWeighting](prefix + "-weight")))
		tc_weight = None[comps.ITCWeighting]()
	}

	return &WalkingProfile{
		metric:  meta.Metric,
		vehicle: meta.Vehicle,

		base:      base,
		weight:    weight,
		tc_weight: tc_weight,
	}
}

func BuildWalkingProfile(out_path string, source_ SourceOptions, options_ IProfileOptions, prep_cache PrepDict) IRoutingProfile {
	options := options_.(WalkingOptions)
	osm := source_.OSM

	var base *comps.GraphBase
	var attributes *attr.GraphAttributes
	if prep_cache.ContainsKey(WALKING) {
		item := prep_cache.Get(WALKING)
		base = item.A
		attributes = item.B
	} else {
		// parse graph from osm
		base, attributes = parser.ParseGraph(osm)
		// remove closely connected components
		remove := RemoveConnectedComponents(base)
		fmt.Println("remove", remove.Length(), "nodes")
		base = comps.RemoveNodes(base, remove)
		attributes.RemoveNodes(remove)
		prep_cache.Set(WALKING, MakeTuple(base, attributes))
	}

	// build profile
	profile := &WalkingProfile{
		metric:  options.Metric,
		vehicle: options.Vehicle,
	}

	// build metric
	var weight *comps.DefaultWeighting
	switch profile.metric {
	case FASTEST:
		weight = BuildPedestrianWeighting(base, attributes)
	case SHORTEST:
		weight = BuildShortestWeighting(base, attributes)
	default:
		panic("unknown metric-type")
	}

	// store prefix
	prefix := out_path

	// node mapping of attributes of nodes are reordered
	profile.base = base
	profile.weight = Some(comps.IWeighting(weight))
	comps.Store(base, prefix+"-base")
	comps.Store(weight, prefix+"-weight")

	return profile
}

//**********************************************************
// walking profile
//**********************************************************

type TransitProfile struct {
	manager *RoutingManager
	metric  MetricType
	vehicle VehicleType

	base            comps.IGraphBase
	index           Optional[comps.IGraphIndex]
	tc_weight       comps.ITCWeighting
	transit         *comps.Transit
	transit_weights Dict[string, *comps.TransitWeighting]
}

func (self *TransitProfile) Profile() ProfileType {
	return TRANSIT
}
func (self *TransitProfile) Vehicle() VehicleType {
	return self.vehicle
}
func (self *TransitProfile) Metric() MetricType {
	return self.metric
}
func (self *TransitProfile) SetManager(manager *RoutingManager) {
	self.manager = manager
}
func (self *TransitProfile) GetGraph() Optional[graph.IGraph] {
	base := self.base
	if !self.index.HasValue() {
		self.index = Some(comps.NewGraphIndex(base))
	}
	index := self.index
	g := graph.BuildTCGraph(base, self.tc_weight, index)
	return Some(graph.IGraph(g))
}
func (self *TransitProfile) GetCHGraph() Optional[graph.ICHGraph] {
	return None[graph.ICHGraph]()
}
func (self *TransitProfile) GetTiledGraph() Optional[graph.ITiledGraph] {
	return None[graph.ITiledGraph]()
}
func (self *TransitProfile) GetTransitGraph(schedule string) Optional[*graph.TransitGraph] {
	base := self.base
	if !self.index.HasValue() {
		self.index = Some(comps.NewGraphIndex(base))
	}
	index := self.index
	transit := self.transit
	if !self.transit_weights.ContainsKey(schedule) {
		return None[*graph.TransitGraph]()
	}
	transit_weight := self.transit_weights[schedule]
	g := graph.BuildTransitGraph(base, self.tc_weight, index, transit, transit_weight)
	return Some(g)
}
func (self *TransitProfile) GetAttributes() attr.IAttributes {
	att := self.manager._GetAttributes(WALKING)
	return attr.NewMappedAttributes(att, None[structs.IDMapping](), None[structs.IDMapping]())
}
func (self *TransitProfile) _GetMetadata() ProfileMeta {
	weights := make([]string, 0, 7)
	for k, _ := range self.transit_weights {
		weights = append(weights, k)
	}
	meta := TransitMeta{
		Metric:  self.metric,
		Vehicle: self.vehicle,

		Weights: weights,
	}
	meta_str, _ := json.Marshal(meta)
	return ProfileMeta{
		Type: TRANSIT,
		Meta: meta_str,
	}
}

type TransitMeta struct {
	Metric  MetricType  `json:"metric"`
	Vehicle VehicleType `json:"vehicle"`

	Weights []string `json:"weights"`
}

func LoadTransitProfile(path string, p_meta ProfileMeta) IRoutingProfile {
	if p_meta.Type != TRANSIT {
		panic("not a transit profile")
	}
	meta := TransitMeta{}
	json.Unmarshal(p_meta.Meta, &meta)

	prefix := path

	base := comps.Load[*comps.GraphBase](prefix + "-base")
	tc_weight := comps.Load[*comps.TCWeighting](prefix + "-weight")

	transit := comps.Load[*comps.Transit](prefix + "-transit")
	transit_weights := NewDict[string, *comps.TransitWeighting](7)
	for _, w := range meta.Weights {
		transit_weights[w] = comps.Load[*comps.TransitWeighting](prefix + "-transit-weight-" + w)
	}

	return &TransitProfile{
		metric:  meta.Metric,
		vehicle: meta.Vehicle,

		base:            base,
		tc_weight:       tc_weight,
		transit:         transit,
		transit_weights: transit_weights,
	}
}

func BuildTransitProfile(out_path string, source_ SourceOptions, options_ IProfileOptions, prep_cache PrepDict) IRoutingProfile {
	options := options_.(TransitOptions)
	osm := source_.OSM
	gtfs := source_.GTFS

	var base *comps.GraphBase
	var attributes *attr.GraphAttributes
	if prep_cache.ContainsKey(WALKING) {
		item := prep_cache.Get(WALKING)
		base = item.A
		attributes = item.B
	} else {
		// parse graph from osm
		base, attributes = parser.ParseGraph(osm)
		// remove closely connected components
		remove := RemoveConnectedComponents(base)
		fmt.Println("remove", remove.Length(), "nodes")
		base = comps.RemoveNodes(base, remove)
		attributes.RemoveNodes(remove)
		prep_cache.Set(WALKING, MakeTuple(base, attributes))
	}

	// build profile
	profile := &TransitProfile{
		metric:  FASTEST,
		vehicle: options.Vehicle,
	}

	// build metric
	var weight *comps.DefaultWeighting
	switch profile.metric {
	case FASTEST:
		weight = BuildPedestrianWeighting(base, attributes)
	case SHORTEST:
		weight = BuildShortestWeighting(base, attributes)
	default:
		panic("unknown metric-type")
	}

	// store prefix
	prefix := out_path

	// node mapping of attributes of nodes are reordered
	profile.base = base
	profile.tc_weight = comps.ITCWeighting(weight)
	comps.Store(base, prefix+"-base")
	comps.Store(weight, prefix+"-weight")

	stops, conns, schedules := parser.ParseGtfs(gtfs, options.Preperation.FilterPolygon)
	g := graph.BuildGraph(base, weight, None[comps.IGraphIndex]())
	transit := preproc.PrepareTransit(g, stops, conns, options.Preperation.MaxTransferRange)
	profile.transit = transit
	comps.Store(transit, prefix+"-transit")
	transit_weights := NewDict[string, *comps.TransitWeighting](7)
	for k, schedule := range schedules {
		transit_weight := comps.NewTransitWeighting(transit)
		for i := 0; i < transit.ConnectionCount(); i++ {
			s := schedule[i]
			transit_weight.SetWeights(int32(i), s)
		}
		transit_weights[k] = transit_weight
		comps.Store(transit_weight, prefix+"-transit-weight-"+k)
	}
	profile.transit_weights = transit_weights

	return profile
}

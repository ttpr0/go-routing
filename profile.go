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
	"golang.org/x/exp/slog"
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
	WALKING: {
		Build: BuildWalkingProfile,
		Load:  LoadWalkingProfile,
	},
	TRANSIT: {
		Build: BuildTransitProfile,
		Load:  LoadTransitProfile,
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
	if self.weight.HasValue() {
		g := graph.BuildGraph(base, self.weight.Value)
		return Some(graph.IGraph(g))
	}
	if self.tc_weight.HasValue() {
		g := graph.BuildTCGraph(base, self.tc_weight.Value)
		return Some(graph.IGraph(g))
	}
	return None[graph.IGraph]()
}
func (self *DrivingProfile) GetCHGraph() Optional[graph.ICHGraph] {
	base := self.base
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
	g := graph.BuildCHGraph(base, weight, ch, ch_index)
	return Some(graph.ICHGraph(g))
}
func (self *DrivingProfile) GetTiledGraph() Optional[graph.ITiledGraph] {
	base := self.base
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
	g := graph.BuildTiledGraph(base, weight, partition, overlay, Some(cell_index))
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
	slog.Info("Building driving profile from " + osm)

	var base *comps.GraphBase
	var attributes *attr.GraphAttributes
	if prep_cache.ContainsKey(DRIVING) {
		slog.Info("Using cached graph")
		item := prep_cache.Get(DRIVING)
		base = item.A
		attributes = item.B
	} else {
		slog.Info("Parsing graph...")
		// parse graph from osm
		base, attributes = parser.ParseGraph(osm, &parser.DrivingDecoder{})
		// remove closely connected components
		slog.Info("Removing unconnected components...")
		remove_nodes, remove_edges := RemoveConnectedComponents(base)
		slog.Info(fmt.Sprintf("removed %v nodes", remove_nodes.Length()))
		base = comps.RemoveNodes(base, remove_nodes)
		attributes.RemoveNodes(remove_nodes)
		attributes.RemoveEdges(remove_edges)
		prep_cache.Set(DRIVING, MakeTuple(base, attributes))
		slog.Info("Successfully parsed graph")
	}

	// build profile
	profile := &DrivingProfile{
		metric:  options.Metric,
		vehicle: options.Vehicle,
	}

	// build metric
	slog.Info("Building metric: " + profile.metric.String())
	var weight *comps.DefaultWeighting
	switch profile.metric {
	case FASTEST:
		switch profile.vehicle {
		case CAR:
			weight = BuildCarWeighting(base, attributes)
		default:
			weight = BuildCarWeighting(base, attributes)
		}
	case SHORTEST:
		weight = BuildShortestWeighting(base, attributes)
	default:
		panic("unknown metric-type")
	}

	// store prefix
	prefix := out_path

	// node mapping of attributes of nodes are reordered
	attr_node_mapping := structs.NewIdendityMapping(base.NodeCount())

	if options.Preparation.Contraction {
		slog.Info("Building contraction hierarchy")
		new_base, ch, ordering := CreateCH(base, weight)
		slog.Info("Contraction hierarchy successfully built")
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
	} else if options.Preparation.Overlay {
		slog.Info("Building overlay")
		partition := CreatePartition(base, options.Preparation.MaxNodesPerCell)
		slog.Info("Overlay successfully built")
		var overlay *comps.Overlay
		var cell_index *comps.CellIndex
		var ordering Array[int32]
		slog.Info("Building Cell-Index with method: " + options.Preparation.OverlayMethod)
		switch options.Preparation.OverlayMethod {
		case "skeleton":
			base, partition, overlay, cell_index, ordering = CreateGRASP(base, weight, partition, true)
		case "isophast":
			base, partition, overlay, cell_index, ordering = CreateIsoPHAST(base, weight, partition)
		default:
			base, partition, overlay, cell_index, ordering = CreateGRASP(base, weight, partition, false)
		}
		slog.Info("Cell-Index successfully built")
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
	weight    Optional[comps.IWeighting]
	tc_weight Optional[comps.ITCWeighting]
}

func (self *WalkingProfile) Profile() ProfileType {
	return WALKING
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
	if self.weight.HasValue() {
		g := graph.BuildGraph(base, self.weight.Value)
		return Some(graph.IGraph(g))
	}
	if self.tc_weight.HasValue() {
		g := graph.BuildTCGraph(base, self.tc_weight.Value)
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
		base, attributes = parser.ParseGraph(osm, &parser.WalkingDecoder{})
		// remove closely connected components
		remove_nodes, remove_edges := RemoveConnectedComponents(base)
		slog.Info(fmt.Sprintf("removed %v nodes", remove_nodes.Length()))
		base = comps.RemoveNodes(base, remove_nodes)
		attributes.RemoveNodes(remove_nodes)
		attributes.RemoveEdges(remove_edges)
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
		switch profile.vehicle {
		case FOOT:
			weight = BuildFootWeighting(base, attributes)
		default:
			weight = BuildFootWeighting(base, attributes)
		}
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
// cycling profile
//**********************************************************

type CyclingProfile struct {
	manager *RoutingManager
	metric  MetricType
	vehicle VehicleType

	base      comps.IGraphBase
	weight    Optional[comps.IWeighting]
	tc_weight Optional[comps.ITCWeighting]
}

func (self *CyclingProfile) Profile() ProfileType {
	return CYCLING
}
func (self *CyclingProfile) Vehicle() VehicleType {
	return self.vehicle
}
func (self *CyclingProfile) Metric() MetricType {
	return self.metric
}
func (self *CyclingProfile) SetManager(manager *RoutingManager) {
	self.manager = manager
}
func (self *CyclingProfile) GetGraph() Optional[graph.IGraph] {
	base := self.base
	if self.weight.HasValue() {
		g := graph.BuildGraph(base, self.weight.Value)
		return Some(graph.IGraph(g))
	}
	if self.tc_weight.HasValue() {
		g := graph.BuildTCGraph(base, self.tc_weight.Value)
		return Some(graph.IGraph(g))
	}
	return None[graph.IGraph]()
}
func (self *CyclingProfile) GetCHGraph() Optional[graph.ICHGraph] {
	return None[graph.ICHGraph]()
}
func (self *CyclingProfile) GetTiledGraph() Optional[graph.ITiledGraph] {
	return None[graph.ITiledGraph]()
}
func (self *CyclingProfile) GetTransitGraph(schedule string) Optional[*graph.TransitGraph] {
	return None[*graph.TransitGraph]()
}
func (self *CyclingProfile) GetAttributes() attr.IAttributes {
	att := self.manager._GetAttributes(WALKING)
	return attr.NewMappedAttributes(att, None[structs.IDMapping](), None[structs.IDMapping]())
}
func (self *CyclingProfile) _GetMetadata() ProfileMeta {
	meta := CyclingMeta{
		Metric:  self.metric,
		Vehicle: self.vehicle,

		TurnCosts: self.tc_weight.HasValue(),
	}
	meta_str, _ := json.Marshal(meta)
	return ProfileMeta{
		Type: CYCLING,
		Meta: meta_str,
	}
}

type CyclingMeta struct {
	Metric  MetricType  `json:"metric"`
	Vehicle VehicleType `json:"vehicle"`

	TurnCosts bool `json:"turn-costs"`
}

func LoadCyclingProfile(path string, p_meta ProfileMeta) IRoutingProfile {
	if p_meta.Type != CYCLING {
		panic("not a cycling profile")
	}
	meta := CyclingMeta{}
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

	return &CyclingProfile{
		metric:  meta.Metric,
		vehicle: meta.Vehicle,

		base:      base,
		weight:    weight,
		tc_weight: tc_weight,
	}
}

func BuildCyclingProfile(out_path string, source_ SourceOptions, options_ IProfileOptions, prep_cache PrepDict) IRoutingProfile {
	options := options_.(CyclingOptions)
	osm := source_.OSM

	var base *comps.GraphBase
	var attributes *attr.GraphAttributes
	if prep_cache.ContainsKey(CYCLING) {
		item := prep_cache.Get(CYCLING)
		base = item.A
		attributes = item.B
	} else {
		// parse graph from osm
		base, attributes = parser.ParseGraph(osm, &parser.CyclingDecoder{})
		// remove closely connected components
		remove_nodes, remove_edges := RemoveConnectedComponents(base)
		slog.Info(fmt.Sprintf("removed %v nodes", remove_nodes.Length()))
		base = comps.RemoveNodes(base, remove_nodes)
		attributes.RemoveNodes(remove_nodes)
		attributes.RemoveEdges(remove_edges)
		prep_cache.Set(CYCLING, MakeTuple(base, attributes))
	}

	// build profile
	profile := &CyclingProfile{
		metric:  options.Metric,
		vehicle: options.Vehicle,
	}

	// build metric
	var weight *comps.DefaultWeighting
	switch profile.metric {
	case FASTEST:
		switch profile.vehicle {
		case BIKE:
			weight = BuildBikeWeighting(base, attributes)
		default:
			weight = BuildBikeWeighting(base, attributes)
		}
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
// transit profile
//**********************************************************

type TransitProfile struct {
	manager *RoutingManager
	metric  MetricType
	vehicle VehicleType

	base            comps.IGraphBase
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
	g := graph.BuildTCGraph(base, self.tc_weight)
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
	transit := self.transit
	if !self.transit_weights.ContainsKey(schedule) {
		return None[*graph.TransitGraph]()
	}
	transit_weight := self.transit_weights[schedule]
	g := graph.BuildTransitGraph(base, self.tc_weight, transit, transit_weight)
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
	tc_weight := comps.Load[*comps.DefaultWeighting](prefix + "-weight")

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

	var prep_type ProfileType
	switch options.Vehicle {
	case CAR:
		prep_type = DRIVING
	case BIKE:
		prep_type = CYCLING
	default:
		prep_type = WALKING
	}

	var base *comps.GraphBase
	var attributes *attr.GraphAttributes
	if prep_cache.ContainsKey(prep_type) {
		item := prep_cache.Get(prep_type)
		base = item.A
		attributes = item.B
	} else {
		// parse graph from osm
		base, attributes = parser.ParseGraph(osm, GetDecoder(prep_type))
		// remove closely connected components
		remove_nodes, remove_edges := RemoveConnectedComponents(base)
		slog.Info(fmt.Sprintf("removed %v nodes", remove_nodes.Length()))
		base = comps.RemoveNodes(base, remove_nodes)
		attributes.RemoveNodes(remove_nodes)
		attributes.RemoveEdges(remove_edges)
		prep_cache.Set(prep_type, MakeTuple(base, attributes))
	}

	// build profile
	profile := &TransitProfile{
		metric:  FASTEST,
		vehicle: options.Vehicle,
	}

	// build metric
	var weight *comps.DefaultWeighting
	switch options.Vehicle {
	case FOOT:
		weight = BuildFootWeighting(base, attributes)
	case BIKE:
		weight = BuildBikeWeighting(base, attributes)
	case CAR:
		weight = BuildCarWeighting(base, attributes)
	default:
		weight = BuildFootWeighting(base, attributes)
	}

	// store prefix
	prefix := out_path

	// node mapping of attributes of nodes are reordered
	profile.base = base
	profile.tc_weight = comps.ITCWeighting(weight)
	comps.Store(base, prefix+"-base")
	comps.Store(weight, prefix+"-weight")

	stops, conns, schedules := parser.ParseGtfs(gtfs, options.Preparation.FilterPolygon)
	g := graph.BuildGraph(base, weight)
	transit := preproc.PrepareTransit(g, stops, conns, options.Preparation.MaxTransferRange)
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

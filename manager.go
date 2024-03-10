package main

import (
	"github.com/ttpr0/go-routing/attr"
	"github.com/ttpr0/go-routing/comps"
	. "github.com/ttpr0/go-routing/util"
)

func NewRoutingManager(path string, config Config) *RoutingManager {
	build := config.BuildGraphs
	if IsDirectoryEmpty(path) {
		build = true
	}
	graph_path := path + "/"

	// create manager
	manager := &RoutingManager{
		config: config,
	}

	profiles := NewDict[string, IRoutingProfile](10)
	attributes := NewDict[ProfileType, attr.IAttributes](10)
	// build/load profiles
	if build {
		prep_cache := NewDict[ProfileType, Tuple[*comps.GraphBase, *attr.GraphAttributes]](10)
		profile_meta := NewDict[string, ProfileMeta](10)
		for name, options := range config.Build.Profiles {
			if options.Value == nil {
				continue
			}
			handler := PROFILE_HANDLERS[options.Value.Type()]
			profile := handler.Build(graph_path+name, config.Build.Source, options.Value, prep_cache)
			profile.SetManager(manager)
			profiles.Set(name, profile)
			profile_meta[name] = profile._GetMetadata()
		}
		attr_meta := NewList[ProfileType](4)
		for typ, data := range prep_cache {
			att := data.B
			attr.Store(att, graph_path+"attr-"+typ.String())
			attributes.Set(typ, att)
			attr_meta.Add(typ)
		}
		meta := RoutingManagerMeta{
			Profiles:   profile_meta,
			Attributes: attr_meta,
		}
		WriteJSONToFile(meta, graph_path+"meta")
	} else {
		meta := ReadJSONFromFile[RoutingManagerMeta](graph_path + "meta")
		for name, item := range meta.Profiles {
			handler := PROFILE_HANDLERS[item.Type]
			profile := handler.Load(graph_path+name, item)
			profile.SetManager(manager)
			profiles.Set(name, profile)
		}
		for _, typ := range meta.Attributes {
			attributes.Set(typ, attr.Load(graph_path+"attr-"+typ.String()))
		}
	}

	manager.profiles = profiles
	manager.attributes = attributes
	return manager
}

type RoutingManagerMeta struct {
	Profiles   Dict[string, ProfileMeta] `json:"profiles"`
	Attributes List[ProfileType]         `json:"attributes"`
}

type RoutingManager struct {
	config     Config
	profiles   Dict[string, IRoutingProfile]
	attributes Dict[ProfileType, attr.IAttributes]
}

func (self *RoutingManager) GetProfile(profile string) Optional[IRoutingProfile] {
	if self.profiles.ContainsKey(profile) {
		return Some(self.profiles.Get(profile))
	}
	return None[IRoutingProfile]()
}

func (self *RoutingManager) GetMatchingProfile(profile ProfileType, vehicle VehicleType, metric MetricType) Optional[IRoutingProfile] {
	for _, p := range self.profiles {
		if p.Profile() == profile && p.Vehicle() == vehicle && p.Metric() == metric {
			return Some(p)
		}
	}
	return None[IRoutingProfile]()
}

func (self *RoutingManager) _GetAttributes(profile ProfileType) attr.IAttributes {
	return self.attributes.Get(profile)
}

func (self *RoutingManager) _GetServiceConfig() Config {
	return self.config
}

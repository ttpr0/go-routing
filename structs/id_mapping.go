package structs

import (
	. "github.com/ttpr0/go-routing/util"
)

func NewIDMapping(mapping Array[[2]int32]) IDMapping {
	return IDMapping{
		mapping: mapping,
	}
}

func NewIdendityMapping(size int) IDMapping {
	mapping := NewArray[[2]int32](size)
	for i := 0; i < size; i++ {
		mapping[i] = [2]int32{int32(i), int32(i)}
	}
	return IDMapping{
		mapping: mapping,
	}
}

// Maps indices from a source to a target and reversed.
type IDMapping struct {
	// Contains two mappings:
	//
	// -> first value maps source s to target t: mapping[s][0] = t
	//
	// -> second value maps target t to source s: mapping[t][1] = s
	mapping Array[[2]int32]
}

func (self *IDMapping) GetTarget(source int32) int32 {
	return self.mapping[source][0]
}

func (self *IDMapping) GetSource(target int32) int32 {
	return self.mapping[target][1]
}

// reorders sources,
// mapping: old id -> new id
func (self *IDMapping) ReorderSources(mapping Array[int32]) {
	if self.mapping.Length() != mapping.Length() {
		panic("invalid mapping")
	}
	temp := NewArray[int32](self.mapping.Length())
	for i := 0; i < self.mapping.Length(); i++ {
		// store source to target mapping in temporary array
		temp[i] = self.mapping[i][0]

		// remap target to source mapping
		s := self.mapping[i][1]
		self.mapping[i][1] = mapping[s]
	}
	for i := 0; i < self.mapping.Length(); i++ {
		// remap source to target mapping
		s := int32(i)
		t := temp[s]
		new_s := mapping[s]
		self.mapping[new_s][0] = t
	}
}

// reorders targets,
// mapping: old id -> new id
func (self *IDMapping) ReorderTargets(mapping Array[int32]) {
	if self.mapping.Length() != mapping.Length() {
		panic("invalid mapping")
	}
	temp := NewArray[int32](self.mapping.Length())
	for i := 0; i < self.mapping.Length(); i++ {
		// store target to source mapping in temporary array
		temp[i] = self.mapping[i][1]

		// remap source to target mapping
		t := self.mapping[i][0]
		self.mapping[i][0] = mapping[t]
	}
	for i := 0; i < self.mapping.Length(); i++ {
		// remap target to source mapping
		t := int32(i)
		s := temp[t]
		new_t := mapping[t]
		self.mapping[new_t][1] = s
	}
}

func StoreIDMapping(store IDMapping, file string) {
	WriteArrayToFile[[2]int32](store.mapping, file)
}
func LoadIDMapping(file string) IDMapping {
	store := ReadArrayFromFile[[2]int32](file)
	return IDMapping{
		mapping: store,
	}
}

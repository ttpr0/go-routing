package routing

import (
	"github.com/ttpr0/go-routing/attr"
	"github.com/ttpr0/go-routing/geo"
	"github.com/ttpr0/go-routing/graph"
	. "github.com/ttpr0/go-routing/util"
)

type Path struct {
	path    []int32
	lines   []geo.CoordArray
	graph   graph.IGraph
	changed bool
}

func (self *Path) GetGeometry(attributes attr.IAttributes) []geo.CoordArray {
	if self.lines == nil || self.changed {
		self.lines = make([]geo.CoordArray, 0, 10)
		for _, edge_id := range self.path {
			self.lines = append(self.lines, attributes.GetEdgeGeom(edge_id))
		}
	}
	return self.lines
}
func (self *Path) EdgeIterator() IIterator[int32] {
	return &EdgeIterator{&self.path, 0}
}

type EdgeIterator struct {
	path *[]int32
	curr int
}

func (self *EdgeIterator) Next() (int32, bool) {
	if len(*self.path) < self.curr {
		return 0, false
	} else {
		self.curr += 1
		return (*self.path)[self.curr-1], true
	}
}

func NewPath(graph graph.IGraph, path []int32) Path {
	return Path{graph: graph, path: path}
}

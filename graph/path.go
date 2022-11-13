package graph

import "github.com/ttpr0/simple-routing-visualizer/src/go-routing/util"

type Path struct {
	path      []int32
	lines     []CoordArray
	graph     IGraph
	geometry  IGeometry
	weighting IWeighting
	curr      int32
	changed   bool
}

func (self *Path) GetGeometry() []CoordArray {
	if self.lines == nil || self.changed {
		self.lines = []CoordArray{}
		for i := int(self.curr); i < len(self.path); i = i + 2 {
			self.lines = append(self.lines, self.geometry.GetEdge(self.path[i]))
		}
	}
	return self.lines
}
func (self *Path) EdgeIterator() util.IIterator[int32] {
	return &EdgeIterator{&self.path, int(self.curr)}
}
func (self *Path) Step() bool {
	if int(self.curr) >= len(self.path)-2 {
		return false
	}
	self.curr = self.curr + 2
	return true
}
func (self *Path) GetCurrent() (int32, int32, int32) {
	if int(self.curr) == len(self.path)-2 {
		return self.path[self.curr], self.path[self.curr+1], -1
	}
	return self.path[self.curr], self.path[self.curr+1], self.path[self.curr+2]
}

type EdgeIterator struct {
	path *[]int32
	curr int
}

func (self *EdgeIterator) Next() (int32, bool) {
	if len(*self.path) < self.curr {
		return 0, false
	} else {
		self.curr += 2
		return (*self.path)[self.curr-2], true
	}
}

func NewPath(graph IGraph, path []int32) Path {
	return Path{graph: graph, weighting: graph.GetWeighting(), geometry: graph.GetGeometry(), path: path, curr: 1}
}

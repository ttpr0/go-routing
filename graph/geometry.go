package graph

import "github.com/ttpr0/simple-routing-visualizer/src/go-routing/geo"

type IGeometry interface {
	GetNode(node int32) geo.Coord
	GetEdge(edge int32) geo.CoordArray
	GetAllNodes() []geo.Coord
	GetAllEdges() []geo.CoordArray
}

type Geometry struct {
	NodeGeometry []geo.Coord
	EdgeGeometry []geo.CoordArray
}

func (self *Geometry) GetNode(node int32) geo.Coord {
	return self.NodeGeometry[node]
}
func (self *Geometry) GetEdge(edge int32) geo.CoordArray {
	return self.EdgeGeometry[edge]
}
func (self *Geometry) GetAllNodes() []geo.Coord {
	return self.NodeGeometry
}
func (self *Geometry) GetAllEdges() []geo.CoordArray {
	return self.EdgeGeometry
}

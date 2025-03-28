package isochrone

import (
	"fmt"
	"math"

	"github.com/ttpr0/go-routing/attr"
	"github.com/ttpr0/go-routing/geo"
	"github.com/ttpr0/go-routing/graph"
	"github.com/ttpr0/go-routing/routing"
	. "github.com/ttpr0/go-routing/util"
	"golang.org/x/exp/slog"
)

//**********************************************************
// isochrone handler
//**********************************************************

func ComputeIsochrone(g graph.IGraph, att attr.IAttributes, location [2]float32, ranges []int32) *geo.FeatureCollection {
	start := geo.Coord{location[0], location[1]}
	cellsize := int32(500)
	projection := &WebMercatorProjection{}
	rasterizer := NewRasterizer(cellsize)
	centerx, centery := rasterizer.PointToIndex(projection.Proj(start))
	isosize := int32(2000)
	consumer := &SPTIsochroneConsumer{
		points:     NewIsoTree[int]([4]int32{centerx - isosize + 1, centery - isosize + 1, centerx + isosize - 1, centery + isosize - 1}),
		rasterizer: rasterizer,
		projection: projection,
		att:        att,
	}
	s_node, _ := att.GetClosestNode(start)
	spt := routing.NewShortestPathTree(g, s_node, ranges[len(ranges)-1], consumer)
	slog.Debug(fmt.Sprintf("Start Caluclating shortest-path-tree from %v", start))
	spt.CalcShortestPathTree()
	slog.Debug("shortest-path-tree finished")
	slog.Debug("start building isochrone")
	// create tree containing marching square cells
	mq_tree := NewIsoTree[int]([4]int32{centerx - isosize, centery - isosize, centerx + isosize, centery + isosize})
	features := NewList[geo.Feature](len(ranges))
	for i := len(ranges) - 1; i >= 0; i-- {
		_features := _ExtractIsochrone(consumer.points, mq_tree, int(ranges[i]), rasterizer, projection)
		for _, feature := range _features {
			features.Add(feature)
		}
	}
	resp := geo.NewFeatureCollection(features)
	slog.Debug("reponse build")
	return &resp
}

//**********************************************************
// isochrone builder
//**********************************************************

type SPTIsochroneConsumer struct {
	points     *IsoTree[int]
	rasterizer IRasterizer
	projection IProjection
	att        attr.IAttributes
	linecache  []geo.Coord
}

func (self *SPTIsochroneConsumer) ConsumePoint(point geo.Coord, value int) {
	point = self.projection.Proj(point)
	x, y := self.rasterizer.PointToIndex(point)
	valuefunc := func(other int) int {
		if other == 0 {
			return value
		}
		if value < other {
			return value
		} else {
			return other
		}
	}
	self.points.Insert(x, y, valuefunc)
}

func (self *SPTIsochroneConsumer) ConsumeEdge(edge int32, start_value int, end_value int) {
	geom := self.att.GetEdgeGeom(edge)
	if self.linecache == nil {
		self.linecache = make([]geo.Coord, 0, len(geom))
	} else {
		self.linecache = self.linecache[:0]
	}
	for i, _ := range geom {
		self.linecache = append(self.linecache, self.projection.Proj(geom[i]))
	}
	geom = self.linecache
	total_length := float32(0)
	for i := 0; i < len(geom)-1; i++ {
		total_length += _Dist(geom[i], geom[i+1])
	}
	callback := func(point geo.Coord, length float32) {
		x, y := self.rasterizer.PointToIndex(point)
		value := start_value + int(float32(end_value-start_value)*length/total_length)
		valuefunc := func(other int) int {
			if other == 0 {
				return value
			}
			if value < other {
				return value
			} else {
				return other
			}
		}
		self.points.Insert(x, y, valuefunc)
	}
	_SampleAlongLine(geom, 50, callback)
}

type IProjection interface {
	Proj(geo.Coord) geo.Coord
	ReProj(geo.Coord) geo.Coord
}

type IRasterizer interface {
	PointToIndex(geo.Coord) (int32, int32)
	IndexToPoint(int32, int32) geo.Coord
	GetCellSize() float32
}

type DefaultRasterizer struct {
	factor   float32
	cellsize float32
}

func NewRasterizer(precession int32) *DefaultRasterizer {
	cellsize := float32(precession)
	return &DefaultRasterizer{
		factor:   1 / cellsize,
		cellsize: cellsize,
	}
}

func (self *DefaultRasterizer) PointToIndex(point geo.Coord) (int32, int32) {
	c := point
	return int32(c[0] * self.factor), int32(c[1] * self.factor)
}
func (self *DefaultRasterizer) IndexToPoint(x, y int32) geo.Coord {
	point := geo.Coord{float32(x) / self.factor, float32(y) / self.factor}
	return point
}

func (self *DefaultRasterizer) GetCellSize() float32 {
	return self.cellsize
}

type WebMercatorProjection struct{}

func (self *WebMercatorProjection) Proj(point geo.Coord) geo.Coord {
	a := 6378137.0
	c := geo.Coord{}
	c[0] = float32(a * float64(point[0]) * math.Pi / 180)
	c[1] = float32(a * math.Log(math.Tan(math.Pi/4+float64(point[1])*math.Pi/360)))
	return c
}
func (self *WebMercatorProjection) ReProj(point geo.Coord) geo.Coord {
	a := 6378137.0
	c := geo.Coord{}
	c[0] = float32(float64(point[0]) * 180 / (a * math.Pi))
	c[1] = float32(360 * (math.Atan(math.Exp(float64(point[1])/a)) - math.Pi/4) / math.Pi)
	return c
}

//**********************************************************
// isochrone extractor
//**********************************************************

func _ExtractIsochrone(points *IsoTree[int], mq_tree *IsoTree[int], max_range int, rasterizer IRasterizer, projection IProjection) List[geo.Feature] {
	// create marching-square tree
	mq_tree.UpdateAll(-1)
	for cell := range points.Traverse() {
		x := cell.A
		y := cell.B
		if cell.C > max_range {
			continue
		}
		index, exist := mq_tree.Get(x, y)
		if !exist || index == -1 {
			mq_tree.InsertValue(x, y, _GetMarchingSquareIndex(points, x, y, max_range))
		}
		x = cell.A
		y = cell.B - 1
		index, exist = mq_tree.Get(x, y)
		if !exist || index == -1 {
			mq_tree.InsertValue(x, y, _GetMarchingSquareIndex(points, x, y, max_range))
		}
		x = cell.A - 1
		y = cell.B
		index, exist = mq_tree.Get(x, y)
		if !exist || index == -1 {
			mq_tree.InsertValue(x, y, _GetMarchingSquareIndex(points, x, y, max_range))
		}
		x = cell.A - 1
		y = cell.B - 1
		index, exist = mq_tree.Get(x, y)
		if !exist || index == -1 {
			mq_tree.InsertValue(x, y, _GetMarchingSquareIndex(points, x, y, max_range))
		}
	}
	// extract polygons
	polygons := NewList[geo.CoordArray](0)
	for node := range mq_tree.Traverse() {
		if node.C == -1 {
			continue
		}
		startx := node.A
		starty := node.B
		currindex := node.C
		if currindex == 0 || currindex == 15 {
			mq_tree.UpdateValue(startx, starty, -1)
			continue
		}
		currx := startx
		curry := starty
		currdirection := -1
		polygon := make(geo.CoordArray, 0)
		for {
			nextx, nexty, nextdirection, coord, replace := _ProcessSquare(currindex, currdirection, currx, curry, rasterizer)
			if replace == -1 {
				mq_tree.UpdateValue(currx, curry, -1)
			} else {
				mq_tree.UpdateValue(currx, curry, replace)
			}
			// add coordinate to polygon
			polygon = append(polygon, projection.ReProj(coord))
			if nextx == startx && nexty == starty {
				polygons.Add(polygon)
				break
			}
			nextindex, ok := mq_tree.Get(nextx, nexty)
			if !ok || nextindex == -1 {
				// break
				panic("not implemented")
			}
			currx = nextx
			curry = nexty
			currdirection = nextdirection
			currindex = nextindex
		}
	}
	// seperate outer and inner polygons
	outers := NewList[geo.CoordArray](0)
	inners := NewList[Tuple[geo.CoordArray, int]](0)
	for _, polygon := range polygons {
		if _PolygonOrientation(polygon) {
			outers.Add(polygon)
		} else {
			inners.Add(MakeTuple(polygon, -1))
		}
	}
	// match inner polygons to outer polygons
	for i, tuple := range inners {
		inner := tuple.A
		outerid := -1
		outerlength := math.MaxInt
		for j, outer := range outers {
			// inner polygons will always be either fully within or not due to marching squares
			if geo.SimplePointInPolygon(inner[0], [][]geo.Coord{outer}) {
				// take the smallest outer polygons
				if outerid == -1 || len(outer) < outerlength {
					outerid = j
					outerlength = len(outer)
				}
			}
		}
		if outerid != -1 {
			inners[i] = MakeTuple(inner, outerid)
		}
	}
	// build geojson features
	features := NewList[geo.Feature](outers.Length())
	for _, outer := range outers {
		polygon := [][]geo.Coord{outer}
		// for _, tuple := range inners {
		// 	inner := tuple.A
		// 	outerid := tuple.B
		// 	if outerid == i {
		// 		polygon = append(polygon, inner)
		// 	}
		// }
		geometry := geo.NewPolygon(polygon)
		properties := NewDict[string, any](1)
		properties["value"] = max_range
		features.Add(geo.NewFeature(&geometry, properties))
	}
	return features
}
